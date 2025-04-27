/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>

*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"regexp"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a shell.nix to flake.nix, migrating packages if any",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if shell.nix exists
		if _, err := os.Stat("shell.nix"); os.IsNotExist(err) {
			fmt.Println("❌ No shell.nix found in the current directory.")
			return
		}

		// Read the content of shell.nix
		content, err := os.ReadFile("shell.nix")
		if err != nil {
			fmt.Println("❌ Error reading shell.nix:", err)
			return
		}

		// Convert content to a string
		shellNixContent := string(content)

		// Extract packages using regex (improved for your format)
		var packages []string
		re := regexp.MustCompile(`packages\s*=\s*with\s*pkgs;\s*\[([^\]]+)\]`)
		matches := re.FindStringSubmatch(shellNixContent)
		if len(matches) > 1 {
			// Split by spaces or newlines to extract individual packages
			packages = append(packages, strings.Fields(matches[1])...)
		}

		// If no packages found, inform the user
		if len(packages) == 0 {
			fmt.Println("❌ No packages found in shell.nix. No migration necessary.")
			return
		}

		// Create the flake.nix content with migrated packages
		flakeContent := fmt.Sprintf(`{
  description = "A flake for my project";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = nixpkgs.mkShell {
      buildInputs = [ %s ];
    };
  };
}
`, strings.Join(packages, " "))

		// Write to flake.nix
		err = os.WriteFile("flake.nix", []byte(flakeContent), 0644)
		if err != nil {
			fmt.Println("❌ Error writing flake.nix:", err)
			return
		}

		fmt.Println("✅ shell.nix successfully converted and migrated to flake.nix")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
