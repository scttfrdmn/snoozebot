# Azure Cloud Provider Plugin

This plugin provides Azure integration for the Snoozebot agent, allowing it to manage Azure virtual machines.

## Features

- Instance management (start/stop operations)
- Instance information retrieval
- Resource group-based VM discovery
- Status monitoring

## Configuration

The plugin is configured using environment variables:

| Environment Variable | Description | Required | Default |
|---|---|---|---|
| AZURE_SUBSCRIPTION_ID | The Azure subscription ID | Yes | N/A |
| AZURE_RESOURCE_GROUP | The resource group containing the VM | Yes | N/A |
| AZURE_VM_NAME | The name of the VM to manage | No | "default-vm" |
| AZURE_LOCATION | The Azure region | No | "eastus" |

## Authentication

The plugin uses DefaultAzureCredential from the Azure SDK for Go, which supports:

- Environment variables (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
- Managed Identity
- Azure CLI credentials
- Visual Studio Code credentials

## Building

```bash
make plugin PLUGIN=azure
```

## Testing

```bash
make test-azure
```

## Installation

```bash
make install
```

This will install the plugin to `/etc/snoozebot/plugins/azure`

## Usage

The plugin is automatically discovered and can be loaded by the Snoozebot agent:

```bash
snooze-agent --plugins-dir=/etc/snoozebot/plugins
```

Or you can manually load it through the API:

```bash
curl -X POST http://localhost:8080/api/v1/plugins/load -d '{"name": "azure"}'
```

## Implementation Details

The plugin uses the Azure SDK for Go to interact with Azure services, specifically:

- Azure Compute API for VM operations
- Azure Identity for authentication

It implements the CloudProvider interface defined in the Snoozebot plugin system.