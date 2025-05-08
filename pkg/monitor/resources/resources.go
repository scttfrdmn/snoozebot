package resources

import (
	"fmt"
	"runtime"
	"time"
)

// ResourceMonitor is the common interface for all resource monitors
type ResourceMonitor interface {
	// GetUsage returns the current resource usage as a percentage (0-100)
	GetUsage() (float64, error)
}

// ResourceType represents the type of resource being monitored
type ResourceType string

const (
	// CPU represents CPU utilization
	CPU ResourceType = "cpu"
	// Memory represents memory utilization
	Memory ResourceType = "memory"
	// Network represents network I/O
	Network ResourceType = "network"
	// Disk represents disk I/O
	Disk ResourceType = "disk"
	// UserInput represents user keyboard and mouse activity
	UserInput ResourceType = "user_input"
	// GPU represents GPU utilization
	GPU ResourceType = "gpu"
)

// ResourceUsage represents the usage of a resource
type ResourceUsage struct {
	// Type is the resource type
	Type ResourceType
	// Value is the resource usage value as a percentage (0-100)
	Value float64
	// Timestamp is when the measurement was taken
	Timestamp time.Time
}

// NewResourceMonitor creates a new resource monitor for the specified resource type
func NewResourceMonitor(resourceType ResourceType) (ResourceMonitor, error) {
	switch resourceType {
	case CPU:
		return newCPUMonitor()
	case Memory:
		return newMemoryMonitor()
	case Network:
		return newNetworkMonitor()
	case Disk:
		return newDiskMonitor()
	case UserInput:
		return newUserInputMonitor()
	case GPU:
		return newGPUMonitor()
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// newCPUMonitor creates a new CPU monitor based on the current platform
func newCPUMonitor() (ResourceMonitor, error) {
	return NewCPUMonitor()
}

// newMemoryMonitor creates a new memory monitor based on the current platform
func newMemoryMonitor() (ResourceMonitor, error) {
	return NewMemoryMonitor()
}

// newNetworkMonitor creates a new network monitor based on the current platform
func newNetworkMonitor() (ResourceMonitor, error) {
	// Currently only implemented for Linux
	switch runtime.GOOS {
	case "linux":
		return NewNetworkMonitor()
	default:
		return &dummyMonitor{resourceType: Network}, nil
	}
}

// newDiskMonitor creates a new disk monitor based on the current platform
func newDiskMonitor() (ResourceMonitor, error) {
	// Currently only implemented for Linux
	switch runtime.GOOS {
	case "linux":
		return NewDiskMonitor()
	default:
		return &dummyMonitor{resourceType: Disk}, nil
	}
}

// newUserInputMonitor creates a new user input monitor based on the current platform
func newUserInputMonitor() (ResourceMonitor, error) {
	// Currently only implemented for Linux
	switch runtime.GOOS {
	case "linux":
		// Try to create a user input monitor, but fall back to dummy if it fails
		monitor, err := NewUserInputMonitor()
		if err != nil {
			return &dummyMonitor{resourceType: UserInput}, nil
		}
		return monitor, nil
	default:
		return &dummyMonitor{resourceType: UserInput}, nil
	}
}

// newGPUMonitor creates a new GPU monitor based on the current platform
func newGPUMonitor() (ResourceMonitor, error) {
	// Currently only implemented for Linux
	switch runtime.GOOS {
	case "linux":
		// Try to create a GPU monitor, but fall back to dummy if it fails
		monitor, err := NewGPUMonitor()
		if err != nil {
			return &dummyMonitor{resourceType: GPU}, nil
		}
		return monitor, nil
	default:
		return &dummyMonitor{resourceType: GPU}, nil
	}
}

// dummyMonitor is a placeholder monitor that returns fixed values
type dummyMonitor struct {
	resourceType ResourceType
}

// GetUsage returns a fixed usage value
func (m *dummyMonitor) GetUsage() (float64, error) {
	// Return different values for different resource types
	// to simulate different usage patterns
	switch m.resourceType {
	case Memory:
		return 25.0, nil
	case Network:
		return 10.0, nil
	case Disk:
		return 15.0, nil
	case UserInput:
		return 0.0, nil
	case GPU:
		return 5.0, nil
	default:
		return 0.0, nil
	}
}