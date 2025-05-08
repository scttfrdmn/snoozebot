#!/bin/bash

# Script to build all plugins with versioning information
# Usage: 
#   ./build_versioned_plugins.sh
#   SNOOZEBOT_PLUGIN=aws ./build_versioned_plugins.sh  # Build only the AWS plugin

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
PLUGIN_BIN_DIR="$REPO_ROOT/bin/plugins"
VERSION=$(cat "$REPO_ROOT/VERSION")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Create output directory
mkdir -p "$PLUGIN_BIN_DIR"

# Build flags to include version information
BUILD_FLAGS=(
  "-ldflags"
  "-X github.com/scttfrdmn/snoozebot/pkg/plugin/version.BuildVersion=$VERSION"
  "-X github.com/scttfrdmn/snoozebot/pkg/plugin/version.BuildTimestamp=$TIMESTAMP"
  "-X github.com/scttfrdmn/snoozebot/pkg/plugin/version.GitCommit=$GIT_COMMIT"
)

echo "Building plugins with version information:"
echo "Version: $VERSION"
echo "Git Commit: $GIT_COMMIT"
echo "Build Timestamp: $TIMESTAMP"

# Function to build a specific plugin
build_plugin() {
    local plugin=$1
    local plugin_dir="$REPO_ROOT/plugins/$plugin"
    local output_path="$PLUGIN_BIN_DIR/$plugin"
    
    # Check for custom plugin
    if [ "$plugin" == "custom-provider" ]; then
        plugin_dir="$REPO_ROOT/examples/custom_plugin"
    fi
    
    # Check if plugin directory exists
    if [ ! -d "$plugin_dir" ]; then
        echo "Error: Plugin directory does not exist: $plugin_dir"
        return 1
    fi
    
    echo "Building $plugin plugin with version information..."
    cd "$plugin_dir"
    go build -o "$output_path" "${BUILD_FLAGS[@]}" .
    chmod +x "$output_path"
    echo "✅ $plugin plugin built successfully"
}

# Check if we should build a specific plugin
if [ -n "$SNOOZEBOT_PLUGIN" ]; then
    build_plugin "$SNOOZEBOT_PLUGIN"
    exit 0
fi

# Build all plugins
build_plugin "aws"
build_plugin "azure"
build_plugin "gcp"

# Build Custom Plugin (if it exists)
if [ -d "$REPO_ROOT/examples/custom_plugin" ]; then
    build_plugin "custom-provider"
fi

# Generate manifests for all plugins
echo "Generating plugin manifests..."
MANIFEST_DIR="$REPO_ROOT/bin/manifests"
mkdir -p "$MANIFEST_DIR"

# Build the version check tool if needed
if [ ! -f "$REPO_ROOT/bin/versioncheck" ]; then
    echo "Building version check tool..."
    cd "$REPO_ROOT"
    go build -o "$REPO_ROOT/bin/versioncheck" ./cmd/versioncheck
fi

# Generate manifests
"$REPO_ROOT/bin/versioncheck" manifest "$MANIFEST_DIR/aws.manifest.json" create "aws" "$VERSION" "AWS cloud provider plugin for Snoozebot"
"$REPO_ROOT/bin/versioncheck" manifest "$MANIFEST_DIR/azure.manifest.json" create "azure" "$VERSION" "Azure cloud provider plugin for Snoozebot" 
"$REPO_ROOT/bin/versioncheck" manifest "$MANIFEST_DIR/gcp.manifest.json" create "gcp" "$VERSION" "GCP cloud provider plugin for Snoozebot"

if [ -f "$PLUGIN_BIN_DIR/custom-provider" ]; then
    "$REPO_ROOT/bin/versioncheck" manifest "$MANIFEST_DIR/custom-provider.manifest.json" create "custom-provider" "$VERSION" "Custom provider example plugin for Snoozebot"
fi

echo "✅ All plugin manifests generated"

echo "✅ All plugins built successfully with version information"
echo "Plugin binaries are in: $PLUGIN_BIN_DIR"
echo "Plugin manifests are in: $MANIFEST_DIR"