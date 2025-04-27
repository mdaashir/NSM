// Package utils provides utility functions for NSM configuration management.
package utils

import (
	"fmt"
	"os/exec"

	"github.com/spf13/viper"
)

// ConfigValidationError represents a configuration validation error
type ConfigValidationError struct {
	Key     string
	Message string
}

func (e ConfigValidationError) Error() string {
	return fmt.Sprintf("config validation error for %s: %s", e.Key, e.Message)
}

// ValidateConfig checks if the configuration has all required values
func ValidateConfig() []ConfigValidationError {
	var errors []ConfigValidationError

	// Check channel URL
	channelURL := viper.GetString("channel.url")
	if channelURL == "" {
		errors = append(errors, ConfigValidationError{
			Key:     "channel.url",
			Message: "channel URL is required",
		})
	}

	// Check a shell format
	shellFormat := viper.GetString("shell.format")
	if shellFormat != "shell.nix" && shellFormat != "flake.nix" {
		errors = append(errors, ConfigValidationError{
			Key:     "shell.format",
			Message: "shell format must be either 'shell.nix' or 'flake.nix'",
		})
	}

	// Check default packages format
	if !viper.IsSet("default.packages") {
		errors = append(errors, ConfigValidationError{
			Key:     "default.packages",
			Message: "default.packages setting is required (can be empty list)",
		})
	} else {
		defaultPkgs := viper.GetStringSlice("default.packages")
		if defaultPkgs == nil {
			errors = append(errors, ConfigValidationError{
				Key:     "default.packages",
				Message: "default packages must be a list (can be empty)",
			})
		}
	}

	return errors
}

// Config represents the NSM configuration structure
type Config struct {
	Pins map[string]string
}

// LoadConfig loads and returns the NSM configuration
func LoadConfig() (*Config, error) {
	config := &Config{
		Pins: make(map[string]string),
	}

	// Get pins from viper
	if pins := viper.GetStringMapString("pins"); pins != nil {
		config.Pins = pins
	}

	return config, nil
}

// SaveConfig saves the NSM configuration
func SaveConfig(config *Config) error {
	// Set pins in a viper
	if config.Pins != nil {
		viper.Set("pins", config.Pins)
	}

	return viper.WriteConfig()
}

// GetConfigSummary returns a human-readable summary of the current configuration
func GetConfigSummary() map[string]interface{} {
	// Import circular reference resolved by moving CheckNixInstallation call logic here
	_, nixErr := exec.LookPath("nix-env")

	return map[string]interface{}{
		"channel.url":      viper.GetString("channel.url"),
		"shell.format":     viper.GetString("shell.format"),
		"default.packages": viper.GetStringSlice("default.packages"),
		"config_file":      viper.ConfigFileUsed(),
		"environment":      viper.GetString("environment"),
		"flakes_enabled":   CheckFlakeSupport(),
		"nix_installed":    nixErr == nil,
		"config_validated": len(ValidateConfig()) == 0,
	}
}

// MigrateConfig handles configuration format changes
func MigrateConfig() error {
	// Always check and migrate channel URL first, regardless of config version
	if viper.IsSet("channel") && !viper.IsSet("channel.url") {
		channelValue := viper.GetString("channel")
		if channelValue != "" {
			// Set the new format and clear the old one
			viper.Set("channel.url", channelValue)
			viper.Set("channel", nil)
		}
	}

	// Check if we need to migrate other settings
	if !viper.IsSet("config_version") {
		// Set the initial version
		viper.Set("config_version", "1.0.0")

		// Ensure default.packages exists as empty slice if not set
		if !viper.IsSet("default.packages") {
			viper.Set("default.packages", []string{})
		}

		// Ensure shell.format is set
		if !viper.IsSet("shell.format") {
			viper.Set("shell.format", "shell.nix")
		}

		// Save changes
		if err := viper.SafeWriteConfig(); err != nil {
			// If SafeWriteConfig fails, try WriteConfig
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("failed to save migrated config: %v", err)
			}
		}
	}

	return nil
}
