#!/bin/bash

# Script to test cloud provider credentials
# Usage: ./test_credentials.sh <provider> <profile>
# Example: ./test_credentials.sh aws snoozebot-test

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Parse command line arguments
PROVIDER=${1:-}
PROFILE=${2:-}

if [ -z "$PROVIDER" ]; then
  echo -e "${RED}Error: Provider not specified${NC}"
  echo "Usage: $0 <provider> <profile>"
  echo "Example: $0 aws snoozebot-test"
  exit 1
fi

if [ -z "$PROFILE" ]; then
  echo -e "${YELLOW}Warning: Profile not specified, using default${NC}"
  PROFILE="default"
fi

# Function to check AWS credentials
test_aws_credentials() {
  echo -e "${YELLOW}Testing AWS credentials for profile: $PROFILE${NC}"
  export AWS_PROFILE=$PROFILE
  
  # Run the AWS CLI to test credentials
  echo "Checking AWS identity..."
  aws sts get-caller-identity
  
  echo "Listing EC2 regions..."
  aws ec2 describe-regions --query "Regions[].RegionName" --output text
  
  # If a specific region is set, list instances in that region
  if [ -n "$AWS_REGION" ]; then
    echo "Listing EC2 instances in $AWS_REGION..."
    aws ec2 describe-instances --query "Reservations[].Instances[].[InstanceId,State.Name]" --output table
  fi
  
  echo -e "${GREEN}✓ AWS credentials are working properly${NC}"
}

# Function to check Azure credentials
test_azure_credentials() {
  echo -e "${YELLOW}Testing Azure credentials for profile: $PROFILE${NC}"
  export AZURE_PROFILE=$PROFILE
  
  # Check if we have the required environment variables
  if [ -z "$AZURE_SUBSCRIPTION_ID" ]; then
    echo -e "${RED}Error: AZURE_SUBSCRIPTION_ID environment variable is not set${NC}"
    exit 1
  fi
  
  if [ -z "$AZURE_RESOURCE_GROUP" ]; then
    echo -e "${RED}Error: AZURE_RESOURCE_GROUP environment variable is not set${NC}"
    exit 1
  fi
  
  # Run the Azure CLI to test credentials
  echo "Checking Azure identity..."
  az account show
  
  echo "Listing Azure subscriptions..."
  az account list --query "[].name" --output tsv
  
  echo "Listing resource groups in the current subscription..."
  az group list --query "[].name" --output tsv
  
  echo "Listing VMs in resource group $AZURE_RESOURCE_GROUP..."
  az vm list -g $AZURE_RESOURCE_GROUP --query "[].name" --output tsv
  
  echo -e "${GREEN}✓ Azure credentials are working properly${NC}"
}

# Function to check GCP credentials
test_gcp_credentials() {
  echo -e "${YELLOW}Testing GCP credentials for profile: $PROFILE${NC}"
  export GCP_PROFILE=$PROFILE
  
  # Check if we have the required environment variables
  if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Error: PROJECT_ID environment variable is not set${NC}"
    exit 1
  fi
  
  # Determine credentials file path
  HOME_DIR=$(eval echo ~$USER)
  CREDS_FILE="$HOME_DIR/.gcp/$PROFILE.json"
  
  if [ ! -f "$CREDS_FILE" ]; then
    # Try alternate location
    CREDS_FILE="$HOME_DIR/.config/gcloud/profiles/$PROFILE.json"
    
    if [ ! -f "$CREDS_FILE" ]; then
      echo -e "${RED}Error: Could not find credentials file for profile $PROFILE${NC}"
      echo "Expected locations:"
      echo "  $HOME_DIR/.gcp/$PROFILE.json"
      echo "  $HOME_DIR/.config/gcloud/profiles/$PROFILE.json"
      exit 1
    fi
  fi
  
  export GOOGLE_APPLICATION_CREDENTIALS=$CREDS_FILE
  
  # Run the GCP CLI to test credentials
  echo "Checking GCP identity..."
  gcloud auth activate-service-account --key-file=$CREDS_FILE
  
  echo "Setting project to $PROJECT_ID..."
  gcloud config set project $PROJECT_ID
  
  echo "Listing GCP zones..."
  gcloud compute zones list --limit=5
  
  # If a specific zone is set, list instances in that zone
  if [ -n "$ZONE" ]; then
    echo "Listing GCP instances in zone $ZONE..."
    gcloud compute instances list --zones=$ZONE
  fi
  
  echo -e "${GREEN}✓ GCP credentials are working properly${NC}"
}

# Main execution
case "$PROVIDER" in
  aws)
    test_aws_credentials
    ;;
  azure)
    test_azure_credentials
    ;;
  gcp)
    test_gcp_credentials
    ;;
  *)
    echo -e "${RED}Error: Unsupported provider: $PROVIDER${NC}"
    echo "Supported providers: aws, azure, gcp"
    exit 1
    ;;
esac

echo -e "${GREEN}All credential tests completed successfully!${NC}"