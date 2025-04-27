/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "NSM",
	Short: "NSM (Nix Shell Manager) - A tool to manage Nix development environments",
	Long: `NSM (Nix Shell Manager) is a powerful CLI tool that helps you manage Nix development environments.

Features:
- Initialize new Nix shell environments
- Add/remove packages to your environment
- List installed packages
- Convert between shell.nix and flake.nix
- Run Nix shells
- Manage Nix channel versions
- Clean up unused packages

Example Usage:
  nsm init              # Initialize a new shell.nix
  nsm add gcc python3   # Add packages
  nsm list              # List installed packages
  nsm run              # Enter the Nix shell
  nsm clean            # Clean up unused packages`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.NSM.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
