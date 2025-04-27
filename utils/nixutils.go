// Package utils provides utility functions for NSM
package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"unicode"
)

// CheckNixInstallation verifies Nix is properly installed
func CheckNixInstallation() error {
	if _, err := exec.LookPath("nix-env"); err != nil {
		return fmt.Errorf("nix-env not found in PATH: %v", err)
	}

	if _, err := exec.LookPath("nix-shell"); err != nil {
		return fmt.Errorf("nix-shell not found in PATH: %v", err)
	}

	return nil
}

// ValidatePackage checks if a package name is valid
func ValidatePackage(pkg string) bool {
	// Basic validation
	if pkg == "" || len(pkg) > 128 {
		return false
	}

	// Check for valid characters
	for _, r := range pkg {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' && r != '.' {
			return false
		}
	}

	// Check if package exists
	cmd := exec.Command("nix-env", "-qa", pkg)
	if err := cmd.Run(); err != nil {
		Debug("Package validation failed for %s: %v", pkg, err)
		return false
	}

	return true
}

// GetPackageVersion gets the installed version of a package
func GetPackageVersion(pkg string) (string, error) {
	if !ValidatePackage(pkg) {
		return "", fmt.Errorf("invalid package name: %s", pkg)
	}

	cmd := exec.Command("nix-env", "-qa", "--json", pkg)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get package version: %v", err)
	}

	var pkgInfo map[string]interface{}
	if err := json.Unmarshal(output, &pkgInfo); err != nil {
		return "", fmt.Errorf("failed to parse package info: %v", err)
	}

	for _, info := range pkgInfo {
		if version, ok := info.(map[string]interface{})["version"].(string); ok {
			return version, nil
		}
	}

	return "", fmt.Errorf("version not found for package: %s", pkg)
}

// GetChannelInfo gets information about the current Nix channel
func GetChannelInfo() (string, error) {
	cmd := exec.Command("nix-channel", "--list")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get channel info: %v", err)
	}

	channels := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, channel := range channels {
		if strings.HasPrefix(channel, "nixpkgs ") {
			return strings.TrimPrefix(channel, "nixpkgs "), nil
		}
	}

	return "", fmt.Errorf("nixpkgs channel not found")
}

// GetNixpkgsRevision gets the current Git revision of nixpkgs
func GetNixpkgsRevision() (string, error) {
	cmd := exec.Command("nix-instantiate", "--eval", "-E", "with import <nixpkgs> {}; lib.version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get nixpkgs revision: %v", err)
	}

	revision := strings.Trim(string(output), "\"'\n")
	if revision == "" {
		return "", fmt.Errorf("empty nixpkgs revision")
	}

	return revision, nil
}

// CheckFlakeSupport checks if Nix flakes are enabled
func CheckFlakeSupport() bool {
	cmd := exec.Command("nix", "--version")
	output, err := cmd.Output()
	if err != nil {
		Debug("Failed to get Nix version: %v", err)
		return false
	}

	version := string(output)
	if strings.Contains(version, "nix (Nix) 2.4") || strings.Contains(version, "nix (Nix) 2.3") {
		// Check if experimental features are enabled
		cmd = exec.Command("nix", "show-config")
		output, err = cmd.Output()
		if err != nil {
			Debug("Failed to get Nix config: %v", err)
			return false
		}

		return strings.Contains(string(output), "experimental-features = nix-command flakes")
	}

	// Nix 2.5+ has flakes enabled by default
	return true
}

// GetInstalledPackages returns a list of installed packages
func GetInstalledPackages() ([]string, error) {
	cmd := exec.Command("nix-env", "--query", "--installed", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get installed packages: %v", err)
	}

	var pkgInfo map[string]interface{}
	if err := json.Unmarshal(output, &pkgInfo); err != nil {
		return nil, fmt.Errorf("failed to parse package info: %v", err)
	}

	var packages []string
	for name := range pkgInfo {
		packages = append(packages, name)
	}

	return packages, nil
}

// UpdateChannel updates the Nix channel
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

// GetSystemInfo returns information about the Nix installation
func GetSystemInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Get Nix version
	cmd := exec.Command("nix", "--version")
	if output, err := cmd.Output(); err == nil {
		info["nix_version"] = strings.TrimSpace(string(output))
	}

	// Get channel info
	if channel, err := GetChannelInfo(); err == nil {
		info["channel"] = channel
	}

	// Get nixpkgs revision
	if revision, err := GetNixpkgsRevision(); err == nil {
		info["nixpkgs_revision"] = revision
	}

	// Get system features
	info["flakes_enabled"] = fmt.Sprintf("%v", CheckFlakeSupport())

	return info, nil
}
