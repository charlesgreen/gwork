# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gwork is a Go CLI tool for auditing Google Workspace Drive files. It uses service account authentication with domain-wide delegation to scan files across an organization and generate reports on file ownership and external sharing.

## Build and Development Commands

```bash
make build          # Build the gwork binary
make test           # Run all tests with verbose output
make lint           # Run golangci-lint
make check          # Run fmt, vet, lint, and test
make ci             # CI checks (fmt-check, vet, lint, test) - no auto-formatting
make coverage       # Generate HTML coverage report
make deps           # Download and tidy dependencies
```

Run a single test:
```bash
go test -v -run TestFunctionName ./internal/package/...
```

## Architecture

### Package Structure

- **main.go**: CLI entry point using cobra. Defines commands: `audit files`, `audit sharing`, `audit all`, `config init`, `version`
- **internal/audit/**: Orchestrates audit operations. `Auditor` coordinates file discovery and external sharing analysis
- **internal/drive/**: Google Drive API client. Handles file listing and permission retrieval
- **internal/auth/**: Service account authentication with domain-wide delegation
- **internal/config/**: YAML configuration loading and validation
- **internal/reporter/**: CSV output generation for audit results
- **pkg/exitcode/**: Standardized exit codes (0=success, 1=config error, 2=auth error, 3=API error, 10=internal error)

### Dependency Injection Pattern

The codebase uses interfaces for testability:

- **drive.DriveAPI**: Abstracts Google Drive API calls (`ListFiles`, `ListPermissions`)
- **audit.DriveClient**: Abstracts drive client operations for the auditor

Production constructors use real implementations:
- `drive.NewClient()` - creates client with real Google API
- `audit.NewAuditor()` - creates auditor with production drive client

Test constructors accept mock implementations:
- `drive.NewClientWithAPI(mockAPI, ...)` - accepts mock DriveAPI
- `audit.NewAuditorWithClient(cfg, mockClient)` - accepts mock DriveClient

### Key Types

- `drive.FileInfo`: File metadata (ID, name, owner, type, timestamps, size)
- `drive.Permission`: File sharing permission details
- `audit.AuditResult`: Contains file records, external shares, counts, and errors
- `config.Config`: YAML configuration with google/audit/output sections

## Testing

Tests use testify for mocking. See `internal/audit/interfaces_example_test.go` for mock patterns.

The Makefile provides various test targets including race detection (`make test-race`) and coverage threshold checking (`make coverage-check` at 70% threshold).
