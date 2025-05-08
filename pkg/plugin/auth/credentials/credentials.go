// Package credentials provides standardized credential handling for cloud providers
package credentials

import (
	"fmt"
	"os"
	"path/filepath"
)

// CredentialSource defines where credentials are loaded from
type CredentialSource string

const (
	// SourceEnv indicates credentials are loaded from environment variables
	SourceEnv CredentialSource = "environment"
	
	// SourceProfile indicates credentials are loaded from a profile/config file
	SourceProfile CredentialSource = "profile"
	
	// SourceInstance indicates credentials are loaded from instance metadata
	SourceInstance CredentialSource = "instance"
)

// CloudProviderType defines the cloud provider type
type CloudProviderType string

const (
	// ProviderAWS is Amazon Web Services
	ProviderAWS CloudProviderType = "aws"
	
	// ProviderAzure is Microsoft Azure
	ProviderAzure CloudProviderType = "azure"
	
	// ProviderGCP is Google Cloud Platform
	ProviderGCP CloudProviderType = "gcp"
)

// CredentialOptions contains options for loading credentials
type CredentialOptions struct {
	// Provider is the cloud provider type
	Provider CloudProviderType
	
	// Source is where to load credentials from
	Source CredentialSource
	
	// Profile is the profile name for profile-based credentials
	Profile string
	
	// Region is the default region to use
	Region string
	
	// Zone is the default zone to use (if applicable)
	Zone string
	
	// ProjectID is the GCP project ID (if applicable)
	ProjectID string
	
	// SubscriptionID is the Azure subscription ID (if applicable)
	SubscriptionID string
	
	// ResourceGroup is the Azure resource group (if applicable)
	ResourceGroup string
	
	// CredentialsFile is the path to a credentials file (for GCP service account)
	CredentialsFile string
}

// GetAWSEnvironment returns AWS credential environment variables
func GetAWSEnvironment(opts *CredentialOptions) map[string]string {
	env := make(map[string]string)
	
	// Set AWS profile if specified
	if opts.Source == SourceProfile && opts.Profile != "" {
		env["AWS_PROFILE"] = opts.Profile
	}
	
	// Set region if specified
	if opts.Region != "" {
		env["AWS_REGION"] = opts.Region
	}
	
	return env
}

// GetGCPEnvironment returns GCP credential environment variables
func GetGCPEnvironment(opts *CredentialOptions) (map[string]string, error) {
	env := make(map[string]string)
	
	// Set GCP project ID
	if opts.ProjectID != "" {
		env["PROJECT_ID"] = opts.ProjectID
	} else {
		return nil, fmt.Errorf("GCP project ID is required")
	}
	
	// Set zone if specified
	if opts.Zone != "" {
		env["ZONE"] = opts.Zone
	} else {
		env["ZONE"] = "us-central1-a" // Default zone
	}
	
	// Handle credential source
	switch opts.Source {
	case SourceEnv:
		// Environment variables are already set
	
	case SourceProfile:
		// For GCP, profile means a service account JSON file
		// Try common locations for service account file
		if opts.CredentialsFile != "" {
			// Use provided credentials file
			env["GOOGLE_APPLICATION_CREDENTIALS"] = opts.CredentialsFile
		} else if opts.Profile != "" {
			// Look for a profile-specific credentials file
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %w", err)
			}
			
			// Check in ~/.gcp/{profile}.json
			credsDir := filepath.Join(homeDir, ".gcp")
			credsFile := filepath.Join(credsDir, fmt.Sprintf("%s.json", opts.Profile))
			
			if _, err := os.Stat(credsFile); err == nil {
				env["GOOGLE_APPLICATION_CREDENTIALS"] = credsFile
			} else {
				// Check in ~/.config/gcloud/profiles/{profile}.json
				gcloudDir := filepath.Join(homeDir, ".config", "gcloud", "profiles")
				credsFile = filepath.Join(gcloudDir, fmt.Sprintf("%s.json", opts.Profile))
				
				if _, err := os.Stat(credsFile); err == nil {
					env["GOOGLE_APPLICATION_CREDENTIALS"] = credsFile
				} else {
					return nil, fmt.Errorf("could not find GCP credentials file for profile %s", opts.Profile)
				}
			}
		} else {
			return nil, fmt.Errorf("GCP credentials file or profile name is required")
		}
	
	case SourceInstance:
		// No special handling for instance metadata, GCP client will use it automatically
	
	default:
		return nil, fmt.Errorf("unsupported credential source: %s", opts.Source)
	}
	
	return env, nil
}

// GetAzureEnvironment returns Azure credential environment variables
func GetAzureEnvironment(opts *CredentialOptions) (map[string]string, error) {
	env := make(map[string]string)
	
	// Set Azure subscription ID
	if opts.SubscriptionID != "" {
		env["AZURE_SUBSCRIPTION_ID"] = opts.SubscriptionID
	} else {
		return nil, fmt.Errorf("Azure subscription ID is required")
	}
	
	// Set Azure resource group
	if opts.ResourceGroup != "" {
		env["AZURE_RESOURCE_GROUP"] = opts.ResourceGroup
	} else {
		return nil, fmt.Errorf("Azure resource group is required")
	}
	
	// Set region if specified
	if opts.Region != "" {
		env["AZURE_LOCATION"] = opts.Region
	}
	
	// Handle credential source
	switch opts.Source {
	case SourceEnv:
		// Environment variables are already set (AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET)
	
	case SourceProfile:
		if opts.Profile != "" {
			// Look for Azure profile in ~/.azure/profiles/{profile}.json
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %w", err)
			}
			
			profilesDir := filepath.Join(homeDir, ".azure", "profiles")
			profileFile := filepath.Join(profilesDir, fmt.Sprintf("%s.json", opts.Profile))
			
			if _, err := os.Stat(profileFile); err == nil {
				// Found profile file, mark it for the Azure SDK to find
				env["AZURE_AUTH_LOCATION"] = profileFile
			} else {
				return nil, fmt.Errorf("could not find Azure profile file for profile %s", opts.Profile)
			}
		} else {
			return nil, fmt.Errorf("Azure profile name is required")
		}
	
	case SourceInstance:
		// For Azure managed identity
		env["AZURE_USE_MSI"] = "true"
	
	default:
		return nil, fmt.Errorf("unsupported credential source: %s", opts.Source)
	}
	
	return env, nil
}

// ApplyCredentials sets the appropriate environment variables for the specified provider
func ApplyCredentials(opts *CredentialOptions) error {
	var env map[string]string
	var err error
	
	switch opts.Provider {
	case ProviderAWS:
		env = GetAWSEnvironment(opts)
	
	case ProviderGCP:
		env, err = GetGCPEnvironment(opts)
		if err != nil {
			return fmt.Errorf("failed to get GCP environment: %w", err)
		}
	
	case ProviderAzure:
		env, err = GetAzureEnvironment(opts)
		if err != nil {
			return fmt.Errorf("failed to get Azure environment: %w", err)
		}
	
	default:
		return fmt.Errorf("unsupported provider: %s", opts.Provider)
	}
	
	// Apply environment variables
	for key, value := range env {
		os.Setenv(key, value)
	}
	
	return nil
}

// LoadFromProfile loads credentials from a profile
func LoadFromProfile(provider CloudProviderType, profile string) (*CredentialOptions, error) {
	opts := &CredentialOptions{
		Provider: provider,
		Source:   SourceProfile,
		Profile:  profile,
	}
	
	// Load additional options from profile-specific configuration
	switch provider {
	case ProviderAWS:
		// AWS profiles are handled directly by the AWS SDK
		// We could parse ~/.aws/config here to get region information
	
	case ProviderGCP:
		// Load GCP profile information from ~/.gcp/{profile}.json
	
	case ProviderAzure:
		// Load Azure profile information from ~/.azure/profiles/{profile}.json
	
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	
	return opts, nil
}