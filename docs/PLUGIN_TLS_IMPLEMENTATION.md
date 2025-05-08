# Plugin TLS Implementation

This document details the implementation of TLS (Transport Layer Security) for the Snoozebot plugin system.

## Overview

TLS support has been added to secure communication between the Snoozebot core application and its plugins. This implementation provides:

1. Mutual TLS authentication
2. Certificate generation and management
3. Certificate verification
4. Configuration options for different deployment environments

## Architecture

The TLS implementation includes several components:

1. **Certificate Management**: 
   - Located in `/pkg/plugin/tls/cert.go`
   - Handles certificate generation, loading, and verification
   - Creates a self-signed Certificate Authority (CA)
   - Issues certificates for each plugin signed by the CA

2. **TLS Manager**:
   - Located in `/pkg/plugin/tls/manager.go` 
   - Manages TLS certificates and configuration for plugins
   - Provides a unified interface for certificate operations
   - Handles certificate rotation and cleanup

3. **Secure Plugin**:
   - Located in `/pkg/plugin/tls_plugin.go`
   - Wraps plugin communication with TLS security
   - Provides verification of certificates
   - Supports both automatic and manual certificate configuration

4. **Plugin Manager Integration**:
   - Located in `/agent/provider/manager_auth.go`
   - Integrates TLS with the plugin authentication system
   - Handles secure loading and unloading of plugins
   - Manages TLS environment variables for plugins

## Securing Plugin Communication

### Certificate Flow

1. A Certificate Authority (CA) is created in the specified certificate directory
2. Each plugin is issued a certificate signed by the CA
3. The CA certificate is used to verify plugin certificates
4. Communication is encrypted using TLS 1.2+

### Plugin Loading Process

1. The plugin manager initializes the TLS manager
2. Certificates are ensured for both the manager and the plugin
3. A secure connection is established with certificate verification
4. Plugin operations are performed over the encrypted connection

### Certificate Verification

Certificate verification happens in both directions:

1. **Server Verification**:
   - The plugin verifies the server's certificate against the CA
   - The common name is checked to ensure it matches the expected name
   - Certificate validity period is verified

2. **Client Verification**:
   - The server verifies the client's certificate against the CA
   - Certificate chains are validated to ensure trust
   - Certificate revocation status is checked (when implemented)

## Configuration Options

The TLS system supports flexible configuration:

1. **Environment Variables**:
   - `SNOOZEBOT_TLS_ENABLED`: Enable/disable TLS
   - `SNOOZEBOT_TLS_CERT_DIR`: Directory for certificates
   - `SNOOZEBOT_TLS_CERT_FILE`: Custom certificate path
   - `SNOOZEBOT_TLS_KEY_FILE`: Custom key path
   - `SNOOZEBOT_TLS_CA_FILE`: Custom CA certificate path
   - `SNOOZEBOT_TLS_SKIP_VERIFY`: Skip verification (not recommended)
   - `SNOOZEBOT_TLS_CLEANUP_CERTS`: Clean up certificates on unload

2. **Programmatic Configuration**:
   - `EnableTLS()`: Enable/disable TLS in the plugin manager
   - `InitializeTLS()`: Initialize TLS configuration
   - `TLSOptions`: Configure TLS parameters for plugins

## Implementation in Cloud Providers

TLS support has been added to all cloud provider plugins:

1. **AWS Plugin**:
   - Detects TLS configuration from environment variables
   - Uses `ServePluginWithTLS` for secure operation
   - Supports both automatic and manual certificate configuration

2. **Azure Plugin**:
   - Implements TLS support with certificate verification
   - Handles environment variable configuration
   - Provides informative logging for TLS operations

3. **GCP Plugin**:
   - Supports secure communication with TLS
   - Handles certificate management
   - Maintains compatibility with the plugin interface

## Testing

A test script (`scripts/test_tls_plugin.sh`) has been created to verify TLS functionality:

1. Builds plugins with TLS support
2. Sets up a test environment with certificates
3. Starts a plugin with TLS enabled
4. Verifies successful plugin operation
5. Cleans up test resources

## Future Enhancements

The TLS implementation lays the groundwork for future security enhancements:

1. **Certificate Rotation**: Automated rotation of certificates
2. **Certificate Revocation**: Support for checking certificate revocation status
3. **External PKI Integration**: Integration with external certificate authorities
4. **Hardware Security Module Support**: Integration with HSMs for key storage

## Conclusion

The TLS implementation provides a robust security layer for plugin communication, ensuring data privacy and integrity while maintaining flexibility for different deployment scenarios. This implementation completes a key security milestone for the Snoozebot project.