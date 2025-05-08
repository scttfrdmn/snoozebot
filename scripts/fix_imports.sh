#!/bin/bash
set -e

echo "Fixing import paths after username change..."

find . -name "*.go" -type f -exec sed -i '' 's|github.com/scottfridman/snoozebot|github.com/scttfrdmn/snoozebot|g' {} \;
find . -name "*.go" -type f -exec sed -i '' 's|github.com/scttfrdmn/snoozebot|github.com/scttfrdmn/snoozebot|g' {} \;

echo "Running go mod tidy..."
go mod tidy

echo "Import paths fixed!"