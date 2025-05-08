#!/bin/bash

# Script to test plugin version compatibility

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
PLUGIN_BIN_DIR="$REPO_ROOT/bin/plugins"

# Check if plugin name was provided
if [ -z "$1" ]; then
  echo "Error: No plugin name provided"
  echo "Usage: $0 <plugin_name> [plugin_directory]"
  echo "Example: $0 aws"
  echo "Example: $0 custom-provider ./examples/custom_plugin"
  exit 1
fi

PLUGIN_NAME="$1"
PLUGIN_DIR="$REPO_ROOT/plugins/$PLUGIN_NAME"

# Use provided plugin directory if specified
if [ -n "$2" ]; then
  PLUGIN_DIR="$2"
fi

# Check if plugin directory exists
if [ ! -d "$PLUGIN_DIR" ]; then
  echo "Error: Plugin directory does not exist: $PLUGIN_DIR"
  exit 1
fi

# Load the current API version from the main package
CURRENT_API_VERSION=$(grep -r "CurrentAPIVersion = " "$REPO_ROOT/pkg/plugin/plugin.go" | sed 's/.*CurrentAPIVersion = "\(.*\)".*/\1/')

echo "Testing plugin compatibility for: $PLUGIN_NAME"
echo "Current API version: $CURRENT_API_VERSION"
echo "Plugin directory: $PLUGIN_DIR"
echo

# Check if the plugin implements GetAPIVersion
if ! grep -q "GetAPIVersion" "$PLUGIN_DIR/main.go"; then
  echo "❌ Plugin does not implement GetAPIVersion method"
  echo "Please add the following method to your plugin:"
  echo
  echo "func (p *YourProvider) GetAPIVersion() string {"
  echo "    return snoozePlugin.CurrentAPIVersion"
  echo "}"
  exit 1
fi

echo "✅ Plugin implements GetAPIVersion method"

# Check if API version is correctly returned
if ! grep -q "CurrentAPIVersion" "$PLUGIN_DIR/main.go"; then
  echo "⚠️ Plugin may not be using the current API version constant"
  echo "Consider using snoozePlugin.CurrentAPIVersion for better compatibility"
fi

# Build plugin
echo "Building plugin..."
mkdir -p "$PLUGIN_BIN_DIR"
cd "$PLUGIN_DIR"
go build -o "$PLUGIN_BIN_DIR/$PLUGIN_NAME" .
echo "✅ Plugin built successfully"

# Run the plugin with version check environment variable
echo "Testing plugin version compatibility..."
SNOOZEBOT_VERSION_CHECK_ONLY=true "$PLUGIN_BIN_DIR/$PLUGIN_NAME" 2>&1 | grep -q "API version is compatible" || {
  echo "❌ Plugin API version is not compatible"
  exit 1
}

echo "✅ Plugin API version is compatible with current API version"

# Check plugin interface implementation
echo "Verifying CloudProvider interface implementation..."
go vet -vettool="$REPO_ROOT/bin/ifacecheck" "$PLUGIN_DIR/main.go" || {
  echo "⚠️ Plugin may not implement all required CloudProvider methods"
  echo "Please check interface implementation"
}

echo
echo "✅ All compatibility checks passed for plugin: $PLUGIN_NAME"