# libnextimage - Go Bindings

High-performance WebP, AVIF, and GIF image processing library for Go.

## Features

- **WebP Encoding & Decoding**: Fast WebP image processing with full control over quality and encoding options
- **AVIF Encoding & Decoding**: Modern AVIF format support with quality and speed presets
- **GIF Conversion**: Convert between GIF and WebP formats, including animated GIFs
- **CGO-based Performance**: Direct C library bindings for maximum performance
- **Automatic Resource Management**: Automatic cleanup using `runtime.SetFinalizer`
- **Idiomatic Go API**: Clean, error-returning Go interfaces
- **Cross-Platform**: Supports macOS (ARM64/x64), Linux (x64/ARM64), and Windows (x64)

## Installation

```bash
go get github.com/ideamans/libnextimage/golang
```

## Quick Start

### WebP Encoding

```go
package main

import (
    "fmt"
    "os"

    "github.com/ideamans/libnextimage/golang/webp"
)

func main() {
    // Read input image (JPEG, PNG, etc.)
    inputData, err := os.ReadFile("input.jpg")
    if err != nil {
        panic(err)
    }

    // Create encoder with options
    opts := webp.NewEncoderOptions()
    opts.Quality = 80
    opts.Method = 6

    encoder, err := webp.NewEncoder(opts)
    if err != nil {
        panic(err)
    }
    defer encoder.Close()

    // Encode to WebP
    webpData, err := encoder.Encode(inputData)
    if err != nil {
        panic(err)
    }

    // Save output
    if err := os.WriteFile("output.webp", webpData, 0644); err != nil {
        panic(err)
    }

    fmt.Printf("Converted: %d bytes → %d bytes\n", len(inputData), len(webpData))
}
```

### WebP Decoding

```go
package main

import (
    "fmt"
    "os"

    "github.com/ideamans/libnextimage/golang/webp"
)

func main() {
    // Read WebP file
    webpData, err := os.ReadFile("image.webp")
    if err != nil {
        panic(err)
    }

    // Create decoder
    decoder, err := webp.NewDecoder()
    if err != nil {
        panic(err)
    }
    defer decoder.Close()

    // Decode to RGBA
    decoded, err := decoder.Decode(webpData)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Width: %d, Height: %d\n", decoded.Width, decoded.Height)
    fmt.Printf("Data size: %d bytes\n", len(decoded.Data))
}
```

### AVIF Encoding

```go
package main

import (
    "os"

    "github.com/ideamans/libnextimage/golang/avif"
)

func main() {
    inputData, err := os.ReadFile("input.jpg")
    if err != nil {
        panic(err)
    }

    // Create encoder with options
    opts := avif.NewEncoderOptions()
    opts.Quality = 60
    opts.Speed = 6

    encoder, err := avif.NewEncoder(opts)
    if err != nil {
        panic(err)
    }
    defer encoder.Close()

    // Encode to AVIF
    avifData, err := encoder.Encode(inputData)
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("output.avif", avifData, 0644); err != nil {
        panic(err)
    }
}
```

### AVIF Decoding

```go
package main

import (
    "fmt"
    "os"

    "github.com/ideamans/libnextimage/golang/avif"
)

func main() {
    avifData, err := os.ReadFile("image.avif")
    if err != nil {
        panic(err)
    }

    decoder, err := avif.NewDecoder()
    if err != nil {
        panic(err)
    }
    defer decoder.Close()

    decoded, err := decoder.Decode(avifData)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Decoded %dx%d AVIF image\n", decoded.Width, decoded.Height)
}
```

### GIF to WebP Conversion

```go
package main

import (
    "fmt"
    "os"

    "github.com/ideamans/libnextimage/golang/gif2webp"
)

func main() {
    gifData, err := os.ReadFile("animated.gif")
    if err != nil {
        panic(err)
    }

    // Create converter with options
    opts := gif2webp.NewOptions()
    opts.Quality = 80
    opts.Method = 6

    cmd, err := gif2webp.NewCommand(opts)
    if err != nil {
        panic(err)
    }
    defer cmd.Close()

    // Convert GIF to WebP (preserves animation)
    webpData, err := cmd.Run(gifData)
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("animated.webp", webpData, 0644); err != nil {
        panic(err)
    }

    compression := (1.0 - float64(len(webpData))/float64(len(gifData))) * 100
    fmt.Printf("Compression: %.1f%%\n", compression)
}
```

### WebP to GIF Conversion

```go
package main

import (
    "os"

    "github.com/ideamans/libnextimage/golang/webp2gif"
)

func main() {
    webpData, err := os.ReadFile("image.webp")
    if err != nil {
        panic(err)
    }

    cmd, err := webp2gif.NewCommand(nil)
    if err != nil {
        panic(err)
    }
    defer cmd.Close()

    gifData, err := cmd.Run(webpData)
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("output.gif", gifData, 0644); err != nil {
        panic(err)
    }
}
```

## API Reference

### WebP Package

#### Encoder

```go
package webp

type EncoderOptions struct {
    Quality         float32
    Lossless        bool
    Method          int
    Preset          Preset
    ImageHint       ImageHint
    TargetSize      int
    TargetPSNR      float32
    Segments        int
    SNSStrength     int
    FilterStrength  int
    FilterSharpness int
    FilterType      int
    Autofilter      bool
    AlphaQuality    int
    Pass            int
    Exact           bool
    // ... many more options
}

func NewEncoderOptions() *EncoderOptions
func NewEncoder(opts *EncoderOptions) (*Encoder, error)
func (e *Encoder) Encode(data []byte) ([]byte, error)
func (e *Encoder) Close() error
```

**Common Options:**
- `Quality` (0-100): Quality level, default 75
- `Lossless` (bool): Use lossless encoding
- `Method` (0-6): Compression method, higher is slower but better
- `Preset`: Predefined configurations (Default, Picture, Photo, Drawing, Icon, Text)

#### Decoder

```go
type DecoderOptions struct {
    Format            PixelFormat
    UseThreads        bool
    BypassFiltering   bool
    NoFancyUpsampling bool
    CropX, CropY      int
    CropWidth         int
    CropHeight        int
    ScaleWidth        int
    ScaleHeight       int
}

type DecodedImage struct {
    Width  int
    Height int
    Data   []byte
    Format PixelFormat
}

func NewDecoderOptions() *DecoderOptions
func NewDecoder(opts *DecoderOptions) (*Decoder, error)
func (d *Decoder) Decode(data []byte) (*DecodedImage, error)
func (d *Decoder) Close() error
```

### AVIF Package

#### Encoder

```go
package avif

type EncoderOptions struct {
    Quality       int
    QualityAlpha  int
    Speed         int
    BitDepth      int
    YUVFormat     YUVFormat
    Lossless      bool
    Jobs          int
    AutoTiling    bool
    // ... more options
}

func NewEncoderOptions() *EncoderOptions
func NewEncoder(opts *EncoderOptions) (*Encoder, error)
func (e *Encoder) Encode(data []byte) ([]byte, error)
func (e *Encoder) Close() error
```

**Common Options:**
- `Quality` (0-100): Quality level, default 60
- `Speed` (0-10): Encoding speed, higher is faster (0=slowest/best, 10=fastest/worst)
- `BitDepth` (8/10/12): Bit depth per channel
- `YUVFormat`: Color format (YUV444, YUV422, YUV420, YUV400)

#### Decoder

```go
type DecoderOptions struct {
    Format              PixelFormat
    Jobs                int
    ChromaUpsampling    ChromaUpsampling
    IgnoreExif          bool
    IgnoreXMP           bool
    IgnoreICC           bool
    ImageSizeLimit      int
    ImageDimensionLimit int
}

func NewDecoderOptions() *DecoderOptions
func NewDecoder(opts *DecoderOptions) (*Decoder, error)
func (d *Decoder) Decode(data []byte) (*DecodedImage, error)
func (d *Decoder) Close() error
```

### GIF2WebP Package

```go
package gif2webp

type Options struct {
    // Same as WebP encoder options
    Quality  float32
    Lossless bool
    Method   int
    // ... etc
}

type Command struct {
    // Internal implementation
}

func NewOptions() *Options
func NewCommand(opts *Options) (*Command, error)
func (c *Command) Run(gifData []byte) ([]byte, error)
func (c *Command) Close() error
```

**Example: Batch GIF Conversion**

```go
func convertGIFs(files []string) error {
    opts := gif2webp.NewOptions()
    opts.Quality = 80
    opts.Method = 6

    cmd, err := gif2webp.NewCommand(opts)
    if err != nil {
        return err
    }
    defer cmd.Close()

    for _, file := range files {
        gifData, err := os.ReadFile(file)
        if err != nil {
            return err
        }

        webpData, err := cmd.Run(gifData)
        if err != nil {
            log.Printf("Failed to convert %s: %v", file, err)
            continue
        }

        outFile := strings.TrimSuffix(file, ".gif") + ".webp"
        if err := os.WriteFile(outFile, webpData, 0644); err != nil {
            return err
        }

        fmt.Printf("✓ %s: %d → %d bytes\n", file, len(gifData), len(webpData))
    }

    return nil
}
```

### WebP2GIF Package

```go
package webp2gif

type Options struct {
    Reserved int  // Reserved for future use
}

type Command struct {
    // Internal implementation
}

func NewOptions() *Options
func NewCommand(opts *Options) (*Command, error)
func (c *Command) Run(webpData []byte) ([]byte, error)
func (c *Command) Close() error
```

## Advanced Usage

### Reusing Encoder/Decoder Instances

For better performance when processing multiple images, reuse instances:

```go
func convertBatch(files []string) error {
    encoder, err := webp.NewEncoder(&webp.EncoderOptions{
        Quality: 80,
        Method:  4,
    })
    if err != nil {
        return err
    }
    defer encoder.Close()

    for _, file := range files {
        inputData, err := os.ReadFile(file)
        if err != nil {
            log.Printf("Failed to read %s: %v", file, err)
            continue
        }

        webpData, err := encoder.Encode(inputData)
        if err != nil {
            log.Printf("Failed to encode %s: %v", file, err)
            continue
        }

        outFile := strings.TrimSuffix(file, filepath.Ext(file)) + ".webp"
        if err := os.WriteFile(outFile, webpData, 0644); err != nil {
            log.Printf("Failed to write %s: %v", outFile, err)
            continue
        }

        fmt.Printf("✓ %s\n", file)
    }

    return nil
}
```

### Quality vs. File Size Trade-offs

```go
// High quality (larger files)
hqOpts := webp.NewEncoderOptions()
hqOpts.Quality = 95
hqOpts.Method = 6    // Slowest but best compression
hqOpts.Pass = 10     // More analysis passes

// Balanced (good for most use cases)
balancedOpts := webp.NewEncoderOptions()
balancedOpts.Quality = 80
balancedOpts.Method = 4

// Small files (lower quality)
smallOpts := webp.NewEncoderOptions()
smallOpts.Quality = 60
smallOpts.Method = 2
smallOpts.Preprocessing = 2
```

### Lossless Encoding

```go
opts := webp.NewEncoderOptions()
opts.Lossless = true
opts.Quality = 100
opts.Exact = true  // Preserve exact RGB values

encoder, err := webp.NewEncoder(opts)
if err != nil {
    return err
}
defer encoder.Close()

// Perfect quality, good for graphics/diagrams
webpData, err := encoder.Encode(inputData)
```

### Error Handling

```go
import (
    "errors"
    "github.com/ideamans/libnextimage/golang/webp"
    "github.com/ideamans/libnextimage/golang/types"
)

func encodeImage(data []byte) ([]byte, error) {
    encoder, err := webp.NewEncoder(&webp.EncoderOptions{
        Quality: 80,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create encoder: %w", err)
    }
    defer encoder.Close()

    webpData, err := encoder.Encode(data)
    if err != nil {
        var nextErr *types.NextImageError
        if errors.As(err, &nextErr) {
            switch nextErr.Status {
            case types.StatusInvalidParameter:
                return nil, fmt.Errorf("invalid encoding parameters: %w", err)
            case types.StatusOutOfMemory:
                return nil, fmt.Errorf("out of memory: %w", err)
            case types.StatusEncodeFailed:
                return nil, fmt.Errorf("encoding failed: %w", err)
            default:
                return nil, fmt.Errorf("encoding error: %w", err)
            }
        }
        return nil, err
    }

    return webpData, nil
}
```

### Concurrent Processing

Process multiple images concurrently using goroutines:

```go
func convertConcurrent(files []string, workers int) error {
    jobs := make(chan string, len(files))
    results := make(chan error, len(files))

    // Start workers
    for w := 0; w < workers; w++ {
        go func() {
            encoder, err := webp.NewEncoder(&webp.EncoderOptions{
                Quality: 80,
            })
            if err != nil {
                results <- err
                return
            }
            defer encoder.Close()

            for file := range jobs {
                inputData, err := os.ReadFile(file)
                if err != nil {
                    results <- fmt.Errorf("%s: %w", file, err)
                    continue
                }

                webpData, err := encoder.Encode(inputData)
                if err != nil {
                    results <- fmt.Errorf("%s: %w", file, err)
                    continue
                }

                outFile := strings.TrimSuffix(file, filepath.Ext(file)) + ".webp"
                if err := os.WriteFile(outFile, webpData, 0644); err != nil {
                    results <- fmt.Errorf("%s: %w", outFile, err)
                    continue
                }

                results <- nil
                fmt.Printf("✓ %s\n", file)
            }
        }()
    }

    // Send jobs
    for _, file := range files {
        jobs <- file
    }
    close(jobs)

    // Collect results
    var errs []error
    for range files {
        if err := <-results; err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("encountered %d errors during conversion", len(errs))
    }

    return nil
}
```

### Streaming from io.Reader

```go
func encodeFromReader(r io.Reader) ([]byte, error) {
    // Read all data into memory
    inputData, err := io.ReadAll(r)
    if err != nil {
        return nil, fmt.Errorf("failed to read input: %w", err)
    }

    encoder, err := webp.NewEncoder(&webp.EncoderOptions{
        Quality: 80,
    })
    if err != nil {
        return nil, err
    }
    defer encoder.Close()

    return encoder.Encode(inputData)
}

func encodeFile(inputPath, outputPath string) error {
    f, err := os.Open(inputPath)
    if err != nil {
        return err
    }
    defer f.Close()

    webpData, err := encodeFromReader(f)
    if err != nil {
        return err
    }

    return os.WriteFile(outputPath, webpData, 0644)
}
```

## Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| macOS    | ARM64 (M1/M2/M3) | ✅ |
| macOS    | x64          | ✅ |
| Linux    | x64          | ✅ |
| Linux    | ARM64        | ✅ |
| Windows  | x64          | ✅ |

## Performance Tips

1. **Reuse encoder/decoder instances** - Creating new instances has overhead
2. **Choose appropriate quality settings** - Higher quality = larger files + slower encoding
3. **Use appropriate method values** - Higher method = better compression but slower
4. **Consider lossless only when necessary** - Lossless produces much larger files
5. **Process in batches** - Reuse instances when processing multiple images
6. **Use goroutines for parallelism** - Distribute work across CPU cores

## Resource Management

All encoder/decoder/command instances use `runtime.SetFinalizer` for automatic cleanup when garbage collected. However, it's best practice to explicitly call `Close()`:

```go
encoder, err := webp.NewEncoder(opts)
if err != nil {
    return err
}
defer encoder.Close()  // Explicitly release resources

// Use encoder...
```

## Testing

Run the test suite:

```bash
cd golang
go test ./...
```

Run tests with race detector:

```bash
go test -race ./...
```

Run benchmarks:

```bash
go test -bench=. ./...
```

## Examples

See the `examples/golang/` directory for complete working examples:
- `jpeg-to-webp/` - JPEG to WebP conversion
- `jpeg-to-avif/` - JPEG to AVIF conversion
- `batch-convert/` - Batch conversion with progress
- `gif-conversion/` - GIF to WebP and WebP to GIF examples

## CGO Requirements

This package requires CGO to be enabled:

```bash
CGO_ENABLED=1 go build
```

The native library (libnextimage) must be available in your system's library path, or you can set the library path:

```bash
# macOS
export DYLD_LIBRARY_PATH=/path/to/libnextimage

# Linux
export LD_LIBRARY_PATH=/path/to/libnextimage

# Or set CGO flags
export CGO_LDFLAGS="-L/path/to/libnextimage"
```

## License

BSD-3-Clause

## Links

- GitHub: https://github.com/ideamans/libnextimage
- Issues: https://github.com/ideamans/libnextimage/issues
- Documentation: https://pkg.go.dev/github.com/ideamans/libnextimage/golang

## Contributing

Contributions are welcome! Please see the main repository for contribution guidelines.
