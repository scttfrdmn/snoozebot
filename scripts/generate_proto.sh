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

# Clean the generated directory first
rm -rf "$GEN_DIR"/*

# Generate the Go code
protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$GEN_DIR" --go_opt=paths=source_relative \
  --go-grpc_out="$GEN_DIR" --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/agent.proto"

echo "Generated gRPC code successfully!"