package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	snoozePlugin "github.com/scttfrdmn/snoozebot/pkg/plugin"
)

// AzureProvider is an implementation of CloudProvider for Azure
type AzureProvider struct {
	logger             hclog.Logger
	vmClient           *armcompute.VirtualMachinesClient
	subscriptionID     string
	resourceGroupName  string
	currentInstanceID  string
	currentVMName      string
	location           string
}

// StateMapping maps Azure VM states to internal states
var StateMapping = map[string]string{
	"PowerState/running":      "running",
	"PowerState/deallocated":  "stopped",
	"PowerState/deallocating": "stopping",
	"PowerState/starting":     "starting",
	"PowerState/stopped":      "stopped",
}

// New Azure provider creates a new Azure provider
func NewAzureProvider(logger hclog.Logger) (*AzureProvider, error) {
	// Load configuration from environment variables
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return nil, fmt.Errorf("AZURE_SUBSCRIPTION_ID environment variable is required")
	}

	resourceGroupName := os.Getenv("AZURE_RESOURCE_GROUP")
	if resourceGroupName == "" {
		return nil, fmt.Errorf("AZURE_RESOURCE_GROUP environment variable is required")
	}

	location := os.Getenv("AZURE_LOCATION")
	if location == "" {
		location = "eastus" // Default to East US
		logger.Warn("AZURE_LOCATION not set, using default", "location", location)
	}

	// In a real implementation, we would get the VM name from metadata
	// For simplicity, we'll get it from an environment variable
	vmName := os.Getenv("AZURE_VM_NAME")
	if vmName == "" {
		vmName = "default-vm"
		logger.Warn("AZURE_VM_NAME not set, using default", "vm_name", vmName)
	}

	// Create a credential using DefaultAzureCredential
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Create the VM client
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM client: %w", err)
	}

	return &AzureProvider{
		logger:            logger,
		vmClient:          vmClient,
		subscriptionID:    subscriptionID,
		resourceGroupName: resourceGroupName,
		currentVMName:     vmName,
		location:          location,
	}, nil
}

// extractInstanceID extracts the instance ID from the full resource ID
func extractInstanceID(resourceID string) string {
	// Azure resource IDs have the format:
	// /subscriptions/{subId}/resourceGroups/{resourceGroup}/providers/Microsoft.Compute/virtualMachines/{vmName}
	// We'll just return the VM name as the instance ID for simplicity
	parts := strings.Split(resourceID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return resourceID
}

// GetInstanceInfo gets information about the current instance
func (p *AzureProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.InstanceInfo, error) {
	p.logger.Info("Getting instance info", "vm_name", p.currentVMName)

	// Get the VM
	resp, err := p.vmClient.Get(ctx, p.resourceGroupName, p.currentVMName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	// Extract the instance ID
	instanceID := extractInstanceID(*resp.ID)
	p.currentInstanceID = instanceID

	// Get instance view to determine power state
	view, err := p.vmClient.InstanceView(ctx, p.resourceGroupName, p.currentVMName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM instance view: %w", err)
	}

	// Determine state
	state := "unknown"
	if view.Statuses != nil {
		for _, s := range view.Statuses {
			if s.Code != nil && strings.HasPrefix(*s.Code, "PowerState/") {
				state = StateMapping[*s.Code]
				break
			}
		}
	}

	// Determine instance type (size)
	instanceType := "unknown"
	if resp.Properties != nil && resp.Properties.HardwareProfile != nil && resp.Properties.HardwareProfile.VMSize != nil {
		instanceType = string(*resp.Properties.HardwareProfile.VMSize)
	}

	// Extract tags if available
	name := p.currentVMName
	if resp.Tags != nil {
		if nameTag, ok := resp.Tags["Name"]; ok && nameTag != nil {
			name = *nameTag
		}
	}

	// Extract zone if available
	zone := ""
	if resp.Zones != nil && len(resp.Zones) > 0 {
		zone = *resp.Zones[0]
	}

	// Create the instance info
	info := &snoozePlugin.InstanceInfo{
		ID:         instanceID,
		Name:       name,
		Type:       instanceType,
		Region:     p.location,
		Zone:       zone,
		State:      state,
		LaunchTime: time.Now(), // Azure doesn't provide this easily, so we use current time
	}

	return info, nil
}

// StopInstance stops the current instance
func (p *AzureProvider) StopInstance(ctx context.Context) error {
	p.logger.Info("Stopping instance", "vm_name", p.currentVMName)

	// Start a deallocate operation
	pollerResp, err := p.vmClient.BeginDeallocate(ctx, p.resourceGroupName, p.currentVMName, nil)
	if err != nil {
		return fmt.Errorf("failed to start VM deallocate operation: %w", err)
	}

	// Poll for operation completion with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if strings.Contains(err.Error(), "context deadline exceeded") {
			p.logger.Warn("Timeout waiting for VM to stop, operation is still in progress")
			return nil
		} else if strings.Contains(err.Error(), "OperationNotAllowed") && strings.Contains(err.Error(), "stopped") {
			p.logger.Info("VM is already stopped")
			return nil
		} else if strings.Contains(err.Error(), "ResourceNotFound") {
			p.logger.Warn("VM not found", "error", err)
			return fmt.Errorf("VM not found: %w", err)
		} else if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			p.logger.Warn("VM not found", "error", err)
			return fmt.Errorf("VM not found: %w", err)
		}
		return fmt.Errorf("failed to deallocate VM: %w", err)
	}

	p.logger.Info("VM stopped successfully", "vm_name", p.currentVMName)
	return nil
}

// StartInstance starts the current instance
func (p *AzureProvider) StartInstance(ctx context.Context) error {
	p.logger.Info("Starting instance", "vm_name", p.currentVMName)

	// Start a start operation
	pollerResp, err := p.vmClient.BeginStart(ctx, p.resourceGroupName, p.currentVMName, nil)
	if err != nil {
		return fmt.Errorf("failed to start VM start operation: %w", err)
	}

	// Poll for operation completion with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if strings.Contains(err.Error(), "context deadline exceeded") {
			p.logger.Warn("Timeout waiting for VM to start, operation is still in progress")
			return nil
		} else if strings.Contains(err.Error(), "OperationNotAllowed") && strings.Contains(err.Error(), "running") {
			p.logger.Info("VM is already running")
			return nil
		} else if strings.Contains(err.Error(), "ResourceNotFound") {
			p.logger.Warn("VM not found", "error", err)
			return fmt.Errorf("VM not found: %w", err)
		} else if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			p.logger.Warn("VM not found", "error", err)
			return fmt.Errorf("VM not found: %w", err)
		}
		return fmt.Errorf("failed to start VM: %w", err)
	}

	p.logger.Info("VM started successfully", "vm_name", p.currentVMName)
	return nil
}

// GetProviderName returns the name of the cloud provider
func (p *AzureProvider) GetProviderName() string {
	return "azure"
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *AzureProvider) GetProviderVersion() string {
	return "0.1.0"
}

// Shutdown performs cleanup when the plugin is being unloaded
func (p *AzureProvider) Shutdown() {
	p.logger.Info("Shutting down Azure provider")
	// No explicit cleanup needed for Azure client, but we could add it here if needed
}

// ListInstances lists all VMs in the resource group
func (p *AzureProvider) ListInstances(ctx context.Context) ([]*snoozePlugin.InstanceInfo, error) {
	p.logger.Info("Listing instances", "resource_group", p.resourceGroupName)
	
	// List VMs in the resource group
	pager := p.vmClient.NewListPager(p.resourceGroupName, nil)
	
	var instances []*snoozePlugin.InstanceInfo
	
	// Iterate through pages
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list VMs: %w", err)
		}
		
		// Process each VM
		for _, vm := range page.Value {
			// Extract the instance ID
			instanceID := extractInstanceID(*vm.ID)
			
			// Extract tags if available
			name := *vm.Name
			if vm.Tags != nil {
				if nameTag, ok := vm.Tags["Name"]; ok && nameTag != nil {
					name = *nameTag
				}
			}
			
			// Determine instance type (size)
			instanceType := "unknown"
			if vm.Properties != nil && vm.Properties.HardwareProfile != nil && vm.Properties.HardwareProfile.VMSize != nil {
				instanceType = string(*vm.Properties.HardwareProfile.VMSize)
			}
			
			// Extract zone if available
			zone := ""
			if vm.Zones != nil && len(vm.Zones) > 0 {
				zone = *vm.Zones[0]
			}
			
			// Create the instance info
			info := &snoozePlugin.InstanceInfo{
				ID:         instanceID,
				Name:       name,
				Type:       instanceType,
				Region:     p.location,
				Zone:       zone,
				State:      "unknown", // Would need instance view for accurate state
				LaunchTime: time.Now(), // Azure doesn't provide this easily
			}
			
			instances = append(instances, info)
		}
	}
	
	return instances, nil
}

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "azure-provider",
		Level:      hclog.Info,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	// Create the provider
	azureProvider, err := NewAzureProvider(logger)
	if err != nil {
		logger.Error("Failed to create Azure provider", "error", err)
		os.Exit(1)
	}

	// Serve the plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: snoozePlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			"cloud_provider": &snoozePlugin.CloudProviderPlugin{
				Impl: azureProvider,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}