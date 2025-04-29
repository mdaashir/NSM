// Package utils provides utility functions for NSM
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// IsNixInstalled checks if Nix is installed
func IsNixInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// IsInNixShell checks if currently in a Nix shell environment
func IsInNixShell() bool {
	return os.Getenv("IN_NIX_SHELL") != ""
}

// GetNixPath returns the path to the Nix installation
func GetNixPath() (string, error) {
	path, err := exec.LookPath("nix")
	if err != nil {
		return "", fmt.Errorf("nix not found in PATH: %v", err)
	}
	return path, nil
}

// CheckFlakeSupport checks if Nix flakes are enabled
func CheckFlakeSupport() bool {
	cmd := exec.Command("nix", "flake", "--version")
	err := cmd.Run()
	return err == nil
}

// UpdateChannel updates the current Nix channel
func UpdateChannel() error {
	cmd := exec.Command("nix-channel", "--update")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update channel: %v\nOutput: %s", err, output)
	}
	return nil
}

// CollectGarbage runs the Nix garbage collector
func CollectGarbage() error {
	cmd := exec.Command("nix-collect-garbage", "-d")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to collect garbage: %v\nOutput: %s", err, output)
	}
	return nil
}

// GetSystemInfo returns basic system information
func GetSystemInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Get Nix version
	if out, err := exec.Command("nix", "--version").Output(); err == nil {
		info["nix_version"] = strings.TrimSpace(string(out))
	}

	// Get system architecture
	if out, err := exec.Command("nix", "eval", "--impure", "--expr", "builtins.currentSystem").Output(); err == nil {
		info["system"] = strings.Trim(string(out), "\"")
	}

	return info, nil
}

// ValidatePackage checks if a package name is valid
func ValidatePackage(pkg string) bool {
	// Check if package exists in nixpkgs
	cmd := exec.Command("nix-env", "-qaP", pkg)
	return cmd.Run() == nil
}

// CheckNixInstallation verifies Nix is properly installed
func CheckNixInstallation() error {
	if !IsNixInstalled() {
		return fmt.Errorf("nix is not installed")
	}

	// Check if nix-env is available
	if _, err := exec.LookPath("nix-env"); err != nil {
		return fmt.Errorf("nix-env not found: %v", err)
	}

	// Check if nix-store is available
	if _, err := exec.LookPath("nix-store"); err != nil {
		return fmt.Errorf("nix-store not found: %v", err)
	}

	return nil
}

// GetChannelInfo gets information about the current Nix channel
func GetChannelInfo() (string, error) {
	cmd := exec.Command("nix-channel", "--list")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get channel info: %v", err)
	}

	channel := strings.TrimSpace(string(output))
	if channel == "" {
		return "", fmt.Errorf("no channels configured")
	}

	return channel, nil
}

// GetNixVersion returns the installed Nix version
func GetNixVersion() (string, error) {
	cmd := exec.Command("nix", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Nix version: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetInstalledPackages returns a list of installed packages
func GetInstalledPackages() ([]string, error) {
	var packages []string

	// Check shell.nix first
	if FileExists("shell.nix") {
		pkgs, err := ExtractShellNixPackages("shell.nix")
		if err == nil {
			packages = append(packages, pkgs...)
		}
	}

	// Check flake.nix if it exists
	if FileExists("flake.nix") {
		pkgs, err := ExtractFlakePackages("flake.nix")
		if err == nil {
			packages = append(packages, pkgs...)
		}
	}

	return packages, nil
}

// GetPackageVersion returns the version of a package
func GetPackageVersion(pkg string) (string, error) {
	cmd := exec.Command("nix-env", "-qa", "--json", pkg)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get package version: %v", err)
	}

	// Parse JSON output to get version
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse package info: %v", err)
	}

	for _, info := range result {
		if m, ok := info.(map[string]interface{}); ok {
			if version, ok := m["version"].(string); ok {
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("version not found for package %s", pkg)
}

// GetNixpkgsRevision gets the current Git revision of nixpkgs
func GetNixpkgsRevision() (string, error) {
	cmd := exec.Command("nix", "eval", "--raw", "nixpkgs.lib.version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get nixpkgs revision: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ExtractShellNixPackages extracts package list from shell.nix
func ExtractShellNixPackages(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read shell.nix: %v", err)
	}

	// Regular expression to match package lines
	re := regexp.MustCompile(`(?m)^\s*([a-zA-Z0-9_-]+)\s*$`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	var packages []string
	for _, match := range matches {
		if len(match) > 1 {
			packages = append(packages, match[1])
		}
	}

	return packages, nil
}

// ExtractFlakePackages extracts package list from flake.nix
func ExtractFlakePackages(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read flake.nix: %v", err)
	}

	// Regular expression to match package lines in buildInputs
	re := regexp.MustCompile(`buildInputs\s*=\s*\[\s*([^\]]+)\s*\]`)
	match := re.FindStringSubmatch(string(content))
	if len(match) < 2 {
		return nil, fmt.Errorf("no packages found in flake.nix")
	}

	// Split package names and clean them
	packages := strings.Split(match[1], "\n")
	var result []string
	for _, pkg := range packages {
		pkg = strings.TrimSpace(pkg)
		if pkg != "" && !strings.HasPrefix(pkg, "#") {
			result = append(result, pkg)
		}
	}

	return result, nil
}

// PinPackage pins a package to a specific version
func PinPackage(pkg string, version string) error {
	cmd := exec.Command("nix-env", "--set", pkg, version)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pin package: %v\nOutput: %s", err, output)
	}
	return nil
}

// InitFlake initializes a Nix flake in the given directory
func InitFlake(dir string) error {
	cmd := exec.Command("nix", "flake", "init")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to init flake: %v\nOutput: %s", err, output)
	}
	return nil
}

// UpdateFlake updates a Nix flake in the given directory
func UpdateFlake(dir string) error {
	cmd := exec.Command("nix", "flake", "update")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update flake: %v\nOutput: %s", err, output)
	}
	return nil
}

// ParsePackageList parses package list from a Nix file
func ParsePackageList(path string) ([]string, error) {
	if strings.HasSuffix(path, "shell.nix") {
		return ExtractShellNixPackages(path)
	}
	if strings.HasSuffix(path, "flake.nix") {
		return ExtractFlakePackages(path)
	}
	return nil, fmt.Errorf("unsupported file type: %s", path)
}

// GenerateShellNix generates a shell.nix file with given packages
func GenerateShellNix(dir string, packages []string) error {
	content := fmt.Sprintf(`{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    %s
  ];
}`, strings.Join(packages, "\n    "))

	return WriteFile(filepath.Join(dir, "shell.nix"), content)
}

// GetNixShellEnv gets environment variables for a Nix shell
func GetNixShellEnv(dir string) (map[string]string, error) {
	cmd := exec.Command("nix-shell", "--show-trace", "--run", "env")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get shell env: %v", err)
	}

	env := make(map[string]string)
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	return env, nil
}

// GetNixCacheDir gets the Nix cache directory
func GetNixCacheDir() (string, error) {
	cmd := exec.Command("nix", "show-config")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get nix config: %v", err)
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "store-dir") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "/nix/store", nil // fallback to default
}

// CleanNixCache cleans the Nix store cache
func CleanNixCache() error {
	cmd := exec.Command("nix-store", "--gc")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clean cache: %v\nOutput: %s", err, output)
	}
	return nil
}

// InvalidateNixCache invalidates Nix binary cache
func InvalidateNixCache() error {
	cmd := exec.Command("nix-store", "--verify", "--check-contents")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to invalidate cache: %v\nOutput: %s", err, output)
	}
	return nil
}

// GetCurrentProfile gets the current Nix profile path
func GetCurrentProfile() (string, error) {
	cmd := exec.Command("nix-env", "--profile")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current profile: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ListProfileGenerations lists all profile generations
func ListProfileGenerations() ([]string, error) {
	cmd := exec.Command("nix-env", "--list-generations")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list generations: %v", err)
	}

	var generations []string
	for _, line := range strings.Split(string(output), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			generations = append(generations, line)
		}
	}
	return generations, nil
}

// RollbackProfile rolls back to the previous generation
func RollbackProfile() error {
	cmd := exec.Command("nix-env", "--rollback")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to rollback: %v\nOutput: %s", err, output)
	}
	return nil
}
