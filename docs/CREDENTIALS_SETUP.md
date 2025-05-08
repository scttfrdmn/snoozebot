# Setting Up Cloud Provider Credentials for Testing

This document explains how to set up cloud provider credentials for testing Snoozebot plugins.

## Overview

Snoozebot supports profile-based authentication for all major cloud providers:

- **AWS**: Profiles in `~/.aws/config` and `~/.aws/credentials`
- **Azure**: Profile files in `~/.azure/profiles/`
- **GCP**: Service account files in `~/.gcp/` or `~/.config/gcloud/profiles/`

Using profiles allows for:
- Testing with limited-permission accounts
- Switching between different environments (dev, test, prod)
- Secure credential management without hardcoding keys

## AWS Credentials Setup

### Creating AWS Profiles

1. **Create a credentials file** if it doesn't exist:
   ```bash
   mkdir -p ~/.aws
   touch ~/.aws/credentials
   ```

2. **Add a profile** to `~/.aws/credentials`:
   ```ini
   [default]
   aws_access_key_id = YOUR_DEFAULT_ACCESS_KEY
   aws_secret_access_key = YOUR_DEFAULT_SECRET_KEY

   [snoozebot-test]
   aws_access_key_id = YOUR_TEST_ACCESS_KEY
   aws_secret_access_key = YOUR_TEST_SECRET_KEY
   ```

3. **Create a config file** for region information:
   ```bash
   touch ~/.aws/config
   ```

4. **Add region configuration**:
   ```ini
   [default]
   region = us-west-2

   [profile snoozebot-test]
   region = us-east-1
   ```

### Required AWS Permissions

The test user/role should have the following minimum permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:StartInstances",
        "ec2:StopInstances"
      ],
      "Resource": "*"
    }
  ]
}
```

### Using AWS Profiles with Snoozebot

```bash
# Set the profile environment variable
export AWS_PROFILE=snoozebot-test

# Optionally override the region
export AWS_REGION=us-west-2
```

## Azure Credentials Setup

### Creating Azure Profiles

1. **Create Azure CLI profiles directory**:
   ```bash
   mkdir -p ~/.azure/profiles
   ```

2. **Create a profile using Azure CLI** (recommended method):
   ```bash
   # Login to Azure with the account you want to use for testing
   az login

   # Create a service principal for testing
   az ad sp create-for-rbac --name "snoozebot-test" --role contributor \
     --scopes /subscriptions/YOUR_SUBSCRIPTION_ID/resourceGroups/YOUR_RESOURCE_GROUP \
     --output json > ~/.azure/profiles/snoozebot-test.json
   ```

3. **Alternative: Create profile file manually**:

   Create a file at `~/.azure/profiles/snoozebot-test.json`:
   ```json
   {
     "clientId": "YOUR_CLIENT_ID",
     "clientSecret": "YOUR_CLIENT_SECRET",
     "subscriptionId": "YOUR_SUBSCRIPTION_ID",
     "tenantId": "YOUR_TENANT_ID",
     "activeDirectoryEndpointUrl": "https://login.microsoftonline.com",
     "resourceManagerEndpointUrl": "https://management.azure.com/"
   }
   ```

### Required Azure Permissions

The service principal should have the following minimum permissions:
- `Microsoft.Compute/virtualMachines/read`
- `Microsoft.Compute/virtualMachines/start/action`
- `Microsoft.Compute/virtualMachines/deallocate/action`

### Using Azure Profiles with Snoozebot

```bash
# Set the profile environment variable
export AZURE_PROFILE=snoozebot-test

# Set required environment variables
export AZURE_SUBSCRIPTION_ID=your-subscription-id
export AZURE_RESOURCE_GROUP=your-resource-group

# Optionally set VM name and location
export AZURE_VM_NAME=your-vm-name
export AZURE_LOCATION=eastus
```

## GCP Credentials Setup

### Creating GCP Profiles

1. **Create GCP profiles directory**:
   ```bash
   mkdir -p ~/.gcp
   ```

2. **Create a service account in GCP Console**:
   - Go to GCP Console > IAM & Admin > Service Accounts
   - Create a new service account named "snoozebot-test"
   - Assign "Compute Admin" role
   - Create and download a JSON key file

3. **Save service account key file**:
   ```bash
   mv ~/Downloads/project-123456-snoozebot-test.json ~/.gcp/snoozebot-test.json
   ```

### Required GCP Permissions

The service account should have the following minimum permissions:
- `compute.instances.get`
- `compute.instances.list`
- `compute.instances.start`
- `compute.instances.stop`

### Using GCP Profiles with Snoozebot

```bash
# Set the profile environment variable
export GCP_PROFILE=snoozebot-test

# Set project and zone
export PROJECT_ID=your-project-id
export ZONE=us-central1-a
```

## Testing Credential Setup

You can test your credentials by running the credential test script:

```bash
# Test AWS credentials
./scripts/test_credentials.sh aws snoozebot-test

# Test Azure credentials
./scripts/test_credentials.sh azure snoozebot-test

# Test GCP credentials
./scripts/test_credentials.sh gcp snoozebot-test
```

## Environment Setup Script

For convenience, you can use the following script to set up your testing environment:

```bash
#!/bin/bash
# setup_test_env.sh

# Choose which cloud provider to test
PROVIDER=${1:-aws}

case "$PROVIDER" in
  aws)
    export AWS_PROFILE=snoozebot-test
    export AWS_REGION=us-west-2
    ;;
  azure)
    export AZURE_PROFILE=snoozebot-test
    export AZURE_SUBSCRIPTION_ID=your-subscription-id
    export AZURE_RESOURCE_GROUP=your-resource-group
    export AZURE_LOCATION=eastus
    ;;
  gcp)
    export GCP_PROFILE=snoozebot-test
    export PROJECT_ID=your-project-id
    export ZONE=us-central1-a
    ;;
  *)
    echo "Unknown provider: $PROVIDER"
    echo "Usage: $0 [aws|azure|gcp]"
    exit 1
    ;;
esac

echo "Environment set up for $PROVIDER"
```

## Troubleshooting

### AWS

- **Credential not found errors**: Check your `~/.aws/credentials` file.
- **Permission errors**: Verify IAM permissions for your access key.
- **Region errors**: Make sure the region in your config matches the region where your resources are located.

### Azure

- **Authentication failed errors**: Check the contents of your profile file.
- **Service principal expired**: Azure service principals have expiration dates; check if you need to create a new one.
- **Resource not found**: Verify your subscription ID and resource group name.

### GCP

- **Invalid credentials**: Check the path to your service account JSON file.
- **Permission denied**: Verify the service account has the correct IAM roles.
- **Project not found**: Make sure the project ID is correct.

## Security Best Practices

- Use separate accounts/profiles for testing with minimal permissions
- Never commit credential files to version control
- Rotate credentials regularly
- Use environment variables instead of hardcoded credentials
- Consider using a secrets manager for team environments