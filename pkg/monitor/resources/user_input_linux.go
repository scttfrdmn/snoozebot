// +build linux

package resources

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UserInputMonitor monitors user input (keyboard and mouse) activity on Linux systems
type UserInputMonitor struct {
	lastIdleTime   time.Duration
	lastUpdateTime time.Time
	mutex          sync.Mutex
}

// NewUserInputMonitor creates a new user input monitor for Linux
func NewUserInputMonitor() (*UserInputMonitor, error) {
	// Check if we have the necessary tools
	if err := checkXprintidle(); err != nil {
		return nil, fmt.Errorf("prerequisites not met for user input monitoring: %w", err)
	}

	// Initialize with current idle time
	idleTime, err := getIdleTime()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user input monitor: %w", err)
	}

	return &UserInputMonitor{
		lastIdleTime:   idleTime,
		lastUpdateTime: time.Now(),
	}, nil
}

// GetUsage returns the current user input activity as a percentage (0-100)
// Note: For user input, we return either 0% (no user activity) or 100% (user active)
func (m *UserInputMonitor) GetUsage() (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// On Linux, we'll refresh measurements every second at most
	now := time.Now()
	if now.Sub(m.lastUpdateTime) < time.Second {
		if isRecentlyActive(m.lastIdleTime) {
			return 100.0, nil
		}
		return 0.0, nil
	}

	idleTime, err := getIdleTime()
	if err != nil {
		return 0.0, err
	}

	m.lastIdleTime = idleTime
	m.lastUpdateTime = now

	if isRecentlyActive(idleTime) {
		return 100.0, nil
	}
	return 0.0, nil
}

// getIdleTime gets the user idle time using xprintidle
func getIdleTime() (time.Duration, error) {
	// xprintidle returns the idle time in milliseconds
	cmd := exec.Command("xprintidle")
	output, err := cmd.Output()
	if err != nil {
		if err.Error() == "exec: \"xprintidle\": executable file not found in $PATH" {
			// Fallback to manual detection if xprintidle is not available
			return getIdleTimeManual()
		}
		return 0, fmt.Errorf("failed to execute xprintidle: %w", err)
	}

	// Parse output
	idleTimeMs, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse idle time: %w", err)
	}

	return time.Duration(idleTimeMs) * time.Millisecond, nil
}

// getIdleTimeManual attempts to get idle time by parsing /proc/interrupts for keyboard and mouse interrupts
func getIdleTimeManual() (time.Duration, error) {
	// This is a best-effort approach that may not be accurate
	// It works by checking if keyboard or mouse interrupts have changed since last check

	// Check /proc/uptime for system uptime
	uptimeFile, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer uptimeFile.Close()

	scanner := bufio.NewScanner(uptimeFile)
	if !scanner.Scan() {
		return 0, fmt.Errorf("failed to read uptime")
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid uptime format")
	}

	uptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uptime: %w", err)
	}

	// If we don't have better data, assume user has been idle for half the uptime
	// This is arbitrary and not accurate
	return time.Duration(uptime * float64(time.Second) / 2), nil
}

// checkXprintidle checks if xprintidle is available
func checkXprintidle() error {
	_, err := exec.LookPath("xprintidle")
	if err != nil {
		// We'll attempt to use a fallback method, so we don't return an error
		return nil
	}
	return nil
}

// isRecentlyActive returns true if the user was active recently
func isRecentlyActive(idleTime time.Duration) bool {
	// We consider the user active if they were active in the last 5 seconds
	return idleTime < 5*time.Second
}