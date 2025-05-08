# Snoozebot Security Setup Guide

This guide explains how to set up and configure security features for Snoozebot using the security setup wizard.

## Security Features Overview

Snoozebot includes several security features to ensure the safety and integrity of the plugin system:

1. **TLS Encryption**: Secures communication between the main application and plugins
2. **Signature Verification**: Ensures plugins are authentic and have not been tampered with
3. **API Key Authentication**: Controls access to plugins based on roles and permissions

## Using the Security Setup Wizard

The security setup wizard provides a streamlined way to configure all security features at once.

### Prerequisites

- Go 1.18 or newer
- Write access to the configuration directory (defaults to `/etc/snoozebot`)

### Running the Wizard

```bash
# Build the security setup wizard
go build -o bin/securitysetup ./cmd/securitysetup

# Run the wizard (may require sudo if writing to /etc/snoozebot)
./bin/securitysetup

# Alternatively, specify a different configuration directory
SNOOZEBOT_CONFIG_DIR=/path/to/config ./bin/securitysetup

# Use the --force flag to overwrite existing configuration files
./bin/securitysetup --force
```

### Wizard Steps

The wizard will guide you through the following steps:

1. **Directory Setup**: Creates necessary directories for configuration files, certificates, and signatures
2. **TLS Configuration**: Generates TLS certificates for secure communication
3. **Signature Verification**: Creates signing keys and configures signature verification
4. **Authentication**: Generates API keys and configures role-based access control
5. **Environment Configuration**: Creates a file with environment variables for easy setup

After completing the wizard, you'll have a fully configured security environment for Snoozebot.

## Manual Configuration

If you prefer to configure security features manually, refer to the following documentation:

- [TLS Configuration](./PLUGIN_TLS.md)
- [Signature Verification](./PLUGIN_SIGNATURES.md)
- [Authentication](./PLUGIN_AUTHENTICATION.md)

## Environment Variables

The security features can be configured using the following environment variables:

### TLS Configuration

```
SNOOZEBOT_TLS_ENABLED=true
SNOOZEBOT_TLS_CERT_DIR=/etc/snoozebot/certs
SNOOZEBOT_TLS_CERT_FILE=/path/to/cert.pem  # Optional, for custom certificates
SNOOZEBOT_TLS_KEY_FILE=/path/to/key.pem    # Optional, for custom certificates
SNOOZEBOT_TLS_CA_FILE=/path/to/ca.pem      # Optional, for custom CA certificates
SNOOZEBOT_TLS_SKIP_VERIFY=false            # Optional, not recommended for production
```

### Signature Verification

```
SNOOZEBOT_SIGNATURE_ENABLED=true
SNOOZEBOT_SIGNATURE_DIR=/etc/snoozebot/signatures
```

### Authentication

```
SNOOZEBOT_AUTH_ENABLED=true
SNOOZEBOT_AUTH_CONFIG=/etc/snoozebot/config/auth.json
SNOOZEBOT_API_KEY=your-api-key  # Client-side API key
```

## Using the Generated Configuration

After running the wizard, you can source the generated environment file to enable all security features:

```bash
source /etc/snoozebot/security.env
```

You can also add this to your application's startup script or systemd service.

## Troubleshooting

If you encounter issues with the security setup:

1. **TLS Issues**:
   - Check certificate permissions (should be readable by the application)
   - Verify that certificates are valid and not expired
   - Ensure the CA certificate is available to both the application and plugins

2. **Signature Verification Issues**:
   - Ensure plugins are signed with a trusted key
   - Check that the signature directory is accessible
   - Verify that signatures are valid and not expired

3. **Authentication Issues**:
   - Confirm that the API key is valid and not revoked
   - Check that the configured roles have the necessary permissions
   - Ensure the authentication configuration file is accessible

For more detailed troubleshooting information, see the [Troubleshooting Guide](./TROUBLESHOOTING.md).

## Security Best Practices

1. **Production Environments**:
   - Enable all security features in production
   - Use custom certificates from a trusted CA if possible
   - Rotate keys and certificates regularly
   - Limit access to configuration files and private keys

2. **Development Environments**:
   - Use self-signed certificates for development
   - Set up separate keys for development and production
   - Use the security wizard to quickly set up a secure environment

3. **Logging and Monitoring**:
   - Enable security event logging
   - Monitor for certificate expiration
   - Set up alerts for security-related issues