#!/bin/bash

# WebP互換性テストスクリプト
# cwebpコマンドとlibnextiamge WebPライブラリの出力を比較

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$PROJECT_ROOT/bin"
TESTDATA_DIR="$PROJECT_ROOT/testdata"
GOLANG_DIR="$PROJECT_ROOT/golang"
REPORT_DIR="$PROJECT_ROOT/test-reports"

# テスト用一時ディレクトリ
TEMP_DIR="$PROJECT_ROOT/build-test-compat"

echo "=== WebP互換性テスト開始 ==="
echo "コマンドツールディレクトリ: $BIN_DIR"
echo "テストデータ: $TESTDATA_DIR"
echo "一時ディレクトリ: $TEMP_DIR"
echo ""

# 必要なツールの確認
if [ ! -f "$BIN_DIR/cwebp" ]; then
    echo "ERROR: cwebp が見つかりません"
    echo "先に scripts/build-cli-tools.sh を実行してください"
    exit 1
fi

if [ ! -f "$BIN_DIR/dwebp" ]; then
    echo "ERROR: dwebp が見つかりません"
    echo "先に scripts/build-cli-tools.sh を実行してください"
    exit 1
fi

# ディレクトリ作成
mkdir -p "$TEMP_DIR/cmd-output"
mkdir -p "$TEMP_DIR/lib-output"
mkdir -p "$REPORT_DIR"

# レポートファイル
REPORT_FILE="$REPORT_DIR/webp-compat-$(date +%Y%m%d-%H%M%S).txt"

echo "=== テストレポート: WebP互換性検証 ===" > "$REPORT_FILE"
echo "日時: $(date)" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# カウンタ
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# テスト関数: cwebpコマンドとライブラリ出力を比較
# Usage: test_webp_encode <test_name> <input_file> <cwebp_options> <go_quality> <go_lossless>
test_webp_encode() {
    local test_name="$1"
    local input_file="$2"
    local cwebp_opts="$3"
    local go_quality="$4"
    local go_lossless="${5:-0}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo ""
    echo "--- Test: $test_name ---"
    echo "入力: $input_file"
    echo "cwebp options: $cwebp_opts"

    local cmd_output="$TEMP_DIR/cmd-output/${test_name}.webp"
    local lib_output="$TEMP_DIR/lib-output/${test_name}.webp"

    # コマンド版の実行
    if ! "$BIN_DIR/cwebp" $cwebp_opts "$input_file" -o "$cmd_output" &>/dev/null; then
        echo "  ❌ FAILED: cwebp コマンド実行エラー"
        echo "Test: $test_name - FAILED (cwebp error)" >> "$REPORT_FILE"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi

    local cmd_size=$(stat -f%z "$cmd_output" 2>/dev/null || stat -c%s "$cmd_output" 2>/dev/null)
    echo "  cwebp output: $cmd_size bytes"

    # ライブラリ版の実行
    cd "$GOLANG_DIR"

    # Goテストプログラムで同等のエンコードを実行
    cat > /tmp/test_compat_encode.go <<'GOEOF'
package main

import (
    "fmt"
    "os"
    libnextimage "github.com/ideamans/libnextimage/golang"
)

func main() {
    opts := libnextimage.DefaultWebPEncodeOptions()
    opts.Quality = QUALITY_PLACEHOLDER
    if LOSSLESS_PLACEHOLDER != 0 {
        opts.Lossless = true
    }

    data, err := libnextimage.WebPEncodeFile("INPUT_FILE_PLACEHOLDER", opts)
    if err != nil {
        fmt.Fprintf(os.Stderr, "encode error: %v\n", err)
        os.Exit(1)
    }

    if err := os.WriteFile("OUTPUT_FILE_PLACEHOLDER", data, 0644); err != nil {
        fmt.Fprintf(os.Stderr, "write error: %v\n", err)
        os.Exit(1)
    }
}
GOEOF

    # Replace placeholders
    sed -i.bak "s|QUALITY_PLACEHOLDER|$go_quality|g" /tmp/test_compat_encode.go
    sed -i.bak "s|LOSSLESS_PLACEHOLDER|$go_lossless|g" /tmp/test_compat_encode.go
    sed -i.bak "s|INPUT_FILE_PLACEHOLDER|$input_file|g" /tmp/test_compat_encode.go
    sed -i.bak "s|OUTPUT_FILE_PLACEHOLDER|$lib_output|g" /tmp/test_compat_encode.go
    rm /tmp/test_compat_encode.go.bak

    if ! go run /tmp/test_compat_encode.go 2>/dev/null; then
        echo "  ❌ FAILED: ライブラリ実行エラー"
        echo "Test: $test_name - FAILED (library error)" >> "$REPORT_FILE"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        cd "$PROJECT_ROOT"
        return 1
    fi

    cd "$PROJECT_ROOT"

    local lib_size=$(stat -f%z "$lib_output" 2>/dev/null || stat -c%s "$lib_output" 2>/dev/null)
    echo "  library output: $lib_size bytes"

    # ファイルサイズ比較 (±2%の許容範囲)
    local size_diff=$((lib_size - cmd_size))
    local size_diff_abs=${size_diff#-}
    local size_diff_percent=$(awk "BEGIN {printf \"%.2f\", ($size_diff_abs * 100.0) / $cmd_size}")

    echo "  size difference: ${size_diff} bytes (${size_diff_percent}%)"

    # バイナリ比較
    if cmp -s "$cmd_output" "$lib_output"; then
        echo "  ✓ PASSED: バイナリ完全一致"
        echo "Test: $test_name - PASSED (binary match)" >> "$REPORT_FILE"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    fi

    # バイナリ不一致の場合、ファイルサイズで判定
    if (( $(echo "$size_diff_percent < 2.0" | bc -l) )); then
        echo "  ⚠ PASSED (size match): バイナリ不一致だがファイルサイズ許容範囲内"
        echo "Test: $test_name - PASSED (size match, ${size_diff_percent}% diff)" >> "$REPORT_FILE"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo "  ❌ FAILED: ファイルサイズが許容範囲外 (${size_diff_percent}%)"
        echo "Test: $test_name - FAILED (size mismatch, ${size_diff_percent}% diff)" >> "$REPORT_FILE"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

echo "=== 1. Quality オプションテスト ==="

# 代表的な画像でqualityをテスト
test_webp_encode "quality-0-medium" "$TESTDATA_DIR/source/sizes/medium-512x512.png" "-q 0" 0
test_webp_encode "quality-25-medium" "$TESTDATA_DIR/source/sizes/medium-512x512.png" "-q 25" 25
test_webp_encode "quality-50-medium" "$TESTDATA_DIR/source/sizes/medium-512x512.png" "-q 50" 50
test_webp_encode "quality-75-medium" "$TESTDATA_DIR/source/sizes/medium-512x512.png" "-q 75" 75
test_webp_encode "quality-90-medium" "$TESTDATA_DIR/source/sizes/medium-512x512.png" "-q 90" 90
test_webp_encode "quality-100-medium" "$TESTDATA_DIR/source/sizes/medium-512x512.png" "-q 100" 100

echo ""
echo "=== 2. Lossless モードテスト ==="

test_webp_encode "lossless-medium" "$TESTDATA_DIR/source/sizes/medium-512x512.png" "-lossless" 75 1
test_webp_encode "lossless-small" "$TESTDATA_DIR/source/sizes/small-128x128.png" "-lossless" 75 1
test_webp_encode "lossless-alpha" "$TESTDATA_DIR/source/alpha/alpha-gradient.png" "-lossless" 75 1

echo ""
echo "=== 3. サイズバリエーションテスト ==="

# 異なるサイズで同じquality=80をテスト
test_webp_encode "size-tiny-q80" "$TESTDATA_DIR/source/sizes/tiny-16x16.png" "-q 80" 80
test_webp_encode "size-small-q80" "$TESTDATA_DIR/source/sizes/small-128x128.png" "-q 80" 80
test_webp_encode "size-medium-q80" "$TESTDATA_DIR/source/sizes/medium-512x512.png" "-q 80" 80
test_webp_encode "size-large-q80" "$TESTDATA_DIR/source/sizes/large-2048x2048.png" "-q 80" 80

echo ""
echo "=== 4. 透明度テスト ==="

test_webp_encode "alpha-opaque-q75" "$TESTDATA_DIR/source/alpha/opaque.png" "-q 75" 75
test_webp_encode "alpha-transparent-q75" "$TESTDATA_DIR/source/alpha/transparent.png" "-q 75" 75
test_webp_encode "alpha-gradient-q75" "$TESTDATA_DIR/source/alpha/alpha-gradient.png" "-q 75" 75
test_webp_encode "alpha-complex-q75" "$TESTDATA_DIR/source/alpha/alpha-complex.png" "-q 75" 75

echo ""
echo "=== 5. 圧縮特性テスト ==="

# 圧縮しやすい画像
test_webp_encode "compress-flat-q75" "$TESTDATA_DIR/source/compression/flat-color.png" "-q 75" 75

# 圧縮しにくい画像
test_webp_encode "compress-noisy-q75" "$TESTDATA_DIR/source/compression/noisy.png" "-q 75" 75

# エッジが多い画像
test_webp_encode "compress-edges-q75" "$TESTDATA_DIR/source/compression/edges.png" "-q 75" 75

# テキスト画像
test_webp_encode "compress-text-q75" "$TESTDATA_DIR/source/compression/text.png" "-q 75" 75

echo ""
echo "=== テスト結果サマリー ==="
echo ""
echo "Total tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"

if [ $FAILED_TESTS -eq 0 ]; then
    echo ""
    echo "✓ すべてのテストが成功しました！"
    PASS_RATE=100
else
    PASS_RATE=$(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS * 100.0) / $TOTAL_TESTS}")
    echo ""
    echo "⚠ 一部のテストが失敗しました"
fi

echo ""
echo "成功率: ${PASS_RATE}%"

# レポートにサマリー追加
echo "" >> "$REPORT_FILE"
echo "=== サマリー ===" >> "$REPORT_FILE"
echo "Total: $TOTAL_TESTS" >> "$REPORT_FILE"
echo "Passed: $PASSED_TESTS" >> "$REPORT_FILE"
echo "Failed: $FAILED_TESTS" >> "$REPORT_FILE"
echo "Pass rate: ${PASS_RATE}%" >> "$REPORT_FILE"

echo ""
echo "詳細レポート: $REPORT_FILE"

# 成功率が95%未満の場合はエラー終了
if (( $(echo "$PASS_RATE < 95.0" | bc -l) )); then
    echo ""
    echo "ERROR: 成功率が95%未満です (${PASS_RATE}%)"
    exit 1
fi

echo ""
echo "✓ WebP互換性テスト完了"
