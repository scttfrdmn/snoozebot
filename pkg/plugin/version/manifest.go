package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PluginManifest contains metadata about a plugin
type PluginManifest struct {
	// API version that the plugin implements
	APIVersion string `json:"api_version"`
	
	// Plugin name
	Name string `json:"name"`
	
	// Plugin version
	Version string `json:"version"`
	
	// Plugin description
	Description string `json:"description,omitempty"`
	
	// Author information
	Author string `json:"author,omitempty"`
	
	// License information
	License string `json:"license,omitempty"`
	
	// Homepage URL
	Homepage string `json:"homepage,omitempty"`
	
	// Repository URL
	Repository string `json:"repository,omitempty"`
	
	// Build information
	BuildTimestamp string `json:"build_timestamp,omitempty"`
	GitCommit      string `json:"git_commit,omitempty"`
	
	// Compatibility information
	MinHostVersion string `json:"min_host_version,omitempty"`
	MaxHostVersion string `json:"max_host_version,omitempty"`
	
	// Dependencies
	Dependencies map[string]string `json:"dependencies,omitempty"`
	
	// Capabilities
	Capabilities []string `json:"capabilities,omitempty"`
	
	// Supported cloud providers
	SupportedProviders []string `json:"supported_providers,omitempty"`
}

// NewPluginManifest creates a new plugin manifest
func NewPluginManifest(name, version, description string) *PluginManifest {
	return &PluginManifest{
		APIVersion:     CurrentVersion,
		Name:           name,
		Version:        version,
		Description:    description,
		BuildTimestamp: time.Now().Format(time.RFC3339),
		Capabilities:   []string{},
		Dependencies:   make(map[string]string),
	}
}

// IsCompatibleWithHost checks if the plugin is compatible with the host
func (m *PluginManifest) IsCompatibleWithHost() (bool, error) {
	return IsCompatible(m.APIVersion)
}

// SaveManifest saves the manifest to a file
func SaveManifest(manifest *PluginManifest, directory string) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal the manifest to JSON
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	// Write the manifest to a file
	filePath := filepath.Join(directory, fmt.Sprintf("%s.manifest.json", manifest.Name))
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}
	
	return nil
}

// LoadManifest loads a manifest from a file
func LoadManifest(filePath string) (*PluginManifest, error) {
	// Read the manifest file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	
	// Unmarshal the manifest
	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}
	
	return &manifest, nil
}

// FindManifest finds a plugin manifest by name
func FindManifest(name, directory string) (*PluginManifest, error) {
	// Check if the manifest file exists
	filePath := filepath.Join(directory, fmt.Sprintf("%s.manifest.json", name))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest not found: %s", name)
	}
	
	// Load the manifest
	return LoadManifest(filePath)
}