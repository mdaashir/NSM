/*
Copyright © 2025 Mohamed Aashir s <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all packages in the nix environment",
	Long: `Display all packages currently configured in your Nix environment.

This command shows:
- All packages defined in shell.nix
- Their current status in the environment
- Package names exactly as they appear in nixpkgs

Example:
  nsm list    # Show all configured packages

The output is formatted for easy reading and shows packages
in the same format they should be used with 'nsm add'.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileName := "shell.nix"

		data, err := os.Open(fileName)
		if err != nil {
			fmt.Println("❌ Error opening shell.nix:", err)
			return
		}
		defer data.Close()

		scanner := bufio.NewScanner(data)
		inPackages := false
		packages := []string{}

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, "packages = with pkgs; [") {
				inPackages = true
				continue
			}
			if inPackages {
				if strings.Contains(line, "];") {
					break
				}
				if line != "" {
					packages = append(packages, line)
				}
			}
		}

		if len(packages) == 0 {
			fmt.Println("ℹ️  No packages found.")
			return
		}

		fmt.Println("📦 Installed packages:")
		for _, pkg := range packages {
			fmt.Println("- " + pkg)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
