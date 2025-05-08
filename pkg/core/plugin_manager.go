package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-hclog"
	hashicorpPlugin "github.com/hashicorp/go-plugin"
	"github.com/scottfridman/snoozebot/pkg/plugin"
)

// PluginManager handles the loading and management of plugins
type PluginManager struct {
	pluginsDir string
	plugins    map[string]*pluginInstance
	mutex      sync.RWMutex
	logger     hclog.Logger
}

// pluginInstance represents a running plugin instance
type pluginInstance struct {
	client     hashicorpPlugin.ClientProtocol
	provider   plugin.CloudProvider
	pluginPath string
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginsDir string, logger hclog.Logger) *PluginManager {
	if logger == nil {
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin-manager",
			Level:  hclog.Info,
			Output: os.Stderr,
		})
	}

	return &PluginManager{
		pluginsDir: pluginsDir,
		plugins:    make(map[string]*pluginInstance),
		logger:     logger,
	}
}

// LoadPlugin loads a plugin from the plugins directory
func (pm *PluginManager) LoadPlugin(ctx context.Context, pluginName string) (plugin.CloudProvider, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Check if plugin is already loaded
	if instance, ok := pm.plugins[pluginName]; ok {
		return instance.provider, nil
	}

	// Find the plugin in the plugins directory
	pluginPath := filepath.Join(pm.pluginsDir, pluginName)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	// Configure the plugin client
	client := hashicorpPlugin.NewClient(&hashicorpPlugin.ClientConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins:         plugin.PluginMap,
		Cmd:             hashicorpPlugin.Command{Path: pluginPath},
		Logger:          pm.logger,
		AllowedProtocols: []hashicorpPlugin.Protocol{
			hashicorpPlugin.ProtocolGRPC,
		},
	})

	// Connect to the plugin
	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("error connecting to plugin: %w", err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("cloud_provider")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("error dispensing plugin: %w", err)
	}

	// Cast to the correct interface
	provider, ok := raw.(plugin.CloudProvider)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin does not implement CloudProvider interface")
	}

	// Store the plugin instance
	pm.plugins[pluginName] = &pluginInstance{
		client:     rpcClient,
		provider:   provider,
		pluginPath: pluginPath,
	}

	return provider, nil
}

// UnloadPlugin unloads a plugin
func (pm *PluginManager) UnloadPlugin(pluginName string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	instance, ok := pm.plugins[pluginName]
	if !ok {
		return fmt.Errorf("plugin not loaded: %s", pluginName)
	}

	// Kill the client, which will clean up the plugin process
	client, ok := instance.client.(*hashicorpPlugin.GRPCClient)
	if ok {
		client.Close()
	}

	delete(pm.plugins, pluginName)
	return nil
}

// ListLoadedPlugins returns a list of loaded plugins
func (pm *PluginManager) ListLoadedPlugins() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugins := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		plugins = append(plugins, name)
	}

	return plugins
}

// DiscoverPlugins discovers plugins in the plugins directory
func (pm *PluginManager) DiscoverPlugins() ([]string, error) {
	// Ensure plugins directory exists
	if _, err := os.Stat(pm.pluginsDir); os.IsNotExist(err) {
		return nil, nil
	}

	// Read plugins directory
	entries, err := os.ReadDir(pm.pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading plugins directory: %w", err)
	}

	plugins := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if the file is executable
		pluginPath := filepath.Join(pm.pluginsDir, entry.Name())
		info, err := os.Stat(pluginPath)
		if err != nil {
			continue
		}

		// Check if the file is executable on Unix systems
		if info.Mode()&0111 != 0 {
			plugins = append(plugins, entry.Name())
		}
	}

	return plugins, nil
}

// Cleanup cleans up all loaded plugins
func (pm *PluginManager) Cleanup() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for name, instance := range pm.plugins {
		client, ok := instance.client.(*hashicorpPlugin.GRPCClient)
		if ok {
			client.Close()
		}
		delete(pm.plugins, name)
	}
}