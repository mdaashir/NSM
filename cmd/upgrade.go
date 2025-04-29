/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"os/exec"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
)

// Interactive flag for upgrade command
var upgradeInteractive bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Update nixpkgs channel",
	Long: `Update your Nixpkgs channel to the latest version.

This command will:
- Update your configured Nixpkgs channel
- Fetch the latest package definitions
- Ensure access to the newest packages
- Maintain channel consistency

Example:
  nsm upgrade    # Update nixpkgs to latest version

Note: After upgrading, you may need to rebuild your
environment by running 'nsm run' again.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		// Get current channel info for comparison
		oldChannel, err := utils.GetChannelInfo()
		if err != nil {
			utils.Error("Could not get current channel info: %v", err)
			return
		}

		utils.Info("🔄 Updating nixpkgs channel...")

		// Run nix-channel --update
		c := exec.Command("nix-channel", "--update")
		output, err := c.CombinedOutput()
		if err != nil {
			utils.Error("Failed to update nixpkgs: %v", err)
			utils.Tip("Try running 'nsm doctor' to check your installation")
			return
		}

		// Get new channel info
		newChannel, err := utils.GetChannelInfo()
		if err != nil {
			utils.Error("Could not get updated channel info: %v", err)
			return
		}

		utils.Success("Updated nixpkgs channel!")
		if len(output) > 0 {
			utils.Debug("Update details:\n%s", string(output))
		}

		if oldChannel != newChannel {
			utils.Info("Channel changed from:\n%s\nto:\n%s", oldChannel, newChannel)
		}

		// Interactive workflow
		if upgradeInteractive {
			utils.Tip("Run 'nsm run' to enter shell with updated packages")
			if utils.PromptContinue("enter the shell") {
				runCmd.Run(cmd, args)
			}
		} else {
			utils.Tip("Run 'nsm run' to enter shell with updated packages")
		}
	},
}

func init() {
	RootCmd.AddCommand(upgradeCmd)

	// Add interactive flag
	upgradeCmd.Flags().BoolVarP(&upgradeInteractive, "interactive", "i", false, "Run in interactive mode")
}
