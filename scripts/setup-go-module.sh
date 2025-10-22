#!/bin/bash

# Setup script for libnextimage Go module
# This script downloads and installs the pre-built library to the Go module cache
#
# Usage:
#   bash <(curl -fsSL https://raw.githubusercontent.com/ideamans/libnextimage/main/scripts/setup-go-module.sh)
#
# Or locally:
#   bash scripts/setup-go-module.sh

set -e

echo "=== libnextimage Go Module Setup ==="
echo ""

# Detect platform
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin*)
    PLATFORM="darwin"
    ;;
  Linux*)
    PLATFORM="linux"
    ;;
  MINGW*|MSYS*|CYGWIN*)
    PLATFORM="windows"
    ;;
  *)
    echo "Error: Unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

PLATFORM_ARCH="${PLATFORM}-${ARCH}"
echo "Platform: $PLATFORM_ARCH"
echo ""

# Get version (default to latest or specified version)
VERSION="${1:-v0.4.0}"
if [[ "$VERSION" != v* ]]; then
  VERSION="v$VERSION"
fi

echo "Version: $VERSION"
echo ""

# Find Go module cache
GOMODCACHE=$(go env GOMODCACHE)
if [ -z "$GOMODCACHE" ]; then
  echo "Error: Could not determine GOMODCACHE"
  echo "Please run: go env GOMODCACHE"
  exit 1
fi

echo "Go module cache: $GOMODCACHE"
echo ""

# Determine module path (handle both v0.3.0 and golang/v0.3.0 tags)
MODULE_PATH="$GOMODCACHE/github.com/ideamans/libnextimage@$VERSION"
if [ ! -d "$MODULE_PATH" ]; then
  echo "Module not found in cache. Downloading..."
  go get github.com/ideamans/libnextimage/golang@$VERSION
  # Update module path after download
  MODULE_PATH=$(find "$GOMODCACHE/github.com/ideamans" -name "libnextimage@$VERSION" -o -name "libnextimage*" | head -1)
fi

if [ ! -d "$MODULE_PATH" ]; then
  echo "Error: Module still not found after go get"
  echo "Try: go mod tidy in your project directory first"
  exit 1
fi

echo "Module path: $MODULE_PATH"
echo ""

# Check if library already exists
# v0.4.0+: lib/libnextimage.a (new structure)
# v0.3.0 and earlier: lib/$PLATFORM_ARCH/libnextimage.a (old structure)
LIB_PATH="$MODULE_PATH/lib/libnextimage.a"
OLD_LIB_PATH="$MODULE_PATH/lib/$PLATFORM_ARCH/libnextimage.a"

if [ -f "$LIB_PATH" ] || [ -f "$OLD_LIB_PATH" ]; then
  INSTALLED_PATH="${LIB_PATH}"
  if [ ! -f "$LIB_PATH" ]; then
    INSTALLED_PATH="${OLD_LIB_PATH}"
  fi
  echo "✅ Library already installed at: $INSTALLED_PATH"
  read -p "Do you want to reinstall? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup complete."
    exit 0
  fi
fi

# Download release archive
ARCHIVE_NAME="libnextimage-${VERSION}-${PLATFORM_ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/ideamans/libnextimage/releases/download/${VERSION}/${ARCHIVE_NAME}"

echo "Downloading: $ARCHIVE_NAME"
echo "From: $DOWNLOAD_URL"
echo ""

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

if command -v curl > /dev/null; then
  curl -L -o "$TEMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"
elif command -v wget > /dev/null; then
  wget -O "$TEMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"
else
  echo "Error: Neither curl nor wget is available"
  exit 1
fi

# Extract to module directory
echo "Installing to: $MODULE_PATH"
echo ""

# Need to make module cache writable temporarily
if [ ! -w "$MODULE_PATH" ]; then
  echo "Module cache is read-only. Attempting to make writable..."
  chmod +w "$MODULE_PATH" || {
    echo "Error: Cannot make module cache writable"
    echo "You may need to run this script with sudo:"
    echo "  sudo bash $0"
    exit 1
  }
fi

# Clean up old directory structure if it exists
# v0.3.0 and earlier had: lib/$PLATFORM_ARCH/, include/
# v0.4.0+ has: lib/libnextimage.a, include/
if [ -d "$MODULE_PATH/lib/$PLATFORM_ARCH" ]; then
  echo "Removing old platform-specific directory structure..."
  chmod -R +w "$MODULE_PATH/lib" 2>/dev/null || true
  rm -rf "$MODULE_PATH/lib/$PLATFORM_ARCH"
fi

# Make existing directories writable for overwrite
if [ -d "$MODULE_PATH/lib" ]; then
  chmod -R +w "$MODULE_PATH/lib" 2>/dev/null || true
fi
if [ -d "$MODULE_PATH/include" ]; then
  chmod -R +w "$MODULE_PATH/include" 2>/dev/null || true
fi

cd "$MODULE_PATH"
tar xzf "$TEMP_DIR/$ARCHIVE_NAME"

echo ""
echo "✅ Setup complete!"
echo ""
echo "Library installed to: $LIB_PATH"
echo ""
echo "You can now use libnextimage in your Go project:"
echo "  import \"github.com/ideamans/libnextimage/golang\""
echo ""
