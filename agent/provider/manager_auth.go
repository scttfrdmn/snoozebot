package provider

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/auth"
)

// PluginManagerWithAuth extends the Plugin Manager with authentication capabilities
type PluginManagerWithAuth struct {
	baseManager   PluginManager
	authService   auth.PluginAuthService
	logger        hclog.Logger
	configDir     string
	authEnabled   bool
}

// NewPluginManagerWithAuth creates a new plugin manager with authentication
func NewPluginManagerWithAuth(baseManager PluginManager, configDir string, logger hclog.Logger) (*PluginManagerWithAuth, error) {
	if logger == nil {
		logger = defaultPluginLogger().Named("auth")
	}

	// Create the auth service
	authService, err := auth.NewPluginAuthService(configDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	return &PluginManagerWithAuth{
		baseManager: baseManager,
		authService: authService,
		logger:      logger,
		configDir:   configDir,
		authEnabled: true,
	}, nil
}

// EnableAuthentication enables authentication for plugin operations
func (pm *PluginManagerWithAuth) EnableAuthentication(enabled bool) {
	pm.authEnabled = enabled
	pm.logger.Info("Plugin authentication", "enabled", enabled)
}

// IsAuthenticationEnabled returns whether authentication is enabled
func (pm *PluginManagerWithAuth) IsAuthenticationEnabled() bool {
	return pm.authEnabled
}

// LoadPlugin loads a cloud provider plugin with authentication
func (pm *PluginManagerWithAuth) LoadPlugin(ctx context.Context, pluginName string) (CloudProvider, error) {
	// Load the plugin using the base manager
	provider, err := pm.baseManager.LoadPlugin(ctx, pluginName)
	if err != nil {
		return nil, err
	}

	// If authentication is not enabled, return the provider directly
	if !pm.authEnabled {
		return provider, nil
	}

	// Create a new authenticated provider that wraps the base provider
	return &AuthenticatedProvider{
		baseProvider: provider,
		pluginName:   pluginName,
		authService:  pm.authService,
		logger:       pm.logger.Named(pluginName),
	}, nil
}

// UnloadPlugin unloads a cloud provider plugin
func (pm *PluginManagerWithAuth) UnloadPlugin(pluginName string) error {
	return pm.baseManager.UnloadPlugin(pluginName)
}

// GetPlugin gets a loaded cloud provider plugin
func (pm *PluginManagerWithAuth) GetPlugin(pluginName string) (CloudProvider, error) {
	provider, err := pm.baseManager.GetPlugin(pluginName)
	if err != nil {
		return nil, err
	}

	// If authentication is not enabled, return the provider directly
	if !pm.authEnabled {
		return provider, nil
	}

	// Check if the provider is already authenticated
	if authProvider, ok := provider.(*AuthenticatedProvider); ok {
		return authProvider, nil
	}

	// Create a new authenticated provider that wraps the base provider
	return &AuthenticatedProvider{
		baseProvider: provider,
		pluginName:   pluginName,
		authService:  pm.authService,
		logger:       pm.logger.Named(pluginName),
	}, nil
}

// ListPlugins lists all loaded plugins
func (pm *PluginManagerWithAuth) ListPlugins() []string {
	return pm.baseManager.ListPlugins()
}

// DiscoverPlugins discovers plugins in the plugins directory
func (pm *PluginManagerWithAuth) DiscoverPlugins() ([]string, error) {
	return pm.baseManager.DiscoverPlugins()
}

// GenerateAPIKey generates an API key for a plugin
func (pm *PluginManagerWithAuth) GenerateAPIKey(pluginName, role, description string, expiresInDays int) (string, error) {
	// Access the API key manager through the auth service
	authService, ok := pm.authService.(*auth.PluginAuthServiceImpl)
	if !ok {
		return "", fmt.Errorf("auth service does not support API key generation")
	}

	// Generate an API key using the API key manager
	configPath := filepath.Join(pm.configDir, "security.json")
	apiKeyManager, err := auth.NewAPIKeyManager(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to create API key manager: %w", err)
	}

	return apiKeyManager.GenerateAPIKey(pluginName, role, description, expiresInDays)
}

// RevokeAPIKey revokes an API key for a plugin
func (pm *PluginManagerWithAuth) RevokeAPIKey(pluginName string) error {
	// Access the API key manager through the auth service
	authService, ok := pm.authService.(*auth.PluginAuthServiceImpl)
	if !ok {
		return fmt.Errorf("auth service does not support API key revocation")
	}

	// Revoke the API key using the API key manager
	configPath := filepath.Join(pm.configDir, "security.json")
	apiKeyManager, err := auth.NewAPIKeyManager(configPath)
	if err != nil {
		return fmt.Errorf("failed to create API key manager: %w", err)
	}

	return apiKeyManager.RevokeAPIKey(pluginName)
}

// AuthenticatedProvider is a wrapper around a cloud provider that adds authentication
type AuthenticatedProvider struct {
	baseProvider CloudProvider
	pluginName   string
	authService  auth.PluginAuthService
	logger       hclog.Logger
	authenticated bool
	role         string
}

// Authenticate authenticates the provider
func (p *AuthenticatedProvider) Authenticate(ctx context.Context, apiKey string) (bool, error) {
	// Call the auth service to authenticate the plugin
	success, role, err := p.authService.Authenticate(ctx, p.pluginName, apiKey)
	if err != nil {
		p.logger.Error("Authentication failed", "error", err)
		return false, err
	}

	if !success {
		p.logger.Warn("Authentication unsuccessful")
		return false, nil
	}

	p.authenticated = true
	p.role = role
	p.logger.Info("Authentication successful", "role", role)
	return true, nil
}

// checkPermissionAndAuthentication checks if the provider is authenticated and has the required permission
func (p *AuthenticatedProvider) checkPermissionAndAuthentication(ctx context.Context, permission string) error {
	// If not authenticated, return an error
	if !p.authenticated {
		return fmt.Errorf("provider not authenticated")
	}

	// Check if the provider has the required permission
	allowed, err := p.authService.CheckPermission(ctx, p.pluginName, permission)
	if err != nil {
		p.logger.Error("Permission check failed", "permission", permission, "error", err)
		return fmt.Errorf("permission check failed: %w", err)
	}

	if !allowed {
		p.logger.Warn("Permission denied", "permission", permission)
		return fmt.Errorf("permission denied: %s", permission)
	}

	return nil
}

// GetInstanceInfo gets information about an instance
func (p *AuthenticatedProvider) GetInstanceInfo(ctx context.Context, instanceID string) (*InstanceInfo, error) {
	// Check if the provider is authenticated and has the required permission
	if err := p.checkPermissionAndAuthentication(ctx, "cloud_operations"); err != nil {
		return nil, err
	}

	// Call the base provider
	return p.baseProvider.GetInstanceInfo(ctx, instanceID)
}

// StopInstance stops an instance
func (p *AuthenticatedProvider) StopInstance(ctx context.Context, instanceID string) error {
	// Check if the provider is authenticated and has the required permission
	if err := p.checkPermissionAndAuthentication(ctx, "cloud_operations"); err != nil {
		return err
	}

	// Call the base provider
	return p.baseProvider.StopInstance(ctx, instanceID)
}

// StartInstance starts an instance
func (p *AuthenticatedProvider) StartInstance(ctx context.Context, instanceID string) error {
	// Check if the provider is authenticated and has the required permission
	if err := p.checkPermissionAndAuthentication(ctx, "cloud_operations"); err != nil {
		return err
	}

	// Call the base provider
	return p.baseProvider.StartInstance(ctx, instanceID)
}

// GetProviderName returns the name of the cloud provider
func (p *AuthenticatedProvider) GetProviderName() string {
	return p.baseProvider.GetProviderName()
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *AuthenticatedProvider) GetProviderVersion() string {
	return p.baseProvider.GetProviderVersion()
}