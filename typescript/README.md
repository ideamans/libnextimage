# @ideamans/libnextimage

High-performance WebP and AVIF image processing library for Node.js, with TypeScript support.

## Features

- ✅ **WebP encoding/decoding** - Full WebP support with all encoding options
- ✅ **AVIF encoding/decoding** - Next-generation AVIF format support
- ✅ **Zero native compilation** - Pre-built binaries downloaded automatically
- ✅ **TypeScript native** - Full type definitions included
- ✅ **Cross-platform** - macOS (ARM64/Intel), Linux (ARM64/x64), Windows (x64)
- ✅ **High performance** - Direct FFI bindings with minimal overhead
- ✅ **Production ready** - Memory-safe with automatic resource cleanup

## Installation

```bash
npm install @ideamans/libnextimage
```

The package automatically downloads the appropriate pre-built native library for your platform during installation. No compilation required!

### Supported Platforms

- macOS (Apple Silicon M1/M2/M3 and Intel)
- Linux (ARM64 and x64)
- Windows (x64)

## Quick Start

### WebP Encoding

```typescript
import { WebPEncoder } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

// Create encoder with options
const encoder = new WebPEncoder({
  quality: 80,
  method: 6
})

// Encode JPEG to WebP
const jpegData = readFileSync('input.jpg')
const webpData = encoder.encode(jpegData)
writeFileSync('output.webp', webpData)

// Clean up resources
encoder.close()

console.log(`Converted: ${jpegData.length} bytes → ${webpData.length} bytes`)
```

### AVIF Encoding

```typescript
import { AVIFEncoder } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

// Create encoder with options
const encoder = new AVIFEncoder({
  quality: 60,
  speed: 6
})

// Encode JPEG to AVIF
const jpegData = readFileSync('input.jpg')
const avifData = encoder.encode(jpegData)
writeFileSync('output.avif', avifData)

// Clean up resources
encoder.close()

console.log(`Converted: ${jpegData.length} bytes → ${avifData.length} bytes`)
```

### WebP Decoding

```typescript
import { WebPDecoder } from '@ideamans/libnextimage'
import { readFileSync } from 'fs'

const decoder = new WebPDecoder({
  format: 'RGBA'
})

const webpData = readFileSync('input.webp')
const decoded = decoder.decode(webpData)

console.log(`Decoded: ${decoded.width}x${decoded.height}, ${decoded.data.length} bytes`)

decoder.close()
```

### AVIF Decoding

```typescript
import { AVIFDecoder } from '@ideamans/libnextimage'
import { readFileSync } from 'fs'

const decoder = new AVIFDecoder({
  format: 'RGBA'
})

const avifData = readFileSync('input.avif')
const decoded = decoder.decode(avifData)

console.log(`Decoded: ${decoded.width}x${decoded.height}, ${decoded.data.length} bytes`)

decoder.close()
```

## API Reference

### WebPEncoder

#### Constructor Options

```typescript
interface WebPEncodeOptions {
  quality?: number          // 0-100, default: 75
  lossless?: boolean        // default: false
  method?: number           // 0-6, default: 4 (quality/speed tradeoff)
  preset?: WebPPreset       // 'default', 'picture', 'photo', 'drawing', 'icon', 'text'

  // Advanced options
  targetSize?: number       // Target file size in bytes
  targetPSNR?: number       // Target PSNR
  segments?: number         // 1-4, number of segments
  snsStrength?: number      // 0-100, spatial noise shaping
  filterStrength?: number   // 0-100, filter strength
  autofilter?: boolean      // Auto-adjust filter settings

  // Alpha channel
  alphaQuality?: number     // 0-100, alpha compression quality

  // Metadata
  keepMetadata?: number     // MetadataEXIF | MetadataICC | MetadataXMP | MetadataAll

  // Transform
  cropX?: number
  cropY?: number
  cropWidth?: number
  cropHeight?: number
  resizeWidth?: number
  resizeHeight?: number
}
```

#### Methods

```typescript
class WebPEncoder {
  constructor(options?: Partial<WebPEncodeOptions>)

  encode(data: Buffer): Buffer
  encodeFile(path: string): Buffer

  close(): void

  static getDefaultOptions(): WebPEncodeOptions
}
```

### AVIFEncoder

#### Constructor Options

```typescript
interface AVIFEncodeOptions {
  quality?: number          // 0-100, default: 60
  qualityAlpha?: number     // 0-100, default: -1 (use quality)
  speed?: number            // 0-10, default: 6 (0=slowest/best, 10=fastest/worst)

  bitDepth?: number         // 8, 10, or 12 (default: 8)
  yuvFormat?: AVIFYUVFormat // 'YUV444', 'YUV422', 'YUV420', 'YUV400'

  // Advanced options
  lossless?: boolean
  sharpYUV?: boolean
  targetSize?: number

  // Threading
  jobs?: number             // -1=all cores, 0=auto, >0=thread count

  // Tiling
  tileRowsLog2?: number     // 0-6
  tileColsLog2?: number     // 0-6
  autoTiling?: boolean

  // Metadata
  exifData?: Buffer
  xmpData?: Buffer
  iccData?: Buffer
}
```

#### Methods

```typescript
class AVIFEncoder {
  constructor(options?: Partial<AVIFEncodeOptions>)

  encode(data: Buffer): Buffer
  encodeFile(path: string): Buffer

  close(): void

  static getDefaultOptions(): AVIFEncodeOptions
}
```

### WebPDecoder

```typescript
interface WebPDecodeOptions {
  format?: PixelFormat      // 'RGBA', 'BGRA', 'RGB', 'BGR'
  useThreads?: boolean
  bypassFiltering?: boolean
  noFancyUpsampling?: boolean

  cropX?: number
  cropY?: number
  cropWidth?: number
  cropHeight?: number

  scaleWidth?: number
  scaleHeight?: number
}

class WebPDecoder {
  constructor(options?: Partial<WebPDecodeOptions>)

  decode(data: Buffer): DecodedImage
  decodeFile(path: string): DecodedImage

  close(): void

  static getDefaultOptions(): WebPDecodeOptions
}

interface DecodedImage {
  width: number
  height: number
  data: Buffer
  format: PixelFormat
}
```

### AVIFDecoder

```typescript
interface AVIFDecodeOptions {
  format?: PixelFormat      // 'RGBA', 'BGRA', 'RGB', 'BGR'
  jobs?: number             // -1=all cores, 0=auto, >0=thread count

  chromaUpsampling?: ChromaUpsampling

  ignoreExif?: boolean
  ignoreXMP?: boolean
  ignoreICC?: boolean

  imageSizeLimit?: number
  imageDimensionLimit?: number
}

class AVIFDecoder {
  constructor(options?: Partial<AVIFDecodeOptions>)

  decode(data: Buffer): DecodedImage
  decodeFile(path: string): DecodedImage

  close(): void

  static getDefaultOptions(): AVIFDecodeOptions
}
```

## Batch Processing Example

```typescript
import { WebPEncoder } from '@ideamans/libnextimage'
import { readdirSync, readFileSync, writeFileSync } from 'fs'
import { join } from 'path'

const encoder = new WebPEncoder({ quality: 80 })

const files = readdirSync('images')
  .filter(f => f.endsWith('.jpg') || f.endsWith('.png'))

for (const file of files) {
  const inputPath = join('images', file)
  const outputPath = join('output', file.replace(/\.(jpg|png)$/, '.webp'))

  const inputData = readFileSync(inputPath)
  const webpData = encoder.encode(inputData)
  writeFileSync(outputPath, webpData)

  console.log(`✓ ${file}: ${inputData.length} → ${webpData.length} bytes`)
}

encoder.close()
```

## Memory Management

**Important:** Always call `close()` when you're done with an encoder or decoder to free native resources.

```typescript
// Good: Manual cleanup
const encoder = new WebPEncoder({ quality: 80 })
try {
  const result = encoder.encode(data)
  // ... use result
} finally {
  encoder.close()
}

// Good: Reuse encoder for multiple files
const encoder = new WebPEncoder({ quality: 80 })
for (const file of files) {
  const result = encoder.encode(readFileSync(file))
  // ... process result
}
encoder.close()
```

## Version Management

This package uses a dual-version system:

- **Package version** (in package.json): NPM package version
- **Native library version** (in library-version.json): Pre-built library version

This allows patch releases for TypeScript code without rebuilding native libraries.

```typescript
import { getLibraryVersion } from '@ideamans/libnextimage'

console.log(getLibraryVersion()) // e.g., "0.4.0"
```

## Troubleshooting

### "Cannot find libnextimage shared library"

The native library wasn't downloaded during installation.

**Solution:**
```bash
npm install --force @ideamans/libnextimage
```

### "Unsupported platform"

Your platform isn't supported yet. Supported platforms:
- macOS (ARM64, x64)
- Linux (ARM64, x64)
- Windows (x64)

**Solution:** Build from source (see main repository README)

### Memory Issues

If you're processing many images, make sure to:
1. Reuse encoders/decoders instead of creating new ones
2. Call `close()` when done
3. Process images in batches if needed

## Examples

See the [examples/typescript/](../examples/typescript/) directory for complete working examples:

- `jpeg-to-webp.ts` - JPEG to WebP conversion
- `jpeg-to-avif.ts` - JPEG to AVIF conversion
- `batch-convert.ts` - Batch conversion with progress

## Runtime Support

Currently supported:
- ✅ **Node.js** 18+ (Full support)

Planned:
- 🔄 **Deno** (Coming soon)
- 🔄 **Bun** (Coming soon)

## License

BSD-3-Clause

## Links

- [GitHub Repository](https://github.com/ideamans/libnextimage)
- [Examples](../examples/typescript/)
- [Version Management Guide](./VERSION-MANAGEMENT.md)
- [Issue Tracker](https://github.com/ideamans/libnextimage/issues)

## Contributing

Contributions are welcome! Please see the main repository for contribution guidelines.
