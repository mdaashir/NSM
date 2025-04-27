/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Enter the nix-shell / nix develop environment",
	Long: `Enter a Nix development environment based on your configuration.

The command automatically detects and uses the appropriate method:
- For shell.nix: Uses nix-shell
- For flake.nix: Uses nix develop

Examples:
  nsm run            # Enter the development environment

Inside the shell, you'll have access to all packages
specified in your configuration file.

Note: Make sure to run 'nsm init' first if you haven't
created a configuration file yet.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat("shell.nix"); err == nil {
			fmt.Println("🚀 Launching nix-shell...")
			c := exec.Command("nix-shell")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin

			err := c.Run()
			if err != nil {
				fmt.Println("❌ Error running nix-shell:", err)
			}
		} else if _, err := os.Stat("flake.nix"); err == nil {
			fmt.Println("🚀 Launching nix develop...")
			c := exec.Command("nix", "develop")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin

			err := c.Run()
			if err != nil {
				fmt.Println("❌ Error running nix develop:", err)
			}
		} else {
			fmt.Println("❌ No shell.nix or flake.nix found. Run 'NSM init' first!")
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
