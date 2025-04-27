package utils

import (
	"fmt"
	"os/exec"
)

// CheckNixInstallation verifies that Nix is installed on the system
func CheckNixInstallation() error {
	_, err := exec.LookPath("nix-env")
	if err != nil {
		return fmt.Errorf("nix-env command not found. To install Nix:\n\n" +
			"Windows (via WSL2):\n" +
			"1. Enable and setup WSL2\n" +
			"2. Install Ubuntu or another Linux distro from Microsoft Store\n" +
			"3. In WSL2, run: sh <(curl -L https://nixos.org/nix/install) --daemon\n\n" +
			"Linux:\n" +
			"Run: sh <(curl -L https://nixos.org/nix/install) --daemon\n\n" +
			"macOS:\n" +
			"Run: sh <(curl -L https://nixos.org/nix/install)\n\n" +
			"For more information, visit: https://nixos.org/download.html")
	}
	return nil
}
