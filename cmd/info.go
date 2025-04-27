/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"os"
	"os/exec"
	"strings"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system and nix information",
	Long: `Display detailed information about your Nix installation and system.

Information shown:
- Nix version and installation type
- System architecture and OS details
- Active Nixpkgs channel
- Environment status
- Flakes support status
- Current project configuration

Example:
  nsm info    # Show detailed system information`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Info("📊 System Information:")
		utils.Info("==================")

		// Check Nix installation
		if err := utils.CheckNixInstallation(); err != nil {
			utils.Error("Nix is not installed. Please install Nix first!")
			return
		}

		// Show a Nix version
		if version, err := utils.GetNixVersion(); err == nil {
			utils.Success("Nix Version: %s", version)
		} else {
			utils.Error("Could not determine Nix version: %v", err)
		}

		// Show channel information
		if channel, err := utils.GetChannelInfo(); err == nil {
			utils.Success("Channel Info: %s", channel)
		} else {
			utils.Error("Could not get channel info: %v", err)
		}

		// Check flakes support
		if utils.CheckFlakeSupport() {
			utils.Success("Flakes: Supported")
		} else {
			utils.Warn("Flakes: Not enabled")
			utils.Tip("To enable flakes, add 'experimental-features = nix-command flakes' to your Nix config")
		}

		// Show OS information
		if c := exec.Command("uname", "-a"); c != nil {
			if output, err := c.Output(); err == nil {
				utils.Success("OS Info: %s", output)
			}
		}

		// Show the current directory configuration
		utils.Info("\n📁 Project Configuration:")
		utils.Info("=====================")

		configType := utils.GetProjectConfigType()
		switch configType {
		case "shell.nix":
			utils.Success("Configuration: Traditional Nix shell (shell.nix)")
			if content, err := os.ReadFile("shell.nix"); err == nil {
				pkgCount := strings.Count(string(content), "\n    ")
				utils.Info("📦 Packages configured: %d", pkgCount)
			}
		case "flake.nix":
			utils.Success("Configuration: Nix Flake (flake.nix)")
		case "":
			utils.Warn("No Nix configuration found")
			utils.Tip("Run 'nsm init' to create a new environment")
		}

		if utils.FileExists(".envrc") {
			utils.Success("direnv: Configured")
		}

		// Show config file location
		if cfgFile := viper.ConfigFileUsed(); cfgFile != "" {
			utils.Debug("Config file: %s", cfgFile)
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
