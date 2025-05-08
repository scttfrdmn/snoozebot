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

// DiskStats represents disk I/O statistics
type DiskStats struct {
	// Total sectors read
	SectorsRead uint64
	// Total sectors written
	SectorsWritten uint64
	// Timestamp of measurement
	Timestamp time.Time
}

// DiskMonitor monitors disk I/O on Linux systems
type DiskMonitor struct {
	lastStats     *DiskStats
	lastMeasureTs time.Time
	// Max sectors per second (based on reasonable SSD performance)
	// ~500MB/s = ~1,000,000 sectors/s (assuming 512 bytes per sector)
	maxSectorsPerSec uint64
	mutex            sync.Mutex
}

// NewDiskMonitor creates a new disk monitor for Linux
func NewDiskMonitor() (*DiskMonitor, error) {
	// Initialize with current stats
	stats, err := readDiskStats()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize disk monitor: %w", err)
	}

	return &DiskMonitor{
		lastStats:        stats,
		lastMeasureTs:    time.Now(),
		maxSectorsPerSec: 1000000, // 500MB/s assuming 512 bytes per sector
	}, nil
}

// GetUsage returns the current disk I/O usage as a percentage (0-100)
func (m *DiskMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Read current disk stats
	currentStats, err := readDiskStats()
	if err != nil {
		return 0, fmt.Errorf("failed to read disk stats: %w", err)
	}

	// Calculate time difference
	now := time.Now()
	timeDiff := now.Sub(m.lastMeasureTs).Seconds()
	if timeDiff <= 0 {
		return 0, fmt.Errorf("time difference too small")
	}

	// Calculate disk I/O
	readDelta := currentStats.SectorsRead - m.lastStats.SectorsRead
	writeDelta := currentStats.SectorsWritten - m.lastStats.SectorsWritten

	// Calculate total sectors per second
	sectorsPerSec := float64(readDelta+writeDelta) / timeDiff

	// Calculate disk usage as a percentage of maximum capacity
	diskUsage := 100.0 * sectorsPerSec / float64(m.maxSectorsPerSec)

	// Cap at 100%
	if diskUsage > 100.0 {
		diskUsage = 100.0
	}

	// Update last stats
	m.lastStats = currentStats
	m.lastMeasureTs = now

	return diskUsage, nil
}

// readDiskStats reads disk statistics from /proc/diskstats
func readDiskStats() (*DiskStats, error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := &DiskStats{
		Timestamp: time.Now(),
	}

	scanner := bufio.NewScanner(file)
	var totalSectorsRead, totalSectorsWritten uint64

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		// Skip non-device entries and partitions (we only want whole devices)
		deviceName := fields[2]
		if strings.HasPrefix(deviceName, "loop") || strings.HasPrefix(deviceName, "ram") {
			continue
		}

		// Skip partitions (names with numbers)
		if containsDigit(deviceName) {
			continue
		}

		// Parse sectors read (field 6)
		sectorsRead, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			continue
		}

		// Parse sectors written (field 10)
		sectorsWritten, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			continue
		}

		totalSectorsRead += sectorsRead
		totalSectorsWritten += sectorsWritten
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	stats.SectorsRead = totalSectorsRead
	stats.SectorsWritten = totalSectorsWritten

	return stats, nil
}

// containsDigit checks if a string contains any digits
func containsDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}