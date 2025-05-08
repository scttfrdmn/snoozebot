package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AuthenticationManager returns the authenticated plugin manager
func (s *Server) AuthenticationManager() interface{} {
	return s.authenticatedManager
}

// handleAuthStatus handles the GET /api/auth/status endpoint
func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if authentication is enabled
	enabled := false
	if s.authenticatedManager != nil {
		enabled = s.authenticatedManager.IsAuthenticationEnabled()
	}

	// Return the status
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled": enabled,
	})
}

// handleEnableAuth handles the POST /api/auth/enable endpoint
func (s *Server) handleEnableAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if authentication manager is available
	if s.authenticatedManager == nil {
		http.Error(w, "Authentication not available", http.StatusInternalServerError)
		return
	}

	// Enable authentication
	s.authenticatedManager.EnableAuthentication(true)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled": true,
	})
}

// handleDisableAuth handles the POST /api/auth/disable endpoint
func (s *Server) handleDisableAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if authentication manager is available
	if s.authenticatedManager == nil {
		http.Error(w, "Authentication not available", http.StatusInternalServerError)
		return
	}

	// Disable authentication
	s.authenticatedManager.EnableAuthentication(false)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled": false,
	})
}

// APIKeyRequest is used for generating API keys
type APIKeyRequest struct {
	PluginName    string `json:"plugin_name"`
	Role          string `json:"role"`
	Description   string `json:"description"`
	ExpiresInDays int    `json:"expires_in_days"`
}

// handleGenerateAPIKey handles the POST /api/auth/apikey endpoint
func (s *Server) handleGenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if authentication manager is available
	if s.authenticatedManager == nil {
		http.Error(w, "Authentication not available", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req APIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PluginName == "" {
		http.Error(w, "Plugin name is required", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		http.Error(w, "Role is required", http.StatusBadRequest)
		return
	}

	// Set default expiration if not provided
	if req.ExpiresInDays <= 0 {
		req.ExpiresInDays = 365 // Default to 1 year
	}

	// Generate API key
	apiKey, err := s.authenticatedManager.GenerateAPIKey(req.PluginName, req.Role, req.Description, req.ExpiresInDays)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the API key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plugin_name":     req.PluginName,
		"api_key":         apiKey,
		"role":            req.Role,
		"description":     req.Description,
		"expires_in_days": req.ExpiresInDays,
		"created_at":      time.Now().Format(time.RFC3339),
		"expires_at":      time.Now().AddDate(0, 0, req.ExpiresInDays).Format(time.RFC3339),
	})
}

// handleRevokeAPIKey handles the POST /api/auth/apikey/revoke endpoint
func (s *Server) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if authentication manager is available
	if s.authenticatedManager == nil {
		http.Error(w, "Authentication not available", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req APIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PluginName == "" {
		http.Error(w, "Plugin name is required", http.StatusBadRequest)
		return
	}

	// Revoke API key
	err := s.authenticatedManager.RevokeAPIKey(req.PluginName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to revoke API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plugin_name": req.PluginName,
		"success":     true,
	})
}