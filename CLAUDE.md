# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project named "portfolios". The codebase is in early stages of development.

## Development Setup

This project uses Go. Based on the .gitignore configuration:
- Coverage files are excluded (*.out, coverage.*, *.coverprofile, profile.cov)
- Go workspace files are excluded (go.work, go.work.sum)
- Environment files (.env) are excluded

## Common Commands

Once the project structure is established, typical Go commands will include:

**Build:**
```bash
go build ./...
```

**Run tests:**
```bash
go test ./...
```

**Run tests with coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
```

**Run a specific test:**
```bash
go test -run TestName ./path/to/package
```

**Format code:**
```bash
go fmt ./...
```

**Lint (requires golangci-lint):**
```bash
golangci-lint run
```

**Install dependencies:**
```bash
go mod download
go mod tidy
```

## Dependency Management

**CRITICAL: Always use the most up-to-date versions of third-party dependencies.**

When adding or updating dependencies:
1. **Search for the latest version**: Before using any third-party package, search the internet (pkg.go.dev, GitHub releases) to find the latest stable version
2. **Prefer stable releases**: Use the latest stable version, not pre-release versions (e.g., prefer v1.5.1 over v1.6.0-pre.2)
3. **Check regularly**: When working on existing code, verify that dependencies are current and update them if newer versions are available
4. **Update cautiously**: Review changelogs and breaking changes before upgrading major versions

**To update dependencies:**
```bash
# Update a specific dependency to latest
go get -u github.com/package/name@latest

# Update all dependencies
go get -u ./...

# Clean up and verify
go mod tidy
go mod verify
```

**Verify latest versions at:**
- https://pkg.go.dev/[package-path]
- https://github.com/[org]/[repo]/releases

## Architecture Notes

The project structure is not yet established. When developing:
- Follow standard Go project layout conventions
- Keep modules organized by domain or feature
- Place reusable packages in appropriately named directories
