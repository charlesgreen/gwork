// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Client provides GCP project operations.
type Client struct {
	api GCloudAPI
}

// NewClient creates a new GCP client with the real gcloud CLI.
func NewClient() *Client {
	return &Client{
		api: NewGCloudCLI(),
	}
}

// NewClientWithAPI creates a new GCP client with a custom GCloudAPI implementation.
// This is primarily used for testing.
func NewClientWithAPI(api GCloudAPI) *Client {
	return &Client{
		api: api,
	}
}

// GCloudCLI implements GCloudAPI using the real gcloud CLI.
type GCloudCLI struct{}

// NewGCloudCLI creates a new GCloudCLI instance.
func NewGCloudCLI() *GCloudCLI {
	return &GCloudCLI{}
}

// ListProjects lists all GCP projects using gcloud.
func (g *GCloudCLI) ListProjects(ctx context.Context) ([]Project, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "projects", "list",
		"--format=csv[no-heading](projectId,name)")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gcloud projects list failed: %w", err)
	}

	var projects []Project
	seen := make(map[string]bool)

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse CSV properly - name might contain commas
		parts := strings.SplitN(line, ",", 2)
		if len(parts) < 2 {
			continue
		}

		id := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])

		// Deduplicate by project ID
		if seen[id] {
			continue
		}
		seen[id] = true

		projects = append(projects, Project{ID: id, Name: name})
	}

	return projects, nil
}

// GetLastAuditLogEntry queries audit logs for the most recent entry.
func (g *GCloudCLI) GetLastAuditLogEntry(ctx context.Context, projectID string, freshnessDays int) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gcloud", "logging", "read",
		"logName:cloudaudit.googleapis.com",
		"--project="+projectID,
		fmt.Sprintf("--freshness=%dd", freshnessDays),
		"--limit=1",
		"--format=value(timestamp)")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// CheckAllProjects checks activity for all projects concurrently.
// Returns project statuses and any errors encountered (errors don't stop processing).
func (c *Client) CheckAllProjects(ctx context.Context, opts CheckActivityOptions) ([]ProjectStatus, []error) {
	projects, err := c.api.ListProjects(ctx)
	if err != nil {
		return nil, []error{err}
	}

	return c.checkProjectsConcurrently(ctx, projects, opts)
}

// checkProjectsConcurrently checks multiple projects using a worker pool.
func (c *Client) checkProjectsConcurrently(ctx context.Context, projects []Project, opts CheckActivityOptions) ([]ProjectStatus, []error) {
	if len(projects) == 0 {
		return []ProjectStatus{}, nil
	}

	projectChan := make(chan Project, len(projects))
	resultChan := make(chan ProjectStatus, len(projects))
	errorChan := make(chan error, len(projects))

	var wg sync.WaitGroup

	// Start workers
	workers := opts.Workers
	if workers <= 0 {
		workers = 10
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for project := range projectChan {
				status, err := c.checkProject(ctx, project, opts)
				if err != nil {
					errorChan <- fmt.Errorf("project %s: %w", project.ID, err)
					// Still report the status with UNKNOWN
					status = ProjectStatus{
						ProjectID:       project.ID,
						ProjectName:     project.Name,
						LastActivity:    "Error checking",
						DaysSinceActive: -1,
						Status:          "UNKNOWN",
						ConsoleLink:     fmt.Sprintf("https://console.cloud.google.com/home/dashboard?project=%s", project.ID),
					}
				}
				resultChan <- status
			}
		}()
	}

	// Send projects to workers
	for _, p := range projects {
		projectChan <- p
	}
	close(projectChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	var results []ProjectStatus
	var errors []error

	for status := range resultChan {
		results = append(results, status)
	}

	for err := range errorChan {
		errors = append(errors, err)
	}

	return results, errors
}

// checkProject checks activity for a single project.
func (c *Client) checkProject(ctx context.Context, project Project, opts CheckActivityOptions) (ProjectStatus, error) {
	status := ProjectStatus{
		ProjectID:   project.ID,
		ProjectName: project.Name,
		ConsoleLink: fmt.Sprintf("https://console.cloud.google.com/home/dashboard?project=%s", project.ID),
	}

	freshnessDays := opts.FreshnessDays
	if freshnessDays <= 0 {
		freshnessDays = 400
	}

	lastLog, err := c.api.GetLastAuditLogEntry(ctx, project.ID, freshnessDays)

	if err != nil || lastLog == "" {
		status.LastActivity = "No activity found"
		status.DaysSinceActive = -1
		status.Status = "INACTIVE"
		return status, nil // Not an error - just no activity
	}

	// Parse timestamp
	t, parseErr := time.Parse(time.RFC3339, lastLog)
	if parseErr != nil {
		status.LastActivity = lastLog
		status.DaysSinceActive = -1
		status.Status = "UNKNOWN"
		return status, nil
	}

	status.LastActivity = t.Format("2006-01-02")
	status.DaysSinceActive = int(time.Since(t).Hours() / 24)

	inactiveDays := opts.InactiveDays
	if inactiveDays <= 0 {
		inactiveDays = 60
	}

	if status.DaysSinceActive > inactiveDays {
		status.Status = "INACTIVE"
	} else {
		status.Status = "ACTIVE"
	}

	return status, nil
}
