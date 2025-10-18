#!/bin/bash

# C言語ライブラリのビルドスクリプト
# libnextimage.a とその依存ライブラリをビルドします

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/c/build"

echo "=== libnextimage C Library Build ==="
echo "Project root: $PROJECT_ROOT"
echo "Build directory: $BUILD_DIR"
echo ""

# ビルドディレクトリを作成
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# CMake設定
echo "=== Running CMake configuration ==="
cmake .. -DCMAKE_BUILD_TYPE=Release

# ビルド
echo ""
echo "=== Building nextimage library ==="
make nextimage -j$(sysctl -n hw.ncpu 2>/dev/null || nproc 2>/dev/null || echo 4)

echo ""
echo "=== Build complete ==="
echo ""

# 生成されたライブラリの確認
echo "Generated libraries:"
find . -name "*.a" -type f ! -path "./_deps/*" | while read lib; do
    echo "  $(basename $lib): $(du -h "$lib" | cut -f1)"
done

echo ""
echo "Note: For Go bindings, the libraries are linked directly from c/build/"
echo "No need to copy to lib/ directory when using direct linking in common.go"
