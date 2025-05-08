# Snoozebot 0.1.0 Release Plan

This document outlines the plan to prepare Snoozebot for a stable 0.1.0 release. It includes tasks for reverting from the current versioning, fixing compilation errors, completing the GCP plugin, and comprehensive testing.

## Progress Tracking

| Phase | Status | Start Date | Completion Date | Owner |
|-------|--------|------------|----------------|-------|
| 1. Reversion and Stabilization | In Progress | 2025-05-08 | | |
| 2. GCP Plugin Completion | In Progress | 2025-05-08 | | |
| 3. Testing Framework | Not Started | | | |
| 4. Live Provider Testing | Not Started | | | |
| 5. Documentation & Packaging | In Progress | 2025-05-08 | | |

## Detailed Task Breakdown

### Phase 1: Reversion and Stabilization (1 Week)

- [x] **Version Revert**
  - [x] Reset VERSION file to 0.1.0
  - [x] Update API version constants in code (CurrentAPIVersion, etc.)
  - [x] Update plugin manifests to reflect 0.1.0

- [x] **Fix Compilation Errors**
  - [x] Add missing imports in tls_plugin.go:
    - [x] net (already imported)
    - [x] strings (fixed in code)
    - [x] filepath (fixed in code)
    - [x] x509 (fixed in code)
  - [ ] Fix undefined references in plugin/auth/service.go:
    - [ ] UnimplementedPluginAuthServer
    - [ ] PluginAuthClient
    - [ ] AuthenticateRequest/Response
    - [ ] PermissionRequest/Response
  - [ ] Fix struct field issues with TLSConfig in SecureConfig
  - [ ] Resolve type conflicts in monitor code

### Phase 2: GCP Plugin Completion (2 Weeks)

- [x] **Core Functionality**
  - [x] Update GCP plugin to match AWS/Azure interfaces
  - [x] Implement GetAPIVersion method
  - [x] Implement ListInstances method 
  - [x] Add proper error handling and logging
  - [x] Integrate with versioning system
  - [x] Implement Shutdown method

- [x] **GCP Integration**
  - [x] Set up GCP authentication flow
  - [x] Implement VM operations:
    - [x] Start instances
    - [x] Stop instances
    - [x] Get instance info
  - [x] Add instance metadata handling
  - [x] Implement region/zone support

### Phase 3: Testing Framework (2 Weeks)

- [x] **Unit Testing**
  - [ ] Fix existing unit tests
  - [x] Add tests for versioning system:
    - [x] Version parsing
    - [x] Compatibility checking
    - [x] Manifest handling
  - [ ] Create tests for security features:
    - [ ] TLS communication
    - [ ] Signature verification
    - [ ] API key authentication

- [x] **Integration Testing**
  - [ ] Create tests for cross-plugin communication
  - [ ] Test plugin loading with security features
  - [x] Verify version compatibility checking
  - [x] Test plugin discovery and management
  - [x] Create mocks for cloud providers

### Phase 4: Live Provider Testing (3 Weeks)

- [x] **Testing Infrastructure**
  - [x] Set up credential management system
  - [x] Create profile-based authentication
  - [x] Develop test scripts for credentials
  - [x] Create environment setup utilities
  - [x] Document credential setup process
  
- [ ] **Cloud Provider Test Environments**
  - [ ] Set up test environment for AWS:
    - [ ] EC2 instances in test VPC
    - [ ] IAM roles with minimal permissions
  - [ ] Set up test environment for Azure:
    - [ ] VMs in isolated resource group
    - [ ] Service principals with restricted access
  - [ ] Set up test environment for GCP:
    - [ ] Compute Engine instances
    - [ ] Service accounts with proper scoping

- [ ] **Test Scenarios**
  - [ ] Test VM lifecycle management:
    - [ ] Listing instances
    - [ ] Starting instances
    - [ ] Stopping instances
  - [ ] Test error handling:
    - [ ] Invalid credentials
    - [ ] Network failures
    - [ ] Resource not found
  - [ ] Test authentication boundaries
  - [ ] Measure performance under various conditions

- [ ] **Security Validation**
  - [ ] Test plugin communication security
  - [ ] Validate signature verification process
  - [ ] Test TLS certificate validation
  - [ ] Verify authentication controls

### Phase 5: Documentation & Packaging (1 Week)

- [x] **Finalize Documentation**
  - [x] Update all version references to 0.1.0
  - [x] Create provider setup guides:
    - [x] AWS credential setup
    - [x] Azure credential setup
    - [x] GCP credential setup
  - [x] Add credential troubleshooting guide
  - [ ] Create general troubleshooting guide
  - [ ] Update examples to reflect current API

- [ ] **Release Preparation**
  - [ ] Create tagged release v0.1.0
  - [ ] Generate checksum files
  - [ ] Prepare detailed changelog
  - [ ] Create release package

## Prerequisites and Requirements

### Cloud Access Requirements

- AWS account with permissions to:
  - Create/destroy EC2 instances
  - Create VPCs and security groups
  - Generate IAM credentials

- Azure subscription with rights to:
  - Create/manage resource groups
  - Deploy VMs
  - Create service principals

- GCP project with:
  - Compute Engine API enabled
  - Permissions to create/destroy VMs
  - Service account creation capabilities

### Development Environment

- Go 1.18 or newer
- Protocol Buffers compiler
- Access to all three cloud CLIs:
  - AWS CLI
  - Azure CLI
  - Google Cloud SDK

### Testing Resources

- Minimal cloud resources:
  - 1-2 small instances per provider
  - Network configuration for secure testing
  - Temporary storage for test artifacts

## Resuming Work After Interruption

If work on this plan is interrupted, follow these steps to resume:

1. Check the progress tracking table at the top of this document
2. Run `make test` to verify current state
3. Run `./scripts/security_check.sh` to ensure no new vulnerabilities
4. Review git history since last active development
5. Update progress tracking table with current status
6. Continue with the earliest incomplete task

## Timeline Summary

- **Total Duration**: 9 weeks
- **Contingency Buffer**: 2 additional weeks recommended
- **Phase Dependencies**:
  - Phase 2 can start in parallel with Phase 1
  - Phase 3 requires completion of Phase 1 
  - Phase 4 requires completion of Phases 2 and 3
  - Phase 5 requires completion of Phase 4

## Contact Information

- **Project Lead**: (TBD)
- **Technical Contact**: (TBD)
- **Cloud Resources Contact**: (TBD)