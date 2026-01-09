// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package gcp

import "context"

// GCloudAPI abstracts gcloud CLI operations for testing.
// This interface allows mocking gcloud commands in unit tests.
type GCloudAPI interface {
	// ListProjects returns all GCP projects accessible to the current user.
	ListProjects(ctx context.Context) ([]Project, error)

	// GetLastAuditLogEntry returns the timestamp of the most recent audit log entry
	// for the specified project, looking back up to freshnessDays.
	// Returns empty string if no logs found.
	GetLastAuditLogEntry(ctx context.Context, projectID string, freshnessDays int) (string, error)
}
