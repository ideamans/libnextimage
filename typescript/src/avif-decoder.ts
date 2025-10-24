/**
 * AVIF Decoder - Instance-based decoder for efficient batch processing
 */

import koffi from 'koffi';
import {
  nextimage_avif_decoder_create,
  nextimage_avif_decoder_decode,
  nextimage_avif_decoder_destroy,
  nextimage_avif_default_decode_options,
  nextimage_free_decode_buffer,
  nextimage_last_error_message,
  nextimage_clear_error,
  NextImageDecodeBufferStruct,
  NextImageAVIFDecodeOptionsStruct
} from './ffi';
import {
  AVIFDecodeOptions,
  DecodedImage,
  NextImageStatus,
  NextImageError,
  isSuccess,
  normalizePixelFormat
} from './types';

/**
 * FinalizationRegistry to automatically clean up decoder instances if they're garbage collected
 */
const decoderRegistry = new FinalizationRegistry((decoderPtr: any) => {
  if (decoderPtr) {
    nextimage_avif_decoder_destroy(decoderPtr);
  }
});

/**
 * AVIF Decoder class
 *
 * Creates a decoder instance that can be reused for multiple images,
 * reducing initialization overhead.
 *
 * @example
 * ```typescript
 * const decoder = new AVIFDecoder({ format: PixelFormat.RGBA });
 * const decoded1 = decoder.decode(avifData1);
 * const decoded2 = decoder.decode(avifData2);
 * decoder.close();
 * ```
 */
export class AVIFDecoder {
  private decoderPtr: any;
  private closed: boolean = false;

  /**
   * Create a new AVIF decoder with the given options
   * @param options Partial AVIF decoding options (merged with defaults)
   */
  constructor(options: Partial<AVIFDecodeOptions> = {}) {
    const cOptsPtr = this.convertOptions(options);
    this.decoderPtr = nextimage_avif_decoder_create(cOptsPtr);

    if (!this.decoderPtr || koffi.address(this.decoderPtr) === 0n) {
      const errMsg = nextimage_last_error_message();
      nextimage_clear_error(); // Clear error after reading
      throw new NextImageError(
        NextImageStatus.ERROR_OUT_OF_MEMORY,
        `Failed to create AVIF decoder: ${errMsg || 'unknown error'}`
      );
    }

    // Register for automatic cleanup on garbage collection
    decoderRegistry.register(this, this.decoderPtr, this);
  }

  /**
   * Decode AVIF data to pixel data
   * @param avifData Buffer containing AVIF encoded data
   * @returns Decoded image with pixel data
   */
  decode(avifData: Buffer): DecodedImage {
    if (this.closed) {
      throw new Error('Decoder has been closed');
    }

    if (!avifData || avifData.length === 0) {
      throw new NextImageError(NextImageStatus.ERROR_INVALID_PARAM, 'Empty input data');
    }

    const outputPtr = koffi.alloc(NextImageDecodeBufferStruct, 1);

    try {
      const status = nextimage_avif_decoder_decode(
        this.decoderPtr,
        avifData,
        avifData.length,
        outputPtr
      ) as NextImageStatus;

      if (!isSuccess(status)) {
        const errMsg = nextimage_last_error_message();
        nextimage_clear_error(); // Clear error after reading
        throw new NextImageError(status, `AVIF decode failed: ${errMsg || status}`);
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

      return result;
    } finally {
      // Always free C-allocated memory, even on error
      nextimage_free_decode_buffer(outputPtr);
    }
  }

  /**
   * Close the decoder and free resources
   * Must be called when done using the decoder
   */
  close(): void {
    if (!this.closed && this.decoderPtr) {
      // Unregister from finalization registry to avoid double-free
      decoderRegistry.unregister(this);
      nextimage_avif_decoder_destroy(this.decoderPtr);
      this.decoderPtr = null; // Clear pointer to prevent reuse
      this.closed = true;
    }
  }

  /**
   * Get default AVIF decoding options
   */
  static getDefaultOptions(): AVIFDecodeOptions {
    const cOptsPtr = koffi.alloc(NextImageAVIFDecodeOptionsStruct, 1);
    nextimage_avif_default_decode_options(cOptsPtr);
    const cOpts = koffi.decode(cOptsPtr, NextImageAVIFDecodeOptionsStruct) as any;

    return {
      useThreads: cOpts.use_threads !== 0,
      format: cOpts.format,
      ignoreExif: cOpts.ignore_exif !== 0,
      ignoreXMP: cOpts.ignore_xmp !== 0,
      ignoreICC: cOpts.ignore_icc !== 0
    };
  }

  /**
   * Convert TypeScript options to C struct
   */
  private convertOptions(opts: Partial<AVIFDecodeOptions>): any {
    const cOptsPtr = koffi.alloc(NextImageAVIFDecodeOptionsStruct, 1);
    nextimage_avif_default_decode_options(cOptsPtr);

    const cOpts = koffi.decode(cOptsPtr, NextImageAVIFDecodeOptionsStruct) as any;

    // Override with user options
    if (opts.useThreads !== undefined) {
      cOpts.use_threads = opts.useThreads ? 1 : 0;
    }
    if (opts.format !== undefined) {
      cOpts.format = normalizePixelFormat(opts.format);
    }
    if (opts.ignoreExif !== undefined) {
      cOpts.ignore_exif = opts.ignoreExif ? 1 : 0;
    }
    if (opts.ignoreXMP !== undefined) {
      cOpts.ignore_xmp = opts.ignoreXMP ? 1 : 0;
    }
    if (opts.ignoreICC !== undefined) {
      cOpts.ignore_icc = opts.ignoreICC ? 1 : 0;
    }
    if (opts.imageSizeLimit !== undefined) {
      cOpts.image_size_limit = opts.imageSizeLimit;
    }
    if (opts.imageDimensionLimit !== undefined) {
      cOpts.image_dimension_limit = opts.imageDimensionLimit;
    }
    if (opts.strictFlags !== undefined) {
      cOpts.strict_flags = opts.strictFlags;
    }
    if (opts.chromaUpsampling !== undefined) {
      cOpts.chroma_upsampling = opts.chromaUpsampling;
    }

    // Encode back to pointer
    koffi.encode(cOptsPtr, NextImageAVIFDecodeOptionsStruct, cOpts);

    return cOptsPtr;
  }
}
