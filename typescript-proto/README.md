# libnextimage TypeScript Bindings

TypeScript/Node.js bindings for libnextimage - high-performance WebP and AVIF image processing.

## Development Setup

### 1. Build the C library

From the project root:

```bash
# Build both static and shared libraries
make install-c

# This will create:
# - lib/shared/libnextimage.dylib (macOS)
# - lib/shared/libnextimage.so (Linux)
# - lib/shared/libnextimage.dll (Windows)
# - lib/include/*.h (headers)
```

### 2. Install TypeScript dependencies

```bash
cd typescript
npm install
```

### 3. Build TypeScript code

```bash
npm run build
```

### 4. Test library loading

```bash
node -e "console.log(require('./dist/index').getLibraryPath())"
```

Expected output:
```
/path/to/libnextimage/lib/shared/libnextimage.dylib
```

## Project Structure

```
typescript/
├── src/
│   ├── index.ts           # Main entry point
│   └── library.ts         # Library path resolution
├── dist/                  # Compiled JavaScript (generated)
├── package.json
├── tsconfig.json
└── README.md
```

## Library Path Resolution

The TypeScript bindings automatically locate the shared library in development mode:

1. **Development mode** (recommended): `../lib/<platform>/libnextimage.{so,dylib,dll}`
   - Relative to project root
   - Automatically uses the library built by `make install-c`

2. **Installed package**: `./lib/<platform>/libnextimage.{so,dylib,dll}`
   - For published npm packages
   - Includes pre-built binaries

## Usage Example

```typescript
import * as fs from 'fs';
import { encodeWebP, encodeWebPWithQuality } from '@ideamans/libnextimage';

// Read JPEG file
const jpegData = fs.readFileSync('input.jpg');

// Convert to WebP (default quality: 75)
const webpData = encodeWebP(jpegData);
fs.writeFileSync('output.webp', webpData);

// Convert with specific quality
const webpHighQuality = encodeWebPWithQuality(jpegData, 90);
fs.writeFileSync('output-q90.webp', webpHighQuality);

console.log(`JPEG: ${jpegData.length} bytes → WebP: ${webpData.length} bytes`);
```

**Platform detection:**

```typescript
import { getLibraryPath, getPlatform } from '@ideamans/libnextimage';

console.log(`Platform: ${getPlatform()}`); // darwin-arm64, linux-amd64, etc.
console.log(`Library: ${getLibraryPath()}`); // /path/to/libnextimage.dylib
```

## Development Workflow

1. Make changes to C library
2. Rebuild: `make install-c` (from project root)
3. TypeScript code automatically picks up the new library
4. No need to copy files manually!

## Running Tests

```bash
# Run all tests
npm test

# This will:
# 1. Build TypeScript code (tsc)
# 2. Run tests using Node.js test runner
# 3. Convert JPEG images to WebP
# 4. Save outputs to test-output/ directory
```

**Test coverage:**
- Basic JPEG to WebP conversion
- Quality settings (default, 90)
- Lossless encoding
- Multiple file processing

**Test output:**
- `test-output/*.webp` - Generated WebP files for manual inspection

## Next Steps

- [ ] Implement FFI bindings for WebP encode/decode
- [ ] Implement FFI bindings for AVIF encode/decode
- [ ] Add TypeScript type definitions for options
- [ ] Add comprehensive tests
- [ ] Add usage examples

## Notes

- The library path resolution is platform-aware (macOS/Linux/Windows)
- Development mode assumes you're working from the `typescript/` directory
- The shared library is automatically located - no configuration needed!
