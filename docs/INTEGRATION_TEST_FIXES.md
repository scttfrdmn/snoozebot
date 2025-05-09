# Integration Test Fixes for Snoozebot 0.1.0

This document outlines the approach for fixing integration tests in Snoozebot 0.1.0 to handle the different cloud provider interfaces.

## Issue

The integration tests are failing because there are two incompatible cloud provider interfaces:

1. **pkg/plugin.CloudProvider** - Uses `CloudInstanceInfo` struct and doesn't have instance ID parameters
   ```go
   type CloudProvider interface {
       // GetInstanceInfo gets information about the current instance
       GetInstanceInfo(ctx context.Context) (*CloudInstanceInfo, error)
       
       // StopInstance stops the current instance
       StopInstance(ctx context.Context) error
       
       // StartInstance starts the current instance
       StartInstance(ctx context.Context) error
       
       // ListInstances lists all instances
       ListInstances(ctx context.Context) ([]*CloudInstanceInfo, error)
       
       // Other methods...
   }
   ```

2. **agent/provider.CloudProvider** - Uses `InstanceInfo` struct and has instance ID parameters
   ```go
   type CloudProvider interface {
       // GetInstanceInfo gets information about an instance
       GetInstanceInfo(ctx context.Context, instanceID string) (*InstanceInfo, error)
       
       // StopInstance stops an instance
       StopInstance(ctx context.Context, instanceID string) error
       
       // StartInstance starts an instance
       StartInstance(ctx context.Context, instanceID string) error
       
       // Other methods...
   }
   ```

## Approach

There are two options for fixing this issue:

### Option 1: Create Adapter

Create an adapter in the agent/provider package that wraps the pkg/plugin.CloudProvider and adapts it to the agent/provider.CloudProvider interface. This approach maintains backward compatibility with existing code.

```go
// PluginAdapter adapts pkg/plugin.CloudProvider to agent/provider.CloudProvider
type PluginAdapter struct {
    plugin    pluginlib.CloudProvider
    defaultID string
}

// GetInstanceInfo implementation that passes the instance ID
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

// Similar implementations for other methods...
```

### Option 2: Update Plugin Interface

Update the pkg/plugin.CloudProvider interface to match the agent/provider.CloudProvider interface by adding instance ID parameters. This would be a breaking change that requires updating all plugins.

## Decision

We'll implement Option 1 (Create Adapter) because:

1. It maintains backward compatibility with existing plugins
2. It requires fewer changes to the codebase
3. It follows the adapter pattern, which is a clean design solution
4. It allows for a future migration path if we decide to unify the interfaces

## Implementation Plan

1. Create a PluginAdapter struct in agent/provider/adapter.go
2. Implement adapter methods for all CloudProvider interface methods
3. Modify the manager.go LoadPlugin method to wrap returned plugins in the adapter
4. Update integration tests to use the adapter

## Impact

This change will:

1. Allow integration tests to run successfully
2. Maintain backward compatibility with existing plugins
3. Provide a clean separation between the two interfaces
4. Enable a future migration path to a unified interface