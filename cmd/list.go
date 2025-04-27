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
