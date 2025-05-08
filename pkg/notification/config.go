package notification

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"gopkg.in/yaml.v2"
)

// ProviderConfig represents the configuration for a notification provider
type ProviderConfig struct {
	// Enabled indicates whether the provider is enabled
	Enabled bool `yaml:"enabled"`

	// Config is the provider-specific configuration
	Config map[string]interface{} `yaml:"config"`
}

// Config represents the notification configuration
type Config struct {
	// Providers is a map of provider names to their configurations
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// LoadConfig loads the notification configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	// Read the config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// InitManagerFromConfig initializes a notification manager from a config file
func InitManagerFromConfig(configPath string, logger hclog.Logger) (*Manager, error) {
	// Check if the config file exists, if not, create a default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	// Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create the manager
	manager := NewManager(logger)

	// Register and initialize providers
	for name, providerConfig := range config.Providers {
		if !providerConfig.Enabled {
			logger.Info("Provider is disabled, skipping", "name", name)
			continue
		}

		// Get the provider factory
		factory, exists := providerFactories[name]
		if !exists {
			logger.Warn("Unknown provider, skipping", "name", name)
			continue
		}

		// Create and register the provider
		provider := factory(logger)
		if err := manager.RegisterProvider(provider); err != nil {
			logger.Error("Failed to register provider", "name", name, "error", err)
			continue
		}

		// Initialize the provider
		if err := manager.InitProvider(name, providerConfig.Config); err != nil {
			logger.Error("Failed to initialize provider", "name", name, "error", err)
			continue
		}
	}

	return manager, nil
}

// createDefaultConfig creates a default notification config file
func createDefaultConfig(configPath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a default config
	defaultConfig := Config{
		Providers: map[string]ProviderConfig{
			"slack": {
				Enabled: false,
				Config: map[string]interface{}{
					"webhook_url": "https://hooks.slack.com/services/your/webhook/url",
					"channel":     "#snoozebot",
					"username":    "Snoozebot",
					"icon_emoji":  ":robot_face:",
				},
			},
		},
	}

	// Marshal the config
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	// Write the config file
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	return nil
}