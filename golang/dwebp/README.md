# dwebp - Go Bindings for WebP Decoding

This package provides Go bindings for the `dwebp` command, following the SPEC.md specification.

## Features

- ✅ SPEC.md-compliant interface design
- ✅ Convert WebP to PNG format
- ✅ Command reuse for multiple conversions
- ✅ Three usage patterns: `Run()`, `RunFile()`, `RunIO()`
- ✅ Decoding options support
- ✅ Thread-safe command instances

## Installation

```bash
go get github.com/ideamans/libnextimage/golang/dwebp
```

## Quick Start

### Basic Usage

```go
package main

import (
    "os"
    "github.com/ideamans/libnextimage/golang/dwebp"
)

func main() {
    // Create command with default options
    cmd, err := dwebp.NewCommand(nil)
    if err != nil {
        panic(err)
    }
    defer cmd.Close()

    // Read WebP file
    webpData, _ := os.ReadFile("image.webp")

    // Convert to PNG
    pngData, err := cmd.Run(webpData)
    if err != nil {
        panic(err)
    }

    // Save result
    os.WriteFile("image.png", pngData, 0644)
}
```

### Custom Options

```go
// Create custom options
opts := dwebp.NewDefaultOptions()
opts.Format = "RGB"
opts.UseThreads = true

// Create command with custom options
cmd, _ := dwebp.NewCommand(&opts)
defer cmd.Close()

pngData, _ := cmd.Run(webpData)
```

### File Conversion

```go
cmd, _ := dwebp.NewCommand(nil)
defer cmd.Close()

// Direct file conversion
err := cmd.RunFile("input.webp", "output.png")
```

### Stream Conversion

```go
import (
    "os"
    "github.com/ideamans/libnextimage/golang/dwebp"
)

cmd, _ := dwebp.NewCommand(nil)
defer cmd.Close()

input, _ := os.Open("input.webp")
output, _ := os.Create("output.png")

err := cmd.RunIO(input, output)
```

## Options

The `Options` struct supports WebP decoding parameters:

```go
type Options struct {
    Format            string // "RGBA", "RGB", "BGRA"
    BypassFiltering   bool   // Skip in-loop filtering
    NoFancyUpsampling bool   // Don't use fancy upsampler
    UseThreads        bool   // Use multi-threading
}
```

## Output Format

The decoded output is always in PNG format, making it easy to use with standard image processing tools and libraries.

## Command Reuse

Commands can be reused for multiple conversions:

```go
cmd, _ := dwebp.NewCommand(nil)
defer cmd.Close()

// Convert multiple WebP images
for _, webpPath := range webpPaths {
    webpData, _ := os.ReadFile(webpPath)
    pngData, _ := cmd.Run(webpData)
    // ... save pngData
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
