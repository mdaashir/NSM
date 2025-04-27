/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system and nix information",
	Long: `Display detailed information about your Nix installation and system.

Information shown:
- Nix version and installation type
- System architecture and OS details
- Active Nixpkgs channel
- Environment status

Example:
  nsm info    # Show system information

This information is useful for:
- Troubleshooting issues
- Reporting bugs
- Checking compatibility`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show Nix version
		c := exec.Command("nix", "--version")
		output, err := c.Output()
		if err != nil {
			fmt.Println("❌ Error fetching Nix version:", err)
			return
		}
		fmt.Printf("✅ Nix Version: %s\n", output)

		// Show OS information
		c = exec.Command("uname", "-a")
		output, err = c.Output()
		if err != nil {
			fmt.Println("❌ Error fetching OS info:", err)
			return
		}
		fmt.Printf("✅ OS Info: %s\n", output)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
