# Email Notifications for Snoozebot

Snoozebot can send email notifications when important events occur, such as instances becoming idle, actions being scheduled, or instances changing state.

## Setup

### 1. Configure SMTP Settings

To set up email notifications, you need access to an SMTP server. This could be:

- Your organization's SMTP server
- A cloud-based email service (like Gmail, Amazon SES, SendGrid, etc.)
- A self-hosted mail server

You will need the following information:

- SMTP server address (e.g., `smtp.gmail.com`)
- SMTP port (typically 587 for STARTTLS or 465 for SSL)
- Username and password for authentication
- Sender email address

### 2. Configure Snoozebot

Create or edit the `notifications.yaml` file in your Snoozebot config directory (the same directory where you store your authentication configuration). Add or edit the email section:

```yaml
providers:
  email:
    enabled: true
    config:
      smtp_server: "smtp.example.com"
      smtp_port: 587
      username: "notifications@example.com"
      password: "your_password"
      from_address: "Snoozebot <notifications@example.com>"
      to_addresses: ["admin@example.com", "oncall@example.com"]
      subject_prefix: "[Snoozebot]"
      enable_starttls: true
      # enable_ssl: false            # Use SSL/TLS instead of STARTTLS
      # skip_tls_verify: false       # Skip TLS certificate verification (not recommended)
```

### 3. Restart Snoozebot

After creating or modifying the configuration file, restart the Snoozebot agent to apply the changes.

## Configuration Options

### Required Fields

- `smtp_server`: The SMTP server address
- `username`: The SMTP username for authentication
- `password`: The SMTP password for authentication
- `from_address`: The email address to send from
- `to_addresses`: A list of email addresses to send notifications to

### Optional Fields

- `smtp_port`: The SMTP server port (defaults to 587)
- `subject_prefix`: Text to prepend to all email subjects
- `enable_starttls`: Whether to use STARTTLS (defaults to true)
- `enable_ssl`: Whether to use SSL/TLS instead of STARTTLS
- `skip_tls_verify`: Whether to skip TLS certificate verification (not recommended for production)

## Using Gmail as SMTP Server

If you want to use Gmail as your SMTP server, you'll need to use an app password instead of your regular password:

1. Enable 2-Step Verification on your Google account
2. Go to your Google Account > Security > App passwords
3. Generate a new app password for Snoozebot
4. Use the following configuration:

```yaml
providers:
  email:
    enabled: true
    config:
      smtp_server: "smtp.gmail.com"
      smtp_port: 587
      username: "your.email@gmail.com"
      password: "your-app-password"
      from_address: "Snoozebot <your.email@gmail.com>"
      to_addresses: ["recipient@example.com"]
      enable_starttls: true
```

## Testing Email Notifications

You can test if email notifications are working by using the provided test script:

```bash
go run scripts/test_email_notification.go \
  -smtp-server=smtp.example.com \
  -username=user@example.com \
  -password=yourpassword \
  -from="Snoozebot <user@example.com>" \
  -to=recipient@example.com
```

## Email Format

Email notifications include:

1. A subject line with:
   - Optional subject prefix
   - Severity (for non-info notifications)
   - Notification title
   - Instance name (if available)

2. A body with:
   - Notification title
   - Main message
   - Instance details (ID, name, provider, region)
   - Notification-specific details
   - Timestamp and metadata

## Troubleshooting

If email notifications aren't working:

1. Check the Snoozebot logs for any errors related to email notifications
2. Verify your SMTP server address and port are correct
3. Make sure your username and password are correct
4. Ensure your SMTP server allows the connection:
   - Some servers restrict access by IP address
   - Some require specific authentication methods
   - Some have rate limits or other restrictions
5. Try the test script to isolate configuration issues

### Common SMTP Issues

- **Authentication Failed**: Check your username and password. For Gmail, ensure you're using an app password.
- **Connection Timeout**: Check your SMTP server address and port. Make sure your network allows the connection.
- **TLS/SSL Issues**: Try toggling between STARTTLS and SSL, or temporarily enable `skip_tls_verify`.
- **Rate Limiting**: Some SMTP servers limit the number of emails you can send in a period. Check if you've hit a limit.

## Security Considerations

- Never commit SMTP credentials to source control
- Consider using environment variables or a secrets manager
- Use TLS/SSL for secure email transmission
- Regularly rotate SMTP passwords
- Consider using a dedicated email account for notifications