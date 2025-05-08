package resources

import (
	"errors"
	"testing"
)

func TestMonitorManager_GetUsage(t *testing.T) {
	// Create a new monitor manager
	manager, err := NewMonitorManager()
	if err != nil {
		t.Fatalf("Failed to create monitor manager: %v", err)
	}

	// Get usage for CPU
	usage, err := manager.GetUsage(CPU)
	if err != nil {
		t.Fatalf("Failed to get CPU usage: %v", err)
	}

	// Check usage
	if usage.Type != CPU {
		t.Errorf("Expected resource type %s, got %s", CPU, usage.Type)
	}
	if usage.Value < 0 || usage.Value > 100 {
		t.Errorf("CPU usage out of bounds: %f", usage.Value)
	}
	if usage.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}

	// Get usage for Memory
	usage, err = manager.GetUsage(Memory)
	if err != nil {
		t.Fatalf("Failed to get Memory usage: %v", err)
	}

	// Check usage
	if usage.Type != Memory {
		t.Errorf("Expected resource type %s, got %s", Memory, usage.Type)
	}
	if usage.Value < 0 || usage.Value > 100 {
		t.Errorf("Memory usage out of bounds: %f", usage.Value)
	}
	if usage.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}

	// Get usage for non-existent resource
	_, err = manager.GetUsage("invalid")
	if err == nil {
		t.Error("Expected error for invalid resource type, got nil")
	}
}

func TestMonitorManager_AddCustomMonitor(t *testing.T) {
	// Create a new monitor manager
	manager, err := NewMonitorManager()
	if err != nil {
		t.Fatalf("Failed to create monitor manager: %v", err)
	}

	// Add a custom monitor
	customValue := 42.0
	manager.AddCustomMonitor("custom", func() (float64, error) {
		return customValue, nil
	})

	// Get usage for custom monitor
	usage, err := manager.GetCustomUsage("custom")
	if err != nil {
		t.Fatalf("Failed to get custom usage: %v", err)
	}

	// Check usage
	if usage.Type != "custom" {
		t.Errorf("Expected resource type %s, got %s", "custom", usage.Type)
	}
	if usage.Value != customValue {
		t.Errorf("Expected custom value %f, got %f", customValue, usage.Value)
	}

	// Add a failing custom monitor
	expectedErr := errors.New("custom error")
	manager.AddCustomMonitor("failing", func() (float64, error) {
		return 0, expectedErr
	})

	// Get usage for failing custom monitor
	_, err = manager.GetCustomUsage("failing")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Get usage for non-existent custom monitor
	_, err = manager.GetCustomUsage("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent custom monitor, got nil")
	}

	// Remove custom monitor
	manager.RemoveCustomMonitor("custom")

	// Try to get usage for removed custom monitor
	_, err = manager.GetCustomUsage("custom")
	if err == nil {
		t.Error("Expected error for removed custom monitor, got nil")
	}
}

func TestMonitorManager_GetAllUsage(t *testing.T) {
	// Create a new monitor manager
	manager, err := NewMonitorManager()
	if err != nil {
		t.Fatalf("Failed to create monitor manager: %v", err)
	}

	// Add a custom monitor
	customValue := 75.0
	manager.AddCustomMonitor("custom", func() (float64, error) {
		return customValue, nil
	})

	// Get all usage
	allUsage, err := manager.GetAllUsage()
	if err != nil {
		t.Fatalf("Failed to get all usage: %v", err)
	}

	// Check that all expected resource types are present
	expectedTypes := []ResourceType{CPU, Memory, Network, Disk, UserInput, GPU, "custom"}
	for _, expectedType := range expectedTypes {
		usage, ok := allUsage[expectedType]
		if !ok {
			t.Errorf("Missing resource type: %s", expectedType)
			continue
		}

		if usage.Type != expectedType {
			t.Errorf("Expected resource type %s, got %s", expectedType, usage.Type)
		}
		if usage.Value < 0 || usage.Value > 100 {
			t.Errorf("Usage for %s out of bounds: %f", expectedType, usage.Value)
		}
		if usage.Timestamp.IsZero() {
			t.Errorf("Timestamp for %s is zero", expectedType)
		}

		// Check custom value specifically
		if expectedType == "custom" && usage.Value != customValue {
			t.Errorf("Expected custom value %f, got %f", customValue, usage.Value)
		}
	}

	// Add a failing custom monitor
	expectedErr := errors.New("custom error")
	manager.AddCustomMonitor("failing", func() (float64, error) {
		return 0, expectedErr
	})

	// Get all usage should fail because of the failing custom monitor
	_, err = manager.GetAllUsage()
	if err == nil {
		t.Error("Expected error due to failing custom monitor, got nil")
	}
}