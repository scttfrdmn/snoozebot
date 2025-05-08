#!/bin/bash
# Script to test the Azure plugin functionality with real Azure credentials

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Binary paths
PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/azure"
SNOOZE_BIN="$PROJECT_ROOT/bin/snooze"

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
    echo ""
    echo "Additionally, you need to authenticate to Azure using one of these methods:"
    echo "1. Azure CLI: Run 'az login' before this script"
    echo "2. Service Principal: Set AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET"
    echo "3. Managed Identity: Ensure your environment has access to a managed identity"
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

echo "Starting Azure plugin test with real credentials:"
echo "AZURE_SUBSCRIPTION_ID: $AZURE_SUBSCRIPTION_ID"
echo "AZURE_RESOURCE_GROUP: $AZURE_RESOURCE_GROUP" 
echo "AZURE_VM_NAME: $AZURE_VM_NAME"
echo "AZURE_LOCATION: $AZURE_LOCATION"

# Test 1: Basic Plugin Functionality
echo "===== Test 1: Basic Plugin Functionality ====="
echo "Testing direct plugin execution (logs only)..."
output=$($PLUGIN_BIN 2>&1 || true)
if [[ "$output" == *"azure-provider"* ]]; then
    echo "Plugin loads successfully!"
else
    echo "Plugin self-test failed. Output:"
    echo "$output"
    exit 1
fi

# Test 2: Get Instance Information
echo "===== Test 2: Get Instance Information ====="
echo "Running: snooze provider azure get-info"
output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
echo "$output"
if [[ "$output" == *"$AZURE_VM_NAME"* ]] || [[ "$output" == *"instance_id"* ]]; then
    echo "Successfully retrieved instance information!"
else
    echo "Failed to get instance information"
    exit 1
fi

# Test 3: List Instances
echo "===== Test 3: List Instances ====="
echo "Running: snooze provider azure list-instances"
output=$($SNOOZE_BIN provider azure list-instances 2>&1 || true)
echo "$output"
if [[ "$output" == *"instances"* ]] || [[ "$output" == *"$AZURE_VM_NAME"* ]]; then
    echo "Successfully listed instances!"
else
    echo "Failed to list instances"
    exit 1
fi

# Test 4: Stop and Start Instance (OPTIONAL - uncomment if you want to test actual VM operations)
# WARNING: This will stop and start the actual VM
if [ "${TEST_VM_OPERATIONS}" == "true" ]; then
    echo "===== Test 4: Stop and Start Instance ====="
    echo "WARNING: This will STOP your actual Azure VM!"
    echo "Press Ctrl+C now if you don't want to proceed"
    sleep 5
    
    echo "Stopping VM: $AZURE_VM_NAME"
    output=$($SNOOZE_BIN provider azure stop-instance 2>&1 || true)
    echo "$output"
    if [[ "$output" == *"error"* ]]; then
        echo "Failed to stop VM"
        exit 1
    else
        echo "VM stop initiated, waiting 60 seconds..."
        sleep 60
    fi
    
    echo "Getting VM status (should be stopped):"
    output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
    echo "$output"
    
    echo "Starting VM: $AZURE_VM_NAME"
    output=$($SNOOZE_BIN provider azure start-instance 2>&1 || true)
    echo "$output"
    if [[ "$output" == *"error"* ]]; then
        echo "Failed to start VM"
        exit 1
    else
        echo "VM start initiated, waiting 60 seconds..."
        sleep 60
    fi
    
    echo "Getting VM status (should be running):"
    output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
    echo "$output"
    
    echo "Successfully stopped and started VM!"
else
    echo "===== Test 4: Stop and Start Instance ====="
    echo "SKIPPED - Set TEST_VM_OPERATIONS=true to test VM operations"
fi

echo "All Azure plugin tests completed successfully!"