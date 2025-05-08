package signature

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SignatureConfig represents the configuration for signature verification
type SignatureConfig struct {
	Enabled     bool     `json:"enabled"`
	TrustedKeys []string `json:"trusted_keys"`
	KeyDir      string   `json:"key_dir"`
}

// PluginSignature represents a plugin signature
type PluginSignature struct {
	// Version of the signature format
	Version string `json:"version"`
	
	// PluginName is the name of the plugin
	PluginName string `json:"plugin_name"`
	
	// PluginVersion is the version of the plugin
	PluginVersion string `json:"plugin_version"`
	
	// HashAlgorithm is the algorithm used to hash the plugin binary
	HashAlgorithm string `json:"hash_algorithm"`
	
	// Hash is the base64-encoded hash of the plugin binary
	Hash string `json:"hash"`
	
	// SignatureValue is the base64-encoded signature of the hash
	SignatureValue string `json:"signature"`
	
	// SignatureAlgorithm is the algorithm used to sign the hash
	SignatureAlgorithm string `json:"signature_algorithm"`
	
	// KeyID is the identifier for the key used to sign the plugin
	KeyID string `json:"key_id"`
	
	// Issuer is the entity that signed the plugin
	Issuer string `json:"issuer"`
	
	// Timestamp is the time when the signature was created
	Timestamp time.Time `json:"timestamp"`
	
	// ExpiresAt is the time when the signature expires
	ExpiresAt time.Time `json:"expires_at"`
	
	// Non-persistent fields for caching
	decodedHash []byte
	decodedSignature []byte
	hashTimestamp int64
}

// SigningKey represents a key used for signing plugins
type SigningKey struct {
	// ID is the identifier for the key
	ID string `json:"id"`
	
	// Name is a human-readable name for the key
	Name string `json:"name"`
	
	// Algorithm is the algorithm used for the key
	Algorithm string `json:"algorithm"`
	
	// PublicKey is the PEM-encoded public key
	PublicKey string `json:"public_key"`
	
	// PrivateKey is the PEM-encoded private key (only available for key generation)
	PrivateKey string `json:"private_key,omitempty"`
	
	// CreatedAt is the time when the key was created
	CreatedAt time.Time `json:"created_at"`
	
	// ExpiresAt is the time when the key expires
	ExpiresAt time.Time `json:"expires_at"`
	
	// IsRevoked indicates whether the key has been revoked
	IsRevoked bool `json:"is_revoked"`
	
	// RevokedAt is the time when the key was revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// LoadSignature loads a signature from a file
func LoadSignature(filePath string) (*PluginSignature, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signature file: %w", err)
	}
	
	var signature PluginSignature
	if err := json.Unmarshal(data, &signature); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signature: %w", err)
	}
	
	return &signature, nil
}

// SaveSignature saves a signature to a file
func (s *PluginSignature) SaveSignature(filePath string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal signature: %w", err)
	}
	
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write signature file: %w", err)
	}
	
	return nil
}

// ComputeHash computes the hash of a file
func ComputeHash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}
	
	return hash.Sum(nil), nil
}

// LoadKey loads a signing key from a file
func LoadKey(filePath string) (*SigningKey, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}
	
	var key SigningKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, fmt.Errorf("failed to unmarshal key: %w", err)
	}
	
	return &key, nil
}

// SaveKey saves a signing key to a file
func (k *SigningKey) SaveKey(filePath string) error {
	data, err := json.MarshalIndent(k, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}
	
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}
	
	return nil
}

// GetPublicKey returns the parsed public key
func (k *SigningKey) GetPublicKey() (interface{}, error) {
	block, _ := pem.Decode([]byte(k.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key PEM")
	}
	
	return x509.ParsePKIXPublicKey(block.Bytes)
}

// GetPrivateKey returns the parsed private key
func (k *SigningKey) GetPrivateKey() (interface{}, error) {
	if k.PrivateKey == "" {
		return nil, fmt.Errorf("private key not available")
	}
	
	block, _ := pem.Decode([]byte(k.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}
	
	return x509.ParsePKCS8PrivateKey(block.Bytes)
}

// Verify verifies a signature against a plugin binary
func (s *PluginSignature) Verify(pluginPath string, key *SigningKey) error {
	// Check if the signature has expired
	if time.Now().After(s.ExpiresAt) {
		return fmt.Errorf("signature has expired")
	}
	
	// Compute the hash of the plugin binary
	hash, err := ComputeHash(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}
	
	// Decode the stored hash
	storedHash, err := base64.StdEncoding.DecodeString(s.Hash)
	if err != nil {
		return fmt.Errorf("failed to decode stored hash: %w", err)
	}
	
	// Compare the computed hash with the stored hash
	if !compareHashes(hash, storedHash) {
		return fmt.Errorf("hash mismatch")
	}
	
	// Get the public key
	publicKey, err := key.GetPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}
	
	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(s.SignatureValue)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}
	
	// Verify the signature
	switch s.SignatureAlgorithm {
	case "RSA-SHA256":
		rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("public key is not an RSA key")
		}
		
		hashed := sha256.Sum256(storedHash)
		err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signature)
		if err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", s.SignatureAlgorithm)
	}
	
	return nil
}

// compareHashes compares two hashes in constant time
func compareHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	
	return result == 0
}