/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [packages]",
	Short: "Remove one or more packages from the nix environment",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileName := "shell.nix"

		data, err := os.Open(fileName)
		if err != nil {
			fmt.Println("❌ Error opening shell.nix:", err)
			return
		}
		defer data.Close()

		scanner := bufio.NewScanner(data)
		lines := []string{}
		inPackages := false

		toRemove := make(map[string]bool)
		for _, pkg := range args {
			toRemove[pkg] = true
		}

		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			if strings.Contains(trimmed, "packages = with pkgs; [") {
				inPackages = true
				lines = append(lines, line)
				continue
			}
			if inPackages {
				if strings.Contains(trimmed, "];") {
					inPackages = false
					lines = append(lines, line)
					continue
				}
				// Skip package if it matches any argument
				pkgName := strings.Fields(trimmed)
				if len(pkgName) > 0 {
					name := pkgName[0]
					if toRemove[name] {
						continue // skip this package
					}
				}
				lines = append(lines, line)
			} else {
				lines = append(lines, line)
			}
		}

		// Rewrite the file
		err = os.WriteFile(fileName, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			fmt.Println("❌ Error writing to shell.nix:", err)
			return
		}

		fmt.Println("✅ Removed package(s):", strings.Join(args, ", "))
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
