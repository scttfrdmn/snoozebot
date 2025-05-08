# Snoozebot Refactoring Plan

## Overview

This document outlines the plan to refactor Snoozebot from its current architecture to an "Embedded Module + Remote Agent" architecture.

The new architecture will consist of:

1. **Embedded Monitoring Library**: A lightweight Go module that can be embedded in a host application
2. **Remote Agent**: A standalone service that manages cloud resources across multiple instances

This approach provides several advantages:
- Host applications can directly integrate monitoring capabilities
- Instances can be fully stopped (not just hibernated)
- Centralized management of multiple instances
- Separation of monitoring logic from cloud provider interaction

## Project Structure Changes

Current structure:
```
snoozebot/
├── cmd/
│   ├── snoozed/      # Daemon
│   └── snooze/       # CLI
├── pkg/
│   ├── core/         # Core monitoring and plugin management
│   └── plugin/       # Plugin interfaces
└── plugins/          # Cloud provider plugins
```

New structure:
```
snoozebot/
├── pkg/
│   ├── monitor/      # Embeddable monitoring library
│   │   ├── api.go    # Public API for host applications
│   │   ├── resources/# Resource monitoring implementations
│   │   └── config.go # Configuration types
│   ├── common/       # Shared types/utilities
│   │   ├── protocol/ # Communication protocol definitions
│   │   └── models/   # Shared data models
│   └── plugin/       # Plugin system (moves to agent)
├── agent/            # Remote agent implementation
│   ├── cmd/          # Agent command-line application
│   ├── api/          # REST/gRPC API implementation
│   ├── provider/     # Cloud provider interface
│   └── store/        # State storage
├── plugins/          # Cloud provider plugins
│   ├── aws/
│   ├── gcp/
│   └── azure/
└── examples/         # Integration examples
    ├── standalone/   # Standalone usage example
    └── embedded/     # Example host application integration
```

## Core Monitoring Library API

The embedded library will use a fluent API design for easy configuration and extension:

```go
// Public API
type Monitor interface {
    // Configuration
    WithThreshold(resourceType string, threshold float64) Monitor
    WithNapTime(duration time.Duration) Monitor
    WithCheckInterval(duration time.Duration) Monitor
    WithAgentURL(url string) Monitor
    
    // Custom monitoring
    AddResourceMonitor(name string, fn ResourceMonitorFunc) Monitor
    
    // Event handlers
    OnIdleStateChange(fn IdleStateChangeHandler) Monitor
    OnError(fn ErrorHandler) Monitor
    
    // Lifecycle
    Start(ctx context.Context) error
    Stop() error
    
    // State
    GetCurrentState() MonitorState
    IsIdle() bool
    IdleDuration() time.Duration
}

// Easy constructor pattern
func NewMonitor() Monitor {
    // Create monitor instance with defaults
}

func NewMonitorWithConfig(config Config) Monitor {
    // Create monitor with specific config
}
```

## Communication Protocol

The monitoring library and remote agent will communicate using a lightweight protocol:

```go
// Protocol messages
type InstanceRegistration struct {
    InstanceID   string
    InstanceType string
    Region       string
    Provider     string
    Metadata     map[string]string
}

type IdleNotification struct {
    InstanceID    string
    IdleSince     time.Time
    IdleDuration  time.Duration
    ResourceUsage map[string]float64
}

type InstanceCommand struct {
    Command     string // "stop", "start", etc.
    ScheduledAt time.Time
}
```

Implementation will use gRPC with bidirectional streaming for real-time communication.

## Remote Agent Design

The remote agent will:
- Provide APIs for registration and management of instances
- Maintain state of all monitored instances
- Handle authentication and authorization
- Load and manage cloud provider plugins
- Execute cloud operations through plugins
- Expose management interfaces (API, CLI, web UI)

## Migration Strategy

### Phase 1: Split the Codebase
- Separate monitoring code from cloud provider code
- Implement the core monitoring library
- Create initial agent skeleton

### Phase 2: Implement Communication Protocol
- Define protocol messages and API
- Implement client in monitor library
- Implement server in agent

### Phase 3: Move Plugin System to Agent
- Adapt existing plugin system for agent
- Update existing cloud provider plugins

### Phase 4: Complete Agent Implementation
- Add state management
- Implement scheduling
- Create management API

### Phase 5: Create Documentation & Examples
- Document library API
- Create integration examples
- Update developer guides

## Testing Strategy

1. **Unit Tests**: Test core monitoring logic and agent components
2. **Integration Tests**: Test communication protocol
3. **End-to-End Tests**: Test complete workflows with example applications
4. **Mocks**: Create mock implementations for testing
   - Mock agent for library testing
   - Mock cloud providers for agent testing

## Implementation Plan

### Immediate Next Steps:

1. Create the base directory structure for the refactored project
2. Define the core interfaces for the monitoring library
3. Implement the basic resource monitoring functionality
4. Define the communication protocol between monitor and agent
5. Create a simple agent implementation to test the protocol

### Medium-Term Tasks:

1. Fully implement the remote agent with plugin support
2. Migrate existing cloud provider plugins to the new architecture
3. Implement state management in the agent
4. Add authentication and security features

### Long-Term Goals:

1. Create a comprehensive example application
2. Develop a web UI for the agent
3. Add support for additional cloud providers
4. Implement advanced features like scheduling and policies