# Snoozebot Security Monitoring

This document explains how to use the security monitoring features in Snoozebot to track, analyze, and respond to security events.

## Overview

Snoozebot provides comprehensive security event logging and monitoring for:

1. **TLS Encryption**: Certificates, handshakes, and verification events
2. **Signature Verification**: Plugin signatures, key management, and verification events
3. **Authentication**: API key usage, authorization, and access control events
4. **Plugin System**: Plugin loading, unloading, and communication events

## Security Event Logging

Security events are automatically logged to the following locations:

- **File**: `/var/log/snoozebot/security/security-latest.log`
- **Console**: Events are also logged to the console when enabled

### Event Structure

Each security event includes:

- **Timestamp**: When the event occurred
- **Level**: INFO, WARNING, ERROR, or CRITICAL
- **Category**: The event category (tls, signature, auth, plugin, system)
- **EventType**: Specific event type (e.g., AUTH_SUCCESS, TLS_HANDSHAKE)
- **Message**: Human-readable description of the event
- **Component**: The component that generated the event
- **Success**: Whether the operation was successful
- **Details**: Additional information about the event

## Using the Security Monitor Tool

The `securitymon` tool allows you to view, analyze, and monitor security events in real-time.

### Installation

```bash
# Build the security monitor tool
go build -o bin/securitymon ./cmd/securitymon
```

### Basic Usage

```bash
# View the latest security events
./bin/securitymon

# Follow the latest security events in real-time
./bin/securitymon -follow

# Monitor for new security events
./bin/securitymon -watch
```

### Filtering Events

```bash
# Filter by event level
./bin/securitymon -level=ERROR

# Filter by event type
./bin/securitymon -type=AUTH_FAILURE,TLS_HANDSHAKE

# Limit the number of events displayed
./bin/securitymon -limit=50
```

### Output Formats

```bash
# Display events in JSON format
./bin/securitymon -format=json

# Display events in CSV format
./bin/securitymon -format=csv
```

## Security Event Categories

### Authentication Events

| Event Type | Description |
|------------|-------------|
| `AUTH_SUCCESS` | Successful authentication |
| `AUTH_FAILURE` | Failed authentication attempt |
| `APIKEY_CREATED` | API key created |
| `APIKEY_REVOKED` | API key revoked |
| `ROLE_CHANGE` | Role assignment changed |
| `PERMISSION_DENIED` | Permission denied for operation |

### TLS Events

| Event Type | Description |
|------------|-------------|
| `TLS_HANDSHAKE` | TLS handshake attempt |
| `CERT_GENERATED` | Certificate generated |
| `CERT_EXPIRED` | Certificate expired |
| `TLS_VERIFICATION` | Certificate verification |

### Signature Events

| Event Type | Description |
|------------|-------------|
| `SIG_VERIFY` | Signature verification attempt |
| `PLUGIN_SIGNED` | Plugin signed |
| `SIG_INVALID` | Invalid signature detected |
| `KEY_GENERATED` | Signing key generated |
| `KEY_TRUSTED` | Key added to trusted keys |
| `KEY_REVOKED` | Key revoked |

### Plugin Events

| Event Type | Description |
|------------|-------------|
| `PLUGIN_LOADED` | Plugin loaded |
| `PLUGIN_UNLOADED` | Plugin unloaded |
| `PLUGIN_CRASHED` | Plugin crashed |

### System Events

| Event Type | Description |
|------------|-------------|
| `SYSTEM_STARTUP` | System started |
| `SYSTEM_SHUTDOWN` | System shutting down |
| `CONFIG_CHANGED` | Configuration changed |
| `AUDIT_STARTED` | Security audit started |
| `AUDIT_COMPLETED` | Security audit completed |

## Setting Up Security Alerts

### Creating Event Callbacks

You can register callbacks for specific event types to trigger alerts or actions:

```go
// Create a security event manager
eventManager, err := security.NewSecurityEventManager("/var/log/snoozebot/security", logger)
if err != nil {
    // Handle error
}

// Register a callback for authentication failures
eventManager.RegisterCallback(security.EventAuthFailure, func(event *security.SecurityEvent) {
    // Send an alert or take action
    SendAlert(fmt.Sprintf("Authentication failure: %s", event.Message))
})
```

### Integrating with External Systems

The security event log files are in JSON format, making them easy to integrate with external monitoring systems:

```bash
# Send security events to an external log aggregator
tail -f /var/log/snoozebot/security/security-latest.log | your-log-forwarder
```

### Email Alerts

You can configure email alerts for critical security events:

```bash
# Watch for critical events and send email alerts
./bin/securitymon -level=CRITICAL -watch |
    grep -B2 "\[CRITICAL\]" |
    mail -s "Snoozebot Security Alert" admin@example.com
```

## Log Rotation and Retention

Security event logs are automatically rotated based on size (10MB by default) and quantity (5 files by default).

You can configure log rotation settings in your application:

```go
// Set log rotation to 5MB and 10 files
eventManager.SetRotationSettings(5*1024*1024, 10)
```

## Best Practices

1. **Regular Monitoring**: Check security events regularly for suspicious activity
2. **Alert Configuration**: Set up alerts for critical security events
3. **Log Retention**: Configure log retention based on your security policy
4. **Periodic Audits**: Perform periodic security audits using the event logs
5. **Integrate with SIEM**: Forward security events to your security information and event management system

## Troubleshooting

If you're having issues with security event logging:

1. **Check Permissions**: Ensure the application has write access to the log directory
2. **Verify Configuration**: Check that security event logging is enabled
3. **Increase Log Level**: Set the log level to DEBUG for more detailed information
4. **Check Disk Space**: Ensure sufficient disk space for log files

For more detailed troubleshooting, see the [Troubleshooting Guide](./TROUBLESHOOTING.md).