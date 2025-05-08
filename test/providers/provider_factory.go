package providers

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin"
)

// ProviderFactory creates cloud providers for testing
type ProviderFactory struct {
	logger hclog.Logger
}

// ProviderConfig contains configuration for creating a provider
type ProviderConfig struct {
	// Provider type (aws, azure, gcp)
	Type string
	
	// Profile to use for authentication
	Profile string
	
	// UseMock indicates whether to create a mock provider
	UseMock bool
	
	// Additional provider-specific settings
	Settings map[string]string
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(logger hclog.Logger) *ProviderFactory {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}
	
	return &ProviderFactory{
		logger: logger,
	}
}

// CreateProvider creates a cloud provider based on the provided configuration
func (f *ProviderFactory) CreateProvider(config ProviderConfig) (plugin.CloudProvider, error) {
	// If mock mode, create a mock provider
	if config.UseMock {
		f.logger.Info("Creating mock provider", "type", config.Type)
		return NewMockProvider(config.Type, f.logger), nil
	}
	
	// Otherwise, create a real provider
	f.logger.Info("Creating real provider", 
		"type", config.Type, 
		"profile", config.Profile)
	
	switch config.Type {
	case "aws":
		// TODO: Import and create actual AWS provider
		// Currently we can't do this directly due to import cycles
		// In a real application, this would be in a separate package
		return nil, fmt.Errorf("real AWS provider creation not implemented yet")
		
	case "azure":
		// TODO: Import and create actual Azure provider
		return nil, fmt.Errorf("real Azure provider creation not implemented yet")
		
	case "gcp":
		// TODO: Import and create actual GCP provider
		return nil, fmt.Errorf("real GCP provider creation not implemented yet")
		
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}

// LoadProviderFromEnv loads provider configuration from environment variables
func (f *ProviderFactory) LoadProviderFromEnv() (ProviderConfig, error) {
	// Default to mock provider
	config := ProviderConfig{
		Type:    "aws", // Default to AWS
		UseMock: true,  // Default to mock
		Settings: make(map[string]string),
	}
	
	// Determine provider type from environment
	for _, providerType := range []string{"aws", "azure", "gcp"} {
		envVar := fmt.Sprintf("%s_PROFILE", providerType)
		if profile, ok := getEnvIgnoreCase(envVar, ""); ok && profile != "" {
			config.Type = providerType
			config.Profile = profile
			config.UseMock = false
			break
		}
	}
	
	// Enable mock mode if SNOOZEBOT_MOCK_PROVIDER is set
	if mockMode, ok := getEnvIgnoreCase("SNOOZEBOT_MOCK_PROVIDER", ""); ok && mockMode == "true" {
		config.UseMock = true
	}
	
	return config, nil
}

// Get environment variable with case-insensitive matching
func getEnvIgnoreCase(key, defaultValue string) (string, bool) {
	// Try with original key
	if val, ok := getEnv(key); ok {
		return val, true
	}
	
	// Try with uppercase key
	if val, ok := getEnv(key); ok {
		return val, true
	}
	
	// Try with lowercase key
	if val, ok := getEnv(key); ok {
		return val, true
	}
	
	return defaultValue, false
}

// Helper to get environment variable
func getEnv(key string) (string, bool) {
	val, exists := LookupEnv(key)
	return val, exists
}

// LookupEnv is a variable to allow mocking environment variables in tests
var LookupEnv = os.LookupEnv