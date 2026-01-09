// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package config

import "github.com/spf13/viper"

const (
	// DefaultPageSize is the default number of items per API page.
	DefaultPageSize = 1000

	// DefaultOutputFormat is the default output format.
	DefaultOutputFormat = "csv"

	// DefaultOutputDirectory is the default output directory.
	DefaultOutputDirectory = "./output"
)

// setDefaults sets default values in viper.
func setDefaults(v *viper.Viper) {
	v.SetDefault("audit.include_shared_drives", true)
	v.SetDefault("audit.page_size", DefaultPageSize)
	v.SetDefault("output.format", DefaultOutputFormat)
	v.SetDefault("output.directory", DefaultOutputDirectory)
}

// NewDefault creates a new Config with default values.
func NewDefault() *Config {
	return &Config{
		Google: GoogleConfig{
			ServiceAccountFile: "",
			AdminEmail:         "",
			Domain:             "",
		},
		Audit: AuditConfig{
			IncludeSharedDrives: true,
			PageSize:            DefaultPageSize,
		},
		Output: OutputConfig{
			Format:    DefaultOutputFormat,
			Directory: DefaultOutputDirectory,
		},
	}
}
