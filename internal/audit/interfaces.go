// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"

	"github.com/charlesgreen/gwork/internal/drive"
)

// DriveClient defines the operations needed by the auditor.
// The drive.Client implements this interface.
type DriveClient interface {
	ListAllFiles(ctx context.Context) ([]drive.FileInfo, error)
	GetFilePermissions(ctx context.Context, fileID string) ([]drive.Permission, error)
	IsExternalShare(perm drive.Permission) bool
	Domain() string
}
