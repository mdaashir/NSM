#!/bin/bash

# Check if version is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <new-version>"
    echo "Example: $0 1.2.0"
    exit 1
fi

NEW_VERSION=$1
VERSION_NO_V="${NEW_VERSION#v}" # Remove 'v' prefix if present

# Update version in snapcraft.yaml
sed -i "s/^version: .*/version: '$VERSION_NO_V'/" snap/snapcraft.yaml

# Create a new git tag
git tag -a "v$VERSION_NO_V" -m "Release v$VERSION_NO_V"

echo "Version bumped to $VERSION_NO_V"
echo "Don't forget to push the changes with:"
echo "git push && git push --tags"
