# Examples Update for Snoozebot 0.1.0

This document describes the updates made to the example code to align with the current API for Snoozebot 0.1.0.

## Overview of Changes

The following changes were made to the example code:

1. **Updated Custom Plugin Example**: Replaced `InstanceInfo` with `CloudInstanceInfo` for API compatibility
2. **Added Shutdown Method**: Implemented the mandatory Shutdown method added in v0.1.0
3. **Added Documentation**: Added comments and notes about v0.1.0 requirements

## Detailed Changes

### 1. Updated Custom Plugin Example with CloudInstanceInfo

The custom plugin example (`examples/custom_plugin/main.go`) was updated to use `CloudInstanceInfo` instead of `InstanceInfo` to align with the renamed type in the core codebase:

**Before:**
```go
func (p *CustomProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.InstanceInfo, error) {
    // ...
    return &snoozePlugin.InstanceInfo{
        ID:         "i-custom123",
        // ...
    }, nil
}

func (p *CustomProvider) ListInstances(ctx context.Context) ([]*snoozePlugin.InstanceInfo, error) {
    // ...
    instances := []*snoozePlugin.InstanceInfo{
        // ...
    }
    // ...
}
```

**After:**
```go
func (p *CustomProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.CloudInstanceInfo, error) {
    // ...
    return &snoozePlugin.CloudInstanceInfo{
        ID:         "i-custom123",
        // ...
    }, nil
}

func (p *CustomProvider) ListInstances(ctx context.Context) ([]*snoozePlugin.CloudInstanceInfo, error) {
    // ...
    instances := []*snoozePlugin.CloudInstanceInfo{
        // ...
    }
    // ...
}
```

### 2. Added Shutdown Method

Added the mandatory Shutdown method that was introduced in v0.1.0 to properly release resources:

```go
// Shutdown is called when the plugin is being unloaded
func (p *CustomProvider) Shutdown() {
    p.logger.Info("Shutting down custom provider")
    // Perform any cleanup operations here, such as:
    // - Closing connections
    // - Releasing resources
    // - Stopping background workers
}
```

### 3. Added Documentation

Added a note in the custom plugin example to remind users about implementing the Shutdown method:

```go
// Add capabilities
provider.AddCapability("list_instances")
provider.AddCapability("start_instance")
provider.AddCapability("stop_instance")

// Make sure to implement the Shutdown method as well
// This is a new requirement in v0.1.0
```

## Impact

These changes ensure that the example code:

1. Aligns with the current API
2. Includes all required methods
3. Uses the correct type names
4. Provides guidance for proper implementation

Developers using these examples as a basis for their own plugins will now see the correct patterns to follow for Snoozebot 0.1.0.