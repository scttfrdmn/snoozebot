package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scttfrdmn/snoozebot/pkg/monitor"
)

func main() {
	fmt.Println("Starting example host application with embedded Snoozebot monitor")

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and configure the monitor
	mon := monitor.NewMonitor().
		WithThreshold(monitor.CPU, 15.0).
		WithThreshold(monitor.Memory, 25.0).
		WithNapTime(5 * time.Minute). // Short naptime for demo purposes
		WithAgentURL("http://snooze-agent.example.com:8080").
		OnIdleStateChange(func(isIdle bool, duration time.Duration) {
			if isIdle {
				fmt.Printf("System is now idle (duration: %s)\n", duration)
			} else {
				fmt.Println("System is now active")
			}
		}).
		OnError(func(err error) {
			fmt.Printf("Monitor error: %v\n", err)
		})

	// Add a custom resource monitor
	mon.AddResourceMonitor("custom_app_metric", func() (float64, error) {
		// In a real application, this would measure something application-specific
		// For demo purposes, we'll alternate between high and low values
		now := time.Now()
		if now.Second()%30 < 15 {
			return 50.0, nil // High usage
		}
		return 5.0, nil // Low usage
	})

	// Start the monitor
	fmt.Println("Starting monitor...")
	if err := mon.Start(ctx); err != nil {
		fmt.Printf("Failed to start monitor: %v\n", err)
		os.Exit(1)
	}

	// Periodically print the current state
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				state := mon.GetCurrentState()
				fmt.Printf("Current state: idle=%v, duration=%v\n", state.IsIdle, state.IdleDuration)
				
				fmt.Println("Resource usage:")
				for resourceType, usage := range state.CurrentUsage {
					threshold, _ := mon.GetThreshold(resourceType)
					fmt.Printf("  %s: %.1f%% (threshold: %.1f%%)\n", resourceType, usage.Value, threshold)
				}
				fmt.Println()
			}
		}
	}()

	// This represents the main application logic
	fmt.Println("Host application is now running...")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	fmt.Println("Shutting down...")
	
	// Stop the monitor
	if err := mon.Stop(); err != nil {
		fmt.Printf("Error stopping monitor: %v\n", err)
	}
	
	fmt.Println("Monitor stopped")
}