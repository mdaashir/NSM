/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List packages in the current environment",
	Long: `List all packages defined in your Nix environment.

This command will show:
- Package name and version
- Package status (installed/pending)
- Package description
- Installation source (shell.nix/flake.nix)

Examples:
  nsm list                  # List all packages
  nsm list --format json    # Output as JSON
  nsm list --format csv     # Output as CSV
  nsm list --no-color       # Disable colored output`,
	Run: func(cmd *cobra.Command, args []string) {
		formatName, _ := cmd.Flags().GetString("format")
		outputFormat, err := utils.ParseTableFormat(formatName)
		if err != nil {
			utils.Error("%v", err)
			return
		}

		noColor, _ := cmd.Flags().GetBool("no-color")

		// Check for Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		// Get a configuration type
		configType := utils.GetProjectConfigType()
		if configType == "" {
			utils.Error("No shell.nix or flake.nix found in current directory")
			utils.Tip("Run 'nsm init' to create a new environment")
			return
		}

		// Get installed packages
		installedPkgs := make(map[string]bool)
		nixEnvCmd := exec.Command("nix-env", "--query", "--installed")
		output, err := nixEnvCmd.Output()
		if err != nil {
			utils.Debug("Could not query installed packages: %v", err)
		} else {
			for _, line := range strings.Split(string(output), "\n") {
				if pkg := strings.TrimSpace(line); pkg != "" {
					installedPkgs[pkg] = true
				}
			}
		}

		// Read a configuration file
		content, err := utils.ReadFile(configType)
		if err != nil {
			utils.Error("Failed to read %s: %v", configType, err)
			return
		}

		var packages []string
		if configType == "shell.nix" {
			packages = utils.ExtractShellNixPackages(content)
		} else {
			packages = utils.ExtractFlakePackages(content)
		}

		if len(packages) == 0 {
			utils.Info("No packages found in %s", configType)
			return
		}

		// Sort packages alphabetically
		sort.Strings(packages)

		// Create table
		table := utils.NewTableRenderer([]string{"Package", "Status", "Source"})
		table.SetFormat(outputFormat)
		if noColor || outputFormat != utils.FormatText {
			table.DisableColors()
		}

		pendingCount := 0
		for _, pkg := range packages {
			status := "pending"
			if installedPkgs[pkg] {
				status = "installed"
			} else {
				pendingCount++
			}

			table.AddRow([]string{pkg, status, configType})
		}

		if outputFormat == utils.FormatText {
			utils.Info("\n📦 Packages in your Nix environment:")
			fmt.Print(table.String())

			utils.Info("\nTotal packages: %d", len(packages))
			utils.Info("Configuration: %s", configType)

			if pendingCount > 0 {
				utils.Tip("Run 'nsm run' to enter shell with all packages")
			}
			return
		}

		fmt.Print(table.String())
	},
}

func init() {
	listCmd.Flags().String("format", "text", "output format: text, markdown, json, csv")
	listCmd.Flags().Bool("no-color", false, "disable colored output")
	rootCmd.AddCommand(listCmd)
}
