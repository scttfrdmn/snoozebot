package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrAuthDisabled       = errors.New("authentication is disabled")
	ErrPluginNotFound     = errors.New("plugin not found")
	ErrInvalidAPIKey      = errors.New("invalid API key")
	ErrAPIKeyExpired      = errors.New("API key expired")
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrKeyGenerationFail  = errors.New("failed to generate API key")
	ErrConfigSaveFailure  = errors.New("failed to save config")
	ErrConfigLoadFailure  = errors.New("failed to load config")
)

// APIKeyManager manages API keys for plugins
type APIKeyManager struct {
	config     *SecurityConfig
	configPath string
	mutex      sync.RWMutex
}

// NewAPIKeyManager creates a new API key manager
func NewAPIKeyManager(configPath string) (*APIKeyManager, error) {
	config, err := LoadSecurityConfig(configPath)
	if err != nil {
		// If config doesn't exist, create a default one
		if errors.Is(err, ErrConfigLoadFailure) {
			config = &SecurityConfig{
				Auth: AuthConfig{
					Enabled: false,
					APIKeys: []PluginAPIKey{},
					Roles: []PluginRole{
						{
							Name:        "cloud_provider",
							Description: "Role for cloud provider plugins",
							Permissions: []PluginPermission{
								{
									Name:        "cloud_operations",
									Description: "Permission to perform cloud operations",
									Allowed:     true,
								},
								{
									Name:        "filesystem_read",
									Description: "Permission to read from filesystem",
									Allowed:     true,
								},
								{
									Name:        "filesystem_write",
									Description: "Permission to write to filesystem",
									Allowed:     false,
								},
								{
									Name:        "network_outbound",
									Description: "Permission to make outbound network connections",
									Allowed:     true,
								},
							},
						},
					},
				},
				TLS: TLSConfig{
					Enabled: false,
				},
				Signatures: SignatureConfig{
					Enabled: false,
				},
			}

			// Save the default config
			if err := SaveSecurityConfig(config, configPath); err != nil {
				return nil, fmt.Errorf("failed to save default config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load security config: %w", err)
		}
	}

	return &APIKeyManager{
		config:     config,
		configPath: configPath,
		mutex:      sync.RWMutex{},
	}, nil
}

// IsAuthEnabled returns whether authentication is enabled
func (m *APIKeyManager) IsAuthEnabled() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.config.Auth.Enabled
}

// EnableAuth enables authentication
func (m *APIKeyManager) EnableAuth() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.config.Auth.Enabled = true
	return SaveSecurityConfig(m.config, m.configPath)
}

// DisableAuth disables authentication
func (m *APIKeyManager) DisableAuth() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.config.Auth.Enabled = false
	return SaveSecurityConfig(m.config, m.configPath)
}

// GenerateAPIKey generates a new API key for a plugin
func (m *APIKeyManager) GenerateAPIKey(pluginName, role, description string, expiresInDays int) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if role exists
	var roleExists bool
	for _, r := range m.config.Auth.Roles {
		if r.Name == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return "", ErrRoleNotFound
	}

	// Generate a secure API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("%w: %v", ErrKeyGenerationFail, err)
	}
	apiKey := base64.URLEncoding.EncodeToString(keyBytes)

	// Set expiration time
	now := time.Now()
	expiresAt := now.AddDate(0, 0, expiresInDays)

	// Create new API key record
	newKey := PluginAPIKey{
		PluginName:  pluginName,
		APIKey:      apiKey,
		Role:        role,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		Description: description,
	}

	// Remove any existing API key for this plugin
	var filteredKeys []PluginAPIKey
	for _, key := range m.config.Auth.APIKeys {
		if key.PluginName != pluginName {
			filteredKeys = append(filteredKeys, key)
		}
	}
	m.config.Auth.APIKeys = append(filteredKeys, newKey)

	// Save the updated config
	if err := SaveSecurityConfig(m.config, m.configPath); err != nil {
		return "", fmt.Errorf("%w: %v", ErrConfigSaveFailure, err)
	}

	return apiKey, nil
}

// ValidateAPIKey validates an API key for a plugin
func (m *APIKeyManager) ValidateAPIKey(pluginName, apiKey string) (bool, string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check if authentication is enabled
	if !m.config.Auth.Enabled {
		// If auth is disabled, return success but with empty role
		return true, "", nil
	}

	// Find the API key for the plugin
	var pluginKey *PluginAPIKey
	for i, key := range m.config.Auth.APIKeys {
		if key.PluginName == pluginName && key.APIKey == apiKey {
			pluginKey = &m.config.Auth.APIKeys[i]
			break
		}
	}

	if pluginKey == nil {
		return false, "", ErrInvalidAPIKey
	}

	// Check if the API key has expired
	if time.Now().After(pluginKey.ExpiresAt) {
		return false, "", ErrAPIKeyExpired
	}

	// Update last used time (note: this is a read operation so not saved to disk)
	pluginKey.LastUsedAt = time.Now()

	return true, pluginKey.Role, nil
}

// CheckPermission checks if a plugin has a specific permission
func (m *APIKeyManager) CheckPermission(pluginName, permission string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check if authentication is enabled
	if !m.config.Auth.Enabled {
		// If auth is disabled, all permissions are allowed
		return true, nil
	}

	// Find the API key for the plugin
	var pluginKey *PluginAPIKey
	for i, key := range m.config.Auth.APIKeys {
		if key.PluginName == pluginName {
			pluginKey = &m.config.Auth.APIKeys[i]
			break
		}
	}

	if pluginKey == nil {
		return false, ErrPluginNotFound
	}

	// Find the role for the plugin
	var pluginRole *PluginRole
	for i, role := range m.config.Auth.Roles {
		if role.Name == pluginKey.Role {
			pluginRole = &m.config.Auth.Roles[i]
			break
		}
	}

	if pluginRole == nil {
		return false, ErrRoleNotFound
	}

	// Check if the permission is allowed
	for _, perm := range pluginRole.Permissions {
		if perm.Name == permission {
			return perm.Allowed, nil
		}
	}

	// If permission not found, deny by default
	return false, nil
}

// GetAPIKey gets an API key for a plugin
func (m *APIKeyManager) GetAPIKey(pluginName string) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Find the API key for the plugin
	for _, key := range m.config.Auth.APIKeys {
		if key.PluginName == pluginName {
			return key.APIKey, nil
		}
	}

	return "", ErrPluginNotFound
}

// RevokeAPIKey revokes an API key for a plugin
func (m *APIKeyManager) RevokeAPIKey(pluginName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if the plugin exists
	var pluginExists bool
	var filteredKeys []PluginAPIKey
	for _, key := range m.config.Auth.APIKeys {
		if key.PluginName != pluginName {
			filteredKeys = append(filteredKeys, key)
		} else {
			pluginExists = true
		}
	}

	if !pluginExists {
		return ErrPluginNotFound
	}

	// Update the API keys
	m.config.Auth.APIKeys = filteredKeys

	// Save the updated config
	if err := SaveSecurityConfig(m.config, m.configPath); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigSaveFailure, err)
	}

	return nil
}

// GetAllPluginAPIKeys gets all plugin API keys
func (m *APIKeyManager) GetAllPluginAPIKeys() ([]PluginAPIKey, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a copy of the API keys
	keys := make([]PluginAPIKey, len(m.config.Auth.APIKeys))
	copy(keys, m.config.Auth.APIKeys)

	return keys, nil
}

// GetRoles gets all roles
func (m *APIKeyManager) GetRoles() ([]PluginRole, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a copy of the roles
	roles := make([]PluginRole, len(m.config.Auth.Roles))
	copy(roles, m.config.Auth.Roles)

	return roles, nil
}

// AddRole adds a new role
func (m *APIKeyManager) AddRole(role PluginRole) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if role already exists
	for _, r := range m.config.Auth.Roles {
		if r.Name == role.Name {
			return fmt.Errorf("role already exists: %s", role.Name)
		}
	}

	// Add the role
	m.config.Auth.Roles = append(m.config.Auth.Roles, role)

	// Save the updated config
	if err := SaveSecurityConfig(m.config, m.configPath); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigSaveFailure, err)
	}

	return nil
}

// UpdateRole updates an existing role
func (m *APIKeyManager) UpdateRole(role PluginRole) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Find the role
	for i, r := range m.config.Auth.Roles {
		if r.Name == role.Name {
			// Update the role
			m.config.Auth.Roles[i] = role

			// Save the updated config
			if err := SaveSecurityConfig(m.config, m.configPath); err != nil {
				return fmt.Errorf("%w: %v", ErrConfigSaveFailure, err)
			}

			return nil
		}
	}

	return ErrRoleNotFound
}

// DeleteRole deletes a role
func (m *APIKeyManager) DeleteRole(roleName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if role is in use
	for _, key := range m.config.Auth.APIKeys {
		if key.Role == roleName {
			return fmt.Errorf("role is in use by plugin: %s", key.PluginName)
		}
	}

	// Find the role
	var roleExists bool
	var filteredRoles []PluginRole
	for _, r := range m.config.Auth.Roles {
		if r.Name != roleName {
			filteredRoles = append(filteredRoles, r)
		} else {
			roleExists = true
		}
	}

	if !roleExists {
		return ErrRoleNotFound
	}

	// Update the roles
	m.config.Auth.Roles = filteredRoles

	// Save the updated config
	if err := SaveSecurityConfig(m.config, m.configPath); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigSaveFailure, err)
	}

	return nil
}