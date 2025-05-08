package signature

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ComputeHashCached computes the hash of a file with caching
func (s *SignatureServiceImpl) ComputeHashCached(filePath string) ([]byte, error) {
	// Get file info to check if it changed
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	
	fileSize := fileInfo.Size()
	modTime := fileInfo.ModTime()
	
	// Check the cache first with read lock
	s.cacheMutex.RLock()
	cacheEntry, found := s.fileHashCache[filePath]
	s.cacheMutex.RUnlock()
	
	// If found in cache and file hasn't changed, return the cached hash
	if found && 
	   fileSize == cacheEntry.fileSize && 
	   modTime.Equal(cacheEntry.modTime) && 
	   time.Since(cacheEntry.timestamp) < s.hashExpiration {
		s.logger.Debug("Using cached file hash", "file", filePath)
		return cacheEntry.hash, nil
	}
	
	// Compute the hash
	s.logger.Debug("Computing file hash", "file", filePath)
	
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Compute the hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}
	
	hash := hasher.Sum(nil)
	
	// Update the cache with write lock
	s.cacheMutex.Lock()
	s.fileHashCache[filePath] = fileHashCacheEntry{
		hash:      hash,
		timestamp: time.Now(),
		fileSize:  fileSize,
		modTime:   modTime,
	}
	s.cacheMutex.Unlock()
	
	return hash, nil
}

// GetCachedSignature retrieves a cached signature or loads it from disk
func (s *SignatureServiceImpl) GetCachedSignature(pluginName string) (*PluginSignature, error) {
	// Check the cache first with read lock
	s.cacheMutex.RLock()
	signature, found := s.signatureCache[pluginName]
	s.cacheMutex.RUnlock()
	
	if found {
		s.logger.Debug("Using cached signature", "plugin", pluginName)
		return signature, nil
	}
	
	// Load the signature from disk
	s.logger.Debug("Loading signature from disk", "plugin", pluginName)
	signaturePath := filepath.Join(s.signaturesDir, pluginName+".sig")
	
	signature, err := LoadSignature(signaturePath)
	if err != nil {
		return nil, err
	}
	
	// Cache the signature
	s.cacheMutex.Lock()
	s.signatureCache[pluginName] = signature
	s.cacheMutex.Unlock()
	
	return signature, nil
}

// PrepareSignatureForVerification decodes and caches signature components
func (s *SignatureServiceImpl) PrepareSignatureForVerification(signature *PluginSignature) error {
	// If hash is already decoded, return
	if signature.decodedHash != nil && signature.decodedSignature != nil {
		return nil
	}
	
	// Decode the hash if needed
	if signature.decodedHash == nil {
		hash, err := base64.StdEncoding.DecodeString(signature.Hash)
		if err != nil {
			return fmt.Errorf("failed to decode stored hash: %w", err)
		}
		signature.decodedHash = hash
	}
	
	// Decode the signature if needed
	if signature.decodedSignature == nil {
		sig, err := base64.StdEncoding.DecodeString(signature.SignatureValue)
		if err != nil {
			return fmt.Errorf("failed to decode signature: %w", err)
		}
		signature.decodedSignature = sig
	}
	
	return nil
}