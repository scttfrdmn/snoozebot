package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/scottfridman/snoozebot/agent/provider"
)

// AWSProvider is an implementation of CloudProvider for AWS
type AWSProvider struct {
	logger hclog.Logger
}

// GetInstanceInfo gets information about an instance
func (p *AWSProvider) GetInstanceInfo(ctx context.Context, instanceID string) (*provider.InstanceInfo, error) {
	p.logger.Info("Getting instance info", "instanceID", instanceID)
	
	// In a real implementation, this would use the AWS SDK to get instance information
	// For this example, we'll just return dummy data
	return &provider.InstanceInfo{
		ID:         instanceID,
		Name:       fmt.Sprintf("instance-%s", instanceID),
		Type:       "t3.micro",
		Region:     "us-west-2",
		Zone:       "us-west-2a",
		State:      "running",
		LaunchTime: time.Now().Add(-24 * time.Hour),
		Provider:   "aws",
	}, nil
}

// StopInstance stops an instance
func (p *AWSProvider) StopInstance(ctx context.Context, instanceID string) error {
	p.logger.Info("Stopping instance", "instanceID", instanceID)
	
	// In a real implementation, this would use the AWS SDK to stop the instance
	return nil
}

// StartInstance starts an instance
func (p *AWSProvider) StartInstance(ctx context.Context, instanceID string) error {
	p.logger.Info("Starting instance", "instanceID", instanceID)
	
	// In a real implementation, this would use the AWS SDK to start the instance
	return nil
}

// GetProviderName returns the name of the cloud provider
func (p *AWSProvider) GetProviderName() string {
	return "aws"
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *AWSProvider) GetProviderVersion() string {
	return "1.0.0"
}

// CloudProviderPlugin is the implementation of plugin.Plugin
type CloudProviderPlugin struct {
	// Impl is the concrete implementation of CloudProvider
	Impl provider.CloudProvider
}

// GRPCServer registers this plugin for serving with a gRPC server
func (p *CloudProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *plugin.DefaultGRPCServer) error {
	// In a real implementation, we would register the gRPC server
	// For this example, we'll just return nil
	return nil
}

// GRPCClient returns the client for this plugin
func (p *CloudProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *plugin.DefaultGRPCClient) (interface{}, error) {
	// In a real implementation, we would return a gRPC client
	// For this example, we'll just return nil
	return nil, nil
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user-friendly error is shown.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SNOOZEBOT_PLUGIN",
	MagicCookieValue: "snoozebot_provider_v1",
}

// pluginMap is the map of plugins we can dispense
var pluginMap = map[string]plugin.Plugin{
	"cloud_provider": &CloudProviderPlugin{},
}

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "aws-provider",
		Level:  hclog.Info,
		Output: os.Stderr,
	})

	// Create the provider
	awsProvider := &AWSProvider{
		logger: logger,
	}

	// Serve the plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}