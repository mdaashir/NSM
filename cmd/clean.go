/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"os/exec"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up unused nix packages and free space",
	Long: `Clean up your Nix store by removing unused packages and dependencies.

This command runs nix-collect-garbage with the -d flag to:
- Remove unused packages from the Nix store
- Delete old generations of profiles
- Free up disk space
- Remove obsolete dependencies

Example:
  nsm clean    # Clean up unused packages

Note: This operation is safe but irreversible. Make sure
you don't need old generations before cleaning.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		utils.Info("🧹 Running garbage collection...")

		// Run nix-collect-garbage
		c := exec.Command("nix-collect-garbage", "-d")
		output, err := c.CombinedOutput()
		if err != nil {
			utils.Error("Failed to clean packages: %v", err)
			utils.Tip("Try running 'nsm doctor' to check your installation")
			return
		}

		utils.Success("Cleaned up Nix store successfully!")
		if len(output) > 0 {
			utils.Debug("Cleanup details:\n%s", string(output))
		}
		utils.Tip("Run 'nsm info' to check current system state")
	},
}

func init() {
	RootCmd.AddCommand(cleanCmd)
}
