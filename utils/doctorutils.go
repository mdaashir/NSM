package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SystemCheck represents a system diagnostic check
type SystemCheck struct {
	Name  string
	Check func() (bool, string)
}

var systemChecks = []SystemCheck{
	{
		Name: "Nix Installation",
		Check: func() (bool, string) {
			if err := CheckNixInstallation(); err != nil {
				return false, fmt.Sprintf("Nix not properly installed: %v", err)
			}
			return true, "Nix is properly installed"
		},
	},
	{
		Name: "Nix Store Permissions",
		Check: func() (bool, string) {
			if _, err := os.Stat("/nix/store"); err != nil {
				return false, fmt.Sprintf("Cannot access Nix store: %v", err)
			}
			return true, "Nix store is accessible"
		},
	},
	{
		Name: "Configuration",
		Check: func() (bool, string) {
			if errors := ValidateConfig(); len(errors) > 0 {
				var msgs []string
				for _, err := range errors {
					msgs = append(msgs, err.Error())
				}
				return false, fmt.Sprintf("Configuration errors:\n%s", strings.Join(msgs, "\n"))
			}
			return true, "Configuration is valid"
		},
	},
	{
		Name: "Nix Channel",
		Check: func() (bool, string) {
			if _, err := GetChannelInfo(); err != nil {
				return false, fmt.Sprintf("Channel issue: %v", err)
			}
			return true, "Channel is properly configured"
		},
	},
	{
		Name: "Flakes Support",
		Check: func() (bool, string) {
			if !CheckFlakeSupport() {
				return false, "Flakes are not enabled"
			}
			return true, "Flakes are supported"
		},
	},
	{
		Name: "Project Configuration",
		Check: func() (bool, string) {
			configType := GetProjectConfigType()
			if configType == "" {
				return false, "No Nix configuration found in current directory"
			}
			if _, err := os.ReadFile(configType); err != nil {
				return false, fmt.Sprintf("Cannot read %s: %v", configType, err)
			}
			return true, fmt.Sprintf("Found valid %s", configType)
		},
	},
}

// RunSystemChecks performs all system diagnostics
func RunSystemChecks() []CheckResult {
	var results []CheckResult

	for _, check := range systemChecks {
		ok, msg := check.Check()
		results = append(results, CheckResult{
			Name:    check.Name,
			Success: ok,
			Message: msg,
		})
	}

	return results
}

// CheckResult represents the result of a system check
type CheckResult struct {
	Name    string
	Success bool
	Message string
}

// FixCommonIssues attempts to fix common system issues
func FixCommonIssues() []string {
	var fixed []string

	// Try to update channel
	if err := UpdateChannel(); err == nil {
		fixed = append(fixed, "Updated Nix channel")
	}

	// Try to collect garbage
	if err := CollectGarbage(); err == nil {
		fixed = append(fixed, "Cleaned up Nix store")
	}

	// Try to fix configuration
	if config, err := LoadConfig(); err == nil {
		// Reset invalid settings to defaults
		if config.ChannelURL == "" {
			config.ChannelURL = "nixos-unstable"
			fixed = append(fixed, "Reset channel URL to default")
		}
		if config.ShellFormat == "" {
			config.ShellFormat = "shell.nix"
			fixed = append(fixed, "Reset shell format to default")
		}
		if config.Pins == nil {
			config.Pins = make(map[string]string)
			fixed = append(fixed, "Initialized package pins")
		}

		if err := SaveConfig(config); err == nil {
			fixed = append(fixed, "Saved fixed configuration")
		}
	}

	return fixed
}

// GetSystemStatus returns a comprehensive system status report
func GetSystemStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Get system info
	if info, err := GetSystemInfo(); err == nil {
		status["system"] = info
	} else {
		status["system"] = map[string]string{
			"error": err.Error(),
		}
	}

	// Get configuration status
	config, err := LoadConfig()
	if err == nil {
		status["config"] = config
	} else {
		status["config"] = map[string]string{
			"error": err.Error(),
		}
	}

	// Get diagnostic results
	status["checks"] = RunSystemChecks()

	// Get resource usage
	status["resources"] = getResourceUsage()

	return status, nil
}

// getResourceUsage gets system resource usage information
func getResourceUsage() map[string]interface{} {
	usage := make(map[string]interface{})

	// Get Nix store size
	if size, err := getDirSize("/nix/store"); err == nil {
		usage["store_size"] = size
	}

	// Get free disk space
	if space, err := getDiskSpace("/nix"); err == nil {
		usage["free_space"] = space
	}

	return usage
}

// getDirSize gets the size of a directory in bytes
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
