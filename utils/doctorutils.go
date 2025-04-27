package utils

import (
	"fmt"
	"os/exec"
)

// CheckNixInstallation verifies that Nix is installed on the system
func CheckNixInstallation() error {
	_, err := exec.LookPath("nix-env")
	if err != nil {
		return fmt.Errorf("nix-env command not found: %v", err)
	}
	return nil
}
