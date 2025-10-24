# Bun E2E Tests for libnextimage

End-to-end tests for libnextimage on Bun runtime.

## Status

🚧 **Work in Progress** - Basic Bun FFI implementation is available but not yet feature-complete.

Currently implemented:
- ✅ Library path resolution (Bun-specific)
- ✅ FFI bindings using bun:ffi
- ✅ WebPEncoder (basic)

Not yet implemented:
- ⏳ WebPDecoder
- ⏳ AVIFEncoder
- ⏳ AVIFDecoder
- ⏳ Full options support

## Prerequisites

- Bun 1.x or later

## Installation

```bash
cd examples/bun
bun install
```

## Running Tests

```bash
# Basic encoding test
bun test:basic

# Or run directly
bun run scripts/basic-encode.ts
```

## Project Structure

```
examples/bun/
├── package.json        # Package configuration
├── scripts/
│   └── basic-encode.ts # Basic WebP encoding test
└── output/             # Generated files
```

## Example Usage

```typescript
import { WebPEncoder } from '@ideamans/libnextimage/bun'
import { readFileSync, writeFileSync } from 'fs'

// Read JPEG file
const jpegData = readFileSync('input.jpg')

// Create encoder
const encoder = new WebPEncoder({ quality: 80 })

// Encode to WebP
const webpData = encoder.encode(jpegData)

// Write output
writeFileSync('output.webp', webpData)

// Clean up
encoder.close()

console.log(`Converted: ${jpegData.length} → ${webpData.length} bytes`)
```

## Differences from Node.js Version

1. **FFI API**: Uses `bun:ffi` instead of Koffi
2. **Performance**: Bun's FFI may offer better performance
3. **File I/O**: Uses standard Node.js API (fs module)
4. **Import**: Uses `@ideamans/libnextimage/bun` subpath

## Benchmarking

Bun is optimized for performance. You can benchmark encoding speed:

```typescript
import { WebPEncoder } from '@ideamans/libnextimage/bun'
import { readFileSync } from 'fs'

const jpegData = readFileSync('input.jpg')
const encoder = new WebPEncoder({ quality: 80 })

const iterations = 100
const start = Bun.nanoseconds()

for (let i = 0; i < iterations; i++) {
  encoder.encode(jpegData)
}

const elapsed = (Bun.nanoseconds() - start) / 1_000_000
console.log(`Average: ${(elapsed / iterations).toFixed(2)} ms/image`)

encoder.close()
```

## Troubleshooting

### "Cannot find libnextimage shared library"

Make sure you're running from the correct directory and the library is built:

```bash
# Build the C library first
cd ../../
make install-c

# Then run Bun tests
cd examples/bun
bun test:basic
```

### "Dynamic library load failed"

On macOS, you may need to allow the library to run:

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ../../lib/shared/libnextimage.dylib
```

### Type Errors

Bun's TypeScript support is built-in, but you may need to configure `tsconfig.json` for FFI types:

```json
{
  "compilerOptions": {
    "types": ["bun-types"]
  }
}
```

## Development Status

This Bun implementation is a **proof of concept** demonstrating:
- ✅ Bun FFI can load and call libnextimage
- ✅ Platform-specific library resolution works
- ✅ Basic WebP encoding works
- ✅ Compatible with npm package structure

For production use, wait for:
- Full API implementation (decoders, AVIF support)
- Comprehensive testing
- Performance benchmarks vs Node.js

## Performance Notes

Bun's FFI is designed for high performance:
- Zero-copy where possible
- Native TypedArray support
- Fast C interop

Expected benefits over Node.js (Koffi):
- Lower FFI overhead
- Faster startup time
- Better memory efficiency

*Actual benchmarks coming soon*

## Contributing

Help wanted! If you're experienced with Bun FFI, contributions are welcome:
- Implement remaining encoders/decoders
- Add full options support
- Create performance benchmarks
- Improve FFI memory handling

## Links

- [Bun FFI Documentation](https://bun.sh/docs/api/ffi)
- [Main Repository](https://github.com/ideamans/libnextimage)
- [Node.js Version](../nodejs/)
- [Deno Version](../deno/)
