// +build linux

package resources

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GPUMonitor monitors GPU usage on Linux systems with NVIDIA GPUs
type GPUMonitor struct {
	lastMeasurement float64
	lastUpdateTime  time.Time
	mutex           sync.Mutex
	hasGPU          bool
}

// NewGPUMonitor creates a new GPU monitor for Linux
func NewGPUMonitor() (*GPUMonitor, error) {
	// Check if we have an NVIDIA GPU and nvidia-smi
	hasGPU, err := checkNvidiaSmi()
	if err != nil {
		// We don't return an error, just note that we don't have a GPU
		return &GPUMonitor{
			lastMeasurement: 0.0,
			lastUpdateTime:  time.Now(),
			hasGPU:          false,
		}, nil
	}

	// If we have a GPU, initialize with current utilization
	var utilization float64
	if hasGPU {
		utilization, err = getGPUUtilization()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize GPU monitor: %w", err)
		}
	}

	return &GPUMonitor{
		lastMeasurement: utilization,
		lastUpdateTime:  time.Now(),
		hasGPU:          hasGPU,
	}, nil
}

// GetUsage returns the current GPU usage as a percentage (0-100)
func (m *GPUMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If we don't have a GPU, always return 0
	if !m.hasGPU {
		return 0.0, nil
	}

	// On Linux, we'll refresh measurements every second at most
	now := time.Now()
	if now.Sub(m.lastUpdateTime) < time.Second {
		return m.lastMeasurement, nil
	}

	utilization, err := getGPUUtilization()
	if err != nil {
		// Return last known value on error
		return m.lastMeasurement, err
	}

	m.lastMeasurement = utilization
	m.lastUpdateTime = now

	return utilization, nil
}

// checkNvidiaSmi checks if nvidia-smi is available and NVIDIA GPU is present
func checkNvidiaSmi() (bool, error) {
	// Check if nvidia-smi is available
	_, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return false, fmt.Errorf("nvidia-smi not found: %w", err)
	}

	// Run nvidia-smi to check if GPU is accessible
	cmd := exec.Command("nvidia-smi", "--query-gpu=count", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute nvidia-smi: %w", err)
	}

	// Parse output
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return false, fmt.Errorf("failed to parse GPU count: %w", err)
	}

	return count > 0, nil
}

// getGPUUtilization gets GPU utilization using nvidia-smi
func getGPUUtilization() (float64, error) {
	// Run nvidia-smi to get GPU utilization
	cmd := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute nvidia-smi: %w", err)
	}

	// Parse output
	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return 0, fmt.Errorf("no output from nvidia-smi")
	}

	// If we have multiple GPUs, average the utilization
	var totalUtilization float64
	var gpuCount int

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		utilization, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			continue
		}

		totalUtilization += utilization
		gpuCount++
	}

	if gpuCount == 0 {
		return 0, fmt.Errorf("no valid GPU utilization data")
	}

	return totalUtilization / float64(gpuCount), nil
}