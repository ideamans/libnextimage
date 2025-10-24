/**
 * AVIF encode/decode tests with class-based API
 */

import { describe, it } from 'node:test';
import * as assert from 'node:assert';
import * as fs from 'fs';
import * as path from 'path';
import { AVIFEncoder, AVIFDecoder, PixelFormat, NextImageError } from '../src/index';

function getTestImagePath(filename: string): string {
  return path.join(__dirname, '..', '..', '..', 'testdata', 'png', filename);
}

function readTestImage(filename: string): Buffer {
  return fs.readFileSync(getTestImagePath(filename));
}

describe('AVIFEncoder', () => {
  it('should encode PNG to AVIF', () => {
    const pngData = readTestImage('red.png');
    const encoder = new AVIFEncoder({ quality: 60 });
    const avifData = encoder.encode(pngData);
    encoder.close();

    assert.ok(avifData.length > 0);
  });

  it('should reuse encoder for multiple images', () => {
    const encoder = new AVIFEncoder({ quality: 65, speed: 8 });
    const avif1 = encoder.encode(readTestImage('red.png'));
    const avif2 = encoder.encode(readTestImage('blue.png'));
    encoder.close();

    assert.ok(avif1.length > 0);
    assert.ok(avif2.length > 0);
  });

  it('should throw on empty input', () => {
    const encoder = new AVIFEncoder();
    assert.throws(() => encoder.encode(Buffer.alloc(0)), NextImageError);
    encoder.close();
  });
});

describe('AVIFDecoder', () => {
  it('should decode AVIF to pixels', () => {
    const encoder = new AVIFEncoder({ quality: 90 });
    const avifData = encoder.encode(readTestImage('red.png'));
    encoder.close();

    const decoder = new AVIFDecoder();
    const decoded = decoder.decode(avifData);
    decoder.close();

    assert.ok(decoded.width > 0);
    assert.ok(decoded.height > 0);
    assert.ok(decoded.data.length > 0);
  });
});
