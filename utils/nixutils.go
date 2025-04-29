// Package utils provides utility functions for NSM
package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// RetryConfig defines parameters for retry behavior
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Factor      float64 // exponential backoff factor
}

// DefaultRetryConfig provides standard retry parameters
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   100 * time.Millisecond,
	MaxDelay:    5 * time.Second,
	Factor:      1.5,
}

// NixCommand represents a structured command executor for nix operations
type NixCommand struct {
	Command      string
	Args         []string
	Dir          string
	Timeout      time.Duration
	RetryConfig  RetryConfig
	Environment  []string
	IgnoreErrors bool
}

var (
	// ErrCommandTimeout indicates the command exceeded its time limit
	ErrCommandTimeout = errors.New("command execution timed out")

	// ErrMaxRetriesExceeded indicates all retry attempts have failed
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")
)

// ExecuteWithContext runs a nix command with context and returns the output
func (nc *NixCommand) ExecuteWithContext(ctx context.Context) (string, error) {
	var lastErr error
	var output string

	// Apply default retry config if not specified
	if nc.RetryConfig.MaxAttempts == 0 {
		nc.RetryConfig = DefaultRetryConfig
	}

	// Apply default timeout if not specified
	if nc.Timeout == 0 {
		nc.Timeout = 30 * time.Second
	}

	for attempt := 0; attempt < nc.RetryConfig.MaxAttempts; attempt++ {
		// If not the first attempt, delay according to backoff strategy
		if attempt > 0 {
			delay := calculateBackoff(attempt, nc.RetryConfig)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
				// Continue after delay
			}
			Debug("Retrying command execution, attempt %d/%d after %v delay",
				attempt+1, nc.RetryConfig.MaxAttempts, delay)
		}

		// Create execution context with timeout
		execCtx, cancel := context.WithTimeout(ctx, nc.Timeout)
		defer cancel()

		cmd := exec.CommandContext(execCtx, nc.Command, nc.Args...)
		if nc.Dir != "" {
			cmd.Dir = nc.Dir
		}
		if len(nc.Environment) > 0 {
			cmd.Env = nc.Environment
		}

		// Capture stdout and stderr
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		// Check for timeout
		if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			lastErr = ErrCommandTimeout
			Warn("Command execution timed out after %v", nc.Timeout)
			continue
		}

		// If command successful, return output
		if err == nil {
			output = stdout.String()
			return output, nil
		}

		// Command failed
		lastErr = fmt.Errorf("command execution failed: %w\nstderr: %s", err, stderr.String())
		Error("Command failed (attempt %d/%d): %v",
			attempt+1, nc.RetryConfig.MaxAttempts, lastErr)

		// If the error is not retriable, break immediately
		if !isRetriableError(stderr.String()) {
			break
		}
	}

	if nc.IgnoreErrors {
		Warn("Ignoring error: %v", lastErr)
		return "", nil
	}

	return "", fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, lastErr)
}

// Execute runs a nix command with default context
func (nc *NixCommand) Execute() (string, error) {
	return nc.ExecuteWithContext(context.Background())
}

// calculateBackoff determines delay time using exponential backoff with jitter
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	// Calculate exponential delay
	delay := float64(config.BaseDelay)
	for i := 0; i < attempt; i++ {
		delay *= config.Factor
	}

	// Add jitter (±20%) to prevent thundering herd
	jitterPercent := 0.2
	jitter := delay * jitterPercent
	min := delay - jitter
	max := delay + jitter

	// Convert to duration with random jitter
	jitteredDelay := min + (max-min)*float64(time.Now().Nanosecond())/1e9

	// Cap at max delay
	if jitteredDelay > float64(config.MaxDelay) {
		jitteredDelay = float64(config.MaxDelay)
	}

	return time.Duration(jitteredDelay)
}

// isRetriableError determines if an error should trigger a retry
func isRetriableError(stderr string) bool {
	retriablePatterns := []string{
		"connection reset by peer",
		"connection refused",
		"temporarily unavailable",
		"timed out",
		"network is unreachable",
		"resource temporarily unavailable",
	}

	for _, pattern := range retriablePatterns {
		if strings.Contains(strings.ToLower(stderr), pattern) {
			return true
		}
	}

	return false
}

// Additional utility functions for nix operations

// IsNixInstalled checks if nix is installed on the system
func IsNixInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// GetNixVersion returns the installed nix version
func GetNixVersion() (string, error) {
	cmd := NixCommand{
		Command: "nix",
		Args:    []string{"--version"},
		Timeout: 5 * time.Second,
	}
	output, err := cmd.Execute()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// ExecuteNixShell executes a command in a nix-shell
func ExecuteNixShell(nixFile string, command string, args []string) (string, error) {
	shellArgs := []string{"-f", nixFile, "--run", command}
	shellArgs = append(shellArgs, args...)

	cmd := NixCommand{
		Command: "nix-shell",
		Args:    shellArgs,
		Timeout: 2 * time.Minute, // Longer timeout for shell commands
	}

	return cmd.Execute()
}

// IsDirtyWorkingTree checks if the git working tree has uncommitted changes
func IsDirtyWorkingTree(repoPath string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	// If output is not empty, working tree is dirty
	return len(output) > 0, nil
}

// GetNixPlatformSupport returns information about nix support on the current platform
func GetNixPlatformSupport() map[string]interface{} {
	support := map[string]interface{}{
		"platform":     runtime.GOOS,
		"architecture": runtime.GOARCH,
		"isSupported":  true,
		"details":      "",
	}

	// Check platform support
	switch runtime.GOOS {
	case "linux", "darwin":
		support["isSupported"] = true
	case "windows":
		support["isSupported"] = false
		support["details"] = "Windows requires WSL or special configuration for Nix"
	default:
		support["isSupported"] = false
		support["details"] = "Untested platform"
	}

	return support
}

// Command execution pool for running multiple nix commands concurrently
var commandPool = sync.Pool{
	New: func() interface{} {
		return &NixCommand{}
	},
}

// ExecuteNixCommandPool reuses NixCommand objects from a pool
func ExecuteNixCommandPool(command string, args []string) (string, error) {
	cmd := commandPool.Get().(*NixCommand)
	defer commandPool.Put(cmd)

	// Reset the command object
	*cmd = NixCommand{
		Command: command,
		Args:    args,
		Timeout: 30 * time.Second,
	}

	return cmd.Execute()
}

// NixShellCommand creates a command to run in nix-shell with the given package
func NixShellCommand(pkg string, command string) *NixCommand {
	return &NixCommand{
		Command: "nix-shell",
		Args:    []string{"-p", pkg, "--run", command},
		Timeout: 60 * time.Second,
	}
}

// FindNixFiles searches for .nix files in the given directory
func FindNixFiles(directory string) ([]string, error) {
	pattern := filepath.Join(directory, "*.nix")
	return filepath.Glob(pattern)
}

// ValidateNixExpression checks if a nix expression is valid
func ValidateNixExpression(expression string) error {
	cmd := NixCommand{
		Command: "nix-instantiate",
		Args:    []string{"--eval", "--expr", expression},
		Timeout: 10 * time.Second,
	}

	_, err := cmd.Execute()
	return err
}
