#!/bin/bash

# テストデータ生成スクリプト
# ImageMagick (magick) が必要

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TESTDATA_DIR="$PROJECT_ROOT/testdata"

echo "=== テストデータ生成開始 ==="
echo "出力先: $TESTDATA_DIR"

# ImageMagickの確認
if ! command -v magick &> /dev/null; then
    echo "ERROR: ImageMagick (magick) が見つかりません"
    echo "インストール: brew install imagemagick"
    exit 1
fi

# ディレクトリ構造作成
mkdir -p "$TESTDATA_DIR/source"/{sizes,colors,alpha,compression,avif-specific,gif}

cd "$TESTDATA_DIR/source"

echo ""
echo "=== 1. サイズバリエーション ==="

# 極小 (16x16)
magick -size 16x16 gradient:blue-red sizes/tiny-16x16.png
echo "  ✓ tiny-16x16.png"

# 小 (128x128) - 8-bit
magick -size 128x128 -depth 8 gradient:green-yellow sizes/small-128x128.png
echo "  ✓ small-128x128.png"

# 中 (512x512) - 8-bit for better compatibility
magick -size 512x512 -depth 8 gradient:red-blue sizes/medium-512x512.png
echo "  ✓ medium-512x512.png"

# 大 (2048x2048) - largest test image
magick -size 2048x2048 plasma:fractal sizes/large-2048x2048.png
echo "  ✓ large-2048x2048.png"

# 非正方形
magick -size 800x600 gradient:blue-red sizes/rect-800x600.png
echo "  ✓ rect-800x600.png"

magick -size 1920x1080 plasma: sizes/hd-1920x1080.png
echo "  ✓ hd-1920x1080.png"

echo ""
echo "=== 2. 色パターン ==="

# 単色
magick -size 256x256 xc:red colors/solid-red.png
magick -size 256x256 xc:green colors/solid-green.png
magick -size 256x256 xc:blue colors/solid-blue.png
magick -size 256x256 xc:white colors/solid-white.png
magick -size 256x256 xc:black colors/solid-black.png
echo "  ✓ 単色画像 (5種類)"

# グラデーション
magick -size 512x512 gradient:red-blue colors/gradient-horizontal.png
magick -size 512x512 gradient:red-blue -rotate 90 colors/gradient-vertical.png
magick -size 512x512 radial-gradient:red-blue colors/gradient-radial.png
echo "  ✓ グラデーション (3種類)"

# チェッカーボード
magick -size 512x512 pattern:checkerboard colors/checkerboard.png
echo "  ✓ checkerboard.png"

# カラーパレット (256色)
magick -size 512x512 gradient: -colors 256 colors/palette-256.png
echo "  ✓ palette-256.png"

# 写真風 (プラズマ)
magick -size 512x512 plasma:fractal colors/photo-like.png
echo "  ✓ photo-like.png"

echo ""
echo "=== 3. 透明度パターン ==="

# 完全不透明
magick -size 256x256 xc:red alpha/opaque.png
echo "  ✓ opaque.png"

# 完全透明
magick -size 256x256 xc:none alpha/transparent.png
echo "  ✓ transparent.png"

# 半透明グラデーション
magick -size 512x512 gradient:red-blue \
    \( -size 512x512 gradient:white-black \) \
    -compose CopyOpacity -composite \
    alpha/alpha-gradient.png
echo "  ✓ alpha-gradient.png"

# アルファチャンネル付き複雑な画像
magick -size 512x512 plasma: \
    \( -size 512x512 gradient:white-black -evaluate multiply 0.8 \) \
    -compose CopyOpacity -composite \
    alpha/alpha-complex.png
echo "  ✓ alpha-complex.png"

# 透過PNGサンプル (円形)
magick -size 256x256 xc:none \
    -fill red -draw "circle 128,128 200,128" \
    alpha/alpha-circle.png
echo "  ✓ alpha-circle.png"

echo ""
echo "=== 4. 圧縮特性 ==="

# 高圧縮率向け (フラットカラー)
magick -size 512x512 xc:"#3498db" compression/flat-color.png
echo "  ✓ flat-color.png"

# 低圧縮率向け (ノイズ)
magick -size 512x512 xc:white +noise Gaussian compression/noisy.png
echo "  ✓ noisy.png"

# エッジが多い画像
magick -size 512x512 pattern:checkerboard -scale 50% compression/edges.png
echo "  ✓ edges.png"

# テキスト画像
magick -size 512x512 xc:white \
    -pointsize 48 -fill black \
    -annotate +50+100 "Test Text\nCompression\nSample" \
    compression/text.png
echo "  ✓ text.png"

# ディザリング
magick -size 512x512 gradient: -ordered-dither o8x8 compression/dithered.png
echo "  ✓ dithered.png"

echo ""
echo "=== 5. AVIF専用 (ビット深度/色空間) ==="

# 8bit RGB
magick -size 512x512 gradient:red-blue -depth 8 avif-specific/8bit-rgb.png
echo "  ✓ 8bit-rgb.png"

# 10bit RGB (PNGは8bitなので、後でAVIF変換時に10bitにする)
magick -size 512x512 plasma: -depth 16 avif-specific/10bit-source.png
echo "  ✓ 10bit-source.png"

# 12bit RGB (同上)
magick -size 512x512 gradient: -depth 16 avif-specific/12bit-source.png
echo "  ✓ 12bit-source.png"

echo ""
echo "=== 6. GIF専用 (アニメーション) ==="

# 2フレーム (最小)
magick -size 64x64 xc:red -delay 50 \
       -size 64x64 xc:blue -delay 50 \
       -loop 0 gif/anim-2frames.gif
echo "  ✓ anim-2frames.gif"

# 10フレーム (短)
magick -delay 10 \
    \( -size 64x64 xc:red \) \
    \( -size 64x64 xc:orange \) \
    \( -size 64x64 xc:yellow \) \
    \( -size 64x64 xc:green \) \
    \( -size 64x64 xc:cyan \) \
    \( -size 64x64 xc:blue \) \
    \( -size 64x64 xc:purple \) \
    \( -size 64x64 xc:pink \) \
    \( -size 64x64 xc:white \) \
    \( -size 64x64 xc:black \) \
    -loop 0 gif/anim-10frames.gif
echo "  ✓ anim-10frames.gif"

# 静止画GIF (2色)
magick -size 64x64 pattern:checkerboard -colors 2 gif/static-2colors.gif
echo "  ✓ static-2colors.gif"

# 静止画GIF (16色)
magick -size 64x64 gradient: -colors 16 gif/static-16colors.gif
echo "  ✓ static-16colors.gif"

# 静止画GIF (256色)
magick -size 64x64 gradient: -colors 256 gif/static-256colors.gif
echo "  ✓ static-256colors.gif"

echo ""
echo "=== 7. 追加: 実写風サンプル (オプションテスト用) ==="

# 風景風 (グラデーション + ノイズ)
magick -size 1024x768 gradient:skyblue-green \
    +noise Gaussian -blur 0x2 \
    colors/landscape-like.png
echo "  ✓ landscape-like.png"

# 人物風 (楕円形状)
magick -size 512x512 xc:beige \
    -fill tan -draw "ellipse 256,200 150,180 0,360" \
    -fill brown -draw "ellipse 256,180 60,80 0,360" \
    colors/portrait-like.png
echo "  ✓ portrait-like.png"

# テクスチャ
magick -size 512x512 plasma:fractal -blur 0x1 colors/texture.png
echo "  ✓ texture.png"

echo ""
echo "=== 8. JPEG/PNG変換 ==="

# 主要な画像をJPEGにも変換 (WebPエンコードテスト用)
mkdir -p "$TESTDATA_DIR/jpeg-source"
mkdir -p "$TESTDATA_DIR/png-source"

for file in sizes/*.png colors/*.png alpha/opaque.png compression/*.png; do
    if [ -f "$file" ]; then
        basename=$(basename "$file" .png)
        # JPEG (透明度なし画像のみ)
        if [[ "$file" != *"alpha"* ]] || [[ "$file" == *"opaque"* ]]; then
            magick "$file" -quality 95 "$TESTDATA_DIR/jpeg-source/${basename}.jpg"
        fi
        # PNG
        cp "$file" "$TESTDATA_DIR/png-source/${basename}.png"
    fi
done

echo "  ✓ JPEG/PNG変換完了"

echo ""
echo "=== 9. WebP テストデータ (dwebp テスト用) ==="

# cwebpコマンドの確認
CWEBP="$PROJECT_ROOT/bin/cwebp"
if [ ! -f "$CWEBP" ]; then
    echo "  ⚠ cwebp not found at $CWEBP"
    echo "  Run scripts/build-cli-tools.sh first to generate WebP test data"
else
    mkdir -p "$TESTDATA_DIR/webp-samples"

    # ロッシー品質バリエーション
    "$CWEBP" -q 75 "$TESTDATA_DIR/source/sizes/medium-512x512.png" -o "$TESTDATA_DIR/webp-samples/lossy-q75.webp" 2>/dev/null
    echo "  ✓ lossy-q75.webp"

    "$CWEBP" -q 90 "$TESTDATA_DIR/source/sizes/medium-512x512.png" -o "$TESTDATA_DIR/webp-samples/lossy-q90.webp" 2>/dev/null
    echo "  ✓ lossy-q90.webp"

    # ロスレス
    "$CWEBP" -lossless "$TESTDATA_DIR/source/sizes/medium-512x512.png" -o "$TESTDATA_DIR/webp-samples/lossless.webp" 2>/dev/null
    echo "  ✓ lossless.webp"

    # アルファチャンネル付き
    "$CWEBP" -q 75 "$TESTDATA_DIR/source/alpha/alpha-gradient.png" -o "$TESTDATA_DIR/webp-samples/alpha-gradient.webp" 2>/dev/null
    echo "  ✓ alpha-gradient.webp"

    "$CWEBP" -lossless "$TESTDATA_DIR/source/alpha/alpha-gradient.png" -o "$TESTDATA_DIR/webp-samples/alpha-lossless.webp" 2>/dev/null
    echo "  ✓ alpha-lossless.webp"

    # サイズバリエーション
    "$CWEBP" -q 75 "$TESTDATA_DIR/source/sizes/small-128x128.png" -o "$TESTDATA_DIR/webp-samples/small-128x128.webp" 2>/dev/null
    echo "  ✓ small-128x128.webp"

    "$CWEBP" -q 75 "$TESTDATA_DIR/source/sizes/large-2048x2048.png" -o "$TESTDATA_DIR/webp-samples/large-2048x2048.webp" 2>/dev/null
    echo "  ✓ large-2048x2048.webp"
fi

echo ""
echo "=== テストデータ生成完了 ==="
echo ""
echo "生成された画像:"
find "$TESTDATA_DIR/source" -type f | wc -l | xargs echo "  source: "
find "$TESTDATA_DIR/jpeg-source" -type f 2>/dev/null | wc -l | xargs echo "  jpeg-source: "
find "$TESTDATA_DIR/png-source" -type f 2>/dev/null | wc -l | xargs echo "  png-source: "
find "$TESTDATA_DIR/webp-samples" -type f 2>/dev/null | wc -l | xargs echo "  webp-samples: "

echo ""
echo "次のステップ:"
echo "  1. コマンドラインツールのビルド (scripts/build-cli-tools.sh)"
echo "  2. WebPテストデータ生成 (scripts/generate-testdata.sh を再実行)"
echo "  3. 互換性テストの実行"
