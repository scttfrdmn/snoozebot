package provider

import (
	"context"
	"time"
)

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

// CloudProvider defines the interface for cloud provider plugins
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
	
	// ListInstances lists all instances
	ListInstances(ctx context.Context) ([]*InstanceInfo, error)
}

// PluginManager manages cloud provider plugins
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