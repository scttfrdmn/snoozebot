package authtest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/auth"
)

// TestAPIKeyGeneration tests API key generation and validation
func TestAPIKeyGeneration(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "snoozebot-auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the API key manager
	configPath := filepath.Join(tempDir, "security.json")
	apiKeyManager, err := auth.NewAPIKeyManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create API key manager: %v", err)
	}

	// Check if auth is enabled (should be false by default)
	if apiKeyManager.IsAuthEnabled() {
		t.Errorf("Expected auth to be disabled by default")
	}

	// Enable auth
	err = apiKeyManager.EnableAuth()
	if err != nil {
		t.Fatalf("Failed to enable auth: %v", err)
	}

	// Check if auth is enabled
	if !apiKeyManager.IsAuthEnabled() {
		t.Errorf("Expected auth to be enabled")
	}

	// Generate an API key
	apiKey, err := apiKeyManager.GenerateAPIKey("test-plugin", "cloud_provider", "Test API key", 365)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if apiKey == "" {
		t.Errorf("Expected non-empty API key")
	}

	// Validate the API key
	valid, role, err := apiKeyManager.ValidateAPIKey("test-plugin", apiKey)
	if err != nil {
		t.Fatalf("Failed to validate API key: %v", err)
	}

	if !valid {
		t.Errorf("Expected API key to be valid")
	}

	if role != "cloud_provider" {
		t.Errorf("Expected role to be 'cloud_provider', got %s", role)
	}

	// Check permission
	allowed, err := apiKeyManager.CheckPermission("test-plugin", "cloud_operations")
	if err != nil {
		t.Fatalf("Failed to check permission: %v", err)
	}

	if !allowed {
		t.Errorf("Expected cloud_operations permission to be allowed")
	}

	// Revoke the API key
	err = apiKeyManager.RevokeAPIKey("test-plugin")
	if err != nil {
		t.Fatalf("Failed to revoke API key: %v", err)
	}

	// Try to validate the revoked API key
	valid, _, err = apiKeyManager.ValidateAPIKey("test-plugin", apiKey)
	if err == nil {
		t.Errorf("Expected error for revoked API key, got nil")
	}

	if valid {
		t.Errorf("Expected revoked API key to be invalid")
	}
}

// TestAuthService tests the authentication service
func TestAuthService(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "snoozebot-auth-service-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a logger
	logger := hclog.NewNullLogger()

	// Create the auth service
	authService, err := auth.NewPluginAuthService(tempDir, logger)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	// Create the API key manager
	configPath := filepath.Join(tempDir, "security.json")
	apiKeyManager, err := auth.NewAPIKeyManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create API key manager: %v", err)
	}

	// Enable auth
	err = apiKeyManager.EnableAuth()
	if err != nil {
		t.Fatalf("Failed to enable auth: %v", err)
	}

	// Generate an API key
	apiKey, err := apiKeyManager.GenerateAPIKey("test-plugin", "cloud_provider", "Test API key", 365)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// Authenticate using the auth service
	ctx := context.Background()
	success, role, err := authService.Authenticate(ctx, "test-plugin", apiKey)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	if !success {
		t.Errorf("Expected authentication to succeed")
	}

	if role != "cloud_provider" {
		t.Errorf("Expected role to be 'cloud_provider', got %s", role)
	}

	// Check permission using the auth service
	allowed, err := authService.CheckPermission(ctx, "test-plugin", "cloud_operations")
	if err != nil {
		t.Fatalf("Failed to check permission: %v", err)
	}

	if !allowed {
		t.Errorf("Expected cloud_operations permission to be allowed")
	}
}