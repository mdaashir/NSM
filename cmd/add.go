/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [packages...]",
	Short: "Add one or more packages to the nix environment",
	Long: `Add Nixpkgs packages to your development environment.

Usage:
  nsm add <package1> [package2...]  # Add one or more packages

Examples:
  nsm add gcc                     # Add single package
  nsm add python3 nodejs git      # Add multiple packages
  nsm add go rustc cargo         # Add development toolchains

The packages will be added to your shell.nix configuration.
Use 'nsm run' to enter the shell with the new packages.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileName := "shell.nix"

		// Read existing file
		data, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Println("❌ Error reading shell.nix:", err)
			return
		}

		content := string(data)

		// Find insertion point (before closing bracket ])
		pos := strings.Index(content, "];")
		if pos == -1 {
			fmt.Println("❌ Could not find package list in shell.nix")
			return
		}

		// Build the new packages
		newPackages := ""
		for _, pkg := range args {
			newPackages += "    " + pkg + "\n"
		}

		// Insert new packages
		newContent := content[:pos] + newPackages + content[pos:]

		// Write back
		err = os.WriteFile(fileName, []byte(newContent), 0644)
		if err != nil {
			fmt.Println("❌ Error writing to shell.nix:", err)
			return
		}

		fmt.Println("✅ Added package(s):", strings.Join(args, ", "))
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
