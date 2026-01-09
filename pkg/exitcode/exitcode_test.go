// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package exitcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		expected int
	}{
		{
			name:     "Success code",
			exitCode: Success,
			expected: 0,
		},
		{
			name:     "ConfigError code",
			exitCode: ConfigError,
			expected: 1,
		},
		{
			name:     "AuthError code",
			exitCode: AuthError,
			expected: 2,
		},
		{
			name:     "APIError code",
			exitCode: APIError,
			expected: 3,
		},
		{
			name:     "InternalError code",
			exitCode: InternalError,
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.exitCode)
		})
	}
}

func TestExitCodesAreUnique(t *testing.T) {
	codes := map[int]string{
		Success:       "Success",
		ConfigError:   "ConfigError",
		AuthError:     "AuthError",
		APIError:      "APIError",
		InternalError: "InternalError",
	}

	// Ensure all codes are unique
	assert.Equal(t, 5, len(codes), "All exit codes should be unique")
}
