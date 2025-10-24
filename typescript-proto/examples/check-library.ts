/**
 * Simple example to verify library loading
 *
 * Usage:
 *   npm run build
 *   node dist/examples/check-library.js
 */

import { getLibraryPath, getPlatform, getLibraryFileName } from '../src/index';
import * as fs from 'fs';

console.log('=== libnextimage Library Check ===\n');

// 1. Platform detection
console.log(`Platform: ${getPlatform()}`);
console.log(`Library file name: ${getLibraryFileName()}\n`);

// 2. Library path resolution
try {
  const libPath = getLibraryPath();
  console.log(`✓ Library found: ${libPath}`);

  // 3. Verify file exists and get size
  const stats = fs.statSync(libPath);
  console.log(`  Size: ${(stats.size / 1024 / 1024).toFixed(2)} MB`);
  console.log(`  Modified: ${stats.mtime.toISOString()}\n`);

  console.log('✓ Library is ready for use!');
} catch (error) {
  console.error(`✗ Error: ${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);
}
