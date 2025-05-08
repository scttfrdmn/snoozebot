package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/nxadm/tail"
	"github.com/scttfrdmn/snoozebot/pkg/plugin/security"
)

var (
	// Command-line flags
	watchMode    bool
	eventFile    string
	eventDir     string
	logFormat    string
	levelFilter  string
	typeFilter   string
	outputFormat string
	limit        int
	follow       bool
)

func init() {
	flag.BoolVar(&watchMode, "watch", false, "Watch for security events in real-time")
	flag.StringVar(&eventFile, "file", "", "Security event log file to process")
	flag.StringVar(&eventDir, "dir", "/var/log/snoozebot/security", "Security event directory")
	flag.StringVar(&logFormat, "format", "pretty", "Output format (pretty, json, csv)")
	flag.StringVar(&levelFilter, "level", "", "Filter by level (INFO, WARNING, ERROR, CRITICAL)")
	flag.StringVar(&typeFilter, "type", "", "Filter by event type (comma-separated)")
	flag.StringVar(&outputFormat, "output", "terminal", "Output format (terminal, file)")
	flag.IntVar(&limit, "limit", 100, "Limit the number of events to display")
	flag.BoolVar(&follow, "follow", false, "Follow the log file for new events")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Security event monitor and analyzer for Snoozebot\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  %s -watch            Watch for security events in real-time\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -file <file>      Process events from a specific file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -follow           Follow the latest log file for new events\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
}

func main() {
	// Parse command-line flags
	flag.Parse()
	
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "security-monitor",
		Level:  hclog.Info,
		Output: os.Stderr,
	})
	
	// Determine which log file to use
	var logPath string
	if eventFile != "" {
		// Use specified file
		logPath = eventFile
	} else {
		// Use latest log file
		logPath = filepath.Join(eventDir, "security-latest.log")
		
		// Check if the file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			logger.Error("Log file not found", "path", logPath)
			os.Exit(1)
		}
	}
	
	// Watch mode
	if watchMode {
		if err := monitorEvents(logger); err != nil {
			logger.Error("Failed to monitor events", "error", err)
			os.Exit(1)
		}
		return
	}
	
	// Follow mode
	if follow {
		if err := followEvents(logPath, logger); err != nil {
			logger.Error("Failed to follow events", "error", err)
			os.Exit(1)
		}
		return
	}
	
	// Process events from file
	if err := processEvents(logPath, logger); err != nil {
		logger.Error("Failed to process events", "error", err)
		os.Exit(1)
	}
}

// processEvents processes security events from a log file
func processEvents(logPath string, logger hclog.Logger) error {
	// Open log file
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()
	
	// Parse type filters
	typeFilters := make(map[string]bool)
	if typeFilter != "" {
		for _, t := range strings.Split(typeFilter, ",") {
			typeFilters[strings.TrimSpace(t)] = true
		}
	}
	
	// Process events
	decoder := json.NewDecoder(file)
	var events []*security.SecurityEvent
	
	for {
		var event security.SecurityEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			logger.Warn("Failed to decode event", "error", err)
			continue
		}
		
		// Apply filters
		if levelFilter != "" && event.Level != levelFilter {
			continue
		}
		
		if len(typeFilters) > 0 && !typeFilters[event.EventType] {
			continue
		}
		
		events = append(events, &event)
	}
	
	// Limit the number of events
	if limit > 0 && len(events) > limit {
		events = events[len(events)-limit:]
	}
	
	// Display events
	displayEvents(events, logger)
	
	return nil
}

// followEvents follows a log file for new events
func followEvents(logPath string, logger hclog.Logger) error {
	// Setup signal handler for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// Parse type filters
	typeFilters := make(map[string]bool)
	if typeFilter != "" {
		for _, t := range strings.Split(typeFilter, ",") {
			typeFilters[strings.TrimSpace(t)] = true
		}
	}
	
	// Tail the log file
	t, err := tail.TailFile(logPath, tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true, // Use polling instead of inotify
	})
	if err != nil {
		return fmt.Errorf("failed to tail log file: %w", err)
	}
	
	fmt.Println("Following security events. Press Ctrl+C to exit.")
	fmt.Println()
	
	// Process events
	for {
		select {
		case line := <-t.Lines:
			if line.Err != nil {
				logger.Warn("Error reading line", "error", line.Err)
				continue
			}
			
			// Parse event
			var event security.SecurityEvent
			if err := json.Unmarshal([]byte(line.Text), &event); err != nil {
				logger.Warn("Failed to parse event", "error", err)
				continue
			}
			
			// Apply filters
			if levelFilter != "" && event.Level != levelFilter {
				continue
			}
			
			if len(typeFilters) > 0 && !typeFilters[event.EventType] {
				continue
			}
			
			// Display event
			displayEvent(&event, logger)
			
		case <-sigCh:
			fmt.Println("Received signal, shutting down...")
			return nil
		}
	}
}

// monitorEvents monitors security events in real-time
func monitorEvents(logger hclog.Logger) error {
	// Setup signal handler for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// Create security event manager
	eventManager, err := security.NewSecurityEventManager(eventDir, logger)
	if err != nil {
		return fmt.Errorf("failed to create security event manager: %w", err)
	}
	
	// Parse type filters
	typeFilters := make(map[string]bool)
	if typeFilter != "" {
		for _, t := range strings.Split(typeFilter, ",") {
			typeFilters[strings.TrimSpace(t)] = true
		}
	}
	
	// Register callback for all event types
	eventManager.RegisterCallback("", func(event *security.SecurityEvent) {
		// Apply filters
		if levelFilter != "" && event.Level != levelFilter {
			return
		}
		
		if len(typeFilters) > 0 && !typeFilters[event.EventType] {
			return
		}
		
		// Display event
		displayEvent(event, logger)
	})
	
	fmt.Println("Monitoring security events. Press Ctrl+C to exit.")
	fmt.Println()
	
	// Generate test event
	startEvent := security.CreateEvent(
		security.EventAuditStarted,
		"Security monitoring started",
		"security-monitor",
		"audit",
	)
	eventManager.LogEvent(startEvent)
	
	// Wait for signal
	<-sigCh
	fmt.Println("Received signal, shutting down...")
	
	// Generate stop event
	stopEvent := security.CreateEvent(
		security.EventAuditCompleted,
		"Security monitoring stopped",
		"security-monitor",
		"audit",
	)
	eventManager.LogEvent(stopEvent)
	
	return nil
}

// displayEvents displays a list of security events
func displayEvents(events []*security.SecurityEvent, logger hclog.Logger) {
	if len(events) == 0 {
		fmt.Println("No events found")
		return
	}
	
	switch logFormat {
	case "json":
		// Display events as JSON
		for _, event := range events {
			data, err := json.MarshalIndent(event, "", "  ")
			if err != nil {
				logger.Error("Failed to marshal event", "error", err)
				continue
			}
			fmt.Println(string(data))
		}
		
	case "csv":
		// Display events as CSV
		fmt.Println("Timestamp,Level,Category,EventType,Message,Component,Success")
		for _, event := range events {
			fmt.Printf("%s,%s,%s,%s,%s,%s,%t\n",
				event.Timestamp.Format(time.RFC3339),
				event.Level,
				event.Category,
				event.EventType,
				strings.ReplaceAll(event.Message, ",", " "),
				event.Component,
				event.Success,
			)
		}
		
	default:
		// Display events in pretty format
		for _, event := range events {
			displayEvent(event, logger)
		}
	}
	
	fmt.Printf("\nTotal events: %d\n", len(events))
}

// displayEvent displays a single security event
func displayEvent(event *security.SecurityEvent, logger hclog.Logger) {
	switch logFormat {
	case "json":
		// Display event as JSON
		data, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			logger.Error("Failed to marshal event", "error", err)
			return
		}
		fmt.Println(string(data))
		
	case "csv":
		// Display event as CSV
		fmt.Printf("%s,%s,%s,%s,%s,%s,%t\n",
			event.Timestamp.Format(time.RFC3339),
			event.Level,
			event.Category,
			event.EventType,
			strings.ReplaceAll(event.Message, ",", " "),
			event.Component,
			event.Success,
		)
		
	default:
		// Display event in pretty format
		timestamp := event.Timestamp.Format("2006-01-02 15:04:05")
		
		// Choose color based on level
		levelColor := ""
		resetColor := ""
		if outputFormat == "terminal" {
			resetColor = "\033[0m"
			switch event.Level {
			case security.InfoLevel:
				levelColor = "\033[1;32m" // Green
			case security.WarningLevel:
				levelColor = "\033[1;33m" // Yellow
			case security.ErrorLevel:
				levelColor = "\033[1;31m" // Red
			case security.CriticalLevel:
				levelColor = "\033[1;37;41m" // White on red
			}
		}
		
		// Display event header
		fmt.Printf("%s[%s] %s%s %s\n",
			levelColor,
			event.Level,
			event.Message,
			resetColor,
			timestamp,
		)
		
		// Display event details
		fmt.Printf("  Type: %s, Category: %s, Component: %s, Success: %t\n",
			event.EventType,
			event.Category,
			event.Component,
			event.Success,
		)
		
		// Display optional fields
		if event.UserID != "" {
			fmt.Printf("  User: %s\n", event.UserID)
		}
		
		if event.PluginName != "" {
			fmt.Printf("  Plugin: %s\n", event.PluginName)
		}
		
		if event.IPAddress != "" {
			fmt.Printf("  IP: %s\n", event.IPAddress)
		}
		
		if len(event.Details) > 0 {
			fmt.Println("  Details:")
			for k, v := range event.Details {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
		
		fmt.Println()
	}
}