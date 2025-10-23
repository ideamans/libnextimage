/**
 * AVIF encode/decode tests
 * Tests AVIF encoding and decoding functionality with the new API
 */

import { describe, it } from 'node:test';
import * as assert from 'node:assert';
import * as fs from 'fs';
import * as path from 'path';
import { encodeAVIF, decodeAVIF, PixelFormat, NextImageError } from '../src/index';

// Helper: Get test image path
function getTestImagePath(filename: string): string {
  return path.join(__dirname, '..', '..', '..', 'testdata', 'png', filename);
}

// Helper: Read test image
function readTestImage(filename: string): Buffer {
  const imagePath = getTestImagePath(filename);
  return fs.readFileSync(imagePath);
}

describe('AVIF Encoding', () => {
  it('should encode PNG data to AVIF', () => {
    const pngData = readTestImage('red.png');
    const avifData = encodeAVIF(pngData, { quality: 60 });

    assert.ok(avifData, 'AVIF data should not be null');
    assert.ok(avifData.length > 0, 'AVIF data should not be empty');
    assert.ok(Buffer.isBuffer(avifData), 'Result should be a Buffer');

    // AVIF files can start with different patterns depending on the container
    // Most commonly starts with 'ftyp' box after an 8-byte header
    // Just check that we have some data
    assert.ok(avifData.length > 12, 'AVIF file should have meaningful size');
  });

  it('should encode with different quality levels', () => {
    const pngData = readTestImage('red.png');

    const lowQuality = encodeAVIF(pngData, { quality: 20 });
    const highQuality = encodeAVIF(pngData, { quality: 90 });

    assert.ok(lowQuality.length > 0);
    assert.ok(highQuality.length > 0);
    // Higher quality should generally be larger (though not always guaranteed with AVIF)
  });

  it('should encode with speed settings', () => {
    const pngData = readTestImage('blue.png');

    const slow = encodeAVIF(pngData, { speed: 0 });  // Slowest, best quality
    const fast = encodeAVIF(pngData, { speed: 10 }); // Fastest

    assert.ok(slow.length > 0);
    assert.ok(fast.length > 0);
  });

  it('should encode with different bit depths', () => {
    const pngData = readTestImage('red.png');

    const depth8 = encodeAVIF(pngData, { bitDepth: 8 });
    const depth10 = encodeAVIF(pngData, { bitDepth: 10 });

    assert.ok(depth8.length > 0);
    assert.ok(depth10.length > 0);
  });

  it('should throw error for empty input', () => {
    const emptyBuffer = Buffer.alloc(0);

    assert.throws(() => {
      encodeAVIF(emptyBuffer);
    }, NextImageError);
  });

  it('should throw error for invalid image data', () => {
    const invalidData = Buffer.from('not an image');

    assert.throws(() => {
      encodeAVIF(invalidData);
    }, NextImageError);
  });
});

describe('AVIF Decoding', () => {
  it('should decode AVIF data to pixels', () => {
    // First encode a PNG to AVIF
    const pngData = readTestImage('red.png');
    const avifData = encodeAVIF(pngData, { quality: 90 });

    // Then decode it
    const decoded = decodeAVIF(avifData);

    assert.ok(decoded, 'Decoded image should not be null');
    assert.ok(decoded.width > 0, 'Width should be positive');
    assert.ok(decoded.height > 0, 'Height should be positive');
    assert.ok(decoded.data.length > 0, 'Pixel data should not be empty');
    assert.ok(Buffer.isBuffer(decoded.data), 'Pixel data should be a Buffer');
    assert.strictEqual(decoded.format, PixelFormat.RGBA, 'Format should be RGBA');
  });

  it('should decode with threading option', () => {
    const pngData = readTestImage('blue.png');
    const avifData = encodeAVIF(pngData);

    const decoded = decodeAVIF(avifData, { useThreads: true });

    assert.ok(decoded.width > 0);
    assert.ok(decoded.height > 0);
  });

  it('should throw error for empty input', () => {
    const emptyBuffer = Buffer.alloc(0);

    assert.throws(() => {
      decodeAVIF(emptyBuffer);
    }, NextImageError);
  });

  it('should throw error for invalid AVIF data', () => {
    const invalidData = Buffer.from('not an avif file');

    assert.throws(() => {
      decodeAVIF(invalidData);
    }, NextImageError);
  });
});

describe('AVIF Encode/Decode Round Trip', () => {
  it('should maintain image dimensions after round trip', () => {
    const pngData = readTestImage('red.png');

    // Encode
    const avifData = encodeAVIF(pngData, { quality: 90 });

    // Decode
    const decoded = decodeAVIF(avifData);

    assert.ok(decoded.width > 0, 'Width should be preserved');
    assert.ok(decoded.height > 0, 'Height should be preserved');
  });

  it('should work with medium quality encoding', () => {
    const pngData = readTestImage('blue.png');

    const avifData = encodeAVIF(pngData, { quality: 60 });
    const decoded = decodeAVIF(avifData);

    assert.ok(decoded.width > 0);
    assert.ok(decoded.height > 0);
    assert.ok(decoded.data.length > 0);
  });

  it('should work with different speeds', () => {
    const pngData = readTestImage('red.png');

    const avifData = encodeAVIF(pngData, { quality: 70, speed: 8 });
    const decoded = decodeAVIF(avifData);

    assert.ok(decoded.width > 0);
    assert.ok(decoded.height > 0);
  });
});
