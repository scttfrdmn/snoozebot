package provider

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/signature"
)

// PluginManagerWithSignature extends the Plugin Manager with signature verification
type PluginManagerWithSignature struct {
	baseManager     PluginManager
	signatureService signature.SignatureService
	logger          hclog.Logger
	configDir       string
	signatureEnabled bool
}

// NewPluginManagerWithSignature creates a new plugin manager with signature verification
func NewPluginManagerWithSignature(baseManager PluginManager, configDir string, logger hclog.Logger) (*PluginManagerWithSignature, error) {
	if logger == nil {
		logger = defaultPluginLogger().Named("signature")
	}

	// Create the signature service
	sigService, err := signature.NewSignatureService(configDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature service: %w", err)
	}

	return &PluginManagerWithSignature{
		baseManager:      baseManager,
		signatureService: sigService,
		logger:           logger,
		configDir:        configDir,
		signatureEnabled: true,
	}, nil
}

// EnableSignatureVerification enables or disables signature verification for plugins
func (pm *PluginManagerWithSignature) EnableSignatureVerification(enabled bool) {
	pm.signatureEnabled = enabled
	pm.logger.Info("Plugin signature verification", "enabled", enabled)
}

// IsSignatureVerificationEnabled returns whether signature verification is enabled
func (pm *PluginManagerWithSignature) IsSignatureVerificationEnabled() bool {
	return pm.signatureEnabled
}

// LoadPlugin loads a cloud provider plugin with signature verification
func (pm *PluginManagerWithSignature) LoadPlugin(ctx context.Context, pluginName string) (CloudProvider, error) {
	pm.logger.Info("Loading plugin with signature verification", "name", pluginName)

	// Get the plugin path from the base manager
	pluginPaths, err := pm.baseManager.DiscoverPlugins()
	if err != nil {
		return nil, fmt.Errorf("failed to discover plugins: %w", err)
	}

	// Find the plugin path
	var pluginPath string
	for _, path := range pluginPaths {
		if filepath.Base(path) == pluginName {
			pluginPath = path
			break
		}
	}

	if pluginPath == "" {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	// Verify the plugin signature if enabled
	if pm.signatureEnabled {
		pm.logger.Info("Verifying plugin signature", "name", pluginName, "path", pluginPath)
		if err := pm.signatureService.VerifyPluginSignature(pluginName, pluginPath); err != nil {
			return nil, fmt.Errorf("signature verification failed: %w", err)
		}
		pm.logger.Info("Plugin signature verified successfully", "name", pluginName)
	}

	// Load the plugin using the base manager
	provider, err := pm.baseManager.LoadPlugin(ctx, pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	return provider, nil
}

// UnloadPlugin unloads a cloud provider plugin
func (pm *PluginManagerWithSignature) UnloadPlugin(pluginName string) error {
	return pm.baseManager.UnloadPlugin(pluginName)
}

// GetPlugin gets a loaded cloud provider plugin
func (pm *PluginManagerWithSignature) GetPlugin(pluginName string) (CloudProvider, error) {
	return pm.baseManager.GetPlugin(pluginName)
}

// ListPlugins lists all loaded plugins
func (pm *PluginManagerWithSignature) ListPlugins() []string {
	return pm.baseManager.ListPlugins()
}

// DiscoverPlugins discovers plugins in the plugins directory
func (pm *PluginManagerWithSignature) DiscoverPlugins() ([]string, error) {
	return pm.baseManager.DiscoverPlugins()
}

// GetSignatureService returns the signature service
func (pm *PluginManagerWithSignature) GetSignatureService() signature.SignatureService {
	return pm.signatureService
}

// PluginManagerWithAuthAndSignature combines authentication and signature verification
type PluginManagerWithAuthAndSignature struct {
	authManager      *PluginManagerWithAuth
	signatureManager *PluginManagerWithSignature
	logger           hclog.Logger
}

// NewPluginManagerWithAuthAndSignature creates a new plugin manager with authentication and signature verification
func NewPluginManagerWithAuthAndSignature(baseManager PluginManager, configDir string, logger hclog.Logger) (*PluginManagerWithAuthAndSignature, error) {
	if logger == nil {
		logger = defaultPluginLogger().Named("auth_signature")
	}

	// Create the authentication manager
	authManager, err := NewPluginManagerWithAuth(baseManager, configDir, logger.Named("auth"))
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication manager: %w", err)
	}

	// Create the signature manager
	signatureManager, err := NewPluginManagerWithSignature(baseManager, configDir, logger.Named("signature"))
	if err != nil {
		return nil, fmt.Errorf("failed to create signature manager: %w", err)
	}

	return &PluginManagerWithAuthAndSignature{
		authManager:      authManager,
		signatureManager: signatureManager,
		logger:           logger,
	}, nil
}

// LoadPlugin loads a cloud provider plugin with authentication and signature verification
func (pm *PluginManagerWithAuthAndSignature) LoadPlugin(ctx context.Context, pluginName string) (CloudProvider, error) {
	pm.logger.Info("Loading plugin with authentication and signature verification", "name", pluginName)

	// Verify the plugin signature
	if pm.signatureManager.IsSignatureVerificationEnabled() {
		// Get the plugin path from the base manager
		pluginPaths, err := pm.authManager.baseManager.DiscoverPlugins()
		if err != nil {
			return nil, fmt.Errorf("failed to discover plugins: %w", err)
		}

		// Find the plugin path
		var pluginPath string
		for _, path := range pluginPaths {
			if filepath.Base(path) == pluginName {
				pluginPath = path
				break
			}
		}

		if pluginPath == "" {
			return nil, fmt.Errorf("plugin %s not found", pluginName)
		}

		// Verify the plugin signature
		pm.logger.Info("Verifying plugin signature", "name", pluginName, "path", pluginPath)
		if err := pm.signatureManager.GetSignatureService().VerifyPluginSignature(pluginName, pluginPath); err != nil {
			return nil, fmt.Errorf("signature verification failed: %w", err)
		}
		pm.logger.Info("Plugin signature verified successfully", "name", pluginName)
	}

	// Load the plugin with authentication
	provider, err := pm.authManager.LoadPlugin(ctx, pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	return provider, nil
}

// UnloadPlugin unloads a cloud provider plugin
func (pm *PluginManagerWithAuthAndSignature) UnloadPlugin(pluginName string) error {
	return pm.authManager.UnloadPlugin(pluginName)
}

// GetPlugin gets a loaded cloud provider plugin
func (pm *PluginManagerWithAuthAndSignature) GetPlugin(pluginName string) (CloudProvider, error) {
	return pm.authManager.GetPlugin(pluginName)
}

// ListPlugins lists all loaded plugins
func (pm *PluginManagerWithAuthAndSignature) ListPlugins() []string {
	return pm.authManager.ListPlugins()
}

// DiscoverPlugins discovers plugins in the plugins directory
func (pm *PluginManagerWithAuthAndSignature) DiscoverPlugins() ([]string, error) {
	return pm.authManager.DiscoverPlugins()
}

// EnableAuthentication enables or disables authentication for plugins
func (pm *PluginManagerWithAuthAndSignature) EnableAuthentication(enabled bool) {
	pm.authManager.EnableAuthentication(enabled)
}

// IsAuthenticationEnabled returns whether authentication is enabled
func (pm *PluginManagerWithAuthAndSignature) IsAuthenticationEnabled() bool {
	return pm.authManager.IsAuthenticationEnabled()
}

// EnableTLS enables or disables TLS for plugins
func (pm *PluginManagerWithAuthAndSignature) EnableTLS(enabled bool) {
	pm.authManager.EnableTLS(enabled)
}

// IsTLSEnabled returns whether TLS is enabled
func (pm *PluginManagerWithAuthAndSignature) IsTLSEnabled() bool {
	return pm.authManager.IsTLSEnabled()
}

// InitializeTLS initializes the TLS manager
func (pm *PluginManagerWithAuthAndSignature) InitializeTLS() error {
	return pm.authManager.InitializeTLS()
}

// EnableSignatureVerification enables or disables signature verification for plugins
func (pm *PluginManagerWithAuthAndSignature) EnableSignatureVerification(enabled bool) {
	pm.signatureManager.EnableSignatureVerification(enabled)
}

// IsSignatureVerificationEnabled returns whether signature verification is enabled
func (pm *PluginManagerWithAuthAndSignature) IsSignatureVerificationEnabled() bool {
	return pm.signatureManager.IsSignatureVerificationEnabled()
}

// GetSignatureService returns the signature service
func (pm *PluginManagerWithAuthAndSignature) GetSignatureService() signature.SignatureService {
	return pm.signatureManager.GetSignatureService()
}

// GenerateAPIKey generates an API key for a plugin
func (pm *PluginManagerWithAuthAndSignature) GenerateAPIKey(pluginName, role, description string, expiresInDays int) (string, error) {
	return pm.authManager.GenerateAPIKey(pluginName, role, description, expiresInDays)
}

// RevokeAPIKey revokes an API key for a plugin
func (pm *PluginManagerWithAuthAndSignature) RevokeAPIKey(pluginName string) error {
	return pm.authManager.RevokeAPIKey(pluginName)
}