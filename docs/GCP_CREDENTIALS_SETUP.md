# GCP Credentials Setup for Snoozebot

This guide explains how to set up Google Cloud Platform (GCP) credentials for use with Snoozebot's GCP cloud provider plugin.

## Prerequisites

- A Google Cloud Platform account with an active project
- Google Cloud SDK (gcloud CLI) installed on your system
- Appropriate permissions to create service accounts and assign roles

## Step 1: Install Google Cloud SDK

If you don't have the Google Cloud SDK installed, follow these instructions:

### macOS
```bash
brew install --cask google-cloud-sdk
```

Or download the installer from: https://cloud.google.com/sdk/docs/install

### Linux
```bash
# Add the Cloud SDK distribution URI as a package source
echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

# Import the Google Cloud public key
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -

# Update and install the SDK
sudo apt-get update && sudo apt-get install google-cloud-sdk
```

### Windows
Download and install from: https://cloud.google.com/sdk/docs/install-sdk#windows

## Step 2: Initialize the SDK and Authenticate

```bash
gcloud init
```

This will guide you through the initialization process, including:
- Logging into your Google account
- Selecting a project

## Step 3: Create a Service Account

Service accounts are used by applications, rather than individuals, to authenticate with GCP services.

```bash
# Create a service account
gcloud iam service-accounts create snoozebot-service \
    --display-name="Snoozebot Service Account"

# Get your project ID
PROJECT_ID=$(gcloud config get-value project)
```

## Step 4: Assign Roles to the Service Account

Grant the service account the necessary permissions to manage Compute Engine resources:

```bash
# Assign the Compute Admin role
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:snoozebot-service@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/compute.admin"

# For minimal permissions, you could use more specific roles instead:
# gcloud projects add-iam-policy-binding $PROJECT_ID \
#     --member="serviceAccount:snoozebot-service@$PROJECT_ID.iam.gserviceaccount.com" \
#     --role="roles/compute.instanceAdmin.v1"
```

## Step 5: Create and Download Service Account Key

```bash
# Create and download the key file
gcloud iam service-accounts keys create ~/snoozebot-gcp-key.json \
    --iam-account=snoozebot-service@$PROJECT_ID.iam.gserviceaccount.com
```

For security, ensure the key file has restricted permissions:

```bash
chmod 600 ~/snoozebot-gcp-key.json
```

## Step 6: Create a GCP Configuration Directory

```bash
mkdir -p ~/.config/gcloud/snoozebot
```

## Step 7: Copy the Key File to the Configuration Directory

```bash
cp ~/snoozebot-gcp-key.json ~/.config/gcloud/snoozebot/
```

## Step 8: Create an Application Default Credentials File

```bash
# Create the application default credentials directory if it doesn't exist
mkdir -p ~/.config/gcloud/snoozebot/application_default_credentials

# Copy the service account key as the application default credentials
cp ~/snoozebot-gcp-key.json ~/.config/gcloud/snoozebot/application_default_credentials/adc.json
```

## Step 9: Configure Environment Variables for Snoozebot

Add the following lines to your shell configuration file (e.g., `.bashrc`, `.zshrc`):

```bash
# For using GCP with Snoozebot
export GCP_PROFILE=snoozebot
export GOOGLE_APPLICATION_CREDENTIALS=~/.config/gcloud/snoozebot/snoozebot-gcp-key.json
```

Source your shell configuration file or restart your terminal:

```bash
source ~/.bashrc  # or ~/.zshrc, etc.
```

## Step 10: Configure Snoozebot to Use These Credentials

You can set up a specific profile for Snoozebot to use these credentials. Create or modify the Snoozebot configuration file to include:

```json
{
  "provider": "gcp",
  "credentials": {
    "profileName": "snoozebot",
    "credentialsFile": "~/.config/gcloud/snoozebot/snoozebot-gcp-key.json",
    "projectId": "your-gcp-project-id"
  }
}
```

## Verifying Your Setup

To verify your credentials are set up correctly, you can run:

```bash
# Activate the service account using the key
gcloud auth activate-service-account --key-file=~/.config/gcloud/snoozebot/snoozebot-gcp-key.json

# List Compute Engine instances to verify permissions
gcloud compute instances list
```

If the above command lists instances without errors, your setup is correct.

## Additional Security Considerations

1. **Least Privilege**: Assign only the minimum necessary permissions to the service account.

2. **Key Rotation**: Regularly rotate your service account keys.

3. **Secure Key Storage**: Consider using a secrets management system to store and manage your service account keys.

4. **IP-Based Restrictions**: Consider restricting the service account to only be usable from specific IP addresses.

## Troubleshooting

### Error: Permission denied
Ensure the service account has the necessary roles assigned to it. If you still encounter permission issues, you may need additional roles such as `roles/iam.serviceAccountUser`.

### Error: Invalid credentials
Double-check that the path to your credentials file is correct and that the file contains valid credentials.

### Error: Project not found
Ensure you're using the correct project ID in all configuration files and environment variables.

## Next Steps

Once your GCP credentials are set up, you can proceed with configuring the Snoozebot GCP cloud provider plugin in your application.

For more details on GCP service accounts and authentication, see the [Google Cloud documentation](https://cloud.google.com/docs/authentication/getting-started).