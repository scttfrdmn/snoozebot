# Snoozebot Examples

This directory contains examples for using and extending Snoozebot.

## Available Examples

- [Embedded](./embedded/) - Example of embedding the Snoozebot monitoring library in a host application
- [Custom Plugin](./custom_plugin/) - Example of creating a custom cloud provider plugin with versioning support

## Running the Examples

### Embedded Example

```bash
# Build the embedded example
go build -o ../bin/embedded ./embedded

# Run the embedded example
../bin/embedded
```

### Custom Plugin Example

```bash
# Build the custom plugin
go build -o ../bin/plugins/custom-provider ./custom_plugin

# Test the plugin
../scripts/test_plugin_compatibility.sh custom-provider ./custom_plugin
```

## Using with the Agent

The examples can be used with the Snoozebot agent for testing the complete system:

```bash
# Start the agent with plugins
export SNOOZEBOT_PLUGIN_DIR=../bin/plugins
../bin/snoozed
```