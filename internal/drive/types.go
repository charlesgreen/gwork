// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

// Package drive provides a client for Google Drive API operations.
package drive

// FileInfo represents relevant file metadata.
type FileInfo struct {
	ID           string
	Name         string
	MimeType     string
	OwnerEmail   string
	CreatedTime  string
	ModifiedTime string
	Size         int64
}

// Permission represents a file permission.
type Permission struct {
	ID           string
	Type         string // user, group, domain, anyone
	Role         string // owner, organizer, fileOrganizer, writer, commenter, reader
	EmailAddress string
	Domain       string
	DisplayName  string
}
