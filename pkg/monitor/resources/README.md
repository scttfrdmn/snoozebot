# Resource Monitoring Package

This package provides platform-specific resource monitoring capabilities for the Snoozebot project.

## Overview

The resource monitoring package includes:

- A common interface for all resource monitors
- Platform-specific implementations for different resources
- A resource manager to coordinate monitoring activities

## Usage

```go
// Create a resource manager
manager, err := resources.NewMonitorManager()
if err != nil {
    log.Fatalf("Failed to create resource manager: %v", err)
}

// Add a custom resource monitor (optional)
manager.AddCustomMonitor("my_custom_metric", func() (float64, error) {
    // Measure something and return a value between 0 and 100
    return 50.0, nil
})

// Get usage for a specific resource
cpuUsage, err := manager.GetUsage(resources.CPU)
if err != nil {
    log.Fatalf("Failed to get CPU usage: %v", err)
}
fmt.Printf("CPU Usage: %.2f%%\n", cpuUsage.Value)

// Get usage for all resources
allUsage, err := manager.GetAllUsage()
if err != nil {
    log.Fatalf("Failed to get all resource usage: %v", err)
}

for resourceType, usage := range allUsage {
    fmt.Printf("%s Usage: %.2f%%\n", resourceType, usage.Value)
}
```

## Resource Types

The following standard resource types are supported:

- `CPU`: CPU utilization
- `Memory`: Memory utilization
- `Network`: Network I/O
- `Disk`: Disk I/O
- `UserInput`: User keyboard and mouse activity
- `GPU`: GPU utilization

## Platform Support

| Resource  | Linux | macOS | Windows |
|-----------|-------|-------|---------|
| CPU       | ✅    | ✅    | ✅      |
| Memory    | ✅    | ✅    | ✅      |
| Network   | ✅    | ⏳    | ⏳      |
| Disk      | ✅    | ⏳    | ⏳      |
| UserInput | ✅    | ⏳    | ⏳      |
| GPU       | ✅    | ⏳    | ⏳      |

Legend:
- ✅: Implemented
- ⏳: Planned
- ❌: Not planned

## Implementation Details

### CPU Monitoring

- **Linux**: Reads and parses `/proc/stat` to get CPU statistics
- **macOS**: Uses `top` command to get CPU usage
- **Windows**: Uses PowerShell to get CPU usage via WMI

### Memory Monitoring

- **Linux**: Reads and parses `/proc/meminfo` to get memory statistics
- **macOS**: Uses `vm_stat` command to get memory statistics and `sysctl` to get total memory
- **Windows**: Uses PowerShell and WMI to get memory statistics

### Network Monitoring

- **Linux**: Reads and parses `/proc/net/dev` to track network I/O bytes
- **macOS**: Not yet implemented
- **Windows**: Not yet implemented

### Disk I/O Monitoring

- **Linux**: Reads and parses `/proc/diskstats` to track disk I/O activity
- **macOS**: Not yet implemented
- **Windows**: Not yet implemented

### User Input Monitoring

- **Linux**: Uses `xprintidle` if available or falls back to manual detection
- **macOS**: Not yet implemented
- **Windows**: Not yet implemented

### GPU Monitoring

- **Linux**: Uses `nvidia-smi` to get GPU utilization for NVIDIA GPUs
- **macOS**: Not yet implemented
- **Windows**: Not yet implemented

## Testing

Each platform-specific implementation includes unit tests. These tests can be run with:

```bash
go test -v ./pkg/monitor/resources/...
```

To skip platform-specific tests when testing on a different platform, use:

```bash
SKIP_LINUX_TESTS=1 go test -v ./pkg/monitor/resources/...
SKIP_DARWIN_TESTS=1 go test -v ./pkg/monitor/resources/...
SKIP_WINDOWS_TESTS=1 go test -v ./pkg/monitor/resources/...
```