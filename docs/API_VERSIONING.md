# API Versioning in Snoozebot

This document explains the API versioning system used in Snoozebot and provides guidelines for plugin developers.

## Version Format

Snoozebot uses [Semantic Versioning](https://semver.org/) (SemVer) for all API versions. Versions are formatted as:

```
MAJOR.MINOR.PATCH
```

Where:
- **MAJOR**: Incremented for incompatible API changes
- **MINOR**: Incremented for backward-compatible feature additions
- **PATCH**: Incremented for backward-compatible bug fixes

## Current Version

The current API version is `0.1.0`.

## Version Compatibility

Plugins must implement the `GetAPIVersion()` method to declare which API version they support. The host will check compatibility before loading plugins.

### Compatibility Rules

For version `1.x.y`:
- **Major version** must match exactly (plugin must be `1.x.y`)
- **Minor version** must be:
  - Greater than or equal to the minimum compatible version (`1.0.0`)
  - Less than or equal to the host's minor version

For future versions (when major version ≥ 2):
- **Major version** must match exactly
- **Minor version** must be less than or equal to the host's minor version

## Plugin Manifest

Each plugin must provide a manifest that includes version information. The manifest is used for:
- Compatibility checking
- Plugin capability discovery
- Dependency management

Example manifest:

```json
{
  "api_version": "1.0.0",
  "name": "aws-provider",
  "version": "0.1.0",
  "description": "AWS cloud provider for Snoozebot",
  "author": "Your Name",
  "license": "Apache-2.0",
  "capabilities": ["list_instances", "start_instance", "stop_instance"],
  "min_host_version": "0.1.0"
}
```

## Implementing Versioning in Plugins

### Required Methods

All plugins must implement:

```go
// GetAPIVersion returns the API version implemented by the plugin
func (p *YourProvider) GetAPIVersion() string {
    return "1.0.0" // or use the constant: snoozePlugin.CurrentAPIVersion
}
```

### Base Provider

Plugins should use the BaseProvider to handle common functionality:

```go
import (
    "github.com/scttfrdmn/snoozebot/pkg/plugin"
    "github.com/scttfrdmn/snoozebot/pkg/plugin/version"
)

type MyProvider struct {
    *plugin.BaseProvider
    // Custom fields...
}

func NewMyProvider() *MyProvider {
    base := plugin.NewBaseProvider("my-provider", "0.1.0", logger)
    return &MyProvider{
        BaseProvider: base,
        // Initialize custom fields...
    }
}
```

## Testing Version Compatibility

Use the `version.IsCompatible()` function to check compatibility:

```go
compatible, err := version.IsCompatible(pluginVersion)
if err != nil {
    // Handle error
}
if !compatible {
    // Handle incompatibility
}
```

### Security Considerations

We've integrated security scanning for the API versioning system:

1. **Dependency scanning**: Use `make security` or `./scripts/security_check.sh` to:
   - Check for outdated dependencies
   - Scan for known vulnerabilities with govulncheck
   - Perform security code analysis with gosec

2. **CI Integration**: The `.github/workflows/security-scan.yml` file provides automated security scanning on each commit and pull request.

3. **Regular Updates**: Follow the guidance in [Security Maintenance](./SECURITY_MAINTENANCE.md) to keep dependencies up to date.

## Breaking Changes Policy

- **Major Version**: We will only introduce breaking changes with a major version increment.
- **Minor Version**: New features that don't break existing APIs.
- **Patch Version**: Bug fixes that don't break existing APIs.

## Migration Guide

When we release a new major version, we will provide a migration guide with:
- List of breaking changes
- Code examples for upgrading
- Compatibility tools where possible

## Deprecation Policy

1. Features are first marked as deprecated with compiler warnings
2. Deprecated features remain for at least one minor version
3. Deprecated features are removed only in major version updates