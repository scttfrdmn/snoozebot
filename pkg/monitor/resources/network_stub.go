// +build !linux

package resources

import (
	"sync"
)

// NetworkMonitor is a stub implementation for non-Linux platforms
type NetworkMonitor struct {
	mutex sync.Mutex
}

// NewNetworkMonitor creates a new network monitor
// This is a stub implementation for non-Linux platforms
func NewNetworkMonitor() (*NetworkMonitor, error) {
	return &NetworkMonitor{}, nil
}

// GetUsage returns a dummy network usage value
func (m *NetworkMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Return a dummy value
	return 5.0, nil
}