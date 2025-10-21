# libnextimage Go Examples

This directory contains practical examples of using libnextimage in your Go projects.

## Quick Start

These examples demonstrate how a real user would integrate libnextimage into their project using `go get`.

### Prerequisites

- Go 1.24.7 or later
- A JPEG image file to convert (or use the sample command to download one)

### Running the Examples

**Note for local development**: If you're testing from within the libnextimage repository:
1. Uncomment the `replace` directive in `go.mod`
2. This will use your local version instead of the published version

#### 1. JPEG to WebP Conversion

Convert JPEG images to WebP format with customizable quality settings.

```bash
cd examples/golang

# Install dependencies (go.mod already specifies v0.3.0)
go mod tidy

# Download a sample image (or use your own)
curl -o sample.jpg https://raw.githubusercontent.com/ideamans/libnextimage/main/testdata/jpeg-source/gradient-horizontal.jpg

# Run the example
go run jpeg_to_webp.go sample.jpg output.webp

# With custom quality (0-100)
go run jpeg_to_webp.go input.jpg output.webp 85

# With lossless compression
go run jpeg_to_webp.go input.jpg output.webp lossless
```

#### 2. JPEG to AVIF Conversion

Convert JPEG images to AVIF format (next-generation image format).

```bash
go run jpeg_to_avif.go input.jpg output.avif

# With custom quality and speed
go run jpeg_to_avif.go input.jpg output.avif 60 6
```

#### 3. Batch Conversion

Convert multiple images in a directory.

```bash
go run batch_convert.go ./images ./output webp
```

## Example Descriptions

### jpeg_to_webp.go
Basic example showing how to:
- Initialize libnextimage
- Configure WebP encoding options
- Convert a JPEG file to WebP
- Handle errors appropriately

**Use case**: Simple image format conversion for web optimization

### jpeg_to_avif.go
Demonstrates AVIF encoding with:
- Quality and speed settings
- File size comparison
- Compression efficiency reporting

**Use case**: Next-generation image format for modern browsers

### batch_convert.go
Production-ready example featuring:
- Directory traversal
- Concurrent processing
- Progress reporting
- Error handling for multiple files

**Use case**: Bulk image conversion for websites or applications

## What Happens During Installation

When you run `go get github.com/ideamans/libnextimage/golang`:

1. Go downloads the source code
2. The package's `init()` function runs
3. It detects your platform (darwin-arm64, linux-amd64, etc.)
4. Downloads the appropriate pre-built `libnextimage.a` from GitHub Releases
5. Extracts it to the correct location
6. Ready to use!

**No manual compilation needed** - unless you want to build from source.

## Troubleshooting

### "Failed to download pre-built library"

If you see this error:

1. **Wait a few minutes** - CI might still be building the release
2. **Check your internet connection** - Download requires GitHub access
3. **Use a previous version**: `go get github.com/ideamans/libnextimage/golang@v0.2.0`
4. **Build from source**:
   ```bash
   cd ../../  # Go to repository root
   bash scripts/build-c-library.sh
   ```

### "Library file not found during build"

The library will automatically fall back to a previous compatible version if the exact version isn't available yet.

## Learn More

- [Main Documentation](../../README.md)
- [API Reference](../../golang/)
- [C API](../../c/)
- [Release Notes](https://github.com/ideamans/libnextimage/releases)

## Contributing Examples

Have a useful example? PRs welcome!

1. Create a new `.go` file
2. Add clear comments
3. Update this README
4. Test with a fresh `go mod init`
