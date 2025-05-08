# Snoozebot Plugin Development Guide

This guide explains how to develop custom plugins for Snoozebot to extend its functionality to support additional cloud providers or custom actions.

## Plugin Architecture Overview

Snoozebot uses HashiCorp's go-plugin library to implement a robust, process-isolated plugin system. Each plugin:

- Runs as a separate process from the main application
- Communicates with the main application using gRPC
- Implements standard interfaces defined by the core
- Can be managed (started/stopped) independently

## Plugin Interfaces

### CloudProvider Interface

The primary interface that cloud provider plugins must implement is `CloudProvider`:

```go
type CloudProvider interface {
    // GetInstanceInfo gets information about the current instance
    GetInstanceInfo(ctx context.Context) (*InstanceInfo, error)
    
    // StopInstance stops the current instance
    StopInstance(ctx context.Context) error
    
    // StartInstance starts the current instance
    StartInstance(ctx context.Context) error
    
    // GetProviderName returns the name of the cloud provider
    GetProviderName() string
    
    // GetProviderVersion returns the version of the cloud provider plugin
    GetProviderVersion() string
}
```

## Creating a New Plugin

### Step 1: Set Up the Plugin Directory

Create a new directory for your plugin in the `plugins/` directory:

```bash
mkdir -p plugins/myprovider
```

### Step 2: Create the Go Module

Initialize a Go module for your plugin:

```bash
cd plugins/myprovider
go mod init github.com/scottfridman/snoozebot/plugins/myprovider
```

Add a replace directive to point to the local snoozebot module:

```go
// go.mod
module github.com/scottfridman/snoozebot/plugins/myprovider

go 1.18

require (
    github.com/hashicorp/go-hclog v0.14.1
    github.com/hashicorp/go-plugin v1.6.3
    github.com/scottfridman/snoozebot v0.0.0
)

replace github.com/scottfridman/snoozebot => ../../
```

### Step 3: Implement the CloudProvider Interface

Create a `main.go` file that implements the `CloudProvider` interface:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/hashicorp/go-hclog"
    "github.com/hashicorp/go-plugin"
    snoozePlugin "github.com/scottfridman/snoozebot/pkg/plugin"
)

// MyProvider is an implementation of the CloudProvider interface
type MyProvider struct {
    logger hclog.Logger
}

// GetInstanceInfo gets information about the current instance
func (p *MyProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.InstanceInfo, error) {
    p.logger.Info("Getting instance info")
    
    // Implementation specific to your cloud provider
    return &snoozePlugin.InstanceInfo{
        ID:         "my-instance-id",
        Name:       "my-instance",
        Type:       "my-instance-type",
        Region:     "my-region",
        Zone:       "my-zone",
        State:      "running",
        LaunchTime: time.Now().Add(-24 * time.Hour),
    }, nil
}

// StopInstance stops the current instance
func (p *MyProvider) StopInstance(ctx context.Context) error {
    p.logger.Info("Stopping instance")
    
    // Implementation specific to your cloud provider
    return nil
}

// StartInstance starts the current instance
func (p *MyProvider) StartInstance(ctx context.Context) error {
    p.logger.Info("Starting instance")
    
    // Implementation specific to your cloud provider
    return nil
}

// GetProviderName returns the name of the cloud provider
func (p *MyProvider) GetProviderName() string {
    return "myprovider"
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *MyProvider) GetProviderVersion() string {
    return "1.0.0"
}

func main() {
    // Create a logger
    logger := hclog.New(&hclog.LoggerOptions{
        Level:      hclog.Trace,
        Output:     os.Stderr,
        JSONFormat: true,
    })

    // Create the plugin
    myProvider := &MyProvider{
        logger: logger,
    }

    // Set up the plugin map
    plugins := map[string]plugin.Plugin{
        "cloud_provider": &snoozePlugin.CloudProviderPlugin{
            Impl: myProvider,
        },
    }

    // Start the plugin
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: snoozePlugin.Handshake,
        Plugins:         plugins,
        GRPCServer:      plugin.DefaultGRPCServer,
        Logger:          logger,
    })
}
```

### Step 4: Build the Plugin

Build your plugin:

```bash
go build -o myprovider
```

### Step 5: Install the Plugin

Copy the plugin to the Snoozebot plugins directory:

```bash
cp myprovider /etc/snoozebot/plugins/
```

## Plugin Lifecycle

### Loading

Plugins are loaded by the Snoozebot daemon when:
- The daemon starts
- A plugin is explicitly loaded through the API
- A plugin is needed for an operation

### Communication

Communication between the core and plugins follows this flow:
1. Core creates a new plugin client
2. Plugin process is started
3. Handshake is performed to verify compatibility
4. gRPC channel is established
5. Method calls are made over the gRPC channel
6. Results are returned to the core

### Unloading

Plugins are unloaded when:
- The daemon shuts down
- A plugin is explicitly unloaded through the API
- A plugin crashes or becomes unresponsive

## Best Practices

### Error Handling

- Handle errors gracefully and return informative error messages
- Log errors with appropriate context
- Don't crash the plugin process on recoverable errors

### Logging

- Use the provided logger rather than directly writing to stdout/stderr
- Include relevant context in log messages
- Use appropriate log levels (trace, debug, info, warn, error)

### Resource Management

- Clean up resources in case of errors or when the plugin is unloaded
- Don't leave orphaned processes or open connections
- Implement proper timeouts for operations

### Configuration

- Support configuration through environment variables
- Validate configuration at startup
- Fail fast if required configuration is missing

### Testing

- Write unit tests for your plugin
- Test with the actual cloud provider API
- Test error handling and edge cases

## Example Plugins

- [AWS Plugin](../plugins/aws/main.go): Example implementation for AWS
- [GCP Plugin](../plugins/gcp/main.go): Example implementation for GCP (coming soon)
- [Azure Plugin](../plugins/azure/main.go): Example implementation for Azure (coming soon)

## Troubleshooting

### Plugin Won't Load

- Check plugin executable permissions
- Verify plugin is built for the correct architecture
- Check logs for error messages

### Communication Errors

- Ensure plugin is compatible with the core version
- Check for network issues if the plugin runs remotely
- Verify gRPC is not being blocked by a firewall

### Plugin Crashes

- Check logs for panic messages
- Ensure all dependencies are available
- Verify plugin has necessary permissions

## Resources

- [HashiCorp go-plugin Documentation](https://github.com/hashicorp/go-plugin)
- [gRPC Documentation](https://grpc.io/docs/)