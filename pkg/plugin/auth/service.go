package auth

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	snoozePlugin "github.com/scttfrdmn/snoozebot/pkg/plugin"
	"google.golang.org/grpc"
)

// AuthPlugin is the implementation of plugin.Plugin for the auth plugin
type AuthPlugin struct {
	Impl   PluginAuthService
	Logger hclog.Logger
}

// PluginAuthService is the interface that plugins must implement for authentication
type PluginAuthService interface {
	Authenticate(ctx context.Context, pluginName, apiKey string) (bool, string, error)
	CheckPermission(ctx context.Context, pluginName, permission string) (bool, error)
}

// PluginAuthServiceImpl is the implementation of PluginAuthService
type PluginAuthServiceImpl struct {
	apiKeyManager *APIKeyManager
	logger        hclog.Logger
}

// NewPluginAuthService creates a new PluginAuthService
func NewPluginAuthService(configDir string, logger hclog.Logger) (PluginAuthService, error) {
	// Create the API key manager
	configPath := filepath.Join(configDir, "security.json")
	apiKeyManager, err := NewAPIKeyManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key manager: %w", err)
	}

	return &PluginAuthServiceImpl{
		apiKeyManager: apiKeyManager,
		logger:        logger,
	}, nil
}

// Authenticate authenticates a plugin
func (s *PluginAuthServiceImpl) Authenticate(ctx context.Context, pluginName, apiKey string) (bool, string, error) {
	s.logger.Debug("Authenticating plugin", "plugin", pluginName)
	return s.apiKeyManager.ValidateAPIKey(pluginName, apiKey)
}

// CheckPermission checks if a plugin has a specific permission
func (s *PluginAuthServiceImpl) CheckPermission(ctx context.Context, pluginName, permission string) (bool, error) {
	s.logger.Debug("Checking permission", "plugin", pluginName, "permission", permission)
	return s.apiKeyManager.CheckPermission(pluginName, permission)
}

// GRPCServer registers the auth service with the gRPC server
func (p *AuthPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	snoozePlugin.RegisterPluginAuthServer(s, &GRPCPluginAuthServer{
		Impl: p.Impl,
	})
	return nil
}

// GRPCClient returns the auth client
func (p *AuthPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCPluginAuthClient{
		client: snoozePlugin.NewPluginAuthClient(c),
		logger: p.Logger,
	}, nil
}

// GRPCPluginAuthServer is the gRPC server for PluginAuthService
type GRPCPluginAuthServer struct {
	Impl PluginAuthService
	snoozePlugin.UnimplementedPluginAuthServer
}

// Authenticate authenticates a plugin
func (s *GRPCPluginAuthServer) Authenticate(ctx context.Context, req *snoozePlugin.AuthenticateRequest) (*snoozePlugin.AuthenticateResponse, error) {
	success, role, err := s.Impl.Authenticate(ctx, req.PluginName, req.ApiKey)
	if err != nil {
		return &snoozePlugin.AuthenticateResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &snoozePlugin.AuthenticateResponse{
		Success: success,
		Role:    role,
	}, nil
}

// CheckPermission checks if a plugin has a specific permission
func (s *GRPCPluginAuthServer) CheckPermission(ctx context.Context, req *snoozePlugin.PermissionRequest) (*snoozePlugin.PermissionResponse, error) {
	allowed, err := s.Impl.CheckPermission(ctx, req.PluginName, req.Permission)
	if err != nil {
		return &snoozePlugin.PermissionResponse{
			Allowed:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &snoozePlugin.PermissionResponse{
		Allowed: allowed,
	}, nil
}

// GRPCPluginAuthClient is the gRPC client for PluginAuthService
type GRPCPluginAuthClient struct {
	client snoozePlugin.PluginAuthClient
	logger hclog.Logger
}

// Authenticate authenticates a plugin
func (c *GRPCPluginAuthClient) Authenticate(ctx context.Context, pluginName, apiKey string) (bool, string, error) {
	resp, err := c.client.Authenticate(ctx, &snoozePlugin.AuthenticateRequest{
		PluginName: pluginName,
		ApiKey:     apiKey,
	})
	if err != nil {
		return false, "", err
	}

	if !resp.Success && resp.ErrorMessage != "" {
		return false, "", fmt.Errorf(resp.ErrorMessage)
	}

	return resp.Success, resp.Role, nil
}

// CheckPermission checks if a plugin has a specific permission
func (c *GRPCPluginAuthClient) CheckPermission(ctx context.Context, pluginName, permission string) (bool, error) {
	resp, err := c.client.CheckPermission(ctx, &snoozePlugin.PermissionRequest{
		PluginName: pluginName,
		Permission: permission,
	})
	if err != nil {
		return false, err
	}

	if !resp.Allowed && resp.ErrorMessage != "" {
		return false, fmt.Errorf(resp.ErrorMessage)
	}

	return resp.Allowed, nil
}