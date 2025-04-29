//go:build !windows

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	// Minimum required disk space in bytes (1 GB)
	minRequiredDiskSpace uint64 = 1 * 1024 * 1024 * 1024
)

// getDiskSpace returns the available disk space in bytes for a given path
func getDiskSpace(path string) (uint64, error) {
	if path == "" {
		return 0, fmt.Errorf("empty path provided")
	}

	// Ensure the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return 0, fmt.Errorf("path does not exist: %s", path)
	}

	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("failed to get disk stats: %v", err)
	}

	// Available blocks * block size
	freeSpace := stat.Bavail * uint64(stat.Bsize)

	Debug("Disk space info for %s: Free: %d bytes, Total: %d bytes",
		path, freeSpace, stat.Blocks*uint64(stat.Bsize))

	return freeSpace, nil
}

// CheckUnixPermissions checks if the user has proper permissions for Nix operations
func CheckUnixPermissions() DoctorResult {
	result := DoctorResult{
		Name:        "Unix Permissions",
		Description: "Checking permissions for Nix operations",
		Status:      StatusUnknown,
	}

	// Check if user can access /nix directory
	nixDir := "/nix"
	if !DirExists(nixDir) {
		result.Status = StatusError
		result.Message = "The /nix directory does not exist"
		result.Fix = "Install Nix using: sh <(curl -L https://nixos.org/nix/install)"
		return result
	}

	// Check permission to read/write in /nix
	if err := unix.Access(nixDir, unix.R_OK|unix.W_OK); err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Insufficient permissions for /nix: %v", err)

		// Get current user and group
		uid := os.Getuid()
		gid := os.Getgid()

		result.Fix = fmt.Sprintf("Make sure your user (uid=%d, gid=%d) has proper permissions",
			uid, gid)
		return result
	}

	// Check if /nix/store exists and is accessible
	nixStore := "/nix/store"
	if !DirExists(nixStore) {
		result.Status = StatusError
		result.Message = "The /nix/store directory does not exist"
		result.Fix = "Reinstall Nix using: sh <(curl -L https://nixos.org/nix/install)"
		return result
	}

	// Check disk space
	freeSpace, err := getDiskSpace(nixStore)
	if err != nil {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Could not check disk space: %v", err)
		return result
	}

	if freeSpace < minRequiredDiskSpace {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Low disk space: %.2f GB available, recommended at least 1 GB",
			float64(freeSpace)/float64(1024*1024*1024))
		result.Fix = "Free up disk space or increase the size of the partition containing /nix"
		return result
	}

	result.Status = StatusOK
	result.Message = fmt.Sprintf("Proper permissions for Nix directories with %.2f GB available space",
		float64(freeSpace)/float64(1024*1024*1024))
	return result
}

// CheckNixDaemon checks if the Nix daemon is running (multi-user installation)
func CheckNixDaemon() DoctorResult {
	result := DoctorResult{
		Name:        "Nix Daemon",
		Description: "Checking if Nix daemon is running (for multi-user installations)",
		Status:      StatusUnknown,
	}

	// Check if /nix/var/nix/daemon-socket exists (multi-user installation)
	daemonSocket := "/nix/var/nix/daemon-socket/socket"
	if !FileExists(daemonSocket) {
		// Not a multi-user installation, which is fine
		result.Status = StatusOK
		result.Message = "Single-user Nix installation detected (no daemon required)"
		return result
	}

	// Try to check daemon process
	cmd := exec.Command("systemctl", "is-active", "nix-daemon.service")
	output, err := cmd.Output()

	if err != nil {
		// Try another way to check
		cmd = exec.Command("pgrep", "-f", "nix-daemon")
		_, err = cmd.Output()

		if err != nil {
			result.Status = StatusError
			result.Message = "Nix daemon is not running"
			result.Fix = "Start the Nix daemon with: sudo systemctl start nix-daemon.service"
			return result
		}
	}

	status := strings.TrimSpace(string(output))
	if status != "active" && status != "" {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Nix daemon service status: %s", status)
		result.Fix = "Ensure the daemon is running with: sudo systemctl start nix-daemon.service"
		return result
	}

	result.Status = StatusOK
	result.Message = "Nix daemon is running correctly"
	return result
}
