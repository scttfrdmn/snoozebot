#!/bin/bash
# Script to test the Azure plugin functionality

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Binary paths
AGENT_BIN="$PROJECT_ROOT/bin/snooze-agent"
PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/azure"

# Check if files exist
if [ ! -f "$AGENT_BIN" ]; then
    echo "Agent binary not found. Building..."
    (cd "$PROJECT_ROOT" && make agent)
fi

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
$PLUGIN_BIN 2>&1 | grep "azure-provider" || { echo "Plugin self-test failed"; exit 1; }

echo "Tests completed successfully!"