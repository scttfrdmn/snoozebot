package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// handleListPlugins handles listing all loaded plugins
func (s *Server) handleListPlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get list of loaded plugins
	plugins := s.pluginManager.ListPlugins()

	// Build response with more details about each plugin
	type PluginInfo struct {
		Name     string `json:"name"`
		Provider string `json:"provider"`
		Version  string `json:"version"`
		Status   string `json:"status"`
	}

	var response struct {
		Plugins []PluginInfo `json:"plugins"`
		Count   int          `json:"count"`
	}

	for _, pluginName := range plugins {
		providerPlugin, err := s.pluginManager.GetPlugin(pluginName)
		if err != nil {
			s.logger.Error("Error getting plugin info", "plugin", pluginName, "error", err)
			response.Plugins = append(response.Plugins, PluginInfo{
				Name:   pluginName,
				Status: "error",
			})
			continue
		}

		response.Plugins = append(response.Plugins, PluginInfo{
			Name:     pluginName,
			Provider: providerPlugin.GetProviderName(),
			Version:  providerPlugin.GetProviderVersion(),
			Status:   "active",
		})
	}

	response.Count = len(response.Plugins)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDiscoverPlugins handles discovering plugins in the plugins directory
func (s *Server) handleDiscoverPlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get auto-load parameter
	autoLoad := r.URL.Query().Get("autoload") == "true"

	// Discover plugins
	plugins, err := s.pluginManager.DiscoverPlugins()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to discover plugins: %v", err), http.StatusInternalServerError)
		return
	}

	// Get list of already loaded plugins
	loadedPlugins := s.pluginManager.ListPlugins()
	loadedMap := make(map[string]bool)
	for _, name := range loadedPlugins {
		loadedMap[name] = true
	}

	// Build response
	type PluginInfo struct {
		Name     string `json:"name"`
		Loaded   bool   `json:"loaded"`
		Provider string `json:"provider,omitempty"`
		Version  string `json:"version,omitempty"`
	}

	var response struct {
		Plugins []PluginInfo `json:"plugins"`
		Count   int          `json:"count"`
		Directory string     `json:"directory"`
	}

	response.Directory = s.pluginsDir
	
	// Process plugins
	for _, name := range plugins {
		info := PluginInfo{
			Name: name,
			Loaded: loadedMap[name],
		}
		
		// If auto-load is true and plugin is not loaded, try to load it
		if autoLoad && !info.Loaded {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			provider, err := s.pluginManager.LoadPlugin(ctx, name)
			cancel()
			
			if err != nil {
				s.logger.Error("Failed to auto-load plugin", "name", name, "error", err)
			} else {
				info.Loaded = true
				info.Provider = provider.GetProviderName()
				info.Version = provider.GetProviderVersion()
				s.logger.Info("Auto-loaded plugin", "name", name, "provider", info.Provider)
			}
		}
		
		// If plugin is loaded, get provider info
		if info.Loaded && (info.Provider == "" || info.Version == "") {
			provider, err := s.pluginManager.GetPlugin(name)
			if err == nil {
				info.Provider = provider.GetProviderName()
				info.Version = provider.GetProviderVersion()
			}
		}
		
		response.Plugins = append(response.Plugins, info)
	}
	
	response.Count = len(response.Plugins)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLoadPlugin handles loading a plugin
func (s *Server) handleLoadPlugin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var request struct {
		PluginName string `json:"plugin_name"`
		ApiKey     string `json:"api_key,omitempty"`
		Timeout    int    `json:"timeout_seconds,omitempty"`
		Retries    int    `json:"retries,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if request.PluginName == "" {
		http.Error(w, "Plugin name is required", http.StatusBadRequest)
		return
	}
	
	// Check if authentication is required
	authRequired := s.authenticatedManager != nil && s.authenticatedManager.IsAuthenticationEnabled()
	if authRequired && request.ApiKey == "" {
		http.Error(w, "API key is required when authentication is enabled", http.StatusUnauthorized)
		return
	}

	// Set defaults if not provided
	if request.Timeout <= 0 {
		request.Timeout = 30 // Default 30 seconds
	}
	if request.Retries <= 0 {
		request.Retries = 1 // Default 1 try (no retries)
	}

	// Check if plugin is already loaded
	loadedPlugins := s.pluginManager.ListPlugins()
	for _, name := range loadedPlugins {
		if name == request.PluginName {
			// Plugin is already loaded, get info and return success
			provider, err := s.pluginManager.GetPlugin(request.PluginName)
			if err != nil {
				http.Error(w, fmt.Sprintf("Plugin is loaded but info unavailable: %v", err), http.StatusInternalServerError)
				return
			}
			
			response := map[string]interface{}{
				"success":          true,
				"plugin_name":      request.PluginName,
				"provider_name":    provider.GetProviderName(),
				"provider_version": provider.GetProviderVersion(),
				"already_loaded":   true,
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Create context with timeout for loading plugin
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(request.Timeout)*time.Second)
	defer cancel()

	// Try to load the plugin with retries if configured
	var cp provider.CloudProvider
	var err error
	var loadAttempt int
	
	for loadAttempt = 1; loadAttempt <= request.Retries; loadAttempt++ {
		s.logger.Info("Loading plugin", 
			"name", request.PluginName, 
			"attempt", loadAttempt, 
			"max_attempts", request.Retries,
			"auth_required", authRequired)
		
		// Use the appropriate plugin manager based on authentication requirements
		if authRequired && s.authenticatedManager != nil {
			cp, err = s.authenticatedManager.LoadPlugin(ctx, request.PluginName)
			if err == nil {
				// Successfully loaded, now authenticate if needed
				if authProvider, ok := cp.(*provider.AuthenticatedProvider); ok {
					var authSuccess bool
					authSuccess, err = authProvider.Authenticate(ctx, request.ApiKey)
					if !authSuccess {
						err = fmt.Errorf("authentication failed")
					}
				}
			}
		} else {
			// Use the standard plugin manager
			cp, err = s.pluginManager.LoadPlugin(ctx, request.PluginName)
		}
		
		if err == nil {
			// Successfully loaded
			break
		}
		
		s.logger.Error("Failed to load plugin", 
			"name", request.PluginName, 
			"attempt", loadAttempt, 
			"error", err)
		
		// If this was the last attempt or context is done, exit loop
		if loadAttempt >= request.Retries || ctx.Err() != nil {
			break
		}
		
		// Wait for a short time before retrying
		time.Sleep(500 * time.Millisecond)
	}
	
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load plugin after %d attempts: %v", loadAttempt, err), http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"success":          true,
		"plugin_name":      request.PluginName,
		"provider_name":    cp.GetProviderName(),
		"provider_version": cp.GetProviderVersion(),
		"attempts":         loadAttempt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUnloadPlugin handles unloading a plugin
func (s *Server) handleUnloadPlugin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var request struct {
		PluginName string `json:"plugin_name"`
		Force      bool   `json:"force,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if request.PluginName == "" {
		http.Error(w, "Plugin name is required", http.StatusBadRequest)
		return
	}

	// Check if plugin is loaded
	isLoaded := false
	loadedPlugins := s.pluginManager.ListPlugins()
	for _, name := range loadedPlugins {
		if name == request.PluginName {
			isLoaded = true
			break
		}
	}

	if !isLoaded {
		// If plugin is not loaded, return success if force is true, error otherwise
		if request.Force {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     true,
				"plugin_name": request.PluginName,
				"status":      "not_loaded",
				"message":     "Plugin was not loaded, nothing to unload",
			})
			return
		} else {
			http.Error(w, fmt.Sprintf("Plugin '%s' is not loaded", request.PluginName), http.StatusBadRequest)
			return
		}
	}

	// Get plugin info before unloading
	var providerName, providerVersion string
	provider, err := s.pluginManager.GetPlugin(request.PluginName)
	if err == nil {
		providerName = provider.GetProviderName()
		providerVersion = provider.GetProviderVersion()
	}

	// Unload plugin
	start := time.Now()
	err = s.pluginManager.UnloadPlugin(request.PluginName)
	duration := time.Since(start)
	
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to unload plugin: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"success":          true,
		"plugin_name":      request.PluginName,
		"provider_name":    providerName,
		"provider_version": providerVersion,
		"status":           "unloaded",
		"duration_ms":      duration.Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}