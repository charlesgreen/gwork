// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for gwork.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure.
type Config struct {
	Google GoogleConfig `yaml:"google" mapstructure:"google"`
	Audit  AuditConfig  `yaml:"audit" mapstructure:"audit"`
	Output OutputConfig `yaml:"output" mapstructure:"output"`
}

// GoogleConfig contains Google API configuration.
type GoogleConfig struct {
	ServiceAccountFile string `yaml:"service_account_file" mapstructure:"service_account_file"`
	AdminEmail         string `yaml:"admin_email" mapstructure:"admin_email"`
	Domain             string `yaml:"domain" mapstructure:"domain"`
}

// AuditConfig contains audit-specific configuration.
type AuditConfig struct {
	IncludeSharedDrives bool  `yaml:"include_shared_drives" mapstructure:"include_shared_drives"`
	PageSize            int64 `yaml:"page_size" mapstructure:"page_size"`
}

// OutputConfig contains output formatting configuration.
type OutputConfig struct {
	Format    string `yaml:"format" mapstructure:"format"`
	Directory string `yaml:"directory" mapstructure:"directory"`
}

// Load reads and parses the configuration file.
func Load(configPath string) (*Config, error) {
	v := viper.New()
	setDefaults(v)

	v.SetConfigName(".gwork")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(".")
		homeDir, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(homeDir)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to a file.
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
