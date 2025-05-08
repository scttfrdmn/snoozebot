package notification

import (
	"context"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// NotificationTypeIdle is sent when an instance becomes idle
	NotificationTypeIdle NotificationType = "idle"
	
	// NotificationTypeScheduledAction is sent when an action is scheduled for an instance
	NotificationTypeScheduledAction NotificationType = "scheduled_action"
	
	// NotificationTypeActionExecuted is sent when an action is executed
	NotificationTypeActionExecuted NotificationType = "action_executed"
	
	// NotificationTypeError is sent when an error occurs
	NotificationTypeError NotificationType = "error"
	
	// NotificationTypeStateChange is sent when an instance changes state
	NotificationTypeStateChange NotificationType = "state_change"
)

// Severity represents the severity level of a notification
type Severity string

const (
	// SeverityInfo is for informational messages
	SeverityInfo Severity = "info"
	
	// SeverityWarning is for warning messages
	SeverityWarning Severity = "warning"
	
	// SeverityError is for error messages
	SeverityError Severity = "error"
	
	// SeverityCritical is for critical messages
	SeverityCritical Severity = "critical"
)

// Notification represents a notification to be sent
type Notification struct {
	// Type is the type of notification
	Type NotificationType `json:"type"`
	
	// Severity is the severity level of the notification
	Severity Severity `json:"severity"`
	
	// Timestamp is when the notification was created
	Timestamp time.Time `json:"timestamp"`
	
	// InstanceID is the ID of the instance related to the notification
	InstanceID string `json:"instance_id,omitempty"`
	
	// InstanceName is the human-readable name of the instance
	InstanceName string `json:"instance_name,omitempty"`
	
	// Provider is the cloud provider (aws, gcp, azure)
	Provider string `json:"provider,omitempty"`
	
	// Region is the region where the instance is located
	Region string `json:"region,omitempty"`
	
	// Title is the title of the notification
	Title string `json:"title"`
	
	// Message is the message content
	Message string `json:"message"`
	
	// Data is any additional data for the notification
	Data map[string]interface{} `json:"data,omitempty"`
}

// NotificationProvider is the interface that notification providers must implement
type NotificationProvider interface {
	// Name returns the provider name
	Name() string
	
	// Init initializes the provider with the given configuration
	Init(config map[string]interface{}) error
	
	// Send sends a notification
	Send(ctx context.Context, notification *Notification) error
	
	// Close closes the provider and releases any resources
	Close() error
}