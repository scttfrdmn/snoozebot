# Snoozebot Documentation

## Overview

Snoozebot is a system for automatically managing cloud resources by monitoring system activity and hibernating idle instances to save costs. It consists of a monitoring daemon, a command-line interface, and a plugin system for cloud provider integration.

## Architecture

Snoozebot is built with a modular architecture:

1. **Core Daemon (`snoozed`)**: Monitors system resources and triggers hibernation when the system is idle.
2. **CLI Tool (`snooze`)**: Manages configuration and provides status information.
3. **Plugin System**: Extends functionality for different cloud providers.

### Plugin Architecture

Snoozebot uses HashiCorp's go-plugin library to implement a robust, process-isolated plugin system:

- **Process Isolation**: Each plugin runs as a separate process, providing stability and security.
- **gRPC Communication**: Plugins communicate with the core using gRPC.
- **Well-defined Interfaces**: Plugins implement standard interfaces defined by the core.
- **Version Compatibility**: The system checks plugin compatibility during loading.

## Components

### Core

The core components include:

- **Resource Monitor**: Tracks CPU, memory, network, disk, user input, and GPU activity.
- **Plugin Manager**: Handles loading, communication with, and management of plugins.
- **Configuration Manager**: Manages system configuration.

### Plugins

Snoozebot uses plugins to interact with different cloud providers:

- **AWS Plugin**: Manages AWS EC2 instances (stopping/starting).
- **GCP Plugin** (Planned): Will manage Google Cloud Platform instances.
- **Azure Plugin** (Planned): Will manage Microsoft Azure instances.

## Configuration

### Default Configuration

The default configuration values are:

| Resource | Threshold | Description |
| -------- | --------- | ----------- |
| CPU | 10.0% | Maximum CPU usage to consider system idle |
| Memory | 20.0% | Maximum memory usage to consider system idle |
| Network | 5.0% | Maximum network I/O to consider system idle |
| Disk | 5.0% | Maximum disk I/O to consider system idle |
| User Input | 0.0% | Any user input makes the system active |
| GPU | 5.0% | Maximum GPU usage to consider system idle |
| Naptime | 30 minutes | Duration system must be idle before stopping |
| Check Interval | 1 minute | How often resources are checked |

### Configuration File

The configuration is stored in `/etc/snoozebot/config.json` with the following structure:

```json
{
  "thresholds": {
    "cpu": 10.0,
    "memory": 20.0,
    "network": 5.0,
    "disk": 5.0,
    "user_input": 0.0,
    "gpu": 5.0
  },
  "naptime": 1800,
  "check_interval": 60
}
```

## Usage

### Installation

Install using the provided packages or build from source:

```bash
make
sudo make install
```

### Starting the Daemon

```bash
# Start with systemd
sudo systemctl start snoozed

# Or start manually
snoozed --plugins-dir=/etc/snoozebot/plugins
```

### CLI Commands

```bash
# Show current status
snooze status

# View configuration
snooze config list

# Change configuration
snooze config set cpu-threshold 15.0
snooze config set naptime 45

# View snooze history
snooze history

# Control the daemon
snooze start
snooze stop
snooze restart
```

## Developing Plugins

### Plugin Interface

Plugins must implement the `CloudProvider` interface:

```go
type CloudProvider interface {
    // GetInstanceInfo gets information about the current instance
    GetInstanceInfo(ctx context.Context) (*InstanceInfo, error)
    
    // StopInstance stops the current instance
    StopInstance(ctx context.Context) error
    
    // StartInstance starts the current instance
    StartInstance(ctx context.Context) error
    
    // GetProviderName returns the name of the cloud provider
    GetProviderName() string
    
    // GetProviderVersion returns the version of the cloud provider plugin
    GetProviderVersion() string
}
```

### Building a Plugin

1. Create a new directory in `plugins/` for your cloud provider.
2. Implement the `CloudProvider` interface.
3. Set up the plugin server using HashiCorp's go-plugin.
4. Build the plugin using `go build`.

See the AWS plugin in `plugins/aws/` for an example implementation.

## System Requirements

- Linux or MacOS operating system
- Go 1.18 or later (for building from source)
- Administrative privileges (for installing system-wide)

## Cloud Provider Requirements

### AWS

The instance requires an IAM role with the following permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ec2:StopInstances",
            "Resource": "arn:aws:ec2:*:*:instance/*",
            "Condition": {
                "StringEquals": {"ec2:ResourceID": "${ec2:InstanceID}"}
            }
        },
        {
            "Effect": "Allow",
            "Action": "ec2:DescribeInstances",
            "Resource": "*"
        }
    ]
}
```

## Troubleshooting

### Daemon Won't Start

1. Check permissions: `sudo systemctl status snoozed`
2. Verify plugin directory: `ls -la /etc/snoozebot/plugins`
3. Check logs: `journalctl -u snoozed`

### Plugin Errors

1. Ensure plugin compatibility with core version
2. Check plugin permissions: `ls -la /etc/snoozebot/plugins`
3. Verify cloud provider credentials

## License

Apache License 2.0