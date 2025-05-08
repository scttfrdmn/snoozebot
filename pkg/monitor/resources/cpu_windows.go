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

// CPUMonitor monitors CPU usage on Windows systems
type CPUMonitor struct {
	lastMeasurement float64
	lastUpdateTime  time.Time
	mutex           sync.Mutex
}

// NewCPUMonitor creates a new CPU monitor for Windows
func NewCPUMonitor() (*CPUMonitor, error) {
	// Initialize with current stats
	usage, err := getCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize CPU monitor: %w", err)
	}

	return &CPUMonitor{
		lastMeasurement: usage,
		lastUpdateTime:  time.Now(),
	}, nil
}

// GetUsage returns the current CPU usage as a percentage (0-100)
func (m *CPUMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// On Windows, we'll refresh measurements every second at most
	// to avoid excessive PowerShell executions
	now := time.Now()
	if now.Sub(m.lastUpdateTime) < time.Second {
		return m.lastMeasurement, nil
	}

	usage, err := getCPUUsage()
	if err != nil {
		return m.lastMeasurement, err
	}

	m.lastMeasurement = usage
	m.lastUpdateTime = now

	return usage, nil
}

// getCPUUsage gets CPU usage percentage from PowerShell on Windows
func getCPUUsage() (float64, error) {
	// Run PowerShell command to get CPU usage
	cmd := exec.Command("powershell", "-Command", "Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average | Select-Object -ExpandProperty Average")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute PowerShell command: %w", err)
	}

	// Parse output to find CPU usage
	usageStr := strings.TrimSpace(string(output))
	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse CPU usage: %w", err)
	}

	return usage, nil
}