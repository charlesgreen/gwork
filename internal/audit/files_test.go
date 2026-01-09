// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"testing"
	"time"

	"github.com/charlesgreen/gwork/internal/drive"
	"github.com/stretchr/testify/assert"
)

func TestFileInfoToRecord(t *testing.T) {
	tests := []struct {
		name     string
		fileInfo drive.FileInfo
		expected FileRecord
	}{
		{
			name: "complete file info",
			fileInfo: drive.FileInfo{
				ID:           "file123",
				Name:         "test.pdf",
				MimeType:     "application/pdf",
				OwnerEmail:   "owner@example.com",
				CreatedTime:  "2024-01-15T10:30:00Z",
				ModifiedTime: "2024-01-20T15:45:00Z",
				Size:         1024,
			},
			expected: FileRecord{
				OwnerEmail:   "owner@example.com",
				FileID:       "file123",
				FileName:     "test.pdf",
				FileType:     "application/pdf",
				CreatedTime:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ModifiedTime: time.Date(2024, 1, 20, 15, 45, 0, 0, time.UTC),
				SizeBytes:    1024,
			},
		},
		{
			name: "file with no owner",
			fileInfo: drive.FileInfo{
				ID:           "file456",
				Name:         "orphan.txt",
				MimeType:     "text/plain",
				OwnerEmail:   "",
				CreatedTime:  "2024-02-01T08:00:00Z",
				ModifiedTime: "2024-02-01T08:00:00Z",
				Size:         512,
			},
			expected: FileRecord{
				OwnerEmail:   "",
				FileID:       "file456",
				FileName:     "orphan.txt",
				FileType:     "text/plain",
				CreatedTime:  time.Date(2024, 2, 1, 8, 0, 0, 0, time.UTC),
				ModifiedTime: time.Date(2024, 2, 1, 8, 0, 0, 0, time.UTC),
				SizeBytes:    512,
			},
		},
		{
			name: "file with zero size",
			fileInfo: drive.FileInfo{
				ID:           "file789",
				Name:         "empty.txt",
				MimeType:     "text/plain",
				OwnerEmail:   "user@example.com",
				CreatedTime:  "2024-03-01T12:00:00Z",
				ModifiedTime: "2024-03-01T12:00:00Z",
				Size:         0,
			},
			expected: FileRecord{
				OwnerEmail:   "user@example.com",
				FileID:       "file789",
				FileName:     "empty.txt",
				FileType:     "text/plain",
				CreatedTime:  time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
				ModifiedTime: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
				SizeBytes:    0,
			},
		},
		{
			name: "Google Docs file",
			fileInfo: drive.FileInfo{
				ID:           "doc123",
				Name:         "presentation.pptx",
				MimeType:     "application/vnd.google-apps.presentation",
				OwnerEmail:   "presenter@example.com",
				CreatedTime:  "2024-04-10T09:15:00Z",
				ModifiedTime: "2024-04-15T14:30:00Z",
				Size:         2048,
			},
			expected: FileRecord{
				OwnerEmail:   "presenter@example.com",
				FileID:       "doc123",
				FileName:     "presentation.pptx",
				FileType:     "application/vnd.google-apps.presentation",
				CreatedTime:  time.Date(2024, 4, 10, 9, 15, 0, 0, time.UTC),
				ModifiedTime: time.Date(2024, 4, 15, 14, 30, 0, 0, time.UTC),
				SizeBytes:    2048,
			},
		},
		{
			name: "file with invalid timestamps",
			fileInfo: drive.FileInfo{
				ID:           "file999",
				Name:         "invalid.txt",
				MimeType:     "text/plain",
				OwnerEmail:   "user@example.com",
				CreatedTime:  "invalid-timestamp",
				ModifiedTime: "invalid-timestamp",
				Size:         100,
			},
			expected: FileRecord{
				OwnerEmail:   "user@example.com",
				FileID:       "file999",
				FileName:     "invalid.txt",
				FileType:     "text/plain",
				CreatedTime:  time.Time{}, // Zero time for invalid timestamp
				ModifiedTime: time.Time{}, // Zero time for invalid timestamp
				SizeBytes:    100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fileInfoToRecord(tt.fileInfo)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileInfoToRecord_TimestampParsing(t *testing.T) {
	// Test various timestamp formats
	tests := []struct {
		name             string
		createdTime      string
		modifiedTime     string
		expectValidTimes bool
	}{
		{
			name:             "valid RFC3339 timestamps",
			createdTime:      "2024-01-15T10:30:00Z",
			modifiedTime:     "2024-01-20T15:45:00Z",
			expectValidTimes: true,
		},
		{
			name:             "timestamps with timezone offset",
			createdTime:      "2024-01-15T10:30:00+05:00",
			modifiedTime:     "2024-01-20T15:45:00-08:00",
			expectValidTimes: true,
		},
		{
			name:             "empty timestamps",
			createdTime:      "",
			modifiedTime:     "",
			expectValidTimes: false,
		},
		{
			name:             "malformed timestamps",
			createdTime:      "not-a-timestamp",
			modifiedTime:     "also-not-a-timestamp",
			expectValidTimes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo := drive.FileInfo{
				ID:           "test-file",
				Name:         "test.txt",
				MimeType:     "text/plain",
				OwnerEmail:   "owner@example.com",
				CreatedTime:  tt.createdTime,
				ModifiedTime: tt.modifiedTime,
				Size:         100,
			}

			result := fileInfoToRecord(fileInfo)

			if tt.expectValidTimes {
				assert.False(t, result.CreatedTime.IsZero(), "CreatedTime should be parsed")
				assert.False(t, result.ModifiedTime.IsZero(), "ModifiedTime should be parsed")
			} else {
				assert.True(t, result.CreatedTime.IsZero(), "CreatedTime should be zero for invalid input")
				assert.True(t, result.ModifiedTime.IsZero(), "ModifiedTime should be zero for invalid input")
			}
		})
	}
}
