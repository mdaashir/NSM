#!/bin/sh
# Pre-removal script for NSM

set -e

echo "Removing NSM..."
# No need to remove config files, as they might contain user preferences
# Config files are stored in ${HOME}/.config/NSM/

# Optional: Notify user about remaining config files
echo "Note: Configuration files in ${HOME}/.config/NSM/ will be preserved."
echo "To completely remove NSM, delete this directory manually."
