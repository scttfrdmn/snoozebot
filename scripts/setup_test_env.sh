#!/bin/bash

# Script to set up testing environment for cloud provider plugins
# Usage: ./setup_test_env.sh <provider> <profile>
# Example: ./setup_test_env.sh aws snoozebot-test

set -e

# Parse command line arguments
PROVIDER=${1:-aws}
PROFILE=${2:-default}

# Load configuration from .env file if it exists
if [ -f .env ]; then
  source .env
fi

# Set up environment variables based on provider
case "$PROVIDER" in
  aws)
    echo "Setting up AWS environment with profile: $PROFILE"
    export AWS_PROFILE=$PROFILE
    
    # Set default region if not already set
    if [ -z "$AWS_REGION" ]; then
      export AWS_REGION=us-west-2
      echo "Using default AWS region: $AWS_REGION"
    else
      echo "Using AWS region: $AWS_REGION"
    fi
    
    echo "AWS_PROFILE=$PROFILE" > .env
    echo "AWS_REGION=$AWS_REGION" >> .env
    ;;
    
  azure)
    echo "Setting up Azure environment with profile: $PROFILE"
    export AZURE_PROFILE=$PROFILE
    
    # Prompt for required variables if not set
    if [ -z "$AZURE_SUBSCRIPTION_ID" ]; then
      read -p "Enter Azure subscription ID: " AZURE_SUBSCRIPTION_ID
      export AZURE_SUBSCRIPTION_ID
    fi
    
    if [ -z "$AZURE_RESOURCE_GROUP" ]; then
      read -p "Enter Azure resource group: " AZURE_RESOURCE_GROUP
      export AZURE_RESOURCE_GROUP
    fi
    
    if [ -z "$AZURE_LOCATION" ]; then
      export AZURE_LOCATION=eastus
      echo "Using default Azure location: $AZURE_LOCATION"
    fi
    
    if [ -z "$AZURE_VM_NAME" ]; then
      read -p "Enter Azure VM name (or leave blank for default): " AZURE_VM_NAME
      if [ -z "$AZURE_VM_NAME" ]; then
        export AZURE_VM_NAME=default-vm
        echo "Using default VM name: $AZURE_VM_NAME"
      else
        export AZURE_VM_NAME
      fi
    fi
    
    echo "AZURE_PROFILE=$PROFILE" > .env
    echo "AZURE_SUBSCRIPTION_ID=$AZURE_SUBSCRIPTION_ID" >> .env
    echo "AZURE_RESOURCE_GROUP=$AZURE_RESOURCE_GROUP" >> .env
    echo "AZURE_LOCATION=$AZURE_LOCATION" >> .env
    echo "AZURE_VM_NAME=$AZURE_VM_NAME" >> .env
    ;;
    
  gcp)
    echo "Setting up GCP environment with profile: $PROFILE"
    export GCP_PROFILE=$PROFILE
    
    # Determine credentials file path
    HOME_DIR=$(eval echo ~$USER)
    CREDS_FILE="$HOME_DIR/.gcp/$PROFILE.json"
    
    if [ ! -f "$CREDS_FILE" ]; then
      # Try alternate location
      CREDS_FILE="$HOME_DIR/.config/gcloud/profiles/$PROFILE.json"
      
      if [ ! -f "$CREDS_FILE" ]; then
        echo "Could not find credentials file for profile $PROFILE."
        read -p "Enter the path to your service account JSON file: " CUSTOM_CREDS_FILE
        CREDS_FILE=$CUSTOM_CREDS_FILE
      fi
    fi
    
    # Verify credentials file exists
    if [ ! -f "$CREDS_FILE" ]; then
      echo "Error: Could not find credentials file: $CREDS_FILE"
      exit 1
    fi
    
    export GOOGLE_APPLICATION_CREDENTIALS=$CREDS_FILE
    echo "Using credentials file: $CREDS_FILE"
    
    # Prompt for required variables if not set
    if [ -z "$PROJECT_ID" ]; then
      read -p "Enter GCP project ID: " PROJECT_ID
      export PROJECT_ID
    fi
    
    if [ -z "$ZONE" ]; then
      export ZONE=us-central1-a
      echo "Using default GCP zone: $ZONE"
    fi
    
    echo "GCP_PROFILE=$PROFILE" > .env
    echo "GOOGLE_APPLICATION_CREDENTIALS=$CREDS_FILE" >> .env
    echo "PROJECT_ID=$PROJECT_ID" >> .env
    echo "ZONE=$ZONE" >> .env
    ;;
    
  *)
    echo "Error: Unsupported provider: $PROVIDER"
    echo "Supported providers: aws, azure, gcp"
    exit 1
    ;;
esac

# Save environment variables to .env file
echo "# Environment variables for testing $PROVIDER with profile $PROFILE" > .env.test
echo "# Generated at $(date)" >> .env.test
env | grep -E "^(AWS_|AZURE_|GCP_|GOOGLE_|PROJECT_|ZONE)" >> .env.test

echo "Test environment setup complete for $PROVIDER with profile $PROFILE"
echo "Environment variables saved to .env file"
echo "Run the following command to test the credentials:"
echo "./scripts/test_credentials.sh $PROVIDER $PROFILE"