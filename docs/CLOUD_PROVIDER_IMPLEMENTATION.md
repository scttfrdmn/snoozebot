# Cloud Provider Implementation Guidelines

This document outlines the guidelines for implementing new cloud providers in Snoozebot.

## Branching Strategy

Each cloud provider implementation should be developed on its own separate branch:

- AWS plugin: Implemented on the main branch as the first reference implementation
- GCP plugin: Implemented on branch `feature/gcp-provider`, now merged
- Azure plugin: Implemented on branch `feature/azure-provider`, now merged
- Other providers (e.g., DigitalOcean, Oracle Cloud): Each on its own feature branch `feature/{provider-name}`

## Implementation Workflow

1. Create a new feature branch from main:
   ```bash
   git checkout main
   git pull
   git checkout -b feature/{provider-name}-provider
   ```

2. Implement the provider following the pattern established in the AWS provider
   - Follow the interface defined in `/agent/provider/provider.go`
   - Use the cloud provider's official Go SDK
   - Implement all required methods
   - Handle errors appropriately

3. Write comprehensive tests for the new provider
   - Unit tests for all implemented methods
   - Mock the cloud provider's API responses
   - Test error conditions and edge cases

4. Document the provider-specific details
   - Add setup instructions
   - Document required credentials
   - Explain any provider-specific limitations

5. Create a pull request for review before merging to main
   - Ensure all tests pass
   - Address review comments

## Provider Implementation Requirements

Each cloud provider implementation must:

1. Implement the `CloudProvider` interface defined in `/agent/provider/provider.go`
2. Use the cloud provider's official SDK
3. Include proper error handling with detailed error messages
4. Include comprehensive logging
5. Follow the plugin architecture pattern established in the AWS provider
6. Include documentation on credential management
7. Provide proper cleanup in case of errors

## Testing Requirements

All cloud provider implementations must include:

1. Unit tests for all methods
2. Mocks for the cloud provider's API
3. Tests for error conditions and recovery
4. Tests for credential handling
5. Integration tests with real cloud resources (these should be clearly marked and optional)

## Documentation Requirements

Documentation for each provider should include:

1. Setup instructions
2. Credential management instructions
3. Example configurations
4. Any provider-specific limitations or features
5. Troubleshooting guidance

## Code Review Checklist

Before merging a provider implementation:

- [ ] All tests pass
- [ ] Code follows established patterns
- [ ] Documentation is complete
- [ ] Error handling is robust
- [ ] Logging is appropriate and useful
- [ ] Resource cleanup is properly implemented
- [ ] No sensitive information is hardcoded
- [ ] Dependencies are properly managed