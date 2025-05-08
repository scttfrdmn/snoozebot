package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/version"
)

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "version-check",
		Level: hclog.Info,
	})

	// Check for command line arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "check":
		// Check if a version is compatible
		if len(os.Args) < 3 {
			logger.Error("Missing version argument")
			printUsage()
			os.Exit(1)
		}
		
		pluginVersion := os.Args[2]
		
		logger.Info("Checking if version is compatible", 
			"plugin_version", pluginVersion,
			"current_version", version.CurrentVersion)
		
		compatible, err := version.IsCompatible(pluginVersion)
		if err != nil {
			logger.Error("Error checking compatibility", "error", err)
			os.Exit(1)
		}
		
		if compatible {
			logger.Info("Version is compatible")
			fmt.Println("✅ Version", pluginVersion, "is compatible with current API version", version.CurrentVersion)
		} else {
			logger.Error("Version is not compatible")
			fmt.Println("❌ Version", pluginVersion, "is NOT compatible with current API version", version.CurrentVersion)
			os.Exit(1)
		}

	case "info":
		// Display version information
		logger.Info("Displaying version information")
		fmt.Println("Snoozebot API Version Information:")
		fmt.Println("-----------------------------------")
		fmt.Println("Current API Version:", version.CurrentVersion)
		fmt.Println("Minimum Compatible Version:", version.MinimumCompatible)
		fmt.Println("Protocol Version:", 1)

	case "manifest":
		// Check, create or validate a manifest
		if len(os.Args) < 3 {
			logger.Error("Missing manifest file path")
			printUsage()
			os.Exit(1)
		}
		
		manifestPath := os.Args[2]
		
		// Check if we're creating a new manifest or validating an existing one
		if len(os.Args) >= 4 && os.Args[3] == "create" {
			// Create a new manifest
			logger.Info("Creating new manifest", "path", manifestPath)
			
			name := "example-provider"
			pluginVersion := "0.1.0"
			description := "Example cloud provider plugin"
			
			if len(os.Args) >= 5 {
				name = os.Args[4]
			}
			if len(os.Args) >= 6 {
				pluginVersion = os.Args[5]
			}
			if len(os.Args) >= 7 {
				description = os.Args[6]
			}
			
			manifest := version.NewPluginManifest(name, pluginVersion, description)
			manifest.Author = "Your Name"
			manifest.License = "Apache-2.0"
			
			// Save to file
			dir := filepath.Dir(manifestPath)
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				logger.Error("Failed to create directory", "error", err)
				os.Exit(1)
			}
			
			data, err := json.MarshalIndent(manifest, "", "  ")
			if err != nil {
				logger.Error("Failed to marshal manifest", "error", err)
				os.Exit(1)
			}
			
			err = os.WriteFile(manifestPath, data, 0644)
			if err != nil {
				logger.Error("Failed to write manifest", "error", err)
				os.Exit(1)
			}
			
			logger.Info("Manifest created successfully", "path", manifestPath)
			fmt.Println("✅ Manifest created at", manifestPath)
			
		} else {
			// Validate an existing manifest
			logger.Info("Validating manifest", "path", manifestPath)
			
			manifest, err := version.LoadManifest(manifestPath)
			if err != nil {
				logger.Error("Failed to load manifest", "error", err)
				os.Exit(1)
			}
			
			// Display manifest info
			fmt.Println("Manifest Information:")
			fmt.Println("---------------------")
			fmt.Println("Name:", manifest.Name)
			fmt.Println("Version:", manifest.Version)
			fmt.Println("API Version:", manifest.APIVersion)
			fmt.Println("Description:", manifest.Description)
			fmt.Println("Author:", manifest.Author)
			fmt.Println("License:", manifest.License)
			fmt.Println("Capabilities:", manifest.Capabilities)
			
			// Check compatibility
			compatible, err := manifest.IsCompatibleWithHost()
			if err != nil {
				logger.Error("Error checking compatibility", "error", err)
				os.Exit(1)
			}
			
			if compatible {
				logger.Info("Manifest is compatible")
				fmt.Println("\n✅ Manifest API version is compatible with current API version", version.CurrentVersion)
			} else {
				logger.Error("Manifest is not compatible")
				fmt.Println("\n❌ Manifest API version is NOT compatible with current API version", version.CurrentVersion)
				os.Exit(1)
			}
		}

	default:
		logger.Error("Unknown command", "command", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Snoozebot Version Check Tool")
	fmt.Println("---------------------------")
	fmt.Println("Usage:")
	fmt.Println("  versioncheck check <version>             - Check if a version is compatible")
	fmt.Println("  versioncheck info                       - Display version information")
	fmt.Println("  versioncheck manifest <path>            - Validate a manifest file")
	fmt.Println("  versioncheck manifest <path> create [name] [version] [description] - Create a new manifest file")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  versioncheck check 1.0.0")
	fmt.Println("  versioncheck info")
	fmt.Println("  versioncheck manifest /path/to/manifest.json")
	fmt.Println("  versioncheck manifest /path/to/manifest.json create my-provider 0.1.0 \"My cloud provider\"")
}