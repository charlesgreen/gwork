// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name                string
		domain              string
		pageSize            int64
		includeSharedDrives bool
	}{
		{
			name:                "standard configuration",
			domain:              "example.com",
			pageSize:            100,
			includeSharedDrives: true,
		},
		{
			name:                "with shared drives disabled",
			domain:              "example.com",
			pageSize:            500,
			includeSharedDrives: false,
		},
		{
			name:                "maximum page size",
			domain:              "test.org",
			pageSize:            1000,
			includeSharedDrives: true,
		},
		{
			name:                "minimum page size",
			domain:              "test.org",
			pageSize:            1,
			includeSharedDrives: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(nil, tt.domain, tt.pageSize, tt.includeSharedDrives)

			assert.NotNil(t, client)
			assert.Equal(t, tt.domain, client.domain)
			assert.Equal(t, tt.pageSize, client.pageSize)
			assert.Equal(t, tt.includeSharedDrives, client.includeSharedDrives)
		})
	}
}

func TestClient_Domain(t *testing.T) {
	tests := []struct {
		name   string
		domain string
	}{
		{
			name:   "standard domain",
			domain: "example.com",
		},
		{
			name:   "subdomain",
			domain: "mail.example.com",
		},
		{
			name:   "empty domain",
			domain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(nil, tt.domain, 100, true)
			assert.Equal(t, tt.domain, client.Domain())
		})
	}
}
