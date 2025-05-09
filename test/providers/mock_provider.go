package providers

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin"
)

// MockProvider is a mock implementation of the CloudProvider interface for testing
type MockProvider struct {
	name          string
	version       string
	logger        hclog.Logger
	instances     []*plugin.CloudInstanceInfo
	instanceState map[string]string
}

// NewMockProvider creates a new mock cloud provider
func NewMockProvider(name string, logger hclog.Logger) *MockProvider {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	// Create some mock instances
	instances := []*plugin.CloudInstanceInfo{
		MockInstance("i-123456789", "test-instance-1", "t2.micro", "us-west-2", "us-west-2a", "running"),
		MockInstance("i-987654321", "test-instance-2", "t2.small", "us-west-2", "us-west-2b", "stopped"),
		MockInstance("i-555555555", "test-instance-3", "t2.medium", "us-west-2", "us-west-2c", "running"),
	}

	// Initialize instance state map
	instanceState := make(map[string]string)
	for _, inst := range instances {
		instanceState[inst.ID] = inst.State
	}

	return &MockProvider{
		name:          name,
		version:       "0.1.0",
		logger:        logger,
		instances:     instances,
		instanceState: instanceState,
	}
}

// GetAPIVersion returns the API version implemented by the plugin
func (p *MockProvider) GetAPIVersion() string {
	return plugin.CurrentAPIVersion
}

// GetInstanceInfo gets information about the current instance
func (p *MockProvider) GetInstanceInfo(ctx context.Context) (*plugin.CloudInstanceInfo, error) {
	p.logger.Info("Getting instance info (mock)")
	
	// Default to the first instance or use instance ID from environment
	instanceID := "i-123456789" // Default to first mock instance
	
	// Try to find the instance
	for _, inst := range p.instances {
		if inst.ID == instanceID {
			// Update state from instance state map
			inst.State = p.instanceState[inst.ID]
			return inst, nil
		}
	}
	
	return nil, fmt.Errorf("instance not found: %s", instanceID)
}

// StopInstance stops the current instance
func (p *MockProvider) StopInstance(ctx context.Context) error {
	p.logger.Info("Stopping instance (mock)")
	
	// Random delay to simulate API call
	simulateNetworkDelay()
	
	// Default to the first instance
	instanceID := "i-123456789"
	
	// Update the instance state
	p.instanceState[instanceID] = "stopped"
	return nil
}

// StartInstance starts the current instance
func (p *MockProvider) StartInstance(ctx context.Context) error {
	p.logger.Info("Starting instance (mock)")
	
	// Random delay to simulate API call
	simulateNetworkDelay()
	
	// Default to the first instance
	instanceID := "i-123456789"
	
	// Update the instance state
	p.instanceState[instanceID] = "running"
	return nil
}

// GetProviderName returns the name of the cloud provider
func (p *MockProvider) GetProviderName() string {
	return p.name
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *MockProvider) GetProviderVersion() string {
	return p.version
}

// ListInstances lists all instances
func (p *MockProvider) ListInstances(ctx context.Context) ([]*plugin.CloudInstanceInfo, error) {
	p.logger.Info("Listing instances (mock)")
	
	// Random delay to simulate API call
	simulateNetworkDelay()
	
	// Update instance states
	for _, inst := range p.instances {
		inst.State = p.instanceState[inst.ID]
	}
	
	return p.instances, nil
}

// Shutdown performs cleanup when the plugin is being unloaded
func (p *MockProvider) Shutdown() {
	p.logger.Info("Shutting down mock provider")
}

// Simulate network delay for realistic mock behavior
func simulateNetworkDelay() {
	// Random delay between 100-300ms to simulate network latency
	delay := 100 + rand.Intn(200)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}