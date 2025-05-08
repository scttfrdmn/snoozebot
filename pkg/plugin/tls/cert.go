package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultCAValidDays is the default number of days a CA certificate is valid for
	DefaultCAValidDays = 365 * 10 // 10 years

	// DefaultCertValidDays is the default number of days a certificate is valid for
	DefaultCertValidDays = 365 * 2 // 2 years

	// DefaultRSABits is the default number of bits for RSA keys
	DefaultRSABits = 2048

	// DefaultOrganization is the default organization for certificates
	DefaultOrganization = "Snoozebot"
)

// CertificateAuthority represents a certificate authority
type CertificateAuthority struct {
	Cert     *x509.Certificate
	Key      *rsa.PrivateKey
	CertPEM  []byte
	KeyPEM   []byte
	CertFile string
	KeyFile  string
}

// CertificateConfig contains configuration for certificate generation
type CertificateConfig struct {
	CommonName    string
	Organization  string
	Country       string
	Province      string
	Locality      string
	StreetAddress string
	PostalCode    string
	ValidDays     int
	RSABits       int
	IsCA          bool
	CAFile        string
	CAKeyFile     string
	Names         []string // Common names to add as SANs
}

// DefaultCertificateConfig returns a default configuration
func DefaultCertificateConfig() *CertificateConfig {
	return &CertificateConfig{
		CommonName:    "Snoozebot CA",
		Organization:  DefaultOrganization,
		Country:       "US",
		Province:      "California",
		Locality:      "San Francisco",
		StreetAddress: "123 Main St",
		PostalCode:    "94105",
		ValidDays:     DefaultCAValidDays,
		RSABits:       DefaultRSABits,
		IsCA:          true,
	}
}

// GenerateCA generates a new Certificate Authority
func GenerateCA(config *CertificateConfig) (*CertificateAuthority, error) {
	if config == nil {
		config = DefaultCertificateConfig()
	}

	// Create a new private key
	key, err := rsa.GenerateKey(rand.Reader, config.RSABits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(config.ValidDays) * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:    config.CommonName,
			Organization:  []string{config.Organization},
			Country:       []string{config.Country},
			Province:      []string{config.Province},
			Locality:      []string{config.Locality},
			StreetAddress: []string{config.StreetAddress},
			PostalCode:    []string{config.PostalCode},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Self-sign the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// PEM encode the certificate and private key
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	// Parse the certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &CertificateAuthority{
		Cert:    cert,
		Key:     key,
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}, nil
}

// SaveCA saves the CA certificate and private key to files
func (ca *CertificateAuthority) SaveCA(certFile, keyFile string) error {
	// Create directories if they don't exist
	certDir := filepath.Dir(certFile)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	keyDir := filepath.Dir(keyFile)
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write certificate to file
	if err := os.WriteFile(certFile, ca.CertPEM, 0644); err != nil {
		return fmt.Errorf("failed to write certificate to file: %w", err)
	}
	ca.CertFile = certFile

	// Write private key to file
	if err := os.WriteFile(keyFile, ca.KeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key to file: %w", err)
	}
	ca.KeyFile = keyFile

	return nil
}

// GenerateCertificate generates a new certificate signed by the CA
func (ca *CertificateAuthority) GenerateCertificate(config *CertificateConfig) ([]byte, []byte, error) {
	// Create a new private key
	key, err := rsa.GenerateKey(rand.Reader, config.RSABits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(config.ValidDays) * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:    config.CommonName,
			Organization:  []string{config.Organization},
			Country:       []string{config.Country},
			Province:      []string{config.Province},
			Locality:      []string{config.Locality},
			StreetAddress: []string{config.StreetAddress},
			PostalCode:    []string{config.PostalCode},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Add Subject Alternative Names
	if len(config.Names) > 0 {
		template.DNSNames = append(template.DNSNames, config.Names...)
	}

	// Sign the certificate with the CA
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, ca.Cert, &key.PublicKey, ca.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// PEM encode the certificate and private key
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return certPEM, keyPEM, nil
}

// LoadCA loads a Certificate Authority from files
func LoadCA(certFile, keyFile string) (*CertificateAuthority, error) {
	// Read certificate file
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Read key file
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse certificate
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse private key
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to parse key PEM")
	}
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &CertificateAuthority{
		Cert:     cert,
		Key:      key,
		CertPEM:  certPEM,
		KeyPEM:   keyPEM,
		CertFile: certFile,
		KeyFile:  keyFile,
	}, nil
}

// GeneratePluginCertificate generates a certificate for a plugin
func GeneratePluginCertificate(pluginName, certDir string, ca *CertificateAuthority) (string, string, error) {
	// Create config for plugin certificate
	config := &CertificateConfig{
		CommonName:    pluginName,
		Organization:  DefaultOrganization,
		Country:       "US",
		Province:      "California",
		Locality:      "San Francisco",
		StreetAddress: "123 Main St",
		PostalCode:    "94105",
		ValidDays:     DefaultCertValidDays,
		RSABits:       DefaultRSABits,
		IsCA:          false,
		Names:         []string{pluginName, "localhost", "127.0.0.1"},
	}

	// Generate certificate
	certPEM, keyPEM, err := ca.GenerateCertificate(config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate plugin certificate: %w", err)
	}

	// Create plugin certificate directory
	pluginCertDir := filepath.Join(certDir, pluginName)
	if err := os.MkdirAll(pluginCertDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create plugin certificate directory: %w", err)
	}

	// Write certificate and key to files
	certFile := filepath.Join(pluginCertDir, "cert.pem")
	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write certificate to file: %w", err)
	}

	keyFile := filepath.Join(pluginCertDir, "key.pem")
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return "", "", fmt.Errorf("failed to write key to file: %w", err)
	}

	return certFile, keyFile, nil
}

// LoadTLSConfig loads a TLS configuration from certificate and key files
func LoadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate and key: %w", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Load CA certificate if provided
	if caFile != "" {
		// Read CA certificate
		caPEM, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		// Create cert pool and add CA certificate
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to append CA certificate to cert pool")
		}

		// Set cert pool in TLS config
		tlsConfig.RootCAs = certPool
		tlsConfig.ClientCAs = certPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

// EnsureCA ensures that a CA exists, creating one if necessary
func EnsureCA(certDir string) (*CertificateAuthority, error) {
	// Define CA certificate and key file paths
	caCertFile := filepath.Join(certDir, "ca", "cert.pem")
	caKeyFile := filepath.Join(certDir, "ca", "key.pem")

	// Check if CA files exist
	if _, err := os.Stat(caCertFile); err == nil {
		if _, err := os.Stat(caKeyFile); err == nil {
			// Load existing CA
			return LoadCA(caCertFile, caKeyFile)
		}
	}

	// CA files don't exist, generate a new CA
	config := DefaultCertificateConfig()
	ca, err := GenerateCA(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA: %w", err)
	}

	// Save CA to files
	if err := ca.SaveCA(caCertFile, caKeyFile); err != nil {
		return nil, fmt.Errorf("failed to save CA: %w", err)
	}

	return ca, nil
}