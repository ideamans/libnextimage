/**
 * GIF Conversion Tests
 */

import { test } from 'node:test';
import assert from 'node:assert';
import { readFileSync } from 'fs';
import { join } from 'path';
import { GIF2WebPConverter, WebP2GIFConverter, WebPEncoder } from '../src/index';

test('GIF2WebP: Convert static GIF to WebP', () => {
  const gifPath = join(__dirname, '..', '..', '..', 'testdata', 'gif', 'static.gif');
  const gifData = readFileSync(gifPath);

  const converter = new GIF2WebPConverter({ quality: 80 });
  const webpData = converter.convert(gifData);
  converter.close();

  assert.ok(webpData.length > 0, 'WebP data should not be empty');
  // WebP signature: "RIFF....WEBP"
  assert.strictEqual(webpData[0], 0x52, 'Should start with R'); // R
  assert.strictEqual(webpData[1], 0x49, 'Should have I'); // I
  assert.strictEqual(webpData[2], 0x46, 'Should have F'); // F
  assert.strictEqual(webpData[3], 0x46, 'Should have F'); // F
});

test('WebP2GIF: Convert WebP to GIF', () => {
  // First create a WebP from PNG
  const pngPath = join(__dirname, '..', '..', '..', 'testdata', 'png', 'red.png');
  const pngData = readFileSync(pngPath);

  const webpEncoder = new WebPEncoder({ quality: 90 });
  const webpData = webpEncoder.encode(pngData);
  webpEncoder.close();

  // Convert WebP to GIF
  const converter = new WebP2GIFConverter();
  const gifData = converter.convert(webpData);
  converter.close();

  assert.ok(gifData.length > 0, 'GIF data should not be empty');
  // GIF signature: "GIF87a" or "GIF89a"
  assert.strictEqual(gifData.toString('utf8', 0, 3), 'GIF', 'Should have GIF signature');
});

test('GIF2WebP: Converter reuse', () => {
  const gifPath = join(__dirname, '..', '..', '..', 'testdata', 'gif', 'static.gif');
  const gifData = readFileSync(gifPath);

  const converter = new GIF2WebPConverter({ quality: 75 });

  // First conversion
  const webp1 = converter.convert(gifData);
  assert.ok(webp1.length > 0);

  // Second conversion with same converter
  const webp2 = converter.convert(gifData);
  assert.ok(webp2.length > 0);

  converter.close();

  // Should throw after close
  assert.throws(() => converter.convert(gifData), /closed/);
});

test('WebP2GIF: Converter reuse', () => {
  const pngPath = join(__dirname, '..', '..', '..', 'testdata', 'png', 'blue.png');
  const pngData = readFileSync(pngPath);

  // Create WebP
  const webpEncoder = new WebPEncoder({ quality: 85 });
  const webpData = webpEncoder.encode(pngData);
  webpEncoder.close();

  const converter = new WebP2GIFConverter();

  // First conversion
  const gif1 = converter.convert(webpData);
  assert.ok(gif1.length > 0);

  // Second conversion with same converter
  const gif2 = converter.convert(webpData);
  assert.ok(gif2.length > 0);

  converter.close();

  // Should throw after close
  assert.throws(() => converter.convert(webpData), /closed/);
});

test('GIF2WebP: Error on empty data', () => {
  const converter = new GIF2WebPConverter();
  assert.throws(() => converter.convert(Buffer.alloc(0)), /empty/);
  converter.close();
});

test('WebP2GIF: Error on empty data', () => {
  const converter = new WebP2GIFConverter();
  assert.throws(() => converter.convert(Buffer.alloc(0)), /empty/);
  converter.close();
});

test('GIF2WebP: Different quality levels', () => {
  const gifPath = join(__dirname, '..', '..', '..', 'testdata', 'gif', 'static.gif');
  const gifData = readFileSync(gifPath);

  const converter50 = new GIF2WebPConverter({ quality: 50 });
  const webp50 = converter50.convert(gifData);
  converter50.close();

  const converter90 = new GIF2WebPConverter({ quality: 90 });
  const webp90 = converter90.convert(gifData);
  converter90.close();

  // Higher quality should generally produce larger files
  assert.ok(webp50.length < webp90.length || webp50.length === webp90.length);
  assert.ok(webp50.length > 0);
  assert.ok(webp90.length > 0);
});
