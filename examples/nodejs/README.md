# Node.js E2E Tests for libnextimage

End-to-end tests that verify the published (or local) `@ideamans/libnextimage` package works correctly in a real Node.js environment.

## Prerequisites

- Node.js 18 or later
- npm

## Setup

```bash
cd examples/nodejs
npm install
```

This will install the `@ideamans/libnextimage` package. During development, it uses the local version via `file:../../typescript`.

## Running Tests

### All Tests

```bash
npm test
```

This runs all test scripts in sequence:
1. Basic encoding (WebP and AVIF)
2. Batch processing (multiple files)
3. Decode tests (round-trip encoding/decoding)
4. All features (quality levels, lossless, metadata)

### Individual Tests

```bash
# Basic encoding test
npm run test:basic

# Batch processing test
npm run test:batch

# Decode test
npm run test:decode

# All features test
npm run test:all
```

## Test Scripts

### `scripts/basic-encode.js`

Tests basic WebP and AVIF encoding:
- Reads a test JPEG image
- Encodes to WebP (quality 80)
- Encodes to AVIF (quality 60, speed 6)
- Reports compression ratios
- Saves output files

**Expected output:**
```
=== libnextimage Node.js E2E Test: Basic Encode ===

Library version: 0.4.0

Input: 123456 bytes (120.56 KB)

--- WebP Encoding ---
✓ WebP: 89012 bytes (86.93 KB)
  Compression: 27.9% smaller

--- AVIF Encoding ---
✓ AVIF: 76543 bytes (74.75 KB)
  Compression: 38.0% smaller

✅ All basic encoding tests passed!
```

### `scripts/batch-process.js`

Tests batch processing with encoder reuse:
- Finds all JPEG files in testdata/
- Creates a single WebPEncoder instance
- Converts all files to WebP
- Reports individual and total compression

**Expected output:**
```
=== libnextimage Node.js E2E Test: Batch Processing ===

Found 5 JPEG files

Converting...

✓ landscape-like.jpg
  120.56 KB → 86.93 KB (27.9% smaller)
✓ gradient-horizontal.jpg
  45.23 KB → 32.10 KB (29.0% smaller)
...

============================================================
Batch conversion complete!
  Files processed: 5
  Total input: 0.45 MB
  Total output: 0.32 MB
  Overall compression: 28.5%
============================================================

✅ Batch processing test passed!
```

### `scripts/decode-test.js`

Tests encoding and decoding round-trips:
- Encodes JPEG to WebP, then decodes back to raw pixels
- Encodes JPEG to AVIF, then decodes back to raw pixels
- Verifies dimensions and data sizes match

**Expected output:**
```
=== libnextimage Node.js E2E Test: Decode ===

Input: 123456 bytes

--- WebP Round-trip Test ---
✓ Encoded to WebP: 89012 bytes
✓ Decoded from WebP: 800x600, 1920000 bytes
  Format: RGBA
  Expected RGBA size: 1920000 bytes

--- AVIF Round-trip Test ---
✓ Encoded to AVIF: 76543 bytes
✓ Decoded from AVIF: 800x600, 1920000 bytes
  Format: RGBA
  Expected RGBA size: 1920000 bytes

--- Verification ---
✓ Dimensions match: 800x600
✓ Data size matches: 1920000 bytes

✅ All decode tests passed!
```

### `scripts/all-features.js`

Tests various encoding options:
- WebP with different quality levels (50, 75, 90)
- WebP lossless mode
- AVIF with different quality/speed combinations
- Metadata preservation

**Expected output:**
```
=== libnextimage Node.js E2E Test: All Features ===

Input: 123456 bytes

--- WebP: Quality Levels ---
✓ Quality 50: 65.32 KB (47.1% smaller)
✓ Quality 75: 86.93 KB (27.9% smaller)
✓ Quality 90: 108.45 KB (9.8% smaller)

--- WebP: Lossless Mode ---
✓ Lossless: 145.67 KB

--- AVIF: Quality/Speed Combinations ---
✓ Fast/Low (q=50, s=8): 58.23 KB (51.6% smaller)
✓ Balanced (q=60, s=6): 74.75 KB (38.0% smaller)
✓ Quality (q=70, s=4): 89.12 KB (26.0% smaller)

--- WebP: Metadata Preservation ---
✓ With metadata: 87.45 KB

============================================================
Feature test complete!

Generated files in output/:
  - quality-*.webp (different quality levels)
  - lossless.webp (lossless mode)
  - avif-*.avif (different speed/quality)
  - with-metadata.webp (metadata preservation)
============================================================

✅ All feature tests passed!
```

## Output Files

All generated files are saved to the `output/` directory:

```
output/
├── test-basic.webp          # Basic WebP encoding
├── test-basic.avif          # Basic AVIF encoding
├── quality-50.webp          # Low quality
├── quality-75.webp          # Medium quality
├── quality-90.webp          # High quality
├── lossless.webp            # Lossless WebP
├── avif-fast-low.avif       # Fast AVIF encoding
├── avif-balanced.avif       # Balanced AVIF
├── avif-quality.avif        # High quality AVIF
├── with-metadata.webp       # WebP with metadata
└── *.webp                   # Batch converted files
```

## Troubleshooting

### "Test image not found"

The tests use images from `testdata/jpeg-source/` in the repository root. Make sure you're running from the correct directory:

```bash
cd examples/nodejs
npm test
```

### "Cannot find module '@ideamans/libnextimage'"

Install dependencies first:

```bash
npm install
```

### "Cannot find libnextimage shared library"

The native library wasn't downloaded. Try reinstalling:

```bash
rm -rf node_modules
npm install
```

## Using with Published Package

To test with the published npm package instead of the local version:

1. Edit `package.json`:
   ```json
   {
     "dependencies": {
       "@ideamans/libnextimage": "^0.4.0"
     }
   }
   ```

2. Reinstall:
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   npm test
   ```

## CI/CD Integration

These tests can be run in GitHub Actions:

```yaml
- name: Run Node.js E2E tests
  working-directory: examples/nodejs
  run: |
    npm install
    npm test
```

## Performance Notes

- Creating an encoder/decoder instance has some overhead
- For batch processing, **reuse encoder instances** (see `batch-process.js`)
- Always call `close()` to free native resources
- AVIF encoding is slower than WebP (adjust `speed` parameter as needed)

## Next Steps

- Try modifying encoding options in the scripts
- Add your own test images to `input/`
- Compare file sizes and quality
- Measure encoding times for your use case
