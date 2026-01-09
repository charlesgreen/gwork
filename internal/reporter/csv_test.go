// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

package reporter

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlesgreen/gwork/internal/audit"
	"github.com/charlesgreen/gwork/internal/gcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCSVReporter(t *testing.T) {
	tests := []struct {
		name      string
		outputDir string
		wantError bool
	}{
		{
			name:      "valid output directory",
			outputDir: filepath.Join(t.TempDir(), "output"),
			wantError: false,
		},
		{
			name:      "nested output directory",
			outputDir: filepath.Join(t.TempDir(), "parent", "child", "output"),
			wantError: false,
		},
		{
			name:      "current directory",
			outputDir: ".",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := NewCSVReporter(tt.outputDir)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, reporter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reporter)
				assert.Equal(t, tt.outputDir, reporter.OutputDir())
			}
		})
	}
}

func TestCSVReporter_WriteFilesByOwner(t *testing.T) {
	tests := []struct {
		name      string
		records   []audit.FileRecord
		wantError bool
	}{
		{
			name: "multiple records",
			records: []audit.FileRecord{
				{
					OwnerEmail:   "alice@example.com",
					FileID:       "file1",
					FileName:     "document1.pdf",
					FileType:     "application/pdf",
					CreatedTime:  time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					ModifiedTime: time.Date(2024, 1, 20, 15, 0, 0, 0, time.UTC),
					SizeBytes:    1024,
				},
				{
					OwnerEmail:   "bob@example.com",
					FileID:       "file2",
					FileName:     "spreadsheet.xlsx",
					FileType:     "application/vnd.ms-excel",
					CreatedTime:  time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC),
					ModifiedTime: time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC),
					SizeBytes:    2048,
				},
				{
					OwnerEmail:   "alice@example.com",
					FileID:       "file3",
					FileName:     "presentation.pptx",
					FileType:     "application/vnd.ms-powerpoint",
					CreatedTime:  time.Date(2024, 3, 1, 11, 0, 0, 0, time.UTC),
					ModifiedTime: time.Date(2024, 3, 10, 16, 0, 0, 0, time.UTC),
					SizeBytes:    4096,
				},
			},
			wantError: false,
		},
		{
			name:      "empty records",
			records:   []audit.FileRecord{},
			wantError: false,
		},
		{
			name: "single record",
			records: []audit.FileRecord{
				{
					OwnerEmail:   "user@example.com",
					FileID:       "single",
					FileName:     "single.txt",
					FileType:     "text/plain",
					CreatedTime:  time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC),
					ModifiedTime: time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC),
					SizeBytes:    512,
				},
			},
			wantError: false,
		},
		{
			name: "record with zero time",
			records: []audit.FileRecord{
				{
					OwnerEmail:   "user@example.com",
					FileID:       "notime",
					FileName:     "notime.txt",
					FileType:     "text/plain",
					CreatedTime:  time.Time{},
					ModifiedTime: time.Time{},
					SizeBytes:    0,
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			reporter, err := NewCSVReporter(tmpDir)
			require.NoError(t, err)

			err = reporter.WriteFilesByOwner(tt.records)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify the CSV file was created
			csvPath := filepath.Join(tmpDir, "files_by_owner.csv")
			assert.FileExists(t, csvPath)

			// Read and verify CSV content
			file, err := os.Open(csvPath)
			require.NoError(t, err)
			defer file.Close() //nolint:errcheck // test cleanup

			reader := csv.NewReader(file)
			rows, err := reader.ReadAll()
			require.NoError(t, err)

			// Check header
			require.GreaterOrEqual(t, len(rows), 1, "CSV should have at least a header")
			expectedHeader := []string{
				"owner_email", "file_id", "file_name", "file_type",
				"created_time", "modified_time", "size_bytes",
			}
			assert.Equal(t, expectedHeader, rows[0])

			// Check number of data rows
			assert.Equal(t, len(tt.records)+1, len(rows), "CSV should have header + data rows")

			// If we have records, verify they are sorted by owner email
			if len(tt.records) > 0 {
				for i := 1; i < len(rows); i++ {
					if i > 1 {
						// Check sorting
						assert.LessOrEqual(t, rows[i-1][0], rows[i][0], "Rows should be sorted by owner_email")
					}
				}
			}
		})
	}
}

func TestCSVReporter_WriteExternalSharing(t *testing.T) {
	tests := []struct {
		name      string
		records   []audit.ExternalShareRecord
		wantError bool
	}{
		{
			name: "multiple records",
			records: []audit.ExternalShareRecord{
				{
					OwnerEmail:       "alice@example.com",
					FileID:           "file1",
					FileName:         "shared1.pdf",
					SharedWithEmail:  "external@other.com",
					SharedWithDomain: "other.com",
					PermissionType:   "user",
					PermissionRole:   "reader",
					SharedDate:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				},
				{
					OwnerEmail:       "bob@example.com",
					FileID:           "file2",
					FileName:         "shared2.xlsx",
					SharedWithEmail:  "",
					SharedWithDomain: "external.org",
					PermissionType:   "domain",
					PermissionRole:   "writer",
					SharedDate:       time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC),
				},
				{
					OwnerEmail:       "alice@example.com",
					FileID:           "file3",
					FileName:         "public.txt",
					SharedWithEmail:  "",
					SharedWithDomain: "",
					PermissionType:   "anyone",
					PermissionRole:   "reader",
					SharedDate:       time.Time{},
				},
			},
			wantError: false,
		},
		{
			name:      "empty records",
			records:   []audit.ExternalShareRecord{},
			wantError: false,
		},
		{
			name: "single record",
			records: []audit.ExternalShareRecord{
				{
					OwnerEmail:       "user@example.com",
					FileID:           "single",
					FileName:         "single.txt",
					SharedWithEmail:  "guest@external.com",
					SharedWithDomain: "external.com",
					PermissionType:   "user",
					PermissionRole:   "commenter",
					SharedDate:       time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			reporter, err := NewCSVReporter(tmpDir)
			require.NoError(t, err)

			err = reporter.WriteExternalSharing(tt.records)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify the CSV file was created
			csvPath := filepath.Join(tmpDir, "external_sharing.csv")
			assert.FileExists(t, csvPath)

			// Read and verify CSV content
			file, err := os.Open(csvPath)
			require.NoError(t, err)
			defer file.Close() //nolint:errcheck // test cleanup

			reader := csv.NewReader(file)
			rows, err := reader.ReadAll()
			require.NoError(t, err)

			// Check header
			require.GreaterOrEqual(t, len(rows), 1, "CSV should have at least a header")
			expectedHeader := []string{
				"owner_email", "file_id", "file_name", "shared_with_email",
				"shared_with_domain", "permission_type", "permission_role", "shared_date",
			}
			assert.Equal(t, expectedHeader, rows[0])

			// Check number of data rows
			assert.Equal(t, len(tt.records)+1, len(rows), "CSV should have header + data rows")

			// If we have records, verify they are sorted by owner email
			if len(tt.records) > 0 {
				for i := 1; i < len(rows); i++ {
					if i > 1 {
						// Check sorting
						assert.LessOrEqual(t, rows[i-1][0], rows[i][0], "Rows should be sorted by owner_email")
					}
				}
			}
		})
	}
}

func TestCSVReporter_SortingByOwner(t *testing.T) {
	tmpDir := t.TempDir()
	reporter, err := NewCSVReporter(tmpDir)
	require.NoError(t, err)

	// Create records in unsorted order
	records := []audit.FileRecord{
		{OwnerEmail: "charlie@example.com", FileID: "1", FileName: "c.txt"},
		{OwnerEmail: "alice@example.com", FileID: "2", FileName: "a.txt"},
		{OwnerEmail: "bob@example.com", FileID: "3", FileName: "b.txt"},
		{OwnerEmail: "alice@example.com", FileID: "4", FileName: "z.txt"},
		{OwnerEmail: "alice@example.com", FileID: "5", FileName: "m.txt"},
	}

	err = reporter.WriteFilesByOwner(records)
	require.NoError(t, err)

	// Read CSV and verify sorting
	csvPath := filepath.Join(tmpDir, "files_by_owner.csv")
	file, err := os.Open(csvPath)
	require.NoError(t, err)
	defer file.Close() //nolint:errcheck // test cleanup

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	require.NoError(t, err)

	// Verify sorting: alice (3 files), bob (1 file), charlie (1 file)
	require.Equal(t, 6, len(rows)) // header + 5 records
	assert.Equal(t, "alice@example.com", rows[1][0])
	assert.Equal(t, "alice@example.com", rows[2][0])
	assert.Equal(t, "alice@example.com", rows[3][0])
	assert.Equal(t, "bob@example.com", rows[4][0])
	assert.Equal(t, "charlie@example.com", rows[5][0])

	// Verify alice's files are sorted by name
	assert.Equal(t, "a.txt", rows[1][2])
	assert.Equal(t, "m.txt", rows[2][2])
	assert.Equal(t, "z.txt", rows[3][2])
}

func TestCSVReporter_OutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	reporter, err := NewCSVReporter(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, tmpDir, reporter.OutputDir())
}

func TestCSVReporter_TimestampFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	reporter, err := NewCSVReporter(tmpDir)
	require.NoError(t, err)

	records := []audit.FileRecord{
		{
			OwnerEmail:   "user@example.com",
			FileID:       "file1",
			FileName:     "test.txt",
			FileType:     "text/plain",
			CreatedTime:  time.Date(2024, 5, 15, 14, 30, 45, 0, time.UTC),
			ModifiedTime: time.Date(2024, 5, 20, 16, 45, 30, 0, time.UTC),
			SizeBytes:    1024,
		},
	}

	err = reporter.WriteFilesByOwner(records)
	require.NoError(t, err)

	// Read CSV and verify timestamp format
	csvPath := filepath.Join(tmpDir, "files_by_owner.csv")
	file, err := os.Open(csvPath)
	require.NoError(t, err)
	defer file.Close() //nolint:errcheck // test cleanup

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	require.NoError(t, err)

	require.Equal(t, 2, len(rows))
	assert.Equal(t, "2024-05-15T14:30:45Z", rows[1][4]) // created_time
	assert.Equal(t, "2024-05-20T16:45:30Z", rows[1][5]) // modified_time
}

func TestCSVReporter_WriteProjectActivity(t *testing.T) {
	tests := []struct {
		name      string
		statuses  []gcp.ProjectStatus
		wantError bool
	}{
		{
			name: "multiple projects with different statuses",
			statuses: []gcp.ProjectStatus{
				{
					ProjectID:       "active-1",
					ProjectName:     "Active Project",
					LastActivity:    "2024-01-15",
					DaysSinceActive: 30,
					Status:          "ACTIVE",
					ConsoleLink:     "https://console.cloud.google.com/home/dashboard?project=active-1",
				},
				{
					ProjectID:       "inactive-1",
					ProjectName:     "Inactive Project",
					LastActivity:    "2023-06-01",
					DaysSinceActive: 200,
					Status:          "INACTIVE",
					ConsoleLink:     "https://console.cloud.google.com/home/dashboard?project=inactive-1",
				},
				{
					ProjectID:       "no-logs",
					ProjectName:     "No Logs Project",
					LastActivity:    "No activity found",
					DaysSinceActive: -1,
					Status:          "INACTIVE",
					ConsoleLink:     "https://console.cloud.google.com/home/dashboard?project=no-logs",
				},
			},
			wantError: false,
		},
		{
			name:      "empty project list",
			statuses:  []gcp.ProjectStatus{},
			wantError: false,
		},
		{
			name: "single project",
			statuses: []gcp.ProjectStatus{
				{
					ProjectID:       "single",
					ProjectName:     "Single Project",
					LastActivity:    "2024-02-01",
					DaysSinceActive: 15,
					Status:          "ACTIVE",
					ConsoleLink:     "https://console.cloud.google.com/home/dashboard?project=single",
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			reporter, err := NewCSVReporter(tmpDir)
			require.NoError(t, err)

			err = reporter.WriteProjectActivity(tt.statuses)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify the CSV file was created
			csvPath := filepath.Join(tmpDir, "project_activity.csv")
			assert.FileExists(t, csvPath)

			// Read and verify CSV content
			file, err := os.Open(csvPath)
			require.NoError(t, err)
			defer file.Close() //nolint:errcheck // test cleanup

			reader := csv.NewReader(file)
			rows, err := reader.ReadAll()
			require.NoError(t, err)

			// Check header
			require.GreaterOrEqual(t, len(rows), 1, "CSV should have at least a header")
			expectedHeader := []string{
				"project_id", "project_name", "last_activity",
				"days_since_activity", "status", "console_link",
			}
			assert.Equal(t, expectedHeader, rows[0])

			// Check number of data rows
			assert.Equal(t, len(tt.statuses)+1, len(rows), "CSV should have header + data rows")
		})
	}
}

func TestCSVReporter_WriteProjectActivity_NegativeDays(t *testing.T) {
	tmpDir := t.TempDir()
	reporter, err := NewCSVReporter(tmpDir)
	require.NoError(t, err)

	statuses := []gcp.ProjectStatus{
		{
			ProjectID:       "unknown-project",
			ProjectName:     "Unknown",
			LastActivity:    "No activity found",
			DaysSinceActive: -1, // Unknown/N/A
			Status:          "INACTIVE",
			ConsoleLink:     "https://console.cloud.google.com/home/dashboard?project=unknown-project",
		},
	}

	err = reporter.WriteProjectActivity(statuses)
	require.NoError(t, err)

	// Read and verify N/A handling
	csvPath := filepath.Join(tmpDir, "project_activity.csv")
	file, err := os.Open(csvPath)
	require.NoError(t, err)
	defer file.Close() //nolint:errcheck // test cleanup

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	require.NoError(t, err)

	assert.Equal(t, "N/A", rows[1][3]) // days_since_activity column
}

func TestCSVReporter_WriteProjectActivity_Sorting(t *testing.T) {
	tmpDir := t.TempDir()
	reporter, err := NewCSVReporter(tmpDir)
	require.NoError(t, err)

	// Create records in unsorted order
	statuses := []gcp.ProjectStatus{
		{ProjectID: "active-z", ProjectName: "Z Active", Status: "ACTIVE"},
		{ProjectID: "inactive-a", ProjectName: "A Inactive", Status: "INACTIVE"},
		{ProjectID: "active-a", ProjectName: "A Active", Status: "ACTIVE"},
		{ProjectID: "unknown-x", ProjectName: "X Unknown", Status: "UNKNOWN"},
		{ProjectID: "inactive-b", ProjectName: "B Inactive", Status: "INACTIVE"},
	}

	err = reporter.WriteProjectActivity(statuses)
	require.NoError(t, err)

	// Read CSV and verify sorting
	csvPath := filepath.Join(tmpDir, "project_activity.csv")
	file, err := os.Open(csvPath)
	require.NoError(t, err)
	defer file.Close() //nolint:errcheck // test cleanup

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	require.NoError(t, err)

	// Verify sorting: INACTIVE first (sorted by ID), then ACTIVE (sorted by ID), then UNKNOWN
	require.Equal(t, 6, len(rows)) // header + 5 records

	// INACTIVE projects first (sorted by ID)
	assert.Equal(t, "inactive-a", rows[1][0])
	assert.Equal(t, "INACTIVE", rows[1][4])
	assert.Equal(t, "inactive-b", rows[2][0])
	assert.Equal(t, "INACTIVE", rows[2][4])

	// ACTIVE projects next (sorted by ID)
	assert.Equal(t, "active-a", rows[3][0])
	assert.Equal(t, "ACTIVE", rows[3][4])
	assert.Equal(t, "active-z", rows[4][0])
	assert.Equal(t, "ACTIVE", rows[4][4])

	// UNKNOWN last
	assert.Equal(t, "unknown-x", rows[5][0])
	assert.Equal(t, "UNKNOWN", rows[5][4])
}
