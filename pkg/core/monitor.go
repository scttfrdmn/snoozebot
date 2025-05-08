package core

import (
	"context"
	"time"
)

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
	// Value is the resource usage value
	Value float64
	// Timestamp is when the measurement was taken
	Timestamp time.Time
}

// ResourceMonitor is the interface for monitoring resource usage
type ResourceMonitor interface {
	// Start starts the resource monitor
	Start(ctx context.Context) error

	// Stop stops the resource monitor
	Stop() error

	// GetUsage gets the current resource usage
	GetUsage(resourceType ResourceType) (*ResourceUsage, error)

	// GetAllUsage gets all current resource usage
	GetAllUsage() (map[ResourceType]*ResourceUsage, error)

	// SetThreshold sets the threshold for a resource
	SetThreshold(resourceType ResourceType, threshold float64) error

	// GetThreshold gets the threshold for a resource
	GetThreshold(resourceType ResourceType) (float64, error)
}

// MonitorConfig represents the configuration for a resource monitor
type MonitorConfig struct {
	// Thresholds is a map of resource types to thresholds
	Thresholds map[ResourceType]float64

	// NapTime is the duration that resource usage must be below thresholds before taking action
	NapTime time.Duration

	// CheckInterval is how often to check resource usage
	CheckInterval time.Duration
}

// DefaultMonitorConfig returns a default monitor configuration
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		Thresholds: map[ResourceType]float64{
			CPU:       10.0,
			Memory:    20.0,
			Network:   5.0,
			Disk:      5.0,
			UserInput: 0.0,
			GPU:       5.0,
		},
		NapTime:       30 * time.Minute,
		CheckInterval: 1 * time.Minute,
	}
}