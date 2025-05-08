package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
)

const (
	// DefaultKeysDir is the default directory for storing keys
	DefaultKeysDir = "keys"
	
	// DefaultSignaturesDir is the default directory for storing signatures
	DefaultSignaturesDir = "signatures"
	
	// DefaultKeyValidDays is the default number of days a key is valid for
	DefaultKeyValidDays = 365 * 2 // 2 years
	
	// DefaultSignatureValidDays is the default number of days a signature is valid for
	DefaultSignatureValidDays = 365 // 1 year
	
	// DefaultKeyBits is the default number of bits for RSA keys
	DefaultKeyBits = 2048
)

// SignatureService provides signature verification and signing operations
type SignatureService interface {
	// VerifyPluginSignature verifies a plugin's signature
	VerifyPluginSignature(pluginName, pluginPath string) error
	
	// SignPlugin signs a plugin
	SignPlugin(pluginName, pluginPath string, keyID string) (*PluginSignature, error)
	
	// GenerateSigningKey generates a new signing key
	GenerateSigningKey(name string, validDays int) (*SigningKey, error)
	
	// ListSigningKeys lists all signing keys
	ListSigningKeys() ([]*SigningKey, error)
	
	// GetSigningKey gets a signing key by ID
	GetSigningKey(keyID string) (*SigningKey, error)
	
	// RevokeSigningKey revokes a signing key
	RevokeSigningKey(keyID string) error
	
	// IsTrustedKey checks if a key is trusted
	IsTrustedKey(keyID string) (bool, error)
	
	// AddTrustedKey adds a key to the trusted keys list
	AddTrustedKey(keyID string) error
	
	// RemoveTrustedKey removes a key from the trusted keys list
	RemoveTrustedKey(keyID string) error
}

// SignatureServiceImpl implements the SignatureService interface
type SignatureServiceImpl struct {
	config         *SignatureConfig
	baseDir        string
	keysDir        string
	signaturesDir  string
	logger         hclog.Logger
	keyCache       map[string]*SigningKey
	signatureCache map[string]*PluginSignature
	fileHashCache  map[string]fileHashCacheEntry
	cacheMutex     sync.RWMutex
	hashExpiration time.Duration
}

// fileHashCacheEntry holds cached file hash information
type fileHashCacheEntry struct {
	hash      []byte
	timestamp time.Time
	fileSize  int64
	modTime   time.Time
}

// NewSignatureService creates a new signature service
func NewSignatureService(baseDir string, logger hclog.Logger) (SignatureService, error) {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}
	
	// Set default directories
	if baseDir == "" {
		baseDir = "."
	}
	
	keysDir := filepath.Join(baseDir, DefaultKeysDir)
	signaturesDir := filepath.Join(baseDir, DefaultSignaturesDir)
	
	// Create directories if they don't exist
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create keys directory: %w", err)
	}
	
	if err := os.MkdirAll(signaturesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create signatures directory: %w", err)
	}
	
	// Load or create config
	configPath := filepath.Join(baseDir, "signature_config.json")
	var config *SignatureConfig
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config = &SignatureConfig{
			Enabled:     true,
			TrustedKeys: []string{},
			KeyDir:      keysDir,
		}
		
		// Save config
		configData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write config file: %w", err)
		}
	} else {
		// Load existing config
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		
		config = &SignatureConfig{}
		if err := json.Unmarshal(configData, config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	
	return &SignatureServiceImpl{
		config:         config,
		baseDir:        baseDir,
		keysDir:        keysDir,
		signaturesDir:  signaturesDir,
		logger:         logger,
		keyCache:       make(map[string]*SigningKey),
		signatureCache: make(map[string]*PluginSignature),
		fileHashCache:  make(map[string]fileHashCacheEntry),
		hashExpiration: 30 * time.Minute, // Cache file hashes for 30 minutes by default
	}, nil
}

// VerifyPluginSignature verifies a plugin's signature
func (s *SignatureServiceImpl) VerifyPluginSignature(pluginName, pluginPath string) error {
	if !s.config.Enabled {
		s.logger.Warn("Signature verification is disabled")
		return nil
	}
	
	// Get or load the signature (from cache if available)
	signature, err := s.GetCachedSignature(pluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin signature: %w", err)
	}
	
	// Check if the signature has expired
	if time.Now().After(signature.ExpiresAt) {
		return fmt.Errorf("signature has expired")
	}
	
	// Check if the key is trusted (this is a fast operation)
	trusted, err := s.IsTrustedKey(signature.KeyID)
	if err != nil {
		return fmt.Errorf("failed to check if key is trusted: %w", err)
	}
	
	if !trusted {
		return fmt.Errorf("signature key is not trusted: %s", signature.KeyID)
	}
	
	// Get the signing key (uses internal cache)
	key, err := s.GetSigningKey(signature.KeyID)
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}
	
	// Check if the key is revoked
	if key.IsRevoked {
		return fmt.Errorf("signing key is revoked: %s", key.ID)
	}
	
	// Prepare signature for verification (decode cached components)
	if err := s.PrepareSignatureForVerification(signature); err != nil {
		return fmt.Errorf("failed to prepare signature: %w", err)
	}
	
	// Compute the hash of the plugin binary (using cache if possible)
	hash, err := s.ComputeHashCached(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}
	
	// Compare the computed hash with the stored hash
	if !compareHashes(hash, signature.decodedHash) {
		return fmt.Errorf("hash mismatch")
	}
	
	// Get the public key
	publicKeyInterface, err := key.GetPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}
	
	// Verify the signature
	switch signature.SignatureAlgorithm {
	case "RSA-SHA256":
		rsaPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("public key is not an RSA key")
		}
		
		hashed := sha256.Sum256(signature.decodedHash)
		err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signature.decodedSignature)
		if err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", signature.SignatureAlgorithm)
	}
	
	s.logger.Info("Plugin signature verified successfully", 
		"plugin", pluginName,
		"key_id", signature.KeyID,
		"issuer", signature.Issuer)
	
	return nil
}

// SignPlugin signs a plugin
func (s *SignatureServiceImpl) SignPlugin(pluginName, pluginPath string, keyID string) (*PluginSignature, error) {
	// Get the signing key (uses cache)
	key, err := s.GetSigningKey(keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}
	
	// Check if the key is revoked
	if key.IsRevoked {
		return nil, fmt.Errorf("signing key is revoked: %s", keyID)
	}
	
	// Check if the key has expired
	if time.Now().After(key.ExpiresAt) {
		return nil, fmt.Errorf("signing key has expired: %s", keyID)
	}
	
	// Compute the hash of the plugin binary (with caching)
	hash, err := s.ComputeHashCached(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}
	
	// Get the private key
	privateKeyInterface, err := key.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}
	
	// Sign the hash
	var signatureBytes []byte
	var signatureAlgorithm string
	
	switch privateKey := privateKeyInterface.(type) {
	case *rsa.PrivateKey:
		hashed := sha256.Sum256(hash)
		signatureBytes, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
		if err != nil {
			return nil, fmt.Errorf("failed to sign hash: %w", err)
		}
		signatureAlgorithm = "RSA-SHA256"
	default:
		return nil, fmt.Errorf("unsupported private key type")
	}
	
	// Get the plugin version by extracting it from the filename if possible
	pluginVersion := "unknown"
	parts := strings.Split(filepath.Base(pluginPath), "-")
	if len(parts) > 1 {
		pluginVersion = parts[len(parts)-1]
	}
	
	// Create the signature
	signature := &PluginSignature{
		Version:            "1.0",
		PluginName:         pluginName,
		PluginVersion:      pluginVersion,
		HashAlgorithm:      "SHA-256",
		Hash:               base64.StdEncoding.EncodeToString(hash),
		SignatureValue:     base64.StdEncoding.EncodeToString(signatureBytes),
		SignatureAlgorithm: signatureAlgorithm,
		KeyID:              keyID,
		Issuer:             key.Name,
		Timestamp:          time.Now(),
		ExpiresAt:          time.Now().AddDate(0, 0, DefaultSignatureValidDays),
		// Cache the decoded values immediately
		decodedHash:      hash,
		decodedSignature: signatureBytes,
	}
	
	// Save the signature
	signaturePath := filepath.Join(s.signaturesDir, pluginName+".sig")
	if err := signature.SaveSignature(signaturePath); err != nil {
		return nil, fmt.Errorf("failed to save signature: %w", err)
	}
	
	// Update the signature cache
	s.cacheMutex.Lock()
	s.signatureCache[pluginName] = signature
	s.cacheMutex.Unlock()
	
	s.logger.Info("Plugin signed successfully", 
		"plugin", pluginName,
		"key_id", keyID,
		"signature_path", signaturePath)
	
	return signature, nil
}

// GenerateSigningKey generates a new signing key
func (s *SignatureServiceImpl) GenerateSigningKey(name string, validDays int) (*SigningKey, error) {
	if validDays <= 0 {
		validDays = DefaultKeyValidDays
	}
	
	// Generate a key ID
	keyID := uuid.New().String()
	
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, DefaultKeyBits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}
	
	// Marshal the private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	
	// Marshal the public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	
	// PEM encode the keys
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	
	// Create the signing key
	key := &SigningKey{
		ID:         keyID,
		Name:       name,
		Algorithm:  "RSA",
		PublicKey:  string(publicKeyPEM),
		PrivateKey: string(privateKeyPEM),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().AddDate(0, 0, validDays),
		IsRevoked:  false,
	}
	
	// Save the key
	keyPath := filepath.Join(s.keysDir, keyID+".json")
	if err := key.SaveKey(keyPath); err != nil {
		return nil, fmt.Errorf("failed to save key: %w", err)
	}
	
	// Save the public key separately
	publicKeyPath := filepath.Join(s.keysDir, keyID+".pub")
	if err := os.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
		return nil, fmt.Errorf("failed to save public key: %w", err)
	}
	
	// Add the key to the cache
	s.cacheMutex.Lock()
	s.keyCache[keyID] = key
	s.cacheMutex.Unlock()
	
	s.logger.Info("Generated new signing key", 
		"key_id", keyID,
		"name", name,
		"expires", key.ExpiresAt)
	
	return key, nil
}

// ListSigningKeys lists all signing keys
func (s *SignatureServiceImpl) ListSigningKeys() ([]*SigningKey, error) {
	var keys []*SigningKey
	
	// Get all key files
	matches, err := filepath.Glob(filepath.Join(s.keysDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob key files: %w", err)
	}
	
	for _, match := range matches {
		// Extract key ID from filename
		keyID := strings.TrimSuffix(filepath.Base(match), ".json")
		
		// Get the key
		key, err := s.GetSigningKey(keyID)
		if err != nil {
			s.logger.Warn("Failed to load key", "key_id", keyID, "error", err)
			continue
		}
		
		keys = append(keys, key)
	}
	
	return keys, nil
}

// GetSigningKey gets a signing key by ID
func (s *SignatureServiceImpl) GetSigningKey(keyID string) (*SigningKey, error) {
	// Check the cache first
	s.cacheMutex.RLock()
	key, ok := s.keyCache[keyID]
	s.cacheMutex.RUnlock()
	
	if ok {
		return key, nil
	}
	
	// Load the key from file
	keyPath := filepath.Join(s.keysDir, keyID+".json")
	key, err := LoadKey(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load key: %w", err)
	}
	
	// Update the cache
	s.cacheMutex.Lock()
	s.keyCache[keyID] = key
	s.cacheMutex.Unlock()
	
	return key, nil
}

// RevokeSigningKey revokes a signing key
func (s *SignatureServiceImpl) RevokeSigningKey(keyID string) error {
	// Get the key
	key, err := s.GetSigningKey(keyID)
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}
	
	// Revoke the key
	key.IsRevoked = true
	now := time.Now()
	key.RevokedAt = &now
	
	// Save the key
	keyPath := filepath.Join(s.keysDir, keyID+".json")
	if err := key.SaveKey(keyPath); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}
	
	// Update the cache
	s.cacheMutex.Lock()
	s.keyCache[keyID] = key
	s.cacheMutex.Unlock()
	
	// Remove from trusted keys if present
	if s.config.Enabled {
		if err := s.RemoveTrustedKey(keyID); err != nil {
			s.logger.Warn("Failed to remove key from trusted keys", "key_id", keyID, "error", err)
		}
	}
	
	s.logger.Info("Revoked signing key", "key_id", keyID)
	
	return nil
}

// IsTrustedKey checks if a key is trusted
func (s *SignatureServiceImpl) IsTrustedKey(keyID string) (bool, error) {
	if !s.config.Enabled {
		return true, nil
	}
	
	for _, trustedKeyID := range s.config.TrustedKeys {
		if trustedKeyID == keyID {
			return true, nil
		}
	}
	
	return false, nil
}

// AddTrustedKey adds a key to the trusted keys list
func (s *SignatureServiceImpl) AddTrustedKey(keyID string) error {
	// Check if the key exists
	if _, err := s.GetSigningKey(keyID); err != nil {
		return fmt.Errorf("key does not exist: %w", err)
	}
	
	// Check if the key is already trusted
	trusted, err := s.IsTrustedKey(keyID)
	if err != nil {
		return fmt.Errorf("failed to check if key is trusted: %w", err)
	}
	
	if trusted {
		return nil
	}
	
	// Add the key to the trusted keys list
	s.config.TrustedKeys = append(s.config.TrustedKeys, keyID)
	
	// Save the config
	configPath := filepath.Join(s.baseDir, "signature_config.json")
	configData, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	s.logger.Info("Added key to trusted keys", "key_id", keyID)
	
	return nil
}

// RemoveTrustedKey removes a key from the trusted keys list
func (s *SignatureServiceImpl) RemoveTrustedKey(keyID string) error {
	// Check if the key is trusted
	trusted, err := s.IsTrustedKey(keyID)
	if err != nil {
		return fmt.Errorf("failed to check if key is trusted: %w", err)
	}
	
	if !trusted {
		return nil
	}
	
	// Remove the key from the trusted keys list
	var newTrustedKeys []string
	for _, trustedKeyID := range s.config.TrustedKeys {
		if trustedKeyID != keyID {
			newTrustedKeys = append(newTrustedKeys, trustedKeyID)
		}
	}
	
	s.config.TrustedKeys = newTrustedKeys
	
	// Save the config
	configPath := filepath.Join(s.baseDir, "signature_config.json")
	configData, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	s.logger.Info("Removed key from trusted keys", "key_id", keyID)
	
	return nil
}