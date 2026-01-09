// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ValidOutputFormats lists the supported output formats.
var ValidOutputFormats = []string{"csv", "json"}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	var errs []error

	// Validate Google config
	if c.Google.ServiceAccountFile == "" {
		errs = append(errs, errors.New("google.service_account_file is required"))
	} else if _, err := os.Stat(c.Google.ServiceAccountFile); os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("service account file not found: %s", c.Google.ServiceAccountFile))
	}

	if c.Google.AdminEmail == "" {
		errs = append(errs, errors.New("google.admin_email is required for domain-wide delegation"))
	} else if !strings.Contains(c.Google.AdminEmail, "@") {
		errs = append(errs, errors.New("google.admin_email must be a valid email address"))
	}

	if c.Google.Domain == "" {
		errs = append(errs, errors.New("google.domain is required"))
	}

	// Validate audit config
	if c.Audit.PageSize < 1 || c.Audit.PageSize > 1000 {
		errs = append(errs, errors.New("audit.page_size must be between 1 and 1000"))
	}

	// Validate output config
	if !isValidFormat(c.Output.Format) {
		errs = append(errs, fmt.Errorf("output.format must be one of: %s", strings.Join(ValidOutputFormats, ", ")))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func isValidFormat(format string) bool {
	for _, f := range ValidOutputFormats {
		if f == format {
			return true
		}
	}
	return false
}
