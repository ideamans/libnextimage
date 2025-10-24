/**
 * GIF to WebP Converter - Instance-based converter for GIF to WebP conversion
 */

import koffi from 'koffi';
import { getLibraryPath } from './library';
import { WebPEncodeOptions, NextImageStatus, NextImageError, isSuccess } from './types';
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
const CWebPOptionsStruct = koffi.struct('CWebPOptions', {
  quality: koffi.types.float,
  lossless: koffi.types.int,
  method: koffi.types.int,
  preset: koffi.types.int,
  image_hint: koffi.types.int,
  lossless_preset: koffi.types.int,
  target_size: koffi.types.int,
  target_psnr: koffi.types.float,
  segments: koffi.types.int,
  sns_strength: koffi.types.int,
  filter_strength: koffi.types.int,
  filter_sharpness: koffi.types.int,
  filter_type: koffi.types.int,
  autofilter: koffi.types.int,
  alpha_compression: koffi.types.int,
  alpha_filtering: koffi.types.int,
  alpha_quality: koffi.types.int,
  pass: koffi.types.int,
  show_compressed: koffi.types.int,
  preprocessing: koffi.types.int,
  partitions: koffi.types.int,
  partition_limit: koffi.types.int,
  emulate_jpeg_size: koffi.types.int,
  thread_level: koffi.types.int,
  low_memory: koffi.types.int,
  near_lossless: koffi.types.int,
  exact: koffi.types.int,
  use_delta_palette: koffi.types.int,
  use_sharp_yuv: koffi.types.int,
  qmin: koffi.types.int,
  qmax: koffi.types.int,
  keep_metadata: koffi.types.int,
  crop_x: koffi.types.int,
  crop_y: koffi.types.int,
  crop_width: koffi.types.int,
  crop_height: koffi.types.int,
  resize_width: koffi.types.int,
  resize_height: koffi.types.int,
  resize_mode: koffi.types.int,
  blend_alpha: koffi.types.uint32,
  noalpha: koffi.types.int,
  allow_mixed: koffi.types.int,
  minimize_size: koffi.types.int,
  kmin: koffi.types.int,
  kmax: koffi.types.int,
  anim_loop_count: koffi.types.int,
  loop_compatibility: koffi.types.int
});

// FFI function declarations
const gif2webp_create_default_options = lib.func('gif2webp_create_default_options', koffi.pointer(CWebPOptionsStruct), []);
const gif2webp_free_options = lib.func('gif2webp_free_options', 'void', [koffi.pointer(CWebPOptionsStruct)]);
const gif2webp_new_command = lib.func('gif2webp_new_command', 'void*', [koffi.pointer(CWebPOptionsStruct)]);
const gif2webp_run_command = lib.func('gif2webp_run_command', 'int', [
  'void*',
  koffi.pointer(koffi.types.uint8),
  koffi.types.size_t,
  koffi.out(koffi.pointer(NextImageBufferStruct))
]);
const gif2webp_free_command = lib.func('gif2webp_free_command', 'void', ['void*']);

/**
 * FinalizationRegistry to automatically clean up converter instances
 */
const converterRegistry = new FinalizationRegistry((cmdPtr: any) => {
  if (cmdPtr) {
    gif2webp_free_command(cmdPtr);
  }
});

/**
 * GIF to WebP Converter class
 *
 * Converts GIF images (including animated GIFs) to WebP format.
 *
 * @example
 * ```typescript
 * const converter = new GIF2WebPConverter({ quality: 80 });
 * const webpData = converter.convert(gifData);
 * converter.close();
 * ```
 */
export class GIF2WebPConverter {
  private cmdPtr: any;
  private closed: boolean = false;

  /**
   * Create a new GIF to WebP converter with the given options
   * @param options Partial WebP encoding options (merged with defaults)
   */
  constructor(options: Partial<WebPEncodeOptions> = {}) {
    const cOptsPtr = this.convertOptions(options);
    this.cmdPtr = gif2webp_new_command(cOptsPtr);

    // Free the options structure after command creation
    if (cOptsPtr) {
      gif2webp_free_options(cOptsPtr);
    }

    if (!this.cmdPtr || koffi.address(this.cmdPtr) === 0n) {
      const errMsg = nextimage_last_error_message();
      nextimage_clear_error();
      throw new NextImageError(
        NextImageStatus.ERROR_OUT_OF_MEMORY,
        `Failed to create GIF2WebP converter: ${errMsg || 'unknown error'}`
      );
    }

    // Register for automatic cleanup
    converterRegistry.register(this, this.cmdPtr, this);
  }

  /**
   * Convert GIF data to WebP format
   * @param gifData Buffer containing GIF file data
   * @returns WebP encoded data as Buffer
   */
  convert(gifData: Buffer): Buffer {
    if (this.closed) {
      throw new Error('Converter has been closed');
    }

    if (!gifData || gifData.length === 0) {
      throw new Error('GIF data is empty');
    }

    const outputBufPtr = [{ data: null, size: 0 }];
    const status = gif2webp_run_command(
      this.cmdPtr,
      gifData,
      gifData.length,
      outputBufPtr
    );

    if (!isSuccess(status)) {
      const errMsg = nextimage_last_error_message();
      nextimage_clear_error();
      throw new NextImageError(
        status,
        `GIF to WebP conversion failed: ${errMsg || 'unknown error'}`
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
      gif2webp_free_command(this.cmdPtr);
      this.cmdPtr = null;
      this.closed = true;
    }
  }

  /**
   * Convert WebP options to C struct pointer
   */
  private convertOptions(options: Partial<WebPEncodeOptions>): any {
    const optsPtr = gif2webp_create_default_options();
    if (!optsPtr) {
      throw new Error('Failed to create default options');
    }

    // Decode the C struct to JavaScript object
    const cOpts = koffi.decode(optsPtr, CWebPOptionsStruct) as any;

    // Override with user options
    if (options.quality !== undefined) {
      cOpts.quality = options.quality;
    }
    if (options.lossless !== undefined) {
      cOpts.lossless = options.lossless ? 1 : 0;
    }
    if (options.method !== undefined) {
      cOpts.method = options.method;
    }
    if (options.preset !== undefined) {
      cOpts.preset = options.preset;
    }
    if (options.imageHint !== undefined) {
      cOpts.image_hint = options.imageHint;
    }
    if (options.exact !== undefined) {
      cOpts.exact = options.exact ? 1 : 0;
    }

    // Encode back to C struct pointer
    koffi.encode(optsPtr, CWebPOptionsStruct, cOpts);

    return optsPtr;
  }
}
