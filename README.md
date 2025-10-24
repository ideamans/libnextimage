# libnextimage

High-performance WebP and AVIF image encoding/decoding library with bindings for Go, TypeScript/Node.js, Bun, and Deno.

This library provides a unified interface to libwebp and libavif, designed to match the behavior of official CLI tools (`cwebp`, `dwebp`, `avifenc`, `avifdec`) while offering convenient programmatic access across multiple languages.

## Features

- **WebP & AVIF Support**: Full encoding and decoding capabilities
- **Multiple Language Bindings**: Go, TypeScript/Node.js, Bun, Deno
- **Binary Compatibility**: Produces identical output to official CLI tools
- **Cross-Platform**: macOS (Intel/ARM), Linux (x64/ARM64), Windows
- **Easy Installation**: Automatic library download via package managers
- **Zero External Dependencies**: All required libraries bundled

## Installation

### For Go Users

#### Quick Start (Recommended)

**Option 1: Automatic Download in Code**

The library will automatically download pre-built binaries to a writable cache directory (`~/.cache/libnextimage` or `$XDG_CACHE_HOME/libnextimage`) when you call `EnsureLibrary()`:

```go
package main

import (
    "log"
    "github.com/ideamans/libnextimage/golang"
)

func main() {
    // Ensure library is available (downloads if necessary)
    if err := libnextimage.EnsureLibrary(); err != nil {
        log.Fatal(err)
    }

    // Now you can use the library
    // ...
}
```

**Option 2: Manual Pre-Installation**

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

#### Environment Variables

- `LIBNEXTIMAGE_CACHE_DIR`: Custom cache directory for downloaded libraries (default: `~/.cache/libnextimage`)
- `XDG_CACHE_HOME`: Standard XDG cache directory (used if `LIBNEXTIMAGE_CACHE_DIR` is not set)

#### Build from Source

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

### For TypeScript/Node.js Users

Install via npm - the native library will be automatically downloaded during installation:

```bash
npm install @ideamans/libnextimage
```

The package automatically downloads the appropriate native library for your platform (macOS, Linux, or Windows) from GitHub Releases.

**Quick Start:**

```typescript
import { WebPEncoder, AVIFEncoder } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

// Encode PNG to WebP
const pngData = readFileSync('input.png')
const webpEncoder = new WebPEncoder({ quality: 90 })
const webpData = webpEncoder.encode(pngData)
webpEncoder.close()
writeFileSync('output.webp', webpData)

// Encode to AVIF
const avifEncoder = new AVIFEncoder({ quality: 60, speed: 6 })
const avifData = avifEncoder.encode(pngData)
avifEncoder.close()
writeFileSync('output.avif', avifData)
```

See [typescript/README.md](typescript/README.md) for detailed documentation.

### For C/C++ Users

Download pre-built binaries from [GitHub Releases](https://github.com/ideamans/libnextimage/releases) or use the install script:

```bash
# Clone repository
git clone https://github.com/ideamans/libnextimage.git
cd libnextimage

# Install for your platform (downloads from GitHub Releases)
bash scripts/install.sh

# Or install specific version
bash scripts/install.sh v0.4.0
```

This installs:
- `lib/<platform>/libnextimage.a` - Static library (for C/C++ and Go)
- `include/*.h` - Header files

For more details, see the [Building from Source](#building-from-source) section.

## Usage Examples

### Go

```go
import (
    "os"
    "github.com/ideamans/libnextimage/golang"
)

// Ensure library is available (downloads if necessary)
if err := libnextimage.EnsureLibrary(); err != nil {
    panic(err)
}

// Encode PNG to WebP
data, _ := os.ReadFile("input.png")
opts := libnextimage.DefaultWebPEncodeOptions()
opts.Quality = 90

webpData, _ := libnextimage.WebPEncode(data, opts)
os.WriteFile("output.webp", webpData, 0644)

// Encode to AVIF
avifOpts := libnextimage.DefaultAVIFEncodeOptions()
avifOpts.Quality = 60
avifOpts.Speed = 6

avifData, _ := libnextimage.AVIFEncode(data, avifOpts)
os.WriteFile("output.avif", avifData, 0644)
```

### TypeScript/Node.js

```typescript
import { WebPEncoder, AVIFEncoder } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

const imageData = readFileSync('input.png')

// WebP encoding
const webpEncoder = new WebPEncoder({ quality: 90 })
const webpData = webpEncoder.encode(imageData)
webpEncoder.close()
writeFileSync('output.webp', webpData)

// AVIF encoding
const avifEncoder = new AVIFEncoder({ quality: 60, speed: 6 })
const avifData = avifEncoder.encode(imageData)
avifEncoder.close()
writeFileSync('output.avif', avifData)
```

### C

```c
#include "nextimage.h"
#include "webp.h"

NextImageWebPEncodeOptions opts;
nextimage_webp_default_encode_options(&opts);
opts.quality = 90;

NextImageBuffer input, output;
// ... load input data ...

NextImageStatus status = nextimage_webp_encode(
    input.data, input.size, &opts, &output
);

if (status == NEXTIMAGE_STATUS_OK) {
    // Use output.data, output.size
    nextimage_buffer_free(&output);
}
```

## Supported Platforms

- **macOS**: Intel (x64), Apple Silicon (ARM64)
- **Linux**: x64, ARM64
- **Windows**: x64

Pre-built binaries are available for all platforms via [GitHub Releases](https://github.com/ideamans/libnextimage/releases).

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

### Go

```bash
cd golang
go test -v
```

### TypeScript/Node.js

```bash
cd typescript
npm install
npm test
```

All tests verify binary-exact matching with official CLI tools.

## CLI Tool Compatibility

This library produces **byte-for-byte identical** output to official CLI tools:

- ✅ `cwebp` / `dwebp` - WebP encoding/decoding
- ✅ `avifenc` / `avifdec` - AVIF encoding/decoding
- ✅ `gif2webp` / `webp2gif` - GIF conversion

## License

This project is licensed under the BSD 3-Clause License.

- libwebp: BSD License
- libavif: BSD License
- libaom: BSD License

## Documentation

- [TypeScript/Node.js Documentation](typescript/README.md)
- [Examples](examples/)

## Contributing

Contributions are welcome! Please ensure all tests pass before submitting pull requests.

## Support

- Issues: https://github.com/ideamans/libnextimage/issues
- Releases: https://github.com/ideamans/libnextimage/releases

