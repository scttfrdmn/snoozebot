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

// MemoryMonitor monitors memory usage on macOS systems
type MemoryMonitor struct {
	lastMeasurement float64
	lastUpdateTime  time.Time
	mutex           sync.Mutex
}

// NewMemoryMonitor creates a new memory monitor for macOS
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

	// On macOS, we'll refresh measurements every second at most
	// to avoid excessive system calls
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

// getMemoryUsage gets memory usage percentage from vm_stat command on macOS
func getMemoryUsage() (float64, error) {
	// Run vm_stat command to get memory statistics
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute vm_stat command: %w", err)
	}

	// Parse the output
	// vm_stat gives memory info in pages; we need to extract page size and counts
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected vm_stat output format")
	}

	// Extract page size from the first line
	// Format: "Mach Virtual Memory Statistics: (page size of 4096 bytes)"
	pageSizeLine := lines[0]
	pageSizeStr := strings.TrimSuffix(strings.Split(pageSizeLine, "page size of ")[1], " bytes)")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse page size: %w", err)
	}

	// Parse memory statistics
	var freePages, activePages, inactivePages, wiredPages, compressedPages uint64

	for _, line := range lines[1:] {
		if strings.Contains(line, "Pages free:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				freeStr := strings.TrimSpace(strings.Replace(parts[1], ".", "", -1))
				freePages, _ = strconv.ParseUint(freeStr, 10, 64)
			}
		} else if strings.Contains(line, "Pages active:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				activeStr := strings.TrimSpace(strings.Replace(parts[1], ".", "", -1))
				activePages, _ = strconv.ParseUint(activeStr, 10, 64)
			}
		} else if strings.Contains(line, "Pages inactive:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				inactiveStr := strings.TrimSpace(strings.Replace(parts[1], ".", "", -1))
				inactivePages, _ = strconv.ParseUint(inactiveStr, 10, 64)
			}
		} else if strings.Contains(line, "Pages wired down:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				wiredStr := strings.TrimSpace(strings.Replace(parts[1], ".", "", -1))
				wiredPages, _ = strconv.ParseUint(wiredStr, 10, 64)
			}
		} else if strings.Contains(line, "Pages occupied by compressor:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				compressedStr := strings.TrimSpace(strings.Replace(parts[1], ".", "", -1))
				compressedPages, _ = strconv.ParseUint(compressedStr, 10, 64)
			}
		}
	}

	// Calculate total physical memory
	// Since there's no direct way to get total physical memory from vm_stat,
	// we'll use sysctl command as a fallback
	cmd = exec.Command("sysctl", "-n", "hw.memsize")
	output, err = cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute sysctl command: %w", err)
	}

	totalMemoryBytes, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse total memory: %w", err)
	}

	totalPages := totalMemoryBytes / pageSize

	// Calculate used memory (wired + active + inactive + compressed)
	usedPages := wiredPages + activePages + inactivePages + compressedPages

	// Calculate memory usage percentage
	memoryUsage := 100.0 * float64(usedPages) / float64(totalPages)

	return memoryUsage, nil
}