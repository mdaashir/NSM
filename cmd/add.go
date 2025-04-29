/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"os"
	"strings"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [packages...]",
	Short: "Add one or more packages to the nix environment",
	Long: `Add Nixpkgs packages to your development environment.

Usage:
  nsm add <package1> [package2...]  # Add one or more packages

Examples:
  nsm add gcc                     # Add single package
  nsm add python3 nodejs git      # Add multiple packages
  nsm add go rustc cargo         # Add development toolchains`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check for Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		configType := utils.GetProjectConfigType()
		if configType == "" {
			utils.Error("No shell.nix or flake.nix found")
			utils.Tip("Run 'nsm init' to create a new environment")
			return
		}

		utils.Debug("Found configuration file: %s", configType)

		// Validate packages
		var invalidPkgs []string
		for _, pkg := range args {
			if !utils.ValidatePackage(pkg) {
				invalidPkgs = append(invalidPkgs, pkg)
			}
		}

		if len(invalidPkgs) > 0 {
			utils.Error("Invalid package(s): %s", strings.Join(invalidPkgs, ", "))
			utils.Tip("Check package names in https://search.nixos.org")
			return
		}

		// Create backup before modifying
		if err := utils.BackupFile(configType); err != nil {
			utils.Error("Failed to create backup: %v", err)
			return
		}

		// Read an existing file
		content, err := utils.ReadFile(configType)
		if err != nil {
			utils.Error("Error reading %s: %v", configType, err)
			return
		}

		// Find an insertion point based on a file type
		var pos int
		if configType == "shell.nix" {
			pos = strings.Index(content, "];")
		} else {
			pos = strings.Index(content, "];") // For flake.nix, find the buildInputs closure
		}

		if pos == -1 {
			utils.Error("Could not find package list in %s", configType)
			utils.Tip("Run 'nsm init' to create a properly formatted file")
			return
		}

		// Check for duplicate packages
		var duplicates []string
		currentContent := content[:pos]
		for _, pkg := range args {
			if strings.Contains(currentContent, pkg) {
				duplicates = append(duplicates, pkg)
			}
		}

		if len(duplicates) > 0 {
			utils.Warn("Package(s) already installed: %s", strings.Join(duplicates, ", "))
			return
		}

		// Build the new packages section
		newPackages := ""
		for _, pkg := range args {
			newPackages += "    " + pkg + "\n"
		}

		// Insert new packages
		newContent := content[:pos] + newPackages + content[pos:]

		// Write back with secure permissions
		err = os.WriteFile(configType, []byte(newContent), 0600)
		if err != nil {
			utils.Error("Error writing to %s: %v", configType, err)
			return
		}

		utils.Success("Added package(s): %s", strings.Join(args, ", "))
		utils.Tip("Run 'nsm run' to enter the shell with new packages")
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
}
