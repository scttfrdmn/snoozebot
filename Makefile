.PHONY: all daemon cli plugins clean

# Build variables
BINARY_DIR = bin
PLUGINS_DIR = $(BINARY_DIR)/plugins
DAEMON_NAME = snoozed
CLI_NAME = snooze

# Go build flags
GO_BUILD = go build
GO_BUILD_FLAGS = -v

# Paths
DAEMON_SRC = ./cmd/snoozed
CLI_SRC = ./cmd/snooze
PLUGIN_SRC = ./plugins

all: daemon cli plugins

daemon:
	@echo "Building daemon..."
	@mkdir -p $(BINARY_DIR)
	$(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/$(DAEMON_NAME) $(DAEMON_SRC)

cli:
	@echo "Building CLI..."
	@mkdir -p $(BINARY_DIR)
	$(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/$(CLI_NAME) $(CLI_SRC)

plugins:
	@echo "Building plugins..."
	@mkdir -p $(PLUGINS_DIR)
	$(GO_BUILD) $(GO_BUILD_FLAGS) -o $(PLUGINS_DIR)/aws $(PLUGIN_SRC)/aws

clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)

install: all
	@echo "Installing..."
	@install -d /etc/snoozebot/plugins
	@install -m 755 $(BINARY_DIR)/$(DAEMON_NAME) /usr/local/bin/$(DAEMON_NAME)
	@install -m 755 $(BINARY_DIR)/$(CLI_NAME) /usr/local/bin/$(CLI_NAME)
	@install -m 755 $(PLUGINS_DIR)/aws /etc/snoozebot/plugins/aws

run-daemon: daemon
	@echo "Running daemon..."
	$(BINARY_DIR)/$(DAEMON_NAME) --plugins-dir=$(PLUGINS_DIR)

run-cli: cli
	@echo "Running CLI..."
	$(BINARY_DIR)/$(CLI_NAME) status