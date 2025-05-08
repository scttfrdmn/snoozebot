#!/bin/bash
set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Protocol directory
PROTO_DIR="$PROJECT_ROOT/pkg/common/protocol/proto"

# Generated code directory
GEN_DIR="$PROJECT_ROOT/pkg/common/protocol/gen"

# Plugin proto directory
PLUGIN_PROTO_DIR="$PROJECT_ROOT/pkg/plugin/proto"

# Plugin generated code directory
PLUGIN_GEN_DIR="$PROJECT_ROOT/pkg/plugin"

# Clean the generated directories first
rm -rf "$GEN_DIR"/*
mkdir -p "$PLUGIN_GEN_DIR"

# Generate the Go code for agent proto
protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$GEN_DIR" --go_opt=paths=source_relative \
  --go-grpc_out="$GEN_DIR" --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/agent.proto"

# Generate the Go code for plugin proto
protoc \
  --proto_path="$PLUGIN_PROTO_DIR" \
  --go_out="$PLUGIN_GEN_DIR" --go_opt=paths=source_relative \
  --go-grpc_out="$PLUGIN_GEN_DIR" --go-grpc_opt=paths=source_relative \
  "$PLUGIN_PROTO_DIR/cloud_provider.proto"

echo "Generated gRPC code successfully!"