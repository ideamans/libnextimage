import fs from 'fs'
import path from 'path'
import { WebPEncoder, WebPDecoder, AVIFEncoder, AVIFDecoder } from '@ideamans/libnextimage'

console.log('=== libnextimage Node.js E2E Test: Decode ===\n')

// Create output directory
if (!fs.existsSync('output')) {
  fs.mkdirSync('output', { recursive: true })
}

// Use test image
const testImagePath = path.join('..', '..', 'testdata', 'jpeg-source', 'gradient-horizontal.jpg')

if (!fs.existsSync(testImagePath)) {
  console.error(`Test image not found: ${testImagePath}`)
  process.exit(1)
}

const jpegData = fs.readFileSync(testImagePath)
console.log(`Input: ${jpegData.length} bytes\n`)

// WebP encode → decode test
console.log('--- WebP Round-trip Test ---')
const webpEncoder = new WebPEncoder({ quality: 90 })
const webpData = webpEncoder.encode(jpegData)
console.log(`✓ Encoded to WebP: ${webpData.length} bytes`)

const webpDecoder = new WebPDecoder({ format: 'RGBA' })
const webpDecoded = webpDecoder.decode(webpData)
console.log(`✓ Decoded from WebP: ${webpDecoded.width}x${webpDecoded.height}, ${webpDecoded.data.length} bytes`)
console.log(`  Format: ${webpDecoded.format}`)
console.log(`  Expected RGBA size: ${webpDecoded.width * webpDecoded.height * 4} bytes`)

webpEncoder.close()
webpDecoder.close()

// AVIF encode → decode test
console.log('\n--- AVIF Round-trip Test ---')
const avifEncoder = new AVIFEncoder({ quality: 70, speed: 6 })
const avifData = avifEncoder.encode(jpegData)
console.log(`✓ Encoded to AVIF: ${avifData.length} bytes`)

const avifDecoder = new AVIFDecoder({ format: 'RGBA' })
const avifDecoded = avifDecoder.decode(avifData)
console.log(`✓ Decoded from AVIF: ${avifDecoded.width}x${avifDecoded.height}, ${avifDecoded.data.length} bytes`)
console.log(`  Format: ${avifDecoded.format}`)
console.log(`  Expected RGBA size: ${avifDecoded.width * avifDecoded.height * 4} bytes`)

avifEncoder.close()
avifDecoder.close()

// Verify decoded data
console.log('\n--- Verification ---')
if (webpDecoded.width === avifDecoded.width && webpDecoded.height === avifDecoded.height) {
  console.log(`✓ Dimensions match: ${webpDecoded.width}x${webpDecoded.height}`)
} else {
  console.error('✗ Dimension mismatch!')
  process.exit(1)
}

if (webpDecoded.data.length === avifDecoded.data.length) {
  console.log(`✓ Data size matches: ${webpDecoded.data.length} bytes`)
} else {
  console.error('✗ Data size mismatch!')
  process.exit(1)
}

console.log('\n✅ All decode tests passed!')
