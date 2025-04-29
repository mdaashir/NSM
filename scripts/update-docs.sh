#!/bin/bash
# This script updates documentation with the new version

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Error: Version not provided"
    echo "Usage: $0 <version>"
    exit 1
fi

# Update version references in documentation
echo "Updating version references in documentation..."

# Remove 'v' prefix if present for version comparison
VERSION_NUM=${VERSION#v}

# Update README.md with new version
sed -i "s/NSM v[0-9]\+\.[0-9]\+\.[0-9]\+/NSM $VERSION/g" README.md

echo "Documentation updated successfully for version $VERSION."
