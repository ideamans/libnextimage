/**
 * Deno-specific library path resolution
 * Uses import.meta.url instead of __dirname
 */

import { existsSync } from 'https://deno.land/std@0.224.0/fs/mod.ts'
import { join, dirname, fromFileUrl } from 'https://deno.land/std@0.224.0/path/mod.ts'

/**
 * Get the current platform identifier
 */
export function getPlatform(): string {
  const os = Deno.build.os
  const arch = Deno.build.arch

  if (os === 'darwin') {
    return arch === 'aarch64' ? 'darwin-arm64' : 'darwin-amd64'
  } else if (os === 'linux') {
    return arch === 'aarch64' ? 'linux-arm64' : 'linux-amd64'
  } else if (os === 'windows') {
    return 'windows-amd64'
  }

  throw new Error(`Unsupported platform: ${os}-${arch}`)
}

/**
 * Get the shared library file name for the current platform
 */
export function getLibraryFileName(): string {
  const os = Deno.build.os

  if (os === 'darwin') {
    return 'libnextimage.dylib'
  } else if (os === 'linux') {
    return 'libnextimage.so'
  } else if (os === 'windows') {
    return 'libnextimage.dll'
  }

  throw new Error(`Unsupported platform: ${os}`)
}

/**
 * Find the shared library path with fallback chain
 */
export function findLibraryPath(): string {
  const platform = getPlatform()
  const libFileName = getLibraryFileName()

  // Get current module directory from import.meta.url
  const moduleUrl = import.meta.url
  const modulePath = fromFileUrl(moduleUrl)
  const moduleDir = dirname(modulePath)

  // Priority 1: Development mode - shared build
  // typescript/deno/ -> ../../lib/shared/
  const devSharedPath = join(moduleDir, '..', '..', 'lib', 'shared', libFileName)
  if (existsSync(devSharedPath)) {
    return devSharedPath
  }

  // Priority 2: Development mode - platform-specific build
  // typescript/deno/ -> ../../lib/<platform>/
  const devPlatformPath = join(moduleDir, '..', '..', 'lib', platform, libFileName)
  if (existsSync(devPlatformPath)) {
    return devPlatformPath
  }

  // Priority 3: Installed from deno.land/x or npm
  // Look for lib/<platform>/ relative to module
  const installedPath = join(moduleDir, '..', 'lib', platform, libFileName)
  if (existsSync(installedPath)) {
    return installedPath
  }

  throw new Error(
    `Cannot find libnextimage shared library for ${platform}.\n` +
      `Searched paths:\n` +
      `  - ${devSharedPath} (development - shared)\n` +
      `  - ${devPlatformPath} (development - platform)\n` +
      `  - ${installedPath} (installed)\n\n` +
      `Please ensure the library is properly installed.`
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
 * Get the library version from configuration
 */
export function getLibraryVersion(): string {
  try {
    const moduleDir = dirname(fromFileUrl(import.meta.url))
    const versionFilePath = join(moduleDir, '..', 'library-version.json')

    if (existsSync(versionFilePath)) {
      const config = JSON.parse(Deno.readTextFileSync(versionFilePath))
      return config.version
    }
  } catch (_error) {
    // Ignore errors
  }

  return 'unknown'
}
