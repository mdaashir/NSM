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

		var packages []string
		if configType == "shell.nix" {
			packages, err = utils.ExtractShellNixPackages(configType)
		} else {
			packages, err = utils.ExtractFlakePackages(configType)
		}

		if err != nil {
			utils.Error("Failed to extract packages from %s: %v", configType, err)
			return
		}

		if len(packages) == 0 {
			utils.Info("No packages found in %s", configType)
			return
		}

		// Sort packages alphabetically
		sort.Strings(packages)

		// Create table
		table := utils.NewTable([]string{"Package", "Status", "Source"})

		for _, pkg := range packages {
			status := "pending"
			if installedPkgs[pkg] {
				status = "installed"
			}

			table.AddRow([]string{pkg, status, configType})
		}

		// Output as a table
		utils.Info("\n📦 Packages in your Nix environment:")
		utils.Info("\n%s", table.String())

		utils.Info("\nTotal packages: %d", len(packages))
		utils.Info("Configuration: %s", configType)

		// Show tips based on package status
		pendingCount := len(packages) - len(installedPkgs)
		if pendingCount > 0 {
			utils.Tip("Run 'nsm run' to enter shell with all packages")
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
