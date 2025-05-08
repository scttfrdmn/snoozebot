package monitor

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
	// Value is the resource usage value as a percentage (0-100)
	Value float64
	// Timestamp is when the measurement was taken
	Timestamp time.Time
}

// MonitorState contains the current state of the monitor
type MonitorState struct {
	// IsIdle indicates if the system is currently considered idle
	IsIdle bool
	// IdleSince is the time when the system became idle
	IdleSince time.Time
	// IdleDuration is how long the system has been idle
	IdleDuration time.Duration
	// CurrentUsage is a map of resource types to their current usage
	CurrentUsage map[ResourceType]*ResourceUsage
	// Connected indicates if the monitor is connected to an agent
	Connected bool
}

// ResourceMonitorFunc is a function that monitors a resource and returns its usage
type ResourceMonitorFunc func() (float64, error)

// IdleStateChangeHandler is a function that is called when the idle state changes
type IdleStateChangeHandler func(isIdle bool, idleDuration time.Duration)

// ErrorHandler is a function that is called when an error occurs
type ErrorHandler func(err error)

// Monitor is the main interface for the monitoring library
type Monitor interface {
	// Configuration
	WithThreshold(resourceType ResourceType, threshold float64) Monitor
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

// Config contains configuration for the monitor
type Config struct {
	// Thresholds is a map of resource types to thresholds
	Thresholds map[ResourceType]float64
	// NapTime is the duration that resource usage must be below thresholds before taking action
	NapTime time.Duration
	// CheckInterval is how often to check resource usage
	CheckInterval time.Duration
	// AgentURL is the URL of the remote agent
	AgentURL string
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
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
		AgentURL:      "http://localhost:8080",
	}
}

// NewMonitor creates a new monitor with default configuration
func NewMonitor() Monitor {
	return NewMonitorWithConfig(DefaultConfig())
}

// NewMonitorWithConfig creates a new monitor with the specified configuration
func NewMonitorWithConfig(config Config) Monitor {
	return newMonitor(config)
}