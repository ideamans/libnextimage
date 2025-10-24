/**
 * WebP Encoder for Deno
 */

import {
  getLibrary,
  createBufferStruct,
  readBufferStruct,
  copyFromCMemory,
  getLastError
} from './ffi.ts'

export interface WebPEncodeOptions {
  quality?: number
  lossless?: boolean
  method?: number
  [key: string]: unknown
}

export class WebPEncoder {
  private encoder: Deno.PointerValue
  private options: WebPEncodeOptions

  constructor(options: Partial<WebPEncodeOptions> = {}) {
    this.options = { quality: 75, lossless: false, method: 4, ...options }

    // Create options struct (simplified - would need full struct marshalling)
    const optionsBuffer = new ArrayBuffer(256)
    const optionsView = new DataView(optionsBuffer)

    // Set quality (float at offset 0)
    optionsView.setFloat32(0, this.options.quality || 75, true)

    // Set lossless (int at offset 4)
    optionsView.setInt32(4, this.options.lossless ? 1 : 0, true)

    // Set method (int at offset 8)
    optionsView.setInt32(8, this.options.method || 4, true)

    const optionsPtr = Deno.UnsafePointer.of(optionsBuffer)
    const lib = getLibrary()

    this.encoder = lib.symbols.nextimage_webp_encoder_create(optionsPtr)

    if (!this.encoder) {
      throw new Error(`Failed to create WebP encoder: ${getLastError()}`)
    }
  }

  encode(data: Uint8Array): Uint8Array {
    if (!this.encoder) {
      throw new Error('Encoder has been closed')
    }

    const lib = getLibrary()
    const { pointer: outputPtr, view: outputView } = createBufferStruct()

    const status = lib.symbols.nextimage_webp_encoder_encode(
      this.encoder,
      data,
      data.length,
      outputPtr
    )

    if (status !== 0) {
      throw new Error(`WebP encoding failed: ${getLastError()}`)
    }

    const output = readBufferStruct(outputView)
    const result = copyFromCMemory(output.data, output.size)

    // Free C-allocated memory
    lib.symbols.nextimage_free_buffer(outputPtr)

    return result
  }

  encodeFile(path: string): Uint8Array {
    const data = Deno.readFileSync(path)
    return this.encode(data)
  }

  close(): void {
    if (this.encoder) {
      const lib = getLibrary()
      lib.symbols.nextimage_webp_encoder_destroy(this.encoder)
      this.encoder = null!
    }
  }

  static getDefaultOptions(): WebPEncodeOptions {
    return {
      quality: 75,
      lossless: false,
      method: 4
    }
  }
}
