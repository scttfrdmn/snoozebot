package plugin

import (
	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/version"
)

// BaseProvider provides common functionality for all cloud providers
type BaseProvider struct {
	logger        hclog.Logger
	name          string
	version       string
	apiVersion    string
	capabilities  []string
	manifest      *version.PluginManifest
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name, pluginVersion string, logger hclog.Logger) *BaseProvider {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}
	
	return &BaseProvider{
		logger:       logger,
		name:         name,
		version:      pluginVersion,
		apiVersion:   CurrentAPIVersion,
		capabilities: []string{},
	}
}

// GetProviderName returns the name of the cloud provider
func (p *BaseProvider) GetProviderName() string {
	return p.name
}

// GetProviderVersion returns the version of the cloud provider plugin
func (p *BaseProvider) GetProviderVersion() string {
	return p.version
}

// GetAPIVersion returns the API version implemented by the plugin
func (p *BaseProvider) GetAPIVersion() string {
	return p.apiVersion
}

// Shutdown performs any cleanup when the plugin is being unloaded
func (p *BaseProvider) Shutdown() {
	p.logger.Info("Shutting down provider", "name", p.name)
}

// AddCapability adds a capability to the provider
func (p *BaseProvider) AddCapability(capability string) {
	p.capabilities = append(p.capabilities, capability)
}

// HasCapability checks if the provider has a capability
func (p *BaseProvider) HasCapability(capability string) bool {
	for _, c := range p.capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// GetManifest returns the plugin manifest
func (p *BaseProvider) GetManifest() *version.PluginManifest {
	if p.manifest == nil {
		p.manifest = version.NewPluginManifest(p.name, p.version, "")
		p.manifest.Capabilities = p.capabilities
	}
	return p.manifest
}

// SetManifest sets the plugin manifest
func (p *BaseProvider) SetManifest(manifest *version.PluginManifest) {
	p.manifest = manifest
}

// CheckVersionCompatibility checks if the plugin API version is compatible with the host
func (p *BaseProvider) CheckVersionCompatibility() (bool, error) {
	return version.IsCompatible(p.apiVersion)
}