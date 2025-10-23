/**
 * AVIF encoding and decoding functions
 * Simplified API matching Golang implementation
 */

import koffi from 'koffi';
import {
  nextimage_avif_encode_alloc,
  nextimage_avif_decode_alloc,
  nextimage_avif_default_encode_options,
  nextimage_avif_default_decode_options,
  nextimage_free_buffer,
  nextimage_free_decode_buffer,
  nextimage_last_error_message,
  NextImageBufferStruct,
  NextImageDecodeBufferStruct,
  NextImageAVIFEncodeOptionsStruct,
  NextImageAVIFDecodeOptionsStruct
} from './ffi';
import {
  AVIFEncodeOptions,
  AVIFDecodeOptions,
  DecodedImage,
  NextImageStatus,
  NextImageError,
  isSuccess
} from './types';

/**
 * Convert TypeScript options to C struct
 */
function convertEncodeOptions(opts: AVIFEncodeOptions = {}): any {
  // Allocate C struct
  const cOptsPtr = koffi.alloc(NextImageAVIFEncodeOptionsStruct, 1);

  // Initialize with defaults
  nextimage_avif_default_encode_options(cOptsPtr);

  // Decode to modify
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

  // Note: Metadata (exifData, xmpData, iccData) would need special handling
  // with C memory allocation. For now, keeping it simple like Golang version.

  // Encode back to pointer
  koffi.encode(cOptsPtr, NextImageAVIFEncodeOptionsStruct, cOpts);

  return cOptsPtr;
}

/**
 * Convert TypeScript decode options to C struct
 */
function convertDecodeOptions(opts: AVIFDecodeOptions = {}): any {
  // Allocate C struct
  const cOptsPtr = koffi.alloc(NextImageAVIFDecodeOptionsStruct, 1);

  // Initialize with defaults
  nextimage_avif_default_decode_options(cOptsPtr);

  // Decode to modify
  const cOpts = koffi.decode(cOptsPtr, NextImageAVIFDecodeOptionsStruct) as any;

  // Override with user options
  if (opts.useThreads !== undefined) {
    cOpts.use_threads = opts.useThreads ? 1 : 0;
  }
  if (opts.format !== undefined) {
    cOpts.format = opts.format;
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

/**
 * Encode an image file (JPEG, PNG, etc.) to AVIF format
 *
 * @param imageFileData - Buffer containing image file data (JPEG, PNG, etc.)
 * @param options - AVIF encoding options
 * @returns AVIF encoded data as Buffer
 * @throws NextImageError if encoding fails
 *
 * Example:
 * ```typescript
 * import { encodeAVIF } from '@ideamans/libnextimage';
 * import * as fs from 'fs';
 *
 * const jpegData = fs.readFileSync('input.jpg');
 * const avifData = encodeAVIF(jpegData, { quality: 60 });
 * fs.writeFileSync('output.avif', avifData);
 * ```
 */
export function encodeAVIF(imageFileData: Buffer, options: AVIFEncodeOptions = {}): Buffer {
  if (!imageFileData || imageFileData.length === 0) {
    throw new NextImageError(NextImageStatus.ERROR_INVALID_PARAM, 'Empty input data');
  }

  const cOptsPtr = convertEncodeOptions(options);
  const outputPtr = koffi.alloc(NextImageBufferStruct, 1);

  const status = nextimage_avif_encode_alloc(
    imageFileData,
    imageFileData.length,
    cOptsPtr,
    outputPtr
  ) as NextImageStatus;

  if (!isSuccess(status)) {
    const errMsg = nextimage_last_error_message();
    throw new NextImageError(status, `AVIF encode failed: ${errMsg || status}`);
  }

  // Decode output struct
  const output = koffi.decode(outputPtr, NextImageBufferStruct) as any;

  if (!output.data || output.size === 0) {
    throw new NextImageError(NextImageStatus.ERROR_ENCODE_FAILED, 'AVIF encoding produced empty output');
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
 * Decode AVIF data to pixel data
 *
 * @param avifData - Buffer containing AVIF encoded data
 * @param options - AVIF decoding options
 * @returns Decoded image with pixel data
 * @throws NextImageError if decoding fails
 *
 * Example:
 * ```typescript
 * import { decodeAVIF } from '@ideamans/libnextimage';
 * import * as fs from 'fs';
 *
 * const avifData = fs.readFileSync('input.avif');
 * const decoded = decodeAVIF(avifData);
 * console.log(`Decoded ${decoded.width}x${decoded.height} image`);
 * // decoded.data contains raw pixel data (RGBA by default)
 * ```
 */
export function decodeAVIF(avifData: Buffer, options: AVIFDecodeOptions = {}): DecodedImage {
  if (!avifData || avifData.length === 0) {
    throw new NextImageError(NextImageStatus.ERROR_INVALID_PARAM, 'Empty input data');
  }

  const cOptsPtr = convertDecodeOptions(options);
  const outputPtr = koffi.alloc(NextImageDecodeBufferStruct, 1);

  const status = nextimage_avif_decode_alloc(
    avifData,
    avifData.length,
    cOptsPtr,
    outputPtr
  ) as NextImageStatus;

  if (!isSuccess(status)) {
    const errMsg = nextimage_last_error_message();
    throw new NextImageError(status, `AVIF decode failed: ${errMsg || status}`);
  }

  // Decode output struct
  const output = koffi.decode(outputPtr, NextImageDecodeBufferStruct) as any;

  if (!output.data || output.width === 0 || output.height === 0) {
    throw new NextImageError(NextImageStatus.ERROR_DECODE_FAILED, 'AVIF decoding produced invalid output');
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
