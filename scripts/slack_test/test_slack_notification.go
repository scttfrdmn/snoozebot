package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/notification"
	"github.com/scttfrdmn/snoozebot/pkg/notification/providers/slack"
)

func main() {
	// Parse command line flags
	webhookURL := flag.String("webhook", "", "Slack webhook URL")
	channel := flag.String("channel", "", "Slack channel (optional)")
	message := flag.String("message", "Test notification from Snoozebot", "Message to send")
	flag.Parse()

	if *webhookURL == "" {
		fmt.Println("Error: Webhook URL is required")
		fmt.Println("Usage: go run test_slack_notification.go -webhook=https://hooks.slack.com/services/YOUR/WEBHOOK/URL [-channel=#channel] [-message=\"Custom message\"]")
		os.Exit(1)
	}

	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "snoozebot-test",
		Output: os.Stdout,
		Level:  hclog.Info,
	})

	// Create notification manager
	manager := notification.NewManager(logger)

	// Create Slack provider
	slackProvider := slack.New(logger)
	err := manager.RegisterProvider(slackProvider)
	if err != nil {
		log.Fatalf("Failed to register Slack provider: %v", err)
	}

	// Initialize Slack provider
	config := map[string]interface{}{
		"webhook_url": *webhookURL,
		"username":    "Snoozebot Tester",
		"icon_emoji":  ":robot_face:",
	}

	// Add channel if provided
	if *channel != "" {
		config["channel"] = *channel
	}

	err = manager.InitProvider(slack.ProviderName, config)
	if err != nil {
		log.Fatalf("Failed to initialize Slack provider: %v", err)
	}

	// Create test notifications
	ctx := context.Background()

	// Test idle notification
	fmt.Println("Sending idle notification...")
	err = testIdleNotification(ctx, manager, *message)
	if err != nil {
		log.Fatalf("Failed to send idle notification: %v", err)
	}

	// Wait a bit between notifications
	time.Sleep(2 * time.Second)

	// Test scheduled action notification
	fmt.Println("Sending scheduled action notification...")
	err = testScheduledActionNotification(ctx, manager, *message)
	if err != nil {
		log.Fatalf("Failed to send scheduled action notification: %v", err)
	}

	// Wait a bit between notifications
	time.Sleep(2 * time.Second)

	// Test state change notification
	fmt.Println("Sending state change notification...")
	err = testStateChangeNotification(ctx, manager, *message)
	if err != nil {
		log.Fatalf("Failed to send state change notification: %v", err)
	}

	fmt.Println("All notifications sent successfully!")
}

func testIdleNotification(ctx context.Context, manager *notification.Manager, message string) error {
	errors := manager.NotifyIdle(
		ctx,
		"i-12345678",
		"test-instance",
		"aws",
		"us-west-2",
		30*time.Minute,
	)

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func testScheduledActionNotification(ctx context.Context, manager *notification.Manager, message string) error {
	errors := manager.NotifyScheduledAction(
		ctx,
		"i-12345678",
		"test-instance",
		"aws",
		"us-west-2",
		"stop",
		time.Now().Add(5*time.Minute),
		"Instance has been idle for 30 minutes",
	)

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func testStateChangeNotification(ctx context.Context, manager *notification.Manager, message string) error {
	errors := manager.NotifyStateChange(
		ctx,
		"i-12345678",
		"test-instance",
		"aws",
		"us-west-2",
		"running",
		"stopping",
		"Manual stop requested",
	)

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}