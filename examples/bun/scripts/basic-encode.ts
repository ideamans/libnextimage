import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'fs'
import { join } from 'path'
import { WebPEncoder, getLibraryVersion } from '../../typescript/bun/mod.ts'

console.log('=== libnextimage Bun E2E Test: Basic Encode ===\n')
console.log(`Library version: ${getLibraryVersion()}\n`)

// Create output directory
if (!existsSync('output')) {
  mkdirSync('output', { recursive: true })
}

// Use test image from testdata
const testImagePath = join('..', '..', 'testdata', 'jpeg-source', 'landscape-like.jpg')

try {
  const jpegData = readFileSync(testImagePath)
  console.log(`Input: ${jpegData.length} bytes (${(jpegData.length / 1024).toFixed(2)} KB)`)

  // WebP encoding test
  console.log('\n--- WebP Encoding ---')
  const webpEncoder = new WebPEncoder({ quality: 80 })
  const webpData = webpEncoder.encode(jpegData)
  writeFileSync(join('output', 'test-basic.webp'), webpData)

  const webpCompression = ((1 - webpData.length / jpegData.length) * 100).toFixed(1)
  console.log(`✓ WebP: ${webpData.length} bytes (${(webpData.length / 1024).toFixed(2)} KB)`)
  console.log(`  Compression: ${webpCompression}% smaller`)
  webpEncoder.close()

  console.log('\n✅ Bun basic encoding test passed!')
  console.log(`\nOutput file: output/test-basic.webp`)
} catch (error) {
  console.error(`Error: ${error.message}`)
  process.exit(1)
}
