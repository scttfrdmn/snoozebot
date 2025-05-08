package providers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/snoozebot/pkg/plugin"
)

// TestProviderInterface ensures that a provider properly implements the CloudProvider interface
func TestProviderInterface(t *testing.T, provider plugin.CloudProvider) {
	t.Helper()

	// Test GetProviderName
	t.Run("GetProviderName", func(t *testing.T) {
		name := provider.GetProviderName()
		assert.NotEmpty(t, name, "Provider name should not be empty")
	})

	// Test GetProviderVersion
	t.Run("GetProviderVersion", func(t *testing.T) {
		version := provider.GetProviderVersion()
		assert.NotEmpty(t, version, "Provider version should not be empty")
	})

	// Test GetAPIVersion
	t.Run("GetAPIVersion", func(t *testing.T) {
		apiVersion := provider.GetAPIVersion()
		assert.Equal(t, plugin.CurrentAPIVersion, apiVersion, "API version should match current version")
	})

	// Call Shutdown to verify it doesn't panic
	t.Run("Shutdown", func(t *testing.T) {
		assert.NotPanics(t, func() {
			provider.Shutdown()
		}, "Shutdown should not panic")
	})
}

// MockInstance creates a mock instance info for testing
func MockInstance(id, name, instanceType, region, zone, state string) *plugin.InstanceInfo {
	return &plugin.InstanceInfo{
		ID:         id,
		Name:       name,
		Type:       instanceType,
		Region:     region,
		Zone:       zone,
		State:      state,
		LaunchTime: time.Now().Add(-24 * time.Hour), // 1 day ago
	}
}

// ProviderOperationTest provides a standard test for provider operations
type ProviderOperationTest struct {
	Provider      plugin.CloudProvider
	InstanceID    string
	SkipStartStop bool // Set to true for mock providers that can't actually start/stop
}

// Run executes the standard tests for provider operations
func (test *ProviderOperationTest) Run(t *testing.T) {
	if test.Provider == nil {
		t.Fatal("Provider is nil")
	}

	// Setup test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// First, test the CloudProvider interface implementation
	TestProviderInterface(t, test.Provider)

	// Test GetInstanceInfo
	t.Run("GetInstanceInfo", func(t *testing.T) {
		info, err := test.Provider.GetInstanceInfo(ctx)
		if !isLiveTest() {
			// If not a live test, we should get a mock or error
			if err != nil {
				t.Logf("Expected error in mock mode: %v", err)
				return
			}
		} else {
			require.NoError(t, err, "GetInstanceInfo should not error")
			require.NotNil(t, info, "InstanceInfo should not be nil")
			assert.NotEmpty(t, info.ID, "Instance ID should not be empty")
			t.Logf("Instance info: ID=%s, Name=%s, Type=%s, State=%s", 
				info.ID, info.Name, info.Type, info.State)
		}
	})

	// Test ListInstances
	t.Run("ListInstances", func(t *testing.T) {
		instances, err := test.Provider.ListInstances(ctx)
		require.NoError(t, err, "ListInstances should not error")
		
		if isLiveTest() {
			// In live test mode, we may or may not have instances
			t.Logf("Found %d instances", len(instances))
			for i, inst := range instances {
				t.Logf("Instance %d: ID=%s, Name=%s, Type=%s, State=%s", 
					i, inst.ID, inst.Name, inst.Type, inst.State)
			}
		} else {
			// In mock mode, we should have at least some instances
			assert.NotEmpty(t, instances, "ListInstances should return at least one instance in mock mode")
		}
	})

	// Only test Start/Stop in live mode and if not explicitly skipped
	if isLiveTest() && !test.SkipStartStop {
		// Test StartInstance and StopInstance
		// Note: This is potentially destructive, so we have additional safeguards
		if os.Getenv("SNOOZEBOT_TEST_START_STOP") == "true" && test.InstanceID != "" {
			// We need a real instance ID to test with
			os.Setenv("INSTANCE_ID", test.InstanceID)
			
			t.Run("StopInstance", func(t *testing.T) {
				err := test.Provider.StopInstance(ctx)
				require.NoError(t, err, "StopInstance should not error")
			})

			// Wait a bit for instance to transition
			time.Sleep(5 * time.Second)

			t.Run("StartInstance", func(t *testing.T) {
				err := test.Provider.StartInstance(ctx)
				require.NoError(t, err, "StartInstance should not error")
			})
		} else {
			t.Log("Skipping StartInstance/StopInstance tests. Set SNOOZEBOT_TEST_START_STOP=true and provide a real instance ID to enable")
		}
	}
}

// isLiveTest returns true if we're running live tests against actual providers
func isLiveTest() bool {
	return os.Getenv("SNOOZEBOT_LIVE_TESTS") == "true"
}