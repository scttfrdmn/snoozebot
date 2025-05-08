# Azure Cloud Provider Plugin Implementation

This PR implements the Azure cloud provider plugin for Snoozebot, completing the set of major cloud providers (AWS, GCP, and now Azure).

## Features

- Full implementation of the CloudProvider interface for Azure
- Support for VM start and stop operations
- Instance information retrieval
- Resource group-based VM discovery
- Error handling with retries and timeouts
- Comprehensive logging
- Configuration via environment variables
- Unit tests with mocking

## Implementation Details

The plugin uses the Azure SDK for Go to interact with Azure services:

- `azure-sdk-for-go/sdk/azidentity` for authentication
- `azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4` for VM operations

It follows the same pattern as the existing AWS and GCP plugins while accommodating Azure-specific concepts like resource groups and VM power states.

## Testing

The implementation includes:
- Unit tests using mocks
- Integration with the plugin testing framework

## Documentation

The PR includes:
- README for the Azure plugin
- Updates to PLUGIN_SYSTEM_PROGRESS.md
- Makefile additions for building and testing

## Changes

- Added Azure plugin implementation in `/plugins/azure/`
- Added Azure plugin tests in `/test/azure/`
- Updated documentation to reflect completion of Azure plugin
- Enhanced Makefile with Azure-specific targets
- Updated version management to include Azure plugin

## Testing Instructions

1. Set the required environment variables:
   ```
   export AZURE_SUBSCRIPTION_ID="your-subscription-id"
   export AZURE_RESOURCE_GROUP="your-resource-group"
   export AZURE_VM_NAME="your-vm-name"
   ```
   
2. Build the plugin:
   ```
   make plugin PLUGIN=azure
   ```
   
3. Run the tests:
   ```
   make test-azure
   ```
   
4. Load the plugin via the agent API:
   ```
   curl -X POST http://localhost:8080/api/v1/plugins/load -d '{"name": "azure"}'
   ```