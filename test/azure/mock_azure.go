package azuretest

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	snoozePlugin "github.com/scttfrdmn/snoozebot/pkg/plugin"
)

// MockAzureProvider is a mock implementation of the Azure provider for testing
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