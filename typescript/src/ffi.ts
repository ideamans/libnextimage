/**
 * FFI bindings for libnextimage using koffi
 */

import koffi from 'koffi';
import { getLibraryPath } from './library';

// ========================================
// Type Definitions
// ========================================

/**
 * NextImageStatus enum
 */
export enum NextImageStatus {
  OK = 0,
  ERROR_INVALID_PARAM = -1,
  ERROR_ENCODE_FAILED = -2,
  ERROR_DECODE_FAILED = -3,
  ERROR_OUT_OF_MEMORY = -4,
  ERROR_UNSUPPORTED = -5,
  ERROR_BUFFER_TOO_SMALL = -6,
}

/**
 * NextImageBuffer struct (C definition)
 * struct { uint8_t* data; size_t size; }
 */
export interface NextImageBuffer {
  data: Buffer | null;
  size: number;
}

// ========================================
// Koffi Type Definitions
// ========================================

// NextImageBuffer as C struct
const NextImageBufferStruct = koffi.struct('NextImageBuffer', {
  data: koffi.pointer('uint8_t'),
  size: 'size_t',
});

const NextImageBufferPtr = koffi.pointer(NextImageBufferStruct);

// ========================================
// Library Loading
// ========================================

let libraryInstance: any = null;

export function getLibrary() {
  if (libraryInstance) {
    return libraryInstance;
  }

  const libPath = getLibraryPath();
  libraryInstance = koffi.load(libPath);

  return libraryInstance;
}

// ========================================
// Function Bindings
// ========================================

let nextimage_webp_encode_alloc: any = null;
let nextimage_free_buffer: any = null;

/**
 * Initialize function bindings (lazy)
 */
function initBindings() {
  if (nextimage_webp_encode_alloc) {
    return;
  }

  const lib = getLibrary();

  // int nextimage_webp_encode_alloc(const uint8_t* input_data, size_t input_size,
  //                                  const void* options, NextImageBuffer* output)
  nextimage_webp_encode_alloc = lib.func('nextimage_webp_encode_alloc', 'int', [
    koffi.pointer('uint8_t'),
    'size_t',
    koffi.pointer('void'),
    NextImageBufferPtr,
  ]);

  // void nextimage_free_buffer(NextImageBuffer* buffer)
  nextimage_free_buffer = lib.func('nextimage_free_buffer', 'void', [
    NextImageBufferPtr,
  ]);
}

/**
 * Encode image to WebP
 */
export function encodeWebPAlloc(inputData: Buffer): { status: number; output: NextImageBuffer | null } {
  initBindings();

  // Allocate output buffer struct on the heap
  const outputBufferPtr = koffi.alloc(NextImageBufferStruct, 1);

  // Call encode function
  const status = nextimage_webp_encode_alloc(
    inputData,
    inputData.length,
    null, // options (NULL for default)
    outputBufferPtr
  );

  if (status !== NextImageStatus.OK) {
    return { status, output: null };
  }

  // Decode the struct from the pointer
  const outputStruct = koffi.decode(outputBufferPtr, NextImageBufferStruct);

  // The data pointer needs to be read using koffi.decode with the correct size
  const dataSize = Number(outputStruct.size);
  let dataBuffer: Buffer | null = null;

  if (outputStruct.data && dataSize > 0) {
    // Read the data from the pointer
    const rawData = koffi.decode(outputStruct.data, koffi.array('uint8_t', dataSize)) as any;
    // IMPORTANT: Copy the data before we free the C buffer, because koffi may give us
    // a Buffer that points directly to C memory. If we don't copy, the Buffer will
    // reference freed memory after freeBuffer() is called.
    if (Buffer.isBuffer(rawData)) {
      // Create a copy, not a reference
      dataBuffer = Buffer.from(rawData);
    } else if (rawData) {
      dataBuffer = Buffer.from(rawData);
    }
  }

  const result: NextImageBuffer = {
    data: dataBuffer,
    size: dataSize,
  };

  // Free the C-allocated buffer immediately, before returning
  // Now that we've copied the data, it's safe to free the C memory
  // We pass the original pointer directly to the C free function
  nextimage_free_buffer(outputBufferPtr);

  return { status, output: result };
}

/**
 * Free a NextImageBuffer allocated by the library
 */
export function freeBuffer(buffer: NextImageBuffer): void {
  initBindings();

  // Allocate a struct and encode the buffer into it
  const bufferPtr = koffi.alloc(NextImageBufferStruct, 1);
  koffi.encode(bufferPtr, NextImageBufferStruct, buffer);

  nextimage_free_buffer(bufferPtr);
}

/**
 * Get status message from code
 */
export function getStatusMessage(status: NextImageStatus): string {
  switch (status) {
    case NextImageStatus.OK:
      return 'Success';
    case NextImageStatus.ERROR_INVALID_PARAM:
      return 'Invalid parameter';
    case NextImageStatus.ERROR_ENCODE_FAILED:
      return 'Encoding failed';
    case NextImageStatus.ERROR_DECODE_FAILED:
      return 'Decoding failed';
    case NextImageStatus.ERROR_OUT_OF_MEMORY:
      return 'Out of memory';
    case NextImageStatus.ERROR_UNSUPPORTED:
      return 'Unsupported operation';
    case NextImageStatus.ERROR_BUFFER_TOO_SMALL:
      return 'Buffer too small';
    default:
      return `Unknown error (${status})`;
  }
}
