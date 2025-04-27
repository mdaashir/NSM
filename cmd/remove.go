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

// removePackagesFromShellNix removes packages from shell.nix
func removePackagesFromShellNix(content string, toRemove map[string]bool) (string, int) {
	lines := strings.Split(content, "\n")
	result := []string{}
	inPackages := false
	removed := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "packages = with pkgs; [") {
			inPackages = true
			result = append(result, line)
			continue
		}
		if inPackages {
			if strings.Contains(trimmed, "];") {
				inPackages = false
				result = append(result, line)
				continue
			}
			pkgName := strings.Fields(trimmed)
			if len(pkgName) > 0 {
				name := pkgName[0]
				if toRemove[name] {
					removed++
					continue
				}
			}
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n"), removed
}

// removePackagesFromFlake removes packages from flake.nix
func removePackagesFromFlake(content string, toRemove map[string]bool) (string, int) {
	lines := strings.Split(content, "\n")
	result := []string{}
	inBuildInputs := false
	removed := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "buildInputs") {
			inBuildInputs = true
			result = append(result, line)
			continue
		}
		if inBuildInputs && strings.Contains(trimmed, "];") {
			inBuildInputs = false
			result = append(result, line)
			continue
		}
		if inBuildInputs {
			pkgName := strings.Fields(trimmed)
			if len(pkgName) > 0 {
				name := pkgName[0]
				if toRemove[name] {
					removed++
					continue
				}
			}
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n"), removed
}

var removeCmd = &cobra.Command{
	Use:   "remove [packages...]",
	Short: "Remove one or more packages from the nix environment",
	Long: `Remove packages from your Nix development environment.

This command will:
- Remove specified packages from shell.nix or flake.nix
- Keep the environment consistent
- Preserve other package configurations
- Handle multiple package removal safely

Usage:
  nsm remove <package1> [package2...]  # Remove one or more packages

Examples:
  nsm remove gcc              # Remove single package
  nsm remove python3 nodejs   # Remove multiple packages`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check for Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		// Create map of packages to remove
		toRemove := make(map[string]bool)
		for _, pkg := range args {
			toRemove[pkg] = true
		}

		configType := utils.GetProjectConfigType()
		if configType == "" {
			utils.Error("No shell.nix or flake.nix found")
			utils.Tip("Run 'nsm init' to create a new environment")
			return
		}

		utils.Debug("Found configuration file: %s", configType)

		// Read configuration file
		content, err := os.ReadFile(configType)
		if err != nil {
			utils.Error("Error reading %s: %v", configType, err)
			return
		}

		var newContent string
		var removed int

		if configType == "shell.nix" {
			newContent, removed = removePackagesFromShellNix(string(content), toRemove)
		} else {
			newContent, removed = removePackagesFromFlake(string(content), toRemove)
		}

		if removed == 0 {
			utils.Warn("No packages were found to remove")
			return
		}

		// Create backup before modifying
		if err := utils.BackupFile(configType); err != nil {
			utils.Error("Failed to create backup: %v", err)
			return
		}

		// Write changes
		if err := os.WriteFile(configType, []byte(newContent), 0644); err != nil {
			utils.Error("Error writing %s: %v", configType, err)
			return
		}

		utils.Success("Removed %d package(s) from %s", removed, configType)
		utils.Success("Backup created: %s.backup", configType)
		utils.Tip("Run 'nsm run' to enter the updated shell")
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
