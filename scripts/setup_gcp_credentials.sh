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
# First check if we have permission to list services
if gcloud services list --enabled 2>&1 | grep -q "PERMISSION_DENIED\|not found"; then
    echo -e "${YELLOW}Warning: Unable to check if Compute Engine API is enabled.${NC}"
    echo "This could be because:"
    echo "  1. The project '$PROJECT_ID' doesn't exist"
    echo "  2. You don't have permission to access the project"
    echo "  3. You don't have permission to list enabled services"
    
    echo ""
    echo "Would you like to create a new project instead? (Y/n):"
    read -r CREATE_PROJECT
    
    if [[ ! "$CREATE_PROJECT" =~ ^[Nn].* ]]; then
        echo "Enter a new project ID (must be globally unique, e.g., 'snoozebot-12345'):"
        read -r NEW_PROJECT_ID
        
        if [ -z "$NEW_PROJECT_ID" ]; then
            echo -e "${RED}Error: Project ID cannot be empty.${NC}"
            exit 1
        fi
        
        echo "Creating new project: $NEW_PROJECT_ID"
        if gcloud projects create "$NEW_PROJECT_ID" --name="Snoozebot Test Project"; then
            echo -e "${GREEN}Project created successfully!${NC}"
            PROJECT_ID="$NEW_PROJECT_ID"
            
            # Set as the active project
            echo "Setting as the active project..."
            gcloud config set project "$PROJECT_ID"
            
            # Enable billing if possible
            echo ""
            echo "Note: To enable APIs, you need to enable billing for the project."
            echo "Would you like to open the billing page for this project? (y/N):"
            read -r OPEN_BILLING
            
            if [[ "$OPEN_BILLING" =~ ^[Yy].* ]]; then
                gcloud alpha billing projects link "$PROJECT_ID" --billing-account="$(gcloud alpha billing accounts list --format='value(name)' | head -1)"
                echo "Please visit the following URL to set up billing:"
                echo "https://console.cloud.google.com/billing/linkedaccount?project=$PROJECT_ID"
                echo ""
                echo "Press Enter once you've set up billing..."
                read -r
            fi
            
            # Try to enable the API
            echo "Enabling Compute Engine API..."
            if gcloud services enable compute.googleapis.com; then
                echo -e "${GREEN}Compute Engine API enabled!${NC}"
            else
                echo -e "${YELLOW}Warning: Unable to enable Compute Engine API.${NC}"
                echo "You'll need to enable it manually before using Snoozebot with GCP."
                echo "Visit: https://console.cloud.google.com/apis/library/compute.googleapis.com?project=$PROJECT_ID"
            fi
        else
            echo -e "${RED}Error: Failed to create project.${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}Continuing without checking or enabling Compute Engine API.${NC}"
        echo "You'll need to ensure the Compute Engine API is enabled before using Snoozebot with GCP."
    fi
else
    # Check if the API is enabled
    if ! gcloud services list --enabled --filter="name:compute.googleapis.com" | grep -q compute.googleapis.com; then
        echo -e "${YELLOW}Compute Engine API is not enabled. Enabling now...${NC}"
        gcloud services enable compute.googleapis.com
        echo -e "${GREEN}Compute Engine API enabled!${NC}"
    else
        echo -e "${GREEN}Compute Engine API is already enabled.${NC}"
    fi
fi

# Create a service account
SA_NAME="snoozebot-service"
SA_DISPLAY_NAME="Snoozebot Service Account"
SA_EMAIL="$SA_NAME@$PROJECT_ID.iam.gserviceaccount.com"

echo ""
echo -e "${YELLOW}Now creating a service account for Snoozebot...${NC}"

# Check if we have permission to manage service accounts
if gcloud iam service-accounts list 2>&1 | grep -q "PERMISSION_DENIED\|not found"; then
    echo -e "${YELLOW}Warning: Unable to list service accounts.${NC}"
    echo "This could be because you don't have the necessary permissions."
    echo ""
    
    echo "Would you like to continue with a local key file for your own credentials? (Y/n):"
    read -r USE_OWN_CREDS
    
    if [[ ! "$USE_OWN_CREDS" =~ ^[Nn].* ]]; then
        echo "We'll set up a key file using your own user credentials."
        SA_EMAIL="your-own-credentials"
    else
        echo -e "${RED}Cannot proceed without service account or user credentials.${NC}"
        exit 1
    fi
else
    # Check if service account already exists
    if gcloud iam service-accounts list --filter="email:$SA_EMAIL" | grep -q "$SA_EMAIL"; then
        echo -e "${YELLOW}Service account already exists: $SA_EMAIL${NC}"
    else
        echo "Creating service account..."
        if gcloud iam service-accounts create "$SA_NAME" --display-name="$SA_DISPLAY_NAME"; then
            echo -e "${GREEN}Service account created: $SA_EMAIL${NC}"
        else
            echo -e "${RED}Failed to create service account.${NC}"
            echo "Would you like to continue with a local key file for your own credentials? (Y/n):"
            read -r USE_OWN_CREDS
            
            if [[ ! "$USE_OWN_CREDS" =~ ^[Nn].* ]]; then
                echo "We'll set up a key file using your own user credentials."
                SA_EMAIL="your-own-credentials"
            else
                echo -e "${RED}Cannot proceed without service account or user credentials.${NC}"
                exit 1
            fi
        fi
    fi

    # Only try to assign roles if we created a service account
    if [ "$SA_EMAIL" != "your-own-credentials" ]; then
        # Assign roles to the service account
        echo ""
        echo "Assigning Compute Admin role to the service account..."
        if gcloud projects add-iam-policy-binding "$PROJECT_ID" \
            --member="serviceAccount:$SA_EMAIL" \
            --role="roles/compute.admin"; then
            echo -e "${GREEN}Role assigned successfully!${NC}"
        else
            echo -e "${YELLOW}Warning: Unable to assign role to service account.${NC}"
            echo "You may need to do this manually in the Google Cloud Console."
            echo "Visit: https://console.cloud.google.com/iam-admin/iam?project=$PROJECT_ID"
        fi
    fi
fi

# Create and download service account key
KEY_DIR="$HOME/.config/gcloud/snoozebot"
KEY_FILE="$KEY_DIR/snoozebot-gcp-key.json"

# Ensure directory exists
mkdir -p "$KEY_DIR"
mkdir -p "$KEY_DIR/application_default_credentials"

if [ "$SA_EMAIL" = "your-own-credentials" ]; then
    echo ""
    echo "Creating a credentials file with your user credentials..."
    
    # Use application default credentials
    if gcloud auth application-default login --no-launch-browser; then
        # Copy the application default credentials to our location
        DEFAULT_ADC_PATH="$HOME/.config/gcloud/application_default_credentials.json"
        if [ -f "$DEFAULT_ADC_PATH" ]; then
            cp "$DEFAULT_ADC_PATH" "$KEY_FILE"
            chmod 600 "$KEY_FILE"
            
            # Also copy as application default credentials
            ADC_FILE="$KEY_DIR/application_default_credentials/adc.json"
            cp "$KEY_FILE" "$ADC_FILE"
            chmod 600 "$ADC_FILE"
            
            echo ""
            echo -e "${GREEN}User credentials saved at: $KEY_FILE${NC}"
        else
            echo -e "${RED}Error: Application default credentials not found at expected location.${NC}"
            exit 1
        fi
    else
        echo -e "${RED}Error: Failed to create application default credentials.${NC}"
        exit 1
    fi
else
    echo ""
    echo "Creating service account key..."
    if gcloud iam service-accounts keys create "$KEY_FILE" \
        --iam-account="$SA_EMAIL"; then
        
        # Set correct permissions
        chmod 600 "$KEY_FILE"
        
        # Copy as application default credentials
        ADC_FILE="$KEY_DIR/application_default_credentials/adc.json"
        cp "$KEY_FILE" "$ADC_FILE"
        chmod 600 "$ADC_FILE"
        
        echo ""
        echo -e "${GREEN}Service account key created at: $KEY_FILE${NC}"
    else
        echo -e "${RED}Error: Failed to create service account key.${NC}"
        exit 1
    fi
fi

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