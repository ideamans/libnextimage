import fs from 'fs'
import path from 'path'
import { WebPEncoder } from '@ideamans/libnextimage'

console.log('=== libnextimage Node.js E2E Test: Batch Processing ===\n')

// Create output directory
if (!fs.existsSync('output')) {
  fs.mkdirSync('output', { recursive: true })
}

// Find all JPEG test images
const testdataPath = path.join('..', '..', 'testdata', 'jpeg-source')

if (!fs.existsSync(testdataPath)) {
  console.error(`Testdata directory not found: ${testdataPath}`)
  process.exit(1)
}

const files = fs
  .readdirSync(testdataPath)
  .filter((f) => f.endsWith('.jpg') || f.endsWith('.jpeg'))

if (files.length === 0) {
  console.error('No JPEG files found in testdata/jpeg-source/')
  process.exit(1)
}

console.log(`Found ${files.length} JPEG files\n`)

// Create single encoder instance (reuse for efficiency)
const encoder = new WebPEncoder({ quality: 80 })

let totalInput = 0
let totalOutput = 0

console.log('Converting...\n')

for (const file of files) {
  const inputPath = path.join(testdataPath, file)
  const outputPath = path.join('output', file.replace(/\.jpe?g$/i, '.webp'))

  const inputData = fs.readFileSync(inputPath)
  const webpData = encoder.encode(inputData)
  fs.writeFileSync(outputPath, webpData)

  totalInput += inputData.length
  totalOutput += webpData.length

  const compression = ((1 - webpData.length / inputData.length) * 100).toFixed(1)
  console.log(`✓ ${file}`)
  console.log(`  ${(inputData.length / 1024).toFixed(2)} KB → ${(webpData.length / 1024).toFixed(2)} KB (${compression}% smaller)`)
}

encoder.close()

// Summary
console.log('\n' + '='.repeat(60))
console.log('Batch conversion complete!')
console.log(`  Files processed: ${files.length}`)
console.log(`  Total input: ${(totalInput / 1024 / 1024).toFixed(2)} MB`)
console.log(`  Total output: ${(totalOutput / 1024 / 1024).toFixed(2)} MB`)
const overallCompression = ((1 - totalOutput / totalInput) * 100).toFixed(1)
console.log(`  Overall compression: ${overallCompression}%`)
console.log('='.repeat(60))

console.log('\n✅ Batch processing test passed!')
