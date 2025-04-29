package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Status constants for diagnostic results
const (
	StatusUnknown = "UNKNOWN"
	StatusOK      = "OK"
	StatusWarning = "WARNING"
	StatusError   = "ERROR"
)

// DoctorResult represents a diagnostic check result with detailed information
type DoctorResult struct {
	Name        string // Name of the check
	Description string // Description of what is being checked
	Status      string // Status: OK, WARNING, ERROR, UNKNOWN
	Message     string // Detailed message
	Fix         string // Suggested fix if applicable
}

// SystemCheck represents a system diagnostic check
type SystemCheck struct {
	Name  string
	Check func() (bool, string)
}

// CheckResult represents the basic result of a system check (legacy format)
type CheckResult struct {
	Name    string
	Success bool
	Message string
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

// RunSystemChecks performs all system diagnostics (legacy method)
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

// RunDiagnostics runs all diagnostic checks and returns detailed results
func RunDiagnostics() []DoctorResult {
	started := time.Now()
	Debug("Starting diagnostic tests")

	var results []DoctorResult

	// Add OS-specific checks
	if runtime.GOOS == "windows" {
		results = append(results, *CheckWindowsSpecific())
		results = append(results, *CheckDiskSpace())
	} else {
		// Unix-specific checks will be handled by doctorutils_unix.go
		results = append(results, CheckUnixPermissions())
		results = append(results, CheckNixDaemon())
	}

	// Common checks for all platforms
	results = append(results, CheckNixInstalled())
	results = append(results, CheckNixChannels())
	results = append(results, CheckFlakes())
	results = append(results, CheckConfiguration())
	results = append(results, CheckProjectFiles())

	Debug("Completed diagnostic tests in %v", time.Since(started))
	return results
}

// CheckNixInstalled checks if Nix is properly installed
func CheckNixInstalled() DoctorResult {
	result := DoctorResult{
		Name:        "Nix Installation",
		Description: "Checking if Nix is properly installed",
		Status:      StatusUnknown,
	}

	if !IsNixInstalled() {
		result.Status = StatusError
		result.Message = "Nix is not installed"

		if runtime.GOOS == "darwin" {
			result.Fix = "Install Nix using: sh <(curl -L https://nixos.org/nix/install)"
		} else if runtime.GOOS == "linux" {
			result.Fix = "Install Nix using: sh <(curl -L https://nixos.org/nix/install) --daemon"
		} else if runtime.GOOS == "windows" {
			result.Fix = "Install Nix using WSL first, then run: sh <(curl -L https://nixos.org/nix/install)"
		}

		return result
	}

	// Check if required Nix commands are available
	err := CheckNixInstallation()
	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Nix installation issue: %v", err)
		result.Fix = "Reinstall Nix or check your PATH configuration"
		return result
	}

	// Get Nix version
	version, err := GetNixVersion()
	if err != nil {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Could not determine Nix version: %v", err)
		return result
	}

	result.Status = StatusOK
	result.Message = fmt.Sprintf("Nix is properly installed: %s", version)
	return result
}

// CheckNixChannels checks Nix channel configuration
func CheckNixChannels() DoctorResult {
	result := DoctorResult{
		Name:        "Nix Channels",
		Description: "Checking Nix channel configuration",
		Status:      StatusUnknown,
	}

	channel, err := GetChannelInfo()
	if err != nil {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Channel issue: %v", err)
		result.Fix = "Set up channels with: nix-channel --add https://nixos.org/channels/nixos-unstable nixos"
		return result
	}

	result.Status = StatusOK
	result.Message = fmt.Sprintf("Channel configured: %s", channel)
	return result
}

// CheckFlakes checks for Nix flakes support
func CheckFlakes() DoctorResult {
	result := DoctorResult{
		Name:        "Flakes Support",
		Description: "Checking if Nix flakes are enabled",
		Status:      StatusUnknown,
	}

	if !CheckFlakeSupport() {
		result.Status = StatusWarning
		result.Message = "Flakes are not enabled"
		result.Fix = "Add 'experimental-features = nix-command flakes' to your Nix configuration"
		return result
	}

	result.Status = StatusOK
	result.Message = "Flakes are supported"
	return result
}

// CheckConfiguration validates NSM configuration
func CheckConfiguration() DoctorResult {
	result := DoctorResult{
		Name:        "NSM Configuration",
		Description: "Checking NSM configuration",
		Status:      StatusUnknown,
	}

	errors := ValidateConfig()
	if len(errors) > 0 {
		result.Status = StatusError
		var msgs []string
		for _, err := range errors {
			msgs = append(msgs, fmt.Sprintf("- %s", err.Error()))
		}
		result.Message = fmt.Sprintf("Configuration errors:\n%s", strings.Join(msgs, "\n"))
		result.Fix = "Run 'nsm config reset' to reset to defaults"
		return result
	}

	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		result.Status = StatusWarning
		result.Message = "No configuration file found, using defaults"
		return result
	}

	result.Status = StatusOK
	result.Message = fmt.Sprintf("Configuration is valid: %s", configFile)
	return result
}

// CheckProjectFiles checks for Nix project files in current directory
func CheckProjectFiles() DoctorResult {
	result := DoctorResult{
		Name:        "Project Files",
		Description: "Checking for Nix project files",
		Status:      StatusUnknown,
	}

	configType := GetProjectConfigType()
	if configType == "" {
		result.Status = StatusWarning
		result.Message = "No Nix configuration found in current directory"
		result.Fix = "Run 'nsm init' to create a new Nix environment"
		return result
	}

	if !FileExists(configType) {
		result.Status = StatusError
		result.Message = fmt.Sprintf("%s was detected but cannot be read", configType)
		return result
	}

	// Try to parse packages from the file
	packages, err := ParsePackageList(configType)
	if err != nil {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Found %s but couldn't parse packages: %v", configType, err)
		return result
	}

	result.Status = StatusOK
	if len(packages) == 0 {
		result.Message = fmt.Sprintf("Found valid %s with no packages", configType)
	} else {
		result.Message = fmt.Sprintf("Found valid %s with %d packages", configType, len(packages))
	}

	return result
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
	status["diagnostics"] = RunDiagnostics()

	// Keep legacy checks for backward compatibility
	status["checks"] = RunSystemChecks()

	// Get resource usage
	status["resources"] = getResourceUsage()

	return status, nil
}

// getResourceUsage gets system resource usage information
func getResourceUsage() map[string]interface{} {
	usage := make(map[string]interface{})

	// Get Nix store size if available
	storeDir := "/nix/store"
	if runtime.GOOS == "windows" {
		// On Windows, Nix runs inside WSL
		storeDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "Packages", "NixOS.*", "LocalState", "rootfs", "nix", "store")
	}

	if DirExists(storeDir) {
		if size, err := getDirSize(storeDir); err == nil {
			usage["store_size_bytes"] = size
			usage["store_size_gb"] = float64(size) / (1024 * 1024 * 1024)
		}
	}

	// Get free disk space for the Nix store directory
	if space, err := getDiskSpace(storeDir); err == nil {
		usage["free_space_bytes"] = space
		usage["free_space_gb"] = float64(space) / (1024 * 1024 * 1024)
	}

	return usage
}

// getDirSize gets the size of a directory in bytes
func getDirSize(path string) (int64, error) {
	if !DirExists(path) {
		return 0, fmt.Errorf("directory does not exist: %s", path)
	}

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
