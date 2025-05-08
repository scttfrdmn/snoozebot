# Setting Up the Snoozebot Git Repository

This guide walks through the steps to initialize and set up the Git repository for the Snoozebot project.

## Initial Repository Setup

### 1. Initialize the Git Repository

Navigate to the project directory and initialize a new Git repository:

```bash
cd /Users/scttfrdmn/src/snoozebot
git init
```

### 2. Create a .gitignore File

Create a `.gitignore` file to exclude build artifacts, temporary files, and configuration files containing secrets:

```bash
cat > .gitignore << 'EOF'
# Binaries
/bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# OS-specific files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Editor directories and files
.idea/
.vscode/
*.swp
*.swo

# Build files
build/

# Configuration files that might contain secrets
config.local.json
*.env
EOF
```

### 3. Add Files to the Repository

Add all project files to the repository:

```bash
git add .
```

### 4. Make the Initial Commit

Create the initial commit:

```bash
git commit -m "Initial commit: Snoozebot core architecture and plugin system"
```

## Setting Up GitHub Repository

### 1. Create a New Repository on GitHub

Go to GitHub and create a new repository named "snoozebot".

### 2. Add the Remote Repository

Add the GitHub repository as a remote:

```bash
git remote add origin https://github.com/scottfridman/snoozebot.git
```

### 3. Push to GitHub

Push the local repository to GitHub:

```bash
git push -u origin master
```

## Branch Strategy

Consider using the following branch strategy:

- `main` (or `master`): Stable release branch
- `develop`: Development branch
- `feature/*`: Feature branches
- `bugfix/*`: Bug fix branches
- `release/*`: Release preparation branches

### Create the Development Branch

```bash
git checkout -b develop
git push -u origin develop
```

## Collaborative Development

### Working with Feature Branches

For new features, create a feature branch:

```bash
git checkout develop
git checkout -b feature/new-cloud-provider
```

Make changes, commit them, and push to GitHub:

```bash
git add .
git commit -m "Add support for new cloud provider"
git push -u origin feature/new-cloud-provider
```

Create a pull request on GitHub to merge the feature branch into the develop branch.

### Releasing

When ready to release:

1. Create a release branch:
   ```bash
   git checkout develop
   git checkout -b release/v1.0.0
   ```

2. Make any release-specific changes, update version numbers, etc.

3. Merge to master and tag the release:
   ```bash
   git checkout master
   git merge release/v1.0.0
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin master --tags
   ```

4. Merge release branch back to develop:
   ```bash
   git checkout develop
   git merge release/v1.0.0
   git push origin develop
   ```

## GitHub Actions

Consider setting up GitHub Actions for CI/CD. Create a `.github/workflows` directory and add workflow files for:

- Building and testing the code
- Linting and validation
- Creating release assets

Example workflow file for building and testing:

```yaml
name: Build and Test

on:
  push:
    branches: [ master, develop ]
  pull_request:
    branches: [ master, develop ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.18
    - name: Build
      run: make all
    - name: Test
      run: go test -v ./...
```

## GitHub Documentation

Update the GitHub repository with appropriate documentation:

1. Make sure README.md is comprehensive
2. Add CONTRIBUTING.md with contribution guidelines
3. Add SECURITY.md for security policy
4. Add LICENSE file with the MIT license

## Repository Maintenance

Regularly:

1. Update dependencies
2. Review and address issues
3. Merge bug fixes and features
4. Tag releases