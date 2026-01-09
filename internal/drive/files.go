// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"context"
	"fmt"
)

// ListAllFiles retrieves all files in the domain.
func (c *Client) ListAllFiles(ctx context.Context) ([]FileInfo, error) {
	var allFiles []FileInfo
	pageToken := ""

	for {
		select {
		case <-ctx.Done():
			return allFiles, ctx.Err()
		default:
		}

		opts := &ListFilesOptions{
			Corpora:                   "domain",
			PageSize:                  c.pageSize,
			PageToken:                 pageToken,
			Fields:                    "nextPageToken, files(id, name, mimeType, owners, createdTime, modifiedTime, size)",
			SupportsAllDrives:         c.includeSharedDrives,
			IncludeItemsFromAllDrives: c.includeSharedDrives,
		}

		result, err := c.api.ListFiles(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, file := range result.Files {
			ownerEmail := ""
			if len(file.Owners) > 0 {
				ownerEmail = file.Owners[0].EmailAddress
			}

			allFiles = append(allFiles, FileInfo{
				ID:           file.Id,
				Name:         file.Name,
				MimeType:     file.MimeType,
				OwnerEmail:   ownerEmail,
				CreatedTime:  file.CreatedTime,
				ModifiedTime: file.ModifiedTime,
				Size:         file.Size,
			})
		}

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return allFiles, nil
}
