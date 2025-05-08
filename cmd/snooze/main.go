package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/scttfrdmn/snoozebot/pkg/core"
)

const (
	// Exit codes
	ExitSuccess = 0
	ExitError   = 1
)

func main() {
	// Define the command-line flags
	configFile := flag.String("config", "/etc/snoozebot/config.json", "Path to configuration file")
	flag.Parse()

	// Get the command and arguments
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(ExitError)
	}

	command := args[0]
	cmdArgs := args[1:]

	// Execute the appropriate command
	switch command {
	case "status":
		status(cmdArgs)
	case "config":
		if len(cmdArgs) == 0 {
			printConfigUsage()
			os.Exit(ExitError)
		}
		switch cmdArgs[0] {
		case "list":
			configList(*configFile)
		case "set":
			if len(cmdArgs) < 3 {
				fmt.Println("Error: 'set' command requires a key and value")
				printConfigUsage()
				os.Exit(ExitError)
			}
			configSet(*configFile, cmdArgs[1], cmdArgs[2])
		case "get":
			if len(cmdArgs) < 2 {
				fmt.Println("Error: 'get' command requires a key")
				printConfigUsage()
				os.Exit(ExitError)
			}
			configGet(*configFile, cmdArgs[1])
		default:
			fmt.Printf("Error: Unknown config command: %s\n", cmdArgs[0])
			printConfigUsage()
			os.Exit(ExitError)
		}
	case "start":
		startDaemon()
	case "stop":
		stopDaemon()
	case "restart":
		restartDaemon()
	case "history":
		history(cmdArgs)
	case "help":
		printUsage()
	default:
		fmt.Printf("Error: Unknown command: %s\n", command)
		printUsage()
		os.Exit(ExitError)
	}
}

func printUsage() {
	fmt.Println("Usage: snooze [--config=FILE] COMMAND [ARGS]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  status        Show current snooze status")
	fmt.Println("  config        Manage configuration")
	fmt.Println("  start         Start the snooze daemon")
	fmt.Println("  stop          Stop the snooze daemon")
	fmt.Println("  restart       Restart the snooze daemon")
	fmt.Println("  history       Show snooze history")
	fmt.Println("  help          Show this help message")
	fmt.Println("")
	fmt.Println("Run 'snooze COMMAND --help' for more information on a command.")
}

func printConfigUsage() {
	fmt.Println("Usage: snooze [--config=FILE] config COMMAND [ARGS]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list                      List all configuration options")
	fmt.Println("  get KEY                   Get the value of a configuration option")
	fmt.Println("  set KEY VALUE             Set the value of a configuration option")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  snooze config list")
	fmt.Println("  snooze config get cpu-threshold")
	fmt.Println("  snooze config set cpu-threshold 15.0")
	fmt.Println("  snooze config set naptime 45")
}

func status(args []string) {
	// In a real implementation, this would query the daemon status
	fmt.Println("Status: Running")
	fmt.Println("Uptime: 3 days, 2 hours, 15 minutes")
	fmt.Println("")
	fmt.Println("Resource Usage:")
	fmt.Println("  CPU: 5.2% (threshold: 10.0%)")
	fmt.Println("  Memory: 15.7% (threshold: 20.0%)")
	fmt.Println("  Network: 1.3% (threshold: 5.0%)")
	fmt.Println("  Disk: 2.4% (threshold: 5.0%)")
	fmt.Println("  User Input: No (last: 45 minutes ago)")
	fmt.Println("  GPU: 0.0% (threshold: 5.0%)")
	fmt.Println("")
	fmt.Println("Current State: Active (45 minutes until naptime)")
	fmt.Println("Cloud Provider: AWS")
	fmt.Println("Instance ID: i-1234567890abcdef0")
}

func configList(configFile string) {
	// In a real implementation, this would read the configuration file
	config := core.DefaultMonitorConfig()
	
	fmt.Println("Configuration:")
	fmt.Printf("  cpu-threshold: %.1f%%\n", config.Thresholds[core.CPU])
	fmt.Printf("  memory-threshold: %.1f%%\n", config.Thresholds[core.Memory])
	fmt.Printf("  network-threshold: %.1f%%\n", config.Thresholds[core.Network])
	fmt.Printf("  disk-threshold: %.1f%%\n", config.Thresholds[core.Disk])
	fmt.Printf("  user-input-threshold: %.1f%%\n", config.Thresholds[core.UserInput])
	fmt.Printf("  gpu-threshold: %.1f%%\n", config.Thresholds[core.GPU])
	fmt.Printf("  naptime: %d minutes\n", int(config.NapTime.Minutes()))
	fmt.Printf("  check-interval: %d seconds\n", int(config.CheckInterval.Seconds()))
}

func configGet(configFile, key string) {
	// In a real implementation, this would read the configuration file
	config := core.DefaultMonitorConfig()
	
	switch key {
	case "cpu-threshold":
		fmt.Printf("%.1f%%\n", config.Thresholds[core.CPU])
	case "memory-threshold":
		fmt.Printf("%.1f%%\n", config.Thresholds[core.Memory])
	case "network-threshold":
		fmt.Printf("%.1f%%\n", config.Thresholds[core.Network])
	case "disk-threshold":
		fmt.Printf("%.1f%%\n", config.Thresholds[core.Disk])
	case "user-input-threshold":
		fmt.Printf("%.1f%%\n", config.Thresholds[core.UserInput])
	case "gpu-threshold":
		fmt.Printf("%.1f%%\n", config.Thresholds[core.GPU])
	case "naptime":
		fmt.Printf("%d minutes\n", int(config.NapTime.Minutes()))
	case "check-interval":
		fmt.Printf("%d seconds\n", int(config.CheckInterval.Seconds()))
	default:
		fmt.Printf("Error: Unknown configuration key: %s\n", key)
		os.Exit(ExitError)
	}
}

func configSet(configFile, key, value string) {
	// In a real implementation, this would update the configuration file
	switch key {
	case "cpu-threshold", "memory-threshold", "network-threshold", "disk-threshold", "user-input-threshold", "gpu-threshold":
		threshold, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Printf("Error: Invalid value for %s: %s\n", key, value)
			os.Exit(ExitError)
		}
		fmt.Printf("Setting %s to %.1f%%\n", key, threshold)
	case "naptime":
		naptime, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("Error: Invalid value for naptime: %s\n", value)
			os.Exit(ExitError)
		}
		fmt.Printf("Setting naptime to %d minutes\n", naptime)
	case "check-interval":
		interval, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("Error: Invalid value for check-interval: %s\n", value)
			os.Exit(ExitError)
		}
		fmt.Printf("Setting check-interval to %d seconds\n", interval)
	default:
		fmt.Printf("Error: Unknown configuration key: %s\n", key)
		os.Exit(ExitError)
	}
}

func startDaemon() {
	// In a real implementation, this would start the daemon using systemd or another service manager
	fmt.Println("Starting snoozed daemon...")
	// systemctl start snoozed
	fmt.Println("Daemon started successfully")
}

func stopDaemon() {
	// In a real implementation, this would stop the daemon using systemd or another service manager
	fmt.Println("Stopping snoozed daemon...")
	// systemctl stop snoozed
	fmt.Println("Daemon stopped successfully")
}

func restartDaemon() {
	// In a real implementation, this would restart the daemon using systemd or another service manager
	fmt.Println("Restarting snoozed daemon...")
	// systemctl restart snoozed
	fmt.Println("Daemon restarted successfully")
}

func history(args []string) {
	// In a real implementation, this would read the daemon's history log
	fmt.Println("Snooze History:")
	fmt.Println("  2023-06-01 23:15:02 - Instance stopped after 1h30m of inactivity")
	fmt.Println("  2023-06-02 09:32:18 - Instance started by AWS API call")
	fmt.Println("  2023-06-02 18:45:33 - Instance stopped after 2h15m of inactivity")
	fmt.Println("  2023-06-03 10:12:05 - Instance started by AWS API call")
}