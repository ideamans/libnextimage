/**
 * WebP encode/decode tests with class-based API
 */

import { describe, it } from 'node:test';
import * as assert from 'node:assert';
import * as fs from 'fs';
import * as path from 'path';
import { WebPEncoder, WebPDecoder, PixelFormat, NextImageError } from '../src/index';

function getTestImagePath(filename: string): string {
  return path.join(__dirname, '..', '..', '..', 'testdata', 'png', filename);
}

function readTestImage(filename: string): Buffer {
  return fs.readFileSync(getTestImagePath(filename));
}

describe('WebPEncoder', () => {
  it('should encode PNG to WebP', () => {
    const pngData = readTestImage('red.png');
    const encoder = new WebPEncoder({ quality: 80 });
    const webpData = encoder.encode(pngData);
    encoder.close();

    assert.ok(webpData.length > 0);
    assert.strictEqual(webpData.toString('ascii', 0, 4), 'RIFF');
    assert.strictEqual(webpData.toString('ascii', 8, 12), 'WEBP');
  });

  it('should reuse encoder for multiple images', () => {
    const encoder = new WebPEncoder({ quality: 75 });
    const webp1 = encoder.encode(readTestImage('red.png'));
    const webp2 = encoder.encode(readTestImage('blue.png'));
    encoder.close();

    assert.ok(webp1.length > 0);
    assert.ok(webp2.length > 0);
  });

  it('should throw on empty input', () => {
    const encoder = new WebPEncoder();
    assert.throws(() => encoder.encode(Buffer.alloc(0)), NextImageError);
    encoder.close();
  });
});

describe('WebPDecoder', () => {
  it('should decode WebP to pixels', () => {
    const encoder = new WebPEncoder({ quality: 90 });
    const webpData = encoder.encode(readTestImage('red.png'));
    encoder.close();

    const decoder = new WebPDecoder();
    const decoded = decoder.decode(webpData);
    decoder.close();

    assert.ok(decoded.width > 0);
    assert.ok(decoded.height > 0);
    assert.ok(decoded.data.length > 0);
  });
});
