// Package utils provides utility functions for NSM configuration management.
package utils

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

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

	// Channel validation
	channelURL := viper.GetString("channel.url")
	if channelURL == "" {
		errors = append(errors, ConfigValidationError{
			Key:     "channel.url",
			Message: "channel URL is required",
		})
	} else if !strings.HasPrefix(channelURL, "nixos-") && !strings.HasPrefix(channelURL, "nixpkgs-") {
		errors = append(errors, ConfigValidationError{
			Key:     "channel.url",
			Message: "channel URL must start with 'nixos-' or 'nixpkgs-'",
		})
	}

	// Shell format validation
	shellFormat := viper.GetString("shell.format")
	if shellFormat != "shell.nix" && shellFormat != "flake.nix" {
		errors = append(errors, ConfigValidationError{
			Key:     "shell.format",
			Message: "shell format must be either 'shell.nix' or 'flake.nix'",
		})
	}

	// Default packages validation
	if !viper.IsSet("default.packages") {
		errors = append(errors, ConfigValidationError{
			Key:     "default.packages",
			Message: "default.packages setting is required (can be empty list)",
		})
	} else {
		defaultPkgs := viper.GetStringSlice("default.packages")
		for _, pkg := range defaultPkgs {
			if !ValidatePackage(pkg) {
				errors = append(errors, ConfigValidationError{
					Key:     "default.packages",
					Message: fmt.Sprintf("invalid package name: %s", pkg),
				})
			}
		}
	}

	// Version validation
	version := viper.GetString("config_version")
	if version == "" {
		errors = append(errors, ConfigValidationError{
			Key:     "config_version",
			Message: "config version is required",
		})
	} else if !isValidVersion(version) {
		errors = append(errors, ConfigValidationError{
			Key:     "config_version",
			Message: fmt.Sprintf("invalid config version: %s (must be semver)", version),
		})
	}

	// Pins validation
	if pins := viper.GetStringMapString("pins"); pins != nil {
		for pkg, version := range pins {
			if !ValidatePackage(pkg) {
				errors = append(errors, ConfigValidationError{
					Key:     "pins",
					Message: fmt.Sprintf("invalid package name in pins: %s", pkg),
				})
			}
			if !isValidVersion(version) {
				errors = append(errors, ConfigValidationError{
					Key:     "pins",
					Message: fmt.Sprintf("invalid version for package %s: %s", pkg, version),
				})
			}
		}
	}

	return errors
}

// isValidVersion checks if a version string follows semantic versioning
func isValidVersion(version string) bool {
	// Basic semver pattern
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	return true
}

// Config represents the NSM configuration structure
type Config struct {
	Pins            map[string]string `mapstructure:"pins"`
	DefaultPackages []string          `mapstructure:"default.packages"`
	ChannelURL      string            `mapstructure:"channel.url"`
	ShellFormat     string            `mapstructure:"shell.format"`
	Version         string            `mapstructure:"config_version"`
}

// LoadConfig loads and returns the NSM configuration
func LoadConfig() (*Config, error) {
	var config Config

	// Set defaults before loading
	viper.SetDefault("channel.url", "nixos-unstable")
	viper.SetDefault("shell.format", "shell.nix")
	viper.SetDefault("default.packages", []string{})
	viper.SetDefault("config_version", "1.0.0")

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %v", err)
	}

	// Initialize maps if nil
	if config.Pins == nil {
		config.Pins = make(map[string]string)
	}

	return &config, nil
}

// SaveConfig saves the NSM configuration
func SaveConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("cannot save nil config")
	}

	// Validate before saving
	viper.Set("pins", config.Pins)
	viper.Set("default.packages", config.DefaultPackages)
	viper.Set("channel.url", config.ChannelURL)
	viper.Set("shell.format", config.ShellFormat)
	viper.Set("config_version", config.Version)

	if errors := ValidateConfig(); len(errors) > 0 {
		return fmt.Errorf("invalid configuration: %v", errors)
	}

	// Create backup of existing config
	configFile := viper.ConfigFileUsed()
	if configFile != "" && FileExists(configFile) {
		if err := BackupFile(configFile); err != nil {
			Debug("Failed to create config backup: %v", err)
		}
	}

	// Try WriteConfig first
	err := viper.WriteConfig()
	if err != nil {
		// If WriteConfig fails, try SafeWriteConfig
		if err := viper.SafeWriteConfig(); err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}
	}

	return nil
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
	var needsSave bool

	// Initialize missing settings
	if !viper.IsSet("config_version") {
		// Set the initial version
		viper.Set("config_version", "1.0.0")
		needsSave = true
	}

		// Ensure default.packages exists as empty slice if not set
	if !viper.IsSet("default.packages") {
		viper.Set("default.packages", []string{})
		needsSave = true
	}

		// Ensure shell.format is set
	if !viper.IsSet("shell.format") {
		viper.Set("shell.format", "shell.nix")
		needsSave = true
	}

	if !viper.IsSet("pins") {
		viper.Set("pins", make(map[string]string))
		needsSave = true
	}

	// Migrate old channel format
	if viper.IsSet("channel") && !viper.IsSet("channel.url") {
		oldChannel := viper.GetString("channel")
		if oldChannel != "" {
			viper.Set("channel.url", oldChannel)
			viper.Set("channel", nil)
			needsSave = true
			Debug("Migrated channel format from %q to channel.url", oldChannel)
		}
	}

	// Migrate from version 1.0.0 to 1.1.0
	if viper.GetString("config_version") == "1.0.0" {
		viper.Set("config_version", "1.1.0")
		needsSave = true
	}

	if needsSave {
		configFile := viper.ConfigFileUsed()
		if configFile != "" && FileExists(configFile) {
			if err := BackupFile(configFile); err != nil {
				Debug("Failed to create config backup during migration: %v", err)
			}
		}

		if err := viper.WriteConfig(); err != nil {
			if err := viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to save migrated config: %v", err)
			}
		}
		Debug("Successfully migrated configuration")
	}

	return nil
}
