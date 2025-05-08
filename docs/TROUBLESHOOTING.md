# Snoozebot Troubleshooting Guide

This guide helps you diagnose and resolve common issues with Snoozebot, particularly focusing on security features.

## Table of Contents

1. [Security Feature Troubleshooting](#security-feature-troubleshooting)
   - [TLS Issues](#tls-issues)
   - [Signature Verification Issues](#signature-verification-issues)
   - [Authentication Issues](#authentication-issues)
2. [Plugin System Troubleshooting](#plugin-system-troubleshooting)
3. [Common Error Codes](#common-error-codes)
4. [Logging and Debugging](#logging-and-debugging)
5. [Getting Help](#getting-help)

## Security Feature Troubleshooting

### TLS Issues

#### Certificate Not Found

**Symptoms:**
- Error message: "TLS certificate not found"
- Plugin fails to load with TLS errors

**Possible Causes:**
1. Certificates have not been generated
2. Certificate directory path is incorrect
3. Certificate files have been deleted or moved

**Solutions:**
1. Run the security setup wizard to generate certificates:
   ```
   ./bin/securitysetup
   ```
2. Verify the certificate directory exists and has the correct permissions:
   ```
   ls -la /etc/snoozebot/certs
   ```
3. Check that the `SNOOZEBOT_TLS_CERT_DIR` environment variable points to the correct directory:
   ```
   echo $SNOOZEBOT_TLS_CERT_DIR
   ```

#### TLS Handshake Failed

**Symptoms:**
- Error message: "TLS handshake failed"
- Connection between plugin and main application fails
- Plugin fails to start or communicate

**Possible Causes:**
1. TLS version mismatch
2. Certificate not trusted
3. Certificate hostname verification failure
4. Incompatible cipher suites

**Solutions:**
1. Verify certificates are valid and not expired:
   ```
   openssl x509 -in /etc/snoozebot/certs/ca/cert.pem -text -noout
   ```
2. Ensure both client and server certificates are signed by the same CA:
   ```
   openssl verify -CAfile /etc/snoozebot/certs/ca/cert.pem /etc/snoozebot/certs/plugin/cert.pem
   ```
3. For testing purposes only, you can disable certificate verification:
   ```
   export SNOOZEBOT_TLS_SKIP_VERIFY=true
   ```
   **Warning:** Never use this in production environments.

#### Certificate Expired

**Symptoms:**
- Error message: "TLS certificate expired"
- Plugin fails to load with expiration errors

**Solutions:**
1. Check certificate expiration date:
   ```
   openssl x509 -in /etc/snoozebot/certs/plugin/cert.pem -text -noout | grep "Not After"
   ```
2. Regenerate certificates using the security setup wizard:
   ```
   ./bin/securitysetup --force
   ```

### Signature Verification Issues

#### Signature Not Found

**Symptoms:**
- Error message: "Plugin signature not found"
- Plugin fails to load due to missing signature

**Possible Causes:**
1. Plugin has never been signed
2. Signature file was deleted or moved
3. Incorrect signature directory

**Solutions:**
1. Sign the plugin using the snoozesign utility:
   ```
   ./bin/snoozesign -sign -plugin=./bin/plugins/aws -key-id=<key-id>
   ```
2. Verify the signature directory exists and has correct permissions:
   ```
   ls -la /etc/snoozebot/signatures
   ```
3. Check that `SNOOZEBOT_SIGNATURE_DIR` points to the correct directory:
   ```
   echo $SNOOZEBOT_SIGNATURE_DIR
   ```

#### Signature Verification Failed

**Symptoms:**
- Error message: "Signature verification failed"
- Error message: "Plugin binary has been modified"

**Possible Causes:**
1. Plugin binary has been modified after signing
2. Signature is corrupted or tampered with
3. Signing key is not trusted

**Solutions:**
1. Re-sign the plugin with a trusted key:
   ```
   ./bin/snoozesign -sign -plugin=./bin/plugins/aws -key-id=<key-id>
   ```
2. Verify the key is in the trusted keys list:
   ```
   cat /etc/snoozebot/signatures/signature_config.json
   ```
3. Rebuild the plugin to ensure binary integrity:
   ```
   go build -o ./bin/plugins/aws ./plugins/aws
   ```
4. Use the verify command to check the signature:
   ```
   ./bin/snoozesign -verify -plugin=./bin/plugins/aws
   ```

#### Key Expired or Revoked

**Symptoms:**
- Error message: "Expired signing key" or "Revoked signing key"
- Plugin fails to load due to key issues

**Solutions:**
1. Generate a new signing key:
   ```
   ./bin/snoozesign -generate-key -key-name="new-key"
   ```
2. Add the new key to the trusted keys list:
   ```
   ./bin/snoozesign -add-trusted-key -key-id=<new-key-id>
   ```
3. Re-sign the plugin with the new key:
   ```
   ./bin/snoozesign -sign -plugin=./bin/plugins/aws -key-id=<new-key-id>
   ```

### Authentication Issues

#### Invalid API Key

**Symptoms:**
- Error message: "Invalid API key"
- Authentication fails when loading plugins

**Possible Causes:**
1. API key is malformed
2. API key is not in the authentication configuration
3. Authentication configuration is corrupted

**Solutions:**
1. Verify the API key exists in the authentication configuration:
   ```
   cat /etc/snoozebot/config/auth.json
   ```
2. Generate a new API key using the security setup wizard:
   ```
   ./bin/securitysetup
   ```
3. Set the correct API key in the environment:
   ```
   export SNOOZEBOT_API_KEY=<api-key>
   ```

#### Permission Denied

**Symptoms:**
- Error message: "Permission denied"
- Plugin operations fail due to insufficient permissions

**Possible Causes:**
1. API key does not have the required role
2. Required permission is not assigned to the role
3. Authentication configuration is incorrect

**Solutions:**
1. Check the API key roles in the authentication configuration:
   ```
   cat /etc/snoozebot/config/auth.json
   ```
2. Generate a new API key with appropriate roles:
   ```
   ./bin/securitysetup
   ```
3. Modify the authentication configuration to update roles and permissions.

## Plugin System Troubleshooting

### Plugin Not Found

**Symptoms:**
- Error message: "Plugin not found"
- Application fails to load a plugin

**Possible Causes:**
1. Plugin binary does not exist
2. Plugin path is incorrect
3. Permissions issue

**Solutions:**
1. Verify the plugin exists and has the correct permissions:
   ```
   ls -la ./bin/plugins/
   ```
2. Rebuild the plugin if necessary:
   ```
   go build -o ./bin/plugins/aws ./plugins/aws
   ```
3. Check the plugin path configuration.

### Plugin Communication Failed

**Symptoms:**
- Error message: "Failed to communicate with plugin"
- Plugin loads but operations fail

**Possible Causes:**
1. Plugin process has crashed
2. Network or IPC issues
3. TLS configuration mismatch

**Solutions:**
1. Check if the plugin process is running:
   ```
   ps -ef | grep plugin
   ```
2. Enable debug logging to see detailed communication logs:
   ```
   export SNOOZEBOT_DEBUG=true
   ```
3. Ensure TLS configuration is the same on both sides if using TLS.

## Common Error Codes

| Error Code | Description | Resolution |
|------------|-------------|------------|
| `TLS_CERT_NOT_FOUND` | TLS certificate not found | Generate certificates with security setup wizard |
| `TLS_CERT_EXPIRED` | TLS certificate expired | Regenerate certificates |
| `TLS_HANDSHAKE_FAILED` | TLS handshake failed | Check certificate validity and trust |
| `SIG_NOT_FOUND` | Signature not found | Sign the plugin with snoozesign |
| `SIG_VERIFICATION_FAILED` | Signature verification failed | Re-sign the plugin with a trusted key |
| `SIG_HASH_MISMATCH` | Plugin binary modified | Rebuild and re-sign the plugin |
| `AUTH_API_KEY_INVALID` | Invalid API key | Generate a new API key |
| `AUTH_PERMISSION_DENIED` | Permission denied | Use an API key with required roles |
| `PLUGIN_NOT_FOUND` | Plugin not found | Check plugin path and rebuild if necessary |
| `PLUGIN_COMM_FAILED` | Plugin communication failed | Check plugin process and logs |

## Logging and Debugging

Snoozebot uses structured logging to help with troubleshooting. To enable debug logging:

```bash
export SNOOZEBOT_LOG_LEVEL=debug
```

For security-specific logging:

```bash
export SNOOZEBOT_SECURITY_LOG_LEVEL=debug
```

Log files are typically located at:

```
/var/log/snoozebot/snoozebot.log
```

To see real-time logs:

```bash
tail -f /var/log/snoozebot/snoozebot.log | jq
```

## Getting Help

If you continue to experience issues:

1. Check the [Snoozebot documentation](./README.md)
2. Look for similar issues in the issue tracker
3. Collect relevant logs and diagnostics:
   ```bash
   ./bin/snooze diagnose > snoozebot-diag.log
   ```
4. Open a new issue with detailed information about the problem

## Advanced Diagnostics

### TLS Certificate Verification

```bash
# Check CA certificate
openssl x509 -in /etc/snoozebot/certs/ca/cert.pem -text -noout

# Check plugin certificate
openssl x509 -in /etc/snoozebot/certs/plugin/cert.pem -text -noout

# Verify certificate chain
openssl verify -CAfile /etc/snoozebot/certs/ca/cert.pem /etc/snoozebot/certs/plugin/cert.pem
```

### Signature Verification

```bash
# Verify plugin signature
./bin/snoozesign -verify -plugin=./bin/plugins/aws -verbose

# Check trusted keys
./bin/snoozesign -list-trusted-keys

# Compute plugin hash manually
sha256sum ./bin/plugins/aws
```

### Plugin Process Monitoring

```bash
# Find plugin processes
ps -ef | grep -i plugin

# Monitor plugin communication
sudo strace -p <plugin-pid> -e trace=network

# Check plugin memory usage
ps -o pid,user,%mem,%cpu,command -p <plugin-pid>
```