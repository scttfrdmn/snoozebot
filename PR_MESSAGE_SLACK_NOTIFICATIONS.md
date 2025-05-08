# Slack Notifications for Snoozebot

This PR adds a flexible notification system to Snoozebot with Slack integration, allowing the agent to send notifications about important events to Slack channels. This addresses feature request #XX.

## Features Added

1. **Notification Framework**: 
   - Created a flexible, extensible notification system that can support multiple notification providers
   - Implemented a notification manager to handle routing of notifications
   - Created a standard notification format with type, severity, and data fields

2. **Slack Integration**:
   - Added Slack webhook integration for sending formatted messages to Slack channels
   - Customizable appearance with support for channel, username, and icon configuration
   - Color-coded messages based on notification severity

3. **Event Notifications**:
   - Idle state notifications when instances have been idle for certain periods
   - Scheduled action notifications when actions (like stopping instances) are scheduled
   - State change notifications when instances change state

4. **Configuration System**:
   - YAML-based configuration for notification providers
   - Default configuration generation for new setups
   - Support for multiple simultaneous notification providers

5. **Documentation**:
   - Added comprehensive documentation on setting up Slack notifications
   - Documented the notification architecture for developers
   - Updated README to include information about notifications

## Implementation Details

The notification system is designed with extensibility in mind:

- The `notification.NotificationProvider` interface allows for easy addition of new providers
- Notifications are sent asynchronously to avoid blocking the main process
- Common notification types are predefined but the system can be extended
- The notification manager handles provider registration and initialization

## Testing

I've tested this implementation by:

1. Setting up a test Slack workspace and creating an incoming webhook
2. Integrating the notification system into the agent server
3. Manually triggering different notification types and verifying they appear in Slack
4. Testing error handling by intentionally using invalid webhook URLs

## Future Improvements

Potential future improvements could include:

1. Additional notification providers (email, Discord, Microsoft Teams, etc.)
2. More customizable notification templates
3. User-configurable notification filtering
4. Rate limiting for high-volume notification events

## Screenshots

![Slack Notification Example](path_to_screenshot.png)

## Related Issues

- Related to feature request #XX