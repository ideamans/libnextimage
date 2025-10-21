#!/bin/bash

# libnextimage インストールスクリプト
# GitHubリリースからビルド済みライブラリをダウンロードして展開します
#
# 使い方:
#   bash scripts/install.sh [version]
#   version省略時は最新のタグバージョンを使用します

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# プラットフォームとアーキテクチャを検出
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

echo "=== libnextimage Installation Script ==="
echo "Platform: $PLATFORM_ARCH"
echo "Project root: $PROJECT_ROOT"
echo ""

# バージョンを取得（引数がなければversion.goから取得）
VERSION="$1"
if [ -z "$VERSION" ]; then
  # version.goからバージョンを取得
  VERSION=$(grep 'LibraryVersion = ' "$PROJECT_ROOT/golang/version.go" | sed 's/.*"\(.*\)"/\1/')
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine version. Please specify version as argument."
    exit 1
  fi
  echo "Using version from version.go: $VERSION"
fi

# バージョン番号の正規化（vプレフィックスの処理）
if [[ "$VERSION" != v* ]]; then
  VERSION="v${VERSION}"
fi

# GitHub情報
GITHUB_USER="ideamans"
GITHUB_REPO="libnextimage"
ARCHIVE_NAME="libnextimage-${VERSION}-${PLATFORM_ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

echo "Download URL: $DOWNLOAD_URL"
echo ""

# すでにライブラリが存在するかチェック
LIB_PATH="$PROJECT_ROOT/lib/${PLATFORM_ARCH}/libnextimage.a"
if [ -f "$LIB_PATH" ]; then
  echo "Library already exists at: $LIB_PATH"
  read -p "Do you want to overwrite? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Installation cancelled."
    exit 0
  fi
fi

# 一時ディレクトリの作成
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Downloading $ARCHIVE_NAME..."
if command -v curl > /dev/null; then
  curl -L -o "$TEMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"
elif command -v wget > /dev/null; then
  wget -O "$TEMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"
else
  echo "Error: Neither curl nor wget is available. Please install either."
  exit 1
fi

if [ ! -f "$TEMP_DIR/$ARCHIVE_NAME" ]; then
  echo "Error: Download failed"
  exit 1
fi

echo "Download completed: $(du -h "$TEMP_DIR/$ARCHIVE_NAME" | cut -f1)"
echo ""

# プロジェクトルートに展開
echo "Extracting archive to $PROJECT_ROOT..."
cd "$PROJECT_ROOT"
tar xzf "$TEMP_DIR/$ARCHIVE_NAME"

# 確認
echo ""
echo "=== Installation complete ==="
if [ -f "$LIB_PATH" ]; then
  echo "Library installed: $LIB_PATH ($(du -h "$LIB_PATH" | cut -f1))"
else
  echo "Warning: Library file not found at expected location: $LIB_PATH"
  exit 1
fi

if [ -d "$PROJECT_ROOT/include" ]; then
  HEADER_COUNT=$(find "$PROJECT_ROOT/include" -name "*.h" | wc -l)
  echo "Header files: $HEADER_COUNT files in include/"
else
  echo "Warning: include/ directory not found"
fi

echo ""
echo "You can now use the library with Go:"
echo "  import \"github.com/ideamans/libnextimage/golang\""
