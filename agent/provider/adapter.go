package provider

import (
	"context"
	"fmt"

	pluginlib "github.com/scttfrdmn/snoozebot/pkg/plugin"
)

// PluginAdapter adapts pkg/plugin.CloudProvider to agent/provider.CloudProvider
type PluginAdapter struct {
	plugin    pluginlib.CloudProvider
	defaultID string
}

// NewPluginAdapter creates a new adapter for a plugin
func NewPluginAdapter(plugin pluginlib.CloudProvider, defaultID string) CloudProvider {
	if defaultID == "" {
		defaultID = "default-id"
	}
	
	return &PluginAdapter{
		plugin:    plugin,
		defaultID: defaultID,
	}
}

// GetInstanceInfo gets information about an instance
func (p *PluginAdapter) GetInstanceInfo(ctx context.Context, instanceID string) (*InstanceInfo, error) {
	// Call the underlying plugin without an ID
	info, err := p.plugin.GetInstanceInfo(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convert CloudInstanceInfo to InstanceInfo
	return &InstanceInfo{
		ID:         info.ID,
		Name:       info.Name,
		Type:       info.Type,
		Region:     info.Region,
		Zone:       info.Zone,
		State:      info.State,
		LaunchTime: info.LaunchTime,
		Provider:   p.plugin.GetProviderName(),
	}, nil
}

// StopInstance stops an instance
func (p *PluginAdapter) StopInstance(ctx context.Context, instanceID string) error {
	// Call the underlying plugin without an ID
	return p.plugin.StopInstance(ctx)
}

// StartInstance starts an instance
func (p *PluginAdapter) StartInstance(ctx context.Context, instanceID string) error {
	// Call the underlying plugin without an ID
	return p.plugin.StartInstance(ctx)
}

// GetProviderName returns the name of the cloud provider
func (p *PluginAdapter) GetProviderName() string {
	return p.plugin.GetProviderName()
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *PluginAdapter) GetProviderVersion() string {
	return p.plugin.GetProviderVersion()
}

// ListInstances lists all instances
func (p *PluginAdapter) ListInstances(ctx context.Context) ([]*InstanceInfo, error) {
	// This method doesn't exist in the CloudProvider interface,
	// but we can implement it using the ListInstances method of the plugin
	cloudInstances, err := p.plugin.ListInstances(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convert CloudInstanceInfo to InstanceInfo
	instances := make([]*InstanceInfo, len(cloudInstances))
	for i, cloudInstance := range cloudInstances {
		instances[i] = &InstanceInfo{
			ID:         cloudInstance.ID,
			Name:       cloudInstance.Name,
			Type:       cloudInstance.Type,
			Region:     cloudInstance.Region,
			Zone:       cloudInstance.Zone,
			State:      cloudInstance.State,
			LaunchTime: cloudInstance.LaunchTime,
			Provider:   p.plugin.GetProviderName(),
		}
	}
	
	return instances, nil
}