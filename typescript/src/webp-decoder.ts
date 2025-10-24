/**
 * WebP Decoder - Instance-based decoder for efficient batch processing
 */

import koffi from 'koffi';
import {
  nextimage_webp_decoder_create,
  nextimage_webp_decoder_decode,
  nextimage_webp_decoder_destroy,
  nextimage_webp_default_decode_options,
  nextimage_free_decode_buffer,
  nextimage_last_error_message,
  NextImageDecodeBufferStruct,
  NextImageWebPDecodeOptionsStruct
} from './ffi';
import {
  WebPDecodeOptions,
  DecodedImage,
  NextImageStatus,
  NextImageError,
  isSuccess,
  normalizePixelFormat
} from './types';

/**
 * WebP Decoder class
 *
 * Creates a decoder instance that can be reused for multiple images,
 * reducing initialization overhead.
 *
 * @example
 * ```typescript
 * const decoder = new WebPDecoder({ format: PixelFormat.RGBA });
 * const decoded1 = decoder.decode(webpData1);
 * const decoded2 = decoder.decode(webpData2);
 * decoder.close();
 * ```
 */
export class WebPDecoder {
  private decoderPtr: any;
  private closed: boolean = false;

  /**
   * Create a new WebP decoder with the given options
   * @param options Partial WebP decoding options (merged with defaults)
   */
  constructor(options: Partial<WebPDecodeOptions> = {}) {
    const cOptsPtr = this.convertOptions(options);
    this.decoderPtr = nextimage_webp_decoder_create(cOptsPtr);

    if (!this.decoderPtr || koffi.address(this.decoderPtr) === 0n) {
      const errMsg = nextimage_last_error_message();
      throw new NextImageError(
        NextImageStatus.ERROR_OUT_OF_MEMORY,
        `Failed to create WebP decoder: ${errMsg || 'unknown error'}`
      );
    }
  }

  /**
   * Decode WebP data to pixel data
   * @param webpData Buffer containing WebP encoded data
   * @returns Decoded image with pixel data
   */
  decode(webpData: Buffer): DecodedImage {
    if (this.closed) {
      throw new Error('Decoder has been closed');
    }

    if (!webpData || webpData.length === 0) {
      throw new NextImageError(NextImageStatus.ERROR_INVALID_PARAM, 'Empty input data');
    }

    const outputPtr = koffi.alloc(NextImageDecodeBufferStruct, 1);

    const status = nextimage_webp_decoder_decode(
      this.decoderPtr,
      webpData,
      webpData.length,
      outputPtr
    ) as NextImageStatus;

    if (!isSuccess(status)) {
      const errMsg = nextimage_last_error_message();
      throw new NextImageError(status, `WebP decode failed: ${errMsg || status}`);
    }

    // Decode output struct
    const output = koffi.decode(outputPtr, NextImageDecodeBufferStruct) as any;

    if (!output.data || output.width === 0 || output.height === 0) {
      throw new NextImageError(NextImageStatus.ERROR_DECODE_FAILED, 'Decoding produced invalid output');
    }

    // Copy data from C memory to JavaScript Buffer
    const dataSize = Number(output.data_size) || (Number(output.stride) * Number(output.height));
    const rawData = koffi.decode(output.data, koffi.array('uint8_t', dataSize));
    const copiedData = Buffer.from(rawData as any);

    // Create result object
    const result: DecodedImage = {
      data: copiedData,
      width: Number(output.width),
      height: Number(output.height),
      stride: Number(output.stride),
      format: Number(output.format),
      bitDepth: Number(output.bit_depth)
    };

    // Free C-allocated memory
    nextimage_free_decode_buffer(outputPtr);

    return result;
  }

  /**
   * Close the decoder and free resources
   * Must be called when done using the decoder
   */
  close(): void {
    if (!this.closed && this.decoderPtr) {
      nextimage_webp_decoder_destroy(this.decoderPtr);
      this.closed = true;
    }
  }

  /**
   * Get default WebP decoding options
   */
  static getDefaultOptions(): WebPDecodeOptions {
    const cOptsPtr = koffi.alloc(NextImageWebPDecodeOptionsStruct, 1);
    nextimage_webp_default_decode_options(cOptsPtr);
    const cOpts = koffi.decode(cOptsPtr, NextImageWebPDecodeOptionsStruct) as any;

    return {
      useThreads: cOpts.use_threads !== 0,
      format: cOpts.format
    };
  }

  /**
   * Convert TypeScript options to C struct
   */
  private convertOptions(opts: Partial<WebPDecodeOptions>): any {
    const cOptsPtr = koffi.alloc(NextImageWebPDecodeOptionsStruct, 1);
    nextimage_webp_default_decode_options(cOptsPtr);

    const cOpts = koffi.decode(cOptsPtr, NextImageWebPDecodeOptionsStruct) as any;

    // Override with user options
    if (opts.useThreads !== undefined) {
      cOpts.use_threads = opts.useThreads ? 1 : 0;
    }
    if (opts.format !== undefined) {
      cOpts.format = normalizePixelFormat(opts.format);
    }

    // Encode back to pointer
    koffi.encode(cOptsPtr, NextImageWebPDecodeOptionsStruct, cOpts);

    return cOptsPtr;
  }
}
