#!/bin/bash
# Setup Azure credentials for Snoozebot
# This script creates a service principal and generates the necessary credentials file

set -e

# Text colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Azure Credentials Setup for Snoozebot ===${NC}"
echo "This script will help you set up Azure credentials for Snoozebot."
echo "Prerequisites: Azure CLI installed and logged in."
echo ""

# Check if Azure CLI is installed
if ! command -v az &> /dev/null; then
    echo -e "${RED}Error: Azure CLI not found. Please install it first.${NC}"
    echo "Installation instructions: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit 1
fi

# Check if logged in
echo "Checking if you're logged in to Azure..."
if ! az account show &> /dev/null; then
    echo -e "${YELLOW}You're not logged in to Azure. Please log in now.${NC}"
    az login
    if [ $? -ne 0 ]; then
        echo -e "${RED}Login failed. Please try again manually with 'az login'.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}Successfully logged in!${NC}"

# List subscriptions
echo "Getting your Azure subscriptions..."
az account list --output table

# Get subscription ID
echo ""
echo "Please enter the Subscription ID you want to use (copy from the table above):"
read -r SUBSCRIPTION_ID

# Validate subscription
if ! az account show --subscription "$SUBSCRIPTION_ID" &> /dev/null; then
    echo -e "${RED}Error: Invalid subscription ID. Please check and try again.${NC}"
    exit 1
fi

# Set the active subscription
echo "Setting the active subscription..."
az account set --subscription "$SUBSCRIPTION_ID"

# Create service principal
echo ""
echo -e "${YELLOW}Now creating a service principal with Contributor role...${NC}"
echo "This will give Snoozebot the ability to manage resources in your subscription."

SP_NAME="Snoozebot-ServicePrincipal-$(date +%Y%m%d%H%M%S)"
echo "Service principal name: $SP_NAME"

SP_RESPONSE=$(az ad sp create-for-rbac --name "$SP_NAME" --role "Contributor" --scopes "/subscriptions/$SUBSCRIPTION_ID" --output json)

# Extract credentials
CLIENT_ID=$(echo $SP_RESPONSE | jq -r '.appId')
CLIENT_SECRET=$(echo $SP_RESPONSE | jq -r '.password')
TENANT_ID=$(echo $SP_RESPONSE | jq -r '.tenant')

# Create credentials file
AZURE_DIR="$HOME/.azure"
CREDENTIALS_FILE="$AZURE_DIR/snoozebot-credentials.json"

# Ensure directory exists
mkdir -p "$AZURE_DIR"

# Create the credentials file
cat > "$CREDENTIALS_FILE" << EOF
{
  "clientId": "$CLIENT_ID",
  "clientSecret": "$CLIENT_SECRET",
  "tenantId": "$TENANT_ID",
  "subscriptionId": "$SUBSCRIPTION_ID"
}
EOF

# Set correct permissions
chmod 600 "$CREDENTIALS_FILE"

echo ""
echo -e "${GREEN}Credentials file created at: $CREDENTIALS_FILE${NC}"
echo "Please keep this file secure, as it contains sensitive information."

# Create a Snoozebot configuration for Azure
CONFIG_DIR="$HOME/.config/snoozebot"
mkdir -p "$CONFIG_DIR"

CONFIG_FILE="$CONFIG_DIR/azure-config.json"
cat > "$CONFIG_FILE" << EOF
{
  "provider": "azure",
  "credentials": {
    "profileName": "snoozebot",
    "credentialsFile": "$CREDENTIALS_FILE"
  }
}
EOF

# Setting environment variable for the session
export AZURE_PROFILE=snoozebot

echo ""
echo -e "${GREEN}=== Setup Complete! ===${NC}"
echo ""
echo "Your Azure credentials have been set up successfully."
echo "Configuration file created at: $CONFIG_FILE"
echo ""
echo "To use these credentials for tests, add the following to your shell profile:"
echo -e "${YELLOW}export AZURE_PROFILE=snoozebot${NC}"
echo ""
echo "To use these credentials with Snoozebot, configure it to use the profile 'snoozebot'"
echo "or the credentials file at: $CREDENTIALS_FILE"
echo ""
echo -e "${RED}⚠️  IMPORTANT SECURITY NOTE ⚠️${NC}"
echo "The credentials file contains sensitive information."
echo "Make sure to keep it secure and not share it with anyone."
echo "You should also consider setting up more granular permissions in a production environment."