# Slack Notifications for Snoozebot

Snoozebot can send notifications to Slack when important events occur, such as instances becoming idle, actions being scheduled, or instances changing state.

## Setup

### 1. Create a Slack App

1. Go to [https://api.slack.com/apps](https://api.slack.com/apps) and click "Create New App"
2. Choose "From scratch"
3. Name your app (e.g., "Snoozebot") and select your workspace
4. Click "Create App"

### 2. Configure Incoming Webhooks

1. In the app settings, click on "Incoming Webhooks"
2. Toggle "Activate Incoming Webhooks" to On
3. Click "Add New Webhook to Workspace"
4. Select the channel where you want to receive notifications
5. Click "Allow"
6. Copy the webhook URL (it should start with `https://hooks.slack.com/services/`)

### 3. Configure Snoozebot

Create a configuration file named `notifications.yaml` in your Snoozebot config directory (the same directory where you store your authentication configuration). The structure of the file should be:

```yaml
providers:
  slack:
    enabled: true
    config:
      webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
      channel: "#your-channel"  # Optional, overrides the webhook default
      username: "Snoozebot"     # Optional, overrides the webhook default
      icon_emoji: ":robot_face:"  # Optional, overrides the webhook default
      # icon_url: "https://example.com/icon.png"  # Optional, alternative to icon_emoji
```

### 4. Restart Snoozebot

After creating or modifying the configuration file, restart the Snoozebot agent to apply the changes.

## Notification Types

Snoozebot will send the following types of notifications to Slack:

### Idle Notifications

Sent when an instance has been idle for a certain period. The notification includes:

- Instance ID and name (if available)
- Provider and region
- Idle duration

### Scheduled Action Notifications

Sent when an action (usually stopping an instance) is scheduled. The notification includes:

- Instance ID and name (if available)
- Provider and region
- Action type (e.g., "stop")
- Scheduled time
- Reason for the action

### State Change Notifications

Sent when an instance changes state. The notification includes:

- Instance ID and name (if available)
- Provider and region
- Previous state
- Current state
- Reason for the state change

## Customizing Notifications

The Slack notification provider formats messages to include relevant information based on the notification type. You can customize the appearance by modifying the webhook configuration:

- `channel`: The Slack channel to send notifications to
- `username`: The username that will appear as the sender of the message
- `icon_emoji`: An emoji to use as the icon (e.g., `:robot_face:`)
- `icon_url`: A URL to an image to use as the icon (alternative to icon_emoji)

## Testing Notifications

You can test if notifications are working by manually triggering an idle notification. Start a Snoozebot monitor and let it become idle, or manually create an idle state through the API.

## Troubleshooting

If notifications aren't working:

1. Check the Snoozebot logs for any errors related to notifications
2. Verify that the webhook URL is correct and the app is properly installed in your workspace
3. Ensure the Slack app has permission to post to the channel you specified
4. Try recreating the webhook in Slack and updating the configuration

## Security Considerations

The webhook URL provides access to post messages to your Slack workspace. Keep it secure:

- Don't commit the webhook URL to source control
- Consider using environment variables or a secrets manager to store the URL
- Regularly rotate the webhook URL by recreating it in Slack if you suspect it may have been compromised