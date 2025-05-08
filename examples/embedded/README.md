# Embedded Snoozebot Example

This example demonstrates how to embed the Snoozebot monitoring library in a host application.

## Overview

The example shows:

1. Creating and configuring a monitor
2. Adding custom resource monitors
3. Setting up event handlers
4. Managing the monitor lifecycle
5. Accessing monitor state

## Usage

```go
// Create a new monitor with default configuration
monitor := monitor.NewMonitor()

// Or create with custom configuration
monitor := monitor.NewMonitorWithConfig(monitor.Config{
    Thresholds: map[monitor.ResourceType]float64{
        monitor.CPU:       15.0,
        monitor.Memory:    20.0,
        monitor.Network:   5.0,
        monitor.Disk:      5.0,
        monitor.UserInput: 0.0,
        monitor.GPU:       5.0,
    },
    NapTime:       30 * time.Minute,
    CheckInterval: 1 * time.Minute,
    AgentURL:      "http://snooze-agent.example.com:8080",
})

// Configure the monitor using the fluent API
monitor.WithThreshold(monitor.CPU, 15.0).
    WithThreshold(monitor.Memory, 25.0).
    WithNapTime(5 * time.Minute).
    WithAgentURL("http://snooze-agent.example.com:8080")

// Add event handlers
monitor.OnIdleStateChange(func(isIdle bool, duration time.Duration) {
    if isIdle {
        fmt.Printf("System is now idle (duration: %s)\n", duration)
    } else {
        fmt.Println("System is now active")
    }
})

// Add custom resource monitor
monitor.AddResourceMonitor("custom_app_metric", func() (float64, error) {
    // Return a custom metric value between 0-100
    return getCustomMetricValue(), nil
})

// Start the monitor
monitor.Start(ctx)

// Access monitor state
state := monitor.GetCurrentState()
fmt.Printf("Is idle: %v, Duration: %v\n", state.IsIdle, state.IdleDuration)

// Stop the monitor when done
monitor.Stop()
```

## Integration Points

When integrating the Snoozebot monitor into a host application, consider these key integration points:

1. **Application Lifecycle**: Start the monitor when your application starts and stop it when your application stops.

2. **Custom Metrics**: Add custom resource monitors to track application-specific metrics that should influence the idle state.

3. **Idle State Handling**: Implement handlers for idle state changes to perform application-specific actions when the system becomes idle or active.

4. **Error Handling**: Set up error handlers to log or respond to monitoring errors.

5. **Configuration**: Allow users to configure the monitor thresholds, naptime, and other settings through your application's configuration system.

## Best Practices

1. **Resource Efficiency**: The monitor is designed to be lightweight, but be mindful of the check interval and custom resource monitors' performance impact.

2. **Error Handling**: Always handle errors from the monitor, especially when starting and stopping it.

3. **Configuration**: Provide sensible defaults for monitor configuration, but allow users to override them.

4. **Logging**: Integrate monitor events and errors with your application's logging system.

5. **Testing**: Test how your application behaves when the system goes idle and when it becomes active again.