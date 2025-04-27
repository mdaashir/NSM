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
