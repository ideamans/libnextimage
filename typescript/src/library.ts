import * as path from 'path';
import * as fs from 'fs';

/**
 * Platform detection and library path resolution
 */

/**
 * Get the configured native library version
 * @returns Library version string (e.g., "0.4.0")
 */
export function getLibraryVersion(): string {
  try {
    // Try to read from the installed location (dist/../library-version.json)
    const versionFilePath = path.join(__dirname, '..', '..', 'library-version.json');
    if (fs.existsSync(versionFilePath)) {
      const config = JSON.parse(fs.readFileSync(versionFilePath, 'utf8'));
      return config.version;
    }
  } catch (error) {
    // Ignore errors and use fallback
  }

  // Fallback: return package version
  try {
    const packagePath = path.join(__dirname, '..', '..', 'package.json');
    const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    return packageJson.version;
  } catch (error) {
    return 'unknown';
  }
}

/**
 * Get the current platform identifier
 * @returns Platform string in the format: 'darwin-arm64', 'linux-amd64', etc.
 */
export function getPlatform(): string {
  const platform = process.platform;
  const arch = process.arch;

  if (platform === 'darwin') {
    return arch === 'arm64' ? 'darwin-arm64' : 'darwin-amd64';
  } else if (platform === 'linux') {
    // Node.js uses 'arm64' but some systems might report 'aarch64'
    return arch === 'arm64' ? 'linux-arm64' : 'linux-amd64';
  } else if (platform === 'win32') {
    return 'windows-amd64';
  }

  throw new Error(`Unsupported platform: ${platform}-${arch}`);
}

/**
 * Get the shared library file name for the current platform
 * @returns Library filename: 'libnextimage.dylib', 'libnextimage.so', or 'libnextimage.dll'
 */
export function getLibraryFileName(): string {
  const platform = process.platform;

  if (platform === 'darwin') {
    return 'libnextimage.dylib';
  } else if (platform === 'linux') {
    return 'libnextimage.so';
  } else if (platform === 'win32') {
    return 'libnextimage.dll';
  }

  throw new Error(`Unsupported platform: ${platform}`);
}

/**
 * Find the shared library path with fallback chain
 * Looks in the following order:
 * 1. ../../../lib/shared/ (development mode - relative to project root)
 * 2. ../../../lib/<platform>/ (development mode - platform-specific)
 * 3. ../lib/<platform>/ (installed package)
 *
 * Note: __dirname in compiled code (dist/src/) points to:
 * - Development: <project-root>/typescript/dist/src/
 * - Installed: node_modules/@ideamans/libnextimage/dist/src/
 */
export function findLibraryPath(): string {
  const platform = getPlatform();
  const libFileName = getLibraryFileName();

  // Priority 1: Development mode - shared build
  // __dirname = typescript/dist/src/, so ../../../ goes to project root
  const devSharedPath = path.join(__dirname, '..', '..', '..', 'lib', 'shared', libFileName);
  if (fs.existsSync(devSharedPath)) {
    return devSharedPath;
  }

  // Priority 2: Development mode - platform-specific build
  const devPlatformPath = path.join(__dirname, '..', '..', '..', 'lib', platform, libFileName);
  if (fs.existsSync(devPlatformPath)) {
    return devPlatformPath;
  }

  // Priority 3: Installed package
  // __dirname = node_modules/@ideamans/libnextimage/dist/
  const installedPath = path.join(__dirname, '..', 'lib', platform, libFileName);
  if (fs.existsSync(installedPath)) {
    return installedPath;
  }

  throw new Error(
    `Cannot find libnextimage shared library for ${platform}.\n` +
    `Searched paths:\n` +
    `  - ${devSharedPath} (development - shared)\n` +
    `  - ${devPlatformPath} (development - platform)\n` +
    `  - ${installedPath} (installed)\n\n` +
    `Please run 'make install-c' to build the library (development mode)\n` +
    `or ensure the package was installed correctly with prebuilt binaries.`
  );
}

/**
 * Get the library path (cached)
 * This function caches the result to avoid repeated filesystem lookups
 */
let cachedLibraryPath: string | null = null;

export function getLibraryPath(): string {
  if (!cachedLibraryPath) {
    cachedLibraryPath = findLibraryPath();
  }
  return cachedLibraryPath;
}

/**
 * Clear the cached library path (mainly for testing)
 */
export function clearLibraryPathCache(): void {
  cachedLibraryPath = null;
}
