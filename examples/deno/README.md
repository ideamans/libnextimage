# Deno E2E Tests for libnextimage

End-to-end tests for libnextimage on Deno runtime.

## Status

🚧 **Work in Progress** - Basic Deno FFI implementation is available but not yet feature-complete.

Currently implemented:
- ✅ Library path resolution (Deno-specific)
- ✅ FFI bindings using Deno.dlopen()
- ✅ WebPEncoder (basic)

Not yet implemented:
- ⏳ WebPDecoder
- ⏳ AVIFEncoder
- ⏳ AVIFDecoder
- ⏳ Full options support

## Prerequisites

- Deno 1.x or later

## Installation

No installation required! Deno downloads dependencies automatically.

## Running Tests

```bash
# Basic encoding test
deno task test:basic

# Or run directly
deno run --allow-read --allow-write --allow-ffi --allow-env scripts/basic-encode.ts
```

### Required Permissions

- `--allow-read`: Read input files and library files
- `--allow-write`: Write output files
- `--allow-ffi`: Load native library via FFI
- `--allow-env`: Access environment variables (for platform detection)

## Project Structure

```
examples/deno/
├── deno.json           # Deno configuration and tasks
├── scripts/
│   └── basic-encode.ts # Basic WebP encoding test
└── output/             # Generated files
```

## Import Paths

### Local Development

```typescript
// Using import map in deno.json
import { WebPEncoder } from '@ideamans/libnextimage'
```

### From deno.land/x (Not yet published)

```typescript
// Will be available after publishing
import { WebPEncoder } from 'https://deno.land/x/libnextimage@v0.4.0/deno/mod.ts'
```

### From npm specifier

```typescript
// Note: Deno FFI may not work correctly with npm: specifier
// Use the direct import instead
import { WebPEncoder } from 'npm:@ideamans/libnextimage/deno'
```

## Example Usage

```typescript
import { WebPEncoder } from '@ideamans/libnextimage'

// Read JPEG file
const jpegData = await Deno.readFile('input.jpg')

// Create encoder
const encoder = new WebPEncoder({ quality: 80 })

// Encode to WebP
const webpData = encoder.encode(jpegData)

// Write output
await Deno.writeFile('output.webp', webpData)

// Clean up
encoder.close()

console.log(`Converted: ${jpegData.length} → ${webpData.length} bytes`)
```

## Differences from Node.js Version

1. **FFI API**: Uses `Deno.dlopen()` instead of Koffi
2. **File I/O**: Uses `Deno.readFile()` / `Deno.writeFile()` (async)
3. **Path Resolution**: Uses `import.meta.url` instead of `__dirname`
4. **Permissions**: Requires explicit `--allow-ffi` flag

## Troubleshooting

### "Cannot find libnextimage shared library"

Make sure you're running from the correct directory and the library is built:

```bash
# Build the C library first
cd ../../
make install-c

# Then run Deno tests
cd examples/deno
deno task test:basic
```

### "Permission denied"

Deno requires explicit permissions. Make sure to include all required flags:

```bash
deno run --allow-read --allow-write --allow-ffi --allow-env your-script.ts
```

### "Dynamic library load failed"

On macOS, you may need to allow the library to run:

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ../../lib/shared/libnextimage.dylib
```

## Development Status

This Deno implementation is a **proof of concept** demonstrating:
- ✅ Deno FFI can load and call libnextimage
- ✅ Platform-specific library resolution works
- ✅ Basic WebP encoding works

For production use, wait for:
- Full API implementation (decoders, AVIF support)
- Comprehensive testing
- Official release on deno.land/x

## Contributing

Help wanted! If you're experienced with Deno FFI, contributions are welcome:
- Implement remaining encoders/decoders
- Add full options support
- Create comprehensive tests
- Improve error handling

## Links

- [Deno FFI Documentation](https://deno.land/manual/runtime/ffi_api)
- [Main Repository](https://github.com/ideamans/libnextimage)
- [Node.js Version](../nodejs/)
