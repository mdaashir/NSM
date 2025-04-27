package unit

import (
	"os"
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/viper"
)

func setupTestConfig(t *testing.T) func() {
	t.Helper()
	dir := testutils.CreateTempDir(t)

	// Save original config state
	oldConfigFile := viper.ConfigFileUsed()

	// Reset viper completely
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	// Set initial config values
	viper.Set("channel.url", "nixos-unstable")
	viper.Set("shell.format", "shell.nix")
	viper.Set("default.packages", []string{"gcc", "python3"})
	viper.Set("config_version", "1.0.0")
	viper.Set("pins", map[string]string{
		"gcc":     "12.3.0",
		"python3": "3.9.0",
	})

	// Save initial config
	err := viper.SafeWriteConfig()
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cleanup := func() {
		// Reset viper to original state
		viper.Reset()
		if oldConfigFile != "" {
			viper.SetConfigFile(oldConfigFile)
			_ = viper.ReadInConfig()
		}

		// Clean up test directory
		err := os.RemoveAll(dir)
		if err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}

	return cleanup
}

func TestConfigValidation(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	t.Run("valid config", func(t *testing.T) {
		errors := utils.ValidateConfig()
		if len(errors) > 0 {
			t.Errorf("Expected no validation errors, got %v", errors)
		}
	})

	t.Run("missing channel url", func(t *testing.T) {
		viper.Set("channel.url", "")
		errors := utils.ValidateConfig()
		if len(errors) == 0 {
			t.Error("Expected validation error for missing channel URL")
		}
		// Restore valid value
		viper.Set("channel.url", "nixos-unstable")
	})

	t.Run("invalid shell format", func(t *testing.T) {
		viper.Set("shell.format", "invalid")
		errors := utils.ValidateConfig()
		if len(errors) == 0 {
			t.Error("Expected validation error for invalid shell format")
		}
		// Restore valid value
		viper.Set("shell.format", "shell.nix")
	})
}

func TestMigrateConfig(t *testing.T) {
	t.Run("migrate from no version", func(t *testing.T) {
		// Remove version
		viper.Set("config_version", nil)

		if err := utils.MigrateConfig(); err != nil {
			t.Fatalf("MigrateConfig() error = %v", err)
		}

		if !viper.IsSet("config_version") {
			t.Error("config_version was not set during migration")
		}

		if ver := viper.GetString("config_version"); ver != "1.0.0" {
			t.Errorf("config_version = %q, want 1.0.0", ver)
		}
	})
}

func TestLoadConfig(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	config, err := utils.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Check pins
	expectedPins := map[string]string{
		"gcc":     "12.3.0",
		"python3": "3.9.0",
	}

	if len(config.Pins) != len(expectedPins) {
		t.Errorf("Expected %d pins, got %d", len(expectedPins), len(config.Pins))
	}

	for pkg, version := range expectedPins {
		if got := config.Pins[pkg]; got != version {
			t.Errorf("Pin[%q] = %q, want %q", pkg, got, version)
		}
	}
}

func TestSaveConfig(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Modify config
	config := &utils.Config{
		Pins: map[string]string{
			"nodejs": "18.0.0",
			"go":     "1.24.0",
		},
	}

	if err := utils.SaveConfig(config); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify changes were saved
	newConfig, err := utils.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(newConfig.Pins) != len(config.Pins) {
		t.Errorf("Saved config has %d pins, want %d", len(newConfig.Pins), len(config.Pins))
	}

	for pkg, version := range config.Pins {
		if got := newConfig.Pins[pkg]; got != version {
			t.Errorf("Saved pin[%q] = %q, want %q", pkg, got, version)
		}
	}
}
