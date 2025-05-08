package monitor

import (
	"context"
	"testing"
	"time"
)

func TestMonitor_ResourceMonitoring(t *testing.T) {
	// Create a monitor with default configuration
	mon := NewMonitor()

	// Add a custom resource monitor
	customValue := 42.0
	mon.AddResourceMonitor("custom_metric", func() (float64, error) {
		return customValue, nil
	})

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the monitor
	err := mon.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	// Wait for the monitor to collect some data
	time.Sleep(2 * time.Second)

	// Get the current state
	state := mon.GetCurrentState()

	// Check that the standard resource types are present
	standardTypes := []ResourceType{CPU, Memory, Network, Disk, UserInput, GPU}
	for _, resourceType := range standardTypes {
		usage, ok := state.CurrentUsage[resourceType]
		if !ok {
			t.Errorf("Missing resource type: %s", resourceType)
			continue
		}

		if usage.Type != resourceType {
			t.Errorf("Expected resource type %s, got %s", resourceType, usage.Type)
		}
		if usage.Value < 0 || usage.Value > 100 {
			t.Errorf("Usage for %s out of bounds: %f", resourceType, usage.Value)
		}
		if usage.Timestamp.IsZero() {
			t.Errorf("Timestamp for %s is zero", resourceType)
		}
	}

	// Check that the custom resource type is present
	customType := ResourceType("custom_metric")
	usage, ok := state.CurrentUsage[customType]
	if !ok {
		t.Errorf("Missing custom resource type: %s", customType)
	} else {
		if usage.Type != customType {
			t.Errorf("Expected resource type %s, got %s", customType, usage.Type)
		}
		if usage.Value != customValue {
			t.Errorf("Expected custom value %f, got %f", customValue, usage.Value)
		}
		if usage.Timestamp.IsZero() {
			t.Errorf("Timestamp for %s is zero", customType)
		}
	}

	// Stop the monitor
	err = mon.Stop()
	if err != nil {
		t.Fatalf("Failed to stop monitor: %v", err)
	}
}

func TestMonitor_IdleDetection(t *testing.T) {
	// Create a monitor with a short naptime for testing
	config := DefaultConfig()
	config.NapTime = 2 * time.Second            // Shorter naptime for testing
	config.CheckInterval = 500 * time.Millisecond // More frequent checks for testing
	config.Thresholds[CPU] = 99.0               // Set high threshold so we're always "idle"

	mon := NewMonitorWithConfig(config)

	// Create a channel to receive idle state changes
	idleStateChanges := make(chan bool, 10)
	mon.OnIdleStateChange(func(isIdle bool, duration time.Duration) {
		idleStateChanges <- isIdle
	})

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the monitor
	err := mon.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	// We should transition to idle state soon
	select {
	case isIdle := <-idleStateChanges:
		if !isIdle {
			t.Error("Expected transition to idle state, got active state")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for idle state transition")
	}

	// After more time, we should have been idle for longer than the naptime
	time.Sleep(3 * time.Second)

	// Check idle status
	if !mon.IsIdle() {
		t.Error("Expected to be idle, got active")
	}

	idleDuration := mon.IdleDuration()
	if idleDuration < config.NapTime {
		t.Errorf("Expected idle duration >= %v, got %v", config.NapTime, idleDuration)
	}

	// Stop the monitor
	err = mon.Stop()
	if err != nil {
		t.Fatalf("Failed to stop monitor: %v", err)
	}
}