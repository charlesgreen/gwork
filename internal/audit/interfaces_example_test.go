// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package audit_test

import (
	"context"
	"testing"

	"github.com/charlesgreen/gwork/internal/audit"
	"github.com/charlesgreen/gwork/internal/config"
	"github.com/charlesgreen/gwork/internal/drive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDriveClient is a mock implementation of DriveClient for testing.
type MockDriveClient struct {
	mock.Mock
}

func (m *MockDriveClient) ListAllFiles(ctx context.Context) ([]drive.FileInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]drive.FileInfo), args.Error(1)
}

func (m *MockDriveClient) GetFilePermissions(ctx context.Context, fileID string) ([]drive.Permission, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]drive.Permission), args.Error(1)
}

func (m *MockDriveClient) IsExternalShare(perm drive.Permission) bool {
	args := m.Called(perm)
	return args.Bool(0)
}

func (m *MockDriveClient) Domain() string {
	args := m.Called()
	return args.String(0)
}

// ExampleTestWithMockDriveClient demonstrates how to test Auditor with a mock.
func TestExampleWithMockDriveClient(t *testing.T) {
	// Create a mock DriveClient
	mockClient := new(MockDriveClient)

	// Set up expectations
	testFiles := []drive.FileInfo{
		{
			ID:           "file1",
			Name:         "Test File 1",
			MimeType:     "application/pdf",
			OwnerEmail:   "owner@example.com",
			CreatedTime:  "2025-01-01T00:00:00Z",
			ModifiedTime: "2025-01-02T00:00:00Z",
			Size:         1024,
		},
		{
			ID:           "file2",
			Name:         "Test File 2",
			MimeType:     "text/plain",
			OwnerEmail:   "owner@example.com",
			CreatedTime:  "2025-01-03T00:00:00Z",
			ModifiedTime: "2025-01-04T00:00:00Z",
			Size:         2048,
		},
	}

	mockClient.On("ListAllFiles", mock.Anything).Return(testFiles, nil)

	// Create a config
	cfg := &config.Config{
		Google: config.GoogleConfig{
			Domain: "example.com",
		},
	}

	// Create an auditor with the mock client
	auditor := audit.NewAuditorWithClient(cfg, mockClient)

	// Test AuditFiles
	ctx := context.Background()
	result, err := auditor.AuditFiles(ctx)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalFiles)
	assert.Len(t, result.FileRecords, 2)
	assert.Equal(t, "file1", result.FileRecords[0].FileID)
	assert.Equal(t, "Test File 1", result.FileRecords[0].FileName)
	assert.Equal(t, "owner@example.com", result.FileRecords[0].OwnerEmail)

	// Verify mock expectations were met
	mockClient.AssertExpectations(t)
}

// ExampleTestExternalSharingWithMock demonstrates testing external sharing audit.
func TestExampleExternalSharingWithMock(t *testing.T) {
	// Create a mock DriveClient
	mockClient := new(MockDriveClient)

	// Set up test data
	testFiles := []drive.FileInfo{
		{
			ID:         "file1",
			Name:       "Shared File",
			OwnerEmail: "owner@example.com",
		},
	}

	testPermissions := []drive.Permission{
		{
			ID:           "perm1",
			Type:         "user",
			Role:         "reader",
			EmailAddress: "external@otherdomain.com",
		},
	}

	// Set up expectations
	mockClient.On("ListAllFiles", mock.Anything).Return(testFiles, nil)
	mockClient.On("GetFilePermissions", mock.Anything, "file1").Return(testPermissions, nil)
	mockClient.On("IsExternalShare", testPermissions[0]).Return(true)

	// Create a config
	cfg := &config.Config{
		Google: config.GoogleConfig{
			Domain: "example.com",
		},
	}

	// Create an auditor with the mock client
	auditor := audit.NewAuditorWithClient(cfg, mockClient)

	// Test AuditExternalSharing
	ctx := context.Background()
	result, err := auditor.AuditExternalSharing(ctx)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.TotalFiles)
	assert.Equal(t, 1, result.FilesProcessed)
	assert.Equal(t, 1, result.TotalExternalShares)
	assert.Len(t, result.ExternalShares, 1)
	assert.Equal(t, "file1", result.ExternalShares[0].FileID)
	assert.Equal(t, "external@otherdomain.com", result.ExternalShares[0].SharedWithEmail)

	// Verify mock expectations were met
	mockClient.AssertExpectations(t)
}
