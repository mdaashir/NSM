/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init [--flake]",
	Short: "Initialize a new nix environment",
	Long: `Initialize a new Nix development environment.

This command will:
1. Create a new configuration file (shell.nix or flake.nix)
2. Include any default packages from your config
3. Set up the environment ready for use

Options:
  --flake    Create a flake.nix instead of shell.nix
  --force    Overwrite existing configuration files

Examples:
  nsm init            # Create new shell.nix
  nsm init --flake   # Create new flake.nix
  nsm init --force   # Overwrite existing files`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		useFlake, err := cmd.Flags().GetBool("flake")
		if err != nil {
			utils.Error("Failed to get flake flag: %v", err)
			return
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			utils.Error("Failed to get force flag: %v", err)
			return
		}

		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			utils.Debug("Failed to get interactive flag, defaulting to false: %v", err)
			interactive = false
		}

		// Determine a file to create
		var filename string
		if useFlake {
			filename = "flake.nix"
			if !utils.CheckFlakeSupport() {
				utils.Error("Flakes are not enabled in your Nix configuration")
				utils.Tip("Add 'experimental-features = nix-command flakes' to your Nix config")
				return
			}
		} else {
			filename = "shell.nix"
		}

		// Check if files already exist
		if !force {
			if utils.FileExists(filename) {
				utils.Error("%s already exists. Use --force to overwrite", filename)
				return
			}
		} else {
			utils.Debug("Force flag enabled, will overwrite existing files")
		}

		// Create a backup if a file exists and force is enabled
		if force && utils.FileExists(filename) {
			if err := utils.BackupFile(filename); err != nil {
				utils.Error("Failed to create backup: %v", err)
				return
			}
			utils.Success("Created backup: %s.backup", filename)
		}

		// Generate content
		var content string
		if useFlake {
			content = getDefaultFlakeContent()
		} else {
			content = getDefaultShellContent()
		}

		// Write the file
		err = os.WriteFile(filename, []byte(content), 0600)
		if err != nil {
			utils.Error("Failed to create %s: %v", filename, err)
			return
		}

		utils.Success("Created %s with default configuration", filename)

		// Interactive workflow
		if interactive {
			// Ask if user wants to add packages
			if utils.PromptContinue("add packages") {
				// Execute add command
				addCmd.Run(addCmd, []string{})
			}

			// Ask if user wants to run the shell
			if utils.PromptContinue("enter the shell") {
				// Execute run command
				runCmd.Run(runCmd, []string{})
			}
		} else {
			if useFlake {
				utils.Tip("Run 'nsm run' to enter the flake-based shell")
			} else {
				utils.Tip("Run 'nsm run' to enter the shell")
			}
			utils.Tip("Use --interactive flag for an interactive workflow")
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().Bool("flake", false, "Create a flake.nix instead of shell.nix")
	initCmd.Flags().Bool("force", false, "Overwrite existing configuration files")
	initCmd.Flags().Bool("interactive", false, "Enable interactive workflow")
}

// getDefaultShellContent generates shell.nix content with configured defaults
func getDefaultShellContent() string {
	defaultPkgs := viper.GetStringSlice("default.packages")
	pkgList := ""
	for _, pkg := range defaultPkgs {
		if utils.ValidatePackage(pkg) {
			pkgList += "    " + pkg + "\n"
		}
	}

	return fmt.Sprintf(`{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  # Shell name for better identification
  name = "dev-shell";

  # Packages from nixpkgs
  packages = with pkgs; [
%s  ];

  # Shell hook for environment setup
  shellHook = ''
    echo "🚀 Welcome to your Nix development environment!"
    echo "📦 Use 'nsm add <package>' to add more packages"
  '';
}`, pkgList)
}

// getDefaultFlakeContent generates flake.nix content with configured defaults
func getDefaultFlakeContent() string {
	defaultPkgs := viper.GetStringSlice("default.packages")
	var validPkgs []string
	for _, pkg := range defaultPkgs {
		if utils.ValidatePackage(pkg) {
			validPkgs = append(validPkgs, pkg)
		}
	}
	pkgList := strings.Join(validPkgs, "\n      ")

	channel := viper.GetString("channel.url")
	if channel == "" {
		channel = "nixos-unstable"
	}

	return fmt.Sprintf(`{
  description = "Development environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/%s";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system: {
      devShell = nixpkgs.legacyPackages.${system}.mkShell {
        name = "dev-shell";

        buildInputs = with nixpkgs.legacyPackages.${system}; [
      %s
        ];

        shellHook = '''
          echo "🚀 Welcome to your Nix development environment!"
          echo "📦 Use 'nsm add <package>' to add more packages"
        ''';
      };
    });
}`, channel, pkgList)
}
