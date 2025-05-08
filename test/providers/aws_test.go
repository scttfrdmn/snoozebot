package providers

import (
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/scttfrdmn/snoozebot/pkg/plugin"
)

func TestAWSProvider(t *testing.T) {
	// Create a logger for testing
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "aws-test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	// Create the AWS provider
	var provider plugin.CloudProvider
	var err error

	if isLiveTest() {
		// Check if we have an AWS profile
		profile := os.Getenv("AWS_PROFILE")
		if profile == "" {
			t.Skip("Skipping AWS live test: AWS_PROFILE not set")
			return
		}

		// Try to create the provider with real credentials
		// For now, we'll use the mock provider
		provider = NewMockProvider("aws", logger)
		// In a real implementation, we would do something like:
		// provider, err = aws.NewAWSProvider(logger)
		// require.NoError(t, err, "Failed to create AWS provider")
	} else {
		// Skip for now, will use mock provider instead
		t.Skip("Skipping AWS provider test in mock mode")
		return
	}

	// Verify instance ID for start/stop testing
	var instanceID string
	if isLiveTest() && os.Getenv("SNOOZEBOT_TEST_START_STOP") == "true" {
		instanceID = os.Getenv("AWS_TEST_INSTANCE_ID")
		assert.NotEmpty(t, instanceID, "AWS_TEST_INSTANCE_ID must be set for start/stop testing")
	}

	// Run the provider tests
	test := &ProviderOperationTest{
		Provider:      provider,
		InstanceID:    instanceID,
		SkipStartStop: instanceID == "",
	}
	test.Run(t)
}

func TestMockAWSProvider(t *testing.T) {
	if isLiveTest() {
		t.Skip("Skipping mock test in live mode")
		return
	}

	// Create a logger for testing
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "mock-aws-test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	// Create a mock AWS provider
	mockProvider := NewMockProvider("aws", logger)

	// Run the provider tests
	test := &ProviderOperationTest{
		Provider:      mockProvider,
		SkipStartStop: true, // Skip start/stop tests for mock provider
	}
	test.Run(t)
}