/**
 * libnextimage TypeScript/Node.js bindings
 *
 * This module provides TypeScript bindings for the libnextimage C library,
 * enabling WebP and AVIF image encoding/decoding in Node.js.
 *
 * @example
 * ```typescript
 * import { getLibraryPath, encodeWebP } from '@ideamans/libnextimage';
 *
 * // Get the path to the shared library
 * const libPath = getLibraryPath();
 * console.log(`Using library: ${libPath}`);
 *
 * // TODO: Add encoding/decoding examples once FFI bindings are implemented
 * ```
 */

export { getLibraryPath, getPlatform, getLibraryFileName } from './library';

// WebP encoding/decoding
export {
  encodeWebP,
  encodeWebPWithQuality,
  encodeWebPLossless,
  WebPEncodeOptions,
} from './webp';

// FFI low-level bindings (for advanced users)
export { NextImageStatus, getStatusMessage } from './ffi';

/**
 * Version information
 */
export const VERSION = '0.4.0';
