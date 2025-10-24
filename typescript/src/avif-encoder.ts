/**
 * AVIF Encoder - Instance-based encoder for efficient batch processing
 */

import koffi from 'koffi';
import {
  nextimage_avif_encoder_create,
  nextimage_avif_encoder_encode,
  nextimage_avif_encoder_destroy,
  nextimage_avif_default_encode_options,
  nextimage_free_buffer,
  nextimage_last_error_message,
  nextimage_clear_error,
  NextImageBufferStruct,
  NextImageAVIFEncodeOptionsStruct
} from './ffi';
import {
  AVIFEncodeOptions,
  NextImageStatus,
  NextImageError,
  isSuccess
} from './types';

/**
 * FinalizationRegistry to automatically clean up encoder instances if they're garbage collected
 */
const encoderRegistry = new FinalizationRegistry((encoderPtr: any) => {
  if (encoderPtr) {
    nextimage_avif_encoder_destroy(encoderPtr);
  }
});

/**
 * AVIF Encoder class
 *
 * Creates an encoder instance that can be reused for multiple images,
 * reducing initialization overhead.
 *
 * @example
 * ```typescript
 * const encoder = new AVIFEncoder({ quality: 60, speed: 6 });
 * const avif1 = encoder.encode(jpegData1);
 * const avif2 = encoder.encode(pngData);
 * encoder.close();
 * ```
 */
export class AVIFEncoder {
  private encoderPtr: any;
  private closed: boolean = false;

  /**
   * Create a new AVIF encoder with the given options
   * @param options Partial AVIF encoding options (merged with defaults)
   */
  constructor(options: Partial<AVIFEncodeOptions> = {}) {
    const cOptsPtr = this.convertOptions(options);
    this.encoderPtr = nextimage_avif_encoder_create(cOptsPtr);

    if (!this.encoderPtr || koffi.address(this.encoderPtr) === 0n) {
      const errMsg = nextimage_last_error_message();
      nextimage_clear_error(); // Clear error after reading
      throw new NextImageError(
        NextImageStatus.ERROR_OUT_OF_MEMORY,
        `Failed to create AVIF encoder: ${errMsg || 'unknown error'}`
      );
    }

    // Register for automatic cleanup on garbage collection
    encoderRegistry.register(this, this.encoderPtr, this);
  }

  /**
   * Encode image file data (JPEG, PNG, etc.) to AVIF format
   * @param imageFileData Buffer containing image file data
   * @returns AVIF encoded data as Buffer
   */
  encode(imageFileData: Buffer): Buffer {
    if (this.closed) {
      throw new Error('Encoder has been closed');
    }

    if (!imageFileData || imageFileData.length === 0) {
      throw new NextImageError(NextImageStatus.ERROR_INVALID_PARAM, 'Empty input data');
    }

    const outputPtr = koffi.alloc(NextImageBufferStruct, 1);

    try {
      const status = nextimage_avif_encoder_encode(
        this.encoderPtr,
        imageFileData,
        imageFileData.length,
        outputPtr
      ) as NextImageStatus;

      if (!isSuccess(status)) {
        const errMsg = nextimage_last_error_message();
        nextimage_clear_error(); // Clear error after reading
        throw new NextImageError(status, `AVIF encode failed: ${errMsg || status}`);
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

      return result;
    } finally {
      // Always free C-allocated memory, even on error
      nextimage_free_buffer(outputPtr);
    }
  }

  /**
   * Close the encoder and free resources
   * Must be called when done using the encoder
   */
  close(): void {
    if (!this.closed && this.encoderPtr) {
      // Unregister from finalization registry to avoid double-free
      encoderRegistry.unregister(this);
      nextimage_avif_encoder_destroy(this.encoderPtr);
      this.encoderPtr = null; // Clear pointer to prevent reuse
      this.closed = true;
    }
  }

  /**
   * Get default AVIF encoding options
   */
  static getDefaultOptions(): AVIFEncodeOptions {
    const cOptsPtr = koffi.alloc(NextImageAVIFEncodeOptionsStruct, 1);
    nextimage_avif_default_encode_options(cOptsPtr);
    const cOpts = koffi.decode(cOptsPtr, NextImageAVIFEncodeOptionsStruct) as any;

    return {
      quality: cOpts.quality,
      qualityAlpha: cOpts.quality_alpha,
      speed: cOpts.speed,
      bitDepth: cOpts.bit_depth,
      yuvFormat: cOpts.yuv_format,
      yuvRange: cOpts.yuv_range,
      enableAlpha: cOpts.enable_alpha !== 0,
      premultiplyAlpha: cOpts.premultiply_alpha !== 0,
      sharpYUV: cOpts.sharp_yuv !== 0
    };
  }

  /**
   * Convert TypeScript options to C struct
   */
  private convertOptions(opts: Partial<AVIFEncodeOptions>): any {
    const cOptsPtr = koffi.alloc(NextImageAVIFEncodeOptionsStruct, 1);
    nextimage_avif_default_encode_options(cOptsPtr);

    const cOpts = koffi.decode(cOptsPtr, NextImageAVIFEncodeOptionsStruct) as any;

    // Override with user options
    if (opts.quality !== undefined) {
      cOpts.quality = opts.quality;
    }
    if (opts.qualityAlpha !== undefined) {
      cOpts.quality_alpha = opts.qualityAlpha;
    }
    if (opts.speed !== undefined) {
      cOpts.speed = opts.speed;
    }
    if (opts.bitDepth !== undefined) {
      cOpts.bit_depth = opts.bitDepth;
    }
    if (opts.yuvFormat !== undefined) {
      cOpts.yuv_format = opts.yuvFormat;
    }
    if (opts.yuvRange !== undefined) {
      cOpts.yuv_range = opts.yuvRange;
    }
    if (opts.enableAlpha !== undefined) {
      cOpts.enable_alpha = opts.enableAlpha ? 1 : 0;
    }
    if (opts.premultiplyAlpha !== undefined) {
      cOpts.premultiply_alpha = opts.premultiplyAlpha ? 1 : 0;
    }
    if (opts.tileRowsLog2 !== undefined) {
      cOpts.tile_rows_log2 = opts.tileRowsLog2;
    }
    if (opts.tileColsLog2 !== undefined) {
      cOpts.tile_cols_log2 = opts.tileColsLog2;
    }
    if (opts.colorPrimaries !== undefined) {
      cOpts.color_primaries = opts.colorPrimaries;
    }
    if (opts.transferCharacteristics !== undefined) {
      cOpts.transfer_characteristics = opts.transferCharacteristics;
    }
    if (opts.matrixCoefficients !== undefined) {
      cOpts.matrix_coefficients = opts.matrixCoefficients;
    }
    if (opts.sharpYUV !== undefined) {
      cOpts.sharp_yuv = opts.sharpYUV ? 1 : 0;
    }
    if (opts.targetSize !== undefined) {
      cOpts.target_size = opts.targetSize;
    }
    if (opts.irotAngle !== undefined) {
      cOpts.irot_angle = opts.irotAngle;
    }
    if (opts.imirAxis !== undefined) {
      cOpts.imir_axis = opts.imirAxis;
    }

    // Metadata handling
    if (opts.exifData !== undefined && opts.exifData.length > 0) {
      cOpts.exif_data = opts.exifData;
      cOpts.exif_size = opts.exifData.length;
    }
    if (opts.xmpData !== undefined && opts.xmpData.length > 0) {
      cOpts.xmp_data = opts.xmpData;
      cOpts.xmp_size = opts.xmpData.length;
    }
    if (opts.iccData !== undefined && opts.iccData.length > 0) {
      cOpts.icc_data = opts.iccData;
      cOpts.icc_size = opts.iccData.length;
    }

    // Encode back to pointer
    koffi.encode(cOptsPtr, NextImageAVIFEncodeOptionsStruct, cOpts);

    return cOptsPtr;
  }
}
