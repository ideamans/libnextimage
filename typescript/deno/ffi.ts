/**
 * Deno FFI bindings for libnextimage
 * Uses Deno.dlopen() instead of Koffi
 */

import { getLibraryPath } from './library.ts'

// Type definitions matching the C API
export interface NextImageBuffer {
  data: Deno.PointerValue
  size: number
}

export interface NextImageDecodeBuffer {
  data: Deno.PointerValue
  size: number
  width: number
  height: number
  format: number
}

// FFI symbols definition
const symbols = {
  // WebP Encoder
  nextimage_webp_encoder_create: {
    parameters: ['pointer'],
    result: 'pointer'
  },
  nextimage_webp_encoder_encode: {
    parameters: ['pointer', 'buffer', 'usize', 'pointer'],
    result: 'i32'
  },
  nextimage_webp_encoder_destroy: {
    parameters: ['pointer'],
    result: 'void'
  },

  // WebP Decoder
  nextimage_webp_decoder_create: {
    parameters: ['pointer'],
    result: 'pointer'
  },
  nextimage_webp_decoder_decode: {
    parameters: ['pointer', 'buffer', 'usize', 'pointer'],
    result: 'i32'
  },
  nextimage_webp_decoder_destroy: {
    parameters: ['pointer'],
    result: 'void'
  },

  // AVIF Encoder
  nextimage_avif_encoder_create: {
    parameters: ['pointer'],
    result: 'pointer'
  },
  nextimage_avif_encoder_encode: {
    parameters: ['pointer', 'buffer', 'usize', 'pointer'],
    result: 'i32'
  },
  nextimage_avif_encoder_destroy: {
    parameters: ['pointer'],
    result: 'void'
  },

  // AVIF Decoder
  nextimage_avif_decoder_create: {
    parameters: ['pointer'],
    result: 'pointer'
  },
  nextimage_avif_decoder_decode: {
    parameters: ['pointer', 'buffer', 'usize', 'pointer'],
    result: 'i32'
  },
  nextimage_avif_decoder_destroy: {
    parameters: ['pointer'],
    result: 'void'
  },

  // Memory management
  nextimage_free_buffer: {
    parameters: ['pointer'],
    result: 'void'
  },
  nextimage_free_decode_buffer: {
    parameters: ['pointer'],
    result: 'void'
  },

  // Error handling
  nextimage_get_last_error: {
    parameters: [],
    result: 'pointer'
  }
} as const

// Load the dynamic library
let lib: Deno.DynamicLibrary<typeof symbols> | null = null

export function getLibrary(): Deno.DynamicLibrary<typeof symbols> {
  if (!lib) {
    const libPath = getLibraryPath()
    lib = Deno.dlopen(libPath, symbols)
  }
  return lib
}

/**
 * Helper function to read a C string from pointer
 */
export function readCString(ptr: Deno.PointerValue): string {
  if (!ptr) return ''

  const view = new Deno.UnsafePointerView(ptr)
  return view.getCString()
}

/**
 * Helper function to create a buffer struct in memory
 */
export function createBufferStruct(): { pointer: Deno.PointerValue; view: DataView } {
  // NextImageBuffer struct: { data: *u8, size: usize }
  // On 64-bit: 8 bytes (pointer) + 8 bytes (size) = 16 bytes
  const buffer = new ArrayBuffer(16)
  const view = new DataView(buffer)
  const pointer = Deno.UnsafePointer.of(buffer)

  return { pointer, view }
}

/**
 * Helper function to read NextImageBuffer from memory
 */
export function readBufferStruct(view: DataView): NextImageBuffer {
  // Read pointer (8 bytes on 64-bit)
  const dataLow = view.getUint32(0, true)
  const dataHigh = view.getUint32(4, true)
  const data = Deno.UnsafePointer.create(BigInt(dataLow) | (BigInt(dataHigh) << 32n))

  // Read size (8 bytes on 64-bit)
  const sizeLow = view.getUint32(8, true)
  const sizeHigh = view.getUint32(12, true)
  const size = Number(BigInt(sizeLow) | (BigInt(sizeHigh) << 32n))

  return { data, size }
}

/**
 * Helper function to create a decode buffer struct in memory
 */
export function createDecodeBufferStruct(): { pointer: Deno.PointerValue; view: DataView } {
  // NextImageDecodeBuffer struct: { data: *u8, size: usize, width: u32, height: u32, format: i32 }
  // 8 + 8 + 4 + 4 + 4 = 28 bytes (+ padding to 32)
  const buffer = new ArrayBuffer(32)
  const view = new DataView(buffer)
  const pointer = Deno.UnsafePointer.of(buffer)

  return { pointer, view }
}

/**
 * Helper function to read NextImageDecodeBuffer from memory
 */
export function readDecodeBufferStruct(view: DataView): NextImageDecodeBuffer {
  // Read pointer (8 bytes)
  const dataLow = view.getUint32(0, true)
  const dataHigh = view.getUint32(4, true)
  const data = Deno.UnsafePointer.create(BigInt(dataLow) | (BigInt(dataHigh) << 32n))

  // Read size (8 bytes)
  const sizeLow = view.getUint32(8, true)
  const sizeHigh = view.getUint32(12, true)
  const size = Number(BigInt(sizeLow) | (BigInt(sizeHigh) << 32n))

  // Read width, height, format (4 bytes each)
  const width = view.getUint32(16, true)
  const height = view.getUint32(20, true)
  const format = view.getInt32(24, true)

  return { data, size, width, height, format }
}

/**
 * Helper function to copy data from C memory to Uint8Array
 */
export function copyFromCMemory(ptr: Deno.PointerValue, size: number): Uint8Array {
  if (!ptr || size === 0) {
    return new Uint8Array(0)
  }

  const view = new Deno.UnsafePointerView(ptr)
  return new Uint8Array(view.getArrayBuffer(size))
}

/**
 * Get the last error message from the C library
 */
export function getLastError(): string {
  const lib = getLibrary()
  const errorPtr = lib.symbols.nextimage_get_last_error()
  return readCString(errorPtr)
}
