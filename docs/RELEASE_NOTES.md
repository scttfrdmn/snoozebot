# Release Notes - v0.2.0

## Overview

This release introduces significant security enhancements to the Snoozebot plugin system and adds support for Microsoft Azure cloud provider.

## Major Features

### Azure Cloud Provider Support

- Added full support for Azure virtual machines
- Implemented instance management operations (start, stop, get info, list)
- Support for Azure SDK authentication methods
- Comprehensive error handling for Azure-specific scenarios
- Documentation for Azure plugin usage and configuration

### Enhanced Plugin Security

#### TLS Encryption

- Implemented mutual TLS authentication for plugins
- Added automatic certificate generation and management
- Support for custom certificates
- Secure communication between main application and plugins
- Configurable security options

#### Plugin Signature Verification

- Added plugin signing and verification system
- Created the `snoozesign` utility for key management and signing
- Support for key rotation and expiration
- Signature metadata and validation
- Protection against tampered plugins

#### API Key Authentication

- Implemented role-based authentication for plugins
- Support for API key generation and verification
- Permission-based access control
- Secure key storage and validation
- Integration with existing plugin system

## Documentation

- Added comprehensive documentation for security features
- Updated plugin development guides
- Added configuration examples for all new features
- Created troubleshooting information
- Added Azure plugin usage documentation

## Testing Improvements

- Added test scripts for Azure plugin functionality
- Created automated tests for TLS and signature features
- Added integration tests for all security features
- Created test environment setup scripts
- Added performance tests for security overhead

## Configuration

New configuration options are available through environment variables:

### TLS Configuration

```
SNOOZEBOT_TLS_ENABLED=true
SNOOZEBOT_TLS_CERT_DIR=/etc/snoozebot/certs
SNOOZEBOT_TLS_CERT_FILE=/path/to/cert.pem  # Optional
SNOOZEBOT_TLS_KEY_FILE=/path/to/key.pem    # Optional
SNOOZEBOT_TLS_CA_FILE=/path/to/ca.pem      # Optional
SNOOZEBOT_TLS_SKIP_VERIFY=false            # Not recommended for production
```

### Signature Configuration

```
SNOOZEBOT_SIGNATURE_ENABLED=true
SNOOZEBOT_SIGNATURE_DIR=/etc/snoozebot/signatures
```

### Authentication Configuration

```
SNOOZEBOT_AUTH_ENABLED=true
SNOOZEBOT_AUTH_CONFIG=/etc/snoozebot/auth.json
SNOOZEBOT_API_KEY=your-api-key
```

## Breaking Changes

None. All security features are backward compatible and disabled by default.

## Bug Fixes

- Fixed plugin connection handling under high load
- Improved error reporting for plugin failures
- Fixed resource leaks in long-running plugin sessions

## Known Issues

- Key rotation requires a service restart
- Custom certificates require manual distribution to plugin hosts
- Azure plugin requires external Azure CLI authentication in some cases

## Upgrade Instructions

1. Update to the latest version:
   ```bash
   git pull
   go build -o bin/snooze-agent ./agent/cmd
   ```

2. Build the new Azure plugin:
   ```bash
   go build -o bin/plugins/azure ./plugins/azure
   ```

3. Generate API keys if using authentication:
   ```go
   apiKey, err := auth.GenerateAPIKey("admin", []string{"admin"})
   ```

4. Enable security features as needed through environment variables or code.

## Future Plans

- Certificate rotation without service restart
- Enhanced key management
- Additional cloud provider support
- Integration with external certificate authorities
- WebAssembly plugin support