// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"testing"

	"github.com/charlesgreen/gwork/internal/drive"
	"github.com/stretchr/testify/assert"
)

func TestPermissionToRecord(t *testing.T) {
	tests := []struct {
		name       string
		file       drive.FileInfo
		permission drive.Permission
		expected   ExternalShareRecord
	}{
		{
			name: "permission with email address",
			file: drive.FileInfo{
				ID:         "file123",
				Name:       "document.pdf",
				OwnerEmail: "owner@example.com",
			},
			permission: drive.Permission{
				Type:         "user",
				Role:         "reader",
				EmailAddress: "external@other.com",
				Domain:       "",
			},
			expected: ExternalShareRecord{
				OwnerEmail:       "owner@example.com",
				FileID:           "file123",
				FileName:         "document.pdf",
				SharedWithEmail:  "external@other.com",
				SharedWithDomain: "other.com",
				PermissionType:   "user",
				PermissionRole:   "reader",
			},
		},
		{
			name: "permission with domain",
			file: drive.FileInfo{
				ID:         "file456",
				Name:       "spreadsheet.xlsx",
				OwnerEmail: "owner@example.com",
			},
			permission: drive.Permission{
				Type:         "domain",
				Role:         "writer",
				EmailAddress: "",
				Domain:       "external.org",
			},
			expected: ExternalShareRecord{
				OwnerEmail:       "owner@example.com",
				FileID:           "file456",
				FileName:         "spreadsheet.xlsx",
				SharedWithEmail:  "",
				SharedWithDomain: "external.org",
				PermissionType:   "domain",
				PermissionRole:   "writer",
			},
		},
		{
			name: "anyone permission",
			file: drive.FileInfo{
				ID:         "file789",
				Name:       "public.txt",
				OwnerEmail: "owner@example.com",
			},
			permission: drive.Permission{
				Type:         "anyone",
				Role:         "reader",
				EmailAddress: "",
				Domain:       "",
			},
			expected: ExternalShareRecord{
				OwnerEmail:       "owner@example.com",
				FileID:           "file789",
				FileName:         "public.txt",
				SharedWithEmail:  "",
				SharedWithDomain: "",
				PermissionType:   "anyone",
				PermissionRole:   "reader",
			},
		},
		{
			name: "group permission with email",
			file: drive.FileInfo{
				ID:         "fileABC",
				Name:       "team-doc.docx",
				OwnerEmail: "owner@example.com",
			},
			permission: drive.Permission{
				Type:         "group",
				Role:         "commenter",
				EmailAddress: "group@external.com",
				Domain:       "",
			},
			expected: ExternalShareRecord{
				OwnerEmail:       "owner@example.com",
				FileID:           "fileABC",
				FileName:         "team-doc.docx",
				SharedWithEmail:  "group@external.com",
				SharedWithDomain: "external.com",
				PermissionType:   "group",
				PermissionRole:   "commenter",
			},
		},
		{
			name: "permission with both domain and email",
			file: drive.FileInfo{
				ID:         "fileDEF",
				Name:       "shared.pptx",
				OwnerEmail: "owner@example.com",
			},
			permission: drive.Permission{
				Type:         "user",
				Role:         "writer",
				EmailAddress: "user@domain.com",
				Domain:       "domain.com",
			},
			expected: ExternalShareRecord{
				OwnerEmail:       "owner@example.com",
				FileID:           "fileDEF",
				FileName:         "shared.pptx",
				SharedWithEmail:  "user@domain.com",
				SharedWithDomain: "domain.com",
				PermissionType:   "user",
				PermissionRole:   "writer",
			},
		},
		{
			name: "file with no owner",
			file: drive.FileInfo{
				ID:         "orphan123",
				Name:       "orphan.txt",
				OwnerEmail: "",
			},
			permission: drive.Permission{
				Type:         "user",
				Role:         "reader",
				EmailAddress: "someone@other.com",
				Domain:       "",
			},
			expected: ExternalShareRecord{
				OwnerEmail:       "",
				FileID:           "orphan123",
				FileName:         "orphan.txt",
				SharedWithEmail:  "someone@other.com",
				SharedWithDomain: "other.com",
				PermissionType:   "user",
				PermissionRole:   "reader",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := permissionToRecord(tt.file, tt.permission)
			assert.Equal(t, tt.expected.OwnerEmail, result.OwnerEmail)
			assert.Equal(t, tt.expected.FileID, result.FileID)
			assert.Equal(t, tt.expected.FileName, result.FileName)
			assert.Equal(t, tt.expected.SharedWithEmail, result.SharedWithEmail)
			assert.Equal(t, tt.expected.SharedWithDomain, result.SharedWithDomain)
			assert.Equal(t, tt.expected.PermissionType, result.PermissionType)
			assert.Equal(t, tt.expected.PermissionRole, result.PermissionRole)
		})
	}
}

func TestExtractDomainFromEmail(t *testing.T) {
	// This test verifies the drive.ExtractDomain function which is used by
	// permissionToRecord to extract domain from email addresses.
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "standard email",
			email:    "user@example.com",
			expected: "example.com",
		},
		{
			name:     "email with subdomain",
			email:    "user@mail.example.com",
			expected: "mail.example.com",
		},
		{
			name:     "email with multiple @ symbols",
			email:    "user@name@example.com",
			expected: "example.com",
		},
		{
			name:     "email without @ symbol",
			email:    "notanemail",
			expected: "",
		},
		{
			name:     "empty email",
			email:    "",
			expected: "",
		},
		{
			name:     "email with @ at end",
			email:    "user@",
			expected: "",
		},
		{
			name:     "email with @ at start",
			email:    "@example.com",
			expected: "example.com",
		},
		{
			name:     "complex email with plus addressing",
			email:    "user+tag@example.com",
			expected: "example.com",
		},
		{
			name:     "email with country code domain",
			email:    "user@example.co.uk",
			expected: "example.co.uk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := drive.ExtractDomain(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuditResult_Structure(t *testing.T) {
	// Test that AuditResult can be created and has expected fields
	result := &AuditResult{
		TotalFiles:          100,
		TotalExternalShares: 25,
		FilesProcessed:      100,
		Errors:              []error{},
		FileRecords:         []FileRecord{},
		ExternalShares:      []ExternalShareRecord{},
	}

	assert.Equal(t, 100, result.TotalFiles)
	assert.Equal(t, 25, result.TotalExternalShares)
	assert.Equal(t, 100, result.FilesProcessed)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.FileRecords)
	assert.Empty(t, result.ExternalShares)
}

func TestFileRecord_Structure(t *testing.T) {
	// Test that FileRecord can be created with all fields
	record := FileRecord{
		OwnerEmail: "owner@example.com",
		FileID:     "file123",
		FileName:   "test.pdf",
		FileType:   "application/pdf",
		SizeBytes:  1024,
	}

	assert.Equal(t, "owner@example.com", record.OwnerEmail)
	assert.Equal(t, "file123", record.FileID)
	assert.Equal(t, "test.pdf", record.FileName)
	assert.Equal(t, "application/pdf", record.FileType)
	assert.Equal(t, int64(1024), record.SizeBytes)
}

func TestExternalShareRecord_Structure(t *testing.T) {
	// Test that ExternalShareRecord can be created with all fields
	record := ExternalShareRecord{
		OwnerEmail:       "owner@example.com",
		FileID:           "file123",
		FileName:         "test.pdf",
		SharedWithEmail:  "external@other.com",
		SharedWithDomain: "other.com",
		PermissionType:   "user",
		PermissionRole:   "reader",
	}

	assert.Equal(t, "owner@example.com", record.OwnerEmail)
	assert.Equal(t, "file123", record.FileID)
	assert.Equal(t, "test.pdf", record.FileName)
	assert.Equal(t, "external@other.com", record.SharedWithEmail)
	assert.Equal(t, "other.com", record.SharedWithDomain)
	assert.Equal(t, "user", record.PermissionType)
	assert.Equal(t, "reader", record.PermissionRole)
}
