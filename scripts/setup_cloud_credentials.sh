#!/bin/bash
# Main script to set up cloud provider credentials for Snoozebot
# This script guides you through setting up credentials for AWS, Azure, and GCP

set -e

# Text colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Determine script directory 
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo -e "${BLUE}===== Snoozebot Cloud Provider Credentials Setup =====${NC}"
echo "This script will help you set up credentials for cloud providers"
echo "to use with Snoozebot."
echo ""
echo "Available providers:"
echo "1) AWS"
echo "2) Azure"
echo "3) GCP"
echo "4) All providers"
echo "5) Exit"
echo ""
echo -e "Please enter your choice (1-5): ${YELLOW}"
read -r CHOICE
echo -e "${NC}"

setup_aws() {
    echo -e "${BLUE}===== AWS Credentials Setup =====${NC}"
    
    # Check if AWS setup script exists
    AWS_SCRIPT="$SCRIPT_DIR/setup_aws_credentials.sh"
    if [ ! -f "$AWS_SCRIPT" ]; then
        echo -e "${RED}Error: AWS setup script not found at $AWS_SCRIPT${NC}"
        return 1
    fi
    
    # Make the script executable
    chmod +x "$AWS_SCRIPT"
    
    # Run the AWS setup script
    "$AWS_SCRIPT"
    
    echo -e "${GREEN}AWS setup complete!${NC}"
    echo ""
}

setup_azure() {
    echo -e "${BLUE}===== Azure Credentials Setup =====${NC}"
    
    # Check if Azure setup script exists
    AZURE_SCRIPT="$SCRIPT_DIR/setup_azure_credentials.sh"
    if [ ! -f "$AZURE_SCRIPT" ]; then
        echo -e "${RED}Error: Azure setup script not found at $AZURE_SCRIPT${NC}"
        return 1
    fi
    
    # Make the script executable
    chmod +x "$AZURE_SCRIPT"
    
    # Run the Azure setup script
    "$AZURE_SCRIPT"
    
    echo -e "${GREEN}Azure setup complete!${NC}"
    echo ""
}

setup_gcp() {
    echo -e "${BLUE}===== GCP Credentials Setup =====${NC}"
    
    # Check if GCP setup script exists
    GCP_SCRIPT="$SCRIPT_DIR/setup_gcp_credentials.sh"
    if [ ! -f "$GCP_SCRIPT" ]; then
        echo -e "${RED}Error: GCP setup script not found at $GCP_SCRIPT${NC}"
        return 1
    fi
    
    # Make the script executable
    chmod +x "$GCP_SCRIPT"
    
    # Run the GCP setup script
    "$GCP_SCRIPT"
    
    echo -e "${GREEN}GCP setup complete!${NC}"
    echo ""
}

case $CHOICE in
    1)
        setup_aws
        ;;
    2)
        setup_azure
        ;;
    3)
        setup_gcp
        ;;
    4)
        setup_aws
        setup_azure
        setup_gcp
        ;;
    5)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid choice. Exiting.${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}===== All Selected Credentials Setup Complete =====${NC}"
echo ""
echo "You can now use these credentials with Snoozebot for cloud provider testing."
echo "Refer to the documentation for more details on how to configure Snoozebot"
echo "to use these credentials."
echo ""
echo "Documentation files:"
echo "- AWS: docs/AWS_CREDENTIALS_SETUP.md"
echo "- Azure: docs/AZURE_CREDENTIALS_SETUP.md"
echo "- GCP: docs/GCP_CREDENTIALS_SETUP.md"
echo ""
echo "Thank you for using Snoozebot!"