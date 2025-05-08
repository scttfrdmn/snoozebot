# Snoozebot Development Context

This document contains essential context for Claude to continue development of the Snoozebot project across sessions.

## Project Overview

Snoozebot is a system for automatically managing cloud resources by monitoring system activity and hibernating idle instances to save costs. It follows a two-part architecture:

1. **Embedded Monitoring Library**: A lightweight Go module that can be embedded in a host application to monitor system resources and detect idle states.
2. **Remote Agent**: A standalone service that manages cloud resources across multiple instances, handling cloud provider interactions through plugins.

## Current Status

- **Architecture**: Defined and initial implementation in progress
- **Phase 1 (Core Monitoring)**: 
  - Basic interfaces implemented ✅
  - Resource monitoring: CPU monitoring implemented for Linux, macOS, Windows ✅
  - Resource monitoring: Memory monitoring implemented for Linux, macOS, Windows ✅
  - Resource monitoring: Network monitoring implemented for Linux ✅
  - Resource monitoring: Disk I/O monitoring implemented for Linux ✅
  - Resource monitoring: User input monitoring implemented for Linux ✅
  - Resource monitoring: GPU monitoring implemented for Linux ✅
  - Monitoring manager implemented ✅
  - Testing infrastructure and unit tests in place ✅
- **Phase 2 (Agent Communication)**: 
  - Protocol messages defined ✅
  - gRPC service interface defined ✅
  - Client implementation created ✅
  - Server implementation pending ⏳
- **Phase 3 (Remote Agent)**: Basic structure created, full implementation pending
- **Phase 4 (Documentation)**: Architecture and API documentation in place, more needed

See [IMPLEMENTATION_PLAN.md](/Users/scttfrdmn/src/snoozebot/IMPLEMENTATION_PLAN.md) for detailed progress and next steps.

## Key Files and Their Purposes

### Core Monitoring Library
- `/pkg/monitor/api.go`: Public API for the monitor library
- `/pkg/monitor/monitor.go`: Implementation of the monitor
- `/pkg/monitor/resources/resources.go`: Resource monitoring interfaces
- `/pkg/monitor/resources/manager.go`: Resource monitoring manager
- `/pkg/monitor/resources/cpu_linux.go`: Linux CPU monitoring
- `/pkg/monitor/resources/cpu_darwin.go`: macOS CPU monitoring
- `/pkg/monitor/resources/cpu_windows.go`: Windows CPU monitoring
- `/pkg/monitor/resources/memory_linux.go`: Linux memory monitoring
- `/pkg/monitor/resources/memory_darwin.go`: macOS memory monitoring
- `/pkg/monitor/resources/memory_windows.go`: Windows memory monitoring
- `/pkg/monitor/resources/network_linux.go`: Linux network monitoring
- `/pkg/monitor/resources/disk_linux.go`: Linux disk I/O monitoring
- `/pkg/monitor/resources/user_input_linux.go`: Linux user input monitoring
- `/pkg/monitor/resources/gpu_linux.go`: Linux GPU monitoring

### Communication Protocol
- `/pkg/common/protocol/protocol.go`: Protocol message definitions
- `/pkg/common/protocol/client.go`: Agent client implementation
- `/pkg/common/protocol/proto/agent.proto`: gRPC service and message definitions
- `/pkg/common/protocol/gen/`: Generated gRPC code

### Remote Agent
- `/agent/cmd/main.go`: Agent entry point
- `/agent/api/server.go`: API server for the agent
- `/agent/store/store.go`: State storage for the agent
- `/agent/provider/provider.go`: Cloud provider plugin interfaces

### Examples
- `/examples/embedded/main.go`: Example of embedding the monitor in an application

## Implementation Priorities

1. **Resource Monitoring**: Implement actual resource monitoring for CPU, memory, etc.
2. **Testing Infrastructure**: Set up comprehensive testing infrastructure
3. **Agent Communication**: Implement gRPC communication between monitor and agent
4. **Plugin System**: Complete plugin loading and management in the agent
5. **Cloud Provider Plugins**: Implement and test with real cloud providers

## Cloud Testing Configuration

### AWS Testing
- Use temporary credentials through environment variables
- Test region: us-west-2
- Test instance types: t3.micro (minimal cost)
- Clean up resources immediately after tests

### GCP Testing
- Use service account JSON key file
- Test region: us-central1
- Test instance types: e2-micro (minimal cost)
- Clean up resources immediately after tests

### Azure Testing
- Use service principal authentication
- Test region: eastus
- Test VM sizes: Standard_B1s (minimal cost)
- Clean up resources immediately after tests

## Command Patterns

### Building the Project
```bash
go build -o bin/embedded ./examples/embedded
go build -o bin/snooze-agent ./agent/cmd
go build -o bin/plugins/aws ./plugins/aws
```

### Running Tests
```bash
go test -v ./pkg/monitor/...
go test -v ./pkg/common/...
go test -v ./agent/...
go test -v ./plugins/...
```

### Integration Tests
```bash
./test/integration.sh
```

### Cloud Tests (require cloud credentials)
```bash
AWS_ACCESS_KEY_ID=xxx AWS_SECRET_ACCESS_KEY=xxx go test -v ./test/cloud/aws/...
```

## Notes for Next Steps

1. Begin by implementing actual resource monitoring, starting with CPU for Linux
2. Set up the testing infrastructure before proceeding further
3. For each platform, implement monitoring in this order: CPU, memory, disk, network, user input, GPU
4. Maintain cross-platform compatibility throughout
5. Test thoroughly with different load patterns

## Reminders

- Always clean up cloud resources after testing
- Write tests alongside implementation, not after
- Keep the API stable once defined
- Document all significant design decisions
- Ensure error messages are clear and actionable

## Contact Information

Project lead: Scott Fridman