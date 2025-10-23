/**
 * WebP encode/decode tests
 * Tests WebP encoding and decoding functionality with the new API
 */

import { describe, it } from 'node:test';
import * as assert from 'node:assert';
import * as fs from 'fs';
import * as path from 'path';
import { encodeWebP, decodeWebP, PixelFormat, NextImageError } from '../src/index';

// Helper: Get test image path
function getTestImagePath(filename: string): string {
  return path.join(__dirname, '..', '..', '..', 'testdata', 'png', filename);
}

// Helper: Read test image
function readTestImage(filename: string): Buffer {
  const imagePath = getTestImagePath(filename);
  return fs.readFileSync(imagePath);
}

describe('WebP Encoding', () => {
  it('should encode PNG data to WebP', () => {
    const pngData = readTestImage('red.png');
    const webpData = encodeWebP(pngData, { quality: 80 });

    assert.ok(webpData, 'WebP data should not be null');
    assert.ok(webpData.length > 0, 'WebP data should not be empty');
    assert.ok(Buffer.isBuffer(webpData), 'Result should be a Buffer');

    // WebP files start with 'RIFF'
    assert.strictEqual(webpData.toString('ascii', 0, 4), 'RIFF', 'Should have RIFF header');
    // WebP signature at offset 8
    assert.strictEqual(webpData.toString('ascii', 8, 12), 'WEBP', 'Should have WEBP signature');
  });

  it('should encode with lossless option', () => {
    const pngData = readTestImage('blue.png');
    const webpData = encodeWebP(pngData, { lossless: true });

    assert.ok(webpData.length > 0, 'Lossless WebP data should not be empty');
  });

  it('should encode with different quality levels', () => {
    const pngData = readTestImage('red.png');

    const lowQuality = encodeWebP(pngData, { quality: 20 });
    const highQuality = encodeWebP(pngData, { quality: 95 });

    assert.ok(lowQuality.length > 0);
    assert.ok(highQuality.length > 0);
    // High quality should generally be larger
    assert.ok(highQuality.length > lowQuality.length);
  });

  it('should throw error for empty input', () => {
    const emptyBuffer = Buffer.alloc(0);

    assert.throws(() => {
      encodeWebP(emptyBuffer);
    }, NextImageError);
  });

  it('should throw error for invalid image data', () => {
    const invalidData = Buffer.from('not an image');

    assert.throws(() => {
      encodeWebP(invalidData);
    }, NextImageError);
  });
});

describe('WebP Decoding', () => {
  it('should decode WebP data to pixels', () => {
    // First encode a PNG to WebP
    const pngData = readTestImage('red.png');
    const webpData = encodeWebP(pngData, { quality: 100, lossless: true });

    // Then decode it
    const decoded = decodeWebP(webpData);

    assert.ok(decoded, 'Decoded image should not be null');
    assert.ok(decoded.width > 0, 'Width should be positive');
    assert.ok(decoded.height > 0, 'Height should be positive');
    assert.ok(decoded.data.length > 0, 'Pixel data should not be empty');
    assert.ok(Buffer.isBuffer(decoded.data), 'Pixel data should be a Buffer');
    assert.strictEqual(decoded.format, PixelFormat.RGBA, 'Format should be RGBA');
  });

  it('should decode with threading option', () => {
    const pngData = readTestImage('blue.png');
    const webpData = encodeWebP(pngData);

    const decoded = decodeWebP(webpData, { useThreads: true });

    assert.ok(decoded.width > 0);
    assert.ok(decoded.height > 0);
  });

  it('should throw error for empty input', () => {
    const emptyBuffer = Buffer.alloc(0);

    assert.throws(() => {
      decodeWebP(emptyBuffer);
    }, NextImageError);
  });

  it('should throw error for invalid WebP data', () => {
    const invalidData = Buffer.from('not a webp file');

    assert.throws(() => {
      decodeWebP(invalidData);
    }, NextImageError);
  });
});

describe('WebP Encode/Decode Round Trip', () => {
  it('should maintain image dimensions after round trip', () => {
    const pngData = readTestImage('red.png');

    // Encode
    const webpData = encodeWebP(pngData, { quality: 100, lossless: true });

    // Decode
    const decoded = decodeWebP(webpData);

    assert.ok(decoded.width > 0, 'Width should be preserved');
    assert.ok(decoded.height > 0, 'Height should be preserved');
  });

  it('should work with lossy encoding', () => {
    const pngData = readTestImage('blue.png');

    const webpData = encodeWebP(pngData, { quality: 75, lossless: false });
    const decoded = decodeWebP(webpData);

    assert.ok(decoded.width > 0);
    assert.ok(decoded.height > 0);
    assert.ok(decoded.data.length > 0);
  });
});
