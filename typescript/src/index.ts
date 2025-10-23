/**
 * libnextimage TypeScript/Node.js bindings
 * High-performance WebP and AVIF image processing
 */

// Export types
export * from './types';

// Export library utilities
export {
  getPlatform,
  getLibraryFileName,
  getLibraryPath,
  clearLibraryPathCache
} from './library';

// Export WebP functions
export { encodeWebP, decodeWebP } from './webp';

// Export AVIF functions
export { encodeAVIF, decodeAVIF } from './avif';
