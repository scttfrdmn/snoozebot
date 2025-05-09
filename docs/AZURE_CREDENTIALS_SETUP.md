# Azure Credentials Setup for Snoozebot

This guide explains how to set up Azure credentials for use with Snoozebot's Azure cloud provider plugin.

## Prerequisites

- An active Azure subscription
- Azure CLI installed on your system
- Sufficient permissions to create service principals (typically Owner or Contributor role)

## Step 1: Install Azure CLI

If you don't have Azure CLI installed, follow these instructions:

### macOS
```bash
brew update && brew install azure-cli
```

### Linux
```bash
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
```

### Windows
Download and install from: https://aka.ms/installazurecliwindows

## Step 2: Log in to Azure

```bash
az login
```

This will open a browser window where you can authenticate with your Azure account. If you're using a headless environment, you can use:

```bash
az login --use-device-code
```

## Step 3: Check Available Subscriptions

```bash
az account list --output table
```

If you have multiple subscriptions, you can set the default one:

```bash
az account set --subscription "Your-Subscription-Name-or-ID"
```

## Step 4: Create a Service Principal

A service principal is an identity created for use with applications, services, and automation tools to access Azure resources.

```bash
az ad sp create-for-rbac --name "SnoozeBot-ServicePrincipal" --role "Contributor" --scopes "/subscriptions/YOUR-SUBSCRIPTION-ID"
```

Replace `YOUR-SUBSCRIPTION-ID` with your actual subscription ID.

This command will output something like:

```json
{
  "appId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "displayName": "SnoozeBot-ServicePrincipal",
  "password": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "tenant": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

Save this information securely; you won't be able to retrieve the password later.

## Step 5: Create Azure Credentials File

Create a credentials file at `~/.azure/snoozebot-credentials.json` with the following format:

```json
{
  "clientId": "<appId from the output above>",
  "clientSecret": "<password from the output above>",
  "tenantId": "<tenant from the output above>",
  "subscriptionId": "<your subscription ID>"
}
```

For security, ensure the file has restricted permissions:

```bash
chmod 600 ~/.azure/snoozebot-credentials.json
```

## Step 6: Configure Snoozebot to Use These Credentials

You can set up a specific profile for Snoozebot to use these credentials. Create or modify the Snoozebot configuration file to include:

```json
{
  "provider": "azure",
  "credentials": {
    "profileName": "snoozebot",
    "credentialsFile": "~/.azure/snoozebot-credentials.json"
  }
}
```

## Step 7: Set Environment Variable for Tests

For running tests with Azure, set the following environment variable:

```bash
export AZURE_PROFILE=snoozebot
```

## Verifying Your Setup

To verify your credentials are set up correctly, you can run:

```bash
az account show
```

This should display your current active subscription.

## Additional Security Considerations

1. **Limit Permissions**: For production environments, create a service principal with minimal required permissions rather than the broad "Contributor" role.

2. **Credential Rotation**: Regularly rotate the service principal secret (password).

3. **Key Vault**: Consider storing the credentials in Azure Key Vault and configuring Snoozebot to retrieve them securely.

## Troubleshooting

### Error: Insufficient privileges
If you see permission errors, ensure your user account has sufficient privileges to create service principals or request that a global administrator grant you the required permissions.

### Error: Invalid credentials
Double-check that the credential values in your JSON file match exactly what was provided when creating the service principal.

### Expired credentials
Service principal credentials can expire. If you encounter authentication failures, you might need to create a new secret:

```bash
az ad sp credential reset --name "SnoozeBot-ServicePrincipal" --append
```

## Next Steps

Once your Azure credentials are set up, you can proceed with configuring the Snoozebot Azure cloud provider plugin in your application.

For more details on Azure service principals, see the [Azure documentation](https://docs.microsoft.com/en-us/azure/active-directory/develop/app-objects-and-service-principals).