# Snoozebot Plugin Architecture

This document provides an overview of Snoozebot's plugin architecture, detailing the technical design, communication protocols, and implementation patterns.

## Architecture Overview

Snoozebot uses a plugin architecture to support multiple cloud providers and allow for extensibility. The architecture is built on the following key components:

1. **Plugin Manager**: Manages the lifecycle of plugins (discovery, loading, unloading)
2. **Plugin Interface**: Defines the contract between the core application and plugins
3. **gRPC Protocol**: Handles communication between the core and plugins
4. **HashiCorp go-plugin**: Provides the foundation for process isolation and plugin management

```
┌──────────────────────────────────────┐      ┌───────────────────────────────────┐
│                                      │      │                                   │
│             Snoozebot Agent          │      │           Plugin Process          │
│                                      │      │                                   │
│  ┌───────────────┐   ┌────────────┐  │      │  ┌────────────┐   ┌────────────┐  │
│  │               │   │            │  │      │  │            │   │            │  │
│  │ Plugin Manager├───┤ gRPC Client│◄─┼──────┼──►gRPC Server │───┤  Plugin    │  │
│  │               │   │            │  │      │  │            │   │Implementation│  │
│  └───────────────┘   └────────────┘  │      │  └────────────┘   └────────────┘  │
│                                      │      │                                   │
└──────────────────────────────────────┘      └───────────────────────────────────┘
```

## Process Isolation

Each plugin runs as a separate process, providing several benefits:

1. **Stability**: If a plugin crashes, it won't bring down the entire application
2. **Security**: Plugins have limited access to the core application
3. **Versioning**: Different plugins can use different versions of dependencies
4. **Resource Management**: Each plugin's resources are independently managed

## Plugin Interface

The primary interface that cloud provider plugins implement is `CloudProvider`:

```go
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

## Plugin Manager

The PluginManager is responsible for:

1. **Discovery**: Finding available plugins in the plugins directory
2. **Loading**: Starting plugin processes and establishing communication
3. **Management**: Tracking loaded plugins and providing access to them
4. **Unloading**: Shutting down plugin processes gracefully

```go
type PluginManager interface {
    // LoadPlugin loads a cloud provider plugin
    LoadPlugin(ctx context.Context, pluginName string) (CloudProvider, error)
    
    // UnloadPlugin unloads a cloud provider plugin
    UnloadPlugin(pluginName string) error
    
    // GetPlugin gets a loaded cloud provider plugin
    GetPlugin(pluginName string) (CloudProvider, error)
    
    // ListPlugins lists all loaded plugins
    ListPlugins() []string
    
    // DiscoverPlugins discovers plugins in the plugins directory
    DiscoverPlugins() ([]string, error)
}
```

## gRPC Protocol

Communication between the core application and plugins uses gRPC:

1. **Protocol Buffers**: Define the service and message types for plugin communication
2. **Bidirectional Streaming**: Allows for efficient, typed communication
3. **Connection Management**: Handles reconnection and error recovery

The Protocol Buffers definition for the CloudProvider service:

```protobuf
service CloudProvider {
  rpc GetInstanceInfo(GetInstanceInfoRequest) returns (GetInstanceInfoResponse);
  rpc StopInstance(StopInstanceRequest) returns (StopInstanceResponse);
  rpc StartInstance(StartInstanceRequest) returns (StartInstanceResponse);
  rpc GetProviderName(GetProviderNameRequest) returns (GetProviderNameResponse);
  rpc GetProviderVersion(GetProviderVersionRequest) returns (GetProviderVersionResponse);
}
```

## Plugin Lifecycle

Plugins go through a specific lifecycle:

### Discovery

1. The PluginManager scans the plugins directory for executable files
2. Each executable is checked for proper permissions
3. A list of available plugins is compiled

### Loading

1. The PluginManager starts the plugin process
2. A gRPC connection is established
3. A handshake verifies compatibility
4. The plugin is registered with the PluginManager

### Usage

1. The core application gets the plugin from the PluginManager
2. Method calls are made over gRPC
3. Results are returned to the caller

### Unloading

1. The PluginManager signals the plugin to shut down
2. The plugin cleans up resources
3. The plugin process exits
4. The PluginManager removes the plugin from its registry

## Error Handling

The plugin system includes robust error handling:

1. **Connection Errors**: Detected and reported with detailed diagnostics
2. **Timeouts**: All operations have configurable timeouts
3. **Retries**: Operations can be retried with backoff
4. **Health Checks**: Plugin health is periodically verified

## Plugin Development

To create a new plugin, developers need to:

1. Create a new Go module in the plugins directory
2. Implement the CloudProvider interface
3. Set up gRPC server handling with HashiCorp go-plugin
4. Build the plugin executable

See [PLUGIN_DEVELOPMENT.md](./PLUGIN_DEVELOPMENT.md) for a detailed guide.

## Testing

The plugin system includes comprehensive testing:

1. **Unit Tests**: Verify the functionality of individual components
2. **Integration Tests**: Test the end-to-end functionality of the plugin system
3. **Mock Plugins**: Allow testing without actual cloud providers

## Security Considerations

The plugin architecture addresses several security considerations:

1. **Process Isolation**: Limits the impact of plugin vulnerabilities
2. **Validated Interface**: Ensures plugins only have access to defined functionality
3. **Resource Limits**: Prevents plugins from consuming excessive resources
4. **Authentication**: Verifies plugin identity through handshake mechanisms

## Future Extensions

The plugin architecture is designed to be extensible:

1. **Additional Interfaces**: Support for new plugin types beyond cloud providers
2. **Versioned APIs**: Allow for API evolution while maintaining backward compatibility
3. **Remote Plugins**: Enable plugins to run on different machines
4. **Plugin Marketplace**: Facilitate discovery and sharing of community-developed plugins