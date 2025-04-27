/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
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
		// Run nix-collect-garbage
		c := exec.Command("nix-collect-garbage", "-d")
		output, err := c.CombinedOutput()
		if err != nil {
			fmt.Println("❌ Error cleaning nix packages:", err)
			return
		}

		fmt.Println("✅ Cleaned up nix packages!")
		fmt.Printf("Details: %s\n", output)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
