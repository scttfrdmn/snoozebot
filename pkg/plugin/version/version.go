// Package version provides versioning utilities for the Snoozebot plugin system
package version

import (
	"fmt"
	"strconv"
	"strings"
)

// APIVersion represents a semantic version for the plugin API
type APIVersion struct {
	Major int
	Minor int
	Patch int
}

// Version constants
const (
	// CurrentVersion is the current API version
	CurrentVersion = "0.1.0"
	
	// MinimumCompatible is the minimum compatible API version
	MinimumCompatible = "0.1.0"
)

// ParseVersion parses a version string into an APIVersion
func ParseVersion(version string) (APIVersion, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return APIVersion{}, fmt.Errorf("invalid version format: %s", version)
	}
	
	// Parse major version
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return APIVersion{}, fmt.Errorf("invalid major version: %s", parts[0])
	}
	
	// Parse minor version
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return APIVersion{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	
	// Parse patch version if provided
	patch := 0
	if len(parts) == 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return APIVersion{}, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}
	
	return APIVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// String returns the string representation of the version
func (v APIVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsCompatible checks if the given version is compatible with the current version
func IsCompatible(pluginVersion string) (bool, error) {
	// Parse versions
	current, err := ParseVersion(CurrentVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse current version: %w", err)
	}
	
	plugin, err := ParseVersion(pluginVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse plugin version: %w", err)
	}
	
	// For 1.x.y releases, plugin Major version must match current Major version
	// and plugin Minor version must be within a compatible range
	if current.Major == 1 {
		if plugin.Major != current.Major {
			return false, nil
		}
		
		// Plugin minor version must be >= minimum compatible version
		minCompat, err := ParseVersion(MinimumCompatible)
		if err != nil {
			return false, fmt.Errorf("failed to parse minimum compatible version: %w", err)
		}
		
		if plugin.Minor < minCompat.Minor {
			return false, nil
		}
		
		// Plugin minor version must be <= current version
		if plugin.Minor > current.Minor {
			return false, nil
		}
		
		return true, nil
	}
	
	// For future releases with major version >= 2, we'll require exact major version match
	// and plugin minor version <= current minor version
	if plugin.Major != current.Major {
		return false, nil
	}
	
	if plugin.Minor > current.Minor {
		return false, nil
	}
	
	return true, nil
}

// VersionInfo provides version info for a component
type VersionInfo struct {
	APIVersion     string `json:"api_version"`
	PluginVersion  string `json:"plugin_version"`
	PluginName     string `json:"plugin_name"`
	BuildTimestamp string `json:"build_timestamp,omitempty"`
	GitCommit      string `json:"git_commit,omitempty"`
	BuildFlags     string `json:"build_flags,omitempty"`
}

// DefaultVersionInfo returns a default version info
func DefaultVersionInfo(pluginName, pluginVersion string) VersionInfo {
	return VersionInfo{
		APIVersion:    CurrentVersion,
		PluginVersion: pluginVersion,
		PluginName:    pluginName,
	}
}

// IsPluginAPICompatible checks if a given plugin version info is compatible with the current API version
func IsPluginAPICompatible(pluginInfo VersionInfo) (bool, error) {
	return IsCompatible(pluginInfo.APIVersion)
}

// APICompatibilityError represents a version compatibility error
type APICompatibilityError struct {
	PluginName     string
	PluginVersion  string
	PluginAPIVersion string
	CurrentAPIVersion string
}

// Error returns the error message
func (e APICompatibilityError) Error() string {
	return fmt.Sprintf(
		"plugin API version %s is not compatible with current API version %s (plugin: %s, version: %s)",
		e.PluginAPIVersion,
		e.CurrentAPIVersion,
		e.PluginName,
		e.PluginVersion,
	)
}