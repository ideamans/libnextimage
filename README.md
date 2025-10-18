# libnextimage

High-performance WebP and AVIF encoding/decoding library with FFI interface for Go.

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/ideamans/libnextimage)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0--alpha-orange)](DEPENDENCIES.txt)

## Overview

`libnextimage` provides direct FFI access to WebP and AVIF encoding/decoding functionality, eliminating the overhead of spawning separate processes for image conversion operations.

### Key Features

- **Zero Process Overhead**: Direct library calls instead of spawning cwebp/dwebp/avifenc/avifdec processes
- **Multi-plane Support**: Full support for YUV planar formats (4:2:0, 4:2:2, 4:4:4)
- **High Bit Depth**: Support for 8-bit, 10-bit, and 12-bit AVIF encoding
- **Memory Safe**: Comprehensive testing with AddressSanitizer and UndefinedBehaviorSanitizer
- **Thread Safe**: Thread-local error handling and concurrent encoding/decoding
- **Go Integration**: Idiomatic Go API with explicit functions for bytes, files, and streams

## Development Status

✅ **Phase 1 Complete** - Foundation layer implemented

- ✅ Core FFI interface
- ✅ Memory management (alloc/into separation)
- ✅ Error handling (thread-local messages)
- ✅ CMake build system (normal/Debug/ASan/UBSan)
- ✅ Basic test suite

✅ **Phase 2 Complete** - WebP integration

- ✅ libwebp integration (git submodule)
- ✅ WebP C FFI (encode/decode)
- ✅ Go bindings with explicit API
- ✅ Comprehensive Go tests (12 tests all passing)

✅ **Phase 3 Complete** - AVIF integration

- ✅ libavif integration (git submodule with AOM codec)
- ✅ AVIF C FFI (encode/decode with YUV format support)
- ✅ Go bindings for AVIF
- ✅ Comprehensive AVIF tests (13 Go tests + C tests all passing)
- ✅ Support for 8/10/12-bit depth, YUV 4:4:4/4:2:2/4:2:0/4:0:0
- ✅ Combined static library (libnextimage.a) with all dependencies

✅ **Phase 4 Complete** - WebP↔GIF conversion

- ✅ WebP to GIF conversion (webp2gif)
- ✅ 256-color quantization (6x6x6 RGB cube + grayscale)
- ✅ Transparency support
- ✅ Memory-based GIF encoding
- ✅ Go bindings for GIF conversion

🚧 **Phase 4.5 In Progress** - Command-line compatibility verification

- ✅ Test data generation (39 test images)
- ✅ CLI tools build automation (cwebp, dwebp, avifenc, avifdec)
- ✅ Go test framework for compatibility testing
- ✅ **WebP Encoding compatibility: 100% COMPLETE** ✨
  - ✅ Quality options (0, 25, 50, 75, 90, 100): **binary-exact match**
  - ✅ Lossless mode: **binary-exact match**
  - ✅ Method options (0, 2, 4, 6): **binary-exact match**
  - ✅ Size variations (16x16 to 2048x2048): **binary-exact match**
  - ✅ Alpha channel variations: **binary-exact match**
  - ✅ Compression characteristics: **binary-exact match**
  - ✅ AlphaQuality (0, 50, 100): **binary-exact match**
  - ✅ Exact mode: **binary-exact match**
  - ✅ Pass options (1, 5, 10): **binary-exact match**
  - ✅ Option combinations: **binary-exact match**
  - **Total: 38/38 encoding tests passing with binary-exact match!**
- 🚧 **WebP Decoding compatibility: 54.5% passing**
  - ✅ Default lossy decoding: **pixel-exact match**
  - ✅ Default lossless decoding: **pixel-exact match**
  - ✅ Multi-threading (large images): **pixel-exact match**
  - 🚧 NoFancy, NoFilter, alpha-gradient: investigating pixel differences
  - **Total: 6/11 decoding tests passing (basic cases working)**
- ⏳ AVIF compatibility testing

## Usage Examples (Go)

### WebP Encoding/Decoding

```go
package main

import (
    "fmt"
    "os"
    "github.com/ideamans/libnextimage/golang"
)

func main() {
    // Create test RGBA image
    width, height := 640, 480
    rgbaData := make([]byte, width*height*4)
    // ... fill with your image data ...

    // Encode to WebP
    opts := libnextimage.DefaultWebPEncodeOptions()
    opts.Quality = 90.0
    opts.Method = 6  // Higher quality

    webpData, err := libnextimage.WebPEncodeBytes(
        rgbaData, width, height,
        libnextimage.FormatRGBA,
        opts,
    )
    if err != nil {
        panic(err)
    }

    // Save WebP file
    os.WriteFile("output.webp", webpData, 0644)
    fmt.Printf("Encoded to WebP: %d bytes\n", len(webpData))

    // Decode WebP
    decoded, err := libnextimage.WebPDecodeBytes(
        webpData,
        libnextimage.DefaultWebPDecodeOptions(),
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Decoded: %dx%d, format=%d, bit_depth=%d\n",
        decoded.Width, decoded.Height,
        decoded.Format, decoded.BitDepth)
}
```

### AVIF Encoding with Advanced Options

```go
import "github.com/ideamans/libnextimage/golang"

// Encode PNG to AVIF with high quality
opts := libnextimage.DefaultAVIFEncodeOptions()
opts.Quality = 90
opts.Speed = 4
opts.YUVFormat = 0  // 4:4:4 for best quality

avifData, err := libnextimage.AVIFEncodeFile("input.png", opts)
if err != nil {
    panic(err)
}

os.WriteFile("output.avif", avifData, 0644)
```

### AVIF to PNG/JPEG Conversion

```go
import "github.com/ideamans/libnextimage/golang"

// Convert AVIF to PNG with compression
avifData, _ := os.ReadFile("input.avif")
decOpts := libnextimage.DefaultAVIFDecodeOptions()
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingBestQuality

err := libnextimage.AVIFDecodeToPNG(
    avifData,
    "output.png",
    decOpts,
    9,  // PNG compression level (0-9, -1=default)
)

// Convert AVIF to JPEG
err = libnextimage.AVIFDecodeToJPEG(
    avifData,
    "output.jpg",
    decOpts,
    90,  // JPEG quality (1-100)
)

// File-based conversion
err = libnextimage.AVIFDecodeFileToPNG(
    "input.avif",
    "output.png",
    decOpts,
    -1,  // default compression
)
```

### Chroma Upsampling Options

```go
import "github.com/ideamans/libnextimage/golang"

decOpts := libnextimage.DefaultAVIFDecodeOptions()

// Available upsampling modes:
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingAutomatic   // 0 (default)
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingFastest     // 1
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingBestQuality // 2
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingNearest     // 3
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingBilinear    // 4
```

### Security Limits for AVIF Decoding

```go
import "github.com/ideamans/libnextimage/golang"

decOpts := libnextimage.DefaultAVIFDecodeOptions()
decOpts.ImageSizeLimit = 100_000_000      // Max 100M pixels
decOpts.ImageDimensionLimit = 16384       // Max 16384px width/height
decOpts.StrictFlags = 1                   // Enable strict validation

decoded, err := libnextimage.AVIFDecodeBytes(avifData, decOpts)
```

## Quick Start

### Building from Source

#### Prerequisites

- CMake 3.15 or later
- C11-compatible compiler (GCC, Clang, or MSVC)
- Git (for submodule management)

#### Basic Build

```bash
# Clone the repository
git clone --recursive https://github.com/ideamans/libnextimage.git
cd libnextimage

# Build the C library
cd c
mkdir build && cd build
cmake ..
cmake --build .

# Run tests
ctest --output-on-failure
```

#### Debug Build (with Leak Counter)

```bash
cd c
mkdir build-debug && cd build-debug
cmake -DCMAKE_BUILD_TYPE=Debug ..
cmake --build .
ctest --output-on-failure
```

#### AddressSanitizer Build

```bash
cd c
mkdir build-asan && cd build-asan
cmake -DENABLE_ASAN=ON ..
cmake --build .
ctest --output-on-failure
```

#### UndefinedBehaviorSanitizer Build

```bash
cd c
mkdir build-ubsan && cd build-ubsan
cmake -DENABLE_UBSAN=ON ..
cmake --build .
ctest --output-on-failure
```

## Project Structure

```
libnextimage/
├── c/                        # C FFI layer
│   ├── include/              # Public headers
│   │   └── nextimage.h       # Main FFI interface
│   ├── src/                  # Implementation
│   │   ├── common.c          # Memory & error handling
│   │   └── internal.h        # Internal helpers
│   ├── test/                 # C tests
│   │   ├── basic_test.c
│   │   ├── leak_counter_test.c
│   │   └── sanitizer/
│   └── CMakeLists.txt
├── deps/                     # Dependencies (git submodules)
│   └── libwebp/              # WebP library
├── golang/                   # Go bindings
│   ├── common.go             # Common types & utilities
│   ├── webp.go               # WebP Go API
│   └── webp_test.go          # Comprehensive tests
├── lib/                      # Pre-compiled libraries (TBD)
├── SPEC.md                   # Detailed specification
├── DEPENDENCIES.txt          # Bill of Materials
└── LICENSE                   # MIT License
```

## Documentation

- [SPEC.md](SPEC.md) - Comprehensive specification and development plan
- [DEPENDENCIES.txt](DEPENDENCIES.txt) - Bill of Materials with all dependencies
- [CLAUDE.md](CLAUDE.md) - AI assistant context for development

## Testing

### Test Categories

1. **Basic Tests** - Core functionality verification (C)
2. **WebP Tests** - WebP encoding/decoding (C)
3. **Leak Counter Tests** - C heap leak detection (Debug build only)
4. **Sanitizer Tests** - Memory safety validation (ASan/UBSan builds)
5. **Go Tests** - WebP Go bindings and integration

### Running C Tests

```bash
# All C tests
cd c/build
ctest --output-on-failure

# Specific test
./basic_test
./webp_test

# Debug build with leak counter
cd build-debug && ./leak_counter_test

# ASan build
cd build-asan && ./sanitizer_test
```

### Running Go Tests

```bash
cd golang

# All tests (verbose)
go test -v

# Specific test
go test -v -run TestWebPEncode

# With race detector
go test -race

# Benchmarks
go test -bench=. -benchmem

# Coverage
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

This project is currently in active development (Phase 3 complete). Contributions will be welcome once the core API is stabilized.

## Roadmap

- [x] **Phase 1**: Foundation (Complete)
  - Core FFI interface
  - Memory management
  - Error handling
  - Build system
- [x] **Phase 2**: WebP Integration (Complete)
  - libwebp integration
  - WebP C FFI (encode/decode)
  - Go bindings with explicit API
  - Comprehensive tests (C & Go)
- [x] **Phase 3**: AVIF Integration (Complete)
  - libavif integration with AOM codec
  - AVIF C FFI (encode/decode)
  - Go bindings for AVIF
  - YUV format support (4:4:4/4:2:2/4:2:0/4:0:0)
  - 8/10/12-bit depth support
  - Comprehensive tests (C & Go)
- [ ] **Phase 4**: New Features (Week 7)
  - webp2gif conversion
- [ ] **Phase 5**: Security & Fuzzing (Weeks 8-9)
- [ ] **Phase 6**: Optimization (Weeks 10-11)
- [ ] **Phase 7**: Release (Week 12)

See [SPEC.md](SPEC.md) for the complete development plan.

## Credits

Developed by [Ideamans Inc.](https://www.ideamans.com/)

Built on top of:
- [libwebp](https://github.com/webmproject/libwebp) (planned)
- [libavif](https://github.com/AOMediaCodec/libavif) (planned)
