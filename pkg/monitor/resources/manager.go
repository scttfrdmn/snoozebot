package resources

import (
	"fmt"
	"sync"
	"time"
)

// MonitorManager manages all resource monitors
type MonitorManager struct {
	monitors map[ResourceType]ResourceMonitor
	custom   map[string]CustomMonitorFunc
	mutex    sync.RWMutex
}

// CustomMonitorFunc is a function that returns a resource usage value
type CustomMonitorFunc func() (float64, error)

// NewMonitorManager creates a new monitor manager
func NewMonitorManager() (*MonitorManager, error) {
	manager := &MonitorManager{
		monitors: make(map[ResourceType]ResourceMonitor),
		custom:   make(map[string]CustomMonitorFunc),
	}

	// Initialize default monitors
	defaultTypes := []ResourceType{CPU, Memory, Network, Disk, UserInput, GPU}
	for _, resourceType := range defaultTypes {
		monitor, err := NewResourceMonitor(resourceType)
		if err != nil {
			return nil, fmt.Errorf("failed to create monitor for %s: %w", resourceType, err)
		}
		manager.monitors[resourceType] = monitor
	}

	return manager, nil
}

// AddCustomMonitor adds a custom resource monitor
func (m *MonitorManager) AddCustomMonitor(name string, fn CustomMonitorFunc) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.custom[name] = fn
}

// RemoveCustomMonitor removes a custom resource monitor
func (m *MonitorManager) RemoveCustomMonitor(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.custom, name)
}

// GetUsage gets the current usage for a resource
func (m *MonitorManager) GetUsage(resourceType ResourceType) (*ResourceUsage, error) {
	m.mutex.RLock()
	monitor, ok := m.monitors[resourceType]
	m.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no monitor found for resource type: %s", resourceType)
	}

	value, err := monitor.GetUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get usage for %s: %w", resourceType, err)
	}

	return &ResourceUsage{
		Type:      resourceType,
		Value:     value,
		Timestamp: time.Now(),
	}, nil
}

// GetCustomUsage gets the current usage for a custom resource
func (m *MonitorManager) GetCustomUsage(name string) (*ResourceUsage, error) {
	m.mutex.RLock()
	monitor, ok := m.custom[name]
	m.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no custom monitor found with name: %s", name)
	}

	value, err := monitor()
	if err != nil {
		return nil, fmt.Errorf("failed to get usage for custom monitor %s: %w", name, err)
	}

	return &ResourceUsage{
		Type:      ResourceType(name),
		Value:     value,
		Timestamp: time.Now(),
	}, nil
}

// GetAllUsage gets the current usage for all resources
func (m *MonitorManager) GetAllUsage() (map[ResourceType]*ResourceUsage, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	usage := make(map[ResourceType]*ResourceUsage)

	// Get usage for all standard monitors
	for resourceType, monitor := range m.monitors {
		value, err := monitor.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("failed to get usage for %s: %w", resourceType, err)
		}

		usage[resourceType] = &ResourceUsage{
			Type:      resourceType,
			Value:     value,
			Timestamp: time.Now(),
		}
	}

	// Get usage for all custom monitors
	for name, monitor := range m.custom {
		value, err := monitor()
		if err != nil {
			return nil, fmt.Errorf("failed to get usage for custom monitor %s: %w", name, err)
		}

		customType := ResourceType(name)
		usage[customType] = &ResourceUsage{
			Type:      customType,
			Value:     value,
			Timestamp: time.Now(),
		}
	}

	return usage, nil
}