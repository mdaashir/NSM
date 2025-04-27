/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new nix environment",
	Long: `Initialize a new Nix development environment by creating a shell.nix file.

The generated shell.nix will be set up with a basic structure ready for package management.
After initialization, you can:
- Add packages using 'nsm add <package>'
- Enter the shell using 'nsm run'
- Convert to flake.nix using 'nsm convert'

Example:
  nsm init            # Create new shell.nix
  nsm add gcc        # Add a package
  nsm run           # Enter the shell`,
	Run: func(cmd *cobra.Command, args []string) {
		content := `{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    # Add your packages here
  ];
}`

		err := os.WriteFile("shell.nix", []byte(content), 0644)
		if err != nil {
			fmt.Println("Error creating shell.nix:", err)
			return
		}

		fmt.Println("✅ Initialized new nix environment (shell.nix created)!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
