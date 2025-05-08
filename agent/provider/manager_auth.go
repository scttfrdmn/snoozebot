package provider

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/auth"
	plugintls "github.com/scttfrdmn/snoozebot/pkg/plugin/tls"
)

// PluginManagerWithAuth extends the Plugin Manager with authentication capabilities
type PluginManagerWithAuth struct {
	baseManager    PluginManager
	authService    auth.PluginAuthService
	logger         hclog.Logger
	configDir      string
	authEnabled    bool
	tlsEnabled     bool
	tlsManager     *plugintls.TLSManager
	securePlugins  map[string]*plugin.SecurePlugin
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

	// Create TLS manager
	tlsManager, err := plugintls.NewTLSManager(filepath.Join(configDir, "certs"))
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS manager: %w", err)
	}

	return &PluginManagerWithAuth{
		baseManager:   baseManager,
		authService:   authService,
		logger:        logger,
		configDir:     configDir,
		authEnabled:   true,
		tlsEnabled:    false, // TLS disabled by default
		tlsManager:    tlsManager,
		securePlugins: make(map[string]*plugin.SecurePlugin),
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

// EnableTLS enables TLS for plugin operations
func (pm *PluginManagerWithAuth) EnableTLS(enabled bool) {
	pm.tlsEnabled = enabled
	pm.logger.Info("Plugin TLS", "enabled", enabled)
}

// IsTLSEnabled returns whether TLS is enabled
func (pm *PluginManagerWithAuth) IsTLSEnabled() bool {
	return pm.tlsEnabled
}

// InitializeTLS initializes the TLS manager
func (pm *PluginManagerWithAuth) InitializeTLS() error {
	if err := pm.tlsManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize TLS manager: %w", err)
	}
	pm.tlsEnabled = true
	pm.logger.Info("TLS initialized successfully")
	return nil
}

// LoadPlugin loads a cloud provider plugin with authentication and TLS
func (pm *PluginManagerWithAuth) LoadPlugin(ctx context.Context, pluginName string) (CloudProvider, error) {
	// If TLS is enabled, use secure plugin loading
	if pm.tlsEnabled {
		return pm.loadSecurePlugin(ctx, pluginName)
	}

	// Standard plugin loading using the base manager
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

// loadSecurePlugin loads a plugin with TLS security
func (pm *PluginManagerWithAuth) loadSecurePlugin(ctx context.Context, pluginName string) (CloudProvider, error) {
	pm.logger.Info("Loading secure plugin", "name", pluginName, "tls", true)

	// Check if the plugin is already loaded
	if securePlugin, ok := pm.securePlugins[pluginName]; ok {
		return securePlugin.GetCloudProvider(), nil
	}

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

	// Initialize TLS manager if not already initialized
	if !pm.tlsManager.IsInitialized() {
		pm.logger.Info("Initializing TLS manager")
		if err := pm.tlsManager.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize TLS manager: %w", err)
		}
	}

	// Ensure plugin certificate exists
	certFile, keyFile, err := pm.tlsManager.EnsurePluginCertificate(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure plugin certificate: %w", err)
	}
	pm.logger.Info("Plugin certificate ensured", "plugin", pluginName, "cert", certFile, "key", keyFile)

	// Get CA certificate
	caCertFile := filepath.Join(pm.configDir, "certs", "ca", "cert.pem")
	caExists, err := fileExists(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to check CA certificate: %w", err)
	}
	if !caExists {
		return nil, fmt.Errorf("CA certificate not found at %s", caCertFile)
	}
	pm.logger.Info("Using CA certificate", "ca", caCertFile)

	// Create TLS options
	tlsOptions := &plugin.TLSOptions{
		Enabled:    true,
		CertFile:   certFile,
		KeyFile:    keyFile,
		CACert:     caCertFile,
		SkipVerify: false,
	}

	// Create secure plugin
	securePlugin, err := plugin.NewSecurePlugin(pluginName, pluginPath, tlsOptions, pm.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure plugin: %w", err)
	}

	// Set environment variables for plugin TLS
	os.Setenv("SNOOZEBOT_TLS_ENABLED", "true")
	os.Setenv("SNOOZEBOT_TLS_CERT_FILE", certFile)
	os.Setenv("SNOOZEBOT_TLS_KEY_FILE", keyFile)
	os.Setenv("SNOOZEBOT_TLS_CA_FILE", caCertFile)

	// Start the plugin
	pm.logger.Info("Starting secure plugin", "name", pluginName, "path", pluginPath)
	if err := securePlugin.Start(); err != nil {
		return nil, fmt.Errorf("failed to start secure plugin: %w", err)
	}

	// Store the secure plugin
	pm.securePlugins[pluginName] = securePlugin

	provider := securePlugin.GetCloudProvider()
	if provider == nil {
		return nil, fmt.Errorf("plugin did not return a valid cloud provider")
	}

	pm.logger.Info("Secure plugin loaded successfully", "name", pluginName)

	// If authentication is not enabled, return the provider directly
	if !pm.authEnabled {
		return provider, nil
	}

	// Create a new authenticated provider that wraps the secure provider
	return &AuthenticatedProvider{
		baseProvider: provider,
		pluginName:   pluginName,
		authService:  pm.authService,
		logger:       pm.logger.Named(pluginName),
	}, nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// UnloadPlugin unloads a cloud provider plugin
func (pm *PluginManagerWithAuth) UnloadPlugin(pluginName string) error {
	// If TLS is enabled, unload secure plugin
	if pm.tlsEnabled {
		if securePlugin, ok := pm.securePlugins[pluginName]; ok {
			pm.logger.Info("Stopping secure plugin", "name", pluginName)
			
			// Stop the plugin gracefully
			securePlugin.Stop()
			
			// Remove from secure plugins map
			delete(pm.securePlugins, pluginName)
			
			// Optionally clean up certificates if temporary
			cleanupCerts := os.Getenv("SNOOZEBOT_TLS_CLEANUP_CERTS") == "true"
			if cleanupCerts && pm.tlsManager.IsInitialized() {
				pm.logger.Info("Cleaning up certificates for plugin", "name", pluginName)
				if err := pm.tlsManager.CleanupPluginCertificate(pluginName); err != nil {
					pm.logger.Warn("Failed to clean up plugin certificates", "name", pluginName, "error", err)
					// Continue with unloading even if cleanup fails
				}
			}
			
			pm.logger.Info("Secure plugin unloaded successfully", "name", pluginName)
			return nil
		}
	}

	// Standard plugin unloading
	pm.logger.Info("Unloading standard plugin", "name", pluginName)
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