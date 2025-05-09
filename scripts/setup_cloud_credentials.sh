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
    echo "AWS credentials are typically set up through the AWS CLI."
    echo ""
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        echo -e "${RED}Error: AWS CLI not found. Please install it first.${NC}"
        echo "Installation instructions: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
        return 1
    fi
    
    # Check if AWS credentials already exist
    if [ -f "$HOME/.aws/credentials" ]; then
        echo -e "${YELLOW}AWS credentials file already exists.${NC}"
        echo "Current profiles:"
        aws configure list-profiles
        echo ""
    else
        echo -e "${YELLOW}AWS credentials file not found. Let's set it up.${NC}"
    fi
    
    # Offer to set up new profile
    echo "Would you like to set up a new AWS profile for Snoozebot? (y/N):"
    read -r SETUP_AWS
    
    if [[ "$SETUP_AWS" =~ ^[Yy].* ]]; then
        echo ""
        echo "Please enter a name for the AWS profile (default: snoozebot):"
        read -r AWS_PROFILE_NAME
        
        if [ -z "$AWS_PROFILE_NAME" ]; then
            AWS_PROFILE_NAME="snoozebot"
        fi
        
        echo "Setting up AWS profile: $AWS_PROFILE_NAME"
        aws configure --profile "$AWS_PROFILE_NAME"
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}AWS profile '$AWS_PROFILE_NAME' created successfully!${NC}"
            
            # Create Snoozebot AWS config
            CONFIG_DIR="$HOME/.config/snoozebot"
            mkdir -p "$CONFIG_DIR"
            
            CONFIG_FILE="$CONFIG_DIR/aws-config.json"
            cat > "$CONFIG_FILE" << EOF
{
  "provider": "aws",
  "credentials": {
    "profileName": "$AWS_PROFILE_NAME"
  }
}
EOF
            echo "Snoozebot AWS configuration created at: $CONFIG_FILE"
            
            # Set environment variable
            export AWS_PROFILE="$AWS_PROFILE_NAME"
            
            echo ""
            echo "To use these credentials with Snoozebot, add the following to your shell profile:"
            echo -e "${YELLOW}export AWS_PROFILE=$AWS_PROFILE_NAME${NC}"
        else
            echo -e "${RED}Failed to set up AWS profile.${NC}"
        fi
    else
        echo "Skipping AWS credentials setup."
    fi
    
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
echo "- AWS: README.md"
echo "- Azure: docs/AZURE_CREDENTIALS_SETUP.md"
echo "- GCP: docs/GCP_CREDENTIALS_SETUP.md"
echo ""
echo "Thank you for using Snoozebot!"