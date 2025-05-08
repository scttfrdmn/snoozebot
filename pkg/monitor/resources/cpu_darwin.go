// +build darwin

package resources

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CPUMonitor monitors CPU usage on macOS systems
type CPUMonitor struct {
	lastMeasurement float64
	lastUpdateTime  time.Time
	mutex           sync.Mutex
}

// NewCPUMonitor creates a new CPU monitor for macOS
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

	// On macOS, we'll refresh measurements every second at most
	// to avoid excessive system calls
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

// getCPUUsage gets CPU usage percentage from top command on macOS
func getCPUUsage() (float64, error) {
	// Run top command with one-shot output
	cmd := exec.Command("top", "-l", "1", "-n", "0", "-s", "0")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute top command: %w", err)
	}

	// Parse output to find CPU usage
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU usage") {
			// Format: "CPU usage: X.XX% user, X.XX% sys, X.XX% idle"
			parts := strings.Split(line, ",")
			if len(parts) < 3 {
				return 0, fmt.Errorf("unexpected top output format: %s", line)
			}

			// Parse the idle percentage
			idlePart := parts[2]
			idleStr := strings.TrimSpace(strings.Split(idlePart, "%")[0])
			idle, err := strconv.ParseFloat(idleStr, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse idle percentage: %w", err)
			}

			// CPU usage is 100 - idle percentage
			return 100.0 - idle, nil
		}
	}

	return 0, fmt.Errorf("could not find CPU usage information in top output")
}