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

# ヘッダファイルのコピー
echo ""
echo "=== Installing header files ==="
mkdir -p "$PROJECT_ROOT/include/nextimage"
cp "$PROJECT_ROOT/c/include"/*.h "$PROJECT_ROOT/include/"
cp "$PROJECT_ROOT/c/include/nextimage"/*.h "$PROJECT_ROOT/include/nextimage/"
echo "Header files installed to include/"

echo ""
# 統合されたライブラリの確認
if [ -d "$PROJECT_ROOT/lib" ]; then
    echo "Combined libraries in lib/ directory:"
    find "$PROJECT_ROOT/lib" -name "*.a" -type f | while read lib; do
        platform=$(basename $(dirname "$lib"))
        echo "  $platform/$(basename $lib): $(du -h "$lib" | cut -f1)"
    done

    echo ""
    echo "Shared libraries in lib/ directory:"
    find "$PROJECT_ROOT/lib" \( -name "*.so" -o -name "*.dylib" -o -name "*.dll" \) -type f | while read lib; do
        platform=$(basename $(dirname "$lib"))
        echo "  $platform/$(basename $lib): $(du -h "$lib" | cut -f1)"
    done
else
    echo "Warning: lib/ directory not created. Install may have failed."
fi

echo ""
echo "Note: Static library (*.a) for Go bindings, shared library (*.so/*.dylib/*.dll) for Node.js/FFI"
