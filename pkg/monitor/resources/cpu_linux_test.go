// +build linux

package resources

import (
	"os"
	"testing"
	"time"
)

func TestCPUMonitor_Linux(t *testing.T) {
	// Skip test if not running on Linux
	if os.Getenv("SKIP_LINUX_TESTS") != "" {
		t.Skip("Skipping Linux-specific test")
	}

	// Create a new CPU monitor
	monitor, err := NewCPUMonitor()
	if err != nil {
		t.Fatalf("Failed to create CPU monitor: %v", err)
	}

	// Get initial CPU usage
	usage1, err := monitor.GetUsage()
	if err != nil {
		t.Fatalf("Failed to get CPU usage: %v", err)
	}

	// Check that usage is within reasonable bounds
	if usage1 < 0 || usage1 > 100 {
		t.Errorf("CPU usage out of bounds: %f", usage1)
	}

	// Generate some CPU load
	go func() {
		end := time.Now().Add(500 * time.Millisecond)
		for time.Now().Before(end) {
			// Do some meaningless work to generate CPU load
			for i := 0; i < 1000000; i++ {
				_ = i * i
			}
		}
	}()

	// Wait a bit to allow for the load generation and for the monitor to measure it
	time.Sleep(1 * time.Second)

	// Get CPU usage again
	usage2, err := monitor.GetUsage()
	if err != nil {
		t.Fatalf("Failed to get CPU usage: %v", err)
	}

	// Check that usage is still within reasonable bounds
	if usage2 < 0 || usage2 > 100 {
		t.Errorf("CPU usage out of bounds: %f", usage2)
	}

	// We can't make strong assertions about the actual values since
	// they depend on the system load, but we can log them for manual review
	t.Logf("CPU Usage 1: %f%%", usage1)
	t.Logf("CPU Usage 2: %f%%", usage2)
}

func TestReadCPUStats(t *testing.T) {
	// Skip test if not running on Linux
	if os.Getenv("SKIP_LINUX_TESTS") != "" {
		t.Skip("Skipping Linux-specific test")
	}

	// Read CPU stats
	stats, err := readCPUStats()
	if err != nil {
		t.Fatalf("Failed to read CPU stats: %v", err)
	}

	// Check that stats are reasonable
	if stats.Total == 0 {
		t.Error("Total CPU time is zero")
	}
	if stats.Idle == 0 {
		t.Error("Idle CPU time is zero")
	}
	if stats.User == 0 {
		t.Error("User CPU time is zero")
	}

	// Log stats for manual review
	t.Logf("CPU Stats: User=%d, Nice=%d, System=%d, Idle=%d, IOWait=%d, IRQ=%d, SoftIRQ=%d, Steal=%d, Guest=%d, Total=%d",
		stats.User, stats.Nice, stats.System, stats.Idle, stats.IOWait, stats.IRQ, stats.SoftIRQ, stats.Steal, stats.Guest, stats.Total)
}