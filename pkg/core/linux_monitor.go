package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LinuxResourceMonitor implements ResourceMonitor for Linux systems
type LinuxResourceMonitor struct {
	config        *MonitorConfig
	thresholds    map[ResourceType]float64
	currentUsage  map[ResourceType]*ResourceUsage
	ctx           context.Context
	cancel        context.CancelFunc
	running       bool
	mutex         sync.RWMutex
	wg            sync.WaitGroup
	lastActivity  time.Time
	idleDuration  time.Duration
}

// NewLinuxResourceMonitor creates a new LinuxResourceMonitor
func NewLinuxResourceMonitor(config *MonitorConfig) *LinuxResourceMonitor {
	if config == nil {
		config = DefaultMonitorConfig()
	}

	return &LinuxResourceMonitor{
		config:       config,
		thresholds:   config.Thresholds,
		currentUsage: make(map[ResourceType]*ResourceUsage),
		lastActivity: time.Now(),
	}
}

// Start starts the resource monitor
func (m *LinuxResourceMonitor) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("monitor already running")
	}

	// Create a new context with cancel function
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true

	// Start the monitoring goroutine
	m.wg.Add(1)
	go m.monitorResources()

	return nil
}

// Stop stops the resource monitor
func (m *LinuxResourceMonitor) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return fmt.Errorf("monitor not running")
	}

	m.cancel()
	m.wg.Wait()
	m.running = false

	return nil
}

// GetUsage gets the current resource usage
func (m *LinuxResourceMonitor) GetUsage(resourceType ResourceType) (*ResourceUsage, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	usage, ok := m.currentUsage[resourceType]
	if !ok {
		return nil, fmt.Errorf("resource type not monitored: %s", resourceType)
	}

	return usage, nil
}

// GetAllUsage gets all current resource usage
func (m *LinuxResourceMonitor) GetAllUsage() (map[ResourceType]*ResourceUsage, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a copy of the current usage map
	usageCopy := make(map[ResourceType]*ResourceUsage)
	for k, v := range m.currentUsage {
		usageCopy[k] = v
	}

	return usageCopy, nil
}

// SetThreshold sets the threshold for a resource
func (m *LinuxResourceMonitor) SetThreshold(resourceType ResourceType, threshold float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.thresholds[resourceType] = threshold
	return nil
}

// GetThreshold gets the threshold for a resource
func (m *LinuxResourceMonitor) GetThreshold(resourceType ResourceType) (float64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	threshold, ok := m.thresholds[resourceType]
	if !ok {
		return 0, fmt.Errorf("no threshold set for resource type: %s", resourceType)
	}

	return threshold, nil
}

// monitorResources is the main monitoring loop
func (m *LinuxResourceMonitor) monitorResources() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateResourceUsage()
			m.checkIdleState()
		}
	}
}

// updateResourceUsage updates the current resource usage
func (m *LinuxResourceMonitor) updateResourceUsage() {
	// This is a placeholder for actual resource monitoring implementation
	// In a real implementation, we would use platform-specific APIs to get resource usage
	
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	
	// Simulate CPU usage
	m.currentUsage[CPU] = &ResourceUsage{
		Type:      CPU,
		Value:     5.0, // Placeholder value
		Timestamp: now,
	}
	
	// Simulate Memory usage
	m.currentUsage[Memory] = &ResourceUsage{
		Type:      Memory,
		Value:     15.0, // Placeholder value
		Timestamp: now,
	}
	
	// Simulate Network usage
	m.currentUsage[Network] = &ResourceUsage{
		Type:      Network,
		Value:     2.0, // Placeholder value
		Timestamp: now,
	}
	
	// Simulate Disk usage
	m.currentUsage[Disk] = &ResourceUsage{
		Type:      Disk,
		Value:     3.0, // Placeholder value
		Timestamp: now,
	}
	
	// Simulate User input (0 means no input)
	m.currentUsage[UserInput] = &ResourceUsage{
		Type:      UserInput,
		Value:     0.0, // Placeholder value
		Timestamp: now,
	}
	
	// Simulate GPU usage
	m.currentUsage[GPU] = &ResourceUsage{
		Type:      GPU,
		Value:     0.0, // Placeholder value
		Timestamp: now,
	}
}

// checkIdleState checks if the system is idle
func (m *LinuxResourceMonitor) checkIdleState() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check if any resource is above its threshold
	isIdle := true
	for resourceType, usage := range m.currentUsage {
		threshold, ok := m.thresholds[resourceType]
		if !ok {
			continue
		}

		if usage.Value > threshold {
			isIdle = false
			break
		}
	}

	// Update idle state
	now := time.Now()
	if isIdle {
		if m.idleDuration == 0 {
			// System just became idle
			m.lastActivity = now
		}
		m.idleDuration = now.Sub(m.lastActivity)
	} else {
		// System is active
		m.lastActivity = now
		m.idleDuration = 0
	}

	// Check if system has been idle for the naptime duration
	if m.idleDuration >= m.config.NapTime {
		// System has been idle for the naptime duration
		// In a real implementation, we would trigger the cloud provider plugin here
		fmt.Printf("System has been idle for %s, ready to be stopped\n", m.idleDuration)
	}
}