# Snoozebot Notification System

The Snoozebot notification system provides a flexible framework for sending notifications about important events to various channels, such as Slack, email, or custom webhooks.

## Architecture

The notification system consists of the following components:

### Notification Manager

The notification manager (`notification.Manager`) is the central component that:

- Maintains a registry of notification providers
- Provides methods for sending different types of notifications
- Routes notifications to the appropriate providers

### Notification Providers

Notification providers implement the `notification.NotificationProvider` interface and handle the actual delivery of notifications. The currently implemented providers are:

- Slack: Sends notifications to Slack channels using webhooks

### Notification Types

The system supports several notification types:

- **Idle**: Sent when an instance has been idle for a certain period
- **Scheduled Action**: Sent when an action (e.g., stopping an instance) is scheduled
- **Action Executed**: Sent when an action has been executed
- **Error**: Sent when an error occurs
- **State Change**: Sent when an instance changes state

### Severity Levels

Each notification has a severity level that can be used to determine the visual appearance:

- **Info**: Normal informational messages
- **Warning**: Warning messages that might require attention
- **Error**: Error messages that require attention
- **Critical**: Critical messages that require immediate attention

## Configuration

The notification system is configured using a YAML file (`notifications.yaml`) in the configuration directory. The file structure is:

```yaml
providers:
  slack:
    enabled: true
    config:
      webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
      channel: "#your-channel"
      username: "Snoozebot"
      icon_emoji: ":robot_face:"
```

The configuration file is loaded during the initialization of the notification manager.

## Integration Points

The notification system is integrated into the Snoozebot agent at the following points:

1. **Idle Notifications**: When an instance reports it's idle
2. **Scheduled Actions**: When an action is scheduled for an instance
3. **State Changes**: When an instance reports a state change

## Extending the System

The notification system is designed to be extensible, making it easy to add new providers or notification types.

## Supported Notification Providers

The following notification providers are included:

### Slack Provider

The Slack provider sends notifications to Slack channels using webhooks. It supports:

- Customizable message formatting with attachments
- Color-coding based on severity
- Channel, username, and icon customization
- Detailed fields based on notification type

### Email Provider

The Email provider sends notifications via SMTP email. It supports:

- SMTP with STARTTLS or SSL/TLS connections
- Multiple recipients
- Customizable subject prefix
- Authentication with username/password
- Formatted email bodies with detailed information

### Adding a New Provider

To add a new notification provider:

1. Create a new package under `pkg/notification/providers/<provider_name>`
2. Implement the `notification.NotificationProvider` interface
3. Register the provider in `pkg/notification/provider_registry.go`
4. Update the configuration documentation

### Adding a New Notification Type

To add a new notification type:

1. Add a new constant in `pkg/notification/notification.go`
2. Add a new method in `pkg/notification/manager.go` for the notification type
3. Update the provider implementations to handle the new type

## Best Practices

When using or extending the notification system:

1. **Asynchronous Processing**: Always send notifications asynchronously (using goroutines) to avoid blocking the main process
2. **Error Handling**: Handle notification errors gracefully, logging them but not failing the main process
3. **Context**: Always pass a context to notification methods to allow for cancellation or timeout
4. **Rate Limiting**: Consider implementing rate limiting for providers that might have API limits
5. **Template Customization**: Allow users to customize notification templates when possible