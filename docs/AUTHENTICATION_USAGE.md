# Plugin Authentication Usage Guide

This document provides instructions for using the Snoozebot plugin authentication system.

## Overview

The Snoozebot plugin authentication system provides security for plugin operations through:

- API key authentication for plugins
- Role-based permission control
- Secure key generation and validation
- REST API for authentication management

## Configuration

Authentication is disabled by default and can be enabled in two ways:

1. Via command line flag:
   ```
   snooze-agent --enable-auth
   ```

2. Via the REST API:
   ```
   curl -X POST http://localhost:8080/api/auth/enable
   ```

## API Key Management

### Generating API Keys

Generate an API key for a plugin using the API:

```bash
curl -X POST http://localhost:8080/api/auth/apikey \
  -H "Content-Type: application/json" \
  -d '{
    "plugin_name": "aws",
    "role": "cloud_provider",
    "description": "AWS plugin key",
    "expires_in_days": 365
  }'
```

The response will include the API key:

```json
{
  "plugin_name": "aws",
  "api_key": "generated-api-key-value",
  "role": "cloud_provider",
  "description": "AWS plugin key",
  "expires_in_days": 365,
  "created_at": "2025-05-08T12:00:00Z",
  "expires_at": "2026-05-08T12:00:00Z"
}
```

### Revoking API Keys

Revoke an API key for a plugin:

```bash
curl -X POST http://localhost:8080/api/auth/apikey/revoke \
  -H "Content-Type: application/json" \
  -d '{
    "plugin_name": "aws"
  }'
```

### Loading Plugins with Authentication

When authentication is enabled, you must provide the API key when loading a plugin:

```bash
curl -X POST http://localhost:8080/api/plugins/load \
  -H "Content-Type: application/json" \
  -d '{
    "plugin_name": "aws",
    "api_key": "your-api-key-here",
    "timeout_seconds": 30,
    "retries": 3
  }'
```

## Role-Based Permissions

Roles define what operations a plugin can perform. The default configuration includes the following roles:

### Cloud Provider Role

The `cloud_provider` role has these permissions:

```json
{
  "name": "cloud_provider",
  "description": "Role for cloud provider plugins",
  "permissions": [
    {
      "name": "cloud_operations",
      "description": "Permission to perform cloud operations",
      "allowed": true
    },
    {
      "name": "filesystem_read",
      "description": "Permission to read from filesystem",
      "allowed": true
    },
    {
      "name": "filesystem_write",
      "description": "Permission to write to filesystem",
      "allowed": false
    },
    {
      "name": "network_outbound",
      "description": "Permission to make outbound network connections",
      "allowed": true
    }
  ]
}
```

## Authentication Management API

### Check Authentication Status

```bash
curl -X GET http://localhost:8080/api/auth/status
```

Response:
```json
{
  "enabled": true
}
```

### Enable Authentication

```bash
curl -X POST http://localhost:8080/api/auth/enable
```

Response:
```json
{
  "enabled": true
}
```

### Disable Authentication

```bash
curl -X POST http://localhost:8080/api/auth/disable
```

Response:
```json
{
  "enabled": false
}
```

## Configuration File

The authentication configuration is stored in `/etc/snoozebot/config/security.json` by default.

Example configuration:
```json
{
  "auth": {
    "enabled": true,
    "api_keys": [
      {
        "plugin_name": "aws",
        "api_key": "aws_plugin_key_123",
        "role": "cloud_provider",
        "created_at": "2025-05-08T12:00:00Z",
        "expires_at": "2026-05-08T12:00:00Z",
        "description": "AWS plugin key"
      },
      {
        "plugin_name": "gcp",
        "api_key": "gcp_plugin_key_456",
        "role": "cloud_provider",
        "created_at": "2025-05-08T12:00:00Z",
        "expires_at": "2026-05-08T12:00:00Z",
        "description": "GCP plugin key"
      },
      {
        "plugin_name": "azure",
        "api_key": "azure_plugin_key_789",
        "role": "cloud_provider",
        "created_at": "2025-05-08T12:00:00Z",
        "expires_at": "2026-05-08T12:00:00Z",
        "description": "Azure plugin key"
      }
    ]
  },
  "roles": [
    {
      "name": "cloud_provider",
      "permissions": [
        {"name": "cloud_operations", "allowed": true},
        {"name": "filesystem_read", "allowed": true},
        {"name": "filesystem_write", "allowed": false},
        {"name": "network_outbound", "allowed": true}
      ]
    }
  ],
  "tls": {
    "enabled": false,
    "cert_path": "/etc/snoozebot/certs",
    "key_path": "/etc/snoozebot/keys"
  },
  "signatures": {
    "enabled": false,
    "public_key_path": "/etc/snoozebot/keys/public.pem",
    "verify_signatures": true
  }
}
```

## Security Considerations

- API keys are stored in the configuration file and should be protected.
- Enable TLS and file permissions to protect API keys.
- Regularly rotate API keys for production environments.
- For highly sensitive environments, consider implementing more advanced authentication methods.

## Future Security Enhancements

The plugin authentication system will be enhanced with:

1. TLS communication for securing plugin communication
2. Plugin signature verification for ensuring plugin integrity
3. Enhanced permission controls for more granular access control
4. Two-factor authentication for plugin operations
5. Audit logging for security-relevant events