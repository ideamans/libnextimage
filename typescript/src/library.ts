import * as path from 'path';
import * as fs from 'fs';

/**
 * Platform detection and library path resolution
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
 * Find the shared library path
 * Looks in the following order:
 * 1. ../lib/<platform>/ (development mode - relative to project root)
 * 2. ./lib/<platform>/ (installed package)
 */
export function findLibraryPath(): string {
  const platformDir = getPlatform();
  const libFileName = getLibraryFileName();

  // Try development mode first
  // __dirname in compiled code will be dist/src/, so we need to go up 3 levels to project root
  const devPath = path.join(__dirname, '..', '..', '..', 'lib', platformDir, libFileName);
  if (fs.existsSync(devPath)) {
    return devPath;
  }

  // Try installed package
  const installedPath = path.join(__dirname, '..', 'lib', platformDir, libFileName);
  if (fs.existsSync(installedPath)) {
    return installedPath;
  }

  throw new Error(
    `Cannot find libnextimage shared library.\n` +
    `Searched paths:\n` +
    `  - ${devPath} (development)\n` +
    `  - ${installedPath} (installed)\n\n` +
    `Please run 'make install-c' to build the library.`
  );
}

/**
 * Get the library path (cached)
 */
let cachedLibraryPath: string | null = null;
export function getLibraryPath(): string {
  if (!cachedLibraryPath) {
    cachedLibraryPath = findLibraryPath();
  }
  return cachedLibraryPath;
}
