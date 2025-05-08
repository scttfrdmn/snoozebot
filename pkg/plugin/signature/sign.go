package signature

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
)

// PluginSigner signs plugin binaries
type PluginSigner struct {
	sigService SignatureService
	logger     hclog.Logger
}

// NewPluginSigner creates a new plugin signer
func NewPluginSigner(sigService SignatureService, logger hclog.Logger) *PluginSigner {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	return &PluginSigner{
		sigService: sigService,
		logger:     logger,
	}
}

// SignPlugin signs a plugin binary
func (s *PluginSigner) SignPlugin(pluginPath, keyID string) error {
	// Check if the plugin exists
	info, err := os.Stat(pluginPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", pluginPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat plugin file: %w", err)
	}

	// Check if the file is executable
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("plugin file is not executable: %s", pluginPath)
	}

	// Extract the plugin name from the path
	pluginName := strings.TrimSuffix(filepath.Base(pluginPath), filepath.Ext(pluginPath))

	// Sign the plugin
	_, err = s.sigService.SignPlugin(pluginName, pluginPath, keyID)
	if err != nil {
		return fmt.Errorf("failed to sign plugin: %w", err)
	}

	s.logger.Info("Plugin signed successfully", "plugin", pluginName, "path", pluginPath, "key_id", keyID)
	return nil
}

// SignPluginsInDirectory signs all plugins in a directory
func (s *PluginSigner) SignPluginsInDirectory(dirPath, keyID string) error {
	// Check if the directory exists
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", dirPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dirPath)
	}

	// Get all files in the directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Sign each executable file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		fileInfo, err := file.Info()
		if err != nil {
			s.logger.Warn("Failed to get file info", "file", filePath, "error", err)
			continue
		}

		// Check if the file is executable
		if fileInfo.Mode()&0111 != 0 {
			if err := s.SignPlugin(filePath, keyID); err != nil {
				s.logger.Error("Failed to sign plugin", "file", filePath, "error", err)
				// Continue with other plugins
			}
		}
	}

	return nil
}

// CreateAndSignPluginBundle creates a signed plugin bundle
func (s *PluginSigner) CreateAndSignPluginBundle(pluginPath, bundlePath, keyID string) error {
	// Copy the plugin to the bundle path
	if err := copyFile(pluginPath, bundlePath); err != nil {
		return fmt.Errorf("failed to copy plugin to bundle: %w", err)
	}

	// Get the plugin name
	pluginName := strings.TrimSuffix(filepath.Base(pluginPath), filepath.Ext(pluginPath))

	// Sign the plugin
	signature, err := s.sigService.SignPlugin(pluginName, bundlePath, keyID)
	if err != nil {
		return fmt.Errorf("failed to sign plugin: %w", err)
	}

	// Create the signature file path
	signaturePath := bundlePath + ".sig"

	// Save the signature
	if err := signature.SaveSignature(signaturePath); err != nil {
		return fmt.Errorf("failed to save signature: %w", err)
	}

	s.logger.Info("Plugin bundle created and signed",
		"plugin", pluginName,
		"bundle", bundlePath,
		"signature", signaturePath,
		"key_id", keyID)

	return nil
}

// VerifyPluginBundle verifies a plugin bundle
func (s *PluginSigner) VerifyPluginBundle(bundlePath string) error {
	// Get the plugin name
	pluginName := strings.TrimSuffix(filepath.Base(bundlePath), filepath.Ext(bundlePath))

	// Create the signature file path
	signaturePath := bundlePath + ".sig"

	// Check if the signature file exists
	if _, err := os.Stat(signaturePath); os.IsNotExist(err) {
		return fmt.Errorf("signature file not found: %s", signaturePath)
	}

	// Load the signature
	signature, err := LoadSignature(signaturePath)
	if err != nil {
		return fmt.Errorf("failed to load signature: %w", err)
	}

	// Check if the key is trusted
	trusted, err := s.sigService.IsTrustedKey(signature.KeyID)
	if err != nil {
		return fmt.Errorf("failed to check if key is trusted: %w", err)
	}

	if !trusted {
		return fmt.Errorf("signature key is not trusted: %s", signature.KeyID)
	}

	// Get the signing key
	key, err := s.sigService.GetSigningKey(signature.KeyID)
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}

	// Verify the signature
	if err := signature.Verify(bundlePath, key); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	s.logger.Info("Plugin bundle verified successfully",
		"plugin", pluginName,
		"bundle", bundlePath,
		"key_id", signature.KeyID,
		"issuer", signature.Issuer)

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Read the source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Create the destination directory if it doesn't exist
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Write the destination file
	if err := os.WriteFile(dst, data, 0755); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}