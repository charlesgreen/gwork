// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewDefault(t *testing.T) {
	cfg := NewDefault()

	// Test Google config defaults
	assert.Equal(t, "", cfg.Google.ServiceAccountFile, "ServiceAccountFile should be empty by default")
	assert.Equal(t, "", cfg.Google.AdminEmail, "AdminEmail should be empty by default")
	assert.Equal(t, "", cfg.Google.Domain, "Domain should be empty by default")

	// Test Audit config defaults
	assert.Equal(t, true, cfg.Audit.IncludeSharedDrives, "IncludeSharedDrives should be true by default")
	assert.Equal(t, int64(DefaultPageSize), cfg.Audit.PageSize, "PageSize should be DefaultPageSize")

	// Test Output config defaults
	assert.Equal(t, DefaultOutputFormat, cfg.Output.Format, "Format should be DefaultOutputFormat")
	assert.Equal(t, DefaultOutputDirectory, cfg.Output.Directory, "Directory should be DefaultOutputDirectory")
}

func TestSetDefaults(t *testing.T) {
	v := viper.New()
	setDefaults(v)

	// Test that defaults are set in viper
	assert.Equal(t, true, v.GetBool("audit.include_shared_drives"))
	assert.Equal(t, int64(DefaultPageSize), v.GetInt64("audit.page_size"))
	assert.Equal(t, DefaultOutputFormat, v.GetString("output.format"))
	assert.Equal(t, DefaultOutputDirectory, v.GetString("output.directory"))
}

func TestDefaultConstants(t *testing.T) {
	// Verify the default constant values are as expected
	assert.Equal(t, 1000, DefaultPageSize, "DefaultPageSize should be 1000")
	assert.Equal(t, "csv", DefaultOutputFormat, "DefaultOutputFormat should be csv")
	assert.Equal(t, "./output", DefaultOutputDirectory, "DefaultOutputDirectory should be ./output")
}

func TestDefaultsAreValid(t *testing.T) {
	// The default config should not be valid because required fields are empty
	// This is expected - users must provide service_account_file, admin_email, and domain
	cfg := NewDefault()
	err := cfg.Validate()

	// We expect validation to fail because required fields are missing
	assert.Error(t, err, "Default config should fail validation (missing required fields)")
	assert.Contains(t, err.Error(), "service_account_file is required")
	assert.Contains(t, err.Error(), "admin_email is required")
	assert.Contains(t, err.Error(), "domain is required")
}

func TestDefaultPageSizeIsWithinValidRange(t *testing.T) {
	// Ensure the default page size is within the valid range
	assert.GreaterOrEqual(t, DefaultPageSize, 1, "DefaultPageSize should be >= 1")
	assert.LessOrEqual(t, DefaultPageSize, 1000, "DefaultPageSize should be <= 1000")
}

func TestDefaultOutputFormatIsValid(t *testing.T) {
	// Ensure the default output format is in the valid list
	assert.Contains(t, ValidOutputFormats, DefaultOutputFormat, "DefaultOutputFormat should be in ValidOutputFormats")
}
