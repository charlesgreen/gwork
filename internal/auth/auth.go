// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

// Package auth provides authentication for Google APIs using service accounts.
package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Scopes required for the audit tool.
var (
	// DriveScopes are the OAuth scopes required for Drive API access.
	DriveScopes = []string{
		drive.DriveReadonlyScope,
		drive.DriveMetadataReadonlyScope,
	}
)

// Authenticator handles service account authentication with domain-wide delegation.
type Authenticator struct {
	serviceAccountFile string
	adminEmail         string
}

// NewAuthenticator creates a new authenticator.
func NewAuthenticator(serviceAccountFile, adminEmail string) (*Authenticator, error) {
	if serviceAccountFile == "" {
		return nil, fmt.Errorf("service account file path is required")
	}
	if adminEmail == "" {
		return nil, fmt.Errorf("admin email is required for domain-wide delegation")
	}

	return &Authenticator{
		serviceAccountFile: serviceAccountFile,
		adminEmail:         adminEmail,
	}, nil
}

// GetDriveService creates an authenticated Drive service.
func (a *Authenticator) GetDriveService(ctx context.Context) (*drive.Service, error) {
	jsonCredentials, err := os.ReadFile(a.serviceAccountFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account file: %w", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, DriveScopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT config: %w", err)
	}

	// Set Subject for domain-wide delegation impersonation
	config.Subject = a.adminEmail

	ts := config.TokenSource(ctx)

	service, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	return service, nil
}
