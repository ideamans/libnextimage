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
echo "=== Building nextimage libraries (static and shared) ==="
# CPUコア数を取得（クロスプラットフォーム対応）
NCPUS=$(sysctl -n hw.ncpu 2>/dev/null || nproc 2>/dev/null || echo 4)
cmake --build . --target nextimage --parallel $NCPUS
cmake --build . --target nextimage_shared --parallel $NCPUS

echo ""
echo "=== Build complete ==="
echo ""

# 生成されたライブラリの確認（ビルドディレクトリ内）
echo "Generated libraries in build directory:"
find . -name "*.a" -type f ! -path "./_deps/*" | while read lib; do
    echo "  $(basename $lib): $(du -h "$lib" | cut -f1)"
done

echo ""
echo "=== Installing combined library ==="
cmake --install . --prefix "$PROJECT_ROOT"

echo ""
# 統合されたライブラリの確認
if [ -d "$PROJECT_ROOT/lib" ]; then
    echo "Library structure:"
    echo ""

    if [ -f "$PROJECT_ROOT/lib/static/libnextimage.a" ]; then
        echo "Static library:"
        echo "  lib/static/libnextimage.a: $(du -h "$PROJECT_ROOT/lib/static/libnextimage.a" | cut -f1)"
    fi

    echo ""
    echo "Shared libraries:"
    find "$PROJECT_ROOT/lib/shared" \( -name "*.so" -o -name "*.dylib" -o -name "*.dll" \) -type f 2>/dev/null | while read lib; do
        echo "  lib/shared/$(basename $lib): $(du -h "$lib" | cut -f1)"
    done || echo "  (none found)"

    echo ""
    echo "Header files:"
    find "$PROJECT_ROOT/lib/include" -name "*.h" -type f 2>/dev/null | while read hdr; do
        echo "  lib/include/$(basename $hdr)"
    done || echo "  (none found)"
else
    echo "Warning: lib/ directory not created. Install may have failed."
fi

echo ""
echo "Note: Static library (lib/static) for Go bindings, shared library (lib/shared) for TypeScript/FFI"
