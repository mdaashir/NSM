/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"os"
	"os/exec"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Enter the nix-shell / nix develop environment",
	Long: `Enter a Nix development environment based on your configuration.

The command automatically detects and uses the appropriate method:
- For shell.nix: Uses nix-shell
- For flake.nix: Uses nix develop

Options:
  --pure    Run in pure mode (no inherited environment)

Examples:
  nsm run            # Enter the development environment
  nsm run --pure    # Enter a pure shell`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		configType := utils.GetProjectConfigType()
		if configType == "" {
			utils.Error("No shell.nix or flake.nix found")
			utils.Tip("Run 'nsm init' to create a new environment")
			return
		}

		utils.Debug("Using configuration file: %s", configType)

		isPure, err := cmd.Flags().GetBool("pure")
		if err != nil {
			utils.Error("Failed to get pure flag: %v", err)
			return
		}

		if isPure {
			utils.Debug("Running in pure mode")
		}

		var c *exec.Cmd
		if configType == "shell.nix" {
			utils.Info("🚀 Launching nix-shell...")
			var cmdArgs []string
			if isPure {
				cmdArgs = append(cmdArgs, "--pure")
			}
			c = exec.Command("nix-shell", cmdArgs...)
		} else {
			utils.Info("🚀 Launching nix develop...")
			if !utils.CheckFlakeSupport() {
				utils.Error("Flakes are not enabled in your Nix configuration")
				utils.Tip("Add 'experimental-features = nix-command flakes' to your Nix config")
				return
			}
			cmdArgs := []string{"develop"}
			if isPure {
				cmdArgs = append(cmdArgs, "--pure")
			}
			c = exec.Command("nix", cmdArgs...)
		}

		// Setup command environment
		c.Env = os.Environ()
		currentDir, err := os.Getwd()
		if err != nil {
			utils.Error("Failed to get current directory: %v", err)
			return
		}
		c.Dir = currentDir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		// Run the command
		err = c.Run()
		if err != nil {
			utils.Error("Error running %s: %v", configType, err)
			utils.Tip("Try running 'nsm doctor' to diagnose issues")
			return
		}
	},
}

func init() {
	runCmd.Flags().Bool("pure", false, "Run in pure mode (no inherited environment)")
	rootCmd.AddCommand(runCmd)
}
