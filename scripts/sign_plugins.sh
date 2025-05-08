#!/bin/bash
# Script to sign all plugins in the plugins directory

set -e

SIGNATURE_DIR="/etc/snoozebot/signatures"
PLUGINS_DIR="./bin/plugins"
KEY_NAME="snoozebot-release-key"

# Create the signature directory if it doesn't exist
mkdir -p "$SIGNATURE_DIR"

# Check if snoozesign exists and build it if it doesn't
if [ ! -f "./bin/snoozesign" ]; then
    echo "Building snoozesign utility..."
    go build -o ./bin/snoozesign ./cmd/snoozesign
fi

# Generate a signing key if one doesn't exist
echo "Checking for existing signing keys..."
KEY_ID=$(./bin/snoozesign -list-keys -signature-dir="$SIGNATURE_DIR" | grep "ID:" | head -1 | awk '{print $2}')

if [ -z "$KEY_ID" ]; then
    echo "Generating a new signing key..."
    KEY_OUTPUT=$(./bin/snoozesign -generate-key -key-name="$KEY_NAME" -signature-dir="$SIGNATURE_DIR")
    KEY_ID=$(echo "$KEY_OUTPUT" | grep "ID:" | awk '{print $2}')
    echo "Generated key: $KEY_ID"
else
    echo "Using existing key: $KEY_ID"
fi

# Sign all plugins
echo "Signing plugins..."
PLUGIN_COUNT=0

for plugin in "$PLUGINS_DIR"/*; do
    if [ -f "$plugin" ] && [ -x "$plugin" ]; then
        plugin_name=$(basename "$plugin")
        echo "Signing plugin: $plugin_name"
        ./bin/snoozesign -sign -plugin="$plugin" -key-id="$KEY_ID" -signature-dir="$SIGNATURE_DIR"
        PLUGIN_COUNT=$((PLUGIN_COUNT + 1))
    fi
done

# Create a bundle directory if it doesn't exist
BUNDLE_DIR="./bin/bundles"
mkdir -p "$BUNDLE_DIR"

# Create plugin bundles for distribution
echo "Creating plugin bundles..."
for plugin in "$PLUGINS_DIR"/*; do
    if [ -f "$plugin" ] && [ -x "$plugin" ]; then
        plugin_name=$(basename "$plugin")
        echo "Creating bundle for plugin: $plugin_name"
        bundle_path="$BUNDLE_DIR/$plugin_name"
        ./bin/snoozesign -bundle -plugin="$plugin" -bundle-path="$bundle_path" -key-id="$KEY_ID" -signature-dir="$SIGNATURE_DIR"
    fi
done

echo "Signed $PLUGIN_COUNT plugins successfully."
echo "Plugin signatures stored in: $SIGNATURE_DIR"
echo "Plugin bundles stored in: $BUNDLE_DIR"

# Verify the signatures
echo "Verifying plugin signatures..."
for plugin in "$PLUGINS_DIR"/*; do
    if [ -f "$plugin" ] && [ -x "$plugin" ]; then
        plugin_name=$(basename "$plugin")
        echo "Verifying signature for: $plugin_name"
        ./bin/snoozesign -verify -plugin="$plugin" -signature-dir="$SIGNATURE_DIR"
    fi
done

echo "All plugin signatures verified successfully."