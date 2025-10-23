/**
 * WebP encoding test - JPEG to WebP conversion
 *
 * Usage:
 *   npm test
 */

import * as fs from 'fs';
import * as path from 'path';
import * as assert from 'assert';
import { test } from 'node:test';
import { encodeWebP, encodeWebPWithQuality, encodeWebPLossless } from '../src/webp';

// Test data paths (relative to project root)
// __dirname in compiled code will be dist/test/, so we need to go up 3 levels
const TEST_DATA_DIR = path.join(__dirname, '..', '..', '..', 'testdata', 'jpeg-source');
const OUTPUT_DIR = path.join(__dirname, '..', '..', 'test-output');

// Ensure output directory exists
if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

test('WebP Encoding - Basic JPEG to WebP conversion', async () => {
  const inputPath = path.join(TEST_DATA_DIR, 'landscape-like.jpg');
  const outputPath = path.join(OUTPUT_DIR, 'landscape-like.webp');

  // Read JPEG file
  const jpegData = fs.readFileSync(inputPath);
  console.log(`  Input: ${inputPath} (${jpegData.length} bytes)`);

  // Encode to WebP (default quality)
  const webpData = encodeWebP(jpegData);
  console.log(`  Output: ${outputPath} (${webpData.length} bytes)`);

  // Verify output
  assert.ok(webpData.length > 0, 'WebP data should not be empty');
  assert.ok(webpData.length < jpegData.length, 'WebP should be smaller than JPEG');

  // Check WebP magic bytes (RIFF....WEBP)
  assert.strictEqual(webpData.toString('ascii', 0, 4), 'RIFF', 'Should start with RIFF');
  assert.strictEqual(webpData.toString('ascii', 8, 12), 'WEBP', 'Should contain WEBP signature');

  // Save output for manual inspection
  fs.writeFileSync(outputPath, webpData);
  console.log(`  ✓ Saved: ${outputPath}`);
});

test('WebP Encoding - Quality 90', async () => {
  const inputPath = path.join(TEST_DATA_DIR, 'gradient-horizontal.jpg');
  const outputPath = path.join(OUTPUT_DIR, 'gradient-horizontal-q90.webp');

  const jpegData = fs.readFileSync(inputPath);
  console.log(`  Input: ${inputPath} (${jpegData.length} bytes)`);

  // Encode with quality 90
  const webpData = encodeWebPWithQuality(jpegData, 90);
  console.log(`  Output: ${outputPath} (${webpData.length} bytes, quality=90)`);

  assert.ok(webpData.length > 0, 'WebP data should not be empty');
  assert.strictEqual(webpData.toString('ascii', 0, 4), 'RIFF', 'Should be valid WebP');

  fs.writeFileSync(outputPath, webpData);
  console.log(`  ✓ Saved: ${outputPath}`);
});

test('WebP Encoding - Lossless mode', async () => {
  const inputPath = path.join(TEST_DATA_DIR, 'solid-black.jpg');
  const outputPath = path.join(OUTPUT_DIR, 'solid-black-lossless.webp');

  const jpegData = fs.readFileSync(inputPath);
  console.log(`  Input: ${inputPath} (${jpegData.length} bytes)`);

  // Encode lossless
  const webpData = encodeWebPLossless(jpegData);
  console.log(`  Output: ${outputPath} (${webpData.length} bytes, lossless)`);

  assert.ok(webpData.length > 0, 'WebP data should not be empty');
  assert.strictEqual(webpData.toString('ascii', 0, 4), 'RIFF', 'Should be valid WebP');

  fs.writeFileSync(outputPath, webpData);
  console.log(`  ✓ Saved: ${outputPath}`);
});

test('WebP Encoding - Multiple files', async () => {
  const testFiles = ['edges.jpg', 'gradient-radial.jpg'];

  for (const filename of testFiles) {
    const inputPath = path.join(TEST_DATA_DIR, filename);
    const outputFilename = filename.replace('.jpg', '.webp');
    const outputPath = path.join(OUTPUT_DIR, outputFilename);

    const jpegData = fs.readFileSync(inputPath);
    const webpData = encodeWebP(jpegData);

    assert.ok(webpData.length > 0, `WebP data for ${filename} should not be empty`);
    assert.strictEqual(webpData.toString('ascii', 0, 4), 'RIFF', `${filename} should be valid WebP`);

    fs.writeFileSync(outputPath, webpData);
    console.log(`  ✓ ${filename} → ${outputFilename} (${jpegData.length} → ${webpData.length} bytes)`);
  }
});
