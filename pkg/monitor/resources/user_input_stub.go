// +build !linux

package resources

import (
	"sync"
)

// UserInputMonitor is a stub implementation for non-Linux platforms
type UserInputMonitor struct {
	mutex sync.Mutex
}

// NewUserInputMonitor creates a new user input monitor
// This is a stub implementation for non-Linux platforms
func NewUserInputMonitor() (*UserInputMonitor, error) {
	return &UserInputMonitor{}, nil
}

// GetUsage returns a dummy user input value
func (m *UserInputMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Return a dummy value (0 means no user activity)
	return 0.0, nil
}