.PHONY: all daemon cli agent plugins plugin clean version bump-version test test-unit test-integration test-azure security install-tools

# Version information
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0")

# Build variables
BINARY_DIR = bin
PLUGINS_DIR = $(BINARY_DIR)/plugins
DAEMON_NAME = snoozed
CLI_NAME = snooze
AGENT_NAME = snooze-agent

# Go build flags
GO_BUILD = go build
GO_BUILD_FLAGS = -v

# Paths
DAEMON_SRC = ./cmd/snoozed
CLI_SRC = ./cmd/snooze
AGENT_SRC = ./agent/cmd
PLUGIN_SRC = ./plugins

all: daemon cli agent plugins

daemon:
	@echo "Building daemon..."
	@mkdir -p $(BINARY_DIR)
	$(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/$(DAEMON_NAME) $(DAEMON_SRC)

cli:
	@echo "Building CLI..."
	@mkdir -p $(BINARY_DIR)
	$(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/$(CLI_NAME) $(CLI_SRC)

agent:
	@echo "Building agent..."
	@mkdir -p $(BINARY_DIR)
	$(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/$(AGENT_NAME) $(AGENT_SRC)

plugins: $(PLUGINS_DIR)
	@echo "Building all plugins..."
	./scripts/build_plugins.sh

# Build a specific plugin
# Usage: make plugin PLUGIN=aws
plugin: $(PLUGINS_DIR)
	@if [ -z "$(PLUGIN)" ]; then \
		echo "Usage: make plugin PLUGIN=aws"; \
		exit 1; \
	fi
	@echo "Building plugin: $(PLUGIN)..."
	./scripts/build_plugins.sh $(PLUGIN)

$(PLUGINS_DIR):
	@mkdir -p $(PLUGINS_DIR)

clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)

install: all
	@echo "Installing..."
	@install -d /etc/snoozebot/plugins
	@install -m 755 $(BINARY_DIR)/$(DAEMON_NAME) /usr/local/bin/$(DAEMON_NAME)
	@install -m 755 $(BINARY_DIR)/$(CLI_NAME) /usr/local/bin/$(CLI_NAME)
	@install -m 755 $(BINARY_DIR)/$(AGENT_NAME) /usr/local/bin/$(AGENT_NAME)
	@install -m 755 $(PLUGINS_DIR)/aws /etc/snoozebot/plugins/aws
	@install -m 755 $(PLUGINS_DIR)/gcp /etc/snoozebot/plugins/gcp
	@install -m 755 $(PLUGINS_DIR)/azure /etc/snoozebot/plugins/azure

run-daemon: daemon
	@echo "Running daemon..."
	$(BINARY_DIR)/$(DAEMON_NAME) --plugins-dir=$(PLUGINS_DIR)

run-cli: cli
	@echo "Running CLI..."
	$(BINARY_DIR)/$(CLI_NAME) status

run-agent: agent plugins
	@echo "Running agent..."
	$(BINARY_DIR)/$(AGENT_NAME) --plugins-dir=$(PLUGINS_DIR)

# Print the current version
version:
	@echo "Snoozebot version $(VERSION)"

# Update the version to a new one
# Usage: make bump-version NEW_VERSION=1.0.0
bump-version:
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "Usage: make bump-version NEW_VERSION=1.0.0"; \
		exit 1; \
	fi
	@echo "Bumping version from $(VERSION) to $(NEW_VERSION)"
	@echo "$(NEW_VERSION)" > VERSION
	@sed -i '' 's/version", "$(VERSION)"/version", "$(NEW_VERSION)"/' cmd/snoozed/main.go || true
	@sed -i '' 's/return "$(VERSION)"/return "$(NEW_VERSION)"/' plugins/aws/main.go || true
	@sed -i '' 's/return "$(VERSION)"/return "$(NEW_VERSION)"/' plugins/gcp/main.go || true
	@sed -i '' 's/return "$(VERSION)"/return "$(NEW_VERSION)"/' plugins/azure/main.go || true
	@sed -i '' 's/Agent v$(VERSION)/Agent v$(NEW_VERSION)/' agent/cmd/main.go || true
	@echo "Version updated to $(NEW_VERSION)"

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./agent/... ./pkg/... 

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	SNOOZEBOT_RUN_INTEGRATION=true go test -v ./test/integration/...

# Run all tests
test: test-unit test-integration

# Run Azure plugin tests
test-azure:
	@echo "Running Azure plugin tests..."
	go test -v ./test/azure/...

# Install security tools
install-tools:
	@echo "Installing security tools..."
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/sonatype-nexus-community/nancy@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security checks
security:
	@echo "Running security checks..."
	./scripts/security_check.sh

# Run complete CI pipeline including tests and security checks
ci: test security
	@echo "CI pipeline completed successfully"