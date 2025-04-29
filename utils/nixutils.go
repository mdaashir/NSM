// Package utils provides utility functions for NSM
package utils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// NixCommand represents a command to be executed with proper error handling
type NixCommand struct {
	Cmd        string
	Args       []string
	WorkingDir string
	Timeout    time.Duration
}

// ExecuteWithTimeout executes a command with a timeout
func ExecuteWithTimeout(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// Kill the process if context canceled or timed out
		if err := cmd.Process.Kill(); err != nil {
			Debug("Failed to kill process: %v", err)
		}
		return nil, fmt.Errorf("command timed out: %v", ctx.Err())
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("command failed: %v\nstderr: %s", err, stderr.String())
		}
	}

	return stdout.Bytes(), nil
}

// Run executes a NixCommand with proper error handling
func (c *NixCommand) Run() (string, error) {
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second // Default timeout
	}

	Debug("Executing command: %s %v", c.Cmd, c.Args)

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)
	if c.WorkingDir != "" {
		cmd.Dir = c.WorkingDir
	}

	output, err := ExecuteWithTimeout(ctx, cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// IsNixInstalled checks if Nix is installed
func IsNixInstalled() bool {
	_, err := exec.LookPath("nix")
	if err != nil {
		Debug("Nix binary not found in PATH: %v", err)
		return false
	}
	return true
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
	return filepath.Clean(path), nil
}

// CheckFlakeSupport checks if Nix flakes are enabled
func CheckFlakeSupport() bool {
	cmd := &NixCommand{
		Cmd:     "nix",
		Args:    []string{"flake", "--version"},
		Timeout: 5 * time.Second,
	}

	_, err := cmd.Run()
	if err != nil {
		Debug("Flakes not supported: %v", err)
		return false
	}
	return true
}

// UpdateChannel updates the current Nix channel
func UpdateChannel() error {
	cmd := &NixCommand{
		Cmd:     "nix-channel",
		Args:    []string{"--update"},
		Timeout: 120 * time.Second, // Channel updates can take time
	}

	_, err := cmd.Run()
	return err
}

// CollectGarbage runs the Nix garbage collector
func CollectGarbage() error {
	cmd := &NixCommand{
		Cmd:     "nix-collect-garbage",
		Args:    []string{"-d"},
		Timeout: 120 * time.Second, // GC can take time
	}

	_, err := cmd.Run()
	return err
}

// GetSystemInfo returns basic system information
func GetSystemInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Get Nix version
	nixVersion, err := GetNixVersion()
	if err == nil {
		info["nix_version"] = nixVersion
	} else {
		Debug("Failed to get Nix version: %v", err)
	}

	// Get system architecture
	cmd := &NixCommand{
		Cmd:     "nix",
		Args:    []string{"eval", "--impure", "--expr", "builtins.currentSystem"},
		Timeout: 5 * time.Second,
	}

	if out, err := cmd.Run(); err == nil {
		info["system"] = strings.Trim(out, "\"\n ")
	} else {
		Debug("Failed to get system architecture: %v", err)
	}

	// Check if running on multi-user installation
	cmd = &NixCommand{
		Cmd:     "nix",
		Args:    []string{"show-config"},
		Timeout: 5 * time.Second,
	}

	if out, err := cmd.Run(); err == nil {
		info["multi_user"] = fmt.Sprintf("%t", strings.Contains(out, "sandbox = true"))
	}

	// Check for flake support
	info["flakes_enabled"] = fmt.Sprintf("%t", CheckFlakeSupport())

	// Check if in nix shell
	info["in_nix_shell"] = fmt.Sprintf("%t", IsInNixShell())

	return info, nil
}

// ValidatePackage checks if a package name is valid
func ValidatePackage(pkg string) bool {
	if pkg == "" {
		return false
	}

	// Basic validation of package name format
	validName := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`).MatchString(pkg)
	if !validName {
		Debug("Invalid package name format: %s", pkg)
		return false
	}

	// Check if package exists in nixpkgs
	cmd := &NixCommand{
		Cmd:     "nix-env",
		Args:    []string{"-qaP", pkg},
		Timeout: 10 * time.Second,
	}

	_, err := cmd.Run()
	if err != nil {
		Debug("Package validation failed for %s: %v", pkg, err)
		return false
	}

	return true
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

	// Try to run a basic nix command
	cmd := &NixCommand{
		Cmd:     "nix",
		Args:    []string{"--version"},
		Timeout: 5 * time.Second,
	}

	if _, err := cmd.Run(); err != nil {
		return fmt.Errorf("nix command failed: %v", err)
	}

	return nil
}

// GetChannelInfo gets information about the current Nix channel
func GetChannelInfo() (string, error) {
	cmd := &NixCommand{
		Cmd:     "nix-channel",
		Args:    []string{"--list"},
		Timeout: 5 * time.Second,
	}

	output, err := cmd.Run()
	if err != nil {
		return "", err
	}

	channel := strings.TrimSpace(output)
	if channel == "" {
		return "", fmt.Errorf("no channels configured")
	}

	return channel, nil
}

// GetNixVersion returns the installed Nix version
func GetNixVersion() (string, error) {
	cmd := &NixCommand{
		Cmd:     "nix",
		Args:    []string{"--version"},
		Timeout: 5 * time.Second,
	}

	output, err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// GetInstalledPackages returns a list of installed packages
func GetInstalledPackages() ([]string, error) {
	var packages []string

	// Check shell.nix first
	if FileExists("shell.nix") {
		pkgs, err := ExtractShellNixPackages("shell.nix")
		if err == nil {
			packages = append(packages, pkgs...)
		} else {
			Debug("Failed to extract packages from shell.nix: %v", err)
		}
	}

	// Check flake.nix if it exists
	if FileExists("flake.nix") {
		pkgs, err := ExtractFlakePackages("flake.nix")
		if err == nil {
			packages = append(packages, pkgs...)
		} else {
			Debug("Failed to extract packages from flake.nix: %v", err)
		}
	}

	if len(packages) == 0 {
		Debug("No packages found in shell.nix or flake.nix")
	}

	// Remove duplicates
	return removeDuplicates(packages), nil
}

// removeDuplicates removes duplicate items from a string slice
func removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// GetPackageVersion returns the version of a package
func GetPackageVersion(pkg string) (string, error) {
	if !ValidatePackage(pkg) {
		return "", fmt.Errorf("invalid package name: %s", pkg)
	}

	cmd := &NixCommand{
		Cmd:     "nix-env",
		Args:    []string{"-qa", "--json", pkg},
		Timeout: 10 * time.Second,
	}

	output, err := cmd.Run()
	if err != nil {
		return "", err
	}

	// Parse JSON output to get version
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
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
	cmd := &NixCommand{
		Cmd:     "nix",
		Args:    []string{"eval", "--raw", "nixpkgs.lib.version"},
		Timeout: 5 * time.Second,
	}

	output, err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// ExtractShellNixPackages extracts package list from shell.nix
func ExtractShellNixPackages(path string) ([]string, error) {
	if !FileExists(path) {
		return nil, fmt.Errorf("shell.nix file not found: %s", path)
	}

	content, err := ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read shell.nix: %v", err)
	}

	// Regular expression to match package lines
	re := regexp.MustCompile(`(?m)^\s*([a-zA-Z0-9_.-]+)\s*$`)
	matches := re.FindAllStringSubmatch(content, -1)

	var packages []string
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" && match[1] != "with" && match[1] != "pkgs" {
			packages = append(packages, match[1])
		}
	}

	return packages, nil
}

// ExtractFlakePackages extracts package list from flake.nix
func ExtractFlakePackages(path string) ([]string, error) {
	if !FileExists(path) {
		return nil, fmt.Errorf("flake.nix file not found: %s", path)
	}

	content, err := ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read flake.nix: %v", err)
	}

	// Regular expression to match package lines in buildInputs
	re := regexp.MustCompile(`buildInputs\s*=\s*(?:with[^;]*;)?\s*\[\s*([^\]]+)\s*\]`)
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return nil, fmt.Errorf("no packages found in flake.nix")
	}

	// Split package names and clean them
	packageSection := match[1]
	scanner := bufio.NewScanner(strings.NewReader(packageSection))

	var result []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract package name from line
		parts := strings.Fields(line)
		if len(parts) > 0 {
			pkg := strings.TrimSuffix(parts[0], ";")
			if pkg != "" {
				result = append(result, pkg)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing flake.nix: %v", err)
	}

	return result, nil
}

// PinPackage pins a package to a specific version
func PinPackage(pkg string, version string) error {
	if !ValidatePackage(pkg) {
		return fmt.Errorf("invalid package name: %s", pkg)
	}

	cmd := &NixCommand{
		Cmd:     "nix-env",
		Args:    []string{"--set", pkg, version},
		Timeout: 30 * time.Second,
	}

	_, err := cmd.Run()
	return err
}

// InitFlake initializes a Nix flake in the given directory
func InitFlake(dir string) error {
	// Check if flakes are supported
	if !CheckFlakeSupport() {
		return fmt.Errorf("nix flakes are not supported on this system")
	}

	if !DirExists(dir) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// Check if flake.nix already exists
	flakePath := filepath.Join(dir, "flake.nix")
	if FileExists(flakePath) {
		return fmt.Errorf("flake.nix already exists in %s", dir)
	}

	cmd := &NixCommand{
		Cmd:        "nix",
		Args:       []string{"flake", "init"},
		WorkingDir: dir,
		Timeout:    10 * time.Second,
	}

	_, err := cmd.Run()
	return err
}

// UpdateFlake updates a Nix flake in the given directory
func UpdateFlake(dir string) error {
	// Check if flakes are supported
	if !CheckFlakeSupport() {
		return fmt.Errorf("nix flakes are not supported on this system")
	}

	if !DirExists(dir) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// Check if flake.nix exists
	flakePath := filepath.Join(dir, "flake.nix")
	if !FileExists(flakePath) {
		return fmt.Errorf("flake.nix does not exist in %s", dir)
	}

	cmd := &NixCommand{
		Cmd:        "nix",
		Args:       []string{"flake", "update"},
		WorkingDir: dir,
		Timeout:    60 * time.Second, // Flake updates can take time
	}

	_, err := cmd.Run()
	return err
}

// ParsePackageList parses package list from a Nix file
func ParsePackageList(path string) ([]string, error) {
	if !FileExists(path) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

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
	if !DirExists(dir) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// Validate packages
	var validPackages []string
	for _, pkg := range packages {
		if ValidatePackage(pkg) {
			validPackages = append(validPackages, pkg)
		} else {
			Debug("Skipping invalid package: %s", pkg)
		}
	}

	content := fmt.Sprintf(`{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "nsm-managed-shell";

  packages = with pkgs; [
    %s
  ];

  shellHook = ''
    echo "🚀 Welcome to your Nix development environment!"
    echo "📦 Use 'nsm add <package>' to add more packages"
  '';
}`, strings.Join(validPackages, "\n    "))

	filePath := filepath.Join(dir, "shell.nix")
	return SafeWrite(filePath, []byte(content), 0600)
}

// GetNixShellEnv gets environment variables for a Nix shell
func GetNixShellEnv(dir string) (map[string]string, error) {
	if !DirExists(dir) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	// Check if shell.nix or flake.nix exists
	shellNixPath := filepath.Join(dir, "shell.nix")
	flakeNixPath := filepath.Join(dir, "flake.nix")

	if !FileExists(shellNixPath) && !FileExists(flakeNixPath) {
		return nil, fmt.Errorf("neither shell.nix nor flake.nix found in %s", dir)
	}

	cmd := &NixCommand{
		Cmd:        "nix-shell",
		Args:       []string{"--show-trace", "--run", "env"},
		WorkingDir: dir,
		Timeout:    30 * time.Second,
	}

	output, err := cmd.Run()
	if err != nil {
		return nil, err
	}

	env := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	return env, nil
}

// GetNixCacheDir gets the Nix cache directory
func GetNixCacheDir() (string, error) {
	cmd := &NixCommand{
		Cmd:     "nix",
		Args:    []string{"show-config"},
		Timeout: 5 * time.Second,
	}

	output, err := cmd.Run()
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "store-dir") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return filepath.Clean(strings.TrimSpace(parts[1])), nil
			}
		}
	}

	return "/nix/store", nil // fallback to default
}

// CleanNixCache cleans the Nix store cache
func CleanNixCache() error {
	cmd := &NixCommand{
		Cmd:     "nix-store",
		Args:    []string{"--gc"},
		Timeout: 120 * time.Second, // GC can take time
	}

	_, err := cmd.Run()
	return err
}

// InvalidateNixCache invalidates Nix binary cache
func InvalidateNixCache() error {
	cmd := &NixCommand{
		Cmd:     "nix-store",
		Args:    []string{"--verify", "--check-contents"},
		Timeout: 120 * time.Second, // Can be slow
	}

	_, err := cmd.Run()
	return err
}

// GetCurrentProfile gets the current Nix profile path
func GetCurrentProfile() (string, error) {
	cmd := &NixCommand{
		Cmd:     "nix-env",
		Args:    []string{"--profile"},
		Timeout: 5 * time.Second,
	}

	output, err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// ListProfileGenerations lists all profile generations
func ListProfileGenerations() ([]string, error) {
	cmd := &NixCommand{
		Cmd:     "nix-env",
		Args:    []string{"--list-generations"},
		Timeout: 5 * time.Second,
	}

	output, err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var generations []string
	for _, line := range strings.Split(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			generations = append(generations, line)
		}
	}
	return generations, nil
}

// RollbackProfile rolls back to the previous generation
func RollbackProfile() error {
	cmd := &NixCommand{
		Cmd:     "nix-env",
		Args:    []string{"--rollback"},
		Timeout: 30 * time.Second,
	}

	_, err := cmd.Run()
	return err
}
