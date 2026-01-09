// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"context"
	"fmt"
	"strings"
)

// GetFilePermissions retrieves all permissions for a file.
func (c *Client) GetFilePermissions(ctx context.Context, fileID string) ([]Permission, error) {
	var allPerms []Permission
	pageToken := ""

	for {
		select {
		case <-ctx.Done():
			return allPerms, ctx.Err()
		default:
		}

		opts := &ListPermissionsOptions{
			Fields:            "nextPageToken, permissions(id, type, role, emailAddress, domain, displayName)",
			PageToken:         pageToken,
			SupportsAllDrives: c.includeSharedDrives,
		}

		result, err := c.api.ListPermissions(ctx, fileID, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list permissions for file %s: %w", fileID, err)
		}

		for _, perm := range result.Permissions {
			allPerms = append(allPerms, Permission{
				ID:           perm.Id,
				Type:         perm.Type,
				Role:         perm.Role,
				EmailAddress: perm.EmailAddress,
				Domain:       perm.Domain,
				DisplayName:  perm.DisplayName,
			})
		}

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return allPerms, nil
}

// IsExternalShare checks if a permission is external to the domain.
func (c *Client) IsExternalShare(perm Permission) bool {
	switch perm.Type {
	case "anyone":
		return true
	case "domain":
		return perm.Domain != c.domain
	case "user", "group":
		if perm.EmailAddress == "" {
			return false
		}
		emailDomain := ExtractDomain(perm.EmailAddress)
		return emailDomain != c.domain
	default:
		return false
	}
}

// ExtractDomain extracts the domain part from an email address.
func ExtractDomain(email string) string {
	idx := strings.LastIndex(email, "@")
	if idx < 0 {
		return ""
	}
	return email[idx+1:]
}
