/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "NSM",
	Short: "NSM (Nix Shell Manager) - A tool to manage Nix development environments",
	Long: `NSM (Nix Shell Manager) is a powerful CLI tool that helps you manage Nix development environments.

Features:
- Initialize new Nix shell environments
- Add/remove packages to your environment
- List installed packages
- Convert between shell.nix and flake.nix
- Run Nix shells
- Manage Nix channel versions
- Clean up unused packages

Example Usage:
  nsm init              # Initialize a new shell.nix
  nsm add gcc python3   # Add packages
  nsm list              # List installed packages
  nsm run              # Enter the Nix shell
  nsm clean            # Clean up unused packages`,
}

var (
	cfgFile   string
	debugMode bool
	quietMode bool
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Initialize logger before executing
	utils.ConfigureLogger(debugMode, quietMode)

	if err := rootCmd.Execute(); err != nil {
		utils.Error("Error executing command: %v", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(setupConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/NSM/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVar(&quietMode, "quiet", false, "suppress non-error output")

	// Remove default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// setupConfig reads in config file and ENV variables if set
func setupConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Use utils to ensure config directory exists
		configDir, err := utils.EnsureConfigDir()
		if err != nil {
			utils.Error("Error creating config directory: %v", err)
			os.Exit(1)
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Read environment variables
	viper.SetEnvPrefix("NSM")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("channel.url", "nixos-unstable")
	viper.SetDefault("shell.format", "shell.nix")

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			utils.Error("Error reading config file: %v", err)
		} else {
			utils.Debug("No config file found, using defaults")
		}
	} else {
		utils.Debug("Using config file: %s", viper.ConfigFileUsed())
	}
	
	// Migrate configuration if needed
	if err := utils.MigrateConfig(); err != nil {
		utils.Error("Error migrating config: %v", err)
	}
}
