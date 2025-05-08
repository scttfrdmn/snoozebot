// +build linux

package resources

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CPUStats represents CPU statistics
type CPUStats struct {
	User    uint64 // user mode
	Nice    uint64 // user mode with low priority
	System  uint64 // system mode
	Idle    uint64 // idle time
	IOWait  uint64 // I/O wait time
	IRQ     uint64 // IRQ time
	SoftIRQ uint64 // softIRQ time
	Steal   uint64 // time spent in other OSes when running in a virtualized environment
	Guest   uint64 // time spent running a virtual CPU for guest OSes
	Total   uint64 // total of all time fields
}

// CPUMonitor monitors CPU usage on Linux systems
type CPUMonitor struct {
	lastStats     *CPUStats
	lastMeasureTs time.Time
	mutex         sync.Mutex
}

// NewCPUMonitor creates a new CPU monitor for Linux
func NewCPUMonitor() (*CPUMonitor, error) {
	// Initialize with current stats
	stats, err := readCPUStats()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize CPU monitor: %w", err)
	}

	return &CPUMonitor{
		lastStats:     stats,
		lastMeasureTs: time.Now(),
	}, nil
}

// GetUsage returns the current CPU usage as a percentage (0-100)
func (m *CPUMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Read current CPU stats
	currentStats, err := readCPUStats()
	if err != nil {
		return 0, fmt.Errorf("failed to read CPU stats: %w", err)
	}

	// Calculate time difference
	now := time.Now()
	timeDiff := now.Sub(m.lastMeasureTs).Seconds()
	if timeDiff <= 0 {
		return 0, fmt.Errorf("time difference too small")
	}

	// Calculate CPU usage
	idleDelta := float64(currentStats.Idle - m.lastStats.Idle)
	totalDelta := float64(currentStats.Total - m.lastStats.Total)

	var cpuUsage float64
	if totalDelta > 0 {
		cpuUsage = 100.0 * (1.0 - idleDelta/totalDelta)
	}

	// Update last stats
	m.lastStats = currentStats
	m.lastMeasureTs = now

	return cpuUsage, nil
}

// readCPUStats reads CPU statistics from /proc/stat
func readCPUStats() (*CPUStats, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			return nil, fmt.Errorf("invalid CPU stats format: %s", line)
		}

		stats := &CPUStats{}
		stats.User, _ = strconv.ParseUint(fields[1], 10, 64)
		stats.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
		stats.System, _ = strconv.ParseUint(fields[3], 10, 64)
		stats.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
		stats.IOWait, _ = strconv.ParseUint(fields[5], 10, 64)
		stats.IRQ, _ = strconv.ParseUint(fields[6], 10, 64)
		stats.SoftIRQ, _ = strconv.ParseUint(fields[7], 10, 64)

		if len(fields) > 8 {
			stats.Steal, _ = strconv.ParseUint(fields[8], 10, 64)
		}
		if len(fields) > 9 {
			stats.Guest, _ = strconv.ParseUint(fields[9], 10, 64)
		}

		// Calculate total time
		stats.Total = stats.User + stats.Nice + stats.System + stats.Idle +
			stats.IOWait + stats.IRQ + stats.SoftIRQ + stats.Steal + stats.Guest

		return stats, nil
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("CPU stats not found in /proc/stat")
}