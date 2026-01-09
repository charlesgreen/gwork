// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockGCloudAPI is a mock implementation of GCloudAPI for testing.
type MockGCloudAPI struct {
	mock.Mock
}

// ListProjects mocks the ListProjects method.
func (m *MockGCloudAPI) ListProjects(ctx context.Context) ([]Project, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Project), args.Error(1)
}

// GetLastAuditLogEntry mocks the GetLastAuditLogEntry method.
func (m *MockGCloudAPI) GetLastAuditLogEntry(ctx context.Context, projectID string, freshnessDays int) (string, error) {
	args := m.Called(ctx, projectID, freshnessDays)
	return args.String(0), args.Error(1)
}
