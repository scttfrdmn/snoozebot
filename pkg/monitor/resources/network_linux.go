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

// NetworkStats represents network statistics
type NetworkStats struct {
	BytesReceived uint64
	BytesSent     uint64
	Timestamp     time.Time
}

// NetworkMonitor monitors network usage on Linux systems
type NetworkMonitor struct {
	lastStats      *NetworkStats
	lastMeasureTs  time.Time
	maxBytesPerSec uint64
	mutex          sync.Mutex
}

// NewNetworkMonitor creates a new network monitor for Linux
func NewNetworkMonitor() (*NetworkMonitor, error) {
	// Initialize with current stats
	stats, err := readNetworkStats()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize network monitor: %w", err)
	}

	// Set a reasonable default for maximum bytes per second
	// 100 Mbps = 12.5 MB/s = 12,500,000 bytes/sec
	maxBytesPerSec := uint64(12500000)

	return &NetworkMonitor{
		lastStats:      stats,
		lastMeasureTs:  time.Now(),
		maxBytesPerSec: maxBytesPerSec,
	}, nil
}

// GetUsage returns the current network usage as a percentage (0-100)
func (m *NetworkMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Read current network stats
	currentStats, err := readNetworkStats()
	if err != nil {
		return 0, fmt.Errorf("failed to read network stats: %w", err)
	}

	// Calculate time difference
	now := time.Now()
	timeDiff := now.Sub(m.lastMeasureTs).Seconds()
	if timeDiff <= 0 {
		return 0, fmt.Errorf("time difference too small")
	}

	// Calculate network usage
	bytesReceivedDelta := currentStats.BytesReceived - m.lastStats.BytesReceived
	bytesSentDelta := currentStats.BytesSent - m.lastStats.BytesSent

	// Calculate total bytes per second
	bytesPerSec := float64(bytesReceivedDelta+bytesSentDelta) / timeDiff

	// Calculate network usage as a percentage of maximum capacity
	networkUsage := 100.0 * bytesPerSec / float64(m.maxBytesPerSec)

	// Cap at 100%
	if networkUsage > 100.0 {
		networkUsage = 100.0
	}

	// Update last stats
	m.lastStats = currentStats
	m.lastMeasureTs = now

	return networkUsage, nil
}

// readNetworkStats reads network statistics from /proc/net/dev
func readNetworkStats() (*NetworkStats, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := &NetworkStats{
		Timestamp: time.Now(),
	}

	scanner := bufio.NewScanner(file)
	// Skip the first two header lines
	scanner.Scan() // Inter-|   Receive                                             |  Transmit
	scanner.Scan() // face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed

	var totalBytesReceived, totalBytesSent uint64

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 17 {
			continue
		}

		// Skip loopback interface
		interfaceName := strings.Trim(fields[0], " :")
		if interfaceName == "lo" {
			continue
		}

		bytesReceived, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		bytesSent, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			continue
		}

		totalBytesReceived += bytesReceived
		totalBytesSent += bytesSent
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	stats.BytesReceived = totalBytesReceived
	stats.BytesSent = totalBytesSent

	return stats, nil
}