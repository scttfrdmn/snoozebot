package util

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/version"
)

// CheckPluginAPICompatibility checks if a plugin's API version is compatible
// with the current API version. This is useful in plugins to verify
// compatibility before attempting to start.
func CheckPluginAPICompatibility(pluginName, pluginAPIVersion string, logger hclog.Logger) bool {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	// Check if we're just doing a version check
	versionCheckOnly := os.Getenv("SNOOZEBOT_VERSION_CHECK_ONLY") == "true"

	// Get the current API version
	currentAPIVersion := version.CurrentVersion

	logger.Info("Checking plugin API compatibility",
		"plugin", pluginName,
		"plugin_api_version", pluginAPIVersion,
		"current_api_version", currentAPIVersion,
	)

	// Check for compatibility
	compatible, err := version.IsCompatible(pluginAPIVersion)
	if err != nil {
		logger.Error("Failed to check API compatibility",
			"error", err,
			"plugin", pluginName,
		)
		if versionCheckOnly {
			fmt.Println("Error: Failed to check API compatibility:", err)
			os.Exit(1)
		}
		return false
	}

	if !compatible {
		logger.Error("Plugin API version is not compatible",
			"plugin", pluginName,
			"plugin_api_version", pluginAPIVersion,
			"current_api_version", currentAPIVersion,
		)
		if versionCheckOnly {
			fmt.Println("Error: Plugin API version is not compatible with current API version")
			os.Exit(1)
		}
		return false
	}

	logger.Info("Plugin API version is compatible",
		"plugin", pluginName,
		"plugin_api_version", pluginAPIVersion,
		"current_api_version", currentAPIVersion,
	)

	if versionCheckOnly {
		fmt.Println("Success: Plugin API version is compatible with current API version")
		os.Exit(0)
	}

	return true
}

// GetVersionedLDFlags returns ldflags for building a plugin with version information
func GetVersionedLDFlags(pluginVersion, gitCommit, buildTime string) string {
	return fmt.Sprintf(
		"-X github.com/scttfrdmn/snoozebot/pkg/plugin/version.BuildVersion=%s "+
			"-X github.com/scttfrdmn/snoozebot/pkg/plugin/version.GitCommit=%s "+
			"-X github.com/scttfrdmn/snoozebot/pkg/plugin/version.BuildTimestamp=%s",
		pluginVersion,
		gitCommit,
		buildTime,
	)
}