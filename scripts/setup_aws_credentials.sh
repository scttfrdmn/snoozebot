#!/bin/bash
# Setup AWS credentials for Snoozebot
# This script sets up AWS CLI profiles and configurations for Snoozebot

set -e

# Text colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== AWS Credentials Setup for Snoozebot ===${NC}"
echo "This script will help you set up AWS credentials for Snoozebot."
echo "Prerequisites: AWS CLI installed."
echo ""

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}Error: AWS CLI not found. Please install it first.${NC}"
    echo "Installation instructions: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
    exit 1
fi

# Check existing AWS config
if [ -f ~/.aws/credentials ]; then
    echo "Found existing AWS credentials file. Available profiles:"
    aws configure list-profiles
    echo ""
    
    echo -e "Do you want to use an existing profile? (y/N): ${YELLOW}"
    read -r USE_EXISTING
    echo -e "${NC}"
    
    if [[ "$USE_EXISTING" =~ ^[Yy].* ]]; then
        echo "Enter the name of the profile to use:"
        read -r PROFILE_NAME
        
        # Check if profile exists
        if ! aws configure list-profiles | grep -q "^$PROFILE_NAME$"; then
            echo -e "${RED}Error: Profile '$PROFILE_NAME' not found.${NC}"
            echo "Available profiles:"
            aws configure list-profiles
            exit 1
        fi
        
        echo -e "${GREEN}Using existing profile: $PROFILE_NAME${NC}"
    else
        echo "Let's create a new AWS profile."
        echo "Enter a name for the new profile (default: snoozebot):"
        read -r PROFILE_NAME
        
        if [ -z "$PROFILE_NAME" ]; then
            PROFILE_NAME="snoozebot"
        fi
        
        # Configure the new profile
        echo -e "${YELLOW}Setting up AWS profile: $PROFILE_NAME${NC}"
        echo "You'll be prompted to enter your AWS Access Key ID, Secret Access Key, default region, and output format."
        aws configure --profile "$PROFILE_NAME"
    fi
else
    echo "No AWS credentials file found. Let's create one."
    echo "Enter a name for the new profile (default: snoozebot):"
    read -r PROFILE_NAME
    
    if [ -z "$PROFILE_NAME" ]; then
        PROFILE_NAME="snoozebot"
    fi
    
    # Configure the new profile
    echo -e "${YELLOW}Setting up AWS profile: $PROFILE_NAME${NC}"
    echo "You'll be prompted to enter your AWS Access Key ID, Secret Access Key, default region, and output format."
    aws configure --profile "$PROFILE_NAME"
fi

# Test the credentials
echo "Testing AWS credentials..."
if aws sts get-caller-identity --profile "$PROFILE_NAME" > /dev/null 2>&1; then
    echo -e "${GREEN}AWS credentials are working!${NC}"
    
    # Get account info
    ACCOUNT_INFO=$(aws sts get-caller-identity --profile "$PROFILE_NAME" --output json)
    ACCOUNT_ID=$(echo "$ACCOUNT_INFO" | grep -o '"Account": "[^"]*' | cut -d'"' -f4)
    USER_ARN=$(echo "$ACCOUNT_INFO" | grep -o '"Arn": "[^"]*' | cut -d'"' -f4)
    
    echo "Connected to AWS account: $ACCOUNT_ID"
    echo "Using IAM identity: $USER_ARN"
else
    echo -e "${RED}Failed to validate AWS credentials. Please check your access keys and try again.${NC}"
    exit 1
fi

# Check for EC2 permissions
echo "Checking EC2 permissions..."
if aws ec2 describe-instances --profile "$PROFILE_NAME" --max-items 1 > /dev/null 2>&1; then
    echo -e "${GREEN}EC2 permissions confirmed!${NC}"
else
    echo -e "${YELLOW}Warning: Unable to list EC2 instances. The user may not have sufficient permissions.${NC}"
    echo "Snoozebot requires permissions to describe, start, and stop EC2 instances."
    echo "Please ensure the IAM user has the appropriate policies attached."
fi

# Create Snoozebot AWS configuration
CONFIG_DIR="$HOME/.config/snoozebot"
mkdir -p "$CONFIG_DIR"

# Get the default region from the AWS config
DEFAULT_REGION=$(aws configure get region --profile "$PROFILE_NAME")
if [ -z "$DEFAULT_REGION" ]; then
    DEFAULT_REGION="us-west-2"
fi

# Create config file
CONFIG_FILE="$CONFIG_DIR/aws-config.json"
cat > "$CONFIG_FILE" << EOF
{
  "provider": "aws",
  "credentials": {
    "profileName": "$PROFILE_NAME",
    "region": "$DEFAULT_REGION"
  }
}
EOF

# Set environment variable for the current session
export AWS_PROFILE="$PROFILE_NAME"

echo ""
echo -e "${GREEN}=== Setup Complete! ===${NC}"
echo ""
echo "Your AWS credentials have been set up successfully."
echo "Configuration file created at: $CONFIG_FILE"
echo ""
echo "To use these credentials for tests, add the following to your shell profile:"
echo -e "${YELLOW}export AWS_PROFILE=$PROFILE_NAME${NC}"
echo ""
echo "To use these credentials with Snoozebot, configure it to use the profile '$PROFILE_NAME'"
echo "or the configuration file at: $CONFIG_FILE"
echo ""

# Security recommendations
echo -e "${RED}⚠️  SECURITY RECOMMENDATIONS ⚠️${NC}"
echo "1. Use a dedicated IAM user with minimal permissions for Snoozebot"
echo "2. Regularly rotate your AWS access keys"
echo "3. Enable multi-factor authentication for your AWS users"
echo "4. Monitor your AWS usage with CloudTrail"
echo ""
echo "For more information, refer to the AWS_CREDENTIALS_SETUP.md document."