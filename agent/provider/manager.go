package provider

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	pluginlib "github.com/scttfrdmn/snoozebot/pkg/plugin"
)

// defaultPluginLogger creates a default logger for plugins
func defaultPluginLogger() hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Info,
	})
}

// PluginManagerImpl implements the PluginManager interface
type PluginManagerImpl struct {
	pluginsDir      string
	loadedPlugins   map[string]*pluginInstance
	logger          hclog.Logger
	mu              sync.RWMutex
}

// pluginInstance represents a loaded plugin instance
type pluginInstance struct {
	client         goplugin.ClientProtocol
	pluginClient   *goplugin.Client
	cloudProvider  CloudProvider
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginsDir string, logger hclog.Logger) PluginManager {
	if logger == nil {
		logger = defaultPluginLogger()
	}
	
	return &PluginManagerImpl{
		pluginsDir:    pluginsDir,
		loadedPlugins: make(map[string]*pluginInstance),
		logger:        logger,
	}
}

// LoadPlugin loads a cloud provider plugin
func (pm *PluginManagerImpl) LoadPlugin(ctx context.Context, pluginName string) (CloudProvider, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if plugin is already loaded
	if instance, ok := pm.loadedPlugins[pluginName]; ok {
		return instance.cloudProvider, nil
	}

	// Check if plugin exists
	pluginPath := filepath.Join(pm.pluginsDir, pluginName)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin %s not found in %s", pluginName, pm.pluginsDir)
	}

	// Create plugin client
	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig: pluginlib.Handshake,
		Plugins:         pluginlib.PluginMap,
		Cmd:             exec.Command(pluginPath),
		Logger:          pm.logger,
		AllowedProtocols: []goplugin.Protocol{
			goplugin.ProtocolGRPC,
		},
	})

	// Connect to the plugin
	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("error connecting to plugin %s: %w", pluginName, err)
	}

	// Get the cloud provider interface
	raw, err := rpcClient.Dispense("cloud_provider")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("error dispensing plugin %s: %w", pluginName, err)
	}

	// Cast to CloudProvider interface
	cp, ok := raw.(CloudProvider)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin %s does not implement CloudProvider interface", pluginName)
	}

	// Store the plugin instance
	pm.loadedPlugins[pluginName] = &pluginInstance{
		client:         rpcClient,
		pluginClient:   client,
		cloudProvider:  cp,
	}

	pm.logger.Info("Loaded plugin", "name", pluginName, "provider", cp.GetProviderName(), "version", cp.GetProviderVersion())
	return cp, nil
}

// UnloadPlugin unloads a cloud provider plugin
func (pm *PluginManagerImpl) UnloadPlugin(pluginName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	instance, ok := pm.loadedPlugins[pluginName]
	if !ok {
		return fmt.Errorf("plugin %s not loaded", pluginName)
	}

	// Kill the plugin process
	instance.pluginClient.Kill()
	delete(pm.loadedPlugins, pluginName)
	
	pm.logger.Info("Unloaded plugin", "name", pluginName)
	return nil
}

// GetPlugin gets a loaded cloud provider plugin
func (pm *PluginManagerImpl) GetPlugin(pluginName string) (CloudProvider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	instance, ok := pm.loadedPlugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("plugin %s not loaded", pluginName)
	}

	return instance.cloudProvider, nil
}

// ListPlugins lists all loaded plugins
func (pm *PluginManagerImpl) ListPlugins() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]string, 0, len(pm.loadedPlugins))
	for name := range pm.loadedPlugins {
		plugins = append(plugins, name)
	}

	return plugins
}

// DiscoverPlugins discovers plugins in the plugins directory
func (pm *PluginManagerImpl) DiscoverPlugins() ([]string, error) {
	// Ensure plugins directory exists
	if _, err := os.Stat(pm.pluginsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugins directory %s does not exist", pm.pluginsDir)
	}

	// Read all files in the plugins directory
	files, err := ioutil.ReadDir(pm.pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading plugins directory %s: %w", pm.pluginsDir, err)
	}

	// Find executable files
	var plugins []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if the file is executable
		filePath := filepath.Join(pm.pluginsDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			pm.logger.Warn("Error accessing file", "file", filePath, "error", err)
			continue
		}

		// On Unix-like systems, check executable bit
		if fileInfo.Mode()&0111 != 0 {
			plugins = append(plugins, file.Name())
		}
	}

	return plugins, nil
}