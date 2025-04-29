package unit

import (
	"testing"

	"github.com/mdaashir/NSM/tests/testutils"
	"github.com/mdaashir/NSM/utils"
	"github.com/spf13/viper"
)

func TestConfigValidation(t *testing.T) {
	_, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedErrs  int
		expectErrKeys []string
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"channel.url":      "nixos-unstable",
				"shell.format":     "shell.nix",
				"default.packages": []string{},
				"config_version":   "1.0.0",
				"pins":             map[string]string{},
			},
			expectedErrs: 0,
		},
		{
			name: "invalid channel",
			config: map[string]interface{}{
				"channel.url":      "invalid-channel",
				"shell.format":     "shell.nix",
				"default.packages": []string{},
				"config_version":   "1.0.0",
			},
			expectedErrs:  1,
			expectErrKeys: []string{"channel.url"},
		},
		{
			name: "invalid shell format",
			config: map[string]interface{}{
				"channel.url":      "nixos-unstable",
				"shell.format":     "invalid.nix",
				"default.packages": []string{},
				"config_version":   "1.0.0",
			},
			expectedErrs:  1,
			expectErrKeys: []string{"shell.format"},
		},
		{
			name: "invalid version",
			config: map[string]interface{}{
				"channel.url":      "nixos-unstable",
				"shell.format":     "shell.nix",
				"default.packages": []string{},
				"config_version":   "invalid",
			},
			expectedErrs:  1,
			expectErrKeys: []string{"config_version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and set up config for each test
			viper.Reset()
			for k, v := range tt.config {
				viper.Set(k, v)
			}

			errors := utils.ValidateConfig()
			if len(errors) != tt.expectedErrs {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrs, len(errors))
			}

			if tt.expectErrKeys != nil {
				errKeys := make(map[string]bool)
				for _, err := range errors {
					errKeys[err.Key] = true
				}
				for _, key := range tt.expectErrKeys {
					if !errKeys[key] {
						t.Errorf("Expected error for key %s not found", key)
					}
				}
			}
		})
	}
}

func TestConfigMigration(t *testing.T) {
	_, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name           string
		initialConfig  map[string]interface{}
		expectedConfig map[string]interface{}
	}{
		{
			name: "migrate old channel format",
			initialConfig: map[string]interface{}{
				"channel": "nixos-unstable",
			},
			expectedConfig: map[string]interface{}{
				"channel.url": "nixos-unstable",
			},
		},
		{
			name: "migrate missing version",
			initialConfig: map[string]interface{}{
				"channel.url": "nixos-unstable",
			},
			expectedConfig: map[string]interface{}{
				"channel.url":    "nixos-unstable",
				"config_version": "1.0.0",
			},
		},
		{
			name: "migrate old version",
			initialConfig: map[string]interface{}{
				"channel.url":    "nixos-unstable",
				"config_version": "1.0.0",
			},
			expectedConfig: map[string]interface{}{
				"channel.url":    "nixos-unstable",
				"config_version": "1.1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and set up config for each test
			viper.Reset()
			for k, v := range tt.initialConfig {
				viper.Set(k, v)
			}

			err := utils.MigrateConfig()
			testutils.AssertNoError(t, err)

			for k, v := range tt.expectedConfig {
				testutils.AssertConfigValue(t, k, v)
			}
		})
	}
}

func TestConfigIO(t *testing.T) {
	_, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Test loading config
	config, err := utils.LoadConfig()
	testutils.AssertNoError(t, err)
	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	// Test saving config
	config.ChannelURL = "nixos-22.05"
	err = utils.SaveConfig(config)
	testutils.AssertNoError(t, err)

	// Verify saved config
	newConfig, err := utils.LoadConfig()
	testutils.AssertNoError(t, err)
	if newConfig.ChannelURL != "nixos-22.05" {
		t.Errorf("Expected channel.url to be nixos-22.05, got %s", newConfig.ChannelURL)
	}

	// Test nil config
	err = utils.SaveConfig(nil)
	testutils.AssertError(t, err)
}

func TestConfigSummary(t *testing.T) {
	_, cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	summary := utils.GetConfigSummary()

	expectedKeys := []string{
		"channel.url",
		"shell.format",
		"default.packages",
		"config_file",
		"environment",
		"flakes_enabled",
		"nix_installed",
		"config_validated",
	}

	for _, key := range expectedKeys {
		if _, ok := summary[key]; !ok {
			t.Errorf("Expected summary to contain key %s", key)
		}
	}
}
