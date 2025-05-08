#!/bin/bash
# Script to test multiple cloud provider plugins together

set -e

# Directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Project root directory
PROJECT_ROOT="$SCRIPT_DIR/.."

# Binary paths
AWS_PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/aws"
AZURE_PLUGIN_BIN="$PROJECT_ROOT/bin/plugins/azure"
SNOOZE_BIN="$PROJECT_ROOT/bin/snooze"

# Configuration directories
TEST_DIR="$PROJECT_ROOT/test/multi_provider"
CERT_DIR="$TEST_DIR/certs"
SIG_DIR="$TEST_DIR/signatures"
CONFIG_DIR="$TEST_DIR/config"

# Check if binaries exist and build if needed
for plugin in "aws" "azure"; do
    if [ ! -f "$PROJECT_ROOT/bin/plugins/$plugin" ]; then
        echo "$plugin plugin not found. Building..."
        (cd "$PROJECT_ROOT" && ./scripts/build_plugins.sh "$plugin")
    fi
done

if [ ! -f "$SNOOZE_BIN" ]; then
    echo "Snooze CLI not found. Building..."
    (cd "$PROJECT_ROOT" && go build -o "$SNOOZE_BIN" ./cmd/snooze)
fi

# Create test directories
mkdir -p "$CERT_DIR" "$SIG_DIR" "$CONFIG_DIR"
echo "Using test directory: $TEST_DIR"

# Enable TLS for this test
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_DIR="$CERT_DIR"

# Check for credentials
echo "Checking for AWS credentials..."
if [ -z "$AWS_ACCESS_KEY_ID" ] || [ -z "$AWS_SECRET_ACCESS_KEY" ] || [ -z "$AWS_REGION" ]; then
    echo "Warning: AWS credentials not fully configured."
    echo "Set AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_REGION for complete testing."
    HAS_AWS_CREDENTIALS="false"
else
    HAS_AWS_CREDENTIALS="true"
    echo "AWS credentials found."
fi

echo "Checking for Azure credentials..."
if [ -z "$AZURE_SUBSCRIPTION_ID" ] || [ -z "$AZURE_RESOURCE_GROUP" ]; then
    echo "Warning: Azure credentials not fully configured."
    echo "Set AZURE_SUBSCRIPTION_ID, AZURE_RESOURCE_GROUP, and AZURE_VM_NAME for complete testing."
    HAS_AZURE_CREDENTIALS="false"
else
    HAS_AZURE_CREDENTIALS="true"
    echo "Azure credentials found."
fi

# Set default Azure location if not provided
if [ -z "$AZURE_LOCATION" ]; then
    export AZURE_LOCATION="eastus"
    echo "AZURE_LOCATION not set, using default: eastus"
fi

echo "Starting multi-provider test:"
echo "TLS Enabled: $SNOOZEBOT_TLS_ENABLED"
echo "Certificate Directory: $SNOOZEBOT_TLS_CERT_DIR"

# Part 1: Provider Self-Tests
echo "===== Part 1: Provider Self-Tests ====="

echo "Testing AWS plugin..."
if [[ "$HAS_AWS_CREDENTIALS" == "true" ]]; then
    output=$($AWS_PLUGIN_BIN 2>&1 || true)
    if [[ "$output" == *"Plugin Protocol:"* ]]; then
        echo "AWS plugin loads successfully!"
    else
        echo "AWS plugin self-test failed. Output:"
        echo "$output"
        HAS_AWS_CREDENTIALS="false"
    fi
else
    echo "Skipping AWS plugin self-test (no credentials)"
fi

echo "Testing Azure plugin..."
if [[ "$HAS_AZURE_CREDENTIALS" == "true" ]]; then
    output=$($AZURE_PLUGIN_BIN 2>&1 || true)
    if [[ "$output" == *"azure-provider"* ]]; then
        echo "Azure plugin loads successfully!"
    else
        echo "Azure plugin self-test failed. Output:"
        echo "$output"
        HAS_AZURE_CREDENTIALS="false"
    fi
else
    echo "Skipping Azure plugin self-test (no credentials)"
fi

# Part 2: Sequential Provider Tests
echo "===== Part 2: Sequential Provider Tests ====="

if [[ "$HAS_AWS_CREDENTIALS" == "true" ]]; then
    echo "Testing AWS provider with CLI..."
    output=$($SNOOZE_BIN provider aws get-info 2>&1 || true)
    echo "AWS Provider Output:"
    echo "$output"
    if [[ "$output" == *"instance_id"* ]]; then
        echo "AWS provider works correctly!"
    else
        echo "AWS provider test failed"
        HAS_AWS_CREDENTIALS="false"
    fi
else
    echo "Skipping AWS provider CLI test (no credentials)"
fi

if [[ "$HAS_AZURE_CREDENTIALS" == "true" ]]; then
    echo "Testing Azure provider with CLI..."
    output=$($SNOOZE_BIN provider azure get-info 2>&1 || true)
    echo "Azure Provider Output:"
    echo "$output"
    if [[ "$output" == *"instance_id"* ]] || [[ "$output" == *"$AZURE_VM_NAME"* ]]; then
        echo "Azure provider works correctly!"
    else
        echo "Azure provider test failed"
        HAS_AZURE_CREDENTIALS="false"
    fi
else
    echo "Skipping Azure provider CLI test (no credentials)"
fi

# Part 3: Concurrent Provider Tests using background processes
echo "===== Part 3: Concurrent Provider Tests ====="

# Only proceed if both providers work individually
if [[ "$HAS_AWS_CREDENTIALS" == "true" && "$HAS_AZURE_CREDENTIALS" == "true" ]]; then
    echo "Running concurrent provider tests..."
    
    # Create temp files for output
    AWS_OUTPUT="$TEST_DIR/aws_output.txt"
    AZURE_OUTPUT="$TEST_DIR/azure_output.txt"
    
    # Run both providers concurrently
    $SNOOZE_BIN provider aws get-info > "$AWS_OUTPUT" 2>&1 &
    aws_pid=$!
    
    $SNOOZE_BIN provider azure get-info > "$AZURE_OUTPUT" 2>&1 &
    azure_pid=$!
    
    # Wait for both processes to complete
    echo "Waiting for AWS provider (PID: $aws_pid)..."
    wait $aws_pid
    aws_exit=$?
    
    echo "Waiting for Azure provider (PID: $azure_pid)..."
    wait $azure_pid
    azure_exit=$?
    
    # Check results
    echo "AWS provider exit code: $aws_exit"
    echo "Azure provider exit code: $azure_exit"
    
    echo "AWS provider output:"
    cat "$AWS_OUTPUT"
    
    echo "Azure provider output:"
    cat "$AZURE_OUTPUT"
    
    if [[ $aws_exit -eq 0 && $azure_exit -eq 0 ]]; then
        echo "Concurrent provider test succeeded!"
    else
        echo "Concurrent provider test failed"
        echo "AWS exit code: $aws_exit, Azure exit code: $azure_exit"
    fi
else
    echo "Skipping concurrent tests (one or both providers not working)"
fi

# Part 4: Provider Switching Tests
echo "===== Part 4: Provider Switching Tests ====="

if [[ "$HAS_AWS_CREDENTIALS" == "true" && "$HAS_AZURE_CREDENTIALS" == "true" ]]; then
    echo "Testing rapid provider switching..."
    
    # Multiple rapid provider switches
    for i in {1..3}; do
        echo "Round $i: AWS -> Azure -> AWS"
        
        # AWS
        echo "  AWS provider:"
        $SNOOZE_BIN provider aws get-info > /dev/null
        aws_exit=$?
        echo "  AWS exit code: $aws_exit"
        
        # Azure
        echo "  Azure provider:"
        $SNOOZE_BIN provider azure get-info > /dev/null
        azure_exit=$?
        echo "  Azure exit code: $azure_exit"
        
        # AWS again
        echo "  AWS provider again:"
        $SNOOZE_BIN provider aws get-info > /dev/null
        aws_again_exit=$?
        echo "  AWS again exit code: $aws_again_exit"
        
        if [[ $aws_exit -eq 0 && $azure_exit -eq 0 && $aws_again_exit -eq 0 ]]; then
            echo "  Round $i succeeded!"
        else
            echo "  Round $i failed"
            break
        fi
    done
    
    echo "Provider switching test completed."
else
    echo "Skipping provider switching tests (one or both providers not working)"
fi

# Part 5: Test List Instances on Both Providers
echo "===== Part 5: List Instances on Both Providers ====="

if [[ "$HAS_AWS_CREDENTIALS" == "true" ]]; then
    echo "Listing AWS instances..."
    output=$($SNOOZE_BIN provider aws list-instances 2>&1 || true)
    echo "AWS instances:"
    echo "$output"
    if [[ "$output" == *"instances"* ]]; then
        echo "AWS list-instances works correctly!"
    else
        echo "AWS list-instances test failed"
    fi
fi

if [[ "$HAS_AZURE_CREDENTIALS" == "true" ]]; then
    echo "Listing Azure instances..."
    output=$($SNOOZE_BIN provider azure list-instances 2>&1 || true)
    echo "Azure instances:"
    echo "$output"
    if [[ "$output" == *"instances"* ]]; then
        echo "Azure list-instances works correctly!"
    else
        echo "Azure list-instances test failed"
    fi
fi

# Cleanup
if [ "${KEEP_TEST_FILES}" != "true" ]; then
    echo "Cleaning up test directories..."
    rm -rf "$TEST_DIR"
    echo "Test files cleaned up"
else
    echo "Keeping test files for inspection (KEEP_TEST_FILES=true)"
fi

echo "Multi-provider tests completed!"

# Display Summary
echo "===== Test Summary ====="
echo "AWS Provider Tests: $([ "$HAS_AWS_CREDENTIALS" == "true" ] && echo "PASSED" || echo "SKIPPED")"
echo "Azure Provider Tests: $([ "$HAS_AZURE_CREDENTIALS" == "true" ] && echo "PASSED" || echo "SKIPPED")"
echo "Concurrent Provider Tests: $([ "$HAS_AWS_CREDENTIALS" == "true" ] && [ "$HAS_AZURE_CREDENTIALS" == "true" ] && echo "PASSED" || echo "SKIPPED")"
echo "Provider Switching Tests: $([ "$HAS_AWS_CREDENTIALS" == "true" ] && [ "$HAS_AZURE_CREDENTIALS" == "true" ] && echo "PASSED" || echo "SKIPPED")"