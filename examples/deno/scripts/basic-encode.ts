import { WebPEncoder, getLibraryVersion } from 'libnextimage'
import { join } from 'https://deno.land/std@0.224.0/path/mod.ts'
import { ensureDirSync } from 'https://deno.land/std@0.224.0/fs/mod.ts'

console.log('=== libnextimage Deno E2E Test: Basic Encode ===\n')
console.log(`Library version: ${getLibraryVersion()}\n`)

// Create output directory
ensureDirSync('output')

// Use test image from testdata
const testImagePath = join('..', '..', 'testdata', 'jpeg-source', 'landscape-like.jpg')

try {
  const jpegData = await Deno.readFile(testImagePath)
  console.log(`Input: ${jpegData.length} bytes (${(jpegData.length / 1024).toFixed(2)} KB)`)

  // WebP encoding test
  console.log('\n--- WebP Encoding ---')
  const webpEncoder = new WebPEncoder({ quality: 80 })
  const webpData = webpEncoder.encode(jpegData)
  await Deno.writeFile(join('output', 'test-basic.webp'), webpData)

  const webpCompression = ((1 - webpData.length / jpegData.length) * 100).toFixed(1)
  console.log(`✓ WebP: ${webpData.length} bytes (${(webpData.length / 1024).toFixed(2)} KB)`)
  console.log(`  Compression: ${webpCompression}% smaller`)
  webpEncoder.close()

  console.log('\n✅ Deno basic encoding test passed!')
  console.log(`\nOutput file: output/test-basic.webp`)
} catch (error) {
  console.error(`Error: ${error.message}`)
  Deno.exit(1)
}
