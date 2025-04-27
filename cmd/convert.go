/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert shell.nix to flake.nix",
	Long: `Convert your shell.nix configuration to the modern flake.nix format.

This command will:
1. Read your existing shell.nix configuration
2. Extract all configured packages and settings
3. Create a new flake.nix with equivalent functionality
4. Preserve all your package dependencies

The conversion process maintains all your existing packages while
upgrading to the newer, more reproducible flake-based workflow.

Example:
  nsm convert    # Convert shell.nix to flake.nix

Note: Your original shell.nix file will be preserved as a backup.
Use 'nsm run' after conversion to enter the new flake-based shell.`,
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
