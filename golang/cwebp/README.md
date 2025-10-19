# cwebp - Go Bindings for WebP Encoding

This package provides Go bindings for the `cwebp` command, following the SPEC.md specification.

## Features

- ✅ SPEC.md-compliant interface design
- ✅ Convert JPEG/PNG to WebP format
- ✅ Command reuse for multiple conversions
- ✅ Three usage patterns: `Run()`, `RunFile()`, `RunIO()`
- ✅ Comprehensive options support
- ✅ Thread-safe command instances

## Installation

```bash
go get github.com/ideamans/libnextimage/golang/cwebp
```

## Quick Start

### Basic Usage

```go
package main

import (
    "os"
    "github.com/ideamans/libnextimage/golang/cwebp"
)

func main() {
    // Create command with default options
    cmd, err := cwebp.NewCommand(nil)
    if err != nil {
        panic(err)
    }
    defer cmd.Close()

    // Read JPEG/PNG file
    imageData, _ := os.ReadFile("image.jpg")

    // Convert to WebP
    webpData, err := cmd.Run(imageData)
    if err != nil {
        panic(err)
    }

    // Save result
    os.WriteFile("image.webp", webpData, 0644)
}
```

### Custom Options

```go
// Create custom options
opts := cwebp.NewDefaultOptions()
opts.Quality = 80
opts.Method = 4
opts.Lossless = false

// Create command with custom options
cmd, _ := cwebp.NewCommand(&opts)
defer cmd.Close()

webpData, _ := cmd.Run(imageData)
```

### File Conversion

```go
cmd, _ := cwebp.NewCommand(nil)
defer cmd.Close()

// Direct file conversion
err := cmd.RunFile("input.jpg", "output.webp")
```

### Stream Conversion

```go
import (
    "os"
    "github.com/ideamans/libnextimage/golang/cwebp"
)

cmd, _ := cwebp.NewCommand(nil)
defer cmd.Close()

input, _ := os.Open("input.jpg")
output, _ := os.Create("output.webp")

err := cmd.RunIO(input, output)
```

## Options

The `Options` struct supports all WebP encoding parameters:

```go
type Options struct {
    Quality          float32  // 0-100, default: 75
    Lossless         bool     // Lossless mode
    Method           int      // 0-6, compression method
    TargetSize       int      // Target size in bytes
    TargetPSNR       float32  // Target PSNR
    Segments         int      // Number of segments
    SNSStrength      int      // Spatial noise shaping
    FilterStrength   int      // Filter strength
    FilterSharpness  int      // Filter sharpness
    FilterType       int      // Filtering type
    Autofilter       bool     // Auto-adjust filter
    AlphaCompression int      // Alpha compression
    AlphaFiltering   int      // Alpha filtering
    AlphaQuality     int      // Alpha quality
    Pass             int      // Number of passes
    Preprocessing    int      // Preprocessing filter
    Partitions       int      // Number of partitions
    PartitionLimit   int      // Quality degradation
    EmulateJPEGSize  bool     // Emulate JPEG size
    ThreadLevel      int      // Multi-threading level
    LowMemory        bool     // Low memory mode
    NearLossless     int      // Near-lossless preset
    Exact            bool     // Preserve RGB values
    UseDeltaPalette  bool     // Use delta palette
    UseSharpYUV      bool     // Use sharp YUV conversion
}
```

## Command Reuse

Commands can be reused for multiple conversions, which is more efficient than creating a new command for each conversion:

```go
cmd, _ := cwebp.NewCommand(nil)
defer cmd.Close()

// Convert multiple images with the same command
for _, imagePath := range imagePaths {
    imageData, _ := os.ReadFile(imagePath)
    webpData, _ := cmd.Run(imageData)
    // ... save webpData
}
```

## Thread Safety

Each `Command` instance is safe for sequential use. For concurrent conversions, create separate command instances for each goroutine.

## Testing

```bash
go test -v
```

All tests pass successfully with comprehensive coverage.

## License

See the main libnextimage LICENSE file.
