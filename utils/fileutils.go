// Package utils provides utility functions for file operations, logging, configuration,
// and Nix-related functionality for the NSM (Nix Shell Manager) application.
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// isSafePath checks if a file path is safe to access
func isSafePath(path string) bool {
	// Convert to an absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Check if a path contains suspicious patterns
	suspicious := []string{
		"..", // Parent directory traversal
		"~",  // Home directory
		"$",  // Environment variables
		"|",  // Pipe
		">",  // Redirection
		"<",  // Redirection
		";",  // Command chaining
		"&",  // Background execution
		"*",  // Wildcards
		"?",  // Single character wildcard
		"[",  // Character classes
		"]",  // Character classes
	}

	for _, pattern := range suspicious {
		if strings.Contains(absPath, pattern) {
			return false
		}
	}

	return true
}

// FileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
	if !isSafePath(filename) {
		return false
	}
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// BackupFile creates a backup copy of the given file
func BackupFile(filename string) error {
	if !isSafePath(filename) {
		return fmt.Errorf("unsafe file path: %s", filename)
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return os.WriteFile(filename+".backup", content, 0600)
}

// EnsureConfigDir ensures the NSM config directory exists and returns its path
func EnsureConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Check XDG_CONFIG_HOME first
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}

	// Create NSM config directory
	nsmConfigDir := filepath.Join(configDir, "NSM")
	err = os.MkdirAll(nsmConfigDir, 0755)
	if err != nil {
		return "", err
	}

	return nsmConfigDir, nil
}

// ReadFile reads the contents of a file as a string
func ReadFile(filename string) (string, error) {
	if !isSafePath(filename) {
		return "", fmt.Errorf("unsafe file path")
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetProjectConfigType determines which type of Nix configuration file exists
// Returns "shell.nix", "flake.nix", or "" if none found
func GetProjectConfigType() string {
	if FileExists("shell.nix") {
		return "shell.nix"
	}
	if FileExists("flake.nix") {
		return "flake.nix"
	}
	return ""
}

// PinPackage pins a package to a specific version
func PinPackage() error {
	// Get current configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Initialize pins map if needed
	if config.Pins == nil {
		config.Pins = make(map[string]string)
	}

	// Get the package version
	version, err := GetPackageVersion("gcc")
	if err != nil {
		return fmt.Errorf("failed to get package version: %v", err)
	}

	// Update the pin
	config.Pins["gcc"] = version

	// Save the configuration
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}
