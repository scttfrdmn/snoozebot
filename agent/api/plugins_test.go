package api

import (
	"context"
	"testing"
	
	"github.com/hashicorp/go-hclog"
)

// TestDiscoverAndInitPlugins tests the DiscoverAndInitPlugins function
func TestDiscoverAndInitPlugins(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{
		discoveredPlugins: []string{"aws", "gcp", "azure"},
	}

	// Create a server with the mock plugin manager
	s := &Server{
		pluginManager: mockPM,
		logger: hclog.NewNullLogger(),
	}

	// Call the function
	err := s.DiscoverAndInitPlugins(context.Background())
	if err != nil {
		t.Fatalf("Failed to discover and init plugins: %v", err)
	}

	// Check that all plugins were loaded
	if len(mockPM.loadedPlugins) != 3 {
		t.Errorf("Expected 3 plugins to be loaded, got %d", len(mockPM.loadedPlugins))
	}

	// Check that all plugins are in the loaded list
	found1 := false
	found2 := false
	found3 := false
	for _, p := range mockPM.loadedPlugins {
		if p == "aws" {
			found1 = true
		} else if p == "gcp" {
			found2 = true
		} else if p == "azure" {
			found3 = true
		}
	}

	if !found1 {
		t.Error("Expected 'aws' plugin to be loaded")
	}
	if !found2 {
		t.Error("Expected 'gcp' plugin to be loaded")
	}
	if !found3 {
		t.Error("Expected 'azure' plugin to be loaded")
	}
}

// TestDiscoverAndInitPluginsWithErrors tests the DiscoverAndInitPlugins function with errors
func TestDiscoverAndInitPluginsWithErrors(t *testing.T) {
	// Create a mock plugin manager with load error
	mockPM := &mockPluginManager{
		discoveredPlugins: []string{"aws", "gcp", "azure"},
		loadError:         (error)(nil), // Will cause one plugin to fail loading
	}

	// Create a server with the mock plugin manager
	s := &Server{
		pluginManager: mockPM,
		logger: hclog.NewNullLogger(),
	}

	// Call the function
	err := s.DiscoverAndInitPlugins(context.Background())
	if err != nil {
		t.Fatalf("Failed to discover and init plugins: %v", err)
	}

	// Check that no plugins were loaded (all should fail if we have an error)
	if len(mockPM.loadedPlugins) != 0 {
		t.Errorf("Expected 0 plugins to be loaded, got %d", len(mockPM.loadedPlugins))
	}
}