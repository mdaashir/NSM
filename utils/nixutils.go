package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// CheckFlakeSupport checks if flakes are supported
func CheckFlakeSupport() bool {
	cmd := exec.Command("nix", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Flakes are supported in Nix 2.4+
	version := string(output)
	return strings.Contains(version, "nix (Nix) 2.") && !strings.Contains(version, "nix (Nix) 2.3")
}

// ValidatePackage checks if a package name is valid
func ValidatePackage(pkg string) bool {
	if pkg == "" {
		return false
	}

	// Basic validation - alphanumeric with some special chars
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9.\-_]+$`, pkg)
	return matched
}

// GetChannelInfo gets current channel information
func GetChannelInfo() (string, error) {
	cmd := exec.Command("nix-channel", "--list")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetNixpkgsRevision gets the current nixpkgs revision
func GetNixpkgsRevision() (string, error) {
	cmd := exec.Command("nix-instantiate", "--eval", "-E", "<nixpkgs>.lib.version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\""), nil
}

// ExtractShellNixPackages extracts package names from shell.nix content
func ExtractShellNixPackages(content string) []string {
	var packages []string

	// Find the packages section using regex
	re := regexp.MustCompile(`packages\s*=\s*with\s+pkgs;\s*\[([\s\S]*?)]`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return packages
	}

	// Extract package names
	pkgSection := matches[1]
	pkgRe := regexp.MustCompile(`\b([a-zA-Z0-9.\-_]+)\b`)
	for _, pkg := range pkgRe.FindAllString(pkgSection, -1) {
		if pkg != "with" && pkg != "pkgs" {
			packages = append(packages, pkg)
		}
	}

	return packages
}

// ExtractFlakePackages extracts package names from flake.nix content
func ExtractFlakePackages(content string) []string {
	var packages []string

	// Find the buildInputs section using regex
	re := regexp.MustCompile(`buildInputs\s*=\s*with\s+[^;]+;\s*\[([\s\S]*?)]`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return packages
	}

	// Extract package names
	pkgSection := matches[1]
	pkgRe := regexp.MustCompile(`\b([a-zA-Z0-9.\-_]+)\b`)
	for _, pkg := range pkgRe.FindAllString(pkgSection, -1) {
		if pkg != "with" && !strings.Contains(pkg, "pkgs") {
			packages = append(packages, pkg)
		}
	}

	return packages
}

// GetInstalledPackages returns a list of packages installed in the current env
func GetInstalledPackages() ([]string, error) {
	cmd := exec.Command("nix-env", "--query", "--installed")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var packages []string
	for _, line := range strings.Split(string(output), "\n") {
		if pkg := strings.TrimSpace(line); pkg != "" {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// PinPackage pins a specific package to a version
func PinPackage(pkg string, version string) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if config.Pins == nil {
		config.Pins = make(map[string]string)
	}

	// Check if a version exists in nixpkgs
	exists, err := checkVersionExists(pkg, version)
	if err != nil {
		return fmt.Errorf("failed to verify version: %w", err)
	}

	if !exists {
		return fmt.Errorf("version %s not found for package %s", version, pkg)
	}

	// Update pin in config
	config.Pins[pkg] = version

	// Save updated config
	err = SaveConfig(config)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func checkVersionExists(pkg string, version string) (bool, error) {
	cmd := exec.Command("nix", "search", "nixpkgs#"+pkg, "--json")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to search package: %w", err)
	}

	var searchResult map[string]interface{}
	if err := json.Unmarshal(output, &searchResult); err != nil {
		return false, fmt.Errorf("failed to parse search results: %w", err)
	}

	// Check if the version exists in search results
	for _, info := range searchResult {
		if pkgInfo, ok := info.(map[string]interface{}); ok {
			if ver, exists := pkgInfo["version"].(string); exists {
				if ver == version {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// GetNixVersion gets the current Nix version
func GetNixVersion() (string, error) {
	cmd := exec.Command("nix", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Nix version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetPackageVersion gets the version of a specific package
func GetPackageVersion(pkg string) (string, error) {
	cmd := exec.Command("nix-env", "-qa", pkg, "--json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get package version: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse package info: %w", err)
	}

	// Get the first matching package version
	for _, info := range result {
		if pkgInfo, ok := info.(map[string]interface{}); ok {
			if version, exists := pkgInfo["version"].(string); exists {
				return version, nil
			}
		}
	}
	return "", fmt.Errorf("no version found for package %s", pkg)
}
