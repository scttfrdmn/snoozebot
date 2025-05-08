package tls

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TLSManager manages TLS certificates for plugins
type TLSManager struct {
	certDir     string
	ca          *CertificateAuthority
	mutex       sync.RWMutex
	tlsConfigs  map[string]*tls.Config
	initialized bool
}

// NewTLSManager creates a new TLS manager
func NewTLSManager(certDir string) (*TLSManager, error) {
	// Create certificate directory if it doesn't exist
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %w", err)
	}

	return &TLSManager{
		certDir:    certDir,
		tlsConfigs: make(map[string]*tls.Config),
	}, nil
}

// Initialize initializes the TLS manager
func (m *TLSManager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Return if already initialized
	if m.initialized {
		return nil
	}

	// Ensure CA exists
	ca, err := EnsureCA(m.certDir)
	if err != nil {
		return fmt.Errorf("failed to ensure CA: %w", err)
	}
	m.ca = ca

	m.initialized = true
	return nil
}

// IsInitialized returns whether the TLS manager is initialized
func (m *TLSManager) IsInitialized() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.initialized
}

// GetCA returns the certificate authority
func (m *TLSManager) GetCA() (*CertificateAuthority, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("TLS manager not initialized")
	}

	return m.ca, nil
}

// EnsurePluginCertificate ensures that a plugin has a certificate
func (m *TLSManager) EnsurePluginCertificate(pluginName string) (string, string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		if err := m.Initialize(); err != nil {
			return "", "", fmt.Errorf("failed to initialize TLS manager: %w", err)
		}
	}

	// Check if plugin certificate exists
	pluginCertDir := filepath.Join(m.certDir, pluginName)
	certFile := filepath.Join(pluginCertDir, "cert.pem")
	keyFile := filepath.Join(pluginCertDir, "key.pem")

	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			// Certificate and key exist
			return certFile, keyFile, nil
		}
	}

	// Generate new plugin certificate
	return GeneratePluginCertificate(pluginName, m.certDir, m.ca)
}

// GetPluginTLSConfig returns a TLS config for a plugin
func (m *TLSManager) GetPluginTLSConfig(pluginName string) (*tls.Config, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if we already have a TLS config for this plugin
	if tlsConfig, ok := m.tlsConfigs[pluginName]; ok {
		return tlsConfig, nil
	}

	if !m.initialized {
		if err := m.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize TLS manager: %w", err)
		}
	}

	// Ensure plugin certificate exists
	certFile, keyFile, err := m.EnsurePluginCertificate(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure plugin certificate: %w", err)
	}

	// Load TLS config
	caCertFile := filepath.Join(m.certDir, "ca", "cert.pem")
	tlsConfig, err := LoadTLSConfig(certFile, keyFile, caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	// Cache TLS config
	m.tlsConfigs[pluginName] = tlsConfig

	return tlsConfig, nil
}

// GetClientTLSConfig returns a TLS config for a client
func (m *TLSManager) GetClientTLSConfig() (*tls.Config, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		if err := m.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize TLS manager: %w", err)
		}
	}

	// Ensure client certificate exists
	certFile, keyFile, err := m.EnsurePluginCertificate("client")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure client certificate: %w", err)
	}

	// Load TLS config
	caCertFile := filepath.Join(m.certDir, "ca", "cert.pem")
	tlsConfig, err := LoadTLSConfig(certFile, keyFile, caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	// Set server name to localhost
	tlsConfig.ServerName = "localhost"

	return tlsConfig, nil
}

// GetServerTLSConfig returns a TLS config for a server
func (m *TLSManager) GetServerTLSConfig() (*tls.Config, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		if err := m.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize TLS manager: %w", err)
		}
	}

	// Ensure server certificate exists
	certFile, keyFile, err := m.EnsurePluginCertificate("server")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure server certificate: %w", err)
	}

	// Load TLS config
	caCertFile := filepath.Join(m.certDir, "ca", "cert.pem")
	tlsConfig, err := LoadTLSConfig(certFile, keyFile, caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	return tlsConfig, nil
}

// GetCACertificate returns the CA certificate
func (m *TLSManager) GetCACertificate() ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("TLS manager not initialized")
	}

	return m.ca.CertPEM, nil
}

// CleanupPluginCertificate removes a plugin's certificate
func (m *TLSManager) CleanupPluginCertificate(pluginName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Remove from cache
	delete(m.tlsConfigs, pluginName)

	// Remove certificate directory
	pluginCertDir := filepath.Join(m.certDir, pluginName)
	if err := os.RemoveAll(pluginCertDir); err != nil {
		return fmt.Errorf("failed to remove plugin certificate directory: %w", err)
	}

	return nil
}