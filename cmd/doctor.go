/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose the nix environment installation",
	Long: `Perform a comprehensive health check of your Nix installation.

This command checks:
- Nix binary presence and version
- Nixpkgs channel availability and status
- Environment configuration
- Common installation issues
- Shell configuration files
- Nix store permissions
- Flakes support status
- NSM configuration
- Platform-specific requirements

Examples:
  nsm doctor          # Run all diagnostics
  nsm doctor --json   # Output results in JSON format
  nsm doctor --fix    # Attempt to fix detected issues
  nsm doctor --md     # Output in markdown format
  nsm doctor --csv    # Output in CSV format
  nsm doctor --table  # Output in table format (default)`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonFormat, _ := cmd.Flags().GetBool("json")
		csvFormat, _ := cmd.Flags().GetBool("csv")
		markdownFormat, _ := cmd.Flags().GetBool("md")
		tableFormat, _ := cmd.Flags().GetBool("table")
		fixIssues, _ := cmd.Flags().GetBool("fix")
		verbose, _ := cmd.Flags().GetBool("verbose")
		noColor, _ := cmd.Flags().GetBool("no-color")

		// Determine output format
		var outputFormat utils.TableFormat

		if jsonFormat {
			outputFormat = utils.FormatJSON
		} else if csvFormat {
			outputFormat = utils.FormatCSV
		} else if markdownFormat {
			outputFormat = utils.FormatMarkdown
		} else if tableFormat || (!jsonFormat && !csvFormat && !markdownFormat) {
			outputFormat = utils.FormatText
		}

		// Run diagnostics silently for structure output formats
		if outputFormat != utils.FormatText {
			utils.Info("Running diagnostics...")
		} else if !noColor {
			utils.Info("🔍 Running diagnostics...")
			utils.Info("=====================")
		} else {
			utils.Info("Running diagnostics...")
			utils.Info("=====================")
		}

		startTime := time.Now()

		// Run comprehensive diagnostics
		results := utils.RunDiagnostics()

		// Count issues by severity
		var errors, warnings int
		for _, result := range results {
			if result.Status == utils.StatusError {
				errors++
			} else if result.Status == utils.StatusWarning {
				warnings++
			}
		}

		// Process output based on format
		if outputFormat != utils.FormatText {
			// Use enhanced table for structured output
			tableOutput := utils.FormatDiagnosticTable(results, outputFormat)
			fmt.Println(tableOutput)
		} else {
			// Format as human-readable output
			if noColor {
				printDiagnosticResultsNoColor(results, verbose)
			} else {
				printDiagnosticResults(results, verbose)
			}

			// Print summary
			if !noColor {
				utils.Info("\n📊 Diagnostic Summary:")
				utils.Info("=====================")
			} else {
				utils.Info("\nDiagnostic Summary:")
				utils.Info("=====================")
			}

			if errors == 0 && warnings == 0 {
				if !noColor {
					utils.Success("All checks passed! Your Nix installation is healthy.")
				} else {
					fmt.Println("SUCCESS: All checks passed! Your Nix installation is healthy.")
				}
			} else {
				if errors > 0 {
					if !noColor {
						utils.Error("Found %d error(s) that need attention.", errors)
					} else {
						fmt.Printf("ERROR: Found %d error(s) that need attention.\n", errors)
					}
				}
				if warnings > 0 {
					if !noColor {
						utils.Warn("Found %d warning(s) that may need attention.", warnings)
					} else {
						fmt.Printf("WARNING: Found %d warning(s) that may need attention.\n", warnings)
					}
				}

				if fixIssues {
					attemptFixes(results, noColor)
				} else {
					if !noColor {
						utils.Tip("Run with '--fix' to attempt automatic fixes for common issues.")
					} else {
						fmt.Println("TIP: Run with '--fix' to attempt automatic fixes for common issues.")
					}
				}
			}

			utils.Debug("Diagnostics completed in %v", time.Since(startTime))
		}
	},
}

// printDiagnosticResults formats and prints diagnostic results with color
func printDiagnosticResults(results []utils.DoctorResult, verbose bool) {
	for i, result := range results {
		// Add extra newline between checks except for the first one
		if i > 0 {
			fmt.Println()
		}

		// Print check name and description
		utils.Info("🔍 %s:", result.Name)
		if verbose {
			utils.Debug("  Description: %s", result.Description)
		}

		// Print result based on status
		switch result.Status {
		case utils.StatusOK:
			utils.Success("  ✓ %s", result.Message)
		case utils.StatusWarning:
			utils.Warn("  ⚠ %s", result.Message)
			if result.Fix != "" {
				utils.Tip("  Suggestion: %s", result.Fix)
			}
		case utils.StatusError:
			utils.Error("  ✗ %s", result.Message)
			if result.Fix != "" {
				utils.Tip("  Fix: %s", result.Fix)
			}
		default:
			utils.Info("  ? %s", result.Message)
		}
	}
}

// printDiagnosticResultsNoColor formats and prints diagnostic results without color
func printDiagnosticResultsNoColor(results []utils.DoctorResult, verbose bool) {
	for i, result := range results {
		// Add extra newline between checks except for the first one
		if i > 0 {
			fmt.Println()
		}

		// Print check name and description
		fmt.Printf("%s:\n", result.Name)
		if verbose {
			fmt.Printf("  Description: %s\n", result.Description)
		}

		// Print result based on status
		switch result.Status {
		case utils.StatusOK:
			fmt.Printf("  [OK] %s\n", result.Message)
		case utils.StatusWarning:
			fmt.Printf("  [WARNING] %s\n", result.Message)
			if result.Fix != "" {
				fmt.Printf("  Suggestion: %s\n", result.Fix)
			}
		case utils.StatusError:
			fmt.Printf("  [ERROR] %s\n", result.Message)
			if result.Fix != "" {
				fmt.Printf("  Fix: %s\n", result.Fix)
			}
		default:
			fmt.Printf("  [UNKNOWN] %s\n", result.Message)
		}
	}
}

// formatDiagnosticsAsJSON formats diagnostic results as JSON
func formatDiagnosticsAsJSON(results []utils.DoctorResult, duration time.Duration) string {
	var jsonLines []string

	jsonLines = append(jsonLines, "{")
	jsonLines = append(jsonLines, fmt.Sprintf("  \"timestamp\": \"%s\",", time.Now().Format(time.RFC3339)))
	jsonLines = append(jsonLines, fmt.Sprintf("  \"duration_ms\": %d,", duration.Milliseconds()))
	jsonLines = append(jsonLines, fmt.Sprintf("  \"os\": \"%s\",", runtime.GOOS))
	jsonLines = append(jsonLines, fmt.Sprintf("  \"arch\": \"%s\",", runtime.GOARCH))
	jsonLines = append(jsonLines, "  \"results\": [")

	for i, result := range results {
		jsonLines = append(jsonLines, "    {")
		jsonLines = append(jsonLines, fmt.Sprintf("      \"name\": \"%s\",", escapeJSON(result.Name)))
		jsonLines = append(jsonLines, fmt.Sprintf("      \"description\": \"%s\",", escapeJSON(result.Description)))
		jsonLines = append(jsonLines, fmt.Sprintf("      \"status\": \"%s\",", result.Status))
		jsonLines = append(jsonLines, fmt.Sprintf("      \"message\": \"%s\",", escapeJSON(result.Message)))

		if result.Fix != "" {
			jsonLines = append(jsonLines, fmt.Sprintf("      \"fix\": \"%s\"", escapeJSON(result.Fix)))
		} else {
			jsonLines = append(jsonLines, "      \"fix\": null")
		}

		if i < len(results)-1 {
			jsonLines = append(jsonLines, "    },")
		} else {
			jsonLines = append(jsonLines, "    }")
		}
	}

	jsonLines = append(jsonLines, "  ]")
	jsonLines = append(jsonLines, "}")

	return strings.Join(jsonLines, "\n")
}

// escapeJSON escapes a string for JSON output
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// attemptFixes tries to fix detected issues
func attemptFixes(results []utils.DoctorResult, noColor bool) {
	if !noColor {
		utils.Info("\n🔧 Attempting to fix issues:")
		utils.Info("=========================")
	} else {
		utils.Info("\nAttempting to fix issues:")
		utils.Info("=========================")
	}

	fixedCount := 0

	// Try to fix Nix channel issues
	if err := utils.UpdateChannel(); err == nil {
		if !noColor {
			utils.Success("Updated Nix channel")
		} else {
			fmt.Println("SUCCESS: Updated Nix channel")
		}
		fixedCount++
	}

	// Ensure config directory exists
	if _, err := utils.EnsureConfigDir(); err == nil {
		if !noColor {
			utils.Success("Ensured configuration directory exists")
		} else {
			fmt.Println("SUCCESS: Ensured configuration directory exists")
		}
		fixedCount++
	}

	// Fix configuration issues
	if config, err := utils.LoadConfig(); err == nil {
		// Reset invalid settings to defaults
		if config.ChannelURL == "" {
			config.ChannelURL = "nixos-unstable"
			if !noColor {
				utils.Success("Reset channel URL to default")
			} else {
				fmt.Println("SUCCESS: Reset channel URL to default")
			}
			fixedCount++
		}
		if config.ShellFormat == "" {
			config.ShellFormat = "shell.nix"
			if !noColor {
				utils.Success("Reset shell format to default")
			} else {
				fmt.Println("SUCCESS: Reset shell format to default")
			}
			fixedCount++
		}
		if config.Pins == nil {
			config.Pins = make(map[string]string)
			if !noColor {
				utils.Success("Initialized package pins")
			} else {
				fmt.Println("SUCCESS: Initialized package pins")
			}
			fixedCount++
		}

		if err := utils.SaveConfig(config); err == nil {
			if !noColor {
				utils.Success("Saved fixed configuration")
			} else {
				fmt.Println("SUCCESS: Saved fixed configuration")
			}
		}
	}

	// For each error, check if we can fix it
	for _, result := range results {
		if result.Status == utils.StatusError || result.Status == utils.StatusWarning {
			if fixSpecificIssue(result, noColor) {
				fixedCount++
			}
		}
	}

	if fixedCount > 0 {
		if !noColor {
			utils.Success("Fixed %d issue(s). Run 'nsm doctor' again to verify.", fixedCount)
		} else {
			fmt.Printf("SUCCESS: Fixed %d issue(s). Run 'nsm doctor' again to verify.\n", fixedCount)
		}
	} else {
		if !noColor {
			utils.Warn("Could not automatically fix any issues.")
			utils.Tip("Please fix the issues manually following the suggestions above.")
		} else {
			fmt.Println("WARNING: Could not automatically fix any issues.")
			fmt.Println("TIP: Please fix the issues manually following the suggestions above.")
		}
	}
}

// fixSpecificIssue attempts to fix a specific issue based on its name and status
func fixSpecificIssue(result utils.DoctorResult, noColor bool) bool {
	switch result.Name {
	case "Project Files":
		// Create a default shell.nix if none exists
		if !utils.FileExists("shell.nix") && !utils.FileExists("flake.nix") {
			currentDir, err := os.Getwd()
			if err != nil {
				return false
			}
			if err := utils.GenerateShellNix(currentDir, []string{}); err == nil {
				if !noColor {
					utils.Success("Created default shell.nix file")
				} else {
					fmt.Println("SUCCESS: Created default shell.nix file")
				}
				return true
			}
		}
	case "Nix Store Permissions":
		// This usually requires root, so we just show a message
		if result.Status == utils.StatusError {
			if !noColor {
				utils.Warn("Store permission issues require manual intervention:")
				utils.Tip("  %s", result.Fix)
			} else {
				fmt.Println("WARNING: Store permission issues require manual intervention:")
				fmt.Printf("TIP: %s\n", result.Fix)
			}
		}
	case "Flakes Support":
		// If flakes are not supported, we provide instructions
		if !utils.CheckFlakeSupport() {
			nixConfDir := "~/.config/nix"
			if runtime.GOOS == "darwin" {
				nixConfDir = "/etc/nix"
			}
			if !noColor {
				utils.Tip("To enable flakes, add the following to %s/nix.conf:", nixConfDir)
				utils.Tip("  experimental-features = nix-command flakes")
				utils.Tip("Then restart the Nix daemon if using multi-user installation")
			} else {
				fmt.Printf("TIP: To enable flakes, add the following to %s/nix.conf:\n", nixConfDir)
				fmt.Println("TIP: experimental-features = nix-command flakes")
				fmt.Println("TIP: Then restart the Nix daemon if using multi-user installation")
			}
		}
	}

	return false
}

func init() {
	RootCmd.AddCommand(doctorCmd)

	// Add flags
	doctorCmd.Flags().BoolP("json", "j", false, "Output results in JSON format")
	doctorCmd.Flags().Bool("csv", false, "Output results in CSV format")
	doctorCmd.Flags().Bool("md", false, "Output results in Markdown format")
	doctorCmd.Flags().Bool("table", false, "Output results in table format (default)")
	doctorCmd.Flags().BoolP("fix", "f", false, "Attempt to fix detected issues")
	doctorCmd.Flags().BoolP("verbose", "v", false, "Show more detailed output")
	doctorCmd.Flags().Bool("no-color", false, "Disable colored output")
}
