/**
 * Bun-specific library path resolution
 * Bun supports both __dirname and import.meta.url
 */

import * as path from 'path'
import * as fs from 'fs'
import { fileURLToPath } from 'url'

/**
 * Get the current platform identifier
 */
export function getPlatform(): string {
  const platform = process.platform
  const arch = process.arch

  if (platform === 'darwin') {
    return arch === 'arm64' ? 'darwin-arm64' : 'darwin-amd64'
  } else if (platform === 'linux') {
    return arch === 'arm64' ? 'linux-arm64' : 'linux-amd64'
  } else if (platform === 'win32') {
    return 'windows-amd64'
  }

  throw new Error(`Unsupported platform: ${platform}-${arch}`)
}

/**
 * Get the shared library file name for the current platform
 */
export function getLibraryFileName(): string {
  const platform = process.platform

  if (platform === 'darwin') {
    return 'libnextimage.dylib'
  } else if (platform === 'linux') {
    return 'libnextimage.so'
  } else if (platform === 'win32') {
    return 'libnextimage.dll'
  }

  throw new Error(`Unsupported platform: ${platform}`)
}

/**
 * Find the shared library path with fallback chain
 */
export function findLibraryPath(): string {
  const platformId = getPlatform()
  const libFileName = getLibraryFileName()

  // Get current directory using import.meta.url (Bun supports this)
  const currentFile = fileURLToPath(import.meta.url)
  const moduleDir = path.dirname(currentFile)

  // Priority 1: Development mode - shared build
  // typescript/bun/ -> ../../lib/shared/
  const devSharedPath = path.join(moduleDir, '..', '..', 'lib', 'shared', libFileName)
  if (fs.existsSync(devSharedPath)) {
    return devSharedPath
  }

  // Priority 2: Development mode - platform-specific build
  // typescript/bun/ -> ../../lib/<platform>/
  const devPlatformPath = path.join(moduleDir, '..', '..', 'lib', platformId, libFileName)
  if (fs.existsSync(devPlatformPath)) {
    return devPlatformPath
  }

  // Priority 3: Installed package
  // typescript/bun/ -> ../lib/<platform>/
  const installedPath = path.join(moduleDir, '..', 'lib', platformId, libFileName)
  if (fs.existsSync(installedPath)) {
    return installedPath
  }

  throw new Error(
    `Cannot find libnextimage shared library for ${platformId}.\n` +
      `Searched paths:\n` +
      `  - ${devSharedPath} (development - shared)\n` +
      `  - ${devPlatformPath} (development - platform)\n` +
      `  - ${installedPath} (installed)\n\n` +
      `Please run 'make install-c' to build the library (development mode)\n` +
      `or ensure the package was installed correctly with prebuilt binaries.`
  )
}

/**
 * Get the library path (cached)
 */
let cachedLibraryPath: string | null = null

export function getLibraryPath(): string {
  if (!cachedLibraryPath) {
    cachedLibraryPath = findLibraryPath()
  }
  return cachedLibraryPath
}

/**
 * Clear the cached library path (mainly for testing)
 */
export function clearLibraryPathCache(): void {
  cachedLibraryPath = null
}

/**
 * Get the configured native library version
 */
export function getLibraryVersion(): string {
  try {
    const currentFile = fileURLToPath(import.meta.url)
    const moduleDir = path.dirname(currentFile)
    const versionFilePath = path.join(moduleDir, '..', 'library-version.json')

    if (fs.existsSync(versionFilePath)) {
      const config = JSON.parse(fs.readFileSync(versionFilePath, 'utf8'))
      return config.version
    }
  } catch (_error) {
    // Ignore errors
  }

  return 'unknown'
}
