import fs from 'fs'
import path from 'path'
import { WebPEncoder, AVIFEncoder } from 'libnextimage'

console.log('=== libnextimage Node.js E2E Test: All Features ===\n')

// Create output directory
if (!fs.existsSync('output')) {
  fs.mkdirSync('output', { recursive: true })
}

const testImagePath = path.join('..', '..', 'testdata', 'jpeg-source', 'landscape-like.jpg')

if (!fs.existsSync(testImagePath)) {
  console.error(`Test image not found: ${testImagePath}`)
  process.exit(1)
}

const jpegData = fs.readFileSync(testImagePath)
console.log(`Input: ${jpegData.length} bytes\n`)

// WebP: Different quality levels
console.log('--- WebP: Quality Levels ---')
for (const quality of [50, 75, 90]) {
  const encoder = new WebPEncoder({ quality })
  const webpData = encoder.encode(jpegData)
  const outputPath = path.join('output', `quality-${quality}.webp`)
  fs.writeFileSync(outputPath, webpData)
  const compression = ((1 - webpData.length / jpegData.length) * 100).toFixed(1)
  console.log(`✓ Quality ${quality}: ${(webpData.length / 1024).toFixed(2)} KB (${compression}% smaller)`)
  encoder.close()
}

// WebP: Lossless mode
console.log('\n--- WebP: Lossless Mode ---')
const losslessEncoder = new WebPEncoder({ lossless: true })
const losslessData = losslessEncoder.encode(jpegData)
fs.writeFileSync(path.join('output', 'lossless.webp'), losslessData)
console.log(`✓ Lossless: ${(losslessData.length / 1024).toFixed(2)} KB`)
losslessEncoder.close()

// AVIF: Different quality/speed combinations
console.log('\n--- AVIF: Quality/Speed Combinations ---')
const combinations = [
  { quality: 50, speed: 8, label: 'Fast/Low' },
  { quality: 60, speed: 6, label: 'Balanced' },
  { quality: 70, speed: 4, label: 'Quality' }
]

for (const { quality, speed, label } of combinations) {
  const encoder = new AVIFEncoder({ quality, speed })
  const avifData = encoder.encode(jpegData)
  const outputPath = path.join('output', `avif-${label.toLowerCase().replace('/', '-')}.avif`)
  fs.writeFileSync(outputPath, avifData)
  const compression = ((1 - avifData.length / jpegData.length) * 100).toFixed(1)
  console.log(`✓ ${label} (q=${quality}, s=${speed}): ${(avifData.length / 1024).toFixed(2)} KB (${compression}% smaller)`)
  encoder.close()
}

// WebP: Metadata preservation
console.log('\n--- WebP: Metadata Preservation ---')
const metadataEncoder = new WebPEncoder({
  quality: 80,
  keepMetadata: 7 // MetadataAll
})
const metadataWebP = metadataEncoder.encode(jpegData)
fs.writeFileSync(path.join('output', 'with-metadata.webp'), metadataWebP)
console.log(`✓ With metadata: ${(metadataWebP.length / 1024).toFixed(2)} KB`)
metadataEncoder.close()

// Summary
console.log('\n' + '='.repeat(60))
console.log('Feature test complete!')
console.log('\nGenerated files in output/:')
console.log('  - quality-*.webp (different quality levels)')
console.log('  - lossless.webp (lossless mode)')
console.log('  - avif-*.avif (different speed/quality)')
console.log('  - with-metadata.webp (metadata preservation)')
console.log('='.repeat(60))

console.log('\n✅ All feature tests passed!')
