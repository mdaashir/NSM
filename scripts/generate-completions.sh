#!/bin/bash
# Script to generate shell completions for various shells

set -e

# Ensure output directory exists
mkdir -p completions

echo "Generating shell completions..."

# Generate completions for various shells
./nsm completion bash >completions/nsm.bash
./nsm completion zsh >completions/nsm.zsh
./nsm completion fish >completions/nsm.fish
./nsm completion powershell >completions/nsm.ps1

echo "Shell completions generated in completions/ directory"
