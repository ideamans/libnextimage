/**
 * WebP encoding and decoding functions
 * Simplified API matching Golang implementation
 */

import koffi from 'koffi';
import {
  nextimage_webp_encode_alloc,
  nextimage_webp_decode_alloc,
  nextimage_webp_default_encode_options,
  nextimage_webp_default_decode_options,
  nextimage_free_buffer,
  nextimage_free_decode_buffer,
  nextimage_last_error_message,
  NextImageBufferStruct,
  NextImageDecodeBufferStruct,
  NextImageWebPEncodeOptionsStruct,
  NextImageWebPDecodeOptionsStruct
} from './ffi';
import {
  WebPEncodeOptions,
  WebPDecodeOptions,
  DecodedImage,
  NextImageStatus,
  NextImageError,
  isSuccess,
  PixelFormat
} from './types';

/**
 * Convert TypeScript options to C struct
 */
function convertEncodeOptions(opts: WebPEncodeOptions = {}): any {
  // Allocate C struct
  const cOptsPtr = koffi.alloc(NextImageWebPEncodeOptionsStruct, 1);

  // Initialize with defaults
  nextimage_webp_default_encode_options(cOptsPtr);

  // Decode to modify
  const cOpts = koffi.decode(cOptsPtr, NextImageWebPEncodeOptionsStruct) as any;

  // Override with user options
  if (opts.quality !== undefined) {
    cOpts.quality = opts.quality;
  }
  if (opts.lossless !== undefined) {
    cOpts.lossless = opts.lossless ? 1 : 0;
  }
  if (opts.method !== undefined) {
    cOpts.method = opts.method;
  }
  if (opts.preset !== undefined) {
    cOpts.preset = opts.preset;
  }
  if (opts.imageHint !== undefined) {
    cOpts.image_hint = opts.imageHint;
  }
  if (opts.exact !== undefined) {
    cOpts.exact = opts.exact ? 1 : 0;
  }

  // Encode back to pointer
  koffi.encode(cOptsPtr, NextImageWebPEncodeOptionsStruct, cOpts);

  return cOptsPtr;
}

/**
 * Convert TypeScript decode options to C struct
 */
function convertDecodeOptions(opts: WebPDecodeOptions = {}): any {
  // Allocate C struct
  const cOptsPtr = koffi.alloc(NextImageWebPDecodeOptionsStruct, 1);

  // Initialize with defaults
  nextimage_webp_default_decode_options(cOptsPtr);

  // Decode to modify
  const cOpts = koffi.decode(cOptsPtr, NextImageWebPDecodeOptionsStruct) as any;

  // Override with user options
  if (opts.useThreads !== undefined) {
    cOpts.use_threads = opts.useThreads ? 1 : 0;
  }
  if (opts.format !== undefined) {
    cOpts.format = opts.format;
  }

  // Encode back to pointer
  koffi.encode(cOptsPtr, NextImageWebPDecodeOptionsStruct, cOpts);

  return cOptsPtr;
}

/**
 * Encode an image file (JPEG, PNG, etc.) to WebP format
 *
 * @param imageFileData - Buffer containing image file data (JPEG, PNG, etc.)
 * @param options - WebP encoding options
 * @returns WebP encoded data as Buffer
 * @throws NextImageError if encoding fails
 *
 * Example:
 * ```typescript
 * import { encodeWebP } from '@ideamans/libnextimage';
 * import * as fs from 'fs';
 *
 * const jpegData = fs.readFileSync('input.jpg');
 * const webpData = encodeWebP(jpegData, { quality: 80 });
 * fs.writeFileSync('output.webp', webpData);
 * ```
 */
export function encodeWebP(imageFileData: Buffer, options: WebPEncodeOptions = {}): Buffer {
  if (!imageFileData || imageFileData.length === 0) {
    throw new NextImageError(NextImageStatus.ERROR_INVALID_PARAM, 'Empty input data');
  }

  const cOptsPtr = convertEncodeOptions(options);
  const outputPtr = koffi.alloc(NextImageBufferStruct, 1);

  const status = nextimage_webp_encode_alloc(
    imageFileData,
    imageFileData.length,
    cOptsPtr,
    outputPtr
  ) as NextImageStatus;

  if (!isSuccess(status)) {
    const errMsg = nextimage_last_error_message();
    throw new NextImageError(status, `WebP encode failed: ${errMsg || status}`);
  }

  // Decode output struct
  const output = koffi.decode(outputPtr, NextImageBufferStruct) as any;

  if (!output.data || output.size === 0) {
    throw new NextImageError(NextImageStatus.ERROR_ENCODE_FAILED, 'WebP encoding produced empty output');
  }

  // Copy data from C memory to JavaScript Buffer
  const dataSize = Number(output.size);
  const rawData = koffi.decode(output.data, koffi.array('uint8_t', dataSize));
  const result = Buffer.from(rawData as any);

  // Free C-allocated memory
  nextimage_free_buffer(outputPtr);

  return result;
}

/**
 * Decode WebP data to pixel data
 *
 * @param webpData - Buffer containing WebP encoded data
 * @param options - WebP decoding options
 * @returns Decoded image with pixel data
 * @throws NextImageError if decoding fails
 *
 * Example:
 * ```typescript
 * import { decodeWebP } from '@ideamans/libnextimage';
 * import * as fs from 'fs';
 *
 * const webpData = fs.readFileSync('input.webp');
 * const decoded = decodeWebP(webpData);
 * console.log(`Decoded ${decoded.width}x${decoded.height} image`);
 * // decoded.data contains raw pixel data (RGBA by default)
 * ```
 */
export function decodeWebP(webpData: Buffer, options: WebPDecodeOptions = {}): DecodedImage {
  if (!webpData || webpData.length === 0) {
    throw new NextImageError(NextImageStatus.ERROR_INVALID_PARAM, 'Empty input data');
  }

  const cOptsPtr = convertDecodeOptions(options);
  const outputPtr = koffi.alloc(NextImageDecodeBufferStruct, 1);

  const status = nextimage_webp_decode_alloc(
    webpData,
    webpData.length,
    cOptsPtr,
    outputPtr
  ) as NextImageStatus;

  if (!isSuccess(status)) {
    const errMsg = nextimage_last_error_message();
    throw new NextImageError(status, `WebP decode failed: ${errMsg || status}`);
  }

  // Decode output struct
  const output = koffi.decode(outputPtr, NextImageDecodeBufferStruct) as any;

  if (!output.data || output.width === 0 || output.height === 0) {
    throw new NextImageError(NextImageStatus.ERROR_DECODE_FAILED, 'WebP decoding produced invalid output');
  }

  // Copy data from C memory to JavaScript Buffer
  const dataSize = Number(output.data_size) || (Number(output.stride) * Number(output.height));
  const rawData = koffi.decode(output.data, koffi.array('uint8_t', dataSize));
  const copiedData = Buffer.from(rawData as any);

  // Create result object
  const result: DecodedImage = {
    data: copiedData,
    width: Number(output.width),
    height: Number(output.height),
    stride: Number(output.stride),
    format: Number(output.format) as PixelFormat,
    bitDepth: Number(output.bit_depth)
  };

  // Free C-allocated memory
  nextimage_free_decode_buffer(outputPtr);

  return result;
}
