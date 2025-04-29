/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"os/exec"
	"path/filepath"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Check represents a diagnostic check
type Check struct {
	Name        string
	Description string
	Run         func() (bool, string)
	Fix         string
}

// runDiagnostics runs all diagnostic checks
func runDiagnostics() []Check {
	return []Check{
		{
			Name:        "Nix Installation",
			Description: "Check if Nix is installed",
			Run: func() (bool, string) {
				err := utils.CheckNixInstallation()
				return err == nil, "nix"
			},
			Fix: "Install Nix from https://nixos.org/download.html",
		},
		{
			Name:        "Nix Version",
			Description: "Check Nix version",
			Run: func() (bool, string) {
				version, err := utils.GetNixVersion()
				if err != nil {
					return false, ""
				}
				return true, version
			},
			Fix: "Try reinstalling Nix",
		},
		{
			Name:        "Nixpkgs Channel",
			Description: "Check if nixpkgs channel is configured",
			Run: func() (bool, string) {
				channel, err := utils.GetChannelInfo()
				if err != nil || channel == "" {
					return false, ""
				}
				return true, channel
			},
			Fix: "Run 'nix-channel --add https://nixos.org/channels/nixpkgs-unstable nixpkgs'",
		},
		{
			Name:        "Nix Store",
			Description: "Verify Nix store",
			Run: func() (bool, string) {
				cmd := exec.Command("nix-store", "--verify")
				err := cmd.Run()
				return err == nil, ""
			},
			Fix: "Run 'nix-store --verify --repair'",
		},
		{
			Name:        "Flakes Support",
			Description: "Check if flakes are enabled",
			Run: func() (bool, string) {
				return utils.CheckFlakeSupport(), ""
			},
			Fix: "Add 'experimental-features = nix-command flakes' to your Nix config",
		},
		{
			Name:        "Project Configuration",
			Description: "Check for shell.nix or flake.nix",
			Run: func() (bool, string) {
				configType := utils.GetProjectConfigType()
				return configType != "", configType
			},
			Fix: "Run 'nsm init' to create a new environment",
		},
		{
			Name:        "NSM Configuration",
			Description: "Check NSM configuration",
			Run: func() (bool, string) {
				configFile := viper.ConfigFileUsed()
				return configFile != "", filepath.Base(configFile)
			},
			Fix: "Run any NSM command to create default configuration",
		},
	}
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose the nix environment installation",
	Long: `Perform a comprehensive health check of your Nix installation.

This command checks:
- Nix binary presence and version
- Nixpkgs channel availability and status
- Environment configuration
- Common installation issues
- Shell configuration files
- Nix store permissions
- Flakes support status
- NSM configuration

Examples:
  nsm doctor    # Run all diagnostics`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Info("🔍 Running diagnostics...")
		utils.Info("=====================")

		checks := runDiagnostics()
		issues := 0

		for _, check := range checks {
			utils.Info("\n🔍 %s:", check.Name)
			utils.Debug("  Description: %s", check.Description)

			ok, details := check.Run()
			if ok {
				if details != "" {
					utils.Success("  ✓ %s: %s", check.Description, details)
				} else {
					utils.Success("  ✓ %s", check.Description)
				}
			} else {
				issues++
				utils.Error("  ✗ %s failed", check.Description)
				utils.Tip("  Fix: %s", check.Fix)
			}
		}

		utils.Info("\n📊 Diagnostic Summary:")
		utils.Info("=====================")
		if issues == 0 {
			utils.Success("All checks passed! Your Nix installation is healthy.")
		} else {
			utils.Warn("Found %d issue(s) that need attention.", issues)
			utils.Tip("Fix the issues above to ensure proper operation.")
		}
	},
}

func init() {
	RootCmd.AddCommand(doctorCmd)
}
