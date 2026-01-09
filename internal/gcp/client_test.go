// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_CheckAllProjects_MixedStatuses(t *testing.T) {
	mockAPI := new(MockGCloudAPI)

	projects := []Project{
		{ID: "active-project", Name: "Active Project"},
		{ID: "inactive-project", Name: "Inactive Project"},
		{ID: "no-logs-project", Name: "No Logs Project"},
	}

	mockAPI.On("ListProjects", mock.Anything).Return(projects, nil)

	// Active project - recent activity
	activeTime := time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "active-project", 400).Return(activeTime, nil)

	// Inactive project - old activity
	inactiveTime := time.Now().Add(-90 * 24 * time.Hour).Format(time.RFC3339)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "inactive-project", 400).Return(inactiveTime, nil)

	// No logs project - empty response
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "no-logs-project", 400).Return("", nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, errs := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       2,
	})

	assert.Len(t, errs, 0)
	assert.Len(t, results, 3)

	// Count statuses
	statusCounts := make(map[string]int)
	for _, r := range results {
		statusCounts[r.Status]++
	}

	assert.Equal(t, 1, statusCounts["ACTIVE"])
	assert.Equal(t, 2, statusCounts["INACTIVE"])

	mockAPI.AssertExpectations(t)
}

func TestClient_CheckAllProjects_EmptyList(t *testing.T) {
	mockAPI := new(MockGCloudAPI)
	mockAPI.On("ListProjects", mock.Anything).Return([]Project{}, nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, errs := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       2,
	})

	assert.Len(t, errs, 0)
	assert.Len(t, results, 0)

	mockAPI.AssertExpectations(t)
}

func TestClient_CheckAllProjects_AllActive(t *testing.T) {
	mockAPI := new(MockGCloudAPI)

	projects := []Project{
		{ID: "proj1", Name: "Project 1"},
		{ID: "proj2", Name: "Project 2"},
	}

	mockAPI.On("ListProjects", mock.Anything).Return(projects, nil)

	recentTime := time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "proj1", 400).Return(recentTime, nil)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "proj2", 400).Return(recentTime, nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, errs := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       2,
	})

	assert.Len(t, errs, 0)
	assert.Len(t, results, 2)

	for _, r := range results {
		assert.Equal(t, "ACTIVE", r.Status)
	}

	mockAPI.AssertExpectations(t)
}

func TestClient_CheckAllProjects_UnparseableTimestamp(t *testing.T) {
	mockAPI := new(MockGCloudAPI)

	projects := []Project{
		{ID: "bad-timestamp", Name: "Bad Timestamp"},
	}

	mockAPI.On("ListProjects", mock.Anything).Return(projects, nil)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "bad-timestamp", 400).Return("not-a-valid-timestamp", nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, errs := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       1,
	})

	assert.Len(t, errs, 0)
	assert.Len(t, results, 1)
	assert.Equal(t, "UNKNOWN", results[0].Status)
	assert.Equal(t, "not-a-valid-timestamp", results[0].LastActivity)

	mockAPI.AssertExpectations(t)
}

func TestClient_CheckProject_DaysCalculation_ExactlyAtThreshold(t *testing.T) {
	mockAPI := new(MockGCloudAPI)

	projects := []Project{{ID: "test-project", Name: "Test"}}
	mockAPI.On("ListProjects", mock.Anything).Return(projects, nil)

	// Exactly 60 days ago
	timestamp := time.Now().Add(-60 * 24 * time.Hour).Format(time.RFC3339)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "test-project", 400).Return(timestamp, nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, _ := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       1,
	})

	assert.Len(t, results, 1)
	// At exactly 60 days, it should be INACTIVE (> not >=)
	// Actually checking if DaysSinceActive > inactiveDays, so 60 > 60 is false = ACTIVE
	assert.Equal(t, "ACTIVE", results[0].Status)
}

func TestClient_CheckProject_DaysCalculation_OneOverThreshold(t *testing.T) {
	mockAPI := new(MockGCloudAPI)

	projects := []Project{{ID: "test-project", Name: "Test"}}
	mockAPI.On("ListProjects", mock.Anything).Return(projects, nil)

	// 61 days ago
	timestamp := time.Now().Add(-61 * 24 * time.Hour).Format(time.RFC3339)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "test-project", 400).Return(timestamp, nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, _ := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       1,
	})

	assert.Len(t, results, 1)
	assert.Equal(t, "INACTIVE", results[0].Status)
}

func TestClient_CheckProject_DaysCalculation_WellUnderThreshold(t *testing.T) {
	mockAPI := new(MockGCloudAPI)

	projects := []Project{{ID: "test-project", Name: "Test"}}
	mockAPI.On("ListProjects", mock.Anything).Return(projects, nil)

	// 1 day ago
	timestamp := time.Now().Add(-1 * 24 * time.Hour).Format(time.RFC3339)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "test-project", 400).Return(timestamp, nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, _ := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       1,
	})

	assert.Len(t, results, 1)
	assert.Equal(t, "ACTIVE", results[0].Status)
}

func TestClient_ListProjects_Error(t *testing.T) {
	mockAPI := new(MockGCloudAPI)
	mockAPI.On("ListProjects", mock.Anything).Return(nil, errors.New("gcloud not found"))

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, errs := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       1,
	})

	assert.Nil(t, results)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "gcloud not found")

	mockAPI.AssertExpectations(t)
}

func TestProjectStatus_ConsoleLink(t *testing.T) {
	mockAPI := new(MockGCloudAPI)

	projects := []Project{{ID: "my-project-123", Name: "My Project"}}
	mockAPI.On("ListProjects", mock.Anything).Return(projects, nil)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "my-project-123", 400).Return("", nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	results, _ := client.CheckAllProjects(ctx, CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       1,
	})

	assert.Len(t, results, 1)
	assert.Equal(t, "https://console.cloud.google.com/home/dashboard?project=my-project-123",
		results[0].ConsoleLink)
}

func TestClient_CheckAllProjects_DefaultOptions(t *testing.T) {
	mockAPI := new(MockGCloudAPI)

	projects := []Project{{ID: "test", Name: "Test"}}
	mockAPI.On("ListProjects", mock.Anything).Return(projects, nil)
	mockAPI.On("GetLastAuditLogEntry", mock.Anything, "test", 400).Return("", nil)

	client := NewClientWithAPI(mockAPI)
	ctx := context.Background()

	// Use zero values - should use defaults
	results, errs := client.CheckAllProjects(ctx, CheckActivityOptions{})

	assert.Len(t, errs, 0)
	assert.Len(t, results, 1)
}

func TestDefaultCheckActivityOptions(t *testing.T) {
	opts := DefaultCheckActivityOptions()

	assert.Equal(t, 60, opts.InactiveDays)
	assert.Equal(t, 400, opts.FreshnessDays)
	assert.Equal(t, 10, opts.Workers)
}

func TestNewClient(t *testing.T) {
	// Test that NewClient creates a client with real GCloudCLI
	client := NewClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.api)
}

func TestNewClientWithAPI(t *testing.T) {
	mockAPI := new(MockGCloudAPI)
	client := NewClientWithAPI(mockAPI)

	assert.NotNil(t, client)
	assert.Equal(t, mockAPI, client.api)
}
