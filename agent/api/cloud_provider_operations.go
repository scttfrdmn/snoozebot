package api

import (
	"context"
	"fmt"
	"time"

	"github.com/scttfrdmn/snoozebot/agent/provider"
	"github.com/scttfrdmn/snoozebot/pkg/common/protocol/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetInstanceInfo gets information about a cloud instance
func (s *GRPCServer) GetInstanceInfo(ctx context.Context, req *gen.GetInstanceInfoRequest) (*gen.GetInstanceInfoResponse, error) {
	// First determine which cloud provider to use
	instance, err := s.instanceStore.GetInstance(req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Get the plugin for the provider
	pluginName := instance.Registration.Provider
	plugin, err := s.pluginManager.GetPlugin(pluginName)
	if err != nil {
		// Try to load the plugin if it's not loaded
		plugin, err = s.pluginManager.LoadPlugin(ctx, pluginName)
		if err != nil {
			return nil, fmt.Errorf("failed to load cloud provider plugin %s: %w", pluginName, err)
		}
	}

	// Get instance info from the provider
	info, err := plugin.GetInstanceInfo(ctx, req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance info: %w", err)
	}

	// Convert to response format
	return &gen.GetInstanceInfoResponse{
		Id:         info.ID,
		Name:       info.Name,
		Type:       info.Type,
		Region:     info.Region,
		Zone:       info.Zone,
		State:      info.State,
		LaunchTime: timestamppb.New(info.LaunchTime),
		Provider:   info.Provider,
	}, nil
}

// StopInstance stops a cloud instance
func (s *GRPCServer) StopInstance(ctx context.Context, req *gen.StopInstanceRequest) (*gen.StopInstanceResponse, error) {
	// First determine which cloud provider to use
	instance, err := s.instanceStore.GetInstance(req.InstanceId)
	if err != nil {
		return &gen.StopInstanceResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get instance: %v", err),
		}, nil
	}

	// Get the plugin for the provider
	pluginName := instance.Registration.Provider
	plugin, err := s.pluginManager.GetPlugin(pluginName)
	if err != nil {
		// Try to load the plugin if it's not loaded
		plugin, err = s.pluginManager.LoadPlugin(ctx, pluginName)
		if err != nil {
			return &gen.StopInstanceResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to load cloud provider plugin %s: %v", pluginName, err),
			}, nil
		}
	}

	// Stop the instance
	err = plugin.StopInstance(ctx, req.InstanceId)
	if err != nil {
		return &gen.StopInstanceResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to stop instance: %v", err),
		}, nil
	}

	// Update instance state in store
	err = s.instanceStore.UpdateInstanceState(req.InstanceId, "stopping")
	if err != nil {
		// Log the error but don't fail the operation
		fmt.Printf("Warning: failed to update instance state: %v\n", err)
	}

	return &gen.StopInstanceResponse{
		Success: true,
	}, nil
}

// StartInstance starts a cloud instance
func (s *GRPCServer) StartInstance(ctx context.Context, req *gen.StartInstanceRequest) (*gen.StartInstanceResponse, error) {
	// First determine which cloud provider to use
	instance, err := s.instanceStore.GetInstance(req.InstanceId)
	if err != nil {
		return &gen.StartInstanceResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get instance: %v", err),
		}, nil
	}

	// Get the plugin for the provider
	pluginName := instance.Registration.Provider
	plugin, err := s.pluginManager.GetPlugin(pluginName)
	if err != nil {
		// Try to load the plugin if it's not loaded
		plugin, err = s.pluginManager.LoadPlugin(ctx, pluginName)
		if err != nil {
			return &gen.StartInstanceResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to load cloud provider plugin %s: %v", pluginName, err),
			}, nil
		}
	}

	// Start the instance
	err = plugin.StartInstance(ctx, req.InstanceId)
	if err != nil {
		return &gen.StartInstanceResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to start instance: %v", err),
		}, nil
	}

	// Update instance state in store
	err = s.instanceStore.UpdateInstanceState(req.InstanceId, "starting")
	if err != nil {
		// Log the error but don't fail the operation
		fmt.Printf("Warning: failed to update instance state: %v\n", err)
	}

	return &gen.StartInstanceResponse{
		Success: true,
	}, nil
}

// PerformCloudAction performs a cloud provider action
func (s *GRPCServer) PerformCloudAction(ctx context.Context, req *gen.CloudActionRequest) (*gen.CloudActionResponse, error) {
	// First determine which cloud provider to use
	instance, err := s.instanceStore.GetInstance(req.InstanceId)
	if err != nil {
		return &gen.CloudActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get instance: %v", err),
		}, nil
	}

	// Handle built-in actions
	switch req.Action {
	case "stop":
		return s.StopInstance(ctx, &gen.StopInstanceRequest{
			InstanceId: req.InstanceId,
		})
	case "start":
		return s.StartInstance(ctx, &gen.StartInstanceRequest{
			InstanceId: req.InstanceId,
		})
	}

	// For other actions, we need to check if the provider supports them
	// This would require extending the provider interface to support custom actions
	return &gen.CloudActionResponse{
		Success: false,
		Error:   fmt.Sprintf("Unsupported action: %s", req.Action),
	}, nil
}

// ListCloudProviders lists all loaded cloud providers
func (s *GRPCServer) ListCloudProviders(ctx context.Context, req *gen.ListCloudProvidersRequest) (*gen.ListCloudProvidersResponse, error) {
	// Get list of loaded plugins
	plugins := s.pluginManager.ListPlugins()

	providers := make([]*gen.CloudProviderInfo, 0, len(plugins))
	for _, pluginName := range plugins {
		provider, err := s.pluginManager.GetPlugin(pluginName)
		if err != nil {
			// Skip plugins that can't be loaded
			continue
		}

		providers = append(providers, &gen.CloudProviderInfo{
			Name:    provider.GetProviderName(),
			Version: provider.GetProviderVersion(),
			Plugin:  pluginName,
		})
	}

	return &gen.ListCloudProvidersResponse{
		Providers: providers,
	}, nil
}