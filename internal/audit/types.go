// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

// Package audit provides audit functionality for Google Drive files.
package audit

import "time"

// FileRecord represents a file in the files-by-owner report.
type FileRecord struct {
	OwnerEmail   string
	FileID       string
	FileName     string
	FileType     string
	CreatedTime  time.Time
	ModifiedTime time.Time
	SizeBytes    int64
}

// ExternalShareRecord represents an external sharing entry.
type ExternalShareRecord struct {
	OwnerEmail       string
	FileID           string
	FileName         string
	SharedWithEmail  string
	SharedWithDomain string
	PermissionType   string
	PermissionRole   string
	SharedDate       time.Time // Note: Drive API doesn't provide this directly
}

// AuditResult contains the results of an audit operation.
type AuditResult struct {
	TotalFiles          int
	TotalExternalShares int
	FilesProcessed      int
	Errors              []error
	FileRecords         []FileRecord
	ExternalShares      []ExternalShareRecord
}
