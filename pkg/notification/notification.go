package notification

import (
	"github.com/scttfrdmn/snoozebot/pkg/notification/types"
)

// Reexport types from the types package for backward compatibility
type (
	NotificationType = types.NotificationType
	Severity         = types.Severity
	Notification     = types.Notification
	NotificationProvider = types.NotificationProvider
)

// Constants reexported from the types package
const (
	NotificationTypeIdle            = types.NotificationTypeIdle
	NotificationTypeScheduledAction = types.NotificationTypeScheduledAction
	NotificationTypeActionExecuted  = types.NotificationTypeActionExecuted
	NotificationTypeError           = types.NotificationTypeError
	NotificationTypeStateChange     = types.NotificationTypeStateChange

	SeverityInfo     = types.SeverityInfo
	SeverityWarning  = types.SeverityWarning
	SeverityError    = types.SeverityError
	SeverityCritical = types.SeverityCritical
)