/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
)

var freezeCmd = &cobra.Command{
	Use:   "freeze",
	Short: "Freeze current package versions",
	Long: `Freeze the current versions of all installed packages.
This creates a lock file that can be used to reproduce the exact
same environment later.

The lock file contains:
- Package versions
- Channel information
- Nixpkgs revision
- Shell configuration type

Examples:
  nsm freeze              # Create/update lock file
  nsm freeze --json      # Output in JSON format`,
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

		// Get installed packages and their versions
		packages, err := utils.GetInstalledPackages()
		if err != nil {
			utils.Error("Failed to get installed packages: %v", err)
			return
		}

		lockData := make(map[string]interface{})
		packageVersions := make(map[string]string)

		for _, pkg := range packages {
			version, err := utils.GetPackageVersion(pkg)
			if err != nil {
				utils.Warn("Could not get version for %s: %v", pkg, err)
				continue
			}
			packageVersions[pkg] = version
		}

		// Get channel and revision info
		channel, err := utils.GetChannelInfo()
		if err != nil {
			utils.Warn("Could not get channel info: %v", err)
		}

		revision, err := utils.GetNixpkgsRevision()
		if err != nil {
			utils.Warn("Could not get nixpkgs revision: %v", err)
		}

		// Build lock data
		lockData["packages"] = packageVersions
		lockData["channel"] = channel
		lockData["nixpkgs_revision"] = revision
		lockData["config_type"] = configType
		lockData["version"] = "1.0.0"

		// Convert to JSON
		lockContent, err := json.MarshalIndent(lockData, "", "  ")
		if err != nil {
			utils.Error("Failed to create lock file content: %v", err)
			return
		}

		// Write a lock file
		lockFile := "nsm.lock.json"
		if err := os.WriteFile(lockFile, lockContent, 0600); err != nil {
			utils.Error("Failed to write lock file: %v", err)
			return
		}

		utils.Success("Created lock file: %s", lockFile)
		utils.Info("Found %d packages", len(packageVersions))

		if jsonOutput, _ := cmd.Flags().GetBool("json"); jsonOutput {
			fmt.Println(string(lockContent))
			return
		}

		// Show summary
		utils.Info("\n📦 Package versions:")
		for pkg, version := range packageVersions {
			utils.Info("  %s: %s", pkg, version)
		}

		utils.Info("\nChannel: %s", channel)
		utils.Info("Nixpkgs revision: %s", revision)
		utils.Tip("Use 'nsm pin' to restore these exact versions later")
	},
}

func init() {
	RootCmd.AddCommand(freezeCmd)
}
