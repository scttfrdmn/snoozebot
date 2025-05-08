package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/signature"
)

const (
	defaultSignatureDir = "/etc/snoozebot/signatures"
)

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "snoozesign",
		Level:      hclog.Info,
		Output:     os.Stderr,
		JSONFormat: false,
	})

	// Parse command line flags
	var (
		generateKey   = flag.Bool("generate-key", false, "Generate a new signing key")
		keyName       = flag.String("key-name", "", "Name for the new signing key")
		validDays     = flag.Int("valid-days", signature.DefaultKeyValidDays, "Number of days the key is valid for")
		signPlugin    = flag.Bool("sign", false, "Sign a plugin")
		pluginPath    = flag.String("plugin", "", "Path to the plugin to sign")
		keyID         = flag.String("key-id", "", "ID of the key to use for signing")
		verifyPlugin  = flag.Bool("verify", false, "Verify a plugin signature")
		trustKey      = flag.Bool("trust-key", false, "Add a key to the trusted keys list")
		revokeKey     = flag.Bool("revoke-key", false, "Revoke a signing key")
		listKeys      = flag.Bool("list-keys", false, "List all signing keys")
		signatureDir  = flag.String("signature-dir", defaultSignatureDir, "Directory for signatures")
		bundlePlugin  = flag.Bool("bundle", false, "Create a signed plugin bundle")
		bundlePath    = flag.String("bundle-path", "", "Path to the bundle output")
	)

	flag.Parse()

	// Create signature service
	sigService, err := signature.NewSignatureService(*signatureDir, logger)
	if err != nil {
		logger.Error("Failed to create signature service", "error", err)
		os.Exit(1)
	}

	// Create plugin signer
	pluginSigner := signature.NewPluginSigner(sigService, logger)

	// Process commands
	switch {
	case *generateKey:
		if *keyName == "" {
			logger.Error("Key name is required")
			os.Exit(1)
		}

		key, err := sigService.GenerateSigningKey(*keyName, *validDays)
		if err != nil {
			logger.Error("Failed to generate signing key", "error", err)
			os.Exit(1)
		}

		// Add the key to the trusted keys list
		if err := sigService.AddTrustedKey(key.ID); err != nil {
			logger.Error("Failed to add key to trusted keys", "error", err)
			os.Exit(1)
		}

		fmt.Printf("Generated new signing key:\n")
		fmt.Printf("  ID:        %s\n", key.ID)
		fmt.Printf("  Name:      %s\n", key.Name)
		fmt.Printf("  Algorithm: %s\n", key.Algorithm)
		fmt.Printf("  Created:   %s\n", key.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Expires:   %s\n", key.ExpiresAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Status:    %s\n", "trusted")

	case *signPlugin:
		if *pluginPath == "" {
			logger.Error("Plugin path is required")
			os.Exit(1)
		}

		if *keyID == "" {
			logger.Error("Key ID is required")
			os.Exit(1)
		}

		if err := pluginSigner.SignPlugin(*pluginPath, *keyID); err != nil {
			logger.Error("Failed to sign plugin", "error", err)
			os.Exit(1)
		}

		fmt.Printf("Plugin signed successfully: %s\n", *pluginPath)

	case *verifyPlugin:
		if *pluginPath == "" {
			logger.Error("Plugin path is required")
			os.Exit(1)
		}

		// Get the plugin name
		pluginName := filepath.Base(*pluginPath)

		if err := sigService.VerifyPluginSignature(pluginName, *pluginPath); err != nil {
			logger.Error("Signature verification failed", "error", err)
			os.Exit(1)
		}

		fmt.Printf("Plugin signature verified successfully: %s\n", *pluginPath)

	case *trustKey:
		if *keyID == "" {
			logger.Error("Key ID is required")
			os.Exit(1)
		}

		if err := sigService.AddTrustedKey(*keyID); err != nil {
			logger.Error("Failed to add key to trusted keys", "error", err)
			os.Exit(1)
		}

		fmt.Printf("Key added to trusted keys: %s\n", *keyID)

	case *revokeKey:
		if *keyID == "" {
			logger.Error("Key ID is required")
			os.Exit(1)
		}

		if err := sigService.RevokeSigningKey(*keyID); err != nil {
			logger.Error("Failed to revoke key", "error", err)
			os.Exit(1)
		}

		fmt.Printf("Key revoked: %s\n", *keyID)

	case *listKeys:
		keys, err := sigService.ListSigningKeys()
		if err != nil {
			logger.Error("Failed to list keys", "error", err)
			os.Exit(1)
		}

		if len(keys) == 0 {
			fmt.Println("No signing keys found")
			return
		}

		fmt.Println("Signing keys:")
		for _, key := range keys {
			trusted, err := sigService.IsTrustedKey(key.ID)
			if err != nil {
				logger.Error("Failed to check if key is trusted", "key_id", key.ID, "error", err)
				continue
			}

			status := "valid"
			if key.IsRevoked {
				status = "revoked"
			} else if trusted {
				status = "trusted"
			}

			fmt.Printf("  ID:        %s\n", key.ID)
			fmt.Printf("  Name:      %s\n", key.Name)
			fmt.Printf("  Algorithm: %s\n", key.Algorithm)
			fmt.Printf("  Created:   %s\n", key.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Expires:   %s\n", key.ExpiresAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Status:    %s\n", status)
			fmt.Println()
		}

	case *bundlePlugin:
		if *pluginPath == "" {
			logger.Error("Plugin path is required")
			os.Exit(1)
		}

		if *bundlePath == "" {
			logger.Error("Bundle path is required")
			os.Exit(1)
		}

		if *keyID == "" {
			logger.Error("Key ID is required")
			os.Exit(1)
		}

		if err := pluginSigner.CreateAndSignPluginBundle(*pluginPath, *bundlePath, *keyID); err != nil {
			logger.Error("Failed to create signed plugin bundle", "error", err)
			os.Exit(1)
		}

		fmt.Printf("Plugin bundle created and signed: %s\n", *bundlePath)

	default:
		flag.Usage()
	}
}