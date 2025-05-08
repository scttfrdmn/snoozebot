package providers

import (
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGCPProvider(t *testing.T) {
	// Create a logger for testing
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "gcp-test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	if isLiveTest() {
		// Check if we have GCP credentials
		projectID := os.Getenv("PROJECT_ID")
		credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if projectID == "" || credentialsFile == "" {
			t.Skip("Skipping GCP live test: PROJECT_ID or GOOGLE_APPLICATION_CREDENTIALS not set")
			return
		}

		// Check if credentialsFile exists
		_, err := os.Stat(credentialsFile)
		if os.IsNotExist(err) {
			t.Skipf("Skipping GCP live test: credentials file %s does not exist", credentialsFile)
			return
		}

		// Check if we should import the GCP plugin
		// For now, skip to avoid import cycles
		t.Skip("GCP live test not implemented yet")

		// TODO: Implement GCP provider test with real credentials
		// provider, err := gcp.NewGCPProvider(logger)
		// require.NoError(t, err, "Failed to create GCP provider")
	} else {
		// Skip for now, will use mock provider instead
		t.Skip("Skipping GCP provider test in mock mode")
		return
	}
}

func TestMockGCPProvider(t *testing.T) {
	if isLiveTest() {
		t.Skip("Skipping mock test in live mode")
		return
	}

	// Create a logger for testing
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "mock-gcp-test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	// Create a mock GCP provider
	mockProvider := NewMockProvider("gcp", logger)

	// Run the provider tests
	test := &ProviderOperationTest{
		Provider:      mockProvider,
		SkipStartStop: true, // Skip start/stop tests for mock provider
	}
	test.Run(t)
}