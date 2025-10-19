# libnextimage

A high-performance C library for WebP and AVIF image encoding/decoding with Go bindings.

This library provides a unified, command-line compatible interface to libwebp and libavif, designed to match the behavior of official CLI tools (`cwebp`, `dwebp`, `avifenc`, etc.) while offering convenient programmatic access.

## Features

- **WebP Support**: Full encoding and decoding with all `cwebp`/`dwebp` options
- **AVIF Support**: Complete `avifenc`/`avifdec` functionality
- **GIF ↔ WebP**: Animated GIF conversion (gif2webp, webp2gif)
- **Binary Compatibility**: Produces identical output to official CLI tools
- **Zero Dependencies**: Single static library with all dependencies included
- **Cross-Platform**: macOS (Intel/ARM), Linux (x64/ARM64), Windows

## Installation

### For Go Users

#### Option 1: Automatic Download (Recommended)

When you use `go get`, the library will automatically download pre-built binaries from GitHub Releases:

```bash
go get github.com/ideamans/libnextimage/golang
```

The library will automatically:
1. Detect your platform (darwin-arm64, linux-amd64, etc.)
2. Download the appropriate `libnextimage.a` from GitHub Releases
3. Extract it to the correct location

#### Option 2: Build from Source

If you prefer to build from source or the auto-download fails:

```bash
git clone --recursive https://github.com/ideamans/libnextimage.git
cd libnextimage
bash scripts/build-c-library.sh
```

Then use it in your Go project:

```go
import "github.com/ideamans/libnextimage/golang"
```

### For C/C++ Users

1. Download the pre-built library for your platform from [Releases](https://github.com/ideamans/libnextimage/releases):
   - `libnextimage-v0.1.0-darwin-arm64.tar.gz` (macOS Apple Silicon)
   - `libnextimage-v0.1.0-darwin-amd64.tar.gz` (macOS Intel)
   - `libnextimage-v0.1.0-linux-amd64.tar.gz` (Linux x64)
   - `libnextimage-v0.1.0-linux-arm64.tar.gz` (Linux ARM64)

2. Extract the archive:
   ```bash
   tar xzf libnextimage-v0.1.0-<platform>.tar.gz
   ```

3. Copy files to your project:
   ```bash
   cp lib/libnextimage.a /path/to/your/project/
   cp -r include/* /path/to/your/project/include/
   ```

4. Link in your build:
   ```bash
   gcc your_code.c -I./include -L. -lnextimage -ljpeg -lpng -lgif -lz -lc++ -o your_program
   ```

## Usage

### Go API

#### WebP Encoding

```go
import "github.com/ideamans/libnextimage/golang"

// Encode PNG to WebP
data, err := os.ReadFile("input.png")
opts := libnextimage.DefaultWebPEncodeOptions()
opts.Quality = 80
opts.Lossless = false

webpData, err := libnextimage.WebPEncode(data, opts)
os.WriteFile("output.webp", webpData, 0644)
```

#### AVIF Encoding

```go
// Encode JPEG to AVIF
data, err := os.ReadFile("input.jpg")
opts := libnextimage.DefaultAVIFEncodeOptions()
opts.Quality = 60
opts.Speed = 6

avifData, err := libnextimage.AVIFEncode(data, opts)
os.WriteFile("output.avif", avifData, 0644)
```

#### File-based API

```go
// Encode file directly
opts := libnextimage.DefaultWebPEncodeOptions()
opts.Quality = 90
webpData, err := libnextimage.WebPEncodeFile("input.png", opts)
```

### C API

```c
#include "nextimage.h"
#include "webp.h"

// Encode WebP
NextImageWebPEncodeOptions opts;
nextimage_webp_default_encode_options(&opts);
opts.quality = 80;
opts.lossless = 0;

NextImageBuffer input, output;
// ... load input data ...

NextImageStatus status = nextimage_webp_encode(
    input.data, input.size,
    &opts,
    &output
);

if (status == NEXTIMAGE_STATUS_OK) {
    // Use output.data, output.size
    nextimage_buffer_free(&output);
}
```

## Architecture

### Single Library Design

Unlike typical C libraries that require linking multiple dependencies, libnextimage provides a **single static library** (`libnextimage.a`) containing:

- nextimage core (100KB)
- libwebp (640KB)
- libavif (280KB)
- libaom (7.8MB) - AV1 codec
- Helper libraries (libwebpdemux, libwebpmux, libsharpyuv, etc.)

**Total: ~8.9MB** - Everything you need in one file.

### Directory Structure

```
libnextimage/
├── lib/
│   ├── darwin-arm64/
│   │   └── libnextimage.a    # macOS Apple Silicon
│   ├── darwin-amd64/
│   │   └── libnextimage.a    # macOS Intel
│   ├── linux-amd64/
│   │   └── libnextimage.a    # Linux x64
│   └── linux-arm64/
│       └── libnextimage.a    # Linux ARM64
├── include/
│   ├── nextimage.h           # Core API
│   ├── webp.h                # WebP API
│   ├── avif.h                # AVIF API
│   └── nextimage/            # Command-line compatible APIs
│       ├── cwebp.h
│       ├── dwebp.h
│       ├── avifenc.h
│       └── avifdec.h
├── c/                        # C source code
├── golang/                   # Go bindings
├── scripts/
│   └── build-c-library.sh    # Build script
└── deps/                     # Git submodules
    ├── libwebp/
    └── libavif/
```

## Building from Source

### Requirements

- CMake 3.15+
- C11 compiler (GCC, Clang, or MSVC)
- System libraries: libjpeg, libpng, libgif

### macOS

```bash
brew install cmake jpeg libpng giflib
bash scripts/build-c-library.sh
```

### Linux

```bash
sudo apt-get install cmake build-essential libjpeg-dev libpng-dev libgif-dev
bash scripts/build-c-library.sh
```

### Build Output

The script will generate:
- `lib/<platform>/libnextimage.a` - Combined static library
- `include/*.h` - Header files

## Testing

### Test Images

All test images are committed to the repository, so no additional generation is needed. The test suite includes various image sizes and formats:

- `testdata/png-source/large-2048x2048.png` (20MB) - Large image for memory and performance testing
- `testdata/png-source/hd-1920x1080.png` (10MB) - HD resolution testing
- Various smaller images for different test scenarios (transparency, gradients, compression, etc.)

### Go Tests

```bash
cd golang
go test -v
```

The test suite includes:
- 160+ compatibility tests verifying binary-exact matching with CLI tools
- Integration tests
- Round-trip encoding/decoding tests

**All tests produce byte-for-byte identical output to official tools!**

## Compatibility

This library is designed to be a **perfect clone** of the official CLI tools:

| Tool | Status | Binary Match |
|------|--------|--------------|
| cwebp | ✅ Complete | 100% |
| dwebp | ✅ Complete | 100% |
| gif2webp | ✅ Complete | 100% |
| webp2gif | ✅ Complete | N/A |
| avifenc | ✅ Complete | 100% |
| avifdec | ✅ Complete | 100% |

All encoding options produce **byte-for-byte identical** output to the official tools.

## License

This project is licensed under the BSD 3-Clause License.

- libwebp: BSD License
- libavif: BSD License
- libaom: BSD License

## Contributing

Contributions are welcome! Please ensure:
1. All tests pass (`go test -v`)
2. Binary compatibility is maintained (new compatibility tests for new features)
3. Code follows the existing style

## Versioning

We use [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking API changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

Current version: **0.1.0**

## Support

- Issues: https://github.com/ideamans/libnextimage/issues
- Releases: https://github.com/ideamans/libnextimage/releases

