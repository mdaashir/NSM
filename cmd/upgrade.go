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
