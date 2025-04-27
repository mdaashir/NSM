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
	Short: "Manage configuration settings for mytool",
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
