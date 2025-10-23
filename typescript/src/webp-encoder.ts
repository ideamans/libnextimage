/**
 * WebP Encoder - Instance-based encoder for efficient batch processing
 */

import koffi from 'koffi';
import {
  nextimage_webp_encoder_create,
  nextimage_webp_encoder_encode,
  nextimage_webp_encoder_destroy,
  nextimage_webp_default_encode_options,
  nextimage_free_buffer,
  nextimage_last_error_message,
  NextImageBufferStruct,
  NextImageWebPEncodeOptionsStruct
} from './ffi';
import {
  WebPEncodeOptions,
  NextImageStatus,
  NextImageError,
  isSuccess
} from './types';

/**
 * WebP Encoder class
 *
 * Creates an encoder instance that can be reused for multiple images,
 * reducing initialization overhead.
 *
 * @example
 * ```typescript
 * const encoder = new WebPEncoder({ quality: 80, lossless: false });
 * const webp1 = encoder.encode(jpegData1);
 * const webp2 = encoder.encode(pngData);
 * encoder.close();
 * ```
 */
export class WebPEncoder {
  private encoderPtr: any;
  private closed: boolean = false;

  /**
   * Create a new WebP encoder with the given options
   * @param options Partial WebP encoding options (merged with defaults)
   */
  constructor(options: Partial<WebPEncodeOptions> = {}) {
    const cOptsPtr = this.convertOptions(options);
    this.encoderPtr = nextimage_webp_encoder_create(cOptsPtr);

    if (!this.encoderPtr || koffi.address(this.encoderPtr) === 0n) {
      const errMsg = nextimage_last_error_message();
      throw new NextImageError(
        NextImageStatus.ERROR_OUT_OF_MEMORY,
        `Failed to create WebP encoder: ${errMsg || 'unknown error'}`
      );
    }
  }

  /**
   * Encode image file data (JPEG, PNG, etc.) to WebP format
   * @param imageFileData Buffer containing image file data
   * @returns WebP encoded data as Buffer
   */
  encode(imageFileData: Buffer): Buffer {
    if (this.closed) {
      throw new Error('Encoder has been closed');
    }

    if (!imageFileData || imageFileData.length === 0) {
      throw new NextImageError(NextImageStatus.ERROR_INVALID_PARAM, 'Empty input data');
    }

    const outputPtr = koffi.alloc(NextImageBufferStruct, 1);

    const status = nextimage_webp_encoder_encode(
      this.encoderPtr,
      imageFileData,
      imageFileData.length,
      outputPtr
    ) as NextImageStatus;

    if (!isSuccess(status)) {
      const errMsg = nextimage_last_error_message();
      throw new NextImageError(status, `WebP encode failed: ${errMsg || status}`);
    }

    // Decode output struct
    const output = koffi.decode(outputPtr, NextImageBufferStruct) as any;

    if (!output.data || output.size === 0) {
      throw new NextImageError(NextImageStatus.ERROR_ENCODE_FAILED, 'Encoding produced empty output');
    }

    // Copy data from C memory to JavaScript Buffer
    const dataSize = Number(output.size);
    const rawData = koffi.decode(output.data, koffi.array('uint8_t', dataSize));
    const result = Buffer.from(rawData as any);

    // Free C-allocated memory
    nextimage_free_buffer(outputPtr);

    return result;
  }

  /**
   * Close the encoder and free resources
   * Must be called when done using the encoder
   */
  close(): void {
    if (!this.closed && this.encoderPtr) {
      nextimage_webp_encoder_destroy(this.encoderPtr);
      this.closed = true;
    }
  }

  /**
   * Get default WebP encoding options
   */
  static getDefaultOptions(): WebPEncodeOptions {
    const cOptsPtr = koffi.alloc(NextImageWebPEncodeOptionsStruct, 1);
    nextimage_webp_default_encode_options(cOptsPtr);
    const cOpts = koffi.decode(cOptsPtr, NextImageWebPEncodeOptionsStruct) as any;

    return {
      quality: cOpts.quality,
      lossless: cOpts.lossless !== 0,
      method: cOpts.method,
      preset: cOpts.preset,
      imageHint: cOpts.image_hint,
      exact: cOpts.exact !== 0
    };
  }

  /**
   * Convert TypeScript options to C struct
   */
  private convertOptions(opts: Partial<WebPEncodeOptions>): any {
    const cOptsPtr = koffi.alloc(NextImageWebPEncodeOptionsStruct, 1);
    nextimage_webp_default_encode_options(cOptsPtr);

    const cOpts = koffi.decode(cOptsPtr, NextImageWebPEncodeOptionsStruct) as any;

    // Override with user options
    if (opts.quality !== undefined) {
      cOpts.quality = opts.quality;
    }
    if (opts.lossless !== undefined) {
      cOpts.lossless = opts.lossless ? 1 : 0;
    }
    if (opts.method !== undefined) {
      cOpts.method = opts.method;
    }
    if (opts.preset !== undefined) {
      cOpts.preset = opts.preset;
    }
    if (opts.imageHint !== undefined) {
      cOpts.image_hint = opts.imageHint;
    }
    if (opts.exact !== undefined) {
      cOpts.exact = opts.exact ? 1 : 0;
    }

    // Encode back to pointer
    koffi.encode(cOptsPtr, NextImageWebPEncodeOptionsStruct, cOpts);

    return cOptsPtr;
  }
}
