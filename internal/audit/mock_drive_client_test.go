// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"

	"github.com/charlesgreen/gwork/internal/drive"
	"github.com/stretchr/testify/mock"
)

// MockDriveClient is a mock implementation of the drive client interface.
type MockDriveClient struct {
	mock.Mock
}

// ListAllFiles mocks the ListAllFiles method.
func (m *MockDriveClient) ListAllFiles(ctx context.Context) ([]drive.FileInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]drive.FileInfo), args.Error(1)
}

// GetFilePermissions mocks the GetFilePermissions method.
func (m *MockDriveClient) GetFilePermissions(ctx context.Context, fileID string) ([]drive.Permission, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]drive.Permission), args.Error(1)
}

// IsExternalShare mocks the IsExternalShare method.
func (m *MockDriveClient) IsExternalShare(perm drive.Permission) bool {
	args := m.Called(perm)
	return args.Bool(0)
}

// Domain mocks the Domain method.
func (m *MockDriveClient) Domain() string {
	args := m.Called()
	return args.String(0)
}
