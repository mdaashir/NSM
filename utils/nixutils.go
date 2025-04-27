// Package utils provides utility functions for NSM
package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// GetPackageVersion retrieves the version of an installed package by parsing the nix-env output
func GetPackageVersion(pkg string) (string, error) {
	// Run nix-env --query --JSON to get package information
	cmd := exec.Command("nix-env", "--query", "--json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to query package info: %v", err)
	}

	// Parse JSON output
	var pkgInfo map[string]struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(output, &pkgInfo); err != nil {
		return "", fmt.Errorf("failed to parse package info: %v", err)
	}

	// Look for the package version
	for name, info := range pkgInfo {
		if strings.HasPrefix(name, "nixpkgs."+pkg) {
			return info.Version, nil
		}
	}

	return "", fmt.Errorf("package %s not found", pkg)
}

// GetNixVersion gets the installed Nix version
func GetNixVersion() (string, error) {
	cmd := exec.Command("nix", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Nix version: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CheckFlakeSupport checks if Nix flakes are enabled
func CheckFlakeSupport() bool {
	version, err := GetNixVersion()
	if err != nil {
		return false
	}

	// Extract version number from format "nix (Nix) 2.4.0"
	parts := strings.Fields(version)
	if len(parts) < 3 {
		return false
	}

	// Get the version string and remove any trailing characters
	versionStr := strings.TrimRight(parts[2], ")")

	// Parse major and minor versions
	var major, minor int
	if _, err := fmt.Sscanf(versionStr, "%d.%d", &major, &minor); err != nil {
		// Try parsing with trailing .0
		if _, err := fmt.Sscanf(versionStr, "%d.%d.0", &major, &minor); err != nil {
			return false
		}
	}

	// Flakes are supported in Nix 2.4 and above
	return major > 2 || (major == 2 && minor >= 4)
}

// GetChannelInfo gets the current Nixpkgs channel information
func GetChannelInfo() (string, error) {
	cmd := exec.Command("nix-channel", "--list")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get channel info: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetNixpkgsRevision gets the current Nixpkgs revision
func GetNixpkgsRevision() (string, error) {
	cmd := exec.Command("nix-instantiate", "--eval", "-E", "<nixpkgs>.rev")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get nixpkgs revision: %v", err)
	}
	return strings.Trim(string(output), "\"\n"), nil
}

// ValidatePackage checks if a package name is valid
func ValidatePackage(pkg string) bool {
	if pkg == "" {
		return false
	}

	// Allow alphanumeric, dash, dot, and underscore
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_."

	for _, c := range pkg {
		if !strings.ContainsRune(validChars, c) {
			return false
		}
	}
	return true
}

// ExtractShellNixPackages extracts package names from shell.nix content
func ExtractShellNixPackages(content string) []string {
	var packages []string
	lines := strings.Split(content, "\n")
	inPackages := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "packages = with pkgs; [") {
			inPackages = true
			continue
		}
		if inPackages {
			if strings.Contains(trimmed, "];") {
				break
			}
			// Skip empty lines and comments
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				// Split by whitespace in case there are multiple packages per line
				for _, pkg := range strings.Fields(trimmed) {
					pkg = strings.TrimRight(pkg, ",") // Remove trailing comma if present
					if pkg != "" {
						packages = append(packages, pkg)
					}
				}
			}
		}
	}
	return packages
}

// ExtractFlakePackages extracts package names from flake.nix content
func ExtractFlakePackages(content string) []string {
	var packages []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "buildInputs = [") {
			// Extract packages from the current line
			start := strings.Index(trimmed, "[")
			if start != -1 {
				end := strings.Index(trimmed, "]")
				if end != -1 {
					pkgPart := trimmed[start+1 : end]
					// Split by whitespace and clean each package name
					for _, pkg := range strings.Fields(pkgPart) {
						pkg = strings.TrimSpace(pkg)
						if pkg != "" {
							packages = append(packages, pkg)
						}
					}
				}
			}
			break // Since we found and processed the buildInputs line
		}
	}
	return packages
}

// GetInstalledPackages returns a list of installed packages
func GetInstalledPackages() ([]string, error) {
	cmd := exec.Command("nix-env", "--query")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query installed packages: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var packages []string
	for _, line := range lines {
		if pkg := strings.TrimSpace(line); pkg != "" {
			packages = append(packages, pkg)
		}
	}
	return packages, nil
}
