#!/bin/bash
# Script to test the Azure plugin with TLS encryption using real Azure credentials

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Binary paths
PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/azure"
SNOOZE_BIN="$PROJECT_ROOT/bin/snooze"

# Certificate directory
CERT_DIR="$PROJECT_ROOT/test/certs"

# Check if plugin exists
if [ ! -f "$PLUGIN_BIN" ]; then
    echo "Azure plugin not found. Building..."
    (cd "$PROJECT_ROOT" && ./scripts/build_plugins.sh azure)
fi

# Check if snooze CLI exists
if [ ! -f "$SNOOZE_BIN" ]; then
    echo "Snooze CLI not found. Building..."
    (cd "$PROJECT_ROOT" && go build -o "$SNOOZE_BIN" ./cmd/snooze)
fi

# Check for Azure credentials
if [ -z "$AZURE_SUBSCRIPTION_ID" ] || [ -z "$AZURE_RESOURCE_GROUP" ]; then
    echo "Error: Required Azure credentials not found in environment."
    echo "Please set the following environment variables:"
    echo "  AZURE_SUBSCRIPTION_ID: Your Azure subscription ID"
    echo "  AZURE_RESOURCE_GROUP: Your Azure resource group"
    echo "  AZURE_VM_NAME: Your Azure VM name"
    echo "  AZURE_LOCATION: Your Azure location (optional, defaults to eastus)"
    exit 1
fi

# Set default location if not provided
if [ -z "$AZURE_LOCATION" ]; then
    export AZURE_LOCATION="eastus"
    echo "AZURE_LOCATION not set, using default: eastus"
fi

# Verify VM name is provided
if [ -z "$AZURE_VM_NAME" ]; then
    echo "AZURE_VM_NAME not set, please specify the VM to use for testing"
    exit 1
fi

# Create certificate directory if it doesn't exist
mkdir -p "$CERT_DIR"
echo "Using certificate directory: $CERT_DIR"

# Enable TLS for the plugin
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_DIR="$CERT_DIR"

echo "Starting Azure plugin TLS test with real credentials:"
echo "AZURE_SUBSCRIPTION_ID: $AZURE_SUBSCRIPTION_ID"
echo "AZURE_RESOURCE_GROUP: $AZURE_RESOURCE_GROUP" 
echo "AZURE_VM_NAME: $AZURE_VM_NAME"
echo "AZURE_LOCATION: $AZURE_LOCATION"
echo "SNOOZEBOT_TLS_ENABLED: $SNOOZEBOT_TLS_ENABLED"
echo "SNOOZEBOT_TLS_CERT_DIR: $SNOOZEBOT_TLS_CERT_DIR"

# Test 1: Get Instance Information with TLS
echo "===== Test 1: Get Instance Information with TLS ====="
echo "Running: snooze provider azure get-info"
output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
echo "$output"
if [[ "$output" == *"$AZURE_VM_NAME"* ]] || [[ "$output" == *"instance_id"* ]]; then
    echo "Successfully retrieved instance information with TLS encryption!"
else
    echo "Failed to get instance information with TLS"
    exit 1
fi

# Test 2: List Instances with TLS
echo "===== Test 2: List Instances with TLS ====="
echo "Running: snooze provider azure list-instances"
output=$($SNOOZE_BIN provider azure list-instances 2>&1 || true)
echo "$output"
if [[ "$output" == *"instances"* ]] || [[ "$output" == *"$AZURE_VM_NAME"* ]]; then
    echo "Successfully listed instances with TLS encryption!"
else
    echo "Failed to list instances with TLS"
    exit 1
fi

# Cleanup
echo "Cleaning up test certificate directory..."
if [ "${KEEP_CERTS}" != "true" ]; then
    rm -rf "$CERT_DIR"/*
    echo "Certificates cleaned up"
else
    echo "Keeping certificates for inspection (KEEP_CERTS=true)"
fi

echo "All Azure plugin TLS tests completed successfully!"