//go:build windows

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
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

	var free, total, avail uint64

	// Load Windows API
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return 0, fmt.Errorf("failed to load kernel32.dll: %v", err)
	}

	proc, err := kernel32.FindProc("GetDiskFreeSpaceExW")
	if err != nil {
		return 0, fmt.Errorf("failed to find GetDiskFreeSpaceExW function: %v", err)
	}

	// Convert path to UTF16 pointer
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("failed to convert path to UTF16: %v", err)
	}

	// Call Windows API
	ret, _, err := proc.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&free)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&avail)),
	)

	// Windows syscalls return non-zero error even on success
	if ret == 0 {
		return 0, fmt.Errorf("GetDiskFreeSpaceExW failed: %v", err)
	}

	Debug("Disk space info for %s: Free: %d bytes, Total: %d bytes, Available: %d bytes",
		path, free, total, avail)

	return free, nil
}

// CheckDiskSpace checks if there's enough disk space available for Nix operations
func CheckDiskSpace() *DoctorResult {
	result := &DoctorResult{
		Name:        "Disk Space",
		Description: "Checking available disk space for Nix operations",
		Status:      StatusUnknown,
	}

	// Get home directory for config
	home, err := os.UserHomeDir()
	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Failed to get home directory: %v", err)
		return result
	}

	// Check space in home directory
	freeSpace, err := getDiskSpace(home)
	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Failed to check disk space: %v", err)
		return result
	}

	// Check if there's enough space
	if freeSpace < minRequiredDiskSpace {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Low disk space: %.2f GB available, recommended at least 1 GB",
			float64(freeSpace)/float64(1024*1024*1024))
	} else {
		result.Status = StatusOK
		result.Message = fmt.Sprintf("%.2f GB available disk space",
			float64(freeSpace)/float64(1024*1024*1024))
	}

	return result
}

// CheckWindowsSpecific performs Windows-specific checks
func CheckWindowsSpecific() *DoctorResult {
	result := &DoctorResult{
		Name:        "Windows Compatibility",
		Description: "Checking Windows-specific requirements for Nix",
		Status:      StatusUnknown,
	}

	// Check if running Windows 10 or later
	// Version info is in osVersion global variable in Windows
	info := GetWindowsVersionInfo()

	// Check WSL availability (required for Nix on Windows)
	wslEnabled := CheckWSLEnabled()

	if !wslEnabled {
		result.Status = StatusError
		result.Message = "WSL (Windows Subsystem for Linux) is not enabled. Nix requires WSL on Windows."
		result.Fix = "Enable WSL by running 'dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart' in an admin PowerShell"
		return result
	}

	result.Status = StatusOK
	result.Message = fmt.Sprintf("Windows %s build %d. WSL is enabled.",
		info["version"], info["build"])

	return result
}

// GetWindowsVersionInfo returns Windows version information
func GetWindowsVersionInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Call GetVersionExW to get Windows version info
	var osVersionInfo syscall.Osversioninfo
	osVersionInfo.Size = uint32(unsafe.Sizeof(osVersionInfo))

	// NOTE: This API is deprecated but still works for basic version checking
	syscall.GetVersionEx(&osVersionInfo)

	// Parse version information
	info["major"] = int(osVersionInfo.MajorVersion)
	info["minor"] = int(osVersionInfo.MinorVersion)
	info["build"] = int(osVersionInfo.BuildNumber)

	// Determine Windows version
	if osVersionInfo.MajorVersion == 10 {
		info["version"] = "10 or 11"
	} else if osVersionInfo.MajorVersion == 6 && osVersionInfo.MinorVersion == 3 {
		info["version"] = "8.1"
	} else if osVersionInfo.MajorVersion == 6 && osVersionInfo.MinorVersion == 2 {
		info["version"] = "8"
	} else if osVersionInfo.MajorVersion == 6 && osVersionInfo.MinorVersion == 1 {
		info["version"] = "7"
	} else {
		info["version"] = fmt.Sprintf("%d.%d", osVersionInfo.MajorVersion, osVersionInfo.MinorVersion)
	}

	return info
}

// CheckWSLEnabled checks if WSL is enabled on Windows
func CheckWSLEnabled() bool {
	// Try to run a basic WSL command to check if it's available
	cmd := exec.Command("wsl", "--list", "--verbose")
	err := cmd.Run()
	return err == nil
}
