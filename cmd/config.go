/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage NSM configuration",
	Long: `Manage NSM configuration settings.

This command allows you to:
- View current configuration
- Set configuration values
- Reset to defaults
- Validate configuration
- Import/export settings

Examples:
  nsm config                                 # Show current config
  nsm config set channel.url nixos-22.11    # Set channel URL
  nsm config set shell.format flake.nix     # Set default shell format
  nsm config add default.packages gcc        # Add default package
  nsm config remove default.packages gcc     # Remove default package
  nsm config validate                       # Validate current config
  nsm config reset                          # Reset to defaults`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		summary := utils.GetConfigSummary()

		// Convert to JSON for pretty printing
		output, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			utils.Error("Failed to format config: %v", err)
			return
		}

		utils.Info("📝 Current Configuration:")
		fmt.Println(string(output))

		// Show validation status
		if errors := utils.ValidateConfig(); len(errors) > 0 {
			utils.Warn("\n⚠️ Configuration has validation issues:")
			for _, err := range errors {
				utils.Error("- %s", err.Error())
			}
			utils.Tip("Run 'nsm config validate' for more details")
		} else {
			utils.Success("✅ Configuration is valid")
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		// Special handling for different types
		switch key {
		case "shell.format":
			if value != "shell.nix" && value != "flake.nix" {
				utils.Error("Invalid shell format. Must be 'shell.nix' or 'flake.nix'")
				return
			}
		}

		viper.Set(key, value)
		if err := viper.WriteConfig(); err != nil {
			utils.Error("Failed to save config: %v", err)
			return
		}

		utils.Success("Set %s = %s", key, value)

		// Validate the change
		if errors := utils.ValidateConfig(); len(errors) > 0 {
			utils.Warn("New configuration has validation issues")
			utils.Tip("Run 'nsm config validate' to see details")
		}
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		utils.Info("🔍 Validating configuration...")

		errors := utils.ValidateConfig()
		if len(errors) == 0 {
			utils.Success("Configuration is valid!")
			return
		}

		utils.Error("\nFound %d validation issue(s):", len(errors))
		for _, err := range errors {
			utils.Error("- %s", err.Error())
		}
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add [key] [value]",
	Short: "Add a value to a list setting",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		// Only support adding to lists
		if !strings.HasPrefix(key, "default.packages") {
			utils.Error("Can only add to list settings (e.g., default.packages)")
			return
		}

		// Get current list
		current := viper.GetStringSlice(key)

		// Check if value already exists
		for _, v := range current {
			if v == value {
				utils.Warn("Value %s already exists in %s", value, key)
				return
			}
		}

		// Add new value
		current = append(current, value)
		viper.Set(key, current)

		if err := viper.WriteConfig(); err != nil {
			utils.Error("Failed to save config: %v", err)
			return
		}

		utils.Success("Added %s to %s", value, key)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove [key] [value]",
	Short: "Remove a value from a list setting",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		// Only support removing from lists
		if !strings.HasPrefix(key, "default.packages") {
			utils.Error("Can only remove from list settings (e.g., default.packages)")
			return
		}

		// Get current list
		current := viper.GetStringSlice(key)

		// Remove value
		var newList []string
		found := false
		for _, v := range current {
			if v == value {
				found = true
				continue
			}
			newList = append(newList, v)
		}

		if !found {
			utils.Warn("Value %s not found in %s", value, key)
			return
		}

		viper.Set(key, newList)

		if err := viper.WriteConfig(); err != nil {
			utils.Error("Failed to save config: %v", err)
			return
		}

		utils.Success("Removed %s from %s", value, key)
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Run: func(cmd *cobra.Command, args []string) {
		// Set default values
		viper.Set("channel.url", "nixos-unstable")
		viper.Set("shell.format", "shell.nix")
		viper.Set("default.packages", []string{})
		viper.Set("config_version", "1.0.0")

		if err := viper.WriteConfig(); err != nil {
			utils.Error("Failed to save config: %v", err)
			return
		}

		utils.Success("Reset configuration to defaults")
		utils.Tip("Run 'nsm config show' to see the new configuration")
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configResetCmd)
	rootCmd.AddCommand(configCmd)
}
