// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	// Create a temporary service account file for testing
	tmpDir := t.TempDir()
	validServiceAccountFile := filepath.Join(tmpDir, "service-account.json")
	err := os.WriteFile(validServiceAccountFile, []byte(`{"type":"service_account"}`), 0600)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid configuration",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 100,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: false,
		},
		{
			name: "missing service account file",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: "",
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 100,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: true,
			errorMsg:  "service_account_file is required",
		},
		{
			name: "service account file does not exist",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: "/nonexistent/file.json",
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 100,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: true,
			errorMsg:  "service account file not found",
		},
		{
			name: "missing admin email",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 100,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: true,
			errorMsg:  "admin_email is required",
		},
		{
			name: "invalid email format - no @ symbol",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "notanemail",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 100,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: true,
			errorMsg:  "must be a valid email address",
		},
		{
			name: "missing domain",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "",
				},
				Audit: AuditConfig{
					PageSize: 100,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: true,
			errorMsg:  "domain is required",
		},
		{
			name: "page size zero",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 0,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: true,
			errorMsg:  "page_size must be between 1 and 1000",
		},
		{
			name: "page size negative",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: -10,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: true,
			errorMsg:  "page_size must be between 1 and 1000",
		},
		{
			name: "page size too large",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 1001,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: true,
			errorMsg:  "page_size must be between 1 and 1000",
		},
		{
			name: "page size at minimum boundary",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 1,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: false,
		},
		{
			name: "page size at maximum boundary",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 1000,
				},
				Output: OutputConfig{
					Format: "csv",
				},
			},
			wantError: false,
		},
		{
			name: "invalid output format",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 100,
				},
				Output: OutputConfig{
					Format: "xml",
				},
			},
			wantError: true,
			errorMsg:  "output.format must be one of",
		},
		{
			name: "valid json format",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: validServiceAccountFile,
					AdminEmail:         "admin@example.com",
					Domain:             "example.com",
				},
				Audit: AuditConfig{
					PageSize: 100,
				},
				Output: OutputConfig{
					Format: "json",
				},
			},
			wantError: false,
		},
		{
			name: "multiple validation errors",
			config: Config{
				Google: GoogleConfig{
					ServiceAccountFile: "",
					AdminEmail:         "",
					Domain:             "",
				},
				Audit: AuditConfig{
					PageSize: 0,
				},
				Output: OutputConfig{
					Format: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "service_account_file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected bool
	}{
		{
			name:     "csv is valid",
			format:   "csv",
			expected: true,
		},
		{
			name:     "json is valid",
			format:   "json",
			expected: true,
		},
		{
			name:     "xml is invalid",
			format:   "xml",
			expected: false,
		},
		{
			name:     "empty string is invalid",
			format:   "",
			expected: false,
		},
		{
			name:     "uppercase CSV is invalid",
			format:   "CSV",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFormat(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidOutputFormats(t *testing.T) {
	// Ensure ValidOutputFormats contains expected formats
	assert.Contains(t, ValidOutputFormats, "csv")
	assert.Contains(t, ValidOutputFormats, "json")
	assert.Len(t, ValidOutputFormats, 2)
}
