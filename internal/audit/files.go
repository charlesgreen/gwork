// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/charlesgreen/gwork/internal/drive"
)

// AuditFiles performs a files-by-owner audit.
func (a *Auditor) AuditFiles(ctx context.Context) (*AuditResult, error) {
	files, err := a.driveClient.ListAllFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	result := &AuditResult{
		TotalFiles:     len(files),
		FileRecords:    make([]FileRecord, 0, len(files)),
		FilesProcessed: len(files),
	}

	for _, f := range files {
		record := fileInfoToRecord(f)
		result.FileRecords = append(result.FileRecords, record)
	}

	return result, nil
}

// fileInfoToRecord converts a drive.FileInfo to a FileRecord.
func fileInfoToRecord(f drive.FileInfo) FileRecord {
	createdTime, _ := time.Parse(time.RFC3339, f.CreatedTime)
	modifiedTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)

	return FileRecord{
		OwnerEmail:   f.OwnerEmail,
		FileID:       f.ID,
		FileName:     f.Name,
		FileType:     f.MimeType,
		CreatedTime:  createdTime,
		ModifiedTime: modifiedTime,
		SizeBytes:    f.Size,
	}
}
