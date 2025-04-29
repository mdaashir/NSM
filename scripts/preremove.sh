#!/bin/bash
# Pre-removal script for NSM

set -e

echo "Cleaning up NSM before removal..."

# Remove completions
BASH_COMPLETION_DIR="/etc/bash_completion.d"
ZSH_COMPLETION_DIR="/usr/share/zsh/site-functions"
FISH_COMPLETION_DIR="/usr/share/fish/vendor_completions.d"

if [ -f "$BASH_COMPLETION_DIR/nsm" ]; then
    rm "$BASH_COMPLETION_DIR/nsm"
    echo "Removed bash completion"
fi

if [ -f "$ZSH_COMPLETION_DIR/_nsm" ]; then
    rm "$ZSH_COMPLETION_DIR/_nsm"
    echo "Removed zsh completion"
fi

if [ -f "$FISH_COMPLETION_DIR/nsm.fish" ]; then
    rm "$FISH_COMPLETION_DIR/nsm.fish"
    echo "Removed fish completion"
fi

echo "NSM pre-removal cleanup completed successfully."
echo "Note: Your Nix packages and ~/.config/NSM directory have been preserved."
echo "To completely remove all NSM data, manually delete ~/.config/NSM."
