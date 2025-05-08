#!/bin/bash

# Script to run all tests for Snoozebot
# Usage: ./run_all_tests.sh [--no-unit] [--no-integration] [--no-security] [--live]

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse command line arguments
RUN_UNIT=1
RUN_INTEGRATION=1
RUN_SECURITY=1
RUN_LIVE=0

for arg in "$@"; do
  case $arg in
    --no-unit)
      RUN_UNIT=0
      shift
      ;;
    --no-integration)
      RUN_INTEGRATION=0
      shift
      ;;
    --no-security)
      RUN_SECURITY=0
      shift
      ;;
    --live)
      RUN_LIVE=1
      shift
      ;;
  esac
done

# Function to print section header
print_header() {
  echo ""
  echo -e "${BLUE}=========================================="
  echo "  $1"
  echo "==========================================${NC}"
  echo ""
}

# Function to run tests with retry
run_tests() {
  local test_command=$1
  local test_name=$2
  local max_retries=${3:-1}
  local retry_count=0
  
  while [ $retry_count -lt $max_retries ]; do
    echo -e "${YELLOW}Running $test_name (attempt $(($retry_count + 1))/${max_retries})...${NC}"
    
    if eval "$test_command"; then
      echo -e "${GREEN}$test_name completed successfully${NC}"
      return 0
    else
      echo -e "${RED}$test_name failed${NC}"
      retry_count=$((retry_count + 1))
      
      if [ $retry_count -lt $max_retries ]; then
        echo "Retrying in 3 seconds..."
        sleep 3
      fi
    fi
  done
  
  return 1
}

# Main execution
print_header "Starting Snoozebot Tests"

FAILURES=0

# Run unit tests if enabled
if [ $RUN_UNIT -eq 1 ]; then
  print_header "Running Unit Tests"
  if ! run_tests "go test -v ./pkg/... ./agent/..." "Unit tests"; then
    FAILURES=$((FAILURES + 1))
  fi
fi

# Run integration tests if enabled
if [ $RUN_INTEGRATION -eq 1 ]; then
  print_header "Running Integration Tests"
  
  # Use mock or live mode for provider tests
  if [ $RUN_LIVE -eq 1 ]; then
    echo -e "${YELLOW}Running provider tests in LIVE mode${NC}"
    echo "Make sure you have set up credentials for the cloud providers"
    
    if ! run_tests "./scripts/test_providers.sh all live" "Provider tests (live mode)"; then
      FAILURES=$((FAILURES + 1))
    fi
  else
    echo -e "${YELLOW}Running provider tests in MOCK mode${NC}"
    if ! run_tests "./scripts/test_providers.sh all mock" "Provider tests (mock mode)"; then
      FAILURES=$((FAILURES + 1))
    fi
  fi
  
  # Run other integration tests
  SNOOZEBOT_RUN_INTEGRATION=true
  if ! run_tests "go test -v ./test/integration/..." "Integration tests"; then
    FAILURES=$((FAILURES + 1))
  fi
fi

# Run security checks if enabled
if [ $RUN_SECURITY -eq 1 ]; then
  print_header "Running Security Checks"
  if ! run_tests "./scripts/security_check.sh" "Security checks"; then
    FAILURES=$((FAILURES + 1))
  fi
fi

# Print summary
print_header "Test Summary"

if [ $FAILURES -eq 0 ]; then
  echo -e "${GREEN}All tests passed successfully!${NC}"
else
  echo -e "${RED}$FAILURES test groups failed${NC}"
  exit 1
fi