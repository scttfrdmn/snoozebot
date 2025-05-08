package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/notification/types"
)

const (
	// ProviderName is the name of the slack provider
	ProviderName = "slack"

	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 10 * time.Second
)

// Config contains configuration for the Slack provider
type Config struct {
	// WebhookURL is the URL of the Slack webhook
	WebhookURL string `json:"webhook_url"`

	// Channel is the Slack channel to send notifications to (overrides webhook default)
	Channel string `json:"channel,omitempty"`

	// Username is the username to use for notifications (overrides webhook default)
	Username string `json:"username,omitempty"`

	// IconURL is the URL of the icon to use (overrides webhook default)
	IconURL string `json:"icon_url,omitempty"`

	// IconEmoji is the emoji to use as the icon (overrides webhook default)
	IconEmoji string `json:"icon_emoji,omitempty"`

	// Timeout is the timeout for HTTP requests
	Timeout time.Duration `json:"timeout,omitempty"`
}

// Message represents a Slack message
type Message struct {
	Channel     string        `json:"channel,omitempty"`
	Username    string        `json:"username,omitempty"`
	IconURL     string        `json:"icon_url,omitempty"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
	Text        string        `json:"text,omitempty"`
	Attachments []Attachment  `json:"attachments,omitempty"`
	Blocks      []interface{} `json:"blocks,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color      string            `json:"color,omitempty"`
	Title      string            `json:"title,omitempty"`
	TitleLink  string            `json:"title_link,omitempty"`
	Pretext    string            `json:"pretext,omitempty"`
	Text       string            `json:"text,omitempty"`
	Fallback   string            `json:"fallback,omitempty"`
	Fields     []AttachmentField `json:"fields,omitempty"`
	Footer     string            `json:"footer,omitempty"`
	FooterIcon string            `json:"footer_icon,omitempty"`
	TS         int64             `json:"ts,omitempty"`
}

// AttachmentField represents a field in a Slack attachment
type AttachmentField struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

// Provider implements types.NotificationProvider for Slack
type Provider struct {
	config        Config
	client        *http.Client
	logger        hclog.Logger
	initialized   bool
	severityColor map[types.Severity]string
}

// New creates a new Slack notification provider
func New(logger hclog.Logger) *Provider {
	return &Provider{
		logger: logger.Named("slack"),
		severityColor: map[types.Severity]string{
			types.SeverityInfo:     "#36a64f", // Green
			types.SeverityWarning:  "#ffcc00", // Yellow
			types.SeverityError:    "#ff3300", // Red
			types.SeverityCritical: "#ff0000", // Bright Red
		},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return ProviderName
}

// Init initializes the provider with the given configuration
func (p *Provider) Init(config map[string]interface{}) error {
	// Convert config to our Config struct
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var slackConfig Config
	if err := json.Unmarshal(configBytes, &slackConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if slackConfig.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	// Set defaults
	if slackConfig.Timeout == 0 {
		slackConfig.Timeout = DefaultTimeout
	}

	// Create HTTP client
	p.client = &http.Client{
		Timeout: slackConfig.Timeout,
	}

	p.config = slackConfig
	p.initialized = true
	p.logger.Info("Slack notification provider initialized")
	return nil
}

// Send sends a notification to Slack
func (p *Provider) Send(ctx context.Context, n *types.Notification) error {
	if !p.initialized {
		return fmt.Errorf("provider not initialized")
	}

	// Create message based on notification
	message := p.createMessage(n)

	// Marshal message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.config.WebhookURL,
		bytes.NewBuffer(messageJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	p.logger.Debug("Sent notification to Slack",
		"type", n.Type,
		"severity", n.Severity,
		"instance_id", n.InstanceID)

	return nil
}

// Close closes the provider and releases any resources
func (p *Provider) Close() error {
	p.initialized = false
	return nil
}

// createMessage creates a Slack message from a notification
func (p *Provider) createMessage(n *types.Notification) Message {
	// Basic message
	message := Message{
		Channel:   p.config.Channel,
		Username:  p.config.Username,
		IconURL:   p.config.IconURL,
		IconEmoji: p.config.IconEmoji,
	}

	// Set the color based on severity
	color := p.severityColor[n.Severity]
	if color == "" {
		color = p.severityColor[types.SeverityInfo]
	}

	// Create a formatted representation of the instance
	instanceText := n.InstanceID
	if n.InstanceName != "" {
		instanceText = fmt.Sprintf("%s (%s)", n.InstanceName, n.InstanceID)
	}

	// Create fields based on notification data
	fields := []AttachmentField{
		{
			Title: "Instance",
			Value: instanceText,
			Short: true,
		},
	}

	// Add provider and region if available
	if n.Provider != "" {
		fields = append(fields, AttachmentField{
			Title: "Provider",
			Value: n.Provider,
			Short: true,
		})
	}

	if n.Region != "" {
		fields = append(fields, AttachmentField{
			Title: "Region",
			Value: n.Region,
			Short: true,
		})
	}

	// Add additional data fields based on notification type
	switch n.Type {
	case types.NotificationTypeIdle:
		if idleDuration, ok := n.Data["idle_duration"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Idle Duration",
				Value: idleDuration,
				Short: true,
			})
		}

	case types.NotificationTypeScheduledAction:
		if action, ok := n.Data["action"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Action",
				Value: action,
				Short: true,
			})
		}

		if scheduledTime, ok := n.Data["scheduled_time"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Scheduled Time",
				Value: scheduledTime,
				Short: true,
			})
		}

		if reason, ok := n.Data["reason"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Reason",
				Value: reason,
				Short: false,
			})
		}

	case types.NotificationTypeActionExecuted:
		if action, ok := n.Data["action"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Action",
				Value: action,
				Short: true,
			})
		}

		if result, ok := n.Data["result"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Result",
				Value: result,
				Short: true,
			})
		}

	case types.NotificationTypeStateChange:
		if previousState, ok := n.Data["previous_state"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Previous State",
				Value: previousState,
				Short: true,
			})
		}

		if currentState, ok := n.Data["current_state"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Current State",
				Value: currentState,
				Short: true,
			})
		}

		if reason, ok := n.Data["reason"].(string); ok {
			fields = append(fields, AttachmentField{
				Title: "Reason",
				Value: reason,
				Short: false,
			})
		}
	}

	// Create attachment
	attachment := Attachment{
		Color:      color,
		Title:      n.Title,
		Text:       n.Message,
		Fallback:   fmt.Sprintf("%s: %s", n.Title, n.Message),
		Fields:     fields,
		Footer:     "Snoozebot",
		FooterIcon: "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a7/Alarm_Icon.svg/2048px-Alarm_Icon.svg.png",
		TS:         n.Timestamp.Unix(),
	}

	message.Attachments = []Attachment{attachment}
	return message
}