// Copyright 2025 Charles Green, LLC
// SPDX-License-Identifier: Apache-2.0

// Package main provides the gwork CLI tool for Google Workspace security audits.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charlesgreen/gwork/internal/audit"
	"github.com/charlesgreen/gwork/internal/config"
	"github.com/charlesgreen/gwork/internal/gcp"
	"github.com/charlesgreen/gwork/internal/reporter"
	"github.com/charlesgreen/gwork/pkg/exitcode"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"

	cfgFile string
	verbose bool
	quiet   bool

	// Project audit flags
	inactiveDays  int
	workers       int
	freshnessDays int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(exitcode.InternalError)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gwork",
	Short: "Google Workspace security and audit tool",
	Long: `gwork is a CLI tool for auditing Google Workspace Drive files.
It helps identify files shared externally and generates reports
grouped by file owner.`,
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run audit operations",
	Long:  `Run various audit operations on Google Workspace Drive.`,
}

var auditFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Generate files by owner CSV",
	Long:  `Fetch all files from Google Drive across the domain and generate a CSV grouped by owner.`,
	RunE:  runAuditFiles,
}

var auditSharingCmd = &cobra.Command{
	Use:   "sharing",
	Short: "Generate external sharing CSV",
	Long:  `Generate a list of files shared externally (outside the organization domain).`,
	RunE:  runAuditSharing,
}

var auditAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all audits",
	Long:  `Run all audit operations: files by owner and external sharing.`,
	RunE:  runAuditAll,
}

var auditProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Audit GCP project activity",
	Long: `Check all accessible GCP projects for activity and generate a report.

Projects with no audit log activity within the threshold are marked INACTIVE.
Uses your current gcloud CLI credentials (no service account required).

Examples:
  gwork audit projects                    # Use default 60-day threshold
  gwork audit projects --days 90          # Use 90-day threshold
  gwork audit projects --workers 20       # Use 20 concurrent workers`,
	RunE: runAuditProjects,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `Commands for managing gwork configuration.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate sample config file",
	Long:  `Create a new .gwork.yaml configuration file with default values.`,
	RunE:  runConfigInit,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gwork v%s\n", version)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .gwork.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-error output")

	// Build command tree
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)

	auditCmd.AddCommand(auditFilesCmd)
	auditCmd.AddCommand(auditSharingCmd)
	auditCmd.AddCommand(auditAllCmd)
	auditCmd.AddCommand(auditProjectsCmd)

	configCmd.AddCommand(configInitCmd)

	// Project audit flags
	auditProjectsCmd.Flags().IntVar(&inactiveDays, "days", 60,
		"Days of inactivity to consider a project inactive")
	auditProjectsCmd.Flags().IntVar(&workers, "workers", 10,
		"Number of concurrent workers for checking projects")
	auditProjectsCmd.Flags().IntVar(&freshnessDays, "freshness", 400,
		"Days to look back in audit logs")
}

func loadConfig() (*config.Config, error) {
	return config.Load(cfgFile)
}

func runAuditFiles(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	auditor, err := audit.NewAuditor(cfg)
	if err != nil {
		return fmt.Errorf("failed to create auditor: %w", err)
	}

	if !quiet {
		fmt.Println("Fetching files from Google Drive...")
	}

	result, err := auditor.AuditFiles(ctx)
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	rep, err := reporter.NewCSVReporter(cfg.Output.Directory)
	if err != nil {
		return fmt.Errorf("failed to create reporter: %w", err)
	}

	if err := rep.WriteFilesByOwner(result.FileRecords); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	if !quiet {
		fmt.Printf("Files audit complete. Total files: %d\n", result.TotalFiles)
		fmt.Printf("Report saved to: %s/files_by_owner.csv\n", rep.OutputDir())
	}

	return nil
}

func runAuditSharing(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	auditor, err := audit.NewAuditor(cfg)
	if err != nil {
		return fmt.Errorf("failed to create auditor: %w", err)
	}

	if !quiet {
		fmt.Println("Analyzing external sharing...")
	}

	result, err := auditor.AuditExternalSharing(ctx)
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	rep, err := reporter.NewCSVReporter(cfg.Output.Directory)
	if err != nil {
		return fmt.Errorf("failed to create reporter: %w", err)
	}

	if err := rep.WriteExternalSharing(result.ExternalShares); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	if !quiet {
		fmt.Printf("Sharing audit complete. Files processed: %d\n", result.FilesProcessed)
		fmt.Printf("External shares found: %d\n", result.TotalExternalShares)
		fmt.Printf("Report saved to: %s/external_sharing.csv\n", rep.OutputDir())

		if len(result.Errors) > 0 {
			fmt.Printf("Warnings: %d files could not be processed\n", len(result.Errors))
			if verbose {
				for _, e := range result.Errors {
					fmt.Printf("  - %v\n", e)
				}
			}
		}
	}

	return nil
}

func runAuditAll(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	auditor, err := audit.NewAuditor(cfg)
	if err != nil {
		return fmt.Errorf("failed to create auditor: %w", err)
	}

	if !quiet {
		fmt.Println("Running all audits...")
	}

	filesResult, sharingResult, err := auditor.AuditAll(ctx)
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	rep, err := reporter.NewCSVReporter(cfg.Output.Directory)
	if err != nil {
		return fmt.Errorf("failed to create reporter: %w", err)
	}

	if err := rep.WriteFilesByOwner(filesResult.FileRecords); err != nil {
		return fmt.Errorf("failed to write files report: %w", err)
	}

	if err := rep.WriteExternalSharing(sharingResult.ExternalShares); err != nil {
		return fmt.Errorf("failed to write sharing report: %w", err)
	}

	if !quiet {
		fmt.Printf("Files audit complete. Total files: %d\n", filesResult.TotalFiles)
		fmt.Printf("Report saved to: %s/files_by_owner.csv\n", rep.OutputDir())
		fmt.Printf("Sharing audit complete. Files processed: %d\n", sharingResult.FilesProcessed)
		fmt.Printf("External shares found: %d\n", sharingResult.TotalExternalShares)
		fmt.Printf("Report saved to: %s/external_sharing.csv\n", rep.OutputDir())

		if len(sharingResult.Errors) > 0 {
			fmt.Printf("Warnings: %d files could not be processed\n", len(sharingResult.Errors))
			if verbose {
				for _, e := range sharingResult.Errors {
					fmt.Printf("  - %v\n", e)
				}
			}
		}
	}

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	configPath := ".gwork.yaml"

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file %s already exists", configPath)
	}

	cfg := config.NewDefault()
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Printf("Created %s\n", configPath)
	fmt.Println("Please edit the file to add your Google service account credentials.")
	return nil
}

func runAuditProjects(cmd *cobra.Command, args []string) error {
	// Try to load config for output directory, but don't require it
	outputDir := "./output"
	if cfg, err := loadConfig(); err == nil {
		outputDir = cfg.Output.Directory
	}

	ctx := context.Background()

	// Create GCP client (uses gcloud CLI, no service account)
	gcpClient := gcp.NewClient()

	if !quiet {
		fmt.Printf("Checking GCP project activity (inactive = no activity in %d+ days)...\n", inactiveDays)
	}

	opts := gcp.CheckActivityOptions{
		InactiveDays:  inactiveDays,
		FreshnessDays: freshnessDays,
		Workers:       workers,
	}

	results, errors := gcpClient.CheckAllProjects(ctx, opts)
	if results == nil && len(errors) > 0 {
		return fmt.Errorf("failed to check projects: %v", errors[0])
	}

	rep, err := reporter.NewCSVReporter(outputDir)
	if err != nil {
		return fmt.Errorf("failed to create reporter: %w", err)
	}

	if err := rep.WriteProjectActivity(results); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Count by status
	var activeCount, inactiveCount, unknownCount int
	for _, r := range results {
		switch r.Status {
		case "ACTIVE":
			activeCount++
		case "INACTIVE":
			inactiveCount++
		default:
			unknownCount++
		}
	}

	if !quiet {
		fmt.Println()
		fmt.Println("==============================================")
		fmt.Println("  Summary")
		fmt.Println("==============================================")
		fmt.Printf("  Total projects scanned:  %d\n", len(results))
		fmt.Printf("  Active projects:         %d\n", activeCount)
		fmt.Printf("  Inactive projects:       %d\n", inactiveCount)
		if unknownCount > 0 {
			fmt.Printf("  Unknown status:          %d\n", unknownCount)
		}
		fmt.Println()
		fmt.Printf("Report saved to: %s/project_activity.csv\n", rep.OutputDir())
		fmt.Println()
		fmt.Println("To delete a project:")
		fmt.Println("  gcloud projects delete PROJECT_ID")

		if len(errors) > 0 {
			fmt.Printf("\nWarnings: %d projects had errors\n", len(errors))
			if verbose {
				for _, e := range errors {
					fmt.Printf("  - %v\n", e)
				}
			}
		}
	}

	return nil
}
