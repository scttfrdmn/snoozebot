package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/notification/providers/email"
	"github.com/scttfrdmn/snoozebot/pkg/notification/types"
)

func main() {
	// Parse command line flags
	smtpServer := flag.String("smtp-server", "", "SMTP server address")
	smtpPort := flag.Int("smtp-port", 587, "SMTP server port")
	username := flag.String("username", "", "SMTP username")
	password := flag.String("password", "", "SMTP password")
	fromAddress := flag.String("from", "", "From email address")
	toAddress := flag.String("to", "", "To email address")
	enableSSL := flag.Bool("ssl", false, "Use SSL instead of STARTTLS")
	skipVerify := flag.Bool("skip-verify", false, "Skip TLS certificate verification")
	flag.Parse()

	// Validate required parameters
	if *smtpServer == "" || *username == "" || *password == "" || *fromAddress == "" || *toAddress == "" {
		fmt.Println("Error: SMTP server, username, password, from address, and to address are required")
		fmt.Println("Usage: go run test_email_notification.go -smtp-server=smtp.example.com -username=user@example.com -password=yourpassword -from=\"Snoozebot <user@example.com>\" -to=recipient@example.com")
		os.Exit(1)
	}

	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "snoozebot-test",
		Output: os.Stdout,
		Level:  hclog.Info,
	})

	// Create email provider
	emailProvider := email.New(logger)

	// Initialize email provider with configuration
	config := map[string]interface{}{
		"smtp_server":     *smtpServer,
		"smtp_port":       float64(*smtpPort),
		"username":        *username,
		"password":        *password,
		"from_address":    *fromAddress,
		"to_addresses":    []interface{}{*toAddress},
		"enable_ssl":      *enableSSL,
		"enable_starttls": !*enableSSL,
		"skip_tls_verify": *skipVerify,
		"subject_prefix":  "[Snoozebot Test]",
	}

	err := emailProvider.Init(config)
	if err != nil {
		log.Fatalf("Failed to initialize email provider: %v", err)
	}

	// Create test notification
	notification := &types.Notification{
		Type:         types.NotificationTypeIdle,
		Severity:     types.SeverityInfo,
		InstanceID:   "i-12345678",
		InstanceName: "test-instance",
		Provider:     "aws",
		Region:       "us-west-2",
		Title:        "Test Email Notification",
		Message:      "This is a test email notification from Snoozebot.",
		Timestamp:    time.Now(),
		Data: map[string]interface{}{
			"idle_duration": "30m",
		},
	}

	// Send notification
	fmt.Println("Sending test email notification...")
	err = emailProvider.Send(context.Background(), notification)
	if err != nil {
		log.Fatalf("Failed to send email notification: %v", err)
	}

	fmt.Println("Test email notification sent successfully!")
}