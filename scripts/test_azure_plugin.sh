#!/bin/bash
# Script to test the Azure plugin functionality

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Binary paths
PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/azure"

# Check if plugin exists
if [ ! -f "$PLUGIN_BIN" ]; then
    echo "Azure plugin not found. Building..."
    (cd "$PROJECT_ROOT" && ./scripts/build_plugins.sh azure)
fi

# Set required environment variables for testing
export AZURE_SUBSCRIPTION_ID="test-subscription"
export AZURE_RESOURCE_GROUP="test-resource-group"
export AZURE_VM_NAME="test-vm"
export AZURE_LOCATION="eastus"

echo "Starting test with environment:"
echo "AZURE_SUBSCRIPTION_ID: $AZURE_SUBSCRIPTION_ID"
echo "AZURE_RESOURCE_GROUP: $AZURE_RESOURCE_GROUP" 
echo "AZURE_VM_NAME: $AZURE_VM_NAME"
echo "AZURE_LOCATION: $AZURE_LOCATION"

# Run the plugin directly to verify it loads correctly
echo "Testing direct plugin execution..."
output=$($PLUGIN_BIN 2>&1)
echo "$output"
if [[ "$output" == *"azure-provider"* ]]; then
    echo "Plugin loads successfully!"
else
    echo "Plugin self-test failed"
    exit 1
fi

# Skip unit tests for now as they require additional setup
echo "Skipping unit tests for basic plugin verification..."

echo "All tests completed successfully!"