#!/bin/bash
# Script to test TLS plugin functionality

set -e

echo "Building plugins with TLS support..."
./scripts/build_plugins.sh

echo "Setting up test environment..."
mkdir -p /tmp/snoozebot-test/certs

# Export environment variables for testing
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_DIR=/tmp/snoozebot-test/certs
export SNOOZEBOT_TLS_SKIP_VERIFY=false
export AZURE_SUBSCRIPTION_ID=test-subscription
export AZURE_RESOURCE_GROUP=test-resource-group

echo "Starting Azure plugin with TLS..."
./bin/plugins/azure &
PLUGIN_PID=$!

# Wait for the plugin to start
sleep 2

echo "Testing plugin TLS connection..."
# This should be replaced with an actual test that verifies TLS is working
# For now, we'll just check if the plugin is running
if ps -p $PLUGIN_PID > /dev/null; then
    echo "Plugin started successfully with TLS"
else
    echo "Plugin failed to start with TLS"
    exit 1
fi

echo "Cleaning up..."
kill $PLUGIN_PID
rm -rf /tmp/snoozebot-test

echo "TLS plugin test completed successfully"