# Snoozebot Plugin Development Guide

This guide explains how to develop custom plugins for Snoozebot to extend its functionality to support additional cloud providers or custom actions.

> **Note**: For implementing new cloud providers, follow the guidelines in [CLOUD_PROVIDER_IMPLEMENTATION.md](./CLOUD_PROVIDER_IMPLEMENTATION.md), which includes details about the branching strategy and code review requirements.

## Plugin Architecture Overview

Snoozebot uses HashiCorp's go-plugin library to implement a robust, process-isolated plugin system. Each plugin:

- Runs as a separate process from the main application
- Communicates with the main application using gRPC
- Implements standard interfaces defined by the core
- Can be managed (started/stopped) independently

## Plugin System Architecture

Snoozebot's plugin architecture is designed for maximum flexibility and security:

```
┌───────────────────┐      ┌───────────────────┐
│                   │      │                   │
│  Snoozebot Agent  │      │  Plugin Process   │
│                   │      │                   │
│  ┌─────────────┐  │      │  ┌─────────────┐  │
│  │             │  │      │  │             │  │
│  │ PluginMgr   ◄──┼──────┼──► Plugin      │  │
│  │             │  │      │  │ Implementation│  │
│  └─────────────┘  │      │  └─────────────┘  │
│                   │      │                   │
└───────────────────┘      └───────────────────┘
        │                          ▲
        │                          │
        │      ┌──────────────┐    │
        └──────►              ◄────┘
               │    gRPC      │
               │              │
               └──────────────┘
```

## Plugin Interfaces

### CloudProvider Interface

The primary interface that cloud provider plugins must implement is `CloudProvider`:

```go
// CloudProvider is the interface that cloud provider plugins must implement
type CloudProvider interface {
    // GetInstanceInfo gets information about an instance
    GetInstanceInfo(ctx context.Context, instanceID string) (*InstanceInfo, error)
    
    // StopInstance stops an instance
    StopInstance(ctx context.Context, instanceID string) error
    
    // StartInstance starts an instance
    StartInstance(ctx context.Context, instanceID string) error
    
    // GetProviderName returns the name of the cloud provider
    GetProviderName() string
    
    // GetProviderVersion returns the version of the cloud provider plugin
    GetProviderVersion() string
}
```

### InstanceInfo Structure

The `InstanceInfo` structure holds information about a cloud instance:

```go
// InstanceInfo contains information about a cloud instance
type InstanceInfo struct {
    // ID is the unique identifier for the instance
    ID string
    
    // Name is the name of the instance
    Name string
    
    // Type is the instance type
    Type string
    
    // Region is the region where the instance is located
    Region string
    
    // Zone is the availability zone where the instance is located
    Zone string
    
    // State is the current state of the instance
    State string
    
    // LaunchTime is when the instance was launched
    LaunchTime time.Time
    
    // Provider is the cloud provider (aws, gcp, azure)
    Provider string
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

You can build your plugin using the provided build script:

```bash
# Build a specific plugin
./scripts/build_plugins.sh myprovider

# Or using the Makefile
make plugin PLUGIN=myprovider
```

Alternatively, build it manually:

```bash
# Navigate to the project root
cd /path/to/snoozebot

# Build the plugin
go build -o bin/plugins/myprovider ./plugins/myprovider
```

### Step 5: Install the Plugin

The build steps above will place the plugin in the `bin/plugins/` directory, which is the default location for plugins in development mode.

For production deployment, copy the plugin to the system plugins directory:

```bash
sudo mkdir -p /etc/snoozebot/plugins
sudo cp bin/plugins/myprovider /etc/snoozebot/plugins/
sudo chmod +x /etc/snoozebot/plugins/myprovider
```

### Step 6: Test the Plugin

You can test your plugin using the integration test infrastructure:

```bash
# Run the integration tests with your plugin
SNOOZEBOT_RUN_INTEGRATION=true go test -v ./test/integration/...
```

Or manually test it with the agent:

```bash
# Start the agent with your plugin
bin/snooze-agent --plugins-dir=bin/plugins
```

Then use the API to interact with your plugin:

```bash
# List available plugins
curl http://localhost:8080/api/plugins

# Load your plugin
curl -X POST -H "Content-Type: application/json" \
  -d '{"plugin_name":"myprovider"}' \
  http://localhost:8080/api/plugins/load
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

## API Endpoints for Plugin Management

The Snoozebot agent provides RESTful API endpoints for managing plugins:

### List Loaded Plugins

```
GET /api/plugins
```

**Response:**
```json
{
  "plugins": [
    {
      "name": "aws",
      "provider": "aws",
      "version": "0.1.0",
      "status": "active"
    },
    {
      "name": "gcp",
      "provider": "gcp",
      "version": "0.1.0",
      "status": "active"
    },
    {
      "name": "azure",
      "provider": "azure",
      "version": "0.1.0",
      "status": "active"
    }
  ],
  "count": 3
}
```

### Discover Available Plugins

```
GET /api/plugins/discover
```

**Response:**
```json
{
  "plugins": [
    {
      "name": "aws",
      "loaded": true,
      "provider": "aws",
      "version": "0.1.0"
    },
    {
      "name": "gcp",
      "loaded": false
    },
    {
      "name": "azure",
      "loaded": false
    }
  ],
  "count": 3,
  "directory": "/etc/snoozebot/plugins"
}
```

### Load a Plugin

```
POST /api/plugins/load
```

**Request:**
```json
{
  "plugin_name": "aws",
  "timeout_seconds": 30,
  "retries": 3
}
```

**Response:**
```json
{
  "success": true,
  "plugin_name": "aws",
  "provider_name": "aws",
  "provider_version": "0.1.0",
  "attempts": 1
}
```

### Unload a Plugin

```
POST /api/plugins/unload
```

**Request:**
```json
{
  "plugin_name": "aws",
  "force": false
}
```

**Response:**
```json
{
  "success": true,
  "plugin_name": "aws",
  "provider_name": "aws",
  "provider_version": "0.1.0",
  "status": "unloaded",
  "duration_ms": 15
}
```

## Best Practices

### Error Handling

- Handle errors gracefully and return informative error messages
- Log errors with appropriate context
- Don't crash the plugin process on recoverable errors
- Use structured errors with error codes where appropriate
- Consider implementing auto-recovery for transient errors

Example of good error handling:

```go
func (p *MyProvider) GetInstanceInfo(ctx context.Context, instanceID string) (*provider.InstanceInfo, error) {
    p.logger.Info("Getting instance info", "instanceID", instanceID)
    
    // Check for invalid input
    if instanceID == "" {
        return nil, fmt.Errorf("instanceID cannot be empty")
    }
    
    // Call the cloud provider API with timeout and retry
    var instance *cloudprovider.Instance
    var err error
    
    for i := 0; i < 3; i++ {
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()
        
        instance, err = p.client.GetInstance(ctx, instanceID)
        if err == nil {
            break
        }
        
        // Check for permanent error vs. transient error
        if isPermanentError(err) {
            p.logger.Error("Permanent error getting instance", 
                "instanceID", instanceID, 
                "error", err)
            return nil, fmt.Errorf("failed to get instance info: %w", err)
        }
        
        p.logger.Warn("Transient error getting instance, retrying", 
            "instanceID", instanceID, 
            "attempt", i+1, 
            "error", err)
        time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
    }
    
    if err != nil {
        p.logger.Error("Failed to get instance after retries", 
            "instanceID", instanceID, 
            "error", err)
        return nil, fmt.Errorf("failed to get instance info after retries: %w", err)
    }
    
    // Map the cloud provider's instance object to our InstanceInfo
    return mapInstanceToInfo(instance), nil
}
```

### Logging

- Use the provided logger rather than directly writing to stdout/stderr
- Include relevant context in log messages (e.g., instance IDs, operation names)
- Use appropriate log levels (trace, debug, info, warn, error)
- Log the start and end of major operations
- Include timing information for performance tracking

Example of good logging:

```go
func (p *MyProvider) StopInstance(ctx context.Context, instanceID string) error {
    logger := p.logger.With("instanceID", instanceID, "operation", "StopInstance")
    logger.Info("Starting instance stop operation")
    
    startTime := time.Now()
    
    // Make the API call to stop the instance
    err := p.client.StopInstance(ctx, instanceID)
    
    duration := time.Since(startTime)
    logger = logger.With("duration_ms", duration.Milliseconds())
    
    if err != nil {
        logger.Error("Failed to stop instance", "error", err)
        return fmt.Errorf("failed to stop instance: %w", err)
    }
    
    logger.Info("Successfully stopped instance")
    return nil
}
```

### Resource Management

- Clean up resources in case of errors or when the plugin is unloaded
- Don't leave orphaned processes or open connections
- Implement proper timeouts for operations
- Use context cancellation for cleanup
- Implement graceful shutdown
- Close connections properly in `Shutdown` hooks

Example of good resource management:

```go
func (p *MyProvider) Shutdown() {
    p.logger.Info("Shutting down provider")
    
    // Cancel ongoing operations
    if p.cancel != nil {
        p.cancel()
    }
    
    // Close client connections
    if p.client != nil {
        if err := p.client.Close(); err != nil {
            p.logger.Error("Error closing client", "error", err)
        }
    }
    
    // Release other resources
    for _, resource := range p.resources {
        if err := resource.Close(); err != nil {
            p.logger.Error("Error closing resource", "resource", resource.ID, "error", err)
        }
    }
    
    p.logger.Info("Provider shutdown complete")
}
```

### Configuration

- Support configuration through environment variables
- Validate configuration at startup
- Fail fast if required configuration is missing
- Use structured configuration with validation
- Support reloading configuration if possible
- Document configuration options

Example of good configuration handling:

```go
type ProviderConfig struct {
    Region          string
    CredentialsFile string
    Timeout         time.Duration
    MaxRetries      int
}

func LoadConfigFromEnvironment() (*ProviderConfig, error) {
    config := &ProviderConfig{
        Region:          os.Getenv("CLOUD_REGION"),
        CredentialsFile: os.Getenv("CREDENTIALS_FILE"),
        Timeout:         30 * time.Second,
        MaxRetries:      3,
    }
    
    // Override defaults with environment variables if provided
    if timeoutStr := os.Getenv("OPERATION_TIMEOUT"); timeoutStr != "" {
        timeoutSec, err := strconv.Atoi(timeoutStr)
        if err != nil {
            return nil, fmt.Errorf("invalid OPERATION_TIMEOUT value: %w", err)
        }
        config.Timeout = time.Duration(timeoutSec) * time.Second
    }
    
    if retriesStr := os.Getenv("MAX_RETRIES"); retriesStr != "" {
        retries, err := strconv.Atoi(retriesStr)
        if err != nil {
            return nil, fmt.Errorf("invalid MAX_RETRIES value: %w", err)
        }
        config.MaxRetries = retries
    }
    
    // Validate required fields
    if config.Region == "" {
        return nil, fmt.Errorf("CLOUD_REGION environment variable is required")
    }
    
    if config.CredentialsFile == "" {
        return nil, fmt.Errorf("CREDENTIALS_FILE environment variable is required")
    }
    
    // Validate that credentials file exists
    if _, err := os.Stat(config.CredentialsFile); os.IsNotExist(err) {
        return nil, fmt.Errorf("credentials file does not exist: %s", config.CredentialsFile)
    }
    
    return config, nil
}
```

### Testing

- Write unit tests for your plugin
- Test with the actual cloud provider API
- Test error handling and edge cases
- Use mocks for external dependencies
- Test timeouts and cancellation
- Test resource cleanup
- Write integration tests
- Add benchmark tests for performance-critical paths

Example of a good test:

```go
func TestMyProvider_GetInstanceInfo(t *testing.T) {
    // Set up a mock cloud client
    mockClient := NewMockCloudClient()
    
    // Create the provider with the mock client
    provider := &MyProvider{
        logger: hclog.NewNullLogger(),
        client: mockClient,
    }
    
    // Set up the mock to return a specific instance
    mockInstance := &cloudprovider.Instance{
        ID:         "test-instance",
        Name:       "test",
        Type:       "small",
        Region:     "us-west-1",
        Zone:       "us-west-1a",
        State:      "running",
        LaunchTime: time.Now().Add(-1 * time.Hour),
    }
    mockClient.On("GetInstance", mock.Anything, "test-instance").Return(mockInstance, nil)
    
    // Call the method
    info, err := provider.GetInstanceInfo(context.Background(), "test-instance")
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, info)
    assert.Equal(t, "test-instance", info.ID)
    assert.Equal(t, "test", info.Name)
    assert.Equal(t, "small", info.Type)
    assert.Equal(t, "us-west-1", info.Region)
    assert.Equal(t, "us-west-1a", info.Zone)
    assert.Equal(t, "running", info.State)
    
    // Test error case
    mockClient.On("GetInstance", mock.Anything, "nonexistent").Return(nil, fmt.Errorf("instance not found"))
    
    // Call the method
    info, err = provider.GetInstanceInfo(context.Background(), "nonexistent")
    
    // Assertions
    assert.Error(t, err)
    assert.Nil(t, info)
    assert.Contains(t, err.Error(), "failed to get instance info")
}
```

## Development Tools

Snoozebot provides several development tools to help you build and test plugins:

### Build Script

Use the build script to compile your plugins:

```bash
./scripts/build_plugins.sh myprovider
```

### Makefile Targets

The Makefile includes several targets for plugin development:

```bash
# Build all plugins
make plugins

# Build a specific plugin
make plugin PLUGIN=myprovider

# Run all tests (including plugin tests)
make test

# Run just the integration tests
make test-integration
```

### Integration Tests

Integration tests for plugins are located in the `test/integration` directory. Use these tests as examples for testing your own plugins.

## Example Plugins

- [AWS Plugin](../plugins/aws/main.go): Example implementation for AWS
- [GCP Plugin](../plugins/gcp/main.go): Example implementation for GCP
- [Azure Plugin](../plugins/azure/main.go): Example implementation for Azure
- [Integration Tests](../test/integration/plugin_integration_test.go): Examples of plugin testing

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