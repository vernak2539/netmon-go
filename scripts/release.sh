#!/usr/bin/env bash
set -euo pipefail

if [ -z "${1:-}" ]; then
  echo "Usage: make release VERSION=v1.0.0 (or ./scripts/release.sh v1.0.0)"
  exit 1
fi

VERSION="$1"

# Ensure version tag starts with 'v'
if [[ "$VERSION" != v* ]]; then
  VERSION="v$VERSION"
fi

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
  echo "Error: Working directory has uncommitted changes. Please commit or stash them before releasing."
  exit 1
fi

# Check if tag already exists locally
if git rev-parse "$VERSION" >/dev/null 2>&1; then
  echo "Error: Tag '$VERSION' already exists locally."
  exit 1
fi

# Check if tag already exists remotely
if git ls-remote --tags origin | grep -q "refs/tags/$VERSION$"; then
  echo "Error: Tag '$VERSION' already exists on remote 'origin'."
  exit 1
fi

echo "Running tests and linter before release..."
go vet ./...
go test -v -race ./...

echo "Creating release tag '$VERSION'..."
git tag -a "$VERSION" -m "Release $VERSION"

echo "Pushing tag '$VERSION' to origin..."
git push origin "$VERSION"

echo "Successfully tagged and pushed '$VERSION' to remote! GitHub Actions release workflow triggered."
