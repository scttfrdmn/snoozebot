package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/scttfrdmn/snoozebot/agent/provider"
)

// AWSProvider is an implementation of CloudProvider for AWS
type AWSProvider struct {
	logger  hclog.Logger
	ec2Client *ec2.Client
}

// NewAWSProvider creates a new AWS provider
func NewAWSProvider(logger hclog.Logger) (*AWSProvider, error) {
	// Load AWS configuration from environment variables or shared credentials file
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	return &AWSProvider{
		logger:  logger,
		ec2Client: ec2Client,
	}, nil
}

// GetInstanceInfo gets information about an instance
func (p *AWSProvider) GetInstanceInfo(ctx context.Context, instanceID string) (*provider.InstanceInfo, error) {
	p.logger.Info("Getting instance info", "instanceID", instanceID)
	
	// Call AWS EC2 API to get instance information
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	
	result, err := p.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}
	
	// Check if instance was found
	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}
	
	// Extract instance information
	instance := result.Reservations[0].Instances[0]
	
	// Extract name from tags
	var name string
	for _, tag := range instance.Tags {
		if *tag.Key == "Name" {
			name = *tag.Value
			break
		}
	}
	
	// Convert state to string
	state := "unknown"
	if instance.State != nil {
		state = string(instance.State.Name)
	}
	
	// Convert to provider.InstanceInfo
	return &provider.InstanceInfo{
		ID:         instanceID,
		Name:       name,
		Type:       string(instance.InstanceType),
		Region:     "unknown", // Region information is not available from DescribeInstances
		Zone:       aws.ToString(instance.Placement.AvailabilityZone),
		State:      state,
		LaunchTime: aws.ToTime(instance.LaunchTime),
		Provider:   "aws",
	}, nil
}

// StopInstance stops an instance
func (p *AWSProvider) StopInstance(ctx context.Context, instanceID string) error {
	p.logger.Info("Stopping instance", "instanceID", instanceID)
	
	// Call AWS EC2 API to stop the instance
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}
	
	_, err := p.ec2Client.StopInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}
	
	p.logger.Info("Instance stop request sent", "instanceID", instanceID)
	return nil
}

// StartInstance starts an instance
func (p *AWSProvider) StartInstance(ctx context.Context, instanceID string) error {
	p.logger.Info("Starting instance", "instanceID", instanceID)
	
	// Call AWS EC2 API to start the instance
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}
	
	_, err := p.ec2Client.StartInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}
	
	p.logger.Info("Instance start request sent", "instanceID", instanceID)
	return nil
}

// GetProviderName returns the name of the cloud provider
func (p *AWSProvider) GetProviderName() string {
	return "aws"
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *AWSProvider) GetProviderVersion() string {
	return "0.1.0"
}

// ListInstances lists all instances in the account
func (p *AWSProvider) ListInstances(ctx context.Context) ([]provider.InstanceInfo, error) {
	p.logger.Info("Listing instances")
	
	// Call AWS EC2 API to list instances
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "stopped", "stopping"},
			},
		},
	}
	
	result, err := p.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}
	
	// Extract instance information
	instances := make([]provider.InstanceInfo, 0)
	
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Extract name from tags
			var name string
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					name = *tag.Value
					break
				}
			}
			
			// Convert state to string
			state := "unknown"
			if instance.State != nil {
				state = string(instance.State.Name)
			}
			
			instances = append(instances, provider.InstanceInfo{
				ID:         *instance.InstanceId,
				Name:       name,
				Type:       string(instance.InstanceType),
				Region:     "unknown", // Region information is not available from DescribeInstances
				Zone:       aws.ToString(instance.Placement.AvailabilityZone),
				State:      state,
				LaunchTime: aws.ToTime(instance.LaunchTime),
				Provider:   "aws",
			})
		}
	}
	
	return instances, nil
}

// CloudProviderPlugin is the implementation of plugin.Plugin
type CloudProviderPlugin struct {
	// Impl is the concrete implementation of CloudProvider
	Impl provider.CloudProvider
}

// GRPCServer registers this plugin for serving with a gRPC server
func (p *CloudProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *plugin.DefaultGRPCServer) error {
	// TODO: Implement proper gRPC server
	return nil
}

// GRPCClient returns the client for this plugin
func (p *CloudProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *plugin.DefaultGRPCClient) (interface{}, error) {
	// TODO: Implement proper gRPC client
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
	awsProvider, err := NewAWSProvider(logger)
	if err != nil {
		logger.Error("Failed to create AWS provider", "error", err)
		os.Exit(1)
	}

	// Serve the plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"cloud_provider": &CloudProviderPlugin{
				Impl: awsProvider,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}