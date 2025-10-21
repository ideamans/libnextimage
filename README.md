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
- **Dual Distribution**: Both static (.a) and shared (.so/.dylib/.dll) libraries available

## Installation

### For Go Users

#### Quick Start

```bash
# Step 1: Get the module
go get github.com/ideamans/libnextimage/golang

# Step 2: Install pre-built libraries to your Go module cache (ONE-TIME per version)
bash <(curl -fsSL https://raw.githubusercontent.com/ideamans/libnextimage/main/scripts/setup-go-module.sh)

# On Windows (Git Bash/MSYS2), you may need administrator privileges
```

The setup script will:
1. Detect your platform (darwin-arm64, linux-amd64, windows-amd64, etc.)
2. Download the appropriate pre-built `libnextimage.a` from GitHub Releases
3. Install it to your Go module cache (may require sudo/administrator permissions)

**Alternative manual installation**:

```bash
# Using standalone install tool to install to current directory
go run github.com/ideamans/libnextimage/golang/cmd/install@latest

# Or using install script to project directory
git clone https://github.com/ideamans/libnextimage.git
cd libnextimage
bash scripts/install.sh

# Install specific version
bash scripts/install.sh v0.1.0
```

#### Option 2: Build from Source

If you prefer to build from source:

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

#### Quick Install

Use the install script to automatically download and extract the library:

```bash
# Clone the repository (no --recursive needed for pre-built binaries)
git clone https://github.com/ideamans/libnextimage.git
cd libnextimage

# Run install script (automatically detects your platform)
bash scripts/install.sh

# Or install specific version
bash scripts/install.sh v0.1.0
```

The script will download and install:
- `lib/<platform>/libnextimage.a` - Combined static library
- `include/*.h` - Header files

#### Shared Library (for Node.js/FFI)

For Node.js bindings or other FFI use cases, download the shared library version:

```bash
# Download shared library package
# Example for darwin-arm64:
curl -L https://github.com/ideamans/libnextimage/releases/download/v0.4.0/libnextimage-shared-v0.4.0-darwin-arm64.tar.gz | tar xz
```

This provides:
- `lib/<platform>/libnextimage.dylib` (macOS)
- `lib/<platform>/libnextimage.so` (Linux)
- `lib/<platform>/libnextimage.dll` (Windows)
- `include/*.h` - Header files

#### Manual Installation

1. Download the pre-built library for your platform from [Releases](https://github.com/ideamans/libnextimage/releases):

   **Static libraries** (for C/C++ and Go):
   - `libnextimage-v0.1.0-darwin-arm64.tar.gz` (macOS Apple Silicon)
   - `libnextimage-v0.1.0-darwin-amd64.tar.gz` (macOS Intel)
   - `libnextimage-v0.1.0-linux-amd64.tar.gz` (Linux x64)
   - `libnextimage-v0.1.0-linux-arm64.tar.gz` (Linux ARM64)
   - `libnextimage-v0.1.0-windows-amd64.tar.gz` (Windows x64)

   **Shared libraries** (for Node.js/FFI):
   - `libnextimage-shared-v0.1.0-darwin-arm64.tar.gz` (macOS Apple Silicon)
   - `libnextimage-shared-v0.1.0-darwin-amd64.tar.gz` (macOS Intel)
   - `libnextimage-shared-v0.1.0-linux-amd64.tar.gz` (Linux x64)
   - `libnextimage-shared-v0.1.0-linux-arm64.tar.gz` (Linux ARM64)
   - `libnextimage-shared-v0.1.0-windows-amd64.tar.gz` (Windows x64)

2. Extract the archive:
   ```bash
   tar xzf libnextimage-v0.1.0-<platform>.tar.gz
   ```

3. The archive contains:
   ```
   lib/<platform>/libnextimage.a  # Platform-specific library
   include/*.h                     # Header files
   include/nextimage/*.h           # Command-line API headers
   ```

4. Copy files to your project:
   ```bash
   # Example for darwin-arm64
   cp lib/darwin-arm64/libnextimage.a /path/to/your/project/
   cp -r include/* /path/to/your/project/include/
   ```

5. Link in your build:
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

## Installation Tools

### Go Module Setup Script (Recommended for Go users)

Install pre-built libraries directly to your Go module cache:

```bash
# Install for current module version
bash <(curl -fsSL https://raw.githubusercontent.com/ideamans/libnextimage/main/scripts/setup-go-module.sh)

# Install specific version
bash <(curl -fsSL https://raw.githubusercontent.com/ideamans/libnextimage/main/scripts/setup-go-module.sh) v0.3.0
```

This script will:
- Auto-detect your platform (darwin-arm64, linux-amd64, windows-amd64, etc.)
- Find your Go module cache location
- Download and extract the pre-built library to the correct module directory
- Handle read-only module cache (may require sudo/administrator permissions)

**Note**: On Windows, run in Git Bash or MSYS2 with administrator privileges.

### Go Installation Tool

Standalone tool for installing to custom directories:

```bash
# Install to current directory
go run github.com/ideamans/libnextimage/golang/cmd/install@latest

# Install to specific directory
go run github.com/ideamans/libnextimage/golang/cmd/install@latest -dir /path/to/project

# Force re-download even if library exists
go run github.com/ideamans/libnextimage/golang/cmd/install@latest -force

# List available platforms
go run github.com/ideamans/libnextimage/golang/cmd/install@latest -list
```

### Shell Installation Script

For C/C++ projects or manual installation:

```bash
# Install latest version to current directory
bash scripts/install.sh

# Install specific version
bash scripts/install.sh v0.2.0
```

The script features:
- Platform auto-detection (darwin-arm64, linux-amd64, etc.)
- Interactive confirmation before overwriting existing files
- Detailed progress and error reporting
- Works with curl or wget

## Makefile Targets

For convenient building and testing, use the provided Makefile:

```bash
# Show all available targets
make help

# C Library
make build-c      # Build C library (libnextimage.a)
make test-c       # Run C tests
make install-c    # Build and install to lib/ directory
make clean-c      # Clean C build artifacts

# Go Package
make test-go      # Run Go tests

# Combined
make test-all     # Run both C and Go tests
make clean-all    # Clean all build artifacts
```

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

Current version: **0.3.0**

### Release Process

When a new version is tagged, there's a brief window (typically 10-20 minutes) during which:
- The Go code is immediately available via `go get`
- Pre-built binaries are still being compiled by CI

During this window, the auto-download will:
1. First try to download the exact version
2. If unavailable, automatically fall back to the previous stable version
3. Display a notice that binaries will be available soon

Users can also:
- Wait a few minutes for CI to complete
- Manually specify a previous version: `go get github.com/ideamans/libnextimage/golang@v0.2.0`
- Build from source using `bash scripts/build-c-library.sh`

### For Maintainers: Release Checklist

**Recommended Release Process** (eliminates version mismatch window):

```bash
# 1. Update version in code
vim golang/version.go  # Update LibraryVersion

# 2. Commit version bump
git add golang/version.go
git commit -m "Bump version to v0.4.0"
git push

# 3. Create repository tag (triggers CI build)
git tag v0.4.0
git push origin v0.4.0

# 4. Wait for GitHub Actions to complete (~15 minutes)
gh run watch --repo ideamans/libnextimage

# 5. Verify all binaries are available in release
gh release view v0.4.0

# 6. Once confirmed, create Go module tag
git tag golang/v0.4.0
git push origin golang/v0.4.0

# Done! Users can now: go get github.com/ideamans/libnextimage/golang@v0.4.0
```

**Or use the automated script:**
```bash
# After v0.4.0 tag and binaries are ready
bash scripts/release-golang-module.sh v0.4.0
```

The script will:
- Verify repository tag exists
- Check all platform binaries are available
- Create and push golang/v0.4.0 tag
- Provide confirmation and rollback instructions

**Why this order?**
- Step 3 triggers binary builds for all platforms
- Step 6 makes the Go module available only after binaries are ready
- No user will ever experience "binaries not yet available" errors

**Rollback if needed:**
```bash
# If build fails, delete tags and fix issues
git tag -d v0.4.0 golang/v0.4.0
git push origin :refs/tags/v0.4.0 :refs/tags/golang/v0.4.0
```

The library is designed to be backwards compatible, so using v0.3.0 binaries with v0.4.0 code is safe for patch and minor version updates.

## Support

- Issues: https://github.com/ideamans/libnextimage/issues
- Releases: https://github.com/ideamans/libnextimage/releases

