# Security Maintenance Guide for Snoozebot

This document outlines the process for maintaining security across the Snoozebot project, with a focus on dependency management and vulnerability scanning.

## Dependency Management

### Regular Updates

1. **Check for Updates**: Run the following command to check for outdated dependencies:
   ```bash
   go list -m -u all
   ```

2. **Update Dependencies**: Update to newer versions when available:
   ```bash
   go get -u [dependency]  # Update a specific dependency
   go get -u ./...         # Update all direct dependencies
   go mod tidy             # Clean up dependencies
   ```

3. **Testing**: After updates, run all tests to ensure compatibility:
   ```bash
   go test ./...
   ```

### Minimal Version Selection

Go uses minimal version selection (MVS) to determine which versions of dependencies to use. This can be beneficial for security, but also means you should:

- Explicitly update dependencies when security patches are released
- Periodically check for critical updates that MVS wouldn't automatically select

## Security Scanning

### Built-in Go Tools

1. **go mod verify**: Verifies that dependencies haven't been tampered with:
   ```bash
   go mod verify
   ```

2. **govulncheck**: Scans your code and dependencies for known vulnerabilities:
   ```bash
   go install golang.org/x/vuln/cmd/govulncheck@latest
   govulncheck ./...
   ```

### Third-Party Security Tools

1. **Nancy**: Scans dependencies against vulnerability databases:
   ```bash
   go install github.com/sonatype-nexus-community/nancy@latest
   go list -json -deps ./... | nancy sleuth
   ```

2. **gosec**: Static analysis tool for security issues in Go code:
   ```bash
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   gosec ./...
   ```

3. **trivy**: Comprehensive vulnerability scanner that can scan your Go modules:
   ```bash
   # Install trivy first (platform-specific)
   trivy fs --security-checks vuln ./
   ```

## Automated Security Checks

The `security_check.sh` script in the `scripts` directory automates these security checks:

```bash
./scripts/security_check.sh
```

Add this script to your CI/CD pipeline to ensure regular security scanning.

## Vendor Management

Consider vendoring dependencies for critical projects:

```bash
go mod vendor
```

This gives you more control over dependencies but requires more maintenance.

## Best Practices

1. **Dependency Pinning**: Pin dependencies to specific versions in critical systems.

2. **Indirect Dependencies**: Be aware of transitive dependencies that might introduce vulnerabilities.

3. **Go Proxy**: Use the Go module proxy (proxy.golang.org by default) which provides caching and validation.

4. **Private Modules**: For private modules, set up your own proxy server or explicitly configure GOPRIVATE.

5. **Supply Chain Security**: Consider tools like [SLSA](https://slsa.dev/) or [Sigstore](https://www.sigstore.dev/) for supply chain security.

## Response to Vulnerabilities

1. **Assessment**: Evaluate the severity and impact of the vulnerability.

2. **Mitigation**: Update the vulnerable dependency or implement a workaround.

3. **Notification**: If your project has users, notify them of critical vulnerabilities and fixes.

4. **Post-Mortem**: Document the vulnerability and response for future reference.

## Useful Resources

- [Go Security](https://go.dev/security/)
- [Go Vulnerability Database](https://pkg.go.dev/vuln/)
- [OWASP Go Security Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_Security_Cheatsheet.html)
- [GolangCI-Lint](https://golangci-lint.run/) - Includes security linters