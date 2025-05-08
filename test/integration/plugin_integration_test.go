package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/agent/provider"
)

func TestPluginIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("SNOOZEBOT_RUN_INTEGRATION") != "true" {
		t.Skip("Skipping integration test. Set SNOOZEBOT_RUN_INTEGRATION=true to run")
	}

	// Create a temp directory for plugins
	tempDir, err := os.MkdirTemp("", "snoozebot-plugin-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin-test",
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	// Create plugin manager
	pluginManager := provider.NewPluginManager(tempDir, logger)

	// Build a test plugin (mock AWS)
	if err := buildTestPlugin(t, tempDir); err != nil {
		t.Fatalf("Failed to build test plugin: %v", err)
	}

	// Test plugin discovery
	plugins, err := pluginManager.DiscoverPlugins()
	if err != nil {
		t.Fatalf("Failed to discover plugins: %v", err)
	}

	if len(plugins) != 1 {
		t.Fatalf("Expected 1 plugin, got %d", len(plugins))
	}

	if plugins[0] != "mock-aws" {
		t.Fatalf("Expected plugin 'mock-aws', got '%s'", plugins[0])
	}

	// Test plugin loading
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	plugin, err := pluginManager.LoadPlugin(ctx, "mock-aws")
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Test plugin operations
	name := plugin.GetProviderName()
	if name != "mock-aws" {
		t.Errorf("Expected provider name 'mock-aws', got '%s'", name)
	}

	version := plugin.GetProviderVersion()
	if version != "0.1.0" {
		t.Errorf("Expected provider version '0.1.0', got '%s'", version)
	}

	// Test getting instance info
	info, err := plugin.GetInstanceInfo(ctx, "i-mock")
	if err != nil {
		t.Fatalf("Failed to get instance info: %v", err)
	}

	if info.ID != "i-mock" {
		t.Errorf("Expected instance ID 'i-mock', got '%s'", info.ID)
	}

	// Test unloading plugin
	err = pluginManager.UnloadPlugin("mock-aws")
	if err != nil {
		t.Fatalf("Failed to unload plugin: %v", err)
	}

	// Verify plugin was unloaded
	loadedPlugins := pluginManager.ListPlugins()
	if len(loadedPlugins) != 0 {
		t.Errorf("Expected 0 loaded plugins after unload, got %d", len(loadedPlugins))
	}
}

// buildTestPlugin builds a mock AWS plugin for testing
func buildTestPlugin(t *testing.T, pluginsDir string) error {
	// Create a mock plugin
	mockPluginCode := `package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/scttfrdmn/snoozebot/agent/provider"
)

// MockAWSProvider is a mock AWS provider for testing
type MockAWSProvider struct {
	logger hclog.Logger
}

// GetInstanceInfo gets information about an instance
func (p *MockAWSProvider) GetInstanceInfo(ctx context.Context, instanceID string) (*provider.InstanceInfo, error) {
	p.logger.Info("Getting instance info", "instanceID", instanceID)
	
	return &provider.InstanceInfo{
		ID:         instanceID,
		Name:       "mock-instance",
		Type:       "mock.small",
		Region:     "mock-region-1",
		Zone:       "mock-region-1a",
		State:      "running",
		LaunchTime: time.Now(),
		Provider:   "mock-aws",
	}, nil
}

// StopInstance stops an instance
func (p *MockAWSProvider) StopInstance(ctx context.Context, instanceID string) error {
	p.logger.Info("Stopping instance", "instanceID", instanceID)
	return nil
}

// StartInstance starts an instance
func (p *MockAWSProvider) StartInstance(ctx context.Context, instanceID string) error {
	p.logger.Info("Starting instance", "instanceID", instanceID)
	return nil
}

// GetProviderName returns the name of the cloud provider
func (p *MockAWSProvider) GetProviderName() string {
	return "mock-aws"
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *MockAWSProvider) GetProviderVersion() string {
	return "0.1.0"
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user-friendly error is shown.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SNOOZEBOT_PLUGIN",
	MagicCookieValue: "snoozebot_provider_v1",
}

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "mock-aws-provider",
		Level:  hclog.Info,
		Output: os.Stderr,
	})

	// Create the provider
	mockProvider := &MockAWSProvider{
		logger: logger,
	}

	// Set up the plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"cloud_provider": &provider.CloudProviderPlugin{
				Impl: mockProvider,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
`

	// Write the mock plugin to a file
	mockPluginPath := filepath.Join(pluginsDir, "mock-aws.go")
	if err := os.WriteFile(mockPluginPath, []byte(mockPluginCode), 0644); err != nil {
		return fmt.Errorf("failed to write mock plugin: %w", err)
	}

	// Compile the mock plugin
	outputPath := filepath.Join(pluginsDir, "mock-aws")
	cmd := exec.Command("go", "build", "-o", outputPath, mockPluginPath)
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compile mock plugin: %w\nOutput: %s", err, string(output))
	}

	// Make the plugin executable
	if err := os.Chmod(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to make plugin executable: %w", err)
	}

	return nil
}