# Snoozebot

Snoozebot is a Go-based system for automatically managing cloud resources to save costs by monitoring system activity and hibernating idle instances.

## Architecture

Snoozebot consists of two main components:

1. **Embedded Monitoring Library**: A lightweight Go module that can be embedded in a host application to monitor system resources and detect idle states.

2. **Remote Agent**: A standalone service that manages cloud resources across multiple instances, making decisions about when to hibernate instances and handling the cloud provider interactions.

This architecture allows for:
- Full system shutdown (not just hibernation)
- Centralized management of multiple instances
- Separation of monitoring logic from cloud provider interaction
- Easy integration into host applications

## Features

- **Lightweight Monitoring**: Tracks CPU, memory, network, disk, user input, and GPU activity
- **Configurable Thresholds**: Customizable resource usage thresholds for idle detection
- **Extensible**: Support for custom resource monitoring metrics
- **Cloud-Provider Agnostic**: Plugin system for different cloud providers
- **Process Isolation**: Cloud provider plugins run as separate processes for stability and security
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Embeddable**: Can be used as a library in other Go applications
- **Secure Plugin System**: Plugins use TLS encryption, signature verification, and API key authentication
- **Role-Based Access Control**: Fine-grained permission control for plugin operations
- **Versioned Plugin API**: Semantic versioning with compatibility checking for stable plugin interfaces
- **Notification System**: Flexible notification framework with Slack integration

## Embedding in Host Applications

Snoozebot's monitoring library can be embedded in a host application:

```go
// Create and configure the monitor
monitor := monitor.NewMonitor().
    WithThreshold(monitor.CPU, 15.0).
    WithThreshold(monitor.Memory, 25.0).
    WithNapTime(30 * time.Minute).
    WithAgentURL("http://snooze-agent.example.com:8080")

// Add custom resource monitors
monitor.AddResourceMonitor("custom_app_metric", func() (float64, error) {
    return getApplicationMetric(), nil
})

// Handle idle state changes
monitor.OnIdleStateChange(func(isIdle bool, duration time.Duration) {
    if isIdle {
        log.Printf("System idle for %s", duration)
    } else {
        log.Printf("System active")
    }
})

// Start the monitor
monitor.Start(ctx)
```

See the [examples directory](./examples) for more details.

## Cloud Provider Support

Snoozebot supports the following cloud providers through plugins:

- AWS (Amazon Web Services)
- GCP (Google Cloud Platform) - Planned
- Azure (Microsoft Azure Cloud)

## Agent API

The Snoozebot agent provides a REST API for instance registration and management:

- `/api/instances/register` - Register an instance with the agent
- `/api/instances/unregister` - Unregister an instance
- `/api/instances/idle` - Send idle notifications
- `/api/instances/heartbeat` - Send heartbeats
- `/api/instances/state` - Send state changes
- `/api/instances` - List instances

## Building from Source

Prerequisites:
- Go 1.18 or newer

```bash
git clone https://github.com/scottfridman/snoozebot.git
cd snoozebot

# Build the embedded example
go build -o bin/embedded ./examples/embedded

# Build the agent
go build -o bin/snooze-agent ./agent/cmd

# Build plugins
go build -o bin/plugins/aws ./plugins/aws
go build -o bin/plugins/azure ./plugins/azure
```

## Documentation

- [Refactoring Plan](./REFACTORING_PLAN.md) - Details on the architecture and implementation plan
- [Embedding Guide](./examples/embedded/README.md) - Guide to embedding the monitor in applications
- [Plugin Development](./docs/PLUGIN_DEVELOPMENT.md) - Guide to developing cloud provider plugins
- [Plugin TLS](./docs/PLUGIN_TLS.md) - Guide to configuring TLS encryption for secure plugin communication
- [Plugin Signatures](./docs/PLUGIN_SIGNATURES.md) - Guide to signature verification for plugin authenticity
- [Plugin Authentication](./docs/PLUGIN_AUTHENTICATION.md) - Guide to API key authentication for plugins
- [API Versioning](./docs/API_VERSIONING.md) - Guide to the plugin API versioning system
- [Security Maintenance](./docs/SECURITY_MAINTENANCE.md) - Guide to maintaining security with dependency management
- [Notification System](./docs/NOTIFICATION_SYSTEM.md) - Guide to the notification framework architecture
- [Slack Notifications](./docs/SLACK_NOTIFICATIONS.md) - Guide to setting up and using Slack notifications
- [Email Notifications](./docs/EMAIL_NOTIFICATIONS.md) - Guide to setting up and using email notifications

## License

Apache License 2.0