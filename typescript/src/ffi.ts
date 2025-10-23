/**
 * FFI Foundation - Koffi bindings for libnextimage
 * This module loads the C library and defines all function signatures
 */

import koffi from 'koffi';
import { getLibraryPath } from './library';

/**
 * Load the libnextimage shared library
 */
const libPath = getLibraryPath();
const lib = koffi.load(libPath);

/**
 * Define C struct types for Koffi
 */

// struct NextImageBuffer { uint8_t* data; size_t size; }
const NextImageBufferStruct = koffi.struct('NextImageBuffer', {
  data: koffi.pointer(koffi.types.uint8),
  size: koffi.types.size_t
});

// struct NextImageDecodeBuffer (simplified for now)
const NextImageDecodeBufferStruct = koffi.struct('NextImageDecodeBuffer', {
  data: koffi.pointer(koffi.types.uint8),
  data_capacity: koffi.types.size_t,
  data_size: koffi.types.size_t,
  stride: koffi.types.size_t,

  u_plane: koffi.pointer(koffi.types.uint8),
  u_capacity: koffi.types.size_t,
  u_size: koffi.types.size_t,
  u_stride: koffi.types.size_t,

  v_plane: koffi.pointer(koffi.types.uint8),
  v_capacity: koffi.types.size_t,
  v_size: koffi.types.size_t,
  v_stride: koffi.types.size_t,

  width: koffi.types.int,
  height: koffi.types.int,
  bit_depth: koffi.types.int,
  format: koffi.types.int,
  owns_data: koffi.types.int
});

// struct NextImageWebPEncodeOptions
const NextImageWebPEncodeOptionsStruct = koffi.struct('NextImageWebPEncodeOptions', {
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

// struct NextImageAVIFEncodeOptions
const NextImageAVIFEncodeOptionsStruct = koffi.struct('NextImageAVIFEncodeOptions', {
  // Quality settings
  quality: koffi.types.int,
  quality_alpha: koffi.types.int,
  speed: koffi.types.int,

  // Deprecated quantizer settings
  min_quantizer: koffi.types.int,
  max_quantizer: koffi.types.int,
  min_quantizer_alpha: koffi.types.int,
  max_quantizer_alpha: koffi.types.int,

  // Format settings
  bit_depth: koffi.types.int,
  yuv_format: koffi.types.int,
  yuv_range: koffi.types.int,

  // Alpha settings
  enable_alpha: koffi.types.int,
  premultiply_alpha: koffi.types.int,

  // Tiling settings
  tile_rows_log2: koffi.types.int,
  tile_cols_log2: koffi.types.int,

  // CICP color settings
  color_primaries: koffi.types.int,
  transfer_characteristics: koffi.types.int,
  matrix_coefficients: koffi.types.int,

  // Advanced settings
  sharp_yuv: koffi.types.int,
  target_size: koffi.types.int,

  // Metadata settings (pointers)
  exif_data: koffi.pointer(koffi.types.uint8),
  exif_size: koffi.types.size_t,
  xmp_data: koffi.pointer(koffi.types.uint8),
  xmp_size: koffi.types.size_t,
  icc_data: koffi.pointer(koffi.types.uint8),
  icc_size: koffi.types.size_t,

  // Transformation settings
  irot_angle: koffi.types.int,
  imir_axis: koffi.types.int,

  // Pixel aspect ratio (pasp) - array[2]
  pasp: koffi.array('int', 2),

  // Crop rectangle - array[4]
  crop: koffi.array('int', 4),

  // Clean aperture (clap) - array[8]
  clap: koffi.array('int', 8),

  // Content light level information (clli)
  clli_max_cll: koffi.types.int,
  clli_max_pall: koffi.types.int,

  // Animation settings
  timescale: koffi.types.int,
  keyframe_interval: koffi.types.int
});

// struct NextImageAVIFDecodeOptions
const NextImageAVIFDecodeOptionsStruct = koffi.struct('NextImageAVIFDecodeOptions', {
  // Output format options
  output_format: koffi.types.int,
  jpeg_quality: koffi.types.int,

  // Threading
  use_threads: koffi.types.int,
  format: koffi.types.int,
  ignore_exif: koffi.types.int,
  ignore_xmp: koffi.types.int,
  ignore_icc: koffi.types.int,

  // Security limits
  image_size_limit: koffi.types.uint32,
  image_dimension_limit: koffi.types.uint32,

  // Validation flags
  strict_flags: koffi.types.int,

  // Chroma upsampling
  chroma_upsampling: koffi.types.int,

  // Image manipulation
  crop_x: koffi.types.int,
  crop_y: koffi.types.int,
  crop_width: koffi.types.int,
  crop_height: koffi.types.int,
  use_crop: koffi.types.int,

  resize_width: koffi.types.int,
  resize_height: koffi.types.int,
  use_resize: koffi.types.int
});

// struct NextImageWebPDecodeOptions
const NextImageWebPDecodeOptionsStruct = koffi.struct('NextImageWebPDecodeOptions', {
  output_format: koffi.types.int,
  jpeg_quality: koffi.types.int,
  use_threads: koffi.types.int,
  bypass_filtering: koffi.types.int,
  no_fancy_upsampling: koffi.types.int,
  format: koffi.types.int,
  no_dither: koffi.types.int,
  dither_strength: koffi.types.int,
  alpha_dither: koffi.types.int,
  crop_x: koffi.types.int,
  crop_y: koffi.types.int,
  crop_width: koffi.types.int,
  crop_height: koffi.types.int,
  use_crop: koffi.types.int,
  resize_width: koffi.types.int,
  resize_height: koffi.types.int,
  use_resize: koffi.types.int,
  flip: koffi.types.int,
  alpha_only: koffi.types.int,
  incremental: koffi.types.int
});

/**
 * Common Functions
 */

// void nextimage_free_buffer(NextImageBuffer* buffer)
export const nextimage_free_buffer = lib.func(
  'nextimage_free_buffer',
  koffi.types.void,
  [koffi.pointer(NextImageBufferStruct)]
);

// void nextimage_free_decode_buffer(NextImageDecodeBuffer* buffer)
export const nextimage_free_decode_buffer = lib.func(
  'nextimage_free_decode_buffer',
  koffi.types.void,
  [koffi.pointer(NextImageDecodeBufferStruct)]
);

// const char* nextimage_last_error_message(void)
export const nextimage_last_error_message = lib.func(
  'nextimage_last_error_message',
  koffi.types.str,
  []
);

// void nextimage_clear_error(void)
export const nextimage_clear_error = lib.func(
  'nextimage_clear_error',
  koffi.types.void,
  []
);

/**
 * WebP Functions
 */

// void nextimage_webp_default_encode_options(NextImageWebPEncodeOptions* options)
export const nextimage_webp_default_encode_options = lib.func(
  'nextimage_webp_default_encode_options',
  koffi.types.void,
  [koffi.pointer(NextImageWebPEncodeOptionsStruct)]
);

// void nextimage_webp_default_decode_options(NextImageWebPDecodeOptions* options)
export const nextimage_webp_default_decode_options = lib.func(
  'nextimage_webp_default_decode_options',
  koffi.types.void,
  [koffi.pointer(NextImageWebPDecodeOptionsStruct)]
);

// NextImageStatus nextimage_webp_encode_alloc(const uint8_t* input_data, size_t input_size,
//                                              const NextImageWebPEncodeOptions* options, NextImageBuffer* output)
export const nextimage_webp_encode_alloc = lib.func(
  'nextimage_webp_encode_alloc',
  koffi.types.int,  // NextImageStatus
  [
    koffi.pointer(koffi.types.uint8),
    koffi.types.size_t,
    koffi.pointer(NextImageWebPEncodeOptionsStruct),
    koffi.pointer(NextImageBufferStruct)
  ]
);

// NextImageStatus nextimage_webp_decode_alloc(const uint8_t* webp_data, size_t webp_size,
//                                              const NextImageWebPDecodeOptions* options, NextImageDecodeBuffer* output)
export const nextimage_webp_decode_alloc = lib.func(
  'nextimage_webp_decode_alloc',
  koffi.types.int,  // NextImageStatus
  [
    koffi.pointer(koffi.types.uint8),
    koffi.types.size_t,
    koffi.pointer(NextImageWebPDecodeOptionsStruct),
    koffi.pointer(NextImageDecodeBufferStruct)
  ]
);

// NextImageStatus nextimage_webp_decode_size(const uint8_t* webp_data, size_t webp_size,
//                                             int* width, int* height, size_t* required_size)
export const nextimage_webp_decode_size = lib.func(
  'nextimage_webp_decode_size',
  koffi.types.int,  // NextImageStatus
  [
    koffi.pointer(koffi.types.uint8),
    koffi.types.size_t,
    koffi.pointer(koffi.types.int),
    koffi.pointer(koffi.types.int),
    koffi.pointer(koffi.types.size_t)
  ]
);

/**
 * AVIF Functions
 */

// void nextimage_avif_default_encode_options(NextImageAVIFEncodeOptions* options)
export const nextimage_avif_default_encode_options = lib.func(
  'nextimage_avif_default_encode_options',
  koffi.types.void,
  [koffi.pointer(NextImageAVIFEncodeOptionsStruct)]
);

// void nextimage_avif_default_decode_options(NextImageAVIFDecodeOptions* options)
export const nextimage_avif_default_decode_options = lib.func(
  'nextimage_avif_default_decode_options',
  koffi.types.void,
  [koffi.pointer(NextImageAVIFDecodeOptionsStruct)]
);

// NextImageStatus nextimage_avif_encode_alloc(const uint8_t* input_data, size_t input_size,
//                                              const NextImageAVIFEncodeOptions* options, NextImageBuffer* output)
export const nextimage_avif_encode_alloc = lib.func(
  'nextimage_avif_encode_alloc',
  koffi.types.int,  // NextImageStatus
  [
    koffi.pointer(koffi.types.uint8),
    koffi.types.size_t,
    koffi.pointer(NextImageAVIFEncodeOptionsStruct),
    koffi.pointer(NextImageBufferStruct)
  ]
);

// NextImageStatus nextimage_avif_decode_alloc(const uint8_t* avif_data, size_t avif_size,
//                                              const NextImageAVIFDecodeOptions* options, NextImageDecodeBuffer* output)
export const nextimage_avif_decode_alloc = lib.func(
  'nextimage_avif_decode_alloc',
  koffi.types.int,  // NextImageStatus
  [
    koffi.pointer(koffi.types.uint8),
    koffi.types.size_t,
    koffi.pointer(NextImageAVIFDecodeOptionsStruct),
    koffi.pointer(NextImageDecodeBufferStruct)
  ]
);

// NextImageStatus nextimage_avif_decode_size(const uint8_t* avif_data, size_t avif_size,
//                                             int* width, int* height, int* bit_depth, size_t* required_size)
export const nextimage_avif_decode_size = lib.func(
  'nextimage_avif_decode_size',
  koffi.types.int,  // NextImageStatus
  [
    koffi.pointer(koffi.types.uint8),
    koffi.types.size_t,
    koffi.pointer(koffi.types.int),
    koffi.pointer(koffi.types.int),
    koffi.pointer(koffi.types.int),
    koffi.pointer(koffi.types.size_t)
  ]
);

/**
 * Export struct types for use in high-level APIs
 */
export {
  NextImageBufferStruct,
  NextImageDecodeBufferStruct,
  NextImageWebPEncodeOptionsStruct,
  NextImageWebPDecodeOptionsStruct,
  NextImageAVIFEncodeOptionsStruct,
  NextImageAVIFDecodeOptionsStruct
};
