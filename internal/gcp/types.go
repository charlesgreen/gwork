// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

// Package gcp provides GCP project operations for the gwork CLI.
package gcp

// Project represents a GCP project.
type Project struct {
	ID   string
	Name string
}

// ProjectStatus represents a project with its activity status.
type ProjectStatus struct {
	ProjectID       string
	ProjectName     string
	LastActivity    string // Date string (YYYY-MM-DD) or "No activity found"
	DaysSinceActive int    // -1 if unknown/N/A
	Status          string // "ACTIVE", "INACTIVE", "UNKNOWN"
	ConsoleLink     string
}

// CheckActivityOptions contains options for checking project activity.
type CheckActivityOptions struct {
	// InactiveDays is the threshold for marking a project as inactive.
	// Projects with no activity in this many days are marked INACTIVE.
	// Default: 60
	InactiveDays int

	// FreshnessDays is how far back to look in audit logs.
	// Default: 400
	FreshnessDays int

	// Workers is the number of concurrent workers for checking projects.
	// Default: 10
	Workers int
}

// DefaultCheckActivityOptions returns options with sensible defaults.
func DefaultCheckActivityOptions() CheckActivityOptions {
	return CheckActivityOptions{
		InactiveDays:  60,
		FreshnessDays: 400,
		Workers:       10,
	}
}
