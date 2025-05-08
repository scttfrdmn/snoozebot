# Plugin Signature Verification

This document describes the plugin signature verification system in Snoozebot, which provides a way to verify the authenticity and integrity of plugins.

## Overview

Snoozebot's plugin signature system ensures that only trusted plugins can be loaded by the application. The system provides:

1. A mechanism for signing plugins
2. Signature verification during plugin loading
3. Key management for signing and verification
4. Integration with existing security features (authentication and TLS)

## Architecture

The signature verification system consists of the following components:

### Core Components

1. **Signature Models** (`pkg/plugin/signature/models.go`):
   - Defines the data structures for signatures and keys
   - Provides functions for signature creation and verification
   - Implements cryptographic operations for signatures

2. **Signature Service** (`pkg/plugin/signature/service.go`):
   - Manages the signature verification process
   - Maintains a list of trusted keys
   - Provides key generation and management

3. **Plugin Signing** (`pkg/plugin/signature/sign.go`):
   - Utilities for signing plugins
   - Creates signed plugin bundles for distribution

4. **Plugin Manager Integration** (`agent/provider/manager_signature.go`):
   - Extends the plugin manager with signature verification
   - Integrates with the existing authentication and TLS systems

### Command-Line Tools

1. **snoozesign** (`cmd/snoozesign/main.go`):
   - Command-line utility for signing plugins
   - Manages signing keys
   - Verifies plugin signatures
   - Creates signed plugin bundles

2. **sign_plugins.sh** (`scripts/sign_plugins.sh`):
   - Script for signing all plugins in a directory
   - Generates signing keys if needed
   - Creates plugin bundles for distribution
   - Verifies signatures

## Signature Format

Each plugin signature consists of the following information:

- Version: The signature format version
- Plugin Name: The name of the plugin
- Plugin Version: The version of the plugin
- Hash Algorithm: The algorithm used to hash the plugin binary
- Hash: The base64-encoded hash of the plugin binary
- Signature Value: The base64-encoded signature of the hash
- Signature Algorithm: The algorithm used to sign the hash
- Key ID: The identifier for the key used to sign the plugin
- Issuer: The entity that signed the plugin
- Timestamp: The time when the signature was created
- Expires At: The time when the signature expires

Signatures are stored as JSON files alongside the plugins or in a central signature directory.

## Key Management

The signature system uses asymmetric cryptography (RSA) for signing and verification:

1. **Key Generation**:
   - RSA key pairs (2048 bits by default)
   - Keys are stored securely in the key directory
   - Public keys are distributed for verification

2. **Key Trust**:
   - Only trusted keys can be used for verification
   - Keys can be added to or removed from the trusted list
   - Keys can be revoked to prevent their use

3. **Key Expiration**:
   - Keys have an expiration date (2 years by default)
   - Expired keys are automatically rejected

## Using Signature Verification

### Configuration

Signature verification can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SNOOZEBOT_SIGNATURE_ENABLED` | Enable signature verification | `false` |
| `SNOOZEBOT_SIGNATURE_DIR` | Directory for signatures | `/etc/snoozebot/signatures` |

### Enabling Signature Verification

To enable signature verification in your Snoozebot deployment:

```go
// Create a plugin manager with signature verification
configDir := "/etc/snoozebot/config"
pluginManager, err := NewPluginManagerWithSignature(baseManager, configDir, logger)
if err != nil {
    // Handle error
}

// Enable signature verification
pluginManager.EnableSignatureVerification(true)
```

For a complete security solution, you can use the combined manager:

```go
// Create a plugin manager with authentication, TLS, and signature verification
configDir := "/etc/snoozebot/config"
pluginManager, err := NewPluginManagerWithAuthAndSignature(baseManager, configDir, logger)
if err != nil {
    // Handle error
}

// Enable all security features
pluginManager.EnableAuthentication(true)
pluginManager.EnableTLS(true)
pluginManager.EnableSignatureVerification(true)
```

### Signing Plugins

To sign a plugin:

1. Using the command-line utility:

```bash
# Generate a signing key
./bin/snoozesign -generate-key -key-name="release-key"

# Sign a plugin
./bin/snoozesign -sign -plugin="./bin/plugins/aws" -key-id="<key-id>"

# Verify a plugin signature
./bin/snoozesign -verify -plugin="./bin/plugins/aws"
```

2. Using the script:

```bash
# Sign all plugins in the plugins directory
./scripts/sign_plugins.sh
```

3. Programmatically:

```go
// Create a signature service
sigService, err := signature.NewSignatureService("/etc/snoozebot/signatures", logger)
if err != nil {
    // Handle error
}

// Create a plugin signer
pluginSigner := signature.NewPluginSigner(sigService, logger)

// Sign a plugin
err = pluginSigner.SignPlugin("./bin/plugins/aws", keyID)
if err != nil {
    // Handle error
}
```

### Creating Plugin Bundles

Plugin bundles contain a signed plugin and its signature for distribution:

```bash
# Create a plugin bundle
./bin/snoozesign -bundle -plugin="./bin/plugins/aws" -bundle-path="./aws-bundle" -key-id="<key-id>"
```

### Cloud Provider Integration

All cloud provider plugins (AWS, Azure, GCP) have built-in support for signature verification:

```bash
# Run a plugin with signature verification enabled
SNOOZEBOT_SIGNATURE_ENABLED=true ./bin/plugins/aws
```

## Security Considerations

### Best Practices

1. **Key Security**:
   - Store private keys securely
   - Limit access to signing keys
   - Use separate keys for development and production

2. **Signature Management**:
   - Regularly rotate signing keys
   - Verify signatures before loading plugins
   - Keep the trusted key list up to date

3. **Deployment**:
   - Use signature verification in production environments
   - Combine with TLS and authentication for complete security
   - Use signed plugin bundles for distribution

### Integration with TLS and Authentication

For maximum security, use signature verification in combination with TLS and authentication:

1. **TLS** secures the communication channel
2. **Authentication** ensures that only authorized plugins can be used
3. **Signature Verification** ensures that plugins are authentic and unmodified

## Future Enhancements

Planned enhancements to the signature system include:

1. **Revocation Lists**: Certificate Revocation Lists (CRLs) for key revocation
2. **Online Verification**: Online signature verification services
3. **Hardware Security**: Integration with Hardware Security Modules (HSMs)
4. **Signature Delegation**: Allowing trusted signers to delegate signing authority
5. **Policy-Based Verification**: Fine-grained control over signature requirements

## Conclusion

The plugin signature system provides a robust mechanism for verifying the authenticity and integrity of plugins, ensuring that only trusted plugins can be loaded by Snoozebot. When combined with TLS and authentication, it creates a comprehensive security solution for the plugin system.