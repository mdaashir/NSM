/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
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
  nsm list              # List all packages
  nsm list --json      # Output in JSON format
  nsm list --installed # Show only installed packages`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		// Get configuration type
		configType := utils.GetProjectConfigType()
		if configType == "" {
			utils.Error("No shell.nix or flake.nix found in current directory")
			utils.Tip("Run 'nsm init' to create a new environment")
			return
		}

		// Get installed packages
		installedPkgs := make(map[string]bool)
		cmd := exec.Command("nix-env", "--query", "--installed")
		if output, err := cmd.Output(); err == nil {
			for _, line := range strings.Split(string(output), "\n") {
				if pkg := strings.TrimSpace(line); pkg != "" {
					installedPkgs[pkg] = true
				}
			}
		}

		// Read configuration file
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

		// Prepare table data
		headers := []string{"Package", "Status", "Source"}
		var rows [][]string

		for _, pkg := range packages {
			status := "pending"
			if installedPkgs[pkg] {
				status = "installed"
			}

			rows = append(rows, []string{
				pkg,
				status,
				configType,
			})
		}

		// Output as table
		utils.Info("\n📦 Packages in your Nix environment:")
		utils.Table(headers, rows)

		utils.Info("\nTotal packages: %d", len(packages))
		utils.Info("Configuration: %s", configType)

		// Show tips based on package status
		pendingCount := 0
		for _, row := range rows {
			if row[1] == "pending" {
				pendingCount++
			}
		}

		if pendingCount > 0 {
			utils.Tip("Run 'nsm run' to enter shell with all packages")
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
