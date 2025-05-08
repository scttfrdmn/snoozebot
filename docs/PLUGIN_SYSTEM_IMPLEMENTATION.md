# Snoozebot Plugin System Implementation

This document summarizes the implementation of the Snoozebot plugin management system, which allows for dynamic loading, unloading, and discovery of cloud provider plugins.

## Overview

The plugin system uses HashiCorp's go-plugin library to implement a robust, process-isolated plugin architecture. Each plugin:

- Runs as a separate process from the main application
- Communicates with the main application using gRPC
- Implements standard interfaces defined by the core
- Can be managed (loaded/unloaded) independently

## Implemented Components

### Core Plugin System

1. **Plugin Manager** (`agent/provider/manager.go`)
   - Manages the lifecycle of plugins (loading, unloading, discovery)
   - Provides a unified interface for accessing plugins
   - Handles errors and retries

2. **Plugin Interface** (`agent/provider/provider.go`)
   - Defines the `CloudProvider` interface that all plugins must implement
   - Standardizes operations like `GetInstanceInfo`, `StartInstance`, `StopInstance`

3. **Plugin gRPC Implementation** (`pkg/plugin/`)
   - Handles serialization and deserialization of data for cross-process communication
   - Implements the client and server sides of the gRPC protocol

### REST API

1. **Plugin API Handlers** (`agent/api/plugin_handlers.go`)
   - HTTP endpoints for plugin management operations
   - Handles plugin listing, discovery, loading, and unloading

2. **Cloud Provider Operations** (`agent/api/cloud_provider_operations.go`)
   - gRPC endpoints for cloud-specific operations
   - Routes operations to the appropriate plugin
   - Handles errors and retries

3. **Plugin Discovery** (`agent/api/plugins.go`)
   - Automatic discovery and initialization of plugins during agent startup

### Plugin Implementations

1. **AWS Plugin** (`plugins/aws/main.go`)
   - AWS EC2 implementation of the CloudProvider interface
   - Handles AWS-specific instance operations

2. **GCP Plugin** (`plugins/gcp/main.go`)
   - Google Cloud Platform implementation of the CloudProvider interface
   - Handles GCP-specific instance operations

## Test Coverage

Comprehensive tests have been implemented for:

1. **Plugin Manager** (`agent/provider/manager_test.go`)
   - Tests for loading, unloading, and discovering plugins
   - Tests for listing loaded plugins

2. **Plugin API Handlers** (`agent/api/plugin_handlers_test.go`)
   - Tests for HTTP endpoints handling plugin operations
   - Tests error handling and edge cases

3. **Cloud Provider Operations** (`agent/api/cloud_provider_operations_test.go`)
   - Tests for gRPC endpoints handling cloud operations
   - Tests instance state management

4. **Plugin Discovery** (`agent/api/plugins_test.go`)
   - Tests for automatic plugin discovery and initialization
   - Tests error handling during initialization

## Key Features

- **Dynamic Loading/Unloading**: Plugins can be loaded and unloaded at runtime
- **Auto-Discovery**: Plugins are automatically discovered in the plugins directory
- **Process Isolation**: Each plugin runs in its own process for stability and security
- **Standard Interface**: All plugins implement a common interface
- **Robust Error Handling**: Comprehensive error handling and retries
- **Versioning**: Plugin version compatibility checks

## Usage

### Loading a Plugin

```http
POST /api/plugins/load
Content-Type: application/json

{
  "plugin": "aws"
}
```

### Unloading a Plugin

```http
POST /api/plugins/unload
Content-Type: application/json

{
  "plugin": "aws"
}
```

### Listing Loaded Plugins

```http
GET /api/plugins
```

### Discovering Available Plugins

```http
GET /api/plugins/discover
```

## Next Steps

1. Implement authentication and authorization for plugin management API
2. Add metrics and telemetry for plugin operations
3. Implement Azure plugin
4. Create a web UI for plugin management
5. Add plugin dependency management