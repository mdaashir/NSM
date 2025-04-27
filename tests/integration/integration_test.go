package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/viper"
)

// TestPackageManagement tests the full package management lifecycle
func TestPackageManagement(t *testing.T) {
	config, cleanup := testutils.CreateTestConfig(t)
	defer cleanup()

	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Change to test directory
	if err := os.Chdir(config.TempDir); err != nil {
		t.Fatal(err)
	}

	t.Run("package operations", func(t *testing.T) {
		// Verify initial packages
		content, err := os.ReadFile("shell.nix")
		if err != nil {
			t.Fatal(err)
		}

		initialPkgs := utils.ExtractShellNixPackages(string(content))
		if len(initialPkgs) != 2 {
			t.Errorf("Expected 2 initial packages, got %d", len(initialPkgs))
		}

		// Pin a package version
		if err := utils.PinPackage("gcc", "12.3.0"); err != nil {
			t.Errorf("Failed to pin package: %v", err)
		}

		// Verify pin was saved
		config, err := utils.LoadConfig()
		if err != nil {
			t.Fatal(err)
		}

		if version := config.Pins["gcc"]; version != "12.3.0" {
			t.Errorf("Pin version = %q, want 12.3.0", version)
		}
	})
}

// TestConfigurationFlow tests the configuration system end-to-end
func TestConfigurationFlow(t *testing.T) {
	dir := testutils.CreateTempDir(t)
	defer os.RemoveAll(dir)

	// Set up test config file
	configPath := filepath.Join(dir, "config.yaml")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	viper.Reset()
	viper.SetConfigFile(configPath)

	t.Run("config lifecycle", func(t *testing.T) {
		// Initialize with defaults
		viper.Set("channel.url", "nixos-unstable")
		viper.Set("shell.format", "shell.nix")
		viper.Set("default.packages", []string{})

		if err := viper.WriteConfig(); err != nil {
			t.Fatal(err)
		}

		// Validate initial config
		if errs := utils.ValidateConfig(); len(errs) > 0 {
			t.Errorf("Initial config validation failed: %v", errs)
		}

		// Modify configuration
		viper.Set("shell.format", "flake.nix")
		if err := viper.WriteConfig(); err != nil {
			t.Fatal(err)
		}

		// Verify changes
		if format := viper.GetString("shell.format"); format != "flake.nix" {
			t.Errorf("shell.format = %q, want flake.nix", format)
		}

		// Test migration
		viper.Set("config_version", nil)
		if err := utils.MigrateConfig(); err != nil {
			t.Fatal(err)
		}

		if !viper.IsSet("config_version") {
			t.Error("Migration did not set config_version")
		}
	})
}

// TestNixEnvironment tests Nix environment detection and management
func TestNixEnvironment(t *testing.T) {
	config, cleanup := testutils.CreateTestConfig(t)
	defer cleanup()

	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Change to test directory
	if err := os.Chdir(config.TempDir); err != nil {
		t.Fatal(err)
	}

	t.Run("environment detection", func(t *testing.T) {
		// Test shell.nix detection
		configType := utils.GetProjectConfigType()
		if configType != "shell.nix" {
			t.Errorf("GetProjectConfigType() = %q, want shell.nix", configType)
		}

		// Remove shell.nix and test flake.nix detection
		os.Remove("shell.nix")
		configType = utils.GetProjectConfigType()
		if configType != "flake.nix" {
			t.Errorf("GetProjectConfigType() = %q, want flake.nix", configType)
		}

		// Test with no config files
		os.Remove("flake.nix")
		configType = utils.GetProjectConfigType()
		if configType != "" {
			t.Errorf("GetProjectConfigType() = %q, want empty string", configType)
		}
	})
}

// TestLoggingSystem tests the logging system across different components
func TestLoggingSystem(t *testing.T) {
	config, cleanup := testutils.CreateTestConfig(t)
	defer cleanup()

	t.Run("logging across components", func(t *testing.T) {
		// Configure logger
		utils.ConfigureLogger(true, false)

		stdout, stderr := testutils.CaptureOutput(func() {
			// Test error logging from Nix operations
			utils.CheckNixInstallation()

			// Test info logging from file operations
			utils.FileExists("nonexistent-file")

			// Test debug logging from config operations
			utils.ValidateConfig()
		})

		// Verify log output contains expected components
		if !strings.Contains(stdout, "[DEBUG]") {
			t.Error("Debug logging not working")
		}

		// Error messages should go to stderr
		if stderr != "" && !strings.Contains(stderr, "✗") {
			t.Error("Error logging not working")
		}
	})
}
