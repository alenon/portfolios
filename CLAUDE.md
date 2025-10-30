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

## Architecture Notes

The project structure is not yet established. When developing:
- Follow standard Go project layout conventions
- Keep modules organized by domain or feature
- Place reusable packages in appropriately named directories
