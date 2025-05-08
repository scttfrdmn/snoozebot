#!/bin/bash

# Script to run provider tests with various configurations
# Usage: ./test_providers.sh [provider] [mode]
# Examples:
#   ./test_providers.sh           # Run all providers in mock mode
#   ./test_providers.sh aws       # Run AWS provider in mock mode
#   ./test_providers.sh aws live  # Run AWS provider against real AWS
#   ./test_providers.sh all live  # Run all providers against real cloud providers

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

# Parse command line arguments
PROVIDER=${1:-all}
MODE=${2:-mock}

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to run mock tests for a provider
run_mock_tests() {
    local provider=$1
    echo -e "${BLUE}Running mock tests for ${provider}...${NC}"
    
    # Set environment variables for mock mode
    export SNOOZEBOT_LIVE_TESTS=false
    export SNOOZEBOT_MOCK_PROVIDER=true
    
    # Run the tests
    go test -v -count=1 ./test/providers/${provider}_test.go ./test/providers/mock_provider.go ./test/providers/provider_test.go
    
    echo -e "${GREEN}Mock tests for ${provider} completed successfully${NC}\n"
}

# Function to run live tests for a provider
run_live_tests() {
    local provider=$1
    echo -e "${YELLOW}Running live tests for ${provider}...${NC}"
    
    # Make sure we have the provider environment loaded
    if [ ! -f ".env.${provider}" ]; then
        echo -e "${RED}Error: Environment file .env.${provider} not found${NC}"
        echo "Run ./scripts/setup_test_env.sh ${provider} first"
        return 1
    fi
    
    # Load the provider environment
    source ".env.${provider}"
    
    # Set environment variables for live mode
    export SNOOZEBOT_LIVE_TESTS=true
    export SNOOZEBOT_MOCK_PROVIDER=false
    
    # Run the tests
    if go test -v -count=1 ./test/providers/${provider}_test.go ./test/providers/provider_test.go; then
        echo -e "${GREEN}Live tests for ${provider} completed successfully${NC}\n"
    else
        echo -e "${RED}Live tests for ${provider} failed${NC}\n"
        return 1
    fi
}

# Function to check if we have credentials for a provider
check_provider_credentials() {
    local provider=$1
    case "$provider" in
        aws)
            [ -n "$AWS_PROFILE" ] && return 0
            return 1
            ;;
        azure)
            [ -n "$AZURE_PROFILE" ] && [ -n "$AZURE_SUBSCRIPTION_ID" ] && [ -n "$AZURE_RESOURCE_GROUP" ] && return 0
            return 1
            ;;
        gcp)
            [ -n "$GCP_PROFILE" ] && [ -n "$PROJECT_ID" ] && [ -n "$GOOGLE_APPLICATION_CREDENTIALS" ] && return 0
            return 1
            ;;
        *)
            return 1
            ;;
    esac
}

# Main execution
case "$PROVIDER" in
    all)
        echo -e "${BLUE}Testing all providers in ${MODE} mode${NC}\n"
        
        # Run tests for each provider
        for provider in aws azure gcp; do
            if [ "$MODE" = "live" ]; then
                if check_provider_credentials "$provider"; then
                    run_live_tests "$provider" || echo -e "${RED}Live tests for ${provider} failed${NC}"
                else
                    echo -e "${YELLOW}Skipping live tests for ${provider} - credentials not found${NC}"
                fi
            else
                run_mock_tests "$provider"
            fi
        done
        ;;
    
    aws|azure|gcp)
        if [ "$MODE" = "live" ]; then
            if check_provider_credentials "$PROVIDER"; then
                run_live_tests "$PROVIDER"
            else
                echo -e "${RED}Error: Credentials not found for ${PROVIDER}${NC}"
                echo "Run ./scripts/setup_test_env.sh ${PROVIDER} first"
                exit 1
            fi
        else
            run_mock_tests "$PROVIDER"
        fi
        ;;
    
    *)
        echo -e "${RED}Error: Unknown provider ${PROVIDER}${NC}"
        echo "Usage: $0 [provider] [mode]"
        echo "  provider: aws, azure, gcp, all (default: all)"
        echo "  mode: mock, live (default: mock)"
        exit 1
        ;;
esac

echo -e "${GREEN}All tests completed${NC}"