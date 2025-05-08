package azuretest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	snoozePlugin "github.com/scttfrdmn/snoozebot/pkg/plugin"
)

type MockAzureProvider struct {
	logger            hclog.Logger
	subscriptionID    string
	resourceGroupName string
	currentVMName     string
	location          string
}

// NewMockAzureProvider creates a new mock Azure provider for testing
func NewMockAzureProvider(logger hclog.Logger) (*MockAzureProvider, error) {
	return &MockAzureProvider{
		logger:            logger,
		subscriptionID:    "test-subscription",
		resourceGroupName: "test-resource-group",
		currentVMName:     "test-vm",
		location:          "eastus",
	}, nil
}

// GetInstanceInfo returns mock instance info
func (p *MockAzureProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.InstanceInfo, error) {
	return &snoozePlugin.InstanceInfo{
		ID:         "test-vm",
		Name:       "test-vm",
		Type:       "Standard_D2s_v3",
		Region:     "eastus",
		Zone:       "eastus-1",
		State:      "running",
		LaunchTime: time.Now(),
	}, nil
}

// StopInstance simulates stopping an instance
func (p *MockAzureProvider) StopInstance(ctx context.Context) error {
	p.logger.Info("Mock: Stopping instance", "vm_name", p.currentVMName)
	return nil
}

// StartInstance simulates starting an instance
func (p *MockAzureProvider) StartInstance(ctx context.Context) error {
	p.logger.Info("Mock: Starting instance", "vm_name", p.currentVMName)
	return nil
}

// GetProviderName returns the provider name
func (p *MockAzureProvider) GetProviderName() string {
	return "azure"
}

// GetProviderVersion returns the provider version
func (p *MockAzureProvider) GetProviderVersion() string {
	return "0.1.0"
}

// Shutdown performs cleanup when the plugin is being unloaded
func (p *MockAzureProvider) Shutdown() {
	p.logger.Info("Mock: Shutting down Azure provider")
}

func TestAzureProviderCreation(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("AZURE_SUBSCRIPTION_ID", "test-subscription")
	os.Setenv("AZURE_RESOURCE_GROUP", "test-resource-group")
	os.Setenv("AZURE_VM_NAME", "test-vm")
	
	// Create a provider
	provider, err := NewMockAzureProvider(hclog.NewNullLogger())
	if err != nil {
		t.Fatalf("Failed to create Azure provider: %v", err)
	}
	if provider == nil {
		t.Fatal("Provider is nil")
	}
	
	// Test provider name and version
	if provider.GetProviderName() != "azure" {
		t.Errorf("Expected provider name to be 'azure', got %s", provider.GetProviderName())
	}
	if provider.GetProviderVersion() != "0.1.0" {
		t.Errorf("Expected provider version to be '0.1.0', got %s", provider.GetProviderVersion())
	}
}

func TestGetInstanceInfo(t *testing.T) {
	provider, err := NewMockAzureProvider(hclog.NewNullLogger())
	if err != nil {
		t.Fatalf("Failed to create Azure provider: %v", err)
	}
	
	// Test getting instance info
	info, err := provider.GetInstanceInfo(context.Background())
	if err != nil {
		t.Fatalf("Failed to get instance info: %v", err)
	}
	if info == nil {
		t.Fatal("Instance info is nil")
	}
	
	// Validate instance info
	if info.ID != "test-vm" {
		t.Errorf("Expected instance ID to be 'test-vm', got %s", info.ID)
	}
	if info.Name != "test-vm" {
		t.Errorf("Expected instance name to be 'test-vm', got %s", info.Name)
	}
	if info.Type != "Standard_D2s_v3" {
		t.Errorf("Expected instance type to be 'Standard_D2s_v3', got %s", info.Type)
	}
}

func TestStopInstance(t *testing.T) {
	provider, err := NewMockAzureProvider(hclog.NewNullLogger())
	if err != nil {
		t.Fatalf("Failed to create Azure provider: %v", err)
	}
	
	// Test stopping the instance
	err = provider.StopInstance(context.Background())
	if err != nil {
		t.Fatalf("Failed to stop instance: %v", err)
	}
}

func TestStartInstance(t *testing.T) {
	provider, err := NewMockAzureProvider(hclog.NewNullLogger())
	if err != nil {
		t.Fatalf("Failed to create Azure provider: %v", err)
	}
	
	// Test starting the instance
	err = provider.StartInstance(context.Background())
	if err != nil {
		t.Fatalf("Failed to start instance: %v", err)
	}
}