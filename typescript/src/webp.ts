/**
 * WebP encoding/decoding API
 */

import koffi from 'koffi';
import { encodeWebPAlloc, freeBuffer, NextImageStatus, getStatusMessage } from './ffi';

/**
 * WebP encoding options
 */
export interface WebPEncodeOptions {
  quality?: number; // 0-100, default 75
  lossless?: boolean; // default false
  method?: number; // 0-6, default 4 (quality/speed trade-off)
}

/**
 * Encode image data (JPEG, PNG, etc.) to WebP
 *
 * @param inputData - Input image data (JPEG, PNG, GIF, etc.)
 * @param options - Encoding options (currently unused, reserved for future)
 * @returns WebP encoded data as Buffer
 * @throws Error if encoding fails
 */
export function encodeWebP(inputData: Buffer, options?: WebPEncodeOptions): Buffer {
  // Call FFI function
  const { status, output } = encodeWebPAlloc(inputData);

  // Check status
  if (status !== NextImageStatus.OK || !output) {
    throw new Error(`WebP encoding failed: ${getStatusMessage(status)}`);
  }

  // Extract the data pointer and size
  const dataPtr = output.data;
  const dataSize = Number(output.size);

  if (dataSize === 0) {
    throw new Error('WebP encoding returned empty buffer');
  }

  // The dataPtr is already a Buffer from our FFI binding
  // Note: The C-allocated buffer has already been freed in encodeWebPAlloc()
  const resultBuffer = dataPtr!;

  return resultBuffer;
}

/**
 * Convenience function: Encode with quality setting
 */
export function encodeWebPWithQuality(inputData: Buffer, quality: number): Buffer {
  // TODO: Implement quality setting when options struct is added
  return encodeWebP(inputData, { quality });
}

/**
 * Convenience function: Encode lossless
 */
export function encodeWebPLossless(inputData: Buffer): Buffer {
  // TODO: Implement lossless mode when options struct is added
  return encodeWebP(inputData, { lossless: true });
}
