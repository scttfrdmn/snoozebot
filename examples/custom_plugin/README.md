# Custom Cloud Provider Plugin Example

This example demonstrates how to create a custom cloud provider plugin for Snoozebot with proper versioning support.

## Features

This example shows:

- How to implement the CloudProvider interface
- How to use the BaseProvider for common functionality
- How to create and manage a plugin manifest
- How to handle API version compatibility
- How to support security features (TLS, signatures)
- Best practices for plugin development

## Prerequisites

- Go 1.18 or newer
- Snoozebot repository

## Building the Plugin

```bash
# Navigate to the example directory
cd examples/custom_plugin

# Build the plugin
go build -o ../../bin/plugins/custom-provider main.go
```

## Plugin Manifest

The plugin includes a manifest with:

- API version information
- Plugin metadata (name, version, author)
- Capabilities
- Compatibility requirements

This manifest is saved when the plugin starts and is used by the host for compatibility checking and capability discovery.

## Testing with the Agent

1. Build the plugin as described above
2. Start the Snoozebot agent with plugin discovery enabled:

```bash
# Set the plugin directory
export SNOOZEBOT_PLUGIN_DIR=/path/to/snoozebot/bin/plugins

# Start the agent
./bin/snooze-agent
```

3. The agent should discover and load the custom plugin
4. Check the agent logs for successful plugin loading

## Customizing the Plugin

To adapt this example for your own cloud provider:

1. Change the provider name, version, and description
2. Implement the actual cloud provider-specific functionality
3. Add any additional capabilities your provider supports
4. Update the manifest with your information

## Security Features

The plugin supports:

- TLS encryption for secure plugin communication
- Signature verification (when enabled by the host)
- API version compatibility checking

These can be enabled via environment variables:

```bash
# Enable TLS
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_DIR=/path/to/certs

# Enable signature verification (on the host side)
export SNOOZEBOT_SIGNATURE_ENABLED=true
export SNOOZEBOT_SIGNATURE_DIR=/path/to/signatures
```

## Troubleshooting

- If the plugin fails to load, check the API version compatibility
- Verify that the required methods are properly implemented
- Check the logs for detailed error messages
- Ensure the plugin binary is in the expected location