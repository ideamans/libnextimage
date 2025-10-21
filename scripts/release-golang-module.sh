#!/bin/bash

# Release Go module after verifying binary builds are complete
# Usage: bash scripts/release-golang-module.sh v0.4.0

set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v0.4.0"
  exit 1
fi

VERSION="$1"

# Ensure version starts with 'v'
if [[ "$VERSION" != v* ]]; then
  VERSION="v$VERSION"
fi

echo "=== Go Module Release Script ==="
echo "Version: $VERSION"
echo ""

# Check if repository tag exists
echo "Checking if repository tag $VERSION exists..."
if ! git rev-parse "$VERSION" >/dev/null 2>&1; then
  echo "❌ Error: Tag $VERSION does not exist"
  echo "Please create it first:"
  echo "  git tag $VERSION && git push origin $VERSION"
  exit 1
fi
echo "✅ Repository tag $VERSION exists"
echo ""

# Check if golang module tag already exists
GOLANG_TAG="golang/$VERSION"
if git rev-parse "$GOLANG_TAG" >/dev/null 2>&1; then
  echo "⚠️  Warning: Go module tag $GOLANG_TAG already exists"
  read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    git tag -d "$GOLANG_TAG"
    git push origin ":refs/tags/$GOLANG_TAG" 2>/dev/null || true
    echo "Deleted existing tag $GOLANG_TAG"
  else
    echo "Aborted"
    exit 0
  fi
fi

# Check if release exists with binaries
echo "Checking if GitHub Release $VERSION has binaries..."
if ! command -v gh &> /dev/null; then
  echo "⚠️  Warning: 'gh' CLI not found. Cannot verify binaries."
  echo "Please manually verify that all platform binaries are available:"
  echo "  https://github.com/ideamans/libnextimage/releases/tag/$VERSION"
  echo ""
  read -p "Have you verified all binaries are available? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Please wait for binaries to be built and try again."
    exit 1
  fi
else
  # Check release assets
  ASSETS=$(gh release view "$VERSION" --json assets --jq '.assets[].name' 2>/dev/null || echo "")

  if [ -z "$ASSETS" ]; then
    echo "❌ Error: Release $VERSION not found or has no assets"
    echo "Please wait for GitHub Actions to complete building binaries."
    exit 1
  fi

  echo "Found assets:"
  echo "$ASSETS" | sed 's/^/  - /'
  echo ""

  # Check for required platforms
  REQUIRED_PLATFORMS=("darwin-arm64" "darwin-amd64" "linux-amd64" "windows-amd64")
  MISSING_PLATFORMS=()

  for platform in "${REQUIRED_PLATFORMS[@]}"; do
    if ! echo "$ASSETS" | grep -q "$platform"; then
      MISSING_PLATFORMS+=("$platform")
    fi
  done

  if [ ${#MISSING_PLATFORMS[@]} -gt 0 ]; then
    echo "❌ Error: Missing binaries for platforms:"
    printf '  - %s\n' "${MISSING_PLATFORMS[@]}"
    echo ""
    echo "Please wait for GitHub Actions to complete."
    echo "You can watch the progress with:"
    echo "  gh run list --limit 5"
    exit 1
  fi

  echo "✅ All required platform binaries are available"
fi

echo ""
echo "=== Ready to create Go module tag ==="
echo "This will create and push: $GOLANG_TAG"
echo ""
read -p "Continue? (y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Aborted"
  exit 0
fi

# Create and push golang module tag
echo ""
echo "Creating Go module tag..."
git tag "$GOLANG_TAG"
git push origin "$GOLANG_TAG"

echo ""
echo "✅ Success!"
echo ""
echo "Go module $GOLANG_TAG has been released!"
echo "Users can now install it with:"
echo "  go get github.com/ideamans/libnextimage/golang@$VERSION"
echo ""
echo "Verify the release:"
echo "  https://github.com/ideamans/libnextimage/releases/tag/$GOLANG_TAG"
