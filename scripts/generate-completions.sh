#!/bin/bash
# This script generates shell completions for NSM

set -e

BINARY_PATH=$1
if [ -z "$BINARY_PATH" ]; then
    echo "Error: Binary path not provided"
    echo "Usage: $0 <binary-path>"
    exit 1
fi

# Create completions directory if it doesn't exist
mkdir -p completions

# Generate Bash completions
echo "Generating Bash completions..."
$BINARY_PATH completion bash >completions/nsm.bash

# Generate Zsh completions
echo "Generating Zsh completions..."
$BINARY_PATH completion zsh >completions/nsm.zsh

# Generate Fish completions
echo "Generating Fish completions..."
$BINARY_PATH completion fish >completions/nsm.fish

echo "Shell completions generated successfully in the completions/ directory."
