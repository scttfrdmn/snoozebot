#!/bin/bash
# Script to test the Azure plugin with signature verification using real Azure credentials

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Binary paths
PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/azure"
SNOOZE_BIN="$PROJECT_ROOT/bin/snooze"
SNOOZESIGN_BIN="$PROJECT_ROOT/bin/snoozesign"

# Signature and key directories
SIG_DIR="$PROJECT_ROOT/test/signatures"
KEY_DIR="$PROJECT_ROOT/test/keys"

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

# Check if snoozesign exists
if [ ! -f "$SNOOZESIGN_BIN" ]; then
    echo "Snoozesign tool not found. Building..."
    (cd "$PROJECT_ROOT" && go build -o "$SNOOZESIGN_BIN" ./cmd/snoozesign)
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

# Create signature and key directories if they don't exist
mkdir -p "$SIG_DIR"
mkdir -p "$KEY_DIR"
echo "Using signature directory: $SIG_DIR"
echo "Using key directory: $KEY_DIR"

# Enable signature verification for the plugin
export SNOOZEBOT_SIGNATURE_ENABLED=true
export SNOOZEBOT_SIGNATURE_DIR="$SIG_DIR"

echo "Starting Azure plugin signature verification test with real credentials:"
echo "AZURE_SUBSCRIPTION_ID: $AZURE_SUBSCRIPTION_ID"
echo "AZURE_RESOURCE_GROUP: $AZURE_RESOURCE_GROUP" 
echo "AZURE_VM_NAME: $AZURE_VM_NAME"
echo "AZURE_LOCATION: $AZURE_LOCATION"
echo "SNOOZEBOT_SIGNATURE_ENABLED: $SNOOZEBOT_SIGNATURE_ENABLED"
echo "SNOOZEBOT_SIGNATURE_DIR: $SNOOZEBOT_SIGNATURE_DIR"

# Step 1: Generate a test signing key
echo "===== Step 1: Generate a test signing key ====="
KEY_NAME="test-azure-key"
$SNOOZESIGN_BIN -generate-key -key-name="$KEY_NAME" -key-dir="$KEY_DIR" -validity=30

# Get the key ID from output
KEY_ID=$(ls "$KEY_DIR" | grep -v .pub | head -1 | sed 's/\.key$//')
echo "Generated key ID: $KEY_ID"

# Step 2: Sign the Azure plugin
echo "===== Step 2: Sign the Azure plugin ====="
$SNOOZESIGN_BIN -sign -plugin="$PLUGIN_BIN" -key-id="$KEY_ID" -key-dir="$KEY_DIR" -signature-dir="$SIG_DIR"

# Step 3: Verify the signature
echo "===== Step 3: Verify the signature ====="
$SNOOZESIGN_BIN -verify -plugin="$PLUGIN_BIN" -signature-dir="$SIG_DIR"

# Test 4: Get instance information with signature verification
echo "===== Test 4: Get Instance Information with Signature Verification ====="
echo "Running: snooze provider azure get-info"
output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
echo "$output"
if [[ "$output" == *"$AZURE_VM_NAME"* ]] || [[ "$output" == *"instance_id"* ]]; then
    echo "Successfully retrieved instance information with signature verification!"
else
    echo "Failed to get instance information with signature verification"
    exit 1
fi

# Test 5: List Instances with signature verification
echo "===== Test 5: List Instances with Signature Verification ====="
echo "Running: snooze provider azure list-instances"
output=$($SNOOZE_BIN provider azure list-instances 2>&1 || true)
echo "$output"
if [[ "$output" == *"instances"* ]] || [[ "$output" == *"$AZURE_VM_NAME"* ]]; then
    echo "Successfully listed instances with signature verification!"
else
    echo "Failed to list instances with signature verification"
    exit 1
fi

# Cleanup
if [ "${KEEP_SIGNATURES}" != "true" ]; then
    echo "Cleaning up test signature and key directories..."
    rm -rf "$SIG_DIR"/*
    rm -rf "$KEY_DIR"/*
    echo "Signatures and keys cleaned up"
else
    echo "Keeping signatures and keys for inspection (KEEP_SIGNATURES=true)"
fi

echo "All Azure plugin signature verification tests completed successfully!"