# nextimage

High-performance WebP and AVIF image encoding/decoding library for Rust.

[![Crates.io](https://img.shields.io/crates/v/nextimage.svg)](https://crates.io/crates/nextimage)
[![Documentation](https://docs.rs/nextimage/badge.svg)](https://docs.rs/nextimage)
[![License](https://img.shields.io/crates/l/nextimage.svg)](LICENSE)

## Features

- **WebP encoding/decoding** with full libwebp options support
- **AVIF encoding/decoding** with AOM codec
- **GIF to WebP conversion** for animated images
- **WebP to GIF conversion**
- **Automatic library download** (optional feature)
- **Cross-platform support**: macOS (Intel/ARM), Linux (x64/ARM64), Windows (x64)
- **CLI-compatible output**: produces identical output to `cwebp`, `avifenc`, etc.

## Installation

Add to your `Cargo.toml`:

```toml
[dependencies]
nextimage = "0.5"
```

With automatic library download:

```toml
[dependencies]
nextimage = { version = "0.5", features = ["auto-download"] }
```

## Quick Start

```rust
use nextimage::{webp, avif};

fn main() -> nextimage::Result<()> {
    // Ensure the native library is available
    nextimage::ensure_library()?;

    // Read input image (PNG, JPEG, etc.)
    let input = std::fs::read("input.png")?;

    // Encode to WebP
    let webp_data = webp::encode(&input)?;
    std::fs::write("output.webp", webp_data)?;

    // Encode to AVIF
    let avif_data = avif::encode(&input)?;
    std::fs::write("output.avif", avif_data)?;

    Ok(())
}
```

## Usage

### WebP Encoding

```rust
use nextimage::webp::{WebPEncoder, WebPEncodeOptions, WebPPreset};

// Lossy encoding with quality 85
let options = WebPEncodeOptions {
    quality: 85.0,
    ..Default::default()
};
let encoder = WebPEncoder::new(options)?;
let webp_data = encoder.encode(&png_data)?;

// Lossless encoding
let encoder = WebPEncoder::new(WebPEncodeOptions::lossless())?;
let webp_data = encoder.encode(&png_data)?;

// With preset
let options = WebPEncodeOptions {
    preset: WebPPreset::Photo,
    quality: 90.0,
    ..Default::default()
};
```

### WebP Decoding

```rust
use nextimage::webp::{WebPDecoder, WebPDecodeOptions, WebPOutputFormat};

// Decode to PNG (default)
let decoder = WebPDecoder::new(WebPDecodeOptions::default())?;
let png_data = decoder.decode(&webp_data)?;

// Decode to JPEG
let options = WebPDecodeOptions {
    output_format: WebPOutputFormat::Jpeg,
    jpeg_quality: 90,
    ..Default::default()
};
let decoder = WebPDecoder::new(options)?;
let jpeg_data = decoder.decode(&webp_data)?;
```

### AVIF Encoding

```rust
use nextimage::avif::{AVIFEncoder, AVIFEncodeOptions, YuvFormat};

// Default encoding (quality 60, speed 6)
let encoder = AVIFEncoder::new(AVIFEncodeOptions::default())?;
let avif_data = encoder.encode(&png_data)?;

// High quality encoding
let options = AVIFEncodeOptions {
    quality: 80,
    speed: 2,
    yuv_format: YuvFormat::Yuv444,
    ..Default::default()
};
let encoder = AVIFEncoder::new(options)?;
let avif_data = encoder.encode(&png_data)?;

// Small file size
let encoder = AVIFEncoder::new(AVIFEncodeOptions::small_size())?;
```

### AVIF Decoding

```rust
use nextimage::avif::{AVIFDecoder, AVIFDecodeOptions};

let decoder = AVIFDecoder::new(AVIFDecodeOptions::default())?;
let png_data = decoder.decode(&avif_data)?;
```

### GIF Conversion

```rust
use nextimage::webp;

// GIF to WebP
let gif_data = std::fs::read("animation.gif")?;
let webp_data = webp::gif_to_webp(&gif_data)?;

// WebP to GIF
let webp_data = std::fs::read("animation.webp")?;
let gif_data = webp::webp_to_gif(&webp_data)?;
```

### Encoder Reuse

For better performance when processing multiple images:

```rust
use nextimage::webp::{WebPEncoder, WebPEncodeOptions};

let encoder = WebPEncoder::new(WebPEncodeOptions::lossy(85.0))?;

for path in image_paths {
    let input = std::fs::read(path)?;
    let webp = encoder.encode(&input)?;
    std::fs::write(format!("{}.webp", path), webp)?;
}
```

## Library Installation

The native library can be installed in several ways:

### 1. Automatic Download (Recommended)

Enable the `auto-download` feature. When `libnextimage.a` is not found, the build script downloads
`libnextimage-v<crate version>-<platform>.tar.gz` from GitHub Releases into the cache directory
(`<cache>/lib/<platform>/`) and links it. The static library is needed at link time, so the download
happens during `cargo build`, not when your program runs.

`ensure_library()` checks that the library is in place (and downloads it into the cache if it is not):

```rust
nextimage::ensure_library()?;
```

The crate version must match a release tag (`v<crate version>`) that has the prebuilt archives.

### 2. Environment Variable

Set `LIBNEXTIMAGE_LIB_DIR` to point to the directory containing `libnextimage.a`:

```bash
export LIBNEXTIMAGE_LIB_DIR=/path/to/lib/darwin-arm64
```

### 3. Manual Installation

Download from [GitHub Releases](https://github.com/ideamans/libnextimage/releases) and place in the cache directory:

- Linux/macOS: `~/.cache/libnextimage/lib/<platform>/libnextimage.a`
- Windows: `%LOCALAPPDATA%/libnextimage/lib/<platform>/libnextimage.a`

## Supported Platforms

| Platform | Architecture | Status |
|----------|--------------|--------|
| macOS    | ARM64 (M1/M2/M3) | ✅ |
| macOS    | x86_64 (Intel) | ✅ |
| Linux    | x86_64 | ✅ |
| Linux    | ARM64 | ✅ |
| Windows  | x86_64 | ✅ |

## CLI Compatibility

This library produces **byte-for-byte identical** output to official CLI tools:

- `cwebp` / `dwebp` - WebP encoding/decoding
- `avifenc` / `avifdec` - AVIF encoding/decoding
- `gif2webp` / `webp2gif` - GIF conversion

## License

BSD-3-Clause

## Related Projects

- [libnextimage](https://github.com/ideamans/libnextimage) - Core C library
- [libwebp](https://chromium.googlesource.com/webm/libwebp) - WebP codec
- [libavif](https://github.com/AOMediaCodec/libavif) - AVIF codec
