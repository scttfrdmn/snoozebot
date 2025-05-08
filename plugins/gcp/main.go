package main

import (
	"context"
	"fmt"
	"os"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	snoozePlugin "github.com/scttfrdmn/snoozebot/pkg/plugin"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCPProvider is an implementation of CloudProvider for GCP
type GCPProvider struct {
	logger             hclog.Logger
	instancesClient    *compute.InstancesClient
	currentInstanceID  string
	currentZone        string
	currentProject     string
}

// NewGCPProvider creates a new GCP provider
func NewGCPProvider(logger hclog.Logger) (*GCPProvider, error) {
	ctx := context.Background()
	
	// Create the instances client
	instancesClient, err := compute.NewInstancesRESTClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	// In a real implementation, we would get the instance ID from metadata
	// For simplicity, we'll get it from environment variables or use defaults
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "default-instance"
		logger.Warn("Using default instance ID", "instanceID", instanceID)
	}

	zone := os.Getenv("ZONE")
	if zone == "" {
		zone = "us-central1-a"
		logger.Warn("Using default zone", "zone", zone)
	}

	project := os.Getenv("PROJECT_ID")
	if project == "" {
		project = "default-project"
		logger.Warn("Using default project ID", "project", project)
	}

	return &GCPProvider{
		logger:            logger,
		instancesClient:   instancesClient,
		currentInstanceID: instanceID,
		currentZone:       zone,
		currentProject:    project,
	}, nil
}

// GetInstanceInfo gets information about the current instance
func (p *GCPProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.InstanceInfo, error) {
	p.logger.Info("Getting instance info", 
		"instanceID", p.currentInstanceID,
		"zone", p.currentZone,
		"project", p.currentProject)
	
	// Call GCP Compute API to get instance information
	req := &computepb.GetInstanceRequest{
		Instance: p.currentInstanceID,
		Zone:     p.currentZone,
		Project:  p.currentProject,
	}
	
	instance, err := p.instancesClient.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	
	// Convert to snoozePlugin.InstanceInfo
	return &snoozePlugin.InstanceInfo{
		ID:         p.currentInstanceID,
		Name:       instance.GetName(),
		Type:       instance.GetMachineType(),
		Region:     p.currentZone[:len(p.currentZone)-2], // Remove the zone suffix to get the region
		Zone:       p.currentZone,
		State:      instance.GetStatus(),
		LaunchTime: time.Unix(instance.GetCreationTimestamp(), 0),
	}, nil
}

// StopInstance stops the current instance
func (p *GCPProvider) StopInstance(ctx context.Context) error {
	p.logger.Info("Stopping instance", 
		"instanceID", p.currentInstanceID,
		"zone", p.currentZone,
		"project", p.currentProject)
	
	// Call GCP Compute API to stop the instance
	req := &computepb.StopInstanceRequest{
		Instance: p.currentInstanceID,
		Zone:     p.currentZone,
		Project:  p.currentProject,
	}
	
	op, err := p.instancesClient.Stop(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}
	
	// Wait for the operation to complete
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for instance to stop: %w", err)
	}
	
	p.logger.Info("Instance stopped", "instanceID", p.currentInstanceID)
	return nil
}

// StartInstance starts the current instance
func (p *GCPProvider) StartInstance(ctx context.Context) error {
	p.logger.Info("Starting instance", 
		"instanceID", p.currentInstanceID,
		"zone", p.currentZone,
		"project", p.currentProject)
	
	// Call GCP Compute API to start the instance
	req := &computepb.StartInstanceRequest{
		Instance: p.currentInstanceID,
		Zone:     p.currentZone,
		Project:  p.currentProject,
	}
	
	op, err := p.instancesClient.Start(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}
	
	// Wait for the operation to complete
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for instance to start: %w", err)
	}
	
	p.logger.Info("Instance started", "instanceID", p.currentInstanceID)
	return nil
}

// GetProviderName returns the name of the cloud provider
func (p *GCPProvider) GetProviderName() string {
	return "gcp"
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *GCPProvider) GetProviderVersion() string {
	return "0.1.0"
}

// GetAPIVersion returns the API version implemented by the plugin
func (p *GCPProvider) GetAPIVersion() string {
	return "0.1.0" // Match the project version
}

// Shutdown performs cleanup when the plugin is being unloaded
func (p *GCPProvider) Shutdown() {
	p.logger.Info("Shutting down GCP provider")
	if p.instancesClient != nil {
		p.instancesClient.Close()
	}
}

// ListInstances lists all instances in the project and zone
func (p *GCPProvider) ListInstances(ctx context.Context) ([]*snoozePlugin.InstanceInfo, error) {
	p.logger.Info("Listing instances", 
		"zone", p.currentZone,
		"project", p.currentProject)
	
	// Call GCP Compute API to list instances
	req := &computepb.ListInstancesRequest{
		Zone:    p.currentZone,
		Project: p.currentProject,
	}
	
	it := p.instancesClient.List(ctx, req)
	instances := make([]*snoozePlugin.InstanceInfo, 0)
	
	for {
		instance, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list instances: %w", err)
		}
		
		instances = append(instances, &snoozePlugin.InstanceInfo{
			ID:         instance.GetId(),
			Name:       instance.GetName(),
			Type:       instance.GetMachineType(),
			Region:     p.currentZone[:len(p.currentZone)-2], // Remove the zone suffix to get the region
			Zone:       p.currentZone,
			State:      instance.GetStatus(),
			LaunchTime: time.Unix(instance.GetCreationTimestamp(), 0),
		})
	}
	
	return instances, nil
}

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "gcp-provider",
		Level:      hclog.Info,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	// Create the provider
	gcpProvider, err := NewGCPProvider(logger)
	if err != nil {
		logger.Error("Failed to create GCP provider", "error", err)
		os.Exit(1)
	}

	// Check for TLS configuration
	tlsEnabled := os.Getenv("SNOOZEBOT_TLS_ENABLED") == "true"
	// Check for signature verification configuration
	signatureEnabled := os.Getenv("SNOOZEBOT_SIGNATURE_ENABLED") == "true"
	
	if tlsEnabled || signatureEnabled {
		var securityOptions struct {
			TLS       *snoozePlugin.TLSOptions
			Signature bool
		}
		
		// Configure TLS if enabled
		if tlsEnabled {
			logger.Info("TLS enabled for plugin communication")
			
			// Set up TLS options
			securityOptions.TLS = &snoozePlugin.TLSOptions{
				Enabled: true,
			}
			
			// Check for custom cert paths
			certFile := os.Getenv("SNOOZEBOT_TLS_CERT_FILE")
			keyFile := os.Getenv("SNOOZEBOT_TLS_KEY_FILE")
			caFile := os.Getenv("SNOOZEBOT_TLS_CA_FILE")
			certDir := os.Getenv("SNOOZEBOT_TLS_CERT_DIR")
			
			if certFile != "" && keyFile != "" {
				securityOptions.TLS.CertFile = certFile
				securityOptions.TLS.KeyFile = keyFile
				securityOptions.TLS.CACert = caFile
				logger.Info("Using provided TLS certificates", "cert", certFile, "key", keyFile, "ca", caFile)
			} else if certDir != "" {
				securityOptions.TLS.CertDir = certDir
				logger.Info("Using TLS certificates from directory", "dir", certDir)
			} else {
				logger.Warn("TLS is enabled but no certificates specified, falling back to insecure mode")
				tlsEnabled = false
				securityOptions.TLS = nil
			}
			
			// Skip verification in debug mode
			if os.Getenv("SNOOZEBOT_TLS_SKIP_VERIFY") == "true" {
				if securityOptions.TLS != nil {
					securityOptions.TLS.SkipVerify = true
				}
				logger.Warn("TLS certificate verification disabled - INSECURE")
			}
		}
		
		// Configure signatures if enabled
		if signatureEnabled {
			logger.Info("Signature verification enabled for plugin")
			securityOptions.Signature = true
			
			// Check signature-related env vars for plugins (these are for informational purposes as
			// the actual signature is handled by the manager on the host side)
			signatureDir := os.Getenv("SNOOZEBOT_SIGNATURE_DIR")
			if signatureDir != "" {
				logger.Info("Using signature directory", "dir", signatureDir)
			}
		}
		
		// If security features are enabled, use the secure plugin server
		if tlsEnabled {
			// Serve the plugin with TLS (and signature awareness)
			snoozePlugin.ServePluginWithTLS(gcpProvider, securityOptions.TLS, logger)
			return
		}
	}
	
	// Serve the plugin without security features
	logger.Info("Security features disabled for plugin communication")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: snoozePlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			"cloud_provider": &snoozePlugin.CloudProviderPlugin{
				Impl: gcpProvider,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}