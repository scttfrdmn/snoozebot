package api

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

// DiscoverAndInitPlugins discovers plugins and initializes them
func (s *Server) DiscoverAndInitPlugins(ctx context.Context) error {
	s.logger.Info("Discovering plugins", "dir", s.pluginsDir)
	
	// Discover available plugins
	plugins, err := s.pluginManager.DiscoverPlugins()
	if err != nil {
		return fmt.Errorf("failed to discover plugins: %w", err)
	}
	
	s.logger.Info("Discovered plugins", "count", len(plugins), "plugins", plugins)
	
	// For each plugin, try to load it
	for _, pluginName := range plugins {
		pluginPath := filepath.Join(s.pluginsDir, pluginName)
		s.logger.Info("Loading plugin", "name", pluginName, "path", pluginPath)
		
		// Create a context with timeout for loading the plugin
		loadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		
		// Try to load the plugin
		provider, err := s.pluginManager.LoadPlugin(loadCtx, pluginName)
		cancel() // Cancel the timeout context
		
		if err != nil {
			s.logger.Error("Failed to load plugin", "name", pluginName, "error", err)
			continue
		}
		
		s.logger.Info("Successfully loaded plugin", 
			"name", pluginName, 
			"provider", provider.GetProviderName(), 
			"version", provider.GetProviderVersion())
	}
	
	return nil
}

// GetPluginInfo returns detailed information about a plugin
func (s *Server) GetPluginInfo(pluginName string) (map[string]interface{}, error) {
	provider, err := s.pluginManager.GetPlugin(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	
	return map[string]interface{}{
		"name":     pluginName,
		"provider": provider.GetProviderName(),
		"version":  provider.GetProviderVersion(),
	}, nil
}

// LoadPluginWithRetry attempts to load a plugin with retries
func (s *Server) LoadPluginWithRetry(ctx context.Context, pluginName string, maxRetries int, delay time.Duration) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		// Try to load the plugin
		_, err := s.pluginManager.LoadPlugin(ctx, pluginName)
		if err == nil {
			// Successfully loaded
			return nil
		}
		
		lastErr = err
		s.logger.Warn("Failed to load plugin, retrying", 
			"name", pluginName, 
			"attempt", i+1, 
			"maxRetries", maxRetries, 
			"error", err)
		
		// Wait before retrying
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue with retry
		}
	}
	
	return fmt.Errorf("failed to load plugin after %d attempts: %w", maxRetries, lastErr)
}