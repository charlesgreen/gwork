// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"fmt"

	"github.com/charlesgreen/gwork/internal/drive"
)

// AuditExternalSharing performs an external sharing audit.
func (a *Auditor) AuditExternalSharing(ctx context.Context) (*AuditResult, error) {
	files, err := a.driveClient.ListAllFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	result := &AuditResult{
		TotalFiles:     len(files),
		ExternalShares: make([]ExternalShareRecord, 0),
		Errors:         make([]error, 0),
	}

	for _, file := range files {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		perms, err := a.driveClient.GetFilePermissions(ctx, file.ID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("file %s: %w", file.ID, err))
			continue
		}

		result.FilesProcessed++

		for _, perm := range perms {
			if a.driveClient.IsExternalShare(perm) {
				record := permissionToRecord(file, perm)
				result.ExternalShares = append(result.ExternalShares, record)
			}
		}
	}

	result.TotalExternalShares = len(result.ExternalShares)
	return result, nil
}

// permissionToRecord converts a file and permission to an ExternalShareRecord.
func permissionToRecord(file drive.FileInfo, perm drive.Permission) ExternalShareRecord {
	sharedWithDomain := perm.Domain
	if sharedWithDomain == "" && perm.EmailAddress != "" {
		sharedWithDomain = drive.ExtractDomain(perm.EmailAddress)
	}

	return ExternalShareRecord{
		OwnerEmail:       file.OwnerEmail,
		FileID:           file.ID,
		FileName:         file.Name,
		SharedWithEmail:  perm.EmailAddress,
		SharedWithDomain: sharedWithDomain,
		PermissionType:   perm.Type,
		PermissionRole:   perm.Role,
		// SharedDate is not available from Drive API
	}
}
