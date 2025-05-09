// +build !linux

package resources

import (
	"sync"
)

// GPUMonitor is a stub implementation for non-Linux platforms
type GPUMonitor struct {
	mutex sync.Mutex
}

// NewGPUMonitor creates a new GPU monitor
// This is a stub implementation for non-Linux platforms
func NewGPUMonitor() (*GPUMonitor, error) {
	return &GPUMonitor{}, nil
}

// GetUsage returns a dummy GPU usage value
func (m *GPUMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Return a dummy value
	return 5.0, nil
}