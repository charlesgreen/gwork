// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package drive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_IsExternalShare(t *testing.T) {
	client := &Client{
		domain: "example.com",
	}

	tests := []struct {
		name       string
		permission Permission
		expected   bool
	}{
		{
			name: "anyone type is always external",
			permission: Permission{
				Type: "anyone",
			},
			expected: true,
		},
		{
			name: "domain type with same domain is internal",
			permission: Permission{
				Type:   "domain",
				Domain: "example.com",
			},
			expected: false,
		},
		{
			name: "domain type with different domain is external",
			permission: Permission{
				Type:   "domain",
				Domain: "external.com",
			},
			expected: true,
		},
		{
			name: "user type with same domain email is internal",
			permission: Permission{
				Type:         "user",
				EmailAddress: "user@example.com",
			},
			expected: false,
		},
		{
			name: "user type with different domain email is external",
			permission: Permission{
				Type:         "user",
				EmailAddress: "user@external.com",
			},
			expected: true,
		},
		{
			name: "user type with empty email is internal",
			permission: Permission{
				Type:         "user",
				EmailAddress: "",
			},
			expected: false,
		},
		{
			name: "group type with same domain email is internal",
			permission: Permission{
				Type:         "group",
				EmailAddress: "group@example.com",
			},
			expected: false,
		},
		{
			name: "group type with different domain email is external",
			permission: Permission{
				Type:         "group",
				EmailAddress: "group@external.com",
			},
			expected: true,
		},
		{
			name: "group type with empty email is internal",
			permission: Permission{
				Type:         "group",
				EmailAddress: "",
			},
			expected: false,
		},
		{
			name: "user type with subdomain is external",
			permission: Permission{
				Type:         "user",
				EmailAddress: "user@sub.example.com",
			},
			expected: true,
		},
		{
			name: "user type with email containing multiple @ symbols",
			permission: Permission{
				Type:         "user",
				EmailAddress: "user@name@external.com",
			},
			expected: true,
		},
		{
			name: "unknown permission type is internal",
			permission: Permission{
				Type: "unknown",
			},
			expected: false,
		},
		{
			name: "domain type with empty domain is external",
			permission: Permission{
				Type:   "domain",
				Domain: "",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.IsExternalShare(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractDomain(t *testing.T) {
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
			name:     "email with multiple @ symbols uses last one",
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
			name:     "complex email address",
			email:    "user+tag@example.co.uk",
			expected: "example.co.uk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractDomain(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}
