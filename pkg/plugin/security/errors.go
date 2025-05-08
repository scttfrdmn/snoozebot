package security

import (
	"errors"
	"fmt"
	"strings"
)

// Common error codes for security-related issues
const (
	// TLS error codes
	ErrTLSInitializationFailed     = "TLS_INIT_FAILED"
	ErrTLSCertificateNotFound      = "TLS_CERT_NOT_FOUND"
	ErrTLSCertificateInvalid       = "TLS_CERT_INVALID"
	ErrTLSCertificateExpired       = "TLS_CERT_EXPIRED"
	ErrTLSHandshakeFailed          = "TLS_HANDSHAKE_FAILED"
	ErrTLSVerificationFailed       = "TLS_VERIFICATION_FAILED"
	ErrTLSCertificateAuthority     = "TLS_CA_ERROR"
	ErrTLSGenerationFailed         = "TLS_GENERATION_FAILED"
	
	// Signature error codes
	ErrSignatureNotFound           = "SIG_NOT_FOUND"
	ErrSignatureInvalid            = "SIG_INVALID"
	ErrSignatureExpired            = "SIG_EXPIRED"
	ErrSignatureKeyNotFound        = "SIG_KEY_NOT_FOUND"
	ErrSignatureKeyInvalid         = "SIG_KEY_INVALID"
	ErrSignatureKeyExpired         = "SIG_KEY_EXPIRED"
	ErrSignatureKeyRevoked         = "SIG_KEY_REVOKED"
	ErrSignatureGenerationFailed   = "SIG_GENERATION_FAILED"
	ErrSignatureHashMismatch       = "SIG_HASH_MISMATCH"
	ErrSignatureVerificationFailed = "SIG_VERIFICATION_FAILED"
	
	// Authentication error codes
	ErrAuthInitializationFailed    = "AUTH_INIT_FAILED"
	ErrAuthConfigNotFound          = "AUTH_CONFIG_NOT_FOUND"
	ErrAuthConfigInvalid           = "AUTH_CONFIG_INVALID"
	ErrAuthAPIKeyInvalid           = "AUTH_API_KEY_INVALID"
	ErrAuthAPIKeyExpired           = "AUTH_API_KEY_EXPIRED"
	ErrAuthAPIKeyRevoked           = "AUTH_API_KEY_REVOKED"
	ErrAuthPermissionDenied        = "AUTH_PERMISSION_DENIED"
	ErrAuthRoleInvalid             = "AUTH_ROLE_INVALID"
	
	// Plugin error codes
	ErrPluginNotFound              = "PLUGIN_NOT_FOUND"
	ErrPluginInvalid               = "PLUGIN_INVALID"
	ErrPluginCommunicationFailed   = "PLUGIN_COMM_FAILED"
	ErrPluginSecurityDisabled      = "PLUGIN_SECURITY_DISABLED"
)

// SecurityError represents a security-related error with a code and details
type SecurityError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]string
}

// Error implements the error interface
func (e *SecurityError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
	
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}
	
	if len(e.Context) > 0 {
		contextParts := make([]string, 0, len(e.Context))
		for k, v := range e.Context {
			contextParts = append(contextParts, fmt.Sprintf("%s=%s", k, v))
		}
		msg += " (" + strings.Join(contextParts, ", ") + ")"
	}
	
	return msg
}

// Unwrap returns the underlying cause of the error
func (e *SecurityError) Unwrap() error {
	return e.Cause
}

// Is checks if the target error is of the same type as this error
func (e *SecurityError) Is(target error) bool {
	t, ok := target.(*SecurityError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewError creates a new security error
func NewError(code, message string) *SecurityError {
	return &SecurityError{
		Code:    code,
		Message: message,
		Context: make(map[string]string),
	}
}

// WithCause adds a cause to the error
func (e *SecurityError) WithCause(cause error) *SecurityError {
	e.Cause = cause
	return e
}

// WithContext adds context to the error
func (e *SecurityError) WithContext(key, value string) *SecurityError {
	e.Context[key] = value
	return e
}

// IsSecurityError checks if an error is a security error
func IsSecurityError(err error) bool {
	var secErr *SecurityError
	return errors.As(err, &secErr)
}

// GetSecurityErrorCode gets the code from a security error
func GetSecurityErrorCode(err error) (string, bool) {
	var secErr *SecurityError
	if errors.As(err, &secErr) {
		return secErr.Code, true
	}
	return "", false
}

// GetUserFriendlyMessage returns a user-friendly error message
func GetUserFriendlyMessage(err error) string {
	var secErr *SecurityError
	if !errors.As(err, &secErr) {
		return "An error occurred"
	}
	
	switch secErr.Code {
	// TLS errors
	case ErrTLSInitializationFailed:
		return "Failed to initialize TLS security. Check certificate directory permissions and try again."
	case ErrTLSCertificateNotFound:
		return "TLS certificate not found. Run the security setup wizard to generate certificates."
	case ErrTLSCertificateInvalid:
		return "Invalid TLS certificate. The certificate may be corrupted or in the wrong format."
	case ErrTLSCertificateExpired:
		return "Expired TLS certificate. Regenerate certificates using the security setup wizard."
	case ErrTLSHandshakeFailed:
		return "Failed to establish secure connection. Check that TLS is configured correctly on both sides."
	case ErrTLSVerificationFailed:
		return "TLS certificate verification failed. Ensure the certificate is signed by a trusted CA."
	case ErrTLSCertificateAuthority:
		return "CA certificate error. The Certificate Authority may be invalid or misconfigured."
	case ErrTLSGenerationFailed:
		return "Failed to generate TLS certificates. Check permissions and disk space."
	
	// Signature errors
	case ErrSignatureNotFound:
		return "Plugin signature not found. Sign the plugin using the 'snoozesign' utility."
	case ErrSignatureInvalid:
		return "Invalid plugin signature. The signature may be corrupted or tampered with."
	case ErrSignatureExpired:
		return "Expired plugin signature. Re-sign the plugin using the 'snoozesign' utility."
	case ErrSignatureKeyNotFound:
		return "Signing key not found. Generate a new key using the 'snoozesign' utility."
	case ErrSignatureKeyInvalid:
		return "Invalid signing key. The key may be corrupted or in the wrong format."
	case ErrSignatureKeyExpired:
		return "Expired signing key. Generate a new key using the 'snoozesign' utility."
	case ErrSignatureKeyRevoked:
		return "Revoked signing key. The key has been revoked and can no longer be used."
	case ErrSignatureGenerationFailed:
		return "Failed to generate signature. Check permissions and key availability."
	case ErrSignatureHashMismatch:
		return "Plugin binary has been modified. The hash does not match the signature."
	case ErrSignatureVerificationFailed:
		return "Signature verification failed. The plugin may be tampered with or corrupt."
	
	// Authentication errors
	case ErrAuthInitializationFailed:
		return "Failed to initialize authentication. Check configuration and permissions."
	case ErrAuthConfigNotFound:
		return "Authentication configuration not found. Run the security setup wizard to configure authentication."
	case ErrAuthConfigInvalid:
		return "Invalid authentication configuration. The configuration may be corrupted."
	case ErrAuthAPIKeyInvalid:
		return "Invalid API key. Check the key format and ensure it has been properly generated."
	case ErrAuthAPIKeyExpired:
		return "Expired API key. Generate a new key using the security setup wizard."
	case ErrAuthAPIKeyRevoked:
		return "Revoked API key. The key has been revoked and can no longer be used."
	case ErrAuthPermissionDenied:
		return "Permission denied. The API key does not have the required permissions."
	case ErrAuthRoleInvalid:
		return "Invalid role. The specified role does not exist or has been removed."
	
	// Plugin errors
	case ErrPluginNotFound:
		return "Plugin not found. Check the plugin path and ensure it has been built."
	case ErrPluginInvalid:
		return "Invalid plugin. The plugin may be corrupted or built for a different platform."
	case ErrPluginCommunicationFailed:
		return "Failed to communicate with plugin. Check that the plugin process is running."
	case ErrPluginSecurityDisabled:
		return "Plugin security is disabled. Enable security features for better protection."
	
	default:
		return fmt.Sprintf("Security error: %s", secErr.Message)
	}
}

// GetDetailedHelp returns detailed help for an error
func GetDetailedHelp(err error) string {
	var secErr *SecurityError
	if !errors.As(err, &secErr) {
		return "No detailed help available for this error"
	}
	
	switch secErr.Code {
	// TLS errors
	case ErrTLSInitializationFailed:
		return `
TLS initialization failed. This may be due to:
1. Missing or invalid certificate directory
2. Insufficient permissions to read/write certificates
3. Invalid CA certificate

Resolution steps:
- Run the security setup wizard to regenerate certificates
- Check directory permissions (should be readable by the application)
- Ensure the configuration points to the correct certificate directory
`
	case ErrTLSCertificateNotFound:
		return `
TLS certificate not found. This may be due to:
1. Certificates never generated
2. Incorrect certificate path
3. Deleted or moved certificates

Resolution steps:
- Run the security setup wizard to generate certificates
- Check that SNOOZEBOT_TLS_CERT_DIR points to the correct directory
- Ensure certificate files exist and have correct permissions
`
	case ErrTLSHandshakeFailed:
		return `
TLS handshake failed. This may be due to:
1. TLS version mismatch
2. Cipher suite incompatibility
3. Certificate not trusted by the client or server
4. Certificate hostname verification failure

Resolution steps:
- Ensure both client and server are using compatible TLS versions
- Check that certificates are valid and trusted
- Verify hostname configuration matches the certificate
- For testing, you can set SNOOZEBOT_TLS_SKIP_VERIFY=true (not for production)
`
	
	// Signature errors
	case ErrSignatureNotFound:
		return `
Plugin signature not found. This may be due to:
1. Plugin has never been signed
2. Signature file was deleted or moved
3. Incorrect signature directory

Resolution steps:
- Sign the plugin using: ./bin/snoozesign -sign -plugin=./bin/plugins/<plugin> -key-id=<key-id>
- Check that SNOOZEBOT_SIGNATURE_DIR points to the correct directory
- Ensure the plugin name matches the signature file name
`
	case ErrSignatureVerificationFailed:
		return `
Signature verification failed. This may be due to:
1. Plugin binary modified after signing
2. Incorrect or tampered signature
3. Signing key not trusted

Resolution steps:
- Re-sign the plugin with a trusted key
- Verify the key ID is in the trusted keys list
- Check the plugin binary integrity
- Use ./bin/snoozesign -verify -plugin=./bin/plugins/<plugin> for detailed information
`
	
	// Authentication errors
	case ErrAuthAPIKeyInvalid:
		return `
Invalid API key. This may be due to:
1. Malformed API key
2. API key not in the authentication configuration
3. Incorrect authentication configuration

Resolution steps:
- Check the API key format
- Verify the API key is included in the authentication configuration
- Run the security setup wizard to generate a new API key
- Ensure SNOOZEBOT_AUTH_CONFIG points to the correct file
`
	case ErrAuthPermissionDenied:
		return `
Permission denied. This may be due to:
1. API key does not have the required role
2. Required permission is not assigned to the role
3. Authentication configuration mismatch

Resolution steps:
- Check the API key roles in the authentication configuration
- Verify the required permissions are assigned to the roles
- Use an API key with the appropriate roles for the operation
- Use ./bin/securitysetup to reconfigure authentication
`
	
	default:
		return "No detailed help available for this error"
	}
}

// Example helper functions for each security component

// NewTLSError creates a new TLS-related error
func NewTLSError(code, message string) *SecurityError {
	return NewError(code, message)
}

// NewSignatureError creates a new signature-related error
func NewSignatureError(code, message string) *SecurityError {
	return NewError(code, message)
}

// NewAuthError creates a new authentication-related error
func NewAuthError(code, message string) *SecurityError {
	return NewError(code, message)
}

// ConvertError converts a regular error to a SecurityError if possible
func ConvertError(err error, defaultCode string) *SecurityError {
	if err == nil {
		return nil
	}
	
	// If it's already a SecurityError, return it
	var secErr *SecurityError
	if errors.As(err, &secErr) {
		return secErr
	}
	
	// Convert to a security error
	return NewError(defaultCode, err.Error()).WithCause(err)
}