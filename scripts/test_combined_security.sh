#!/bin/bash
# Script to test all security features combined (TLS, Signatures, Authentication)

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Binary paths
PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/azure"
SNOOZE_BIN="$PROJECT_ROOT/bin/snooze"
SNOOZESIGN_BIN="$PROJECT_ROOT/bin/snoozesign"

# Configuration directories
TEST_DIR="$PROJECT_ROOT/test/combined_security"
CERT_DIR="$TEST_DIR/certs"
SIG_DIR="$TEST_DIR/signatures"
KEY_DIR="$TEST_DIR/keys"
CONFIG_DIR="$TEST_DIR/config"
AUTH_FILE="$CONFIG_DIR/auth.json"

# Check if binaries exist
if [ ! -f "$PLUGIN_BIN" ]; then
    echo "Azure plugin not found. Building..."
    (cd "$PROJECT_ROOT" && ./scripts/build_plugins.sh azure)
fi

if [ ! -f "$SNOOZE_BIN" ]; then
    echo "Snooze CLI not found. Building..."
    (cd "$PROJECT_ROOT" && go build -o "$SNOOZE_BIN" ./cmd/snooze)
fi

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

# Create test directories
mkdir -p "$CERT_DIR" "$SIG_DIR" "$KEY_DIR" "$CONFIG_DIR"
echo "Using test directory: $TEST_DIR"

# Enable all security features
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_DIR="$CERT_DIR"
export SNOOZEBOT_SIGNATURE_ENABLED=true
export SNOOZEBOT_SIGNATURE_DIR="$SIG_DIR"
export SNOOZEBOT_AUTH_ENABLED=true
export SNOOZEBOT_AUTH_CONFIG="$AUTH_FILE"

echo "Starting combined security features test with real credentials:"
echo "AZURE_SUBSCRIPTION_ID: $AZURE_SUBSCRIPTION_ID"
echo "AZURE_RESOURCE_GROUP: $AZURE_RESOURCE_GROUP" 
echo "AZURE_VM_NAME: $AZURE_VM_NAME"
echo "AZURE_LOCATION: $AZURE_LOCATION"
echo "SNOOZEBOT_TLS_ENABLED: $SNOOZEBOT_TLS_ENABLED"
echo "SNOOZEBOT_TLS_CERT_DIR: $SNOOZEBOT_TLS_CERT_DIR"
echo "SNOOZEBOT_SIGNATURE_ENABLED: $SNOOZEBOT_SIGNATURE_ENABLED"
echo "SNOOZEBOT_SIGNATURE_DIR: $SNOOZEBOT_SIGNATURE_DIR"
echo "SNOOZEBOT_AUTH_ENABLED: $SNOOZEBOT_AUTH_ENABLED"
echo "SNOOZEBOT_AUTH_CONFIG: $SNOOZEBOT_AUTH_CONFIG"

# Step 1: Generate API key for authentication
echo "===== Step 1: Generate API key for authentication ====="
# Create a temporary Go program to generate API keys
API_KEY_GEN="$TEST_DIR/gen_api_key.go"
mkdir -p "$(dirname "$API_KEY_GEN")"
cat > "$API_KEY_GEN" << 'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/scttfrdmn/snoozebot/pkg/plugin/auth"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run gen_api_key.go <auth_file>")
		os.Exit(1)
	}

	authFile := os.Args[1]
	
	// Create a new API key
	apiKey, err := auth.GenerateAPIKey("admin", []string{"admin"})
	if err != nil {
		fmt.Printf("Failed to generate API key: %v\n", err)
		os.Exit(1)
	}

	// Create the auth config file
	authConfig := auth.PluginAuthConfig{
		APIKeys: []auth.APIKey{*apiKey},
	}

	// Save the config
	data, err := json.MarshalIndent(authConfig, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal auth config: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(authFile, data, 0600)
	if err != nil {
		fmt.Printf("Failed to write auth config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("API key generated successfully: %s\n", apiKey.Key)
}
EOF

# Compile and run the key generator
go run "$API_KEY_GEN" "$AUTH_FILE"
API_KEY=$(cat "$AUTH_FILE" | grep "Key" | cut -d '"' -f 4)
echo "Generated API key: $API_KEY"
export SNOOZEBOT_API_KEY="$API_KEY"

# Step 2: Generate a signing key and sign the plugin
echo "===== Step 2: Generate signing key and sign the plugin ====="
KEY_NAME="test-combined-key"
$SNOOZESIGN_BIN -generate-key -key-name="$KEY_NAME" -key-dir="$KEY_DIR" -validity=30

# Get the key ID
KEY_ID=$(ls "$KEY_DIR" | grep -v .pub | head -1 | sed 's/\.key$//')
echo "Generated key ID: $KEY_ID"

# Sign the Azure plugin
$SNOOZESIGN_BIN -sign -plugin="$PLUGIN_BIN" -key-id="$KEY_ID" -key-dir="$KEY_DIR" -signature-dir="$SIG_DIR"

# Verify the signature
$SNOOZESIGN_BIN -verify -plugin="$PLUGIN_BIN" -signature-dir="$SIG_DIR"

# Step 3: Test with all security features enabled
echo "===== Step 3: Test with all security features enabled ====="
echo "Running: snooze provider azure get-info"
start_time=$(date +%s.%N)
output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
end_time=$(date +%s.%N)
execution_time=$(echo "$end_time - $start_time" | bc)
echo "$output"

echo "Command execution time: $execution_time seconds"

if [[ "$output" == *"$AZURE_VM_NAME"* ]] || [[ "$output" == *"instance_id"* ]]; then
    echo "Successfully retrieved instance information with all security features enabled!"
else
    echo "Failed to get instance information with security features"
    exit 1
fi

# Step 4: Benchmark performance of security features
echo "===== Step 4: Benchmark security features ====="
echo "Running 5 iterations to benchmark performance..."

# Disable security features for baseline
unset SNOOZEBOT_TLS_ENABLED
unset SNOOZEBOT_SIGNATURE_ENABLED 
unset SNOOZEBOT_AUTH_ENABLED
unset SNOOZEBOT_API_KEY

echo "Baseline (no security features):"
total_time=0
for i in {1..5}; do
    start_time=$(date +%s.%N)
    $SNOOZE_BIN provider azure get-info > /dev/null 2>&1 || true
    end_time=$(date +%s.%N)
    execution_time=$(echo "$end_time - $start_time" | bc)
    total_time=$(echo "$total_time + $execution_time" | bc)
    printf "  Run %d: %.3f seconds\n" $i $execution_time
done
avg_time=$(echo "$total_time / 5" | bc -l)
printf "  Average: %.3f seconds\n" $avg_time
baseline_time=$avg_time

# Enable TLS only
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_DIR="$CERT_DIR"
unset SNOOZEBOT_SIGNATURE_ENABLED 
unset SNOOZEBOT_AUTH_ENABLED

echo "TLS only:"
total_time=0
for i in {1..5}; do
    start_time=$(date +%s.%N)
    $SNOOZE_BIN provider azure get-info > /dev/null 2>&1 || true
    end_time=$(date +%s.%N)
    execution_time=$(echo "$end_time - $start_time" | bc)
    total_time=$(echo "$total_time + $execution_time" | bc)
    printf "  Run %d: %.3f seconds\n" $i $execution_time
done
avg_time=$(echo "$total_time / 5" | bc -l)
printf "  Average: %.3f seconds\n" $avg_time
tls_overhead=$(echo "($avg_time - $baseline_time) / $baseline_time * 100" | bc -l)
printf "  Overhead: %.1f%%\n" $tls_overhead

# Enable signatures only
unset SNOOZEBOT_TLS_ENABLED
export SNOOZEBOT_SIGNATURE_ENABLED=true
export SNOOZEBOT_SIGNATURE_DIR="$SIG_DIR"
unset SNOOZEBOT_AUTH_ENABLED

echo "Signatures only:"
total_time=0
for i in {1..5}; do
    start_time=$(date +%s.%N)
    $SNOOZE_BIN provider azure get-info > /dev/null 2>&1 || true
    end_time=$(date +%s.%N)
    execution_time=$(echo "$end_time - $start_time" | bc)
    total_time=$(echo "$total_time + $execution_time" | bc)
    printf "  Run %d: %.3f seconds\n" $i $execution_time
done
avg_time=$(echo "$total_time / 5" | bc -l)
printf "  Average: %.3f seconds\n" $avg_time
sig_overhead=$(echo "($avg_time - $baseline_time) / $baseline_time * 100" | bc -l)
printf "  Overhead: %.1f%%\n" $sig_overhead

# Enable authentication only
unset SNOOZEBOT_TLS_ENABLED
unset SNOOZEBOT_SIGNATURE_ENABLED
export SNOOZEBOT_AUTH_ENABLED=true
export SNOOZEBOT_AUTH_CONFIG="$AUTH_FILE"
export SNOOZEBOT_API_KEY="$API_KEY"

echo "Authentication only:"
total_time=0
for i in {1..5}; do
    start_time=$(date +%s.%N)
    $SNOOZE_BIN provider azure get-info > /dev/null 2>&1 || true
    end_time=$(date +%s.%N)
    execution_time=$(echo "$end_time - $start_time" | bc)
    total_time=$(echo "$total_time + $execution_time" | bc)
    printf "  Run %d: %.3f seconds\n" $i $execution_time
done
avg_time=$(echo "$total_time / 5" | bc -l)
printf "  Average: %.3f seconds\n" $avg_time
auth_overhead=$(echo "($avg_time - $baseline_time) / $baseline_time * 100" | bc -l)
printf "  Overhead: %.1f%%\n" $auth_overhead

# All security features enabled
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_DIR="$CERT_DIR"
export SNOOZEBOT_SIGNATURE_ENABLED=true
export SNOOZEBOT_SIGNATURE_DIR="$SIG_DIR"
export SNOOZEBOT_AUTH_ENABLED=true
export SNOOZEBOT_AUTH_CONFIG="$AUTH_FILE"
export SNOOZEBOT_API_KEY="$API_KEY"

echo "All security features enabled:"
total_time=0
for i in {1..5}; do
    start_time=$(date +%s.%N)
    $SNOOZE_BIN provider azure get-info > /dev/null 2>&1 || true
    end_time=$(date +%s.%N)
    execution_time=$(echo "$end_time - $start_time" | bc)
    total_time=$(echo "$total_time + $execution_time" | bc)
    printf "  Run %d: %.3f seconds\n" $i $execution_time
done
avg_time=$(echo "$total_time / 5" | bc -l)
printf "  Average: %.3f seconds\n" $avg_time
all_overhead=$(echo "($avg_time - $baseline_time) / $baseline_time * 100" | bc -l)
printf "  Overhead: %.1f%%\n" $all_overhead

echo "Security Feature Benchmark Summary:"
printf "  Baseline: %.3f seconds\n" $baseline_time
printf "  TLS Overhead: %.1f%%\n" $tls_overhead
printf "  Signature Overhead: %.1f%%\n" $sig_overhead
printf "  Authentication Overhead: %.1f%%\n" $auth_overhead
printf "  Combined Overhead: %.1f%%\n" $all_overhead

# Cleanup
if [ "${KEEP_TEST_FILES}" != "true" ]; then
    echo "Cleaning up test directories..."
    rm -rf "$TEST_DIR"
    echo "Test files cleaned up"
else
    echo "Keeping test files for inspection (KEEP_TEST_FILES=true)"
fi

echo "All combined security features tests completed successfully!"