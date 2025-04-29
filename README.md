# NSM (Nix Shell Manager)

A command-line tool to manage Nix development environments with ease.

## Features

- Initialize new Nix shell environments
- Add/remove packages easily
- List installed packages
- Convert between shell.nix and flake.nix
- Run Nix shells
- Manage Nix channel versions
- Clean up unused packages
- Freeze package versions for reproducibility
- Cross-platform compatibility (Linux, macOS, Windows with WSL)
- Comprehensive diagnostics with auto-fixing capabilities
- Multiple output formats (JSON, CSV, Markdown) for integration with other tools

## Installation

### Using Nix (recommended)

Using flake:

```bash
nix run github:mdaashir/NSM
```

Install globally:

```bash
nix profile install github:mdaashir/NSM
```

### Package Managers

#### macOS

```bash
brew install mdaashir/tap/nsm
```

#### Linux

Debian/Ubuntu:

```bash
curl -LO https://github.com/mdaashir/NSM/releases/latest/download/nsm_*_linux_amd64.deb
sudo dpkg -i nsm_*_linux_amd64.deb
```

Red Hat/Fedora:

```bash
sudo rpm -i https://github.com/mdaashir/NSM/releases/latest/download/nsm_*_linux_amd64.rpm
```

Alpine:

```bash
curl -LO https://github.com/mdaashir/NSM/releases/latest/download/nsm_*_linux_amd64.apk
sudo apk add --allow-untrusted ./nsm_*_linux_amd64.apk
```

Arch Linux:

```bash
curl -LO https://github.com/mdaashir/NSM/releases/latest/download/nsm_*_linux_amd64.pkg.tar.zst
sudo pacman -U nsm_*_linux_amd64.pkg.tar.zst
```

Snap:

```bash
sudo snap install nsm --classic
```

#### Windows

On Windows, NSM requires WSL (Windows Subsystem for Linux) with Nix installed:

1. Install WSL following [Microsoft's instructions](https://learn.microsoft.com/en-us/windows/wsl/install)
2. Install Nix inside your WSL distribution
3. Download and install the NSM Windows binary from the [releases page](https://github.com/mdaashir/NSM/releases)

#### Android (Termux)

```bash
pkg install nsm
```

### Docker

```bash
docker run ghcr.io/mdaashir/nsm:latest
```

### Binary Installation

Download the appropriate binary for your platform from the [releases page](https://github.com/mdaashir/NSM/releases).

### Build from source

Requirements:

- Go 1.24 or later
- Make (optional)

```bash
git clone https://github.com/mdaashir/NSM.git
cd NSM
make build
```

Or using Go directly:

```bash
go install github.com/mdaashir/NSM@latest
```

## Usage

### Initialize a New Environment

```bash
nsm init              # Create new shell.nix
nsm init --flake      # Create new flake.nix
```

### Manage Packages

```bash
nsm add gcc python3   # Add packages
nsm remove gcc        # Remove packages
nsm list              # List installed packages
nsm list --json       # List packages in JSON format
```

### Development Environment

```bash
nsm run               # Enter the Nix shell
nsm run --cmd "make"  # Run a command in the Nix shell
```

### Maintenance

```bash
nsm clean             # Clean up unused packages
nsm upgrade           # Update nixpkgs channel
nsm doctor            # Check installation health
nsm doctor --fix      # Fix common issues automatically
nsm doctor --json     # Output diagnostics in JSON format
nsm doctor --md       # Output diagnostics in Markdown format
nsm doctor --csv      # Output diagnostics in CSV format
nsm doctor --no-color # Disable colored output
```

### Configuration

```bash
nsm config set default.packages "gcc python3"
nsm config get default.packages
nsm config reset      # Reset to defaults
```

### Advanced Features

```bash
nsm convert           # Convert shell.nix to flake.nix
nsm freeze            # Pin nixpkgs version
nsm info              # Show system information
nsm info --json       # System info in JSON format
nsm pin gcc 13.2.0    # Pin package to specific version
```

## Configuration

Configuration file is stored in:
- Linux/macOS: `$HOME/.config/NSM/config.yaml`
- Windows: `%USERPROFILE%\.config\NSM\config.yaml`

Available settings:

- `default.packages`: Default packages for new environments
- `channel.url`: Default Nixpkgs channel URL
- `shell.format`: Preferred format (shell.nix/flake.nix)
- `logging.level`: Log level (debug, info, warn, error)
- `logging.file`: Path to log file

## Shell File Format

### shell.nix

```nix
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    # Your packages here
    gcc
    python3
  ];

  shellHook = ''
    echo "🚀 Welcome to your Nix development environment!"
  '';
}
```

### flake.nix (after conversion)

```nix
{
  description = "Development environment";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = let
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
    in pkgs.mkShell {
      buildInputs = with pkgs; [
        # Your packages here
        gcc
        python3
      ];

      shellHook = ''
        echo "🚀 Welcome to your Nix development environment!"
      '';
    };
  };
}
```

## Error Handling

NSM now includes more robust error handling, with clear error messages and recovery options. Common issues are detected by the `doctor` command and can often be auto-fixed.

## Cross-Platform Support

NSM now works across Linux, macOS, and Windows (with WSL). The `doctor` command includes platform-specific checks to ensure proper installation and usage.

## Diagnostic Capabilities

The enhanced `doctor` command provides comprehensive diagnostics:

```bash
nsm doctor
```

This checks for:
- Nix installation status
- Channel configuration
- Store permissions
- Available disk space
- Flakes support
- Project configuration
- Platform-specific requirements
- And more...

## License

MIT License - See LICENSE file for details

## Author

Mohamed Aashir S <s.mohamedaashir@gmail.com>
