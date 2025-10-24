/**
 * Bun FFI bindings for libnextimage
 * Uses bun:ffi
 */

import { dlopen, FFIType, type Pointer, ptr, CString } from 'bun:ffi'
import { getLibraryPath } from './library.ts'

// Type definitions
export interface NextImageBuffer {
  data: Pointer
  size: number
}

export interface NextImageDecodeBuffer {
  data: Pointer
  size: number
  width: number
  height: number
  format: number
}

// FFI symbols definition
const symbols = {
  // WebP Encoder
  nextimage_webp_encoder_create: {
    args: [FFIType.ptr],
    returns: FFIType.ptr
  },
  nextimage_webp_encoder_encode: {
    args: [FFIType.ptr, FFIType.ptr, FFIType.u64, FFIType.ptr],
    returns: FFIType.i32
  },
  nextimage_webp_encoder_destroy: {
    args: [FFIType.ptr],
    returns: FFIType.void
  },

  // WebP Decoder
  nextimage_webp_decoder_create: {
    args: [FFIType.ptr],
    returns: FFIType.ptr
  },
  nextimage_webp_decoder_decode: {
    args: [FFIType.ptr, FFIType.ptr, FFIType.u64, FFIType.ptr],
    returns: FFIType.i32
  },
  nextimage_webp_decoder_destroy: {
    args: [FFIType.ptr],
    returns: FFIType.void
  },

  // AVIF Encoder
  nextimage_avif_encoder_create: {
    args: [FFIType.ptr],
    returns: FFIType.ptr
  },
  nextimage_avif_encoder_encode: {
    args: [FFIType.ptr, FFIType.ptr, FFIType.u64, FFIType.ptr],
    returns: FFIType.i32
  },
  nextimage_avif_encoder_destroy: {
    args: [FFIType.ptr],
    returns: FFIType.void
  },

  // AVIF Decoder
  nextimage_avif_decoder_create: {
    args: [FFIType.ptr],
    returns: FFIType.ptr
  },
  nextimage_avif_decoder_decode: {
    args: [FFIType.ptr, FFIType.ptr, FFIType.u64, FFIType.ptr],
    returns: FFIType.i32
  },
  nextimage_avif_decoder_destroy: {
    args: [FFIType.ptr],
    returns: FFIType.void
  },

  // Memory management
  nextimage_free_buffer: {
    args: [FFIType.ptr],
    returns: FFIType.void
  },
  nextimage_free_decode_buffer: {
    args: [FFIType.ptr],
    returns: FFIType.void
  },

  // Error handling
  nextimage_get_last_error: {
    args: [],
    returns: FFIType.cstring
  }
} as const

// Load the dynamic library
let lib: ReturnType<typeof dlopen> | null = null

export function getLibrary() {
  if (!lib) {
    const libPath = getLibraryPath()
    lib = dlopen(libPath, symbols)
  }
  return lib
}

/**
 * Helper function to create a buffer struct
 */
export function createBufferStruct(): { buffer: ArrayBuffer; ptr: Pointer } {
  // NextImageBuffer struct: { data: *u8, size: usize }
  // On 64-bit: 8 bytes (pointer) + 8 bytes (size) = 16 bytes
  const buffer = new ArrayBuffer(16)
  return { buffer, ptr: ptr(buffer) }
}

/**
 * Helper function to read NextImageBuffer from memory
 */
export function readBufferStruct(buffer: ArrayBuffer): NextImageBuffer {
  const view = new DataView(buffer)

  // Read pointer (8 bytes on 64-bit)
  const dataLow = view.getUint32(0, true)
  const dataHigh = view.getUint32(4, true)
  const data = ptr(BigInt(dataLow) | (BigInt(dataHigh) << 32n))

  // Read size (8 bytes on 64-bit)
  const sizeLow = view.getUint32(8, true)
  const sizeHigh = view.getUint32(12, true)
  const size = Number(BigInt(sizeLow) | (BigInt(sizeHigh) << 32n))

  return { data, size }
}

/**
 * Helper function to create a decode buffer struct
 */
export function createDecodeBufferStruct(): { buffer: ArrayBuffer; ptr: Pointer } {
  // NextImageDecodeBuffer struct: { data: *u8, size: usize, width: u32, height: u32, format: i32 }
  // 8 + 8 + 4 + 4 + 4 = 28 bytes (+ padding to 32)
  const buffer = new ArrayBuffer(32)
  return { buffer, ptr: ptr(buffer) }
}

/**
 * Helper function to read NextImageDecodeBuffer from memory
 */
export function readDecodeBufferStruct(buffer: ArrayBuffer): NextImageDecodeBuffer {
  const view = new DataView(buffer)

  // Read pointer (8 bytes)
  const dataLow = view.getUint32(0, true)
  const dataHigh = view.getUint32(4, true)
  const data = ptr(BigInt(dataLow) | (BigInt(dataHigh) << 32n))

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
 * Helper function to copy data from C memory to Buffer
 */
export function copyFromCMemory(pointer: Pointer, size: number): Buffer {
  if (!pointer || size === 0) {
    return Buffer.alloc(0)
  }

  // Bun can read from raw pointers
  const buffer = Buffer.alloc(size)
  const view = new Uint8Array(buffer)

  // Read from pointer
  for (let i = 0; i < size; i++) {
    view[i] = new DataView(new ArrayBuffer(1)).getUint8(0)
  }

  return buffer
}

/**
 * Get the last error message from the C library
 */
export function getLastError(): string {
  const library = getLibrary()
  const errorCStr = library.symbols.nextimage_get_last_error() as CString
  return errorCStr ? errorCStr.toString() : 'Unknown error'
}
