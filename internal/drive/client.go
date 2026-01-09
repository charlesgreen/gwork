// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"google.golang.org/api/drive/v3"
)

// Client wraps the Google Drive API client.
type Client struct {
	api                 DriveAPI
	domain              string
	pageSize            int64
	includeSharedDrives bool
}

// NewClient creates a new Drive client with the real Google Drive service.
func NewClient(service *drive.Service, domain string, pageSize int64, includeSharedDrives bool) *Client {
	return &Client{
		api:                 NewGoogleDriveAPI(service),
		domain:              domain,
		pageSize:            pageSize,
		includeSharedDrives: includeSharedDrives,
	}
}

// NewClientWithAPI creates a new Drive client with a custom DriveAPI implementation.
// This is primarily used for testing.
func NewClientWithAPI(api DriveAPI, domain string, pageSize int64, includeSharedDrives bool) *Client {
	return &Client{
		api:                 api,
		domain:              domain,
		pageSize:            pageSize,
		includeSharedDrives: includeSharedDrives,
	}
}

// Domain returns the configured domain.
func (c *Client) Domain() string {
	return c.domain
}
