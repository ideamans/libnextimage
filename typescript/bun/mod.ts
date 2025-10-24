/**
 * libnextimage for Bun
 * High-performance WebP and AVIF image processing
 */

// Export library utilities
export {
  getLibraryVersion,
  getPlatform,
  getLibraryFileName,
  getLibraryPath,
  clearLibraryPathCache
} from './library.ts'

// Export WebP encoder
export { WebPEncoder } from './webp-encoder.ts'
export type { WebPEncodeOptions } from './webp-encoder.ts'

// Note: Full implementation would include:
// - WebPDecoder
// - AVIFEncoder
// - AVIFDecoder
// - All type definitions

// This is a minimal implementation to demonstrate the structure
