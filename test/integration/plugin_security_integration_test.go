package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/agent/provider"
	"github.com/scttfrdmn/snoozebot/pkg/plugin"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/auth"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/signature"
	plugintls "github.com/scttfrdmn/snoozebot/pkg/plugin/tls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegratedSecurityFeatures tests the integration of all security features
func TestIntegratedSecurityFeatures(t *testing.T) {
	// Create a logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	// Setup test directory
	testDir := filepath.Join(os.TempDir(), "snoozebot-test-"+time.Now().Format("20060102-150405"))
	defer os.RemoveAll(testDir)

	require.NoError(t, os.MkdirAll(testDir, 0755))

	// Prepare subdirectories
	configDir := filepath.Join(testDir, "config")
	certsDir := filepath.Join(testDir, "certs")
	signaturesDir := filepath.Join(testDir, "signatures")

	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(certsDir, 0755))
	require.NoError(t, os.MkdirAll(signaturesDir, 0755))

	// Setup base plugin manager
	baseManager := provider.NewPluginManager(filepath.Join(testDir, "plugins"), logger)

	// Create API key for authentication
	apiKey, err := auth.GenerateAPIKey("admin", []string{"admin"})
	require.NoError(t, err)
	authConfigPath := filepath.Join(configDir, "auth.json")
	authConfig := auth.PluginAuthConfig{
		APIKeys: []auth.APIKey{*apiKey},
	}
	authData, err := json.MarshalIndent(authConfig, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(authConfigPath, authData, 0644))

	// Create a plugin manager with auth
	pluginManager, err := provider.NewPluginManagerWithAuth(baseManager, configDir, logger)
	require.NoError(t, err)

	// Initialize TLS
	require.NoError(t, pluginManager.InitializeTLS())
	pluginManager.EnableTLS(true)

	// Initialize signature verification
	sigService, err := signature.NewSignatureService(signaturesDir, logger)
	require.NoError(t, err)
	require.NoError(t, pluginManager.InitializeSignatureVerification(sigService))
	pluginManager.EnableSignatureVerification(true)

	// Initialize authentication
	pluginManager.EnableAuthentication(true)

	// Set API key for authentication
	os.Setenv("SNOOZEBOT_API_KEY", apiKey.Key)
	defer os.Unsetenv("SNOOZEBOT_API_KEY")

	// Test loading AWS plugin with all security features enabled
	// NOTE: This test requires AWS credentials to be set in the environment
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		t.Run("AWS plugin with all security features", func(t *testing.T) {
			// Load the AWS plugin with all security features enabled
			ctx := context.Background()
			provider, err := pluginManager.LoadPlugin(ctx, "aws")
			assert.NoError(t, err)
			assert.NotNil(t, provider)

			// Get provider name
			name := provider.GetProviderName()
			assert.Equal(t, "aws", name)

			// Unload the plugin
			err = pluginManager.UnloadPlugin("aws")
			assert.NoError(t, err)
		})
	} else {
		t.Skip("Skipping AWS plugin test (no credentials)")
	}

	// Test loading Azure plugin with all security features enabled
	// NOTE: This test requires Azure credentials to be set in the environment
	if os.Getenv("AZURE_SUBSCRIPTION_ID") != "" && os.Getenv("AZURE_RESOURCE_GROUP") != "" {
		t.Run("Azure plugin with all security features", func(t *testing.T) {
			// Load the Azure plugin with all security features enabled
			ctx := context.Background()
			provider, err := pluginManager.LoadPlugin(ctx, "azure")
			assert.NoError(t, err)
			assert.NotNil(t, provider)

			// Get provider name
			name := provider.GetProviderName()
			assert.Equal(t, "azure", name)

			// Unload the plugin
			err = pluginManager.UnloadPlugin("azure")
			assert.NoError(t, err)
		})
	} else {
		t.Skip("Skipping Azure plugin test (no credentials)")
	}

	// Test the performance overhead of security features
	t.Run("Performance benchmark", func(t *testing.T) {
		// Disable security features for baseline measurement
		pluginManager.EnableTLS(false)
		pluginManager.EnableSignatureVerification(false)
		pluginManager.EnableAuthentication(false)

		// Measure baseline performance (plugin loading time without security)
		startBaseline := time.Now()
		ctx := context.Background()
		_, err := pluginManager.LoadPlugin(ctx, "aws")
		assert.NoError(t, err)
		baselineDuration := time.Since(startBaseline)
		err = pluginManager.UnloadPlugin("aws")
		assert.NoError(t, err)

		// Enable all security features
		pluginManager.EnableTLS(true)
		pluginManager.EnableSignatureVerification(true)
		pluginManager.EnableAuthentication(true)

		// Measure performance with all security features
		startSecure := time.Now()
		_, err = pluginManager.LoadPlugin(ctx, "aws")
		assert.NoError(t, err)
		secureDuration := time.Since(startSecure)
		err = pluginManager.UnloadPlugin("aws")
		assert.NoError(t, err)

		// Calculate overhead
		overhead := float64(secureDuration.Milliseconds()) / float64(baselineDuration.Milliseconds())
		t.Logf("Performance overhead with security features: %.2fx (baseline: %s, secure: %s)",
			overhead, baselineDuration, secureDuration)

		// Assert overhead is within acceptable range
		assert.Less(t, overhead, 3.0, "Security overhead should be less than 3x")
	})
}

// TestMultiplePlugins tests loading multiple plugins concurrently
func TestMultiplePlugins(t *testing.T) {
	// Create a logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	// Setup test directory
	testDir := filepath.Join(os.TempDir(), "snoozebot-test-"+time.Now().Format("20060102-150405"))
	defer os.RemoveAll(testDir)

	require.NoError(t, os.MkdirAll(testDir, 0755))

	// Setup base plugin manager
	baseManager := provider.NewPluginManager(filepath.Join(testDir, "plugins"), logger)

	// Check if credentials are available
	awsCredsAvailable := os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != ""
	azureCredsAvailable := os.Getenv("AZURE_SUBSCRIPTION_ID") != "" && os.Getenv("AZURE_RESOURCE_GROUP") != ""

	if !awsCredsAvailable || !azureCredsAvailable {
		t.Skip("Skipping multiple plugin test (credentials missing)")
		return
	}

	t.Run("Load multiple providers", func(t *testing.T) {
		ctx := context.Background()
		
		// Load AWS plugin
		awsProvider, err := baseManager.LoadPlugin(ctx, "aws")
		assert.NoError(t, err)
		assert.NotNil(t, awsProvider)
		assert.Equal(t, "aws", awsProvider.GetProviderName())
		
		// Load Azure plugin
		azureProvider, err := baseManager.LoadPlugin(ctx, "azure")
		assert.NoError(t, err)
		assert.NotNil(t, azureProvider)
		assert.Equal(t, "azure", azureProvider.GetProviderName())
		
		// Test both plugins are working
		awsName := awsProvider.GetProviderName()
		azureName := azureProvider.GetProviderName()
		
		assert.Equal(t, "aws", awsName)
		assert.Equal(t, "azure", azureName)
		
		// Unload plugins
		err = baseManager.UnloadPlugin("aws")
		assert.NoError(t, err)
		
		err = baseManager.UnloadPlugin("azure")
		assert.NoError(t, err)
	})
}