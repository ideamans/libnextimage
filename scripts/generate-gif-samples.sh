#!/bin/bash
# Generate GIF test samples from PNG files

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PNG_DIR="$PROJECT_ROOT/testdata/png-source"
GIF_DIR="$PROJECT_ROOT/testdata/gif-source"

echo "Generating GIF test samples..."

# ImageMagick's convert command is needed
if ! command -v convert &> /dev/null; then
    echo "Error: ImageMagick 'convert' command not found"
    echo "Install with: brew install imagemagick"
    exit 1
fi

mkdir -p "$GIF_DIR"

# Static GIF samples (from PNG)
echo "Creating static GIFs..."

# Small size
convert "$PNG_DIR/small-64x64.png" "$GIF_DIR/static-64x64.gif"
echo "  static-64x64.gif"

# Medium size
convert "$PNG_DIR/medium-512x512.png" "$GIF_DIR/static-512x512.gif"
echo "  static-512x512.gif"

# With transparency
convert "$PNG_DIR/alpha-gradient.png" "$GIF_DIR/static-alpha.gif"
echo "  static-alpha.gif"

# Different content types
convert "$PNG_DIR/solid-red.png" "$GIF_DIR/solid-color.gif"
echo "  solid-color.gif"

convert "$PNG_DIR/gradient-horizontal.png" "$GIF_DIR/gradient.gif"
echo "  gradient.gif"

# Simple animated GIF (3 frames from different PNGs)
echo "Creating simple animated GIF..."
convert -delay 100 \
    "$PNG_DIR/solid-red.png" \
    "$PNG_DIR/solid-green.png" \
    "$PNG_DIR/solid-blue.png" \
    -loop 0 \
    "$GIF_DIR/animated-3frames.gif"
echo "  animated-3frames.gif"

# Animated GIF with alpha
echo "Creating animated GIF with transparency..."
convert -delay 100 \
    "$PNG_DIR/alpha-transparent.png" \
    "$PNG_DIR/alpha-opaque.png" \
    -loop 0 \
    "$GIF_DIR/animated-alpha.gif"
echo "  animated-alpha.gif"

# Small animated GIF
echo "Creating small animated GIF..."
convert -delay 100 \
    "$PNG_DIR/tiny-16x16.png" \
    "$PNG_DIR/small-64x64.png" \
    -resize 16x16 \
    -loop 0 \
    "$GIF_DIR/animated-small.gif"
echo "  animated-small.gif"

echo ""
echo "Generated GIF samples:"
ls -lh "$GIF_DIR"/*.gif | awk '{print "  " $9 " (" $5 ")"}'
echo ""
echo "Total: $(ls -1 "$GIF_DIR"/*.gif | wc -l) files"
