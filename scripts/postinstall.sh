#!/bin/sh
# Post-installation script for NSM

set -e

# Create config directory if needed
mkdir -p "${HOME}/.config/NSM"

# Run the doctor command to check system configuration
nsm doctor || true

echo "NSM has been successfully installed!"
echo "Run 'nsm --help' for usage information."
