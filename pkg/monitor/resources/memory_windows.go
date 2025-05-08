// +build windows

package resources

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MemoryMonitor monitors memory usage on Windows systems
type MemoryMonitor struct {
	lastMeasurement float64
	lastUpdateTime  time.Time
	mutex           sync.Mutex
}

// NewMemoryMonitor creates a new memory monitor for Windows
func NewMemoryMonitor() (*MemoryMonitor, error) {
	// Initialize with current stats
	usage, err := getMemoryUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize memory monitor: %w", err)
	}

	return &MemoryMonitor{
		lastMeasurement: usage,
		lastUpdateTime:  time.Now(),
	}, nil
}

// GetUsage returns the current memory usage as a percentage (0-100)
func (m *MemoryMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// On Windows, we'll refresh measurements every second at most
	// to avoid excessive PowerShell executions
	now := time.Now()
	if now.Sub(m.lastUpdateTime) < time.Second {
		return m.lastMeasurement, nil
	}

	usage, err := getMemoryUsage()
	if err != nil {
		return m.lastMeasurement, err
	}

	m.lastMeasurement = usage
	m.lastUpdateTime = now

	return usage, nil
}

// getMemoryUsage gets memory usage percentage from PowerShell on Windows
func getMemoryUsage() (float64, error) {
	// Run PowerShell command to get memory usage
	// This gets the percentage of physical memory in use
	cmd := exec.Command("powershell", "-Command", 
		"$os = Get-CimInstance Win32_OperatingSystem; "+
		"$pctUsed = (($os.TotalVisibleMemorySize - $os.FreePhysicalMemory) * 100) / $os.TotalVisibleMemorySize; "+
		"$pctUsed")
	
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute PowerShell command: %w", err)
	}

	// Parse output to find memory usage
	usageStr := strings.TrimSpace(string(output))
	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse memory usage: %w", err)
	}

	return usage, nil
}