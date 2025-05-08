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
	snoozePlugin "github.com/scttfrdmn/snoozebot/pkg/plugin"
)

// AWSProvider is an implementation of CloudProvider for AWS
type AWSProvider struct {
	logger    hclog.Logger
	ec2Client *ec2.Client
	// Current instance ID - for the plugin implementation we assume we're running on an EC2 instance
	currentInstanceID string
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

	// In a real implementation, we would get the instance ID from metadata
	// For simplicity, we'll get it from an environment variable or use a default
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "i-default"
		logger.Warn("Using default instance ID", "instanceID", instanceID)
	}

	return &AWSProvider{
		logger:           logger,
		ec2Client:        ec2Client,
		currentInstanceID: instanceID,
	}, nil
}

// GetInstanceInfo gets information about the current instance
func (p *AWSProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.InstanceInfo, error) {
	p.logger.Info("Getting instance info", "instanceID", p.currentInstanceID)
	
	// Call AWS EC2 API to get instance information
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{p.currentInstanceID},
	}
	
	result, err := p.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}
	
	// Check if instance was found
	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance not found: %s", p.currentInstanceID)
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
	
	// Convert to snoozePlugin.InstanceInfo
	return &snoozePlugin.InstanceInfo{
		ID:         p.currentInstanceID,
		Name:       name,
		Type:       string(instance.InstanceType),
		Region:     "unknown", // Region information is not available from DescribeInstances
		Zone:       aws.ToString(instance.Placement.AvailabilityZone),
		State:      state,
		LaunchTime: aws.ToTime(instance.LaunchTime),
	}, nil
}

// StopInstance stops the current instance
func (p *AWSProvider) StopInstance(ctx context.Context) error {
	p.logger.Info("Stopping instance", "instanceID", p.currentInstanceID)
	
	// Call AWS EC2 API to stop the instance
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{p.currentInstanceID},
	}
	
	_, err := p.ec2Client.StopInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}
	
	p.logger.Info("Instance stop request sent", "instanceID", p.currentInstanceID)
	return nil
}

// StartInstance starts the current instance
func (p *AWSProvider) StartInstance(ctx context.Context) error {
	p.logger.Info("Starting instance", "instanceID", p.currentInstanceID)
	
	// Call AWS EC2 API to start the instance
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{p.currentInstanceID},
	}
	
	_, err := p.ec2Client.StartInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}
	
	p.logger.Info("Instance start request sent", "instanceID", p.currentInstanceID)
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

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "aws-provider",
		Level:      hclog.Info,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	// Create the provider
	awsProvider, err := NewAWSProvider(logger)
	if err != nil {
		logger.Error("Failed to create AWS provider", "error", err)
		os.Exit(1)
	}

	// Check for TLS configuration
	tlsEnabled := os.Getenv("SNOOZEBOT_TLS_ENABLED") == "true"
	
	if tlsEnabled {
		logger.Info("TLS enabled for plugin communication")
		
		// Set up TLS options
		tlsOptions := &snoozePlugin.TLSOptions{
			Enabled: true,
		}
		
		// Check for custom cert paths
		certFile := os.Getenv("SNOOZEBOT_TLS_CERT_FILE")
		keyFile := os.Getenv("SNOOZEBOT_TLS_KEY_FILE")
		caFile := os.Getenv("SNOOZEBOT_TLS_CA_FILE")
		certDir := os.Getenv("SNOOZEBOT_TLS_CERT_DIR")
		
		if certFile != "" && keyFile != "" {
			tlsOptions.CertFile = certFile
			tlsOptions.KeyFile = keyFile
			tlsOptions.CACert = caFile
			logger.Info("Using provided TLS certificates", "cert", certFile, "key", keyFile, "ca", caFile)
		} else if certDir != "" {
			tlsOptions.CertDir = certDir
			logger.Info("Using TLS certificates from directory", "dir", certDir)
		} else {
			logger.Warn("TLS is enabled but no certificates specified, falling back to insecure mode")
			tlsEnabled = false
		}
		
		// Skip verification in debug mode
		if os.Getenv("SNOOZEBOT_TLS_SKIP_VERIFY") == "true" {
			tlsOptions.SkipVerify = true
			logger.Warn("TLS certificate verification disabled - INSECURE")
		}
		
		if tlsEnabled {
			// Serve the plugin with TLS
			snoozePlugin.ServePluginWithTLS(awsProvider, tlsOptions, logger)
			return
		}
	}
	
	// Serve the plugin without TLS
	logger.Info("TLS disabled for plugin communication")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: snoozePlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			"cloud_provider": &snoozePlugin.CloudProviderPlugin{
				Impl: awsProvider,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}