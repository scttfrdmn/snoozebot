package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-hclog"
)

// AuthenticatedPluginManager is the interface for plugin manager with authentication
type AuthenticatedPluginManager interface {
	EnableAuthentication(enabled bool)
	IsAuthenticationEnabled() bool
	GenerateAPIKey(pluginName, role, description string, expiresInDays int) (string, error)
	RevokeAPIKey(pluginName string) error
}

// AuthHandlers contains all handlers for plugin authentication
type AuthHandlers struct {
	pluginManager AuthenticatedPluginManager
	logger        hclog.Logger
}

// NewAuthHandlers creates a new AuthHandlers
func NewAuthHandlers(pluginManager AuthenticatedPluginManager, logger hclog.Logger) *AuthHandlers {
	return &AuthHandlers{
		pluginManager: pluginManager,
		logger:        logger,
	}
}

// RegisterRoutes registers all auth routes
func (h *AuthHandlers) RegisterRoutes(router *gin.Engine) {
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.GET("/status", h.getAuthStatus)
		authGroup.POST("/enable", h.enableAuth)
		authGroup.POST("/disable", h.disableAuth)
		authGroup.POST("/apikey", h.generateAPIKey)
		authGroup.DELETE("/apikey/:pluginName", h.revokeAPIKey)
	}
}

// AuthStatusResponse is the response for the auth status endpoint
type AuthStatusResponse struct {
	Enabled bool `json:"enabled"`
}

// getAuthStatus returns the current authentication status
func (h *AuthHandlers) getAuthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, AuthStatusResponse{
		Enabled: h.pluginManager.IsAuthenticationEnabled(),
	})
}

// enableAuth enables authentication
func (h *AuthHandlers) enableAuth(c *gin.Context) {
	h.pluginManager.EnableAuthentication(true)
	c.JSON(http.StatusOK, AuthStatusResponse{
		Enabled: true,
	})
}

// disableAuth disables authentication
func (h *AuthHandlers) disableAuth(c *gin.Context) {
	h.pluginManager.EnableAuthentication(false)
	c.JSON(http.StatusOK, AuthStatusResponse{
		Enabled: false,
	})
}

// GenerateAPIKeyRequest is the request for the generate API key endpoint
type GenerateAPIKeyRequest struct {
	PluginName    string `json:"plugin_name" binding:"required"`
	Role          string `json:"role" binding:"required"`
	Description   string `json:"description"`
	ExpiresInDays int    `json:"expires_in_days"`
}

// GenerateAPIKeyResponse is the response for the generate API key endpoint
type GenerateAPIKeyResponse struct {
	PluginName    string    `json:"plugin_name"`
	APIKey        string    `json:"api_key"`
	Role          string    `json:"role"`
	Description   string    `json:"description"`
	ExpiresInDays int       `json:"expires_in_days"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// generateAPIKey generates an API key for a plugin
func (h *AuthHandlers) generateAPIKey(c *gin.Context) {
	var req GenerateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default expiration if not provided
	if req.ExpiresInDays <= 0 {
		req.ExpiresInDays = 365 // Default to 1 year
	}

	// Generate the API key
	apiKey, err := h.pluginManager.GenerateAPIKey(req.PluginName, req.Role, req.Description, req.ExpiresInDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate API key: %v", err)})
		return
	}

	// Return the API key
	now := time.Now()
	c.JSON(http.StatusOK, GenerateAPIKeyResponse{
		PluginName:    req.PluginName,
		APIKey:        apiKey,
		Role:          req.Role,
		Description:   req.Description,
		ExpiresInDays: req.ExpiresInDays,
		CreatedAt:     now,
		ExpiresAt:     now.AddDate(0, 0, req.ExpiresInDays),
	})
}

// RevokeAPIKeyRequest is the request for the revoke API key endpoint
type RevokeAPIKeyRequest struct {
	PluginName string `json:"plugin_name" binding:"required"`
}

// RevokeAPIKeyResponse is the response for the revoke API key endpoint
type RevokeAPIKeyResponse struct {
	PluginName string `json:"plugin_name"`
	Success    bool   `json:"success"`
}

// revokeAPIKey revokes an API key for a plugin
func (h *AuthHandlers) revokeAPIKey(c *gin.Context) {
	pluginName := c.Param("pluginName")
	if pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin name is required"})
		return
	}

	// Revoke the API key
	err := h.pluginManager.RevokeAPIKey(pluginName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to revoke API key: %v", err)})
		return
	}

	// Return success
	c.JSON(http.StatusOK, RevokeAPIKeyResponse{
		PluginName: pluginName,
		Success:    true,
	})
}