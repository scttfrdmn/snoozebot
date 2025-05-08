package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
)

// Event severity levels
const (
	InfoLevel     = "INFO"
	WarningLevel  = "WARNING"
	ErrorLevel    = "ERROR"
	CriticalLevel = "CRITICAL"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	Timestamp   time.Time         `json:"timestamp"`
	Level       string            `json:"level"`
	Category    string            `json:"category"`
	EventType   string            `json:"event_type"`
	Message     string            `json:"message"`
	Component   string            `json:"component"`
	UserID      string            `json:"user_id,omitempty"`
	PluginName  string            `json:"plugin_name,omitempty"`
	IPAddress   string            `json:"ip_address,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
	RelatedToID string            `json:"related_to_id,omitempty"`
	Success     bool              `json:"success"`
}

// Common security event types
const (
	// Authentication events
	EventAuthSuccess      = "AUTH_SUCCESS"
	EventAuthFailure      = "AUTH_FAILURE"
	EventAPIKeyCreated    = "APIKEY_CREATED"
	EventAPIKeyRevoked    = "APIKEY_REVOKED"
	EventRoleChange       = "ROLE_CHANGE"
	EventPermissionDenied = "PERMISSION_DENIED"

	// TLS events
	EventTLSHandshake     = "TLS_HANDSHAKE"
	EventCertGenerated    = "CERT_GENERATED"
	EventCertExpired      = "CERT_EXPIRED"
	EventTLSVerification  = "TLS_VERIFICATION"

	// Signature events
	EventSignatureVerify   = "SIG_VERIFY"
	EventPluginSigned      = "PLUGIN_SIGNED"
	EventSignatureInvalid  = "SIG_INVALID"
	EventKeyGenerated      = "KEY_GENERATED"
	EventKeyTrusted        = "KEY_TRUSTED"
	EventKeyRevoked        = "KEY_REVOKED"

	// Plugin events
	EventPluginLoaded     = "PLUGIN_LOADED"
	EventPluginUnloaded   = "PLUGIN_UNLOADED"
	EventPluginCrashed    = "PLUGIN_CRASHED"

	// System events
	EventSystemStartup    = "SYSTEM_STARTUP"
	EventSystemShutdown   = "SYSTEM_SHUTDOWN"
	EventConfigChanged    = "CONFIG_CHANGED"
	EventAuditStarted     = "AUDIT_STARTED"
	EventAuditCompleted   = "AUDIT_COMPLETED"
)

// SecurityEventManager handles security events
type SecurityEventManager struct {
	logger       hclog.Logger
	eventsDir    string
	currentFile  *os.File
	mutex        sync.Mutex
	rotateSize   int64
	rotateCount  int
	callbacks    map[string][]EventCallback
	callbackMu   sync.RWMutex
	enabled      bool
	consoleOutput bool
	fileOutput    bool
	currentSize   int64
}

// EventCallback is a function called when events occur
type EventCallback func(event *SecurityEvent)

// NewSecurityEventManager creates a new security event manager
func NewSecurityEventManager(eventsDir string, logger hclog.Logger) (*SecurityEventManager, error) {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	if eventsDir == "" {
		eventsDir = "/var/log/snoozebot/security"
	}

	// Create events directory if it doesn't exist
	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create events directory: %w", err)
	}

	return &SecurityEventManager{
		logger:         logger,
		eventsDir:      eventsDir,
		callbacks:      make(map[string][]EventCallback),
		rotateSize:     10 * 1024 * 1024, // 10MB
		rotateCount:    5,
		enabled:        true,
		consoleOutput:  true,
		fileOutput:     true,
	}, nil
}

// SetRotationSettings sets log rotation settings
func (m *SecurityEventManager) SetRotationSettings(maxSize int64, maxCount int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.rotateSize = maxSize
	m.rotateCount = maxCount
}

// EnableConsoleOutput enables or disables console output
func (m *SecurityEventManager) EnableConsoleOutput(enabled bool) {
	m.consoleOutput = enabled
}

// EnableFileOutput enables or disables file output
func (m *SecurityEventManager) EnableFileOutput(enabled bool) {
	m.fileOutput = enabled
}

// RegisterCallback registers a callback for specific event types
func (m *SecurityEventManager) RegisterCallback(eventType string, callback EventCallback) {
	m.callbackMu.Lock()
	defer m.callbackMu.Unlock()
	
	m.callbacks[eventType] = append(m.callbacks[eventType], callback)
}

// openLogFile opens a new log file
func (m *SecurityEventManager) openLogFile() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Close current file if open
	if m.currentFile != nil {
		m.currentFile.Close()
		m.currentFile = nil
	}
	
	// Create new log file
	timestamp := time.Now().Format("20060102-150405")
	logPath := filepath.Join(m.eventsDir, fmt.Sprintf("security-%s.log", timestamp))
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	
	m.currentFile = file
	m.currentSize = 0
	
	// Create symlink to latest log
	latestPath := filepath.Join(m.eventsDir, "security-latest.log")
	os.Remove(latestPath) // Remove existing symlink if it exists
	if err := os.Symlink(logPath, latestPath); err != nil {
		m.logger.Warn("Failed to create symlink to latest log", "error", err)
	}
	
	// Clean up old log files
	m.cleanupOldLogs()
	
	return nil
}

// cleanupOldLogs removes old log files
func (m *SecurityEventManager) cleanupOldLogs() {
	// List all log files
	pattern := filepath.Join(m.eventsDir, "security-*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		m.logger.Warn("Failed to list log files", "error", err)
		return
	}
	
	// Skip cleanup if we have fewer files than max count
	if len(matches) <= m.rotateCount {
		return
	}
	
	// Sort files by modification time
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	
	files := make([]fileInfo, 0, len(matches))
	for _, match := range matches {
		// Skip latest symlink
		if match == filepath.Join(m.eventsDir, "security-latest.log") {
			continue
		}
		
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}
		
		files = append(files, fileInfo{
			path:    match,
			modTime: stat.ModTime(),
		})
	}
	
	// Sort files by modification time (oldest first)
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
	
	// Remove oldest files until we're under the limit
	for i := 0; i < len(files)-m.rotateCount; i++ {
		os.Remove(files[i].path)
		m.logger.Debug("Removed old log file", "path", files[i].path)
	}
}

// LogEvent logs a security event
func (m *SecurityEventManager) LogEvent(event *SecurityEvent) error {
	if !m.enabled {
		return nil
	}
	
	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	
	// Call registered callbacks
	m.callbackMu.RLock()
	if callbacks, ok := m.callbacks[event.EventType]; ok {
		for _, callback := range callbacks {
			callback(event)
		}
	}
	m.callbackMu.RUnlock()
	
	// Log to console if enabled
	if m.consoleOutput {
		level := hclog.Info
		switch event.Level {
		case WarningLevel:
			level = hclog.Warn
		case ErrorLevel:
			level = hclog.Error
		case CriticalLevel:
			level = hclog.Error
		}
		
		m.logger.Log(level, event.Message,
			"type", event.EventType,
			"category", event.Category,
			"component", event.Component,
			"success", event.Success,
		)
	}
	
	// Return if file output is disabled
	if !m.fileOutput {
		return nil
	}
	
	// Convert event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	data = append(data, '\n')
	
	// Check if we need to open a new log file
	if m.currentFile == nil {
		if err := m.openLogFile(); err != nil {
			return err
		}
	}
	
	// Check if we need to rotate the log file
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.currentSize >= m.rotateSize {
		if err := m.openLogFile(); err != nil {
			return err
		}
	}
	
	// Write the event to the log file
	n, err := m.currentFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	
	m.currentSize += int64(n)
	return nil
}

// CreateEvent creates a new security event
func CreateEvent(eventType, message, component, category string) *SecurityEvent {
	return &SecurityEvent{
		Timestamp: time.Now(),
		Level:     InfoLevel,
		EventType: eventType,
		Message:   message,
		Component: component,
		Category:  category,
		Success:   true,
		Details:   make(map[string]string),
	}
}

// WithLevel sets the event level
func (e *SecurityEvent) WithLevel(level string) *SecurityEvent {
	e.Level = level
	return e
}

// WithDetails adds details to the event
func (e *SecurityEvent) WithDetails(key, value string) *SecurityEvent {
	e.Details[key] = value
	return e
}

// WithUserID sets the user ID
func (e *SecurityEvent) WithUserID(userID string) *SecurityEvent {
	e.UserID = userID
	return e
}

// WithPluginName sets the plugin name
func (e *SecurityEvent) WithPluginName(pluginName string) *SecurityEvent {
	e.PluginName = pluginName
	return e
}

// WithIPAddress sets the IP address
func (e *SecurityEvent) WithIPAddress(ipAddress string) *SecurityEvent {
	e.IPAddress = ipAddress
	return e
}

// WithSuccess sets the success flag
func (e *SecurityEvent) WithSuccess(success bool) *SecurityEvent {
	e.Success = success
	return e
}

// WithRelatedID sets the related ID
func (e *SecurityEvent) WithRelatedID(relatedID string) *SecurityEvent {
	e.RelatedToID = relatedID
	return e
}

// Event helper functions for common events

// LogAuthSuccess logs an authentication success event
func LogAuthSuccess(manager *SecurityEventManager, userID, component string, details map[string]string) error {
	event := CreateEvent(EventAuthSuccess, "Authentication successful", component, "authentication").
		WithLevel(InfoLevel).
		WithUserID(userID).
		WithSuccess(true)
	
	// Add details
	for k, v := range details {
		event.WithDetails(k, v)
	}
	
	return manager.LogEvent(event)
}

// LogAuthFailure logs an authentication failure event
func LogAuthFailure(manager *SecurityEventManager, message, userID, component string, details map[string]string) error {
	event := CreateEvent(EventAuthFailure, message, component, "authentication").
		WithLevel(WarningLevel).
		WithUserID(userID).
		WithSuccess(false)
	
	// Add details
	for k, v := range details {
		event.WithDetails(k, v)
	}
	
	return manager.LogEvent(event)
}

// LogTLSHandshake logs a TLS handshake event
func LogTLSHandshake(manager *SecurityEventManager, success bool, component, pluginName string, details map[string]string) error {
	message := "TLS handshake successful"
	level := InfoLevel
	
	if !success {
		message = "TLS handshake failed"
		level = ErrorLevel
	}
	
	event := CreateEvent(EventTLSHandshake, message, component, "tls").
		WithLevel(level).
		WithPluginName(pluginName).
		WithSuccess(success)
	
	// Add details
	for k, v := range details {
		event.WithDetails(k, v)
	}
	
	return manager.LogEvent(event)
}

// LogSignatureVerify logs a signature verification event
func LogSignatureVerify(manager *SecurityEventManager, success bool, component, pluginName string, details map[string]string) error {
	message := "Signature verification successful"
	level := InfoLevel
	
	if !success {
		message = "Signature verification failed"
		level = ErrorLevel
	}
	
	event := CreateEvent(EventSignatureVerify, message, component, "signature").
		WithLevel(level).
		WithPluginName(pluginName).
		WithSuccess(success)
	
	// Add details
	for k, v := range details {
		event.WithDetails(k, v)
	}
	
	return manager.LogEvent(event)
}

// LogPluginLoaded logs a plugin loaded event
func LogPluginLoaded(manager *SecurityEventManager, pluginName, component string, details map[string]string) error {
	event := CreateEvent(EventPluginLoaded, "Plugin loaded successfully", component, "plugin").
		WithLevel(InfoLevel).
		WithPluginName(pluginName).
		WithSuccess(true)
	
	// Add details
	for k, v := range details {
		event.WithDetails(k, v)
	}
	
	return manager.LogEvent(event)
}