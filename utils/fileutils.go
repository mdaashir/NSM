package utils

import (
	"os"
	"path/filepath"
)

// FileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// BackupFile creates a backup copy of the given file
func BackupFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return os.WriteFile(filename+".backup", content, 0644)
}

// GetProjectConfigType returns "shell.nix", "flake.nix", or "" based on what exists
func GetProjectConfigType() string {
	if FileExists("shell.nix") {
		return "shell.nix"
	}
	if FileExists("flake.nix") {
		return "flake.nix"
	}
	return ""
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
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile writes content to a file, creating it if it doesn't exist
func WriteFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// CreateDirIfNotExists creates a directory if it doesn't exist
func CreateDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}
