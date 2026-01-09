// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"fmt"

	"github.com/charlesgreen/gwork/internal/auth"
	"github.com/charlesgreen/gwork/internal/config"
	"github.com/charlesgreen/gwork/internal/drive"
)

// Auditor orchestrates audit operations.
type Auditor struct {
	config      *config.Config
	driveClient DriveClient
}

// NewAuditor creates a new Auditor instance with the production drive client.
func NewAuditor(cfg *config.Config) (*Auditor, error) {
	authenticator, err := auth.NewAuthenticator(
		cfg.Google.ServiceAccountFile,
		cfg.Google.AdminEmail,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	ctx := context.Background()
	driveService, err := authenticator.GetDriveService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	driveClient := drive.NewClient(
		driveService,
		cfg.Google.Domain,
		cfg.Audit.PageSize,
		cfg.Audit.IncludeSharedDrives,
	)

	return &Auditor{
		config:      cfg,
		driveClient: driveClient,
	}, nil
}

// NewAuditorWithClient creates a new Auditor instance with a custom DriveClient.
// This is primarily used for testing.
func NewAuditorWithClient(cfg *config.Config, client DriveClient) *Auditor {
	return &Auditor{
		config:      cfg,
		driveClient: client,
	}
}

// AuditAll performs all audit operations.
func (a *Auditor) AuditAll(ctx context.Context) (*AuditResult, *AuditResult, error) {
	filesResult, err := a.AuditFiles(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("files audit failed: %w", err)
	}

	sharingResult, err := a.AuditExternalSharing(ctx)
	if err != nil {
		return filesResult, nil, fmt.Errorf("sharing audit failed: %w", err)
	}

	return filesResult, sharingResult, nil
}
