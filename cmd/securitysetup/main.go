package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/auth"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/signature"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/tls"
)

// SecurityWizard is a setup wizard for Snoozebot security features
type SecurityWizard struct {
	baseDir        string
	configDir      string
	certsDir       string
	signaturesDir  string
	keysDir        string
	logger         hclog.Logger
	reader         *bufio.Reader
	writer         io.Writer
	forceOverwrite bool
}

// NewSecurityWizard creates a new security wizard
func NewSecurityWizard(baseDir string, logger hclog.Logger, reader io.Reader, writer io.Writer, forceOverwrite bool) *SecurityWizard {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	if baseDir == "" {
		baseDir = "/etc/snoozebot"
	}

	return &SecurityWizard{
		baseDir:        baseDir,
		configDir:      filepath.Join(baseDir, "config"),
		certsDir:       filepath.Join(baseDir, "certs"),
		signaturesDir:  filepath.Join(baseDir, "signatures"),
		keysDir:        filepath.Join(baseDir, "signatures", "keys"),
		logger:         logger,
		reader:         bufio.NewReader(reader),
		writer:         writer,
		forceOverwrite: forceOverwrite,
	}
}

// Run runs the security wizard
func (w *SecurityWizard) Run() error {
	w.printWelcome()

	// Ensure directories exist
	if err := w.ensureDirectories(); err != nil {
		return err
	}

	// Configure TLS
	if err := w.setupTLS(); err != nil {
		return err
	}

	// Configure signatures
	if err := w.setupSignatures(); err != nil {
		return err
	}

	// Configure authentication
	if err := w.setupAuthentication(); err != nil {
		return err
	}

	// Generate environment configuration
	if err := w.generateEnvironmentConfig(); err != nil {
		return err
	}

	// Show summary
	w.showSummary()

	return nil
}

// printWelcome prints the welcome message
func (w *SecurityWizard) printWelcome() {
	w.println("=================================")
	w.println("Snoozebot Security Setup Wizard")
	w.println("=================================")
	w.println("")
	w.println("This wizard will help you set up security features for Snoozebot:")
	w.println("1. TLS encryption for secure plugin communication")
	w.println("2. Plugin signature verification for integrity checks")
	w.println("3. API key authentication for access control")
	w.println("")
	w.println("Press Enter to continue...")
	w.readLine()
}

// ensureDirectories ensures that the necessary directories exist
func (w *SecurityWizard) ensureDirectories() error {
	dirs := []string{
		w.baseDir,
		w.configDir,
		w.certsDir,
		w.signaturesDir,
		w.keysDir,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			w.printf("Creating directory: %s\n", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}

	return nil
}

// setupTLS sets up TLS for plugin communication
func (w *SecurityWizard) setupTLS() error {
	w.println("\n=== TLS Configuration ===")
	w.println("TLS ensures secure encrypted communication between the main application and plugins.")

	enable, err := w.promptYesNo("Enable TLS encryption?", true)
	if err != nil {
		return err
	}

	if !enable {
		w.println("TLS encryption will not be enabled.")
		return nil
	}

	w.println("Generating TLS certificates...")

	// Create TLS manager
	tlsManager, err := tls.NewTLSManager(w.certsDir)
	if err != nil {
		return fmt.Errorf("failed to create TLS manager: %w", err)
	}

	// Initialize TLS
	if err := tlsManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize TLS: %w", err)
	}

	// Generate CA certificate
	ca, err := tlsManager.GetCA()
	if err != nil {
		return fmt.Errorf("failed to get CA: %w", err)
	}

	w.printf("CA certificate generated at: %s\n", ca.CertFile)
	w.printf("CA private key generated at: %s\n", ca.KeyFile)

	// Generate client and server certificates
	_, _, err = tlsManager.EnsurePluginCertificate("client")
	if err != nil {
		return fmt.Errorf("failed to generate client certificate: %w", err)
	}

	_, _, err = tlsManager.EnsurePluginCertificate("server")
	if err != nil {
		return fmt.Errorf("failed to generate server certificate: %w", err)
	}

	w.println("TLS certificates generated successfully.")

	return nil
}

// setupSignatures sets up plugin signature verification
func (w *SecurityWizard) setupSignatures() error {
	w.println("\n=== Signature Verification Configuration ===")
	w.println("Signature verification ensures that plugins are authentic and have not been tampered with.")

	enable, err := w.promptYesNo("Enable plugin signature verification?", true)
	if err != nil {
		return err
	}

	if !enable {
		w.println("Plugin signature verification will not be enabled.")
		return nil
	}

	// Create signature service
	sigService, err := signature.NewSignatureService(w.signaturesDir, w.logger)
	if err != nil {
		return fmt.Errorf("failed to create signature service: %w", err)
	}

	// Generate signing key
	keyName, err := w.promptString("Enter a name for the signing key", "snoozebot-key")
	if err != nil {
		return err
	}

	w.println("Generating signing key...")
	key, err := sigService.GenerateSigningKey(keyName, signature.DefaultKeyValidDays)
	if err != nil {
		return fmt.Errorf("failed to generate signing key: %w", err)
	}

	// Add the key to the trusted keys list
	if err := sigService.AddTrustedKey(key.ID); err != nil {
		return fmt.Errorf("failed to add key to trusted keys: %w", err)
	}

	w.printf("Signing key generated with ID: %s\n", key.ID)
	w.printf("Key will expire on: %s\n", key.ExpiresAt.Format("2006-01-02"))
	w.println("Key has been added to the trusted keys list.")

	return nil
}

// setupAuthentication sets up API key authentication
func (w *SecurityWizard) setupAuthentication() error {
	w.println("\n=== Authentication Configuration ===")
	w.println("Authentication ensures that only authorized clients can use the plugins.")

	enable, err := w.promptYesNo("Enable API key authentication?", true)
	if err != nil {
		return err
	}

	if !enable {
		w.println("API key authentication will not be enabled.")
		return nil
	}

	// Ask for roles
	roles := []string{"admin"}
	customRoles, err := w.promptYesNo("Would you like to add custom roles?", false)
	if err != nil {
		return err
	}

	if customRoles {
		role, err := w.promptString("Enter role name (or empty to finish)", "")
		if err != nil {
			return err
		}

		for role != "" {
			roles = append(roles, role)
			role, err = w.promptString("Enter role name (or empty to finish)", "")
			if err != nil {
				return err
			}
		}
	}

	// Generate API key
	w.println("Generating API key...")
	apiKey, err := auth.GenerateAPIKey("admin", roles)
	if err != nil {
		return fmt.Errorf("failed to generate API key: %w", err)
	}

	// Save authentication config
	authConfig := auth.PluginAuthConfig{
		APIKeys: []auth.APIKey{*apiKey},
	}

	authConfigPath := filepath.Join(w.configDir, "auth.json")
	authData, err := json.MarshalIndent(authConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %w", err)
	}

	if err := os.WriteFile(authConfigPath, authData, 0600); err != nil {
		return fmt.Errorf("failed to write auth config: %w", err)
	}

	w.printf("API key generated: %s\n", apiKey.Key)
	w.printf("Key has roles: %s\n", strings.Join(apiKey.Roles, ", "))
	w.printf("Authentication configuration saved to: %s\n", authConfigPath)

	return nil
}

// generateEnvironmentConfig generates an environment configuration file
func (w *SecurityWizard) generateEnvironmentConfig() error {
	w.println("\n=== Environment Configuration ===")
	w.println("Generating environment configuration file...")

	envFilePath := filepath.Join(w.baseDir, "security.env")
	
	// Check if file exists
	if _, err := os.Stat(envFilePath); err == nil && !w.forceOverwrite {
		overwrite, err := w.promptYesNo(fmt.Sprintf("File %s already exists. Overwrite?", envFilePath), false)
		if err != nil {
			return err
		}
		
		if !overwrite {
			w.println("Skipping environment file generation.")
			return nil
		}
	}

	// Create environment file
	file, err := os.Create(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to create environment file: %w", err)
	}
	defer file.Close()

	// Write environment variables
	lines := []string{
		"# Snoozebot Security Configuration",
		"# Generated on " + time.Now().Format("2006-01-02 15:04:05"),
		"",
		"# Base configuration directory",
		"SNOOZEBOT_CONFIG_DIR=" + w.baseDir,
		"",
		"# TLS Configuration",
		"SNOOZEBOT_TLS_ENABLED=true",
		"SNOOZEBOT_TLS_CERT_DIR=" + w.certsDir,
		"# SNOOZEBOT_TLS_SKIP_VERIFY=false  # Uncomment to disable certificate verification (not recommended)",
		"",
		"# Signature Verification Configuration",
		"SNOOZEBOT_SIGNATURE_ENABLED=true",
		"SNOOZEBOT_SIGNATURE_DIR=" + w.signaturesDir,
		"",
		"# Authentication Configuration",
		"SNOOZEBOT_AUTH_ENABLED=true",
		"SNOOZEBOT_AUTH_CONFIG=" + filepath.Join(w.configDir, "auth.json"),
		"# SNOOZEBOT_API_KEY=your-api-key  # Uncomment and replace with your API key",
	}

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to environment file: %w", err)
		}
	}

	w.printf("Environment configuration saved to: %s\n", envFilePath)
	return nil
}

// showSummary shows a summary of the configured security features
func (w *SecurityWizard) showSummary() {
	w.println("\n=== Setup Complete ===")
	w.println("Snoozebot security features have been configured successfully.")
	w.println("")
	w.println("To use the security features, source the environment file:")
	w.printf("  source %s/security.env\n", w.baseDir)
	w.println("")
	w.println("Or export the environment variables in your application.")
	w.println("")
	w.println("For more information, see the documentation:")
	w.println("  - TLS: docs/PLUGIN_TLS.md")
	w.println("  - Signatures: docs/PLUGIN_SIGNATURES.md")
	w.println("  - Authentication: docs/PLUGIN_AUTHENTICATION.md")
	w.println("")
}

// promptYesNo prompts for a yes/no answer
func (w *SecurityWizard) promptYesNo(prompt string, defaultValue bool) (bool, error) {
	defaultStr := "Y/n"
	if !defaultValue {
		defaultStr = "y/N"
	}

	w.printf("%s [%s]: ", prompt, defaultStr)
	input, err := w.readLine()
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultValue, nil
	}

	if input == "y" || input == "yes" {
		return true, nil
	}

	if input == "n" || input == "no" {
		return false, nil
	}

	w.println("Please enter 'y' or 'n'")
	return w.promptYesNo(prompt, defaultValue)
}

// promptString prompts for a string
func (w *SecurityWizard) promptString(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		w.printf("%s [%s]: ", prompt, defaultValue)
	} else {
		w.printf("%s: ", prompt)
	}

	input, err := w.readLine()
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}

	return input, nil
}

// println prints a line to the writer
func (w *SecurityWizard) println(a ...interface{}) {
	fmt.Fprintln(w.writer, a...)
}

// printf prints a formatted string to the writer
func (w *SecurityWizard) printf(format string, a ...interface{}) {
	fmt.Fprintf(w.writer, format, a...)
}

// readLine reads a line from the reader
func (w *SecurityWizard) readLine() (string, error) {
	return w.reader.ReadString('\n')
}

func main() {
	// Parse command-line flags
	baseDir := os.Getenv("SNOOZEBOT_CONFIG_DIR")
	if baseDir == "" {
		baseDir = "/etc/snoozebot"
	}

	// Check for force flag
	forceOverwrite := false
	for _, arg := range os.Args {
		if arg == "-f" || arg == "--force" {
			forceOverwrite = true
			break
		}
	}

	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "security-wizard",
		Level:  hclog.Info,
		Output: os.Stderr,
	})

	// Create and run the wizard
	wizard := NewSecurityWizard(baseDir, logger, os.Stdin, os.Stdout, forceOverwrite)
	if err := wizard.Run(); err != nil {
		logger.Error("Failed to run security wizard", "error", err)
		os.Exit(1)
	}
}