/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
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
		fmt.Println("🔄 Updating nixpkgs channel...")

		// Run `nix-channel --update`
		c := exec.Command("nix-channel", "--update")
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.ErrOrStderr()

		err := c.Run()
		if err != nil {
			fmt.Println("❌ Error updating nixpkgs:", err)
			return
		}

		fmt.Println("✅ nixpkgs channel updated!")
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
