package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/notification/types"
)

// Manager manages notification providers and handles sending notifications
type Manager struct {
	providers map[string]types.NotificationProvider
	logger    hclog.Logger
	mu        sync.RWMutex
}

// NewManager creates a new notification manager
func NewManager(logger hclog.Logger) *Manager {
	return &Manager{
		providers: make(map[string]types.NotificationProvider),
		logger:    logger.Named("notification-manager"),
	}
}

// RegisterProvider registers a notification provider
func (m *Manager) RegisterProvider(provider types.NotificationProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := provider.Name()
	if _, exists := m.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	m.providers[name] = provider
	m.logger.Info("Registered notification provider", "name", name)
	return nil
}

// InitProvider initializes a provider with the given configuration
func (m *Manager) InitProvider(name string, config map[string]interface{}) error {
	m.mu.RLock()
	provider, exists := m.providers[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("provider %s not registered", name)
	}

	err := provider.Init(config)
	if err != nil {
		return fmt.Errorf("failed to initialize provider %s: %w", name, err)
	}

	m.logger.Info("Initialized notification provider", "name", name)
	return nil
}

// SendNotification sends a notification to all registered providers
func (m *Manager) SendNotification(ctx context.Context, notification *types.Notification) []error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.providers) == 0 {
		m.logger.Warn("No notification providers registered")
		return nil
	}

	// Set timestamp if not already set
	if notification.Timestamp.IsZero() {
		notification.Timestamp = time.Now()
	}

	var errors []error
	var wg sync.WaitGroup

	for name, provider := range m.providers {
		wg.Add(1)
		go func(name string, provider types.NotificationProvider) {
			defer wg.Done()

			err := provider.Send(ctx, notification)
			if err != nil {
				m.logger.Error("Failed to send notification",
					"provider", name,
					"type", notification.Type,
					"error", err)
				
				m.mu.Lock()
				errors = append(errors, fmt.Errorf("provider %s: %w", name, err))
				m.mu.Unlock()
			}
		}(name, provider)
	}

	wg.Wait()
	return errors
}

// NotifyIdle creates and sends an idle notification
func (m *Manager) NotifyIdle(ctx context.Context, instanceID, instanceName, provider, region string, idleDuration time.Duration) []error {
	notification := &types.Notification{
		Type:         types.NotificationTypeIdle,
		Severity:     types.SeverityInfo,
		InstanceID:   instanceID,
		InstanceName: instanceName,
		Provider:     provider,
		Region:       region,
		Title:        "Instance Idle",
		Message:      fmt.Sprintf("Instance %s has been idle for %s", instanceName, idleDuration),
		Data: map[string]interface{}{
			"idle_duration": idleDuration.String(),
		},
	}

	return m.SendNotification(ctx, notification)
}

// NotifyScheduledAction creates and sends a scheduled action notification
func (m *Manager) NotifyScheduledAction(ctx context.Context, instanceID, instanceName, provider, region, action string, scheduledTime time.Time, reason string) []error {
	notification := &types.Notification{
		Type:         types.NotificationTypeScheduledAction,
		Severity:     types.SeverityWarning,
		InstanceID:   instanceID,
		InstanceName: instanceName,
		Provider:     provider,
		Region:       region,
		Title:        fmt.Sprintf("Scheduled Action: %s", action),
		Message:      fmt.Sprintf("Action %s scheduled for instance %s at %s. Reason: %s", action, instanceName, scheduledTime.Format(time.RFC3339), reason),
		Data: map[string]interface{}{
			"action":         action,
			"scheduled_time": scheduledTime.Format(time.RFC3339),
			"reason":         reason,
		},
	}

	return m.SendNotification(ctx, notification)
}

// NotifyActionExecuted creates and sends an action executed notification
func (m *Manager) NotifyActionExecuted(ctx context.Context, instanceID, instanceName, provider, region, action, result string) []error {
	notification := &types.Notification{
		Type:         types.NotificationTypeActionExecuted,
		Severity:     types.SeverityInfo,
		InstanceID:   instanceID,
		InstanceName: instanceName,
		Provider:     provider,
		Region:       region,
		Title:        fmt.Sprintf("Action Executed: %s", action),
		Message:      fmt.Sprintf("Action %s executed on instance %s. Result: %s", action, instanceName, result),
		Data: map[string]interface{}{
			"action": action,
			"result": result,
		},
	}

	return m.SendNotification(ctx, notification)
}

// NotifyError creates and sends an error notification
func (m *Manager) NotifyError(ctx context.Context, instanceID, instanceName, provider, region, errorType, errorMessage string) []error {
	notification := &types.Notification{
		Type:         types.NotificationTypeError,
		Severity:     types.SeverityError,
		InstanceID:   instanceID,
		InstanceName: instanceName,
		Provider:     provider,
		Region:       region,
		Title:        fmt.Sprintf("Error: %s", errorType),
		Message:      errorMessage,
		Data: map[string]interface{}{
			"error_type": errorType,
		},
	}

	return m.SendNotification(ctx, notification)
}

// NotifyStateChange creates and sends a state change notification
func (m *Manager) NotifyStateChange(ctx context.Context, instanceID, instanceName, provider, region, previousState, currentState, reason string) []error {
	notification := &types.Notification{
		Type:         types.NotificationTypeStateChange,
		Severity:     types.SeverityInfo,
		InstanceID:   instanceID,
		InstanceName: instanceName,
		Provider:     provider,
		Region:       region,
		Title:        "Instance State Change",
		Message:      fmt.Sprintf("Instance %s state changed from %s to %s. Reason: %s", instanceName, previousState, currentState, reason),
		Data: map[string]interface{}{
			"previous_state": previousState,
			"current_state":  currentState,
			"reason":         reason,
		},
	}

	return m.SendNotification(ctx, notification)
}

// Close closes all providers
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, provider := range m.providers {
		err := provider.Close()
		if err != nil {
			m.logger.Error("Failed to close notification provider", "name", name, "error", err)
			lastErr = err
		}
	}

	return lastErr
}