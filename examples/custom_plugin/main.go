package main

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	snoozePlugin "github.com/scttfrdmn/snoozebot/pkg/plugin"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/version"
)

// CustomProvider is an example implementation of CloudProvider
type CustomProvider struct {
	*snoozePlugin.BaseProvider
	logger hclog.Logger
	// Add your custom fields here
}

// NewCustomProvider creates a new custom provider
func NewCustomProvider(logger hclog.Logger) (*CustomProvider, error) {
	// Initialize the base provider
	baseProvider := snoozePlugin.NewBaseProvider("custom-provider", "0.1.0", logger)
	
	// Create the provider instance
	provider := &CustomProvider{
		BaseProvider: baseProvider,
		logger:       logger,
		// Initialize your custom fields here
	}
	
	// Add capabilities
	provider.AddCapability("list_instances")
	provider.AddCapability("start_instance")
	provider.AddCapability("stop_instance")
	
	// Create and set the manifest
	manifest := version.NewPluginManifest("custom-provider", "0.1.0", "Custom cloud provider example")
	manifest.Author = "Your Name"
	manifest.License = "Apache-2.0"
	manifest.Repository = "https://github.com/your-username/snoozebot-custom-plugin"
	manifest.MinHostVersion = "0.1.0"
	manifest.SupportedProviders = []string{"custom-cloud"}
	manifest.APIVersion = "0.1.0" // Make sure API version matches project version
	provider.SetManifest(manifest)
	
	return provider, nil
}

// GetInstanceInfo gets information about the current instance
func (p *CustomProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.InstanceInfo, error) {
	p.logger.Info("Getting instance info")
	
	// In a real implementation, you would query your cloud provider API
	// For this example, we'll return a mock instance
	return &snoozePlugin.InstanceInfo{
		ID:         "i-custom123",
		Name:       "custom-instance",
		Type:       "custom.micro",
		Region:     "custom-region-1",
		Zone:       "custom-zone-a",
		State:      "running",
		LaunchTime: time.Now().Add(-24 * time.Hour), // Launched yesterday
	}, nil
}

// StopInstance stops the current instance
func (p *CustomProvider) StopInstance(ctx context.Context) error {
	p.logger.Info("Stopping instance")
	// In a real implementation, you would call your cloud provider API
	return nil
}

// StartInstance starts the current instance
func (p *CustomProvider) StartInstance(ctx context.Context) error {
	p.logger.Info("Starting instance")
	// In a real implementation, you would call your cloud provider API
	return nil
}

// ListInstances lists all instances
func (p *CustomProvider) ListInstances(ctx context.Context) ([]*snoozePlugin.InstanceInfo, error) {
	p.logger.Info("Listing instances")
	
	// In a real implementation, you would query your cloud provider API
	// For this example, we'll return mock instances
	instances := []*snoozePlugin.InstanceInfo{
		{
			ID:         "i-custom123",
			Name:       "custom-instance-1",
			Type:       "custom.micro",
			Region:     "custom-region-1",
			Zone:       "custom-zone-a",
			State:      "running",
			LaunchTime: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:         "i-custom456",
			Name:       "custom-instance-2",
			Type:       "custom.large",
			Region:     "custom-region-1",
			Zone:       "custom-zone-b",
			State:      "stopped",
			LaunchTime: time.Now().Add(-48 * time.Hour),
		},
	}
	
	return instances, nil
}

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "custom-provider",
		Level:      hclog.Info,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	// Create the provider
	customProvider, err := NewCustomProvider(logger)
	if err != nil {
		logger.Error("Failed to create custom provider", "error", err)
		os.Exit(1)
	}

	// Check API version compatibility
	compatible, err := customProvider.CheckVersionCompatibility()
	if err != nil {
		logger.Error("Failed to check API version compatibility", "error", err)
		os.Exit(1)
	}
	if !compatible {
		logger.Error("Plugin API version is not compatible with host",
			"plugin_version", customProvider.GetAPIVersion(),
			"current_version", snoozePlugin.CurrentAPIVersion)
		os.Exit(1)
	}

	// Save the manifest for the plugin
	manifestDir := os.Getenv("SNOOZEBOT_MANIFEST_DIR")
	if manifestDir == "" {
		manifestDir = "."
	}
	
	err = version.SaveManifest(customProvider.GetManifest(), manifestDir)
	if err != nil {
		logger.Warn("Failed to save plugin manifest", "error", err)
	}

	// Check for TLS configuration
	tlsEnabled := os.Getenv("SNOOZEBOT_TLS_ENABLED") == "true"
	if tlsEnabled {
		logger.Info("TLS enabled for plugin communication")
		
		// Set up TLS options
		tlsOptions := &snoozePlugin.TLSOptions{
			Enabled: true,
		}
		
		// Check for custom cert paths
		certFile := os.Getenv("SNOOZEBOT_TLS_CERT_FILE")
		keyFile := os.Getenv("SNOOZEBOT_TLS_KEY_FILE")
		caFile := os.Getenv("SNOOZEBOT_TLS_CA_FILE")
		certDir := os.Getenv("SNOOZEBOT_TLS_CERT_DIR")
		
		if certFile != "" && keyFile != "" {
			tlsOptions.CertFile = certFile
			tlsOptions.KeyFile = keyFile
			tlsOptions.CACert = caFile
			logger.Info("Using provided TLS certificates", "cert", certFile, "key", keyFile, "ca", caFile)
		} else if certDir != "" {
			tlsOptions.CertDir = certDir
			logger.Info("Using TLS certificates from directory", "dir", certDir)
		} else {
			logger.Warn("TLS is enabled but no certificates specified, falling back to insecure mode")
			tlsEnabled = false
		}
		
		// Serve with TLS if enabled
		if tlsEnabled {
			// Serve the plugin with TLS
			snoozePlugin.ServePluginWithTLS(customProvider, tlsOptions, logger)
			return
		}
	}
	
	// Serve the plugin without TLS
	logger.Info("Serving plugin without TLS")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: snoozePlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			"cloud_provider": &snoozePlugin.CloudProviderPlugin{
				Impl: customProvider,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}