# Snoozebot Installation Guide

This document provides step-by-step instructions for installing and configuring Snoozebot with all its features.

## Prerequisites

- Go 1.18 or newer
- Git
- For Azure plugin: Azure CLI or Service Principal credentials
- For AWS plugin: AWS CLI configured or IAM credentials

## Basic Installation

### 1. Clone the Repository

```bash
git clone https://github.com/scottfridman/snoozebot.git
cd snoozebot
```

### 2. Build the Main Components

```bash
# Build the agent
go build -o bin/snooze-agent ./agent/cmd

# Build the CLI
go build -o bin/snooze ./cmd/snooze

# Build the snoozesign utility (for plugin signing)
go build -o bin/snoozesign ./cmd/snoozesign
```

### 3. Build the Cloud Provider Plugins

```bash
# Build the AWS plugin
go build -o bin/plugins/aws ./plugins/aws

# Build the Azure plugin
go build -o bin/plugins/azure ./plugins/azure
```

### 4. Create Configuration Directories

```bash
# Create configuration directories
mkdir -p /etc/snoozebot/config
mkdir -p /etc/snoozebot/certs
mkdir -p /etc/snoozebot/signatures
chmod 700 /etc/snoozebot/config /etc/snoozebot/certs /etc/snoozebot/signatures
```

## Basic Configuration

### Agent Configuration

Create a basic configuration file at `/etc/snoozebot/config/agent.json`:

```json
{
  "listen_address": "0.0.0.0",
  "listen_port": 8080,
  "log_level": "info",
  "plugin_dir": "/path/to/snoozebot/bin/plugins",
  "idle_timeout": 3600
}
```

### Plugin Configuration

Each cloud provider plugin requires its own configuration. Set up environment variables according to your cloud provider:

#### AWS Plugin

```bash
export AWS_REGION=us-west-2
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
```

#### Azure Plugin

```bash
export AZURE_SUBSCRIPTION_ID=your-subscription-id
export AZURE_RESOURCE_GROUP=your-resource-group
export AZURE_LOCATION=eastus
export AZURE_VM_NAME=your-vm-name
```

## Security Features Configuration

### 1. TLS Encryption

To enable TLS encryption for plugin communication:

```bash
# Enable TLS
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_DIR=/etc/snoozebot/certs

# Optional: For custom certificates
# export SNOOZEBOT_TLS_CERT_FILE=/path/to/cert.pem
# export SNOOZEBOT_TLS_KEY_FILE=/path/to/key.pem
# export SNOOZEBOT_TLS_CA_FILE=/path/to/ca.pem
```

The certificates will be automatically generated the first time you run the application with TLS enabled.

### 2. Signature Verification

To enable plugin signature verification:

```bash
# Enable signature verification
export SNOOZEBOT_SIGNATURE_ENABLED=true
export SNOOZEBOT_SIGNATURE_DIR=/etc/snoozebot/signatures
```

Generate a signing key and sign your plugins:

```bash
# Generate a signing key
./bin/snoozesign -generate-key -key-name="release-key"

# Get the key ID
KEY_ID=$(ls /etc/snoozebot/signatures | grep -v .pub | head -1 | sed 's/\.key$//')

# Sign a plugin
./bin/snoozesign -sign -plugin="./bin/plugins/aws" -key-id="$KEY_ID"
./bin/snoozesign -sign -plugin="./bin/plugins/azure" -key-id="$KEY_ID"

# Verify a plugin signature
./bin/snoozesign -verify -plugin="./bin/plugins/aws"
```

### 3. API Key Authentication

To enable API key authentication:

```bash
# Enable authentication
export SNOOZEBOT_AUTH_ENABLED=true
export SNOOZEBOT_AUTH_CONFIG=/etc/snoozebot/config/auth.json
```

Create the auth configuration file with an API key:

```go
// Save this to a Go file and run it once to generate your API key
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/scttfrdmn/snoozebot/pkg/plugin/auth"
)

func main() {
    // Create a new API key with admin role
    apiKey, err := auth.GenerateAPIKey("admin", []string{"admin"})
    if err != nil {
        fmt.Printf("Failed to generate API key: %v\n", err)
        os.Exit(1)
    }

    // Create the auth config
    authConfig := auth.PluginAuthConfig{
        APIKeys: []auth.APIKey{*apiKey},
    }

    // Save the config
    data, err := json.MarshalIndent(authConfig, "", "  ")
    if err != nil {
        fmt.Printf("Failed to marshal auth config: %v\n", err)
        os.Exit(1)
    }

    err = os.WriteFile("/etc/snoozebot/config/auth.json", data, 0600)
    if err != nil {
        fmt.Printf("Failed to write auth config: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("API key generated successfully: %s\n", apiKey.Key)
}
```

After generating the API key, use it in your requests:

```bash
export SNOOZEBOT_API_KEY=your-api-key
```

## Running Snoozebot

### Start the Agent

```bash
./bin/snooze-agent
```

### Use the CLI to Interact with the Agent

```bash
# Get information about the current instance
./bin/snooze provider aws get-info

# List all instances
./bin/snooze provider aws list-instances

# Stop an instance
./bin/snooze provider aws stop-instance

# Start an instance
./bin/snooze provider aws start-instance
```

## Troubleshooting

### TLS Issues

- Check certificate permissions: Certificates should be readable by the user running Snoozebot
- Verify certificate paths: Ensure paths are correct in environment variables
- Check logs for TLS errors: Run with increased log verbosity
- For testing: Try with `SNOOZEBOT_TLS_SKIP_VERIFY=true` (not for production)

### Signature Verification Issues

- Check signature directory permissions
- Verify that plugins have been signed
- Check key permissions
- Run signature verification manually to see detailed errors

### Authentication Issues

- Verify API key in environment variable
- Check auth configuration file permissions
- Ensure role permissions are correct
- Check logs for authentication failures

## For Production Environments

For production deployments, consider:

1. Using a process manager like systemd
2. Setting up proper log rotation
3. Using secure credential management (not environment variables)
4. Implementing monitoring and alerts
5. Regular certificate and key rotation

## Next Steps

After installation, refer to these documents for more details:

- [Plugin TLS](./PLUGIN_TLS.md) - Detailed TLS configuration guide
- [Plugin Signatures](./PLUGIN_SIGNATURES.md) - Detailed signature verification guide
- [Plugin Authentication](./PLUGIN_AUTHENTICATION.md) - Detailed authentication guide