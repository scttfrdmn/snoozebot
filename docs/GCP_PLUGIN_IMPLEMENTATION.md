# GCP Plugin Implementation Guide

This document outlines the plan for completing the Google Cloud Platform (GCP) plugin for Snoozebot to achieve feature parity with the AWS and Azure plugins.

## Current Status

The GCP plugin currently has a basic structure but is missing key functionality:
- Missing GetAPIVersion method
- Missing or incomplete ListInstances method
- May have outdated imports or dependencies
- May not handle authentication properly

## Implementation Roadmap

### 1. Plugin Structure Updates

Update the GCP plugin to match the structure of the AWS and Azure plugins:

```go
// Main plugin structure
type GCPProvider struct {
    *snoozePlugin.BaseProvider
    logger        hclog.Logger
    computeClient *compute.Service
    projectID     string
    zone          string
}

// Constructor function
func NewGCPProvider(logger hclog.Logger) (*GCPProvider, error) {
    // Initialize the BaseProvider
    baseProvider := snoozePlugin.NewBaseProvider("gcp", "0.1.0", logger)
    
    // Create GCP provider
    provider := &GCPProvider{
        BaseProvider: baseProvider,
        logger:       logger,
        projectID:    os.Getenv("GCP_PROJECT_ID"),
        zone:         getDefaultZone(),
    }
    
    // Add capabilities
    provider.AddCapability("list_instances")
    provider.AddCapability("start_instance")
    provider.AddCapability("stop_instance")
    
    // Set up the compute client
    client, err := createComputeClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create compute client: %w", err)
    }
    provider.computeClient = client
    
    return provider, nil
}
```

### 2. Authentication Implementation

Implement Google Cloud authentication:

```go
func createComputeClient(ctx context.Context) (*compute.Service, error) {
    // Try to use Application Default Credentials
    client, err := google.DefaultClient(ctx, compute.ComputeScope)
    if err != nil {
        return nil, fmt.Errorf("failed to create default client: %w", err)
    }
    
    // Create the compute service
    service, err := compute.New(client)
    if err != nil {
        return nil, fmt.Errorf("failed to create compute service: %w", err)
    }
    
    return service, nil
}
```

### 3. Required Methods Implementation

Implement all required methods from the CloudProvider interface:

```go
// GetAPIVersion returns the API version implemented by the plugin
func (p *GCPProvider) GetAPIVersion() string {
    return snoozePlugin.CurrentAPIVersion
}

// GetInstanceInfo gets information about the current instance
func (p *GCPProvider) GetInstanceInfo(ctx context.Context) (*snoozePlugin.InstanceInfo, error) {
    // Implementation details
}

// StopInstance stops the current instance
func (p *GCPProvider) StopInstance(ctx context.Context) error {
    // Implementation details
}

// StartInstance starts the current instance
func (p *GCPProvider) StartInstance(ctx context.Context) error {
    // Implementation details
}

// GetProviderName returns the name of the cloud provider
func (p *GCPProvider) GetProviderName() string {
    return "gcp"
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *GCPProvider) GetProviderVersion() string {
    return "0.1.0"
}

// ListInstances lists all instances in the current project/zone
func (p *GCPProvider) ListInstances(ctx context.Context) ([]*snoozePlugin.InstanceInfo, error) {
    // Implementation details
}

// Shutdown performs cleanup when the plugin is being unloaded
func (p *GCPProvider) Shutdown() {
    p.logger.Info("Shutting down GCP provider")
}
```

### 4. List Instances Implementation

Detailed implementation of the ListInstances method:

```go
// ListInstances lists all instances in the current project/zone
func (p *GCPProvider) ListInstances(ctx context.Context) ([]*snoozePlugin.InstanceInfo, error) {
    p.logger.Info("Listing instances", "project", p.projectID, "zone", p.zone)
    
    // Create request to list instances
    req := p.computeClient.Instances.List(p.projectID, p.zone)
    
    // Execute the request
    var instances []*snoozePlugin.InstanceInfo
    err := req.Pages(ctx, func(page *compute.InstanceList) error {
        for _, instance := range page.Items {
            // Parse launch time
            launchTime, _ := time.Parse(time.RFC3339, instance.CreationTimestamp)
            
            // Convert state to string
            state := "unknown"
            if instance.Status != "" {
                state = strings.ToLower(instance.Status)
            }
            
            // Create InstanceInfo
            info := &snoozePlugin.InstanceInfo{
                ID:         instance.Id,
                Name:       instance.Name,
                Type:       instance.MachineType,
                Region:     extractRegionFromZone(p.zone),
                Zone:       p.zone,
                State:      state,
                LaunchTime: launchTime,
            }
            
            instances = append(instances, info)
        }
        return nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to list instances: %w", err)
    }
    
    p.logger.Info("Found instances", "count", len(instances))
    return instances, nil
}
```

### 5. Instance Operations Implementation

Implement methods to start and stop instances:

```go
// StopInstance stops an instance
func (p *GCPProvider) StopInstance(ctx context.Context) error {
    instanceID := os.Getenv("INSTANCE_ID")
    if instanceID == "" {
        return fmt.Errorf("no instance ID provided")
    }
    
    p.logger.Info("Stopping instance", "instanceID", instanceID)
    
    // Create and execute stop request
    op, err := p.computeClient.Instances.Stop(p.projectID, p.zone, instanceID).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("failed to stop instance: %w", err)
    }
    
    // Wait for operation to complete
    if err := p.waitForOperation(ctx, op); err != nil {
        return fmt.Errorf("error waiting for instance to stop: %w", err)
    }
    
    p.logger.Info("Instance stopped successfully", "instanceID", instanceID)
    return nil
}

// StartInstance starts an instance
func (p *GCPProvider) StartInstance(ctx context.Context) error {
    instanceID := os.Getenv("INSTANCE_ID")
    if instanceID == "" {
        return fmt.Errorf("no instance ID provided")
    }
    
    p.logger.Info("Starting instance", "instanceID", instanceID)
    
    // Create and execute start request
    op, err := p.computeClient.Instances.Start(p.projectID, p.zone, instanceID).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("failed to start instance: %w", err)
    }
    
    // Wait for operation to complete
    if err := p.waitForOperation(ctx, op); err != nil {
        return fmt.Errorf("error waiting for instance to start: %w", err)
    }
    
    p.logger.Info("Instance started successfully", "instanceID", instanceID)
    return nil
}

// Helper method to wait for an operation to complete
func (p *GCPProvider) waitForOperation(ctx context.Context, op *compute.Operation) error {
    // Implementation to poll the operation until it completes
}
```

### 6. Helper Functions

Add necessary helper functions:

```go
// Extract region from zone (e.g., us-central1-a -> us-central1)
func extractRegionFromZone(zone string) string {
    parts := strings.Split(zone, "-")
    if len(parts) < 3 {
        return "unknown"
    }
    return strings.Join(parts[:len(parts)-1], "-")
}

// Get default zone from environment or use a fallback
func getDefaultZone() string {
    zone := os.Getenv("GCP_ZONE")
    if zone == "" {
        return "us-central1-a" // Default zone
    }
    return zone
}
```

### 7. Error Handling

Implement robust error handling:

```go
// Standardized error handling
func handleGCPError(err error, operation string) error {
    if err == nil {
        return nil
    }
    
    // Check for specific GCP error types
    if gerr, ok := err.(*googleapi.Error); ok {
        switch gerr.Code {
        case 403:
            return fmt.Errorf("%s: permission denied: %w", operation, err)
        case 404:
            return fmt.Errorf("%s: resource not found: %w", operation, err)
        case 429:
            return fmt.Errorf("%s: quota exceeded: %w", operation, err)
        default:
            return fmt.Errorf("%s: GCP API error (code %d): %w", operation, gerr.Code, err)
        }
    }
    
    // General error
    return fmt.Errorf("%s: %w", operation, err)
}
```

## Dependencies

The GCP plugin requires the following dependencies:

```
github.com/hashicorp/go-hclog
github.com/hashicorp/go-plugin
google.golang.org/api/compute/v1
google.golang.org/api/option
golang.org/x/oauth2/google
```

Add these to the go.mod file if not already present.

## Testing

### Unit Tests

Create unit tests for the GCP plugin:

```go
func TestGCPProvider_GetAPIVersion(t *testing.T) {
    provider := createTestProvider(t)
    version := provider.GetAPIVersion()
    assert.Equal(t, snoozePlugin.CurrentAPIVersion, version)
}

func TestGCPProvider_ListInstances(t *testing.T) {
    // Test with mock compute service
}
```

### Integration Tests

Create integration tests that require real GCP credentials:

```go
func TestGCPProvider_Integration(t *testing.T) {
    if os.Getenv("SNOOZEBOT_INTEGRATION_TESTS") != "true" {
        t.Skip("Skipping integration tests")
    }
    
    // Requires valid GCP credentials in environment
    provider, err := NewGCPProvider(hclog.Default())
    require.NoError(t, err)
    
    // Test list instances
    instances, err := provider.ListInstances(context.Background())
    require.NoError(t, err)
    
    // Verify instances list
    t.Logf("Found %d instances", len(instances))
    for _, inst := range instances {
        t.Logf("Instance: %s, State: %s", inst.Name, inst.State)
    }
}
```

## Implementation Sequence

1. Update plugin structure and add GetAPIVersion
2. Implement authentication and basic client setup
3. Implement ListInstances method
4. Add instance operation methods (start/stop)
5. Implement error handling
6. Add helper functions
7. Create tests
8. Update documentation

## Configuration Guide

Document the environment variables used:

- `GCP_PROJECT_ID`: The GCP project ID to use
- `GCP_ZONE`: The GCP zone to use (defaults to us-central1-a)
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to GCP credentials file
- `INSTANCE_ID`: ID of the specific instance for operations