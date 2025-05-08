package tls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TLSManager manages TLS certificates for plugins
type TLSManager struct {
	certDir             string
	ca                  *CertificateAuthority
	mutex               sync.RWMutex
	tlsConfigs          map[string]*tls.Config
	certCache           map[string]certCacheEntry
	clientTLSConfig     *tls.Config
	serverTLSConfig     *tls.Config
	initialized         bool
	caCertPEMCache      []byte
	certExpirationCheck time.Time
}

// certCacheEntry holds a cached certificate info
type certCacheEntry struct {
	certFile  string
	keyFile   string
	timestamp time.Time
	tlsConfig *tls.Config
}

// NewTLSManager creates a new TLS manager
func NewTLSManager(certDir string) (*TLSManager, error) {
	// Create certificate directory if it doesn't exist
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %w", err)
	}

	return &TLSManager{
		certDir:             certDir,
		tlsConfigs:          make(map[string]*tls.Config),
		certCache:           make(map[string]certCacheEntry),
		certExpirationCheck: time.Now(),
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
	
	// Cache the CA certificate PEM for faster access
	m.caCertPEMCache = ca.CertPEM

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

// checkCertificateExpiration checks if it's time to validate certificate expiration
func (m *TLSManager) checkCertificateExpiration() bool {
	// Only check once per day to avoid filesystem operations
	if time.Since(m.certExpirationCheck) > 24*time.Hour {
		m.certExpirationCheck = time.Now()
		return true
	}
	return false
}

// EnsurePluginCertificate ensures that a plugin has a certificate
func (m *TLSManager) EnsurePluginCertificate(pluginName string) (string, string, error) {
	// Check cache first with read lock
	m.mutex.RLock()
	if entry, ok := m.certCache[pluginName]; ok {
		// Check if we need to verify expiration
		if !m.checkCertificateExpiration() {
			// Return cached values if we don't need to check expiration
			m.mutex.RUnlock()
			return entry.certFile, entry.keyFile, nil
		}
	}
	m.mutex.RUnlock()

	// Acquire write lock for generating or checking expiration
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Re-check cache after acquiring write lock (double-check pattern)
	if entry, ok := m.certCache[pluginName]; ok && !m.checkCertificateExpiration() {
		return entry.certFile, entry.keyFile, nil
	}

	if !m.initialized {
		if err := m.Initialize(); err != nil {
			return "", "", fmt.Errorf("failed to initialize TLS manager: %w", err)
		}
	}

	// Check if plugin certificate exists
	pluginCertDir := filepath.Join(m.certDir, pluginName)
	certFile := filepath.Join(pluginCertDir, "cert.pem")
	keyFile := filepath.Join(pluginCertDir, "key.pem")

	// Check if certificate exists and is valid
	certExists := false
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			certExists = true
			
			// Check certificate expiration if it's time to do so
			if m.checkCertificateExpiration() {
				// Read certificate to check expiration
				certPEM, err := os.ReadFile(certFile)
				if err == nil {
					certBlock, _ := pem.Decode(certPEM)
					if certBlock != nil {
						cert, err := x509.ParseCertificate(certBlock.Bytes)
						if err == nil {
							// If certificate is expired or will expire within 30 days, regenerate it
							if time.Now().After(cert.NotAfter) || time.Until(cert.NotAfter) < 30*24*time.Hour {
								certExists = false
							}
						}
					}
				}
			}
		}
	}

	if certExists {
		// Cache the certificate information
		m.certCache[pluginName] = certCacheEntry{
			certFile:  certFile,
			keyFile:   keyFile,
			timestamp: time.Now(),
		}
		return certFile, keyFile, nil
	}

	// Generate new plugin certificate
	certFile, keyFile, err := GeneratePluginCertificate(pluginName, m.certDir, m.ca)
	if err != nil {
		return "", "", err
	}
	
	// Cache the certificate information
	m.certCache[pluginName] = certCacheEntry{
		certFile:  certFile,
		keyFile:   keyFile,
		timestamp: time.Now(),
	}
	
	return certFile, keyFile, nil
}

// GetPluginTLSConfig returns a TLS config for a plugin
func (m *TLSManager) GetPluginTLSConfig(pluginName string) (*tls.Config, error) {
	// Check cache with read lock first
	m.mutex.RLock()
	if entry, ok := m.certCache[pluginName]; ok && entry.tlsConfig != nil {
		tlsConfig := entry.tlsConfig
		m.mutex.RUnlock()
		return tlsConfig, nil
	}
	
	// Also check tlsConfigs cache for backward compatibility
	if tlsConfig, ok := m.tlsConfigs[pluginName]; ok {
		m.mutex.RUnlock()
		return tlsConfig, nil
	}
	m.mutex.RUnlock()
	
	// Need to create a new TLS config, acquire write lock
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Double-check the cache after acquiring the write lock
	if entry, ok := m.certCache[pluginName]; ok && entry.tlsConfig != nil {
		return entry.tlsConfig, nil
	}
	
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

	// Cache TLS config in both caches
	m.tlsConfigs[pluginName] = tlsConfig
	
	// Update the cert cache as well
	if entry, ok := m.certCache[pluginName]; ok {
		entry.tlsConfig = tlsConfig
		m.certCache[pluginName] = entry
	} else {
		m.certCache[pluginName] = certCacheEntry{
			certFile:  certFile,
			keyFile:   keyFile,
			timestamp: time.Now(),
			tlsConfig: tlsConfig,
		}
	}

	return tlsConfig, nil
}

// GetClientTLSConfig returns a TLS config for a client
func (m *TLSManager) GetClientTLSConfig() (*tls.Config, error) {
	// Check if we have a cached client TLS config
	m.mutex.RLock()
	if m.clientTLSConfig != nil {
		tlsConfig := m.clientTLSConfig
		m.mutex.RUnlock()
		return tlsConfig, nil
	}
	m.mutex.RUnlock()
	
	// Need to create a new TLS config
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check again after acquiring write lock
	if m.clientTLSConfig != nil {
		return m.clientTLSConfig, nil
	}

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
	
	// Cache the client TLS config
	m.clientTLSConfig = tlsConfig

	return tlsConfig, nil
}

// GetServerTLSConfig returns a TLS config for a server
func (m *TLSManager) GetServerTLSConfig() (*tls.Config, error) {
	// Check if we have a cached server TLS config
	m.mutex.RLock()
	if m.serverTLSConfig != nil {
		tlsConfig := m.serverTLSConfig
		m.mutex.RUnlock()
		return tlsConfig, nil
	}
	m.mutex.RUnlock()
	
	// Need to create a new TLS config
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check again after acquiring write lock
	if m.serverTLSConfig != nil {
		return m.serverTLSConfig, nil
	}

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
	
	// Cache the server TLS config
	m.serverTLSConfig = tlsConfig

	return tlsConfig, nil
}

// GetCACertificate returns the CA certificate
func (m *TLSManager) GetCACertificate() ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("TLS manager not initialized")
	}

	// Use cached PEM
	if m.caCertPEMCache != nil {
		return m.caCertPEMCache, nil
	}

	return m.ca.CertPEM, nil
}

// CleanupPluginCertificate removes a plugin's certificate
func (m *TLSManager) CleanupPluginCertificate(pluginName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Remove from caches
	delete(m.tlsConfigs, pluginName)
	delete(m.certCache, pluginName)
	
	// Reset client/server TLS config if they match the plugin being cleaned up
	if pluginName == "client" {
		m.clientTLSConfig = nil
	} else if pluginName == "server" {
		m.serverTLSConfig = nil
	}

	// Remove certificate directory
	pluginCertDir := filepath.Join(m.certDir, pluginName)
	if err := os.RemoveAll(pluginCertDir); err != nil {
		return fmt.Errorf("failed to remove plugin certificate directory: %w", err)
	}

	return nil
}