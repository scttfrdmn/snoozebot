# Snoozebot 0.1.0 Reversion Steps

This document provides detailed steps to revert the project from the current version (1.0.0) back to 0.1.0 as part of the release plan. This is a critical first step to ensure we have a stable foundation before proceeding with the full release plan.

## Prerequisites

Before beginning, ensure:
1. You have a clean working directory (`git status` shows no changes)
2. You have the latest code from the main branch
3. You have Go 1.18+ installed

## Reversion Steps

### 1. Update VERSION File

```bash
# Update the VERSION file
echo "0.1.0" > VERSION
```

### 2. Update API Version Constants

Edit the following files:

**pkg/plugin/plugin.go**
```go
// Update this constant
const (
    // CurrentAPIVersion is the current API version
    CurrentAPIVersion = "0.1.0"
    
    // ProtocolVersion is the protocol version
    ProtocolVersion = 1
)
```

**pkg/plugin/version/version.go**
```go
// Update these constants
const (
    // CurrentVersion is the current API version
    CurrentVersion = "0.1.0"
    
    // MinimumCompatible is the minimum compatible API version
    MinimumCompatible = "0.1.0"
)
```

### 3. Fix Version References

These files may reference the version and need to be updated:

```bash
# Find all files with "1.0.0" version references
grep -r "1.0.0" --include="*.go" .

# Find all files with "CurrentVersion" references
grep -r "CurrentVersion" --include="*.go" .

# Find all files with "CurrentAPIVersion" references
grep -r "CurrentAPIVersion" --include="*.go" .
```

### 4. Update Plugin Versions

All plugins should be updated to return the correct API version:

**plugins/aws/main.go**
```go
// GetAPIVersion returns the API version implemented by the plugin
func (p *AWSProvider) GetAPIVersion() string {
    return "0.1.0" // Or use snoozePlugin.CurrentAPIVersion
}
```

Make similar changes to:
- plugins/azure/main.go
- plugins/gcp/main.go
- examples/custom_plugin/main.go

### 5. Rebuild Manifests

```bash
# Clean existing manifests
rm -rf bin/manifests

# Rebuild plugins with new version
./scripts/build_versioned_plugins.sh
```

### 6. Update Documentation

The following files contain version information that should be updated:

- docs/API_VERSIONING.md
- README.md
- PR_MESSAGE_API_VERSIONING.md
- examples/custom_plugin/README.md
- IMPLEMENTATION_SUMMARY.md

### 7. Fix Compilation Errors

Detailed compilation error fixes are in the [Release Plan](./RELEASE_PLAN_0.1.0.md), but here are quick fixes for common errors:

**pkg/plugin/tls_plugin.go** missing imports:
```go
import (
    "context"
    "crypto/tls"
    "crypto/x509"  // Add this
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "net"          // Add this
    "os"
    "path/filepath" // Add this
    "strings"      // Add this
    
    "github.com/hashicorp/go-hclog"
    "github.com/hashicorp/go-plugin"
    "google.golang.org/grpc"
)
```

### 8. Verify Changes

After making these changes:

```bash
# Verify the version
cat VERSION
make version

# Attempt to build
make

# Run specific tests
go test ./pkg/plugin/version/...
```

### 9. Commit Changes

Once the reversion is complete and verified:

```bash
git add .
git commit -m "Revert to version 0.1.0 for stable release preparation"
```

## Troubleshooting

### Common Issues

1. **Build Fails After Reversion**: 
   - Check that all version constants are consistently set to "0.1.0"
   - Verify imports are fixed in tls_plugin.go

2. **Tests Fail**:
   - Some tests may have hardcoded "1.0.0" version expectations
   - Update test files to expect "0.1.0" instead

3. **Plugin Loading Fails**:
   - Check manifest files for correct version
   - Verify plugin version compatibility logic

## Next Steps

After completing the reversion, proceed to fixing all compilation errors as outlined in [Release Plan](./RELEASE_PLAN_0.1.0.md) Phase 1.