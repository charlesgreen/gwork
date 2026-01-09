// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"context"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// DriveAPI abstracts Google Drive operations for testing.
type DriveAPI interface {
	ListFiles(ctx context.Context, opts *ListFilesOptions) (*ListFilesResult, error)
	ListPermissions(ctx context.Context, fileID string, opts *ListPermissionsOptions) (*ListPermissionsResult, error)
}

// ListFilesOptions contains options for listing files.
type ListFilesOptions struct {
	Corpora                   string
	PageSize                  int64
	PageToken                 string
	Fields                    string
	SupportsAllDrives         bool
	IncludeItemsFromAllDrives bool
}

// ListFilesResult contains the result of listing files.
type ListFilesResult struct {
	Files         []*drive.File
	NextPageToken string
}

// ListPermissionsOptions contains options for listing permissions.
type ListPermissionsOptions struct {
	Fields            string
	PageToken         string
	SupportsAllDrives bool
}

// ListPermissionsResult contains the result of listing permissions.
type ListPermissionsResult struct {
	Permissions   []*drive.Permission
	NextPageToken string
}

// GoogleDriveAPI implements DriveAPI using the real Google Drive service.
type GoogleDriveAPI struct {
	service *drive.Service
}

// NewGoogleDriveAPI creates a new GoogleDriveAPI instance.
func NewGoogleDriveAPI(service *drive.Service) *GoogleDriveAPI {
	return &GoogleDriveAPI{service: service}
}

// ListFiles lists files from Google Drive.
func (g *GoogleDriveAPI) ListFiles(ctx context.Context, opts *ListFilesOptions) (*ListFilesResult, error) {
	call := g.service.Files.List().
		Corpora(opts.Corpora).
		PageSize(opts.PageSize).
		Fields(googleapi.Field(opts.Fields)).
		SupportsAllDrives(opts.SupportsAllDrives).
		IncludeItemsFromAllDrives(opts.IncludeItemsFromAllDrives)

	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}

	result, err := call.Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return &ListFilesResult{
		Files:         result.Files,
		NextPageToken: result.NextPageToken,
	}, nil
}

// ListPermissions lists permissions for a file.
func (g *GoogleDriveAPI) ListPermissions(ctx context.Context, fileID string, opts *ListPermissionsOptions) (*ListPermissionsResult, error) {
	call := g.service.Permissions.List(fileID).
		Fields(googleapi.Field(opts.Fields)).
		SupportsAllDrives(opts.SupportsAllDrives)

	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}

	result, err := call.Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return &ListPermissionsResult{
		Permissions:   result.Permissions,
		NextPageToken: result.NextPageToken,
	}, nil
}
