package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// PluginPermission represents a single permission for a plugin
type PluginPermission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Allowed     bool   `json:"allowed"`
}

// PluginRole represents a role with associated permissions
type PluginRole struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Permissions []PluginPermission `json:"permissions"`
}

// PluginAPIKey represents an API key for a plugin
type PluginAPIKey struct {
	PluginName  string    `json:"plugin_name"`
	APIKey      string    `json:"api_key"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
	Description string    `json:"description,omitempty"`
}

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	Enabled bool           `json:"enabled"`
	APIKeys []PluginAPIKey `json:"api_keys"`
	Roles   []PluginRole   `json:"roles"`
}

// TLSConfig represents the TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
}

// SignatureConfig represents the signature verification configuration
type SignatureConfig struct {
	Enabled          bool   `json:"enabled"`
	PublicKeyPath    string `json:"public_key_path"`
	VerifySignatures bool   `json:"verify_signatures"`
}

// SecurityConfig represents the overall security configuration
type SecurityConfig struct {
	Auth       AuthConfig      `json:"auth"`
	TLS        TLSConfig       `json:"tls"`
	Signatures SignatureConfig `json:"signatures"`
}

// LoadSecurityConfig loads the security configuration from a file
func LoadSecurityConfig(configPath string) (*SecurityConfig, error) {
	// Ensure the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("security config file not found: %s", configPath)
	}

	// Read the config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read security config file: %w", err)
	}

	// Parse the config file
	var config SecurityConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse security config file: %w", err)
	}

	return &config, nil
}

// SaveSecurityConfig saves the security configuration to a file
func SaveSecurityConfig(config *SecurityConfig, configPath string) error {
	// Create the directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal security config: %w", err)
	}

	// Write the config file
	if err := ioutil.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write security config file: %w", err)
	}

	return nil
}