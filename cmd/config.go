/*
Copyright © 2025 Mohamed Aashir S <s.mohamedaashir@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage NSM configuration settings",
	Long: `Manage configuration settings for NSM (Nix Shell Manager).

Available Commands:
  set <key> <value>    Set a configuration value
  get <key>            Get a configuration value

Configuration Options:
  default.packages     Default packages for new environments
  channel.url         Default Nixpkgs channel URL
  shell.format        Preferred format (shell.nix/flake.nix)

Examples:
  nsm config set default.packages "gcc python3"
  nsm config set channel.url "nixos-unstable"
  nsm config get default.packages

Settings are stored in $HOME/.config/NSM/config.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if a specific action is requested (e.g., setting default packages)
		if len(args) == 0 {
			fmt.Println("❌ No arguments provided for config command")
			return
		}

		switch args[0] {
		case "set":
			if len(args) < 3 {
				fmt.Println("❌ Set requires key and value: config set <key> <value>")
				return
			}
			key := args[1]
			value := args[2]

			viper.Set(key, value)

			err := viper.WriteConfig()
			if err != nil {
				fmt.Println("❌ Error writing config:", err)
				return
			}

			fmt.Printf("✅ Config key '%s' set to '%s'\n", key, value)

		case "get":
			if len(args) < 2 {
				fmt.Println("❌ Get requires a key: config get <key>")
				return
			}
			key := args[1]

			value := viper.GetString(key)
			if value == "" {
				fmt.Printf("❌ Config key '%s' not found.\n", key)
				return
			}

			fmt.Printf("✅ Config '%s' is set to '%s'\n", key, value)

		default:
			fmt.Println("❌ Unknown config command.")
		}
	},
}

func init() {
	// Set default config path
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.config/NSM")

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("❌ No config file found, starting fresh.")
	}

	rootCmd.AddCommand(configCmd)
}
