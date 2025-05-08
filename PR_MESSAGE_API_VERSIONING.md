# API Versioning for Snoozebot Plugin System

This PR implements API versioning for the Snoozebot plugin system, ensuring backward compatibility and proper version management for future development.

## Changes

- Implemented semantic versioning (SemVer) for the plugin API
- Added version compatibility checking logic between plugins and host
- Created a plugin manifest system to store and communicate version information
- Updated proto definitions to include versioning information
- Implemented a base provider with common functionality including versioning support
- Added GetAPIVersion() method to the AWS plugin
- Created utility functions for checking plugin compatibility 
- Added documentation explaining the versioning system
- Created example custom plugin with full versioning support
- Added tools for version checking and interface validation
- Created scripts for building and testing versioned plugins

## Testing

The API versioning system has been tested with:
- Compatibility checks between different versions
- Plugin manifest generation and validation
- Loading plugins with version verification
- Integration with existing security features

## Documentation

- Added comprehensive API_VERSIONING.md documentation
- Updated POST_MERGE_STEPS.md to reflect completed work
- Created example plugin with documentation for plugin developers
- Added version checking tools with documentation

## Breaking Changes

None. This implementation adds version support without breaking compatibility with existing plugins. The VERSION is now set to 1.0.0 to match the API version.

## Next Steps

After merging:
1. Run `./scripts/build_versioned_plugins.sh` to rebuild all plugins with version information
2. Test plugin compatibility with `./scripts/test_plugin_compatibility.sh`
3. Review the API versioning documentation in `/docs/API_VERSIONING.md`

## Future Work

- Implement plugin auto-update mechanism for version compatibility
- Add dependency management between plugins
- Create plugin discovery and repository system