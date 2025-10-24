import fs from 'fs'
import path from 'path'
import { WebPEncoder, AVIFEncoder, getLibraryVersion } from 'libnextimage'

console.log('=== libnextimage Node.js E2E Test: Basic Encode ===\n')
console.log(`Library version: ${getLibraryVersion()}\n`)

// Create output directory if it doesn't exist
if (!fs.existsSync('output')) {
  fs.mkdirSync('output', { recursive: true })
}

// Use test image from testdata
const testImagePath = path.join('..', '..', 'testdata', 'jpeg-source', 'landscape-like.jpg')

if (!fs.existsSync(testImagePath)) {
  console.error(`Test image not found: ${testImagePath}`)
  console.error('Please ensure testdata/ directory exists in the repository root')
  process.exit(1)
}

const jpegData = fs.readFileSync(testImagePath)
console.log(`Input: ${jpegData.length} bytes (${(jpegData.length / 1024).toFixed(2)} KB)`)

// WebP encoding test
console.log('\n--- WebP Encoding ---')
const webpEncoder = new WebPEncoder({ quality: 80 })
const webpData = webpEncoder.encode(jpegData)
fs.writeFileSync(path.join('output', 'test-basic.webp'), webpData)
const webpCompression = ((1 - webpData.length / jpegData.length) * 100).toFixed(1)
console.log(`✓ WebP: ${webpData.length} bytes (${(webpData.length / 1024).toFixed(2)} KB)`)
console.log(`  Compression: ${webpCompression}% smaller`)
webpEncoder.close()

// AVIF encoding test
console.log('\n--- AVIF Encoding ---')
const avifEncoder = new AVIFEncoder({ quality: 60, speed: 6 })
const avifData = avifEncoder.encode(jpegData)
fs.writeFileSync(path.join('output', 'test-basic.avif'), avifData)
const avifCompression = ((1 - avifData.length / jpegData.length) * 100).toFixed(1)
console.log(`✓ AVIF: ${avifData.length} bytes (${(avifData.length / 1024).toFixed(2)} KB)`)
console.log(`  Compression: ${avifCompression}% smaller`)
avifEncoder.close()

console.log('\n✅ All basic encoding tests passed!')
console.log(`\nOutput files:`)
console.log(`  - output/test-basic.webp`)
console.log(`  - output/test-basic.avif`)
