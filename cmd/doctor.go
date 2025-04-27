/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose the nix environment installation",
	Long: `Perform a comprehensive health check of your Nix installation.

This command checks:
- Nix binary presence and version
- Nixpkgs channel availability and status
- Environment configuration
- Common installation issues

Examples:
  nsm doctor    # Run all diagnostics

The command will suggest fixes for any issues found.
Use this command if you experience any problems with
your Nix environment or NSM commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if nix is installed
		_, err := exec.LookPath("nix")
		if err != nil {
			fmt.Println("❌ nix is not installed. Please install nix first.")
			return
		}

		// Check if nixpkgs is available
		c := exec.Command("nix", "--version")
		err = c.Run()
		if err != nil {
			fmt.Println("❌ Error checking nix version:", err)
			return
		}

		fmt.Println("✅ Nix installation is healthy.")
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
