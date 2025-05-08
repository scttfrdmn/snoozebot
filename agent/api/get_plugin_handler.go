package api

import (
	"fmt"
	"net/http"
	"strings"
)

// handleGetPluginInfo handles getting information about a specific plugin
func (s *Server) handleGetPluginInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract plugin name from path
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/plugins/") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	pluginName := strings.TrimPrefix(path, "/api/plugins/")
	if pluginName == "" {
		// If no plugin name is provided, redirect to the list endpoint
		http.Redirect(w, r, "/api/plugins", http.StatusFound)
		return
	}

	// Check if the plugin is loaded
	loadedPlugins := s.pluginManager.ListPlugins()
	isLoaded := false
	for _, name := range loadedPlugins {
		if name == pluginName {
			isLoaded = true
			break
		}
	}

	if !isLoaded {
		http.Error(w, fmt.Sprintf("Plugin '%s' is not loaded", pluginName), http.StatusNotFound)
		return
	}

	// Get plugin information
	provider, err := s.pluginManager.GetPlugin(pluginName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get plugin info: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := map[string]interface{}{
		"name":     pluginName,
		"provider": provider.GetProviderName(),
		"version":  provider.GetProviderVersion(),
		"status":   "active",
	}

	// Add plugin path
	pluginPath := fmt.Sprintf("%s/%s", s.pluginsDir, pluginName)
	response["path"] = pluginPath

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}