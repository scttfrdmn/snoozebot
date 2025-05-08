package azuretest

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureProviderCreation(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("AZURE_SUBSCRIPTION_ID", "test-subscription")
	os.Setenv("AZURE_RESOURCE_GROUP", "test-resource-group")
	os.Setenv("AZURE_VM_NAME", "test-vm")
	
	// Create a provider (this will create a mock version that doesn't actually connect to Azure)
	// The implementation details for this would be in a mock_azure.go file
	provider, err := NewMockAzureProvider(hclog.NewNullLogger())
	require.NoError(t, err)
	require.NotNil(t, provider)
	
	// Test provider name and version
	assert.Equal(t, "azure", provider.GetProviderName())
	assert.Equal(t, "0.1.0", provider.GetProviderVersion())
}

func TestGetInstanceInfo(t *testing.T) {
	provider, err := NewMockAzureProvider(hclog.NewNullLogger())
	require.NoError(t, err)
	
	// Test getting instance info
	info, err := provider.GetInstanceInfo(context.Background())
	require.NoError(t, err)
	require.NotNil(t, info)
	
	// Validate instance info
	assert.Equal(t, "test-vm", info.ID)
	assert.Equal(t, "test-vm", info.Name)
	assert.Equal(t, "Standard_D2s_v3", info.Type)
	assert.Equal(t, "eastus", info.Region)
	assert.Equal(t, "eastus-1", info.Zone)
	assert.Equal(t, "running", info.State)
}

func TestStopInstance(t *testing.T) {
	provider, err := NewMockAzureProvider(hclog.NewNullLogger())
	require.NoError(t, err)
	
	// Test stopping the instance
	err = provider.StopInstance(context.Background())
	require.NoError(t, err)
}

func TestStartInstance(t *testing.T) {
	provider, err := NewMockAzureProvider(hclog.NewNullLogger())
	require.NoError(t, err)
	
	// Test starting the instance
	err = provider.StartInstance(context.Background())
	require.NoError(t, err)
}