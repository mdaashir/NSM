#!/bin/bash
# Script to update documentation after a new release

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    VERSION=$(git describe --tags --always)
fi

echo "Updating documentation for version $VERSION..."

# Update version in README
sed -i "s/version: .*/version: $VERSION/" README.md

# Generate command documentation
echo "## Commands" >docs/commands.md
echo "" >>docs/commands.md
echo "This documentation is automatically generated." >>docs/commands.md
echo "" >>docs/commands.md

# Add each command's help to documentation
for cmd in add clean config convert doctor freeze info init list pin remove run upgrade; do
    echo "### nsm $cmd" >>docs/commands.md
    echo "" >>docs/commands.md
    echo '```' >>docs/commands.md
    ./nsm $cmd --help >>docs/commands.md
    echo '```' >>docs/commands.md
    echo "" >>docs/commands.md
done

# Update the CHANGELOG
if [ ! -z "$(git tag -l "$VERSION")" ]; then
    # Only update CHANGELOG if this is an actual tagged version
    DATE=$(date +"%Y-%m-%d")

    # Create new entry in CHANGELOG if this version isn't there yet
    if ! grep -q "## \[$VERSION\]" CHANGELOG.md; then
        sed -i "s/# Changelog/# Changelog\n\n## [$VERSION] - $DATE\n\n### Added\n\n### Changed\n\n### Fixed\n/" CHANGELOG.md
    fi
fi

echo "Documentation updated successfully!"
