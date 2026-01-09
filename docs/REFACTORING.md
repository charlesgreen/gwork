# Refactoring for Testability

This document describes the refactoring work done to make the gwork project testable by introducing interfaces and dependency injection.

## Overview

The project previously had tight coupling to Google APIs and filesystem operations, making unit testing impossible. This refactoring introduces interfaces and dependency injection to enable testing with mocks.

## Changes Made

### 1. Drive Package (`internal/drive/`)

#### Created `interfaces.go`
- **`DriveAPI` interface**: Abstracts Google Drive API operations
  - `ListFiles(ctx, opts)`: Lists files from Google Drive
  - `ListPermissions(ctx, fileID, opts)`: Lists permissions for a file

- **`GoogleDriveAPI` struct**: Implements `DriveAPI` using the real Google Drive service
  - Used in production code
  - Wraps the actual `*drive.Service`

- **Option types**: `ListFilesOptions`, `ListFilesResult`, `ListPermissionsOptions`, `ListPermissionsResult`
  - Encapsulate request/response data for cleaner interfaces

#### Modified `client.go`
- Changed `Client.service` from `*drive.Service` to `DriveAPI` interface
- Added `NewClient()`: Creates client with real Google Drive service (production)
- Added `NewClientWithAPI()`: Creates client with custom `DriveAPI` implementation (testing)
- Maintains backward compatibility - existing code continues to work

#### Modified `files.go`
- Updated `ListAllFiles()` to use `DriveAPI` interface methods
- No business logic changes - same functionality

#### Modified `permissions.go`
- Updated `GetFilePermissions()` to use `DriveAPI` interface methods
- `IsExternalShare()` logic unchanged - same business rules

### 2. Audit Package (`internal/audit/`)

#### Created `interfaces.go`
- **`DriveClient` interface**: Defines operations needed by the auditor
  - `ListAllFiles(ctx)`: Returns all files
  - `GetFilePermissions(ctx, fileID)`: Returns permissions for a file
  - `IsExternalShare(perm)`: Checks if permission is external
  - `Domain()`: Returns the configured domain

- The `drive.Client` implements this interface naturally

#### Modified `audit.go`
- Changed `Auditor.driveClient` from `*drive.Client` to `DriveClient` interface
- Added `NewAuditor()`: Creates auditor with real drive client (production)
- Added `NewAuditorWithClient()`: Creates auditor with custom `DriveClient` (testing)
- No changes to `AuditFiles()` or `AuditExternalSharing()` business logic

### 3. Dependencies

#### Updated `go.mod`
- Added `github.com/stretchr/testify v1.9.0` for mocking support

## Testing Examples

### Testing Drive Package

```go
import (
    "github.com/leansecurity-co/gwork/internal/drive"
    "github.com/stretchr/testify/mock"
)

type MockDriveAPI struct {
    mock.Mock
}

func (m *MockDriveAPI) ListFiles(ctx context.Context, opts *drive.ListFilesOptions) (*drive.ListFilesResult, error) {
    args := m.Called(ctx, opts)
    return args.Get(0).(*drive.ListFilesResult), args.Error(1)
}

func TestWithMock(t *testing.T) {
    mockAPI := new(MockDriveAPI)
    mockAPI.On("ListFiles", mock.Anything, mock.Anything).Return(&drive.ListFilesResult{
        Files: []*v3.File{/* test data */},
    }, nil)

    client := drive.NewClientWithAPI(mockAPI, "example.com", 100, false)
    files, err := client.ListAllFiles(context.Background())

    // assertions...
    mockAPI.AssertExpectations(t)
}
```

### Testing Audit Package

```go
import (
    "github.com/leansecurity-co/gwork/internal/audit"
    "github.com/stretchr/testify/mock"
)

type MockDriveClient struct {
    mock.Mock
}

func (m *MockDriveClient) ListAllFiles(ctx context.Context) ([]drive.FileInfo, error) {
    args := m.Called(ctx)
    return args.Get(0).([]drive.FileInfo), args.Error(1)
}

func TestAuditWithMock(t *testing.T) {
    mockClient := new(MockDriveClient)
    mockClient.On("ListAllFiles", mock.Anything).Return([]drive.FileInfo{
        /* test data */
    }, nil)

    cfg := &config.Config{/* test config */}
    auditor := audit.NewAuditorWithClient(cfg, mockClient)
    result, err := auditor.AuditFiles(context.Background())

    // assertions...
    mockClient.AssertExpectations(t)
}
```

See `internal/audit/interfaces_example_test.go` for complete working examples.

## Backward Compatibility

All existing code continues to work without changes:
- `drive.NewClient()` still accepts `*drive.Service` and creates a production client
- `audit.NewAuditor()` still creates a production auditor with real Google APIs
- CLI commands and main.go require no modifications
- All business logic remains unchanged

## Pre-existing Issues

Note: The following pre-existing issues are unrelated to this refactoring:
- **Config package tests**: Type mismatches between `int` and `int64` in test assertions
- **Reporter package**: Missing error checks on deferred `file.Close()` calls

These existed before the refactoring and don't affect the functionality of the refactored code.

## Verification

All refactored packages pass their tests:
```bash
$ go test ./internal/drive/...
PASS

$ go test ./internal/audit/...
PASS

$ golangci-lint run ./internal/drive/... ./internal/audit/...
0 issues.

$ go build .
# Builds successfully
```

## GCP Package (`internal/gcp/`)

The GCP package was added to support auditing GCP project activity. It follows the same patterns established in the drive and audit packages.

### Created `interfaces.go`
- **`GCloudAPI` interface**: Abstracts gcloud CLI operations
  - `ListProjects(ctx)`: Lists all GCP projects
  - `GetLastAuditLogEntry(ctx, projectID, freshnessDays)`: Gets last audit log timestamp

- **`GCloudCLI` struct**: Implements `GCloudAPI` using the real gcloud CLI
  - Used in production code
  - Wraps `exec.Command` for gcloud operations

### Created `types.go`
- **`Project`**: Represents a GCP project (ID, Name)
- **`ProjectStatus`**: Project with activity status (ACTIVE/INACTIVE/UNKNOWN)
- **`CheckActivityOptions`**: Configuration for activity checking (threshold, workers, freshness)

### Created `client.go`
- **`Client`**: Main client wrapping the GCloudAPI interface
- `NewClient()`: Creates client with real gcloud CLI (production)
- `NewClientWithAPI()`: Creates client with custom `GCloudAPI` implementation (testing)
- `CheckAllProjects()`: Checks all projects concurrently using worker pool

### Testing Example

```go
import (
    "github.com/charlesgreen/gwork/internal/gcp"
    "github.com/stretchr/testify/mock"
)

type MockGCloudAPI struct {
    mock.Mock
}

func (m *MockGCloudAPI) ListProjects(ctx context.Context) ([]gcp.Project, error) {
    args := m.Called(ctx)
    return args.Get(0).([]gcp.Project), args.Error(1)
}

func TestWithMock(t *testing.T) {
    mockAPI := new(MockGCloudAPI)
    mockAPI.On("ListProjects", mock.Anything).Return([]gcp.Project{
        {ID: "test-project", Name: "Test"},
    }, nil)

    client := gcp.NewClientWithAPI(mockAPI)
    results, errs := client.CheckAllProjects(context.Background(), gcp.DefaultCheckActivityOptions())

    // assertions...
    mockAPI.AssertExpectations(t)
}
```

## Next Steps

The refactoring is complete for the core drive, audit, and gcp packages. Future work could include:
- Create more comprehensive unit tests using the new interfaces
- Add integration tests that use real Google API responses
- Consider refactoring reporter package for testability
- Consider refactoring config package if filesystem testing is needed
