#!/bin/bash
set -e

VERSION=$(cat VERSION)

echo "🚀 Creating release for version v$VERSION"

# Validate semantic versioning format
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ VERSION file must contain semantic version (e.g., 1.0.0)"
    echo "Current value: $VERSION"
    exit 1
fi

# Check if tag already exists
if git tag -l | grep -q "^v$VERSION$"; then
    echo "❌ Tag v$VERSION already exists"
    echo "Update VERSION file or delete existing tag"
    exit 1
fi

# Validate we have binaries
if [ ! -d "bin" ] || [ -z "$(ls -A bin/)" ]; then
    echo "❌ No binaries found in bin/ directory"
    echo "Run: make build-all"
    exit 1
fi

echo "📦 Generating checksums..."
cd bin
sha256sum * > checksums.txt
cd ..

echo "🏷️  Creating git tag..."
git tag "v$VERSION"
git push origin "v$VERSION"

echo "✅ Release v$VERSION tagged successfully!"
echo "🔗 Binaries and checksums available in bin/"
echo "📝 To create GitHub release manually:"
echo "   1. Go to: https://github.com/bdchatham/AphexCLI/releases/new"
echo "   2. Select tag: v$VERSION"
echo "   3. Upload files from bin/ directory"
echo "   4. Add release notes"
