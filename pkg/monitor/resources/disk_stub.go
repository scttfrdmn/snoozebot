// +build !linux

package resources

import (
	"sync"
)

// DiskMonitor is a stub implementation for non-Linux platforms
type DiskMonitor struct {
	mutex sync.Mutex
}

// NewDiskMonitor creates a new disk monitor
// This is a stub implementation for non-Linux platforms
func NewDiskMonitor() (*DiskMonitor, error) {
	return &DiskMonitor{}, nil
}

// GetUsage returns a dummy disk usage value
func (m *DiskMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Return a dummy value
	return 15.0, nil
}