package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/core"
	"github.com/scttfrdmn/snoozebot/pkg/plugin"
)

func main() {
	// Parse command line flags
	pluginsDir := flag.String("plugins-dir", "/etc/snoozebot/plugins", "Directory containing plugins")
	configFile := flag.String("config", "/etc/snoozebot/config.json", "Path to configuration file")
	logLevel := flag.String("log-level", "info", "Log level (trace, debug, info, warn, error)")
	flag.Parse()

	// Set up logger
	level := hclog.LevelFromString(*logLevel)
	if level == hclog.NoLevel {
		level = hclog.Info
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "snoozed",
		Level:  level,
		Output: os.Stdout,
	})

	logger.Info("Starting snoozed daemon", "version", "1.0.0")

	// Create the plugin manager
	pluginManager := core.NewPluginManager(*pluginsDir, logger.Named("plugin-manager"))

	// Discover available plugins
	availablePlugins, err := pluginManager.DiscoverPlugins()
	if err != nil {
		logger.Error("Failed to discover plugins", "error", err)
	} else {
		logger.Info("Discovered plugins", "count", len(availablePlugins), "plugins", availablePlugins)
	}

	// Create a default monitor configuration
	monitorConfig := core.DefaultMonitorConfig()

	// Override with values from config file if it exists
	// In a real implementation, we would parse the config file here
	logger.Info("Using configuration", "napTime", monitorConfig.NapTime, "checkInterval", monitorConfig.CheckInterval)

	// Create the resource monitor
	monitor := core.NewLinuxResourceMonitor(monitorConfig)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the resource monitor
	if err := monitor.Start(ctx); err != nil {
		logger.Error("Failed to start resource monitor", "error", err)
		os.Exit(1)
	}
	defer monitor.Stop()

	// Load the AWS plugin if available
	awsPluginName := "aws"
	if contains(availablePlugins, awsPluginName) {
		logger.Info("Loading AWS plugin")
		provider, err := pluginManager.LoadPlugin(ctx, awsPluginName)
		if err != nil {
			logger.Error("Failed to load AWS plugin", "error", err)
		} else {
			// Get provider info
			name := provider.GetProviderName()
			version := provider.GetProviderVersion()
			logger.Info("Loaded cloud provider plugin", "name", name, "version", version)

			// Get instance info
			instanceInfo, err := provider.GetInstanceInfo(ctx)
			if err != nil {
				logger.Error("Failed to get instance info", "error", err)
			} else {
				logger.Info("Instance info", 
					"id", instanceInfo.ID,
					"name", instanceInfo.Name,
					"type", instanceInfo.Type,
					"region", instanceInfo.Region,
					"zone", instanceInfo.Zone,
					"state", instanceInfo.State,
					"launchTime", instanceInfo.LaunchTime)
			}
		}
	} else {
		logger.Warn("AWS plugin not found")
	}

	// Set up signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigs
	logger.Info("Received signal, shutting down", "signal", sig)

	// Clean up
	logger.Info("Cleaning up")
	pluginManager.Cleanup()
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}