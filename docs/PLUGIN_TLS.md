# Plugin TLS Communication

This document describes how to configure and use the TLS communication features for Snoozebot plugins.

## Overview

Snoozebot uses TLS (Transport Layer Security) to secure the communication between the main application and its plugins. This ensures that all data exchanged between the components is encrypted and authenticated, protecting against eavesdropping and tampering.

The TLS implementation includes:

1. Certificate generation and management
2. Mutual TLS authentication
3. Certificate verification
4. Configurable security options

## Configuration

### Environment Variables

The TLS functionality can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SNOOZEBOT_TLS_ENABLED` | Enable TLS for plugin communication | `false` |
| `SNOOZEBOT_TLS_CERT_DIR` | Directory for storing certificates | `/etc/snoozebot/certs` |
| `SNOOZEBOT_TLS_CERT_FILE` | Path to custom certificate file | |
| `SNOOZEBOT_TLS_KEY_FILE` | Path to custom key file | |
| `SNOOZEBOT_TLS_CA_FILE` | Path to custom CA certificate file | |
| `SNOOZEBOT_TLS_SKIP_VERIFY` | Skip certificate verification (not recommended) | `false` |
| `SNOOZEBOT_TLS_CLEANUP_CERTS` | Clean up certificates when unloading plugins | `false` |

### Configuration Methods

TLS can be configured in three ways:

1. **Automatic Certificate Generation**: If `SNOOZEBOT_TLS_CERT_DIR` is set, certificates will be automatically generated.

2. **Custom Certificates**: If `SNOOZEBOT_TLS_CERT_FILE` and `SNOOZEBOT_TLS_KEY_FILE` are set, these certificates will be used.

3. **Programmatic Configuration**: TLS can be enabled and configured programmatically using the `PluginManagerWithAuth` API.

## Using TLS with Plugins

### Enabling TLS in the Plugin Manager

To enable TLS for plugin communication:

```go
// Create plugin manager with authentication
configDir := "/etc/snoozebot/config"
pluginManager, err := NewPluginManagerWithAuth(baseManager, configDir, logger)
if err != nil {
    // Handle error
}

// Enable TLS
if err := pluginManager.InitializeTLS(); err != nil {
    // Handle error
}
pluginManager.EnableTLS(true)
```

### Loading and Unloading Plugins with TLS

Once TLS is enabled, plugins will be automatically loaded with TLS security:

```go
// Load plugin with TLS
provider, err := pluginManager.LoadPlugin(ctx, "azure")
if err != nil {
    // Handle error
}

// Use the provider...

// Unload plugin when done
if err := pluginManager.UnloadPlugin("azure"); err != nil {
    // Handle error
}
```

### Implementing TLS Support in Plugins

All Snoozebot plugins (AWS, Azure, GCP) have built-in TLS support. For custom plugins, you can use the `ServePluginWithTLS` function:

```go
// Check for TLS configuration
tlsEnabled := os.Getenv("SNOOZEBOT_TLS_ENABLED") == "true"

if tlsEnabled {
    // Set up TLS options
    tlsOptions := &snoozePlugin.TLSOptions{
        Enabled: true,
        CertDir: os.Getenv("SNOOZEBOT_TLS_CERT_DIR"),
    }
    
    // Serve the plugin with TLS
    snoozePlugin.ServePluginWithTLS(pluginImpl, tlsOptions, logger)
} else {
    // Serve without TLS
    plugin.Serve(/* standard configuration */)
}
```

## Certificate Management

### Certificate Authority

Snoozebot generates a self-signed Certificate Authority (CA) that signs all plugin certificates. The CA certificate is stored in the certificates directory and used to verify plugin certificates.

### Certificate Generation

Certificates are generated with the following properties:

- RSA 2048-bit keys
- Valid for 2 years (plugin certificates) or 10 years (CA certificate)
- TLS 1.2+ compatibility
- Supports both server and client authentication

### Certificate Verification

When TLS is enabled, certificate verification is performed by default:

1. The plugin's certificate is verified against the CA certificate
2. The common name (CN) in the certificate is verified against the plugin name
3. Certificate expiration and validity are checked

### Custom Certificates

For environments that require custom certificates (e.g., using an existing PKI), you can provide your own certificates:

```
export SNOOZEBOT_TLS_ENABLED=true
export SNOOZEBOT_TLS_CERT_FILE=/path/to/cert.pem
export SNOOZEBOT_TLS_KEY_FILE=/path/to/key.pem
export SNOOZEBOT_TLS_CA_FILE=/path/to/ca.pem
```

## Security Considerations

### Production Use

For production use, we recommend:

1. Always enable TLS for plugin communication
2. Never set `SNOOZEBOT_TLS_SKIP_VERIFY` to `true`
3. Protect certificate files with appropriate permissions
4. Consider using custom certificates from a trusted CA
5. Regularly rotate certificates

### Testing Environment

For testing, you can simplify the TLS setup:

1. Use automatic certificate generation
2. Set a consistent certificate directory
3. Consider enabling certificate cleanup

## Troubleshooting

### Common Issues

1. **Certificate verification failures**: Ensure the CA certificate is available to both the main application and plugins
2. **Plugin connection failures**: Check certificate and key permissions
3. **TLS handshake errors**: Verify that both sides are using compatible TLS configurations

### Debugging TLS Issues

For debugging TLS issues, you can enable verbose logging:

```go
logger := hclog.New(&hclog.LoggerOptions{
    Name:   "tls-debug",
    Level:  hclog.Debug,
    Output: os.Stderr,
})
```

And in some cases, you may need to temporarily disable verification:

```
export SNOOZEBOT_TLS_SKIP_VERIFY=true
```

**Important**: Never disable verification in production environments.

## Future Enhancements

Planned enhancements to the TLS implementation include:

1. Certificate rotation
2. Certificate revocation checking
3. Enhanced logging for TLS events
4. Integration with external certificate management systems