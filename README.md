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

## Installation

### Prerequisites

- Nix package manager installed
- Go 1.24 or later

### Install from source

```bash
go install github.com/mdaashir/NSM@latest
```

## Usage

### Initialize a New Environment

```bash
nsm init              # Create new shell.nix
```

### Manage Packages

```bash
nsm add gcc python3   # Add packages
nsm remove gcc        # Remove packages
nsm list              # List installed packages
```

### Development Environment

```bash
nsm run              # Enter the Nix shell
```

### Maintenance

```bash
nsm clean            # Clean up unused packages
nsm upgrade          # Update nixpkgs channel
nsm doctor           # Check installation health
```

### Configuration

```bash
nsm config set default.packages "gcc python3"
nsm config get default.packages
```

### Advanced Features

```bash
nsm convert          # Convert shell.nix to flake.nix
nsm freeze           # Pin nixpkgs version
nsm info             # Show system information
```

## Configuration

Configuration file is stored in `$HOME/.config/NSM/config.yaml`

Available settings:

- `default.packages`: Default packages for new environments
- `channel.url`: Default Nixpkgs channel URL
- `shell.format`: Preferred format (shell.nix/flake.nix)

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
}
```

### flake.nix (after conversion)

```nix
{
  description = "Development environment";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = nixpkgs.mkShell {
      buildInputs = [
        # Your packages here
      ];
    };
  };
}
```

## License

MIT License - See LICENSE file for details

## Author

Mohamed Aashir S <s.mohamedaashir@gmail.com>
