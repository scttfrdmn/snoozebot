package provider

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
)

// TestNewPluginManager tests creating a new plugin manager
func TestNewPluginManager(t *testing.T) {
	// Create a temporary directory for plugins
	tempDir, err := ioutil.TempDir("", "plugins")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	// Create a new plugin manager
	pm := NewPluginManager(tempDir, logger)
	if pm == nil {
		t.Fatal("Expected non-nil PluginManager")
	}

	// Check the plugin manager implementation
	impl, ok := pm.(*PluginManagerImpl)
	if !ok {
		t.Fatal("Expected PluginManagerImpl")
	}

	// Check the plugins directory
	if impl.pluginsDir != tempDir {
		t.Errorf("Expected pluginsDir %s, got %s", tempDir, impl.pluginsDir)
	}

	// Check the loaded plugins map
	if len(impl.loadedPlugins) != 0 {
		t.Errorf("Expected empty loadedPlugins map, got %d entries", len(impl.loadedPlugins))
	}

	// Check the logger
	if impl.logger != logger {
		t.Error("Expected logger to be the same as provided")
	}
}

// TestDiscoverPlugins tests discovering plugins in the plugins directory
func TestDiscoverPlugins(t *testing.T) {
	// Create a temporary directory for plugins
	tempDir, err := ioutil.TempDir("", "plugins")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a plugin executable file
	pluginPath := filepath.Join(tempDir, "test-plugin")
	err = ioutil.WriteFile(pluginPath, []byte("#!/bin/sh\necho 'Test plugin'"), 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin file: %v", err)
	}

	// Create a non-executable file
	nonExecPath := filepath.Join(tempDir, "non-exec")
	err = ioutil.WriteFile(nonExecPath, []byte("Not executable"), 0644)
	if err != nil {
		t.Fatalf("Failed to create non-executable file: %v", err)
	}

	// Create a directory
	dirPath := filepath.Join(tempDir, "dir")
	err = os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create a plugin manager
	pm := NewPluginManager(tempDir, nil)

	// Discover plugins
	plugins, err := pm.DiscoverPlugins()
	if err != nil {
		t.Fatalf("Failed to discover plugins: %v", err)
	}

	// Check the discovered plugins
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}

	if len(plugins) > 0 && plugins[0] != "test-plugin" {
		t.Errorf("Expected plugin 'test-plugin', got '%s'", plugins[0])
	}
}

// TestListPlugins tests listing loaded plugins
func TestListPlugins(t *testing.T) {
	// Create a plugin manager
	pm := &PluginManagerImpl{
		loadedPlugins: map[string]*pluginInstance{
			"plugin1": {},
			"plugin2": {},
		},
		logger: hclog.NewNullLogger(),
	}

	// List plugins
	plugins := pm.ListPlugins()

	// Check the listed plugins
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}

	// Check that both plugins are in the list
	found1 := false
	found2 := false
	for _, p := range plugins {
		if p == "plugin1" {
			found1 = true
		} else if p == "plugin2" {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Expected to find 'plugin1' in the list")
	}
	if !found2 {
		t.Error("Expected to find 'plugin2' in the list")
	}
}

// TestGetPlugin tests getting a loaded plugin
func TestGetPlugin(t *testing.T) {
	// Create a mock cloud provider
	mockCP := &MockCloudProvider{}

	// Create a plugin manager with a loaded plugin
	pm := &PluginManagerImpl{
		loadedPlugins: map[string]*pluginInstance{
			"mock": {
				cloudProvider: mockCP,
			},
		},
		logger: hclog.NewNullLogger(),
	}

	// Get the plugin
	cp, err := pm.GetPlugin("mock")
	if err != nil {
		t.Fatalf("Failed to get plugin: %v", err)
	}

	// Check the cloud provider
	if cp != mockCP {
		t.Error("Expected the same cloud provider instance")
	}

	// Try to get a non-existent plugin
	_, err = pm.GetPlugin("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent plugin")
	}
}

// MockCloudProvider is a mock implementation of CloudProvider for testing
type MockCloudProvider struct{}

func (m *MockCloudProvider) GetInstanceInfo(ctx context.Context, instanceID string) (*InstanceInfo, error) {
	return &InstanceInfo{
		ID:         instanceID,
		Name:       "mock-instance",
		Type:       "mock-type",
		Region:     "mock-region",
		Zone:       "mock-zone",
		State:      "running",
		LaunchTime: time.Now(),
		Provider:   "mock",
	}, nil
}

func (m *MockCloudProvider) StopInstance(ctx context.Context, instanceID string) error {
	return nil
}

func (m *MockCloudProvider) StartInstance(ctx context.Context, instanceID string) error {
	return nil
}

func (m *MockCloudProvider) GetProviderName() string {
	return "mock"
}

func (m *MockCloudProvider) GetProviderVersion() string {
	return "1.0.0"
}