#!/bin/bash
# Post-installation script for NSM

set -e

# Create config directory if it doesn't exist
mkdir -p ~/.config/NSM

# Create completions directory if needed
BASH_COMPLETION_DIR="/etc/bash_completion.d"
ZSH_COMPLETION_DIR="/usr/share/zsh/site-functions"
FISH_COMPLETION_DIR="/usr/share/fish/vendor_completions.d"

# Generate and install shell completions
if [ -d "$BASH_COMPLETION_DIR" ]; then
    nsm completion bash >"$BASH_COMPLETION_DIR/nsm"
    echo "Installed bash completion"
fi

if [ -d "$ZSH_COMPLETION_DIR" ]; then
    nsm completion zsh >"$ZSH_COMPLETION_DIR/_nsm"
    echo "Installed zsh completion"
fi

if [ -d "$FISH_COMPLETION_DIR" ]; then
    nsm completion fish >"$FISH_COMPLETION_DIR/nsm.fish"
    echo "Installed fish completion"
fi

# Add NSM to PATH if not already there
if ! echo "$PATH" | grep -q "$(dirname $(which nsm))"; then
    echo 'export PATH="$PATH:$(dirname $(which nsm))"' >>~/.bashrc
    echo "Added NSM to PATH in ~/.bashrc"
fi

# Verify Nix installation
if ! command -v nix &>/dev/null; then
    echo "Nix not found. NSM requires Nix to function properly."
    echo "Please install Nix using one of these methods:"
    echo "  - Single-user installation: 'curl -L https://nixos.org/nix/install | sh'"
    echo "  - Multi-user installation: 'curl -L https://nixos.org/nix/install | sh -s -- --daemon'"
fi

echo "NSM has been successfully installed!"
echo "Run 'nsm --help' to get started."
