// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

// Package exitcode defines exit codes for the gwork CLI.
package exitcode

const (
	// Success indicates the command completed successfully.
	Success = 0

	// ConfigError indicates a configuration error.
	ConfigError = 1

	// AuthError indicates an authentication error.
	AuthError = 2

	// APIError indicates a Google API error.
	APIError = 3

	// InternalError indicates an internal error.
	InternalError = 10
)
