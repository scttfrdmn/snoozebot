package providers

import (
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureProvider(t *testing.T) {
	// Create a logger for testing
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "azure-test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	if isLiveTest() {
		// Check if we have Azure credentials
		subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
		resourceGroup := os.Getenv("AZURE_RESOURCE_GROUP")
		if subscriptionID == "" || resourceGroup == "" {
			t.Skip("Skipping Azure live test: AZURE_SUBSCRIPTION_ID or AZURE_RESOURCE_GROUP not set")
			return
		}

		// Check if we should import the Azure plugin
		// For now, skip to avoid import cycles
		t.Skip("Azure live test not implemented yet")

		// TODO: Implement Azure provider test with real credentials
		// provider, err := azure.NewAzureProvider(logger)
		// require.NoError(t, err, "Failed to create Azure provider")
	} else {
		// Skip for now, will use mock provider instead
		t.Skip("Skipping Azure provider test in mock mode")
		return
	}
}

func TestMockAzureProvider(t *testing.T) {
	if isLiveTest() {
		t.Skip("Skipping mock test in live mode")
		return
	}

	// Create a logger for testing
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "mock-azure-test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	// Create a mock Azure provider
	mockProvider := NewMockProvider("azure", logger)

	// Run the provider tests
	test := &ProviderOperationTest{
		Provider:      mockProvider,
		SkipStartStop: true, // Skip start/stop tests for mock provider
	}
	test.Run(t)
}