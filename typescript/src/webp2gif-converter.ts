/**
 * WebP to GIF Converter - Instance-based converter for WebP to GIF conversion
 */

import koffi from 'koffi';
import { getLibraryPath } from './library';
import { NextImageStatus, NextImageError, isSuccess } from './types';
import {
  nextimage_free_buffer,
  nextimage_last_error_message,
  nextimage_clear_error,
  NextImageBufferStruct
} from './ffi';

// Load library
const libPath = getLibraryPath();
const lib = koffi.load(libPath);

// Structures
const WebP2GifOptionsStruct = koffi.struct('WebP2GifOptions', {
  reserved: koffi.types.int
});

// FFI function declarations
const webp2gif_create_default_options = lib.func('webp2gif_create_default_options', koffi.pointer(WebP2GifOptionsStruct), []);
const webp2gif_free_options = lib.func('webp2gif_free_options', 'void', [koffi.pointer(WebP2GifOptionsStruct)]);
const webp2gif_new_command = lib.func('webp2gif_new_command', 'void*', [koffi.pointer(WebP2GifOptionsStruct)]);
const webp2gif_run_command = lib.func('webp2gif_run_command', 'int', [
  'void*',
  koffi.pointer(koffi.types.uint8),
  koffi.types.size_t,
  koffi.out(koffi.pointer(NextImageBufferStruct))
]);
const webp2gif_free_command = lib.func('webp2gif_free_command', 'void', ['void*']);

/**
 * FinalizationRegistry to automatically clean up converter instances
 */
const converterRegistry = new FinalizationRegistry((cmdPtr: any) => {
  if (cmdPtr) {
    webp2gif_free_command(cmdPtr);
  }
});

/**
 * WebP to GIF Converter options
 * Currently minimal - reserved for future extensions
 */
export interface WebP2GIFOptions {
  // Reserved for future use
  reserved?: number;
}

/**
 * WebP to GIF Converter class
 *
 * Converts WebP images to GIF format.
 *
 * @example
 * ```typescript
 * const converter = new WebP2GIFConverter();
 * const gifData = converter.convert(webpData);
 * converter.close();
 * ```
 */
export class WebP2GIFConverter {
  private cmdPtr: any;
  private closed: boolean = false;

  /**
   * Create a new WebP to GIF converter
   * @param options Converter options (currently minimal)
   */
  constructor(options: WebP2GIFOptions = {}) {
    const cOptsPtr = webp2gif_create_default_options();

    // Apply user options if provided
    if (cOptsPtr && options.reserved !== undefined) {
      const cOpts = koffi.decode(cOptsPtr, WebP2GifOptionsStruct) as any;
      cOpts.reserved = options.reserved;
      koffi.encode(cOptsPtr, WebP2GifOptionsStruct, cOpts);
    }

    this.cmdPtr = webp2gif_new_command(cOptsPtr);

    // Free the options structure as it's been copied
    if (cOptsPtr) {
      webp2gif_free_options(cOptsPtr);
    }

    if (!this.cmdPtr || koffi.address(this.cmdPtr) === 0n) {
      const errMsg = nextimage_last_error_message();
      nextimage_clear_error();
      throw new NextImageError(
        NextImageStatus.ERROR_OUT_OF_MEMORY,
        `Failed to create WebP2GIF converter: ${errMsg || 'unknown error'}`
      );
    }

    // Register for automatic cleanup
    converterRegistry.register(this, this.cmdPtr, this);
  }

  /**
   * Convert WebP data to GIF format
   * @param webpData Buffer containing WebP file data
   * @returns GIF encoded data as Buffer
   */
  convert(webpData: Buffer): Buffer {
    if (this.closed) {
      throw new Error('Converter has been closed');
    }

    if (!webpData || webpData.length === 0) {
      throw new Error('WebP data is empty');
    }

    const outputBufPtr = [{ data: null, size: 0 }];
    const status = webp2gif_run_command(
      this.cmdPtr,
      webpData,
      webpData.length,
      outputBufPtr
    );

    if (!isSuccess(status)) {
      const errMsg = nextimage_last_error_message();
      nextimage_clear_error();
      throw new NextImageError(
        status,
        `WebP to GIF conversion failed: ${errMsg || 'unknown error'}`
      );
    }

    const outputBuf = outputBufPtr[0];
    if (!outputBuf.data || outputBuf.size === 0) {
      throw new Error('Conversion produced empty output');
    }

    // Copy data to Node.js Buffer
    const result = Buffer.from(koffi.decode(outputBuf.data, 'uint8', outputBuf.size));

    // Free C buffer
    nextimage_free_buffer(outputBuf);

    return result;
  }

  /**
   * Close the converter and release resources
   */
  close(): void {
    if (!this.closed && this.cmdPtr) {
      converterRegistry.unregister(this);
      webp2gif_free_command(this.cmdPtr);
      this.cmdPtr = null;
      this.closed = true;
    }
  }
}
