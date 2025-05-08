#!/bin/bash
# This script builds all plugins or a specific plugin

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Plugins directory
PLUGINS_DIR="$PROJECT_ROOT/plugins"

# Output directory
OUTPUT_DIR="$PROJECT_ROOT/bin/plugins"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Function to build a single plugin
build_plugin() {
    local plugin=$1
    local plugin_dir="$PLUGINS_DIR/$plugin"
    local output_path="$OUTPUT_DIR/$plugin"
    
    echo "Building plugin: $plugin"
    
    # Check if plugin directory exists
    if [ ! -d "$plugin_dir" ]; then
        echo "Error: Plugin directory $plugin_dir does not exist"
        return 1
    fi
    
    # Check if plugin main.go exists
    if [ ! -f "$plugin_dir/main.go" ]; then
        echo "Error: Plugin main.go not found in $plugin_dir"
        return 1
    fi
    
    # Build the plugin
    # Check if plugin has its own go.mod
    if [ -f "$plugin_dir/go.mod" ]; then
        # Build using module-aware mode
        (cd "$plugin_dir" && go build -v -o "$output_path" .)
    else
        # Build using standard mode
        (cd "$PROJECT_ROOT" && go build -v -o "$output_path" "./plugins/$plugin")
    fi
    
    echo "Plugin $plugin built successfully to $output_path"
    
    # Make the plugin executable
    chmod +x "$output_path"
}

# Function to build all plugins
build_all_plugins() {
    # Get all plugin directories
    local plugins=$(find "$PLUGINS_DIR" -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
    
    for plugin in $plugins; do
        build_plugin "$plugin"
    done
}

# Main execution
if [ $# -eq 0 ]; then
    # No arguments, build all plugins
    echo "Building all plugins..."
    build_all_plugins
else
    # Build specific plugin
    build_plugin "$1"
fi

echo "Plugin build completed successfully!"