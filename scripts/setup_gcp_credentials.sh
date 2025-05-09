#!/bin/bash
# Setup GCP credentials for Snoozebot
# This script creates a service account and generates the necessary key file

set -e

# Text colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== GCP Credentials Setup for Snoozebot ===${NC}"
echo "This script will help you set up GCP credentials for Snoozebot."
echo "Prerequisites: Google Cloud SDK installed and initialized."
echo ""

# Check if gcloud CLI is installed
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}Error: Google Cloud SDK not found. Please install it first.${NC}"
    echo "Installation instructions: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Check if logged in
echo "Checking if you're logged in to GCP..."
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" &> /dev/null; then
    echo -e "${YELLOW}You're not logged in to GCP. Please log in now.${NC}"
    gcloud auth login
    if [ $? -ne 0 ]; then
        echo -e "${RED}Login failed. Please try again manually with 'gcloud auth login'.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}Successfully logged in!${NC}"

# Get current project
CURRENT_PROJECT=$(gcloud config get-value project)
if [ -z "$CURRENT_PROJECT" ]; then
    echo -e "${YELLOW}No project currently selected.${NC}"
    
    # List projects
    echo "Listing available projects..."
    gcloud projects list
    
    echo ""
    echo "Please enter the Project ID you want to use (copy from the list above):"
    read -r PROJECT_ID
    
    # Set the project
    echo "Setting the active project..."
    gcloud config set project "$PROJECT_ID"
else
    PROJECT_ID=$CURRENT_PROJECT
    echo "Using current project: $PROJECT_ID"
    
    # Confirm project
    echo "Would you like to use this project? (Y/n):"
    read -r CONFIRM
    if [[ "$CONFIRM" =~ ^[Nn].* ]]; then
        # List projects
        echo "Listing available projects..."
        gcloud projects list
        
        echo ""
        echo "Please enter the Project ID you want to use (copy from the list above):"
        read -r PROJECT_ID
        
        # Set the project
        echo "Setting the active project..."
        gcloud config set project "$PROJECT_ID"
    fi
fi

# Enable Compute Engine API if not already enabled
echo "Checking if Compute Engine API is enabled..."
if ! gcloud services list --enabled --filter="name:compute.googleapis.com" | grep -q compute.googleapis.com; then
    echo -e "${YELLOW}Compute Engine API is not enabled. Enabling now...${NC}"
    gcloud services enable compute.googleapis.com
    echo -e "${GREEN}Compute Engine API enabled!${NC}"
else
    echo -e "${GREEN}Compute Engine API is already enabled.${NC}"
fi

# Create a service account
SA_NAME="snoozebot-service"
SA_DISPLAY_NAME="Snoozebot Service Account"
SA_EMAIL="$SA_NAME@$PROJECT_ID.iam.gserviceaccount.com"

echo ""
echo -e "${YELLOW}Now creating a service account for Snoozebot...${NC}"

# Check if service account already exists
if gcloud iam service-accounts list --filter="email:$SA_EMAIL" | grep -q "$SA_EMAIL"; then
    echo -e "${YELLOW}Service account already exists: $SA_EMAIL${NC}"
else
    gcloud iam service-accounts create "$SA_NAME" --display-name="$SA_DISPLAY_NAME"
    echo -e "${GREEN}Service account created: $SA_EMAIL${NC}"
fi

# Assign roles to the service account
echo ""
echo "Assigning Compute Admin role to the service account..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/compute.admin"

echo -e "${GREEN}Role assigned successfully!${NC}"

# Create and download service account key
KEY_DIR="$HOME/.config/gcloud/snoozebot"
KEY_FILE="$KEY_DIR/snoozebot-gcp-key.json"

# Ensure directory exists
mkdir -p "$KEY_DIR"
mkdir -p "$KEY_DIR/application_default_credentials"

echo ""
echo "Creating service account key..."
gcloud iam service-accounts keys create "$KEY_FILE" \
    --iam-account="$SA_EMAIL"

# Set correct permissions
chmod 600 "$KEY_FILE"

# Copy as application default credentials
ADC_FILE="$KEY_DIR/application_default_credentials/adc.json"
cp "$KEY_FILE" "$ADC_FILE"
chmod 600 "$ADC_FILE"

echo ""
echo -e "${GREEN}Service account key created at: $KEY_FILE${NC}"
echo "Please keep this file secure, as it contains sensitive information."

# Create a Snoozebot configuration for GCP
CONFIG_DIR="$HOME/.config/snoozebot"
mkdir -p "$CONFIG_DIR"

CONFIG_FILE="$CONFIG_DIR/gcp-config.json"
cat > "$CONFIG_FILE" << EOF
{
  "provider": "gcp",
  "credentials": {
    "profileName": "snoozebot",
    "credentialsFile": "$KEY_FILE",
    "projectId": "$PROJECT_ID"
  }
}
EOF

# Setting environment variables for the session
export GCP_PROFILE=snoozebot
export GOOGLE_APPLICATION_CREDENTIALS="$KEY_FILE"

echo ""
echo -e "${GREEN}=== Setup Complete! ===${NC}"
echo ""
echo "Your GCP credentials have been set up successfully."
echo "Configuration file created at: $CONFIG_FILE"
echo ""
echo "To use these credentials for tests, add the following to your shell profile:"
echo -e "${YELLOW}export GCP_PROFILE=snoozebot${NC}"
echo -e "${YELLOW}export GOOGLE_APPLICATION_CREDENTIALS=$KEY_FILE${NC}"
echo ""
echo "To use these credentials with Snoozebot, configure it to use the profile 'snoozebot'"
echo "or the credentials file at: $KEY_FILE"
echo ""
echo -e "${RED}⚠️  IMPORTANT SECURITY NOTE ⚠️${NC}"
echo "The credentials file contains sensitive information."
echo "Make sure to keep it secure and not share it with anyone."
echo "You should also consider setting up more granular permissions in a production environment."