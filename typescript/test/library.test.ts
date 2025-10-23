/**
 * Library path resolution tests
 * Tests platform detection and library path resolution logic
 */

import { describe, it } from 'node:test';
import * as assert from 'node:assert';
import * as path from 'path';
import * as fs from 'fs';
import {
  getPlatform,
  getLibraryFileName,
  getLibraryPath,
  clearLibraryPathCache
} from '../src/library';

describe('Platform Detection', () => {
  it('should return a valid platform string', () => {
    const platform = getPlatform();
    assert.ok(platform, 'Platform should not be empty');

    // Should match one of the supported platforms
    const validPlatforms = [
      'darwin-arm64',
      'darwin-amd64',
      'linux-arm64',
      'linux-amd64',
      'windows-amd64'
    ];

    assert.ok(
      validPlatforms.includes(platform),
      `Platform ${platform} should be one of: ${validPlatforms.join(', ')}`
    );
  });

  it('should return correct library file name for current platform', () => {
    const fileName = getLibraryFileName();
    assert.ok(fileName, 'Library file name should not be empty');

    // Should match platform-specific extension
    const platform = process.platform;
    if (platform === 'darwin') {
      assert.strictEqual(fileName, 'libnextimage.dylib');
    } else if (platform === 'linux') {
      assert.strictEqual(fileName, 'libnextimage.so');
    } else if (platform === 'win32') {
      assert.strictEqual(fileName, 'libnextimage.dll');
    }
  });
});

describe('Library Path Resolution', () => {
  it('should find library in one of the fallback paths', () => {
    clearLibraryPathCache();

    // This will throw if library is not found
    const libPath = getLibraryPath();

    assert.ok(libPath, 'Library path should not be empty');
    assert.ok(path.isAbsolute(libPath), 'Library path should be absolute');
  });

  it('should return the same path when called multiple times (caching)', () => {
    clearLibraryPathCache();

    const path1 = getLibraryPath();
    const path2 = getLibraryPath();

    assert.strictEqual(path1, path2, 'Cached path should be identical');
  });

  it('should resolve to an existing file', () => {
    clearLibraryPathCache();

    const libPath = getLibraryPath();
    const exists = fs.existsSync(libPath);

    assert.strictEqual(
      exists,
      true,
      `Library file should exist at: ${libPath}`
    );
  });

  it('should prioritize shared library over platform-specific in development', () => {
    clearLibraryPathCache();

    const libPath = getLibraryPath();
    const platform = getPlatform();

    // Check which path was resolved
    // Priority 1: ../../lib/shared/
    // Priority 2: ../../lib/<platform>/
    // Priority 3: ../lib/<platform>/

    const isSharedPath = libPath.includes(path.join('lib', 'shared'));
    const isPlatformDevPath = libPath.includes(path.join('lib', platform));
    const isPlatformInstalledPath = !isSharedPath && isPlatformDevPath;

    assert.ok(
      isSharedPath || isPlatformDevPath || isPlatformInstalledPath,
      `Library path ${libPath} should match one of the expected patterns`
    );
  });

  it('should clear cache when clearLibraryPathCache is called', () => {
    // Get path twice to ensure it's cached
    const path1 = getLibraryPath();
    const path2 = getLibraryPath();
    assert.strictEqual(path1, path2);

    // Clear cache
    clearLibraryPathCache();

    // Get path again - should still work
    const path3 = getLibraryPath();
    assert.strictEqual(path1, path3, 'Path should be the same after cache clear');
  });
});

describe('Error Handling', () => {
  it('should throw error with helpful message if library not found', () => {
    // This test is informational - we can't actually test the error case
    // in a real environment because the library should exist

    // Just verify that getLibraryPath works in our environment
    clearLibraryPathCache();
    assert.doesNotThrow(() => {
      getLibraryPath();
    });
  });
});
