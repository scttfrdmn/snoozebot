package protocol

import (
	"time"
)

// InstanceRegistration represents the registration of an instance with the agent
type InstanceRegistration struct {
	// InstanceID is the unique identifier for the instance
	InstanceID string `json:"instance_id"`
	
	// InstanceType is the type of instance (e.g., AWS instance type)
	InstanceType string `json:"instance_type"`
	
	// Region is the region where the instance is located
	Region string `json:"region"`
	
	// Zone is the availability zone where the instance is located
	Zone string `json:"zone"`
	
	// Provider is the cloud provider (aws, gcp, azure)
	Provider string `json:"provider"`
	
	// Metadata is a map of additional metadata about the instance
	Metadata map[string]string `json:"metadata"`
	
	// Thresholds is a map of resource types to thresholds
	Thresholds map[string]float64 `json:"thresholds"`
	
	// NapTime is the duration that the instance must be idle before it can be stopped
	NapTime time.Duration `json:"nap_time"`
}

// RegistrationResponse is the response to an instance registration
type RegistrationResponse struct {
	// Success indicates if the registration was successful
	Success bool `json:"success"`
	
	// Error is an error message if the registration failed
	Error string `json:"error,omitempty"`
	
	// AgentID is the ID of the agent handling this instance
	AgentID string `json:"agent_id"`
	
	// HeartbeatInterval is the interval at which the instance should send heartbeats
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

// IdleNotification represents a notification that an instance is idle
type IdleNotification struct {
	// InstanceID is the ID of the instance
	InstanceID string `json:"instance_id"`
	
	// IdleSince is the time when the instance became idle
	IdleSince time.Time `json:"idle_since"`
	
	// IdleDuration is how long the instance has been idle
	IdleDuration time.Duration `json:"idle_duration"`
	
	// ResourceUsage is a map of resource types to their current usage
	ResourceUsage map[string]float64 `json:"resource_usage"`
}

// IdleNotificationResponse is the response to an idle notification
type IdleNotificationResponse struct {
	// Action is the action to take (none, wait, stop)
	Action string `json:"action"`
	
	// Reason is the reason for the action
	Reason string `json:"reason,omitempty"`
	
	// ScheduledAction is a scheduled action for the instance
	ScheduledAction *ScheduledAction `json:"scheduled_action,omitempty"`
}

// ScheduledAction represents an action scheduled for an instance
type ScheduledAction struct {
	// Action is the action to perform (stop, start, etc.)
	Action string `json:"action"`
	
	// ScheduledTime is when the action is scheduled to occur
	ScheduledTime time.Time `json:"scheduled_time"`
	
	// Reason is the reason for the action
	Reason string `json:"reason,omitempty"`
}

// Heartbeat represents a heartbeat from an instance to the agent
type Heartbeat struct {
	// InstanceID is the ID of the instance
	InstanceID string `json:"instance_id"`
	
	// Timestamp is the time of the heartbeat
	Timestamp time.Time `json:"timestamp"`
	
	// State is the current state of the instance (running, idle, etc.)
	State string `json:"state"`
	
	// ResourceUsage is a map of resource types to their current usage
	ResourceUsage map[string]float64 `json:"resource_usage,omitempty"`
}

// HeartbeatResponse is the response to a heartbeat
type HeartbeatResponse struct {
	// Acknowledged indicates if the heartbeat was acknowledged
	Acknowledged bool `json:"acknowledged"`
	
	// Commands is a list of commands for the instance to execute
	Commands []InstanceCommand `json:"commands,omitempty"`
}

// InstanceCommand represents a command for an instance to execute
type InstanceCommand struct {
	// Command is the command to execute (ping, stop, start, etc.)
	Command string `json:"command"`
	
	// Parameters is a map of parameters for the command
	Parameters map[string]string `json:"parameters,omitempty"`
}

// InstanceStateChange represents a change in the state of an instance
type InstanceStateChange struct {
	// InstanceID is the ID of the instance
	InstanceID string `json:"instance_id"`
	
	// PreviousState is the previous state of the instance
	PreviousState string `json:"previous_state"`
	
	// CurrentState is the current state of the instance
	CurrentState string `json:"current_state"`
	
	// Timestamp is when the state change occurred
	Timestamp time.Time `json:"timestamp"`
	
	// Reason is the reason for the state change
	Reason string `json:"reason,omitempty"`
}

// InstanceStateChangeResponse is the response to an instance state change
type InstanceStateChangeResponse struct {
	// Acknowledged indicates if the state change was acknowledged
	Acknowledged bool `json:"acknowledged"`
	
	// Error is an error message if there was a problem
	Error string `json:"error,omitempty"`
}