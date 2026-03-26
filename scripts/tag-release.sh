#!/bin/bash
# Create git tag from VERSION.txt
# Workflow: update VERSION.txt -> commit -> run this script to tag

set -e

VERSION=$(cat internal/version/VERSION.txt | tr -d '[:space:]')

if [ -z "$VERSION" ]; then
    echo "❌ Error: VERSION.txt is empty"
    exit 1
fi

if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "❌ Error: Tag $VERSION already exists"
    exit 1
fi

echo "🏷️  Creating tag $VERSION..."
git tag "$VERSION"
echo "✅ Tag $VERSION created!"

git push origin "$VERSION"
echo "📤 Pushed tag to remote"
