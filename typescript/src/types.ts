/**
 * Type Definitions for libnextimage
 * These types map to the C library interface (matching golang/webp.go)
 */

/**
 * Status codes returned by libnextimage functions
 */
export enum NextImageStatus {
  OK = 0,
  ERROR_INVALID_PARAM = -1,
  ERROR_ENCODE_FAILED = -2,
  ERROR_DECODE_FAILED = -3,
  ERROR_OUT_OF_MEMORY = -4,
  ERROR_UNSUPPORTED = -5,
  ERROR_BUFFER_TOO_SMALL = -6
}

/**
 * Pixel format types supported by libnextimage
 */
export enum PixelFormat {
  RGBA = 0,      // RGBA 8bit/channel
  RGB = 1,       // RGB 8bit/channel
  BGRA = 2,      // BGRA 8bit/channel
  YUV420 = 3,    // YUV 4:2:0 planar
  YUV422 = 4,    // YUV 4:2:2 planar
  YUV444 = 5     // YUV 4:4:4 planar
}

/**
 * Convert PixelFormat string or number to number
 */
export function normalizePixelFormat(format: PixelFormat | string | number | undefined): number {
  if (format === undefined) {
    return PixelFormat.RGBA;
  }
  if (typeof format === 'number') {
    return format;
  }
  if (typeof format === 'string') {
    const upperFormat = format.toUpperCase();
    switch (upperFormat) {
      case 'RGBA': return PixelFormat.RGBA;
      case 'RGB': return PixelFormat.RGB;
      case 'BGRA': return PixelFormat.BGRA;
      case 'YUV420': return PixelFormat.YUV420;
      case 'YUV422': return PixelFormat.YUV422;
      case 'YUV444': return PixelFormat.YUV444;
      default: throw new Error(`Unknown pixel format: ${format}`);
    }
  }
  return format as number;
}

/**
 * WebP Preset types
 */
export enum WebPPreset {
  DEFAULT = 0,   // default preset
  PICTURE = 1,   // digital picture, like portrait
  PHOTO = 2,     // outdoor photograph
  DRAWING = 3,   // hand or line drawing
  ICON = 4,      // small-sized colorful images
  TEXT = 5       // text-like
}

/**
 * WebP image hint types
 */
export enum WebPImageHint {
  DEFAULT = 0,   // default hint
  PICTURE = 1,   // digital picture, like portrait
  PHOTO = 2,     // outdoor photograph
  GRAPH = 3      // discrete tone image (graph, map-tile)
}

/**
 * WebP encoding options (matching Golang implementation)
 * Simplified version with commonly used options
 */
export interface WebPEncodeOptions {
  // Basic settings
  quality?: number;           // 0-100, default 75
  lossless?: boolean;         // default false
  method?: number;            // 0-6, default 4 (quality/speed trade-off)

  // Preset
  preset?: WebPPreset;        // preset type, default -1 (none)
  imageHint?: WebPImageHint;  // image type hint, default DEFAULT
  exact?: boolean;            // preserve RGB in transparent area, default false
}

/**
 * WebP decoding options (simplified)
 */
export interface WebPDecodeOptions {
  useThreads?: boolean;       // enable multi-threading, default false
  format?: PixelFormat;       // desired pixel format, default RGBA
}

/**
 * AVIF YUV format types
 */
export enum AVIFYUVFormat {
  YUV444 = 0,   // 4:4:4 (no chroma subsampling)
  YUV422 = 1,   // 4:2:2 (horizontal subsampling)
  YUV420 = 2,   // 4:2:0 (both horizontal and vertical subsampling)
  YUV400 = 3    // 4:0:0 (grayscale, no chroma)
}

/**
 * AVIF YUV range types
 */
export enum AVIFYUVRange {
  LIMITED = 0,  // Limited range (16-235 for 8-bit)
  FULL = 1      // Full range (0-255 for 8-bit)
}

/**
 * AVIF encoding options (matching Golang implementation)
 */
export interface AVIFEncodeOptions {
  // Quality settings
  quality?: number;           // 0-100, default 60
  qualityAlpha?: number;      // 0-100, default -1 (use quality if -1)
  speed?: number;             // 0-10, default 6 (0=slowest/best, 10=fastest/worst)

  // Format settings
  bitDepth?: number;          // 8, 10, or 12, default 8
  yuvFormat?: AVIFYUVFormat;  // YUV format, default 444
  yuvRange?: AVIFYUVRange;    // YUV range, default FULL

  // Alpha settings
  enableAlpha?: boolean;      // enable alpha, default true
  premultiplyAlpha?: boolean; // premultiply color by alpha, default false

  // Tiling settings
  tileRowsLog2?: number;      // 0-6, default 0
  tileColsLog2?: number;      // 0-6, default 0

  // CICP color settings
  colorPrimaries?: number;    // CICP color primaries, -1=auto
  transferCharacteristics?: number; // CICP transfer, -1=auto
  matrixCoefficients?: number;      // CICP matrix, -1=auto

  // Advanced settings
  sharpYUV?: boolean;         // use sharp RGB->YUV conversion, default false
  targetSize?: number;        // target file size in bytes, 0=disabled

  // Metadata
  exifData?: Buffer;          // EXIF metadata
  xmpData?: Buffer;           // XMP metadata
  iccData?: Buffer;           // ICC profile

  // Transformation
  irotAngle?: number;         // rotation: 0-3 (90° * angle anti-clockwise), -1=disabled
  imirAxis?: number;          // mirror: 0=vertical, 1=horizontal, -1=disabled
}

/**
 * AVIF decoding options
 */
export interface AVIFDecodeOptions {
  // Threading
  useThreads?: boolean;       // enable multi-threading, default true

  // Output format
  format?: PixelFormat;       // desired pixel format, default RGBA

  // Metadata handling
  ignoreExif?: boolean;       // ignore EXIF metadata, default false
  ignoreXMP?: boolean;        // ignore XMP metadata, default false
  ignoreICC?: boolean;        // ignore ICC profile, default false

  // Security limits
  imageSizeLimit?: number;    // max pixels (default 268435456)
  imageDimensionLimit?: number; // max width/height (default 32768)

  // Validation
  strictFlags?: number;       // strict validation flags, default 1

  // Chroma upsampling
  chromaUpsampling?: number;  // 0=auto, 1=fastest, 2=best, 3=nearest, 4=bilinear
}

/**
 * Decoded image data
 */
export interface DecodedImage {
  data: Buffer;               // Raw pixel data (Y plane for YUV formats)
  width: number;              // Image width in pixels
  height: number;             // Image height in pixels
  stride: number;             // Bytes per row (Y plane stride for YUV)
  format: PixelFormat;        // Pixel format
  bitDepth: number;           // Bit depth (8, 10, 12)

  // YUV plane data (only present for YUV formats)
  uPlane?: Buffer;            // U/Cb plane data
  vPlane?: Buffer;            // V/Cr plane data
  uStride?: number;           // U plane bytes per row
  vStride?: number;           // V plane bytes per row
}

/**
 * Helper: Check if status indicates success
 */
export function isSuccess(status: NextImageStatus): boolean {
  return status === NextImageStatus.OK;
}

/**
 * Helper: Get error message for status code
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
      return 'Unknown error';
  }
}

/**
 * Custom error class for libnextimage operations
 */
export class NextImageError extends Error {
  constructor(
    public status: NextImageStatus,
    message?: string
  ) {
    super(message || getStatusMessage(status));
    this.name = 'NextImageError';
  }
}
