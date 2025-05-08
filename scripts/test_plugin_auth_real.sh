#!/bin/bash
# Script to test the plugin authentication system with API keys

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Binary paths
PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/azure"
SNOOZE_BIN="$PROJECT_ROOT/bin/snooze"

# Config directory
CONFIG_DIR="$PROJECT_ROOT/test/config"
AUTH_FILE="$CONFIG_DIR/auth.json"

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

# Create config directory if it doesn't exist
mkdir -p "$CONFIG_DIR"
echo "Using config directory: $CONFIG_DIR"

# Enable authentication for the plugin
export SNOOZEBOT_AUTH_ENABLED=true
export SNOOZEBOT_AUTH_CONFIG="$AUTH_FILE"

echo "Starting plugin authentication test:"
echo "SNOOZEBOT_AUTH_ENABLED: $SNOOZEBOT_AUTH_ENABLED"
echo "SNOOZEBOT_AUTH_CONFIG: $SNOOZEBOT_AUTH_CONFIG"

# Step 1: Generate an API key
echo "===== Step 1: Generate an API key ====="
# Create a temporary Go program to generate API keys
API_KEY_GEN="$PROJECT_ROOT/test/gen_api_key.go"
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
		fmt.Println("Usage: go run gen_api_key.go <config_dir>")
		os.Exit(1)
	}

	configDir := os.Args[1]
	authFile := filepath.Join(configDir, "auth.json")

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
	fmt.Printf("Auth config written to: %s\n", authFile)
}
EOF

# Compile and run the key generator
go run "$API_KEY_GEN" "$CONFIG_DIR"
API_KEY=$(cat "$AUTH_FILE" | grep "Key" | cut -d '"' -f 4)
echo "Generated API key: $API_KEY"

# Step 2: Test using the API key
echo "===== Step 2: Test using the API key ====="
export SNOOZEBOT_API_KEY="$API_KEY"
echo "Using API key: $SNOOZEBOT_API_KEY"

# Test getting instance information with API key authentication
echo "Running: snooze provider azure get-info"
output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
echo "$output"
if [[ "$output" == *"error"* ]] && [[ "$output" == *"authentication"* ]]; then
    echo "Authentication error as expected - need to use real credentials too"
else
    echo "API key authentication test failed - unexpected output. Should get auth error."
    exit 1
fi

# Step 3: Test with invalid API key
echo "===== Step 3: Test with invalid API key ====="
export SNOOZEBOT_API_KEY="invalid-key"
echo "Using invalid API key: $SNOOZEBOT_API_KEY"

# Test getting instance information with invalid API key
echo "Running: snooze provider azure get-info"
output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
echo "$output"
if [[ "$output" == *"invalid API key"* ]] || [[ "$output" == *"authentication failed"* ]]; then
    echo "Successfully detected invalid API key!"
else
    echo "Invalid API key test failed - unexpected output"
    exit 1
fi

# Cleanup
if [ "${KEEP_CONFIG}" != "true" ]; then
    echo "Cleaning up test config directory..."
    rm -rf "$CONFIG_DIR"/*
    rm "$API_KEY_GEN"
    echo "Config cleaned up"
else
    echo "Keeping config for inspection (KEEP_CONFIG=true)"
fi

echo "All plugin authentication tests completed successfully!"