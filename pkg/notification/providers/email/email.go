package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/notification/types"
)

const (
	// ProviderName is the name of the email provider
	ProviderName = "email"

	// Default values
	DefaultSMTPPort = 587
	DefaultTimeout  = 10 * time.Second
)

// Config contains configuration for the email provider
type Config struct {
	// SMTPServer is the SMTP server address
	SMTPServer string `json:"smtp_server"`

	// SMTPPort is the SMTP server port
	SMTPPort int `json:"smtp_port,omitempty"`

	// Username is the SMTP username
	Username string `json:"username"`

	// Password is the SMTP password
	Password string `json:"password"`

	// FromAddress is the sender email address
	FromAddress string `json:"from_address"`

	// ToAddresses are the recipient email addresses
	ToAddresses []string `json:"to_addresses"`

	// EnableSSL enables SSL/TLS for SMTP
	EnableSSL bool `json:"enable_ssl,omitempty"`

	// EnableStartTLS enables STARTTLS for SMTP
	EnableStartTLS bool `json:"enable_starttls,omitempty"`

	// SkipTLSVerify skips TLS certificate verification
	SkipTLSVerify bool `json:"skip_tls_verify,omitempty"`

	// SubjectPrefix is an optional prefix for email subjects
	SubjectPrefix string `json:"subject_prefix,omitempty"`

	// Timeout is the timeout for SMTP operations
	Timeout time.Duration `json:"timeout,omitempty"`
}

// Provider implements types.NotificationProvider for email
type Provider struct {
	config      Config
	auth        smtp.Auth
	logger      hclog.Logger
	initialized bool
}

// New creates a new email notification provider
func New(logger hclog.Logger) *Provider {
	return &Provider{
		logger: logger.Named("email"),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return ProviderName
}

// Init initializes the provider with the given configuration
func (p *Provider) Init(config map[string]interface{}) error {
	// Ensure configuration is passed
	if config == nil {
		return fmt.Errorf("config is required")
	}

	// Parse server and port
	smtpServer, ok := config["smtp_server"].(string)
	if !ok || smtpServer == "" {
		return fmt.Errorf("smtp_server is required")
	}
	p.config.SMTPServer = smtpServer

	// Parse port with default
	if smtpPort, ok := config["smtp_port"].(float64); ok {
		p.config.SMTPPort = int(smtpPort)
	} else {
		p.config.SMTPPort = DefaultSMTPPort
	}

	// Parse credentials
	username, ok := config["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("username is required")
	}
	p.config.Username = username

	password, ok := config["password"].(string)
	if !ok || password == "" {
		return fmt.Errorf("password is required")
	}
	p.config.Password = password

	// Parse from address
	fromAddress, ok := config["from_address"].(string)
	if !ok || fromAddress == "" {
		return fmt.Errorf("from_address is required")
	}
	p.config.FromAddress = fromAddress

	// Parse to addresses
	toAddressesRaw, ok := config["to_addresses"].([]interface{})
	if !ok || len(toAddressesRaw) == 0 {
		return fmt.Errorf("to_addresses is required and must be a non-empty array")
	}
	toAddresses := make([]string, len(toAddressesRaw))
	for i, addr := range toAddressesRaw {
		if addrStr, ok := addr.(string); ok && addrStr != "" {
			toAddresses[i] = addrStr
		} else {
			return fmt.Errorf("to_addresses must be a list of valid email addresses")
		}
	}
	p.config.ToAddresses = toAddresses

	// Parse SSL/TLS options
	if enableSSL, ok := config["enable_ssl"].(bool); ok {
		p.config.EnableSSL = enableSSL
	}

	if enableStartTLS, ok := config["enable_starttls"].(bool); ok {
		p.config.EnableStartTLS = enableStartTLS
	} else {
		p.config.EnableStartTLS = true // Default to STARTTLS enabled
	}

	if skipTLSVerify, ok := config["skip_tls_verify"].(bool); ok {
		p.config.SkipTLSVerify = skipTLSVerify
	}

	// Parse other options
	if subjectPrefix, ok := config["subject_prefix"].(string); ok {
		p.config.SubjectPrefix = subjectPrefix
	}

	// Set default timeout
	p.config.Timeout = DefaultTimeout

	// Create SMTP auth
	p.auth = smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.SMTPServer)

	p.initialized = true
	p.logger.Info("Email notification provider initialized")
	return nil
}

// Send sends a notification via email
func (p *Provider) Send(ctx context.Context, n *types.Notification) error {
	if !p.initialized {
		return fmt.Errorf("provider not initialized")
	}

	// Create email subject
	subject := p.createSubject(n)

	// Create email body
	body := p.createBody(n)

	// Build email message
	message := p.buildEmailMessage(subject, body)

	// Build the SMTP address with port
	smtpAddr := fmt.Sprintf("%s:%d", p.config.SMTPServer, p.config.SMTPPort)

	// Send email
	var err error
	if p.config.EnableSSL {
		// Use SSL connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: p.config.SkipTLSVerify,
			ServerName:         p.config.SMTPServer,
		}
		err = p.sendEmailWithSSL(smtpAddr, tlsConfig, message)
	} else {
		// Use standard connection with optional STARTTLS
		err = p.sendEmail(smtpAddr, message)
	}

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	p.logger.Debug("Sent notification via email",
		"type", n.Type,
		"severity", n.Severity,
		"subject", subject,
		"recipients", len(p.config.ToAddresses))

	return nil
}

// Close closes the provider and releases any resources
func (p *Provider) Close() error {
	p.initialized = false
	return nil
}

// createSubject creates the email subject based on the notification
func (p *Provider) createSubject(n *types.Notification) string {
	var subject string

	// Add prefix if configured
	if p.config.SubjectPrefix != "" {
		subject = fmt.Sprintf("%s ", p.config.SubjectPrefix)
	}

	// Add severity for non-info notifications
	if n.Severity != types.SeverityInfo {
		subject += fmt.Sprintf("[%s] ", strings.ToUpper(string(n.Severity)))
	}

	// Add title
	subject += n.Title

	// Add instance info if available
	if n.InstanceName != "" {
		subject += fmt.Sprintf(" - %s", n.InstanceName)
	} else if n.InstanceID != "" {
		subject += fmt.Sprintf(" - %s", n.InstanceID)
	}

	return subject
}

// createBody creates the email body based on the notification
func (p *Provider) createBody(n *types.Notification) string {
	var builder strings.Builder

	// Add title
	builder.WriteString(fmt.Sprintf("# %s\n\n", n.Title))

	// Add message
	builder.WriteString(fmt.Sprintf("%s\n\n", n.Message))

	// Add instance details if available
	if n.InstanceID != "" {
		builder.WriteString("## Instance Details\n\n")
		
		// Instance name and ID
		if n.InstanceName != "" {
			builder.WriteString(fmt.Sprintf("- **Name**: %s\n", n.InstanceName))
			builder.WriteString(fmt.Sprintf("- **ID**: %s\n", n.InstanceID))
		} else {
			builder.WriteString(fmt.Sprintf("- **ID**: %s\n", n.InstanceID))
		}

		// Provider and region
		if n.Provider != "" {
			builder.WriteString(fmt.Sprintf("- **Provider**: %s\n", n.Provider))
		}
		if n.Region != "" {
			builder.WriteString(fmt.Sprintf("- **Region**: %s\n", n.Region))
		}
		
		builder.WriteString("\n")
	}

	// Add notification details based on type
	builder.WriteString("## Details\n\n")

	// Add timestamp
	builder.WriteString(fmt.Sprintf("- **Time**: %s\n", n.Timestamp.Format(time.RFC3339)))
	
	// Add type
	builder.WriteString(fmt.Sprintf("- **Type**: %s\n", n.Type))
	
	// Add severity
	builder.WriteString(fmt.Sprintf("- **Severity**: %s\n", n.Severity))

	// Add type-specific details
	switch n.Type {
	case types.NotificationTypeIdle:
		if idleDuration, ok := n.Data["idle_duration"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Idle Duration**: %s\n", idleDuration))
		}

	case types.NotificationTypeScheduledAction:
		if action, ok := n.Data["action"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Action**: %s\n", action))
		}
		if scheduledTime, ok := n.Data["scheduled_time"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Scheduled Time**: %s\n", scheduledTime))
		}
		if reason, ok := n.Data["reason"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Reason**: %s\n", reason))
		}

	case types.NotificationTypeActionExecuted:
		if action, ok := n.Data["action"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Action**: %s\n", action))
		}
		if result, ok := n.Data["result"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Result**: %s\n", result))
		}

	case types.NotificationTypeStateChange:
		if previousState, ok := n.Data["previous_state"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Previous State**: %s\n", previousState))
		}
		if currentState, ok := n.Data["current_state"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Current State**: %s\n", currentState))
		}
		if reason, ok := n.Data["reason"].(string); ok {
			builder.WriteString(fmt.Sprintf("- **Reason**: %s\n", reason))
		}
	}

	// Add footer
	builder.WriteString("\n---\n")
	builder.WriteString("This is an automated notification from Snoozebot.\n")

	return builder.String()
}

// buildEmailMessage creates a complete email message with headers and body
func (p *Provider) buildEmailMessage(subject, body string) []byte {
	// Build header
	header := make(map[string]string)
	header["From"] = p.config.FromAddress
	header["To"] = strings.Join(p.config.ToAddresses, ", ")
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	// Build message
	var message strings.Builder
	for key, value := range header {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	return []byte(message.String())
}

// sendEmail sends an email using a standard SMTP connection with optional STARTTLS
func (p *Provider) sendEmail(addr string, message []byte) error {
	// Create a client
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	// Use STARTTLS if enabled
	if p.config.EnableStartTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: p.config.SkipTLSVerify,
			ServerName:         p.config.SMTPServer,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}

	// Authenticate
	if err = client.Auth(p.auth); err != nil {
		return err
	}

	// Set sender
	if err = client.Mail(p.config.FromAddress); err != nil {
		return err
	}

	// Set recipients
	for _, recipient := range p.config.ToAddresses {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}

	// Send the message
	writer, err := client.Data()
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write(message)
	return err
}

// sendEmailWithSSL sends an email using an SSL connection
func (p *Provider) sendEmailWithSSL(addr string, tlsConfig *tls.Config, message []byte) error {
	// Connect over TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create client
	client, err := smtp.NewClient(conn, p.config.SMTPServer)
	if err != nil {
		return err
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(p.auth); err != nil {
		return err
	}

	// Set sender
	if err = client.Mail(p.config.FromAddress); err != nil {
		return err
	}

	// Set recipients
	for _, recipient := range p.config.ToAddresses {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}

	// Send the message
	writer, err := client.Data()
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write(message)
	return err
}