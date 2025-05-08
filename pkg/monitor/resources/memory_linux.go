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

// MemoryStats represents memory statistics
type MemoryStats struct {
	Total     uint64
	Free      uint64
	Available uint64
	Buffers   uint64
	Cached    uint64
}

// MemoryMonitor monitors memory usage on Linux systems
type MemoryMonitor struct {
	lastMeasurement float64
	lastUpdateTime  time.Time
	mutex           sync.Mutex
}

// NewMemoryMonitor creates a new memory monitor for Linux
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

	// Refresh measurements every second at most
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

// getMemoryUsage gets memory usage percentage from /proc/meminfo
func getMemoryUsage() (float64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	stats := MemoryStats{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := fields[0]
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "MemTotal:":
			stats.Total = value
		case "MemFree:":
			stats.Free = value
		case "MemAvailable:":
			stats.Available = value
		case "Buffers:":
			stats.Buffers = value
		case "Cached:":
			stats.Cached = value
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	if stats.Total == 0 {
		return 0, fmt.Errorf("could not determine total memory")
	}

	// Calculate memory usage
	// If MemAvailable is available (kernel 3.14+), use it for a more accurate calculation
	var usedMemory uint64
	if stats.Available > 0 {
		usedMemory = stats.Total - stats.Available
	} else {
		// Older kernels: use free + buffers + cached
		usedMemory = stats.Total - (stats.Free + stats.Buffers + stats.Cached)
	}

	// Calculate percentage
	memoryUsage := 100.0 * float64(usedMemory) / float64(stats.Total)

	return memoryUsage, nil
}