/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>

*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// pinCmd represents the pin command
var pinCmd = &cobra.Command{
	Use:   "pin",
	Short: "Pin a specific nixpkgs commit version",
	Run: func(cmd *cobra.Command, args []string) {
		// Run nix-channel --list to get the current channel URL
		c := exec.Command("nix-channel", "--list")
		output, err := c.Output()
		if err != nil {
			fmt.Println("❌ Error fetching nix channel:", err)
			return
		}

		// Print current channel URL
		fmt.Printf("✅ Current nixpkgs channel: %s\n", output)
	},
}

func init() {
	rootCmd.AddCommand(pinCmd)
}
