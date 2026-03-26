#!/bin/bash
# Bump version in internal/version/VERSION.txt and commit
# Usage: ./scripts/bump-version.sh [patch|minor|major|vX.Y.Z]

set -e

VERSION_FILE="internal/version/VERSION.txt"
CURRENT=$(cat "$VERSION_FILE" | tr -d '[:space:]')

# Remove 'v' prefix for parsing
CURRENT_NUM=${CURRENT#v}

# Parse current version
MAJOR=$(echo "$CURRENT_NUM" | cut -d. -f1)
MINOR=$(echo "$CURRENT_NUM" | cut -d. -f2)
PATCH=$(echo "$CURRENT_NUM" | cut -d. -f3)

# Determine new version
if [ "$1" == "major" ]; then
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
elif [ "$1" == "minor" ]; then
    MINOR=$((MINOR + 1))
    PATCH=0
elif [ "$1" == "patch" ]; then
    PATCH=$((PATCH + 1))
elif [[ "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    NEW_VERSION="$1"
else
    echo "Usage: $0 [patch|minor|major|vX.Y.Z]"
    echo "Current version: $CURRENT"
    exit 1
fi

if [ -z "$NEW_VERSION" ]; then
    NEW_VERSION="v$MAJOR.$MINOR.$PATCH"
fi

echo "$NEW_VERSION" > "$VERSION_FILE"
echo "📝 Version bumped: $CURRENT -> $NEW_VERSION"

git add "$VERSION_FILE"
git commit -m "Bump version to $NEW_VERSION"
echo "✅ Committed version bump"

git push
echo "📤 Pushed commit to remote"
