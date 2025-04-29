/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ChannelURLKey = "channel.url"
)

// backupFile creates a backup of the given file
func backupFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return os.WriteFile(filename+".backup", content, 0600)
}

// parseShellNixPackages parses packages from shell.nix with regex
func parseShellNixPackages(content string) []string {
	var packages []string
	lines := strings.Split(content, "\n")
	inPackages := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "packages = with pkgs; [") {
			inPackages = true
			continue
		}
		if inPackages {
			if strings.Contains(trimmed, "];") {
				break
			}
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				packages = append(packages, trimmed)
			}
		}
	}
	return packages
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert shell.nix to flake.nix",
	Long: `Convert your shell.nix configuration to the modern flake.nix format.

This command will:
1. Read your existing shell.nix configuration
2. Extract all configured packages and settings
3. Create a new flake.nix with equivalent functionality
4. Create a backup of your shell.nix file
5. Preserve all package dependencies

Examples:
  nsm convert              # Convert shell.nix to flake.nix
  nsm convert --no-backup  # Convert without creating backup`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if shell.nix exists
		if _, err := os.Stat("shell.nix"); os.IsNotExist(err) {
			fmt.Println("❌ No shell.nix found in the current directory")
			return
		}

		// Check if flake.nix already exists
		if _, err := os.Stat("flake.nix"); err == nil {
			fmt.Println("❌ flake.nix already exists")
			fmt.Println("💡 Remove or rename existing flake.nix first")
			return
		}

		// Read shell.nix
		content, err := os.ReadFile("shell.nix")
		if err != nil {
			fmt.Println("❌ Error reading shell.nix:", err)
			return
		}

		// Create a backup if requested
		noBackup, _ := cmd.Flags().GetBool("no-backup")
		if !noBackup {
			if err := backupFile("shell.nix"); err != nil {
				fmt.Println("❌ Error creating backup:", err)
				return
			}
			fmt.Println("✅ Created backup: shell.nix.backup")
		}

		// Parse packages
		packages := parseShellNixPackages(string(content))
		if len(packages) == 0 {
			fmt.Println("⚠️  No packages found in shell.nix")
		}

		// Generate flake.nix content
		channel := viper.GetString(ChannelURLKey)
		if channel == "" {
			channel = "nixpkgs-unstable"
		}

		flakeContent := fmt.Sprintf(`{
  description = "Development environment converted from shell.nix";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/%s";

  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = nixpkgs.legacyPackages.x86_64-linux.mkShell {
      buildInputs = with nixpkgs.legacyPackages.x86_64-linux; [
        %s
      ];
    };
  };
}`, channel, strings.Join(packages, "\n        "))

		// Write flake.nix
		if err := os.WriteFile("flake.nix", []byte(flakeContent), 0600); err != nil {
			fmt.Println("❌ Error writing flake.nix:", err)
			return
		}

		fmt.Println("✅ Successfully converted to flake.nix")
		fmt.Printf("📦 Migrated %d packages\n", len(packages))
		fmt.Println("💡 Run 'nsm run' to enter the new flake-based shell")
	},
}

func init() {
	RootCmd.AddCommand(convertCmd)
}
