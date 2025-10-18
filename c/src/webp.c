#include "webp.h"
#include "internal.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

// libwebp headers
#include "webp/encode.h"
#include "webp/decode.h"
#include "webp/demux.h"
#include "webp/mux.h"

// imageio headers for reading JPEG/PNG/etc
#include "image_dec.h"

// giflib header
#include <gif_lib.h>

// GIF helper functions from libwebp examples
#include "../../deps/libwebp/examples/gifdec.h"

// デフォルトエンコードオプション (全WebPConfigフィールドに対応)
void nextimage_webp_default_encode_options(NextImageWebPEncodeOptions* options) {
    if (!options) return;

    memset(options, 0, sizeof(NextImageWebPEncodeOptions));

    // 基本設定
    options->quality = 75.0f;
    options->lossless = 0;
    options->method = 4;

    // プリセット
    options->preset = -1;  // -1 = none (use manual config)
    options->image_hint = NEXTIMAGE_WEBP_HINT_DEFAULT;
    options->lossless_preset = -1;  // -1 = don't use preset

    // ターゲット設定
    options->target_size = 0;
    options->target_psnr = 0.0f;

    // セグメント/フィルタ設定 (WebPConfigの初期値)
    options->segments = 4;
    options->sns_strength = 50;
    options->filter_strength = 60;
    options->filter_sharpness = 0;
    options->filter_type = 1;  // strong filter
    options->autofilter = 0;

    // アルファチャンネル設定
    options->alpha_compression = 1;
    options->alpha_filtering = 1;  // fast
    options->alpha_quality = 100;

    // エントロピー設定
    options->pass = 1;

    // その他の設定
    options->show_compressed = 0;
    options->preprocessing = 0;
    options->partitions = 0;
    options->partition_limit = 0;
    options->emulate_jpeg_size = 0;
    options->thread_level = 0;
    options->low_memory = 0;
    options->near_lossless = -1;  // -1 = not set, 0-100 = use near lossless (auto-enables lossless)
    options->exact = 0;
    options->use_delta_palette = 0;
    options->use_sharp_yuv = 0;
    options->qmin = 0;
    options->qmax = 100;

    // メタデータ設定
    options->keep_metadata = -1;  // default (none for compatibility)

    // 画像変換設定
    options->crop_x = -1;         // -1 = disabled
    options->crop_y = -1;
    options->crop_width = -1;
    options->crop_height = -1;
    options->resize_width = -1;   // -1 = disabled
    options->resize_height = -1;
    options->resize_mode = 0;     // 0 = always (default)

    // アルファチャンネル特殊処理
    options->blend_alpha = (uint32_t)-1;  // -1 = disabled
    options->noalpha = 0;         // default: keep alpha

    // アニメーション設定
    options->allow_mixed = 0;     // default: no mixed mode
    options->minimize_size = 0;   // default: off (faster)
    options->kmin = -1;           // -1 = auto (will be set based on lossless)
    options->kmax = -1;           // -1 = auto (will be set based on lossless)
    options->anim_loop_count = 0; // default: infinite loop
    options->loop_compatibility = 0; // default: off
}

// デフォルトデコードオプション (全dwebpオプションに対応)
void nextimage_webp_default_decode_options(NextImageWebPDecodeOptions* options) {
    if (!options) return;

    memset(options, 0, sizeof(NextImageWebPDecodeOptions));

    // 基本設定
    options->use_threads = 0;
    options->bypass_filtering = 0;
    options->no_fancy_upsampling = 0;
    options->format = NEXTIMAGE_FORMAT_RGBA;

    // ディザリング設定
    options->no_dither = 0;
    options->dither_strength = 100;
    options->alpha_dither = 0;

    // 画像操作
    options->crop_x = 0;
    options->crop_y = 0;
    options->crop_width = 0;
    options->crop_height = 0;
    options->use_crop = 0;

    options->resize_width = 0;
    options->resize_height = 0;
    options->use_resize = 0;

    options->flip = 0;

    // 特殊モード
    options->alpha_only = 0;
    options->incremental = 0;
}

// WebP config を NextImageWebPEncodeOptions から設定 (全フィールド対応)
static int setup_webp_config(WebPConfig* config, const NextImageWebPEncodeOptions* options) {
    if (!options) {
        // No options provided, use default
        if (!WebPConfigInit(config)) {
            nextimage_set_error("Failed to initialize WebP config");
            return 0;
        }
        return 1;
    }

    // Check if preset is specified (-preset flag in cwebp)
    int using_preset = 0;
    int using_lossless_preset = 0;

    // Always initialize config first to ensure all fields have valid defaults
    if (options->preset >= 0 && options->preset <= NEXTIMAGE_WEBP_PRESET_TEXT) {
        // Use preset initialization (quality is passed to preset)
        if (!WebPConfigPreset(config, (WebPPreset)options->preset, options->quality)) {
            nextimage_set_error("Failed to initialize WebP config with preset %d", options->preset);
            return 0;
        }
        using_preset = 1;
    } else {
        // Use default initialization
        if (!WebPConfigInit(config)) {
            nextimage_set_error("Failed to initialize WebP config");
            return 0;
        }
    }

    // Apply lossless preset after initialization if specified (-z flag in cwebp)
    if (options->lossless_preset >= 0 && options->lossless_preset <= 9) {
        if (!WebPConfigLosslessPreset(config, options->lossless_preset)) {
            nextimage_set_error("Invalid lossless preset level: %d", options->lossless_preset);
            return 0;
        }
        using_lossless_preset = 1;
    }

    // When using preset or lossless preset, some parameters are already set by the preset
    // We should only override them if the user explicitly provided non-default values
    //
    // Strategy: When using a preset, only override parameters that differ from our C defaults
    // This allows user-specified values to override preset values, while keeping preset
    // values for any parameters that weren't explicitly set by the user
    if (!using_preset && !using_lossless_preset) {
        // No preset: use all options values
        config->lossless = options->lossless;
        config->quality = options->quality;
        config->method = options->method;
        config->segments = options->segments;
        config->sns_strength = options->sns_strength;
        config->filter_strength = options->filter_strength;
        config->filter_sharpness = options->filter_sharpness;
        config->filter_type = options->filter_type;
        config->autofilter = options->autofilter;
    } else if (using_preset) {
        // When using preset, only override if value differs from C defaults
        // This allows user-specified non-default values to override the preset
        if (options->lossless != 0) {  // default is 0
            config->lossless = options->lossless;
        }
        // Quality from preset is already set, don't override unless user specified different value
        // Note: preset quality is already applied in WebPConfigPreset()
        // We don't override it even if options->quality == 75, because that's the preset's quality too
        if (options->method != 4) {  // default is 4
            config->method = options->method;
        }
        if (options->segments != 4) {  // default is 4
            config->segments = options->segments;
        }
        if (options->sns_strength != 50) {  // default is 50
            config->sns_strength = options->sns_strength;
        }
        if (options->filter_strength != 60) {  // default is 60
            config->filter_strength = options->filter_strength;
        }
        if (options->filter_sharpness != 0) {  // default is 0
            config->filter_sharpness = options->filter_sharpness;
        }
        if (options->filter_type != 1) {  // default is 1
            config->filter_type = options->filter_type;
        }
        if (options->autofilter != 0) {  // default is 0
            config->autofilter = options->autofilter;
        }
    }
    // When using lossless_preset, don't override anything as WebPConfigLosslessPreset sets everything

    // Only set image_hint if not using preset (preset sets its own hint)
    // or if user explicitly specified a non-default hint
    if (!using_preset || options->image_hint != NEXTIMAGE_WEBP_HINT_DEFAULT) {
        config->image_hint = (WebPImageHint)options->image_hint;
    }

    // These parameters are always safe to set (not affected by presets)
    config->target_size = options->target_size;
    config->target_PSNR = options->target_psnr;

    config->alpha_compression = options->alpha_compression;
    config->alpha_filtering = options->alpha_filtering;
    config->alpha_quality = options->alpha_quality;

    config->pass = options->pass;

    config->show_compressed = options->show_compressed;

    // Only override preprocessing if not using preset OR if value differs from default
    if (!using_preset) {
        config->preprocessing = options->preprocessing;
    } else if (options->preprocessing != 0) {  // default is 0
        // User explicitly set preprocessing to non-default value
        config->preprocessing = options->preprocessing;
    }
    // Otherwise: using preset and options->preprocessing == 0 (default), so keep preset's value

    config->partitions = options->partitions;
    config->partition_limit = options->partition_limit;
    config->emulate_jpeg_size = options->emulate_jpeg_size;
    config->thread_level = options->thread_level;
    config->low_memory = options->low_memory;

    // near_lossless only works in lossless mode
    // If near_lossless is explicitly set (not the default -1), enable lossless mode
    // This matches cwebp behavior where -near_lossless flag automatically enables lossless
    // even when the value is 100 (off)
    if (options->near_lossless >= 0 && options->near_lossless <= 100) {
        config->lossless = 1;
        config->near_lossless = options->near_lossless;
    }

    config->exact = options->exact;

    // Only override use_delta_palette if not using preset OR if value differs from default
    if (!using_preset) {
        config->use_delta_palette = options->use_delta_palette;
    } else if (options->use_delta_palette != 0) {  // default is 0
        config->use_delta_palette = options->use_delta_palette;
    }

    config->use_sharp_yuv = options->use_sharp_yuv;
    config->qmin = options->qmin;
    config->qmax = options->qmax;

    if (!WebPValidateConfig(config)) {
        nextimage_set_error("Invalid WebP configuration");
        return 0;
    }

    return 1;
}

// メモリライターコールバック
static int webp_memory_writer(const uint8_t* data, size_t data_size, const WebPPicture* picture) {
    NextImageEncodeBuffer* output = (NextImageEncodeBuffer*)picture->custom_ptr;

    // Reallocate buffer
    size_t new_size = output->size + data_size;
    uint8_t* new_data = (uint8_t*)nextimage_realloc(output->data, new_size);
    if (!new_data) {
        return 0;
    }

    memcpy(new_data + output->size, data, data_size);
    output->data = new_data;
    output->size = new_size;

    return 1;
}

// エンコード実装（画像ファイルデータから）
NextImageStatus nextimage_webp_encode_alloc(
    const uint8_t* input_data,
    size_t input_size,
    const NextImageWebPEncodeOptions* options,
    NextImageEncodeBuffer* output
) {
    if (!input_data || input_size == 0 || !output) {
        nextimage_set_error("Invalid parameters: NULL input or output");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Clear output
    memset(output, 0, sizeof(NextImageEncodeBuffer));

    // 画像フォーマットを推測
    WebPInputFileFormat format = WebPGuessImageType(input_data, input_size);
    if (format == WEBP_UNSUPPORTED_FORMAT) {
        nextimage_set_error("Unsupported or unrecognized image format");
        return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    // 適切なリーダーを取得
    WebPImageReader reader = WebPGetImageReader(format);
    if (!reader) {
        nextimage_set_error("No reader available for this image format");
        return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    // WebPPictureを初期化
    WebPPicture picture;
    if (!WebPPictureInit(&picture)) {
        nextimage_set_error("Failed to initialize WebPPicture");
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // WebP設定を先に行う（use_argbを決定するため）
    WebPConfig config;
    if (!setup_webp_config(&config, options)) {
        WebPPictureFree(&picture);
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Set use_argb BEFORE reading the image (matches cwebp.c line 1030)
    // We need to decide if we prefer ARGB or YUVA samples, depending on the
    // expected compression mode (this saves some conversion steps)
    picture.use_argb = (config.lossless || config.use_sharp_yuv ||
                        config.preprocessing > 0);

    // 画像を読み込む（keep_alpha=1, metadata=NULL）
    // noalpha オプションが有効な場合は keep_alpha=0 で読み込む
    int keep_alpha = (options && options->noalpha) ? 0 : 1;
    if (!reader(input_data, input_size, &picture, keep_alpha, NULL)) {
        WebPPictureFree(&picture);
        nextimage_set_error("Failed to read input image");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // 画像変換処理: crop, resize, blend_alpha (cwebp.c と同じ順序)
    if (options) {
        // 1. Crop処理 (cwebp.c line 1065-1073)
        if (options->crop_x >= 0 && options->crop_y >= 0 &&
            options->crop_width > 0 && options->crop_height > 0) {
            if (!WebPPictureCrop(&picture, options->crop_x, options->crop_y,
                                 options->crop_width, options->crop_height)) {
                WebPPictureFree(&picture);
                nextimage_set_error("Crop failed (invalid crop dimensions)");
                return NEXTIMAGE_ERROR_INVALID_PARAM;
            }
        }

        // 2. Resize処理 (cwebp.c line 1075-1091)
        if (options->resize_width > 0 && options->resize_height > 0) {
            int should_resize = 1;
            int orig_width = picture.width;
            int orig_height = picture.height;

            // resize_mode による条件チェック
            if (options->resize_mode == 1) {  // up_only
                should_resize = (options->resize_width > orig_width ||
                                options->resize_height > orig_height);
            } else if (options->resize_mode == 2) {  // down_only
                should_resize = (options->resize_width < orig_width ||
                                options->resize_height < orig_height);
            }
            // mode==0 (always) の場合は常にリサイズ

            if (should_resize) {
                if (!WebPPictureRescale(&picture, options->resize_width, options->resize_height)) {
                    WebPPictureFree(&picture);
                    nextimage_set_error("Resize failed");
                    return NEXTIMAGE_ERROR_ENCODE_FAILED;
                }
            }
        }

        // 3. Blend alpha処理 (cwebp.c line 1093-1104)
        if (options->blend_alpha != (uint32_t)-1 && picture.use_argb) {
            // WebPBlendAlpha expects 0xRRGGBB format
            // options->blend_alpha is already in 0xRRGGBB format
            WebPBlendAlpha(&picture, options->blend_alpha);
        }
    }

    // カスタムライターを設定
    picture.writer = webp_memory_writer;
    picture.custom_ptr = output;

    // エンコード
    if (!WebPEncode(&config, &picture)) {
        WebPPictureFree(&picture);
        if (output->data) {
            nextimage_free(output->data);
            output->data = NULL;
            output->size = 0;
        }
        nextimage_set_error("WebP encoding failed: %d", picture.error_code);
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    WebPPictureFree(&picture);
    return NEXTIMAGE_OK;
}

// WebPデコード実装 - dwebp.cの実装に基づく
NextImageStatus nextimage_webp_decode_alloc(
    const uint8_t* webp_data,
    size_t webp_size,
    const NextImageWebPDecodeOptions* options,
    NextImageDecodeBuffer* output
) {
    if (!webp_data || webp_size == 0 || !output) {
        nextimage_set_error("Invalid parameters: NULL input or output");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    memset(output, 0, sizeof(NextImageDecodeBuffer));

    // Initialize decoder config (same as dwebp.c)
    WebPDecoderConfig config;
    if (!WebPInitDecoderConfig(&config)) {
        nextimage_set_error("WebP library version mismatch");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Get bitstream features (same as dwebp.c)
    VP8StatusCode status = WebPGetFeatures(webp_data, webp_size, &config.input);
    if (status != VP8_STATUS_OK) {
        nextimage_set_error("Failed to get WebP features: %d", status);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Determine output format
    NextImagePixelFormat format = options ? options->format : NEXTIMAGE_FORMAT_RGBA;

    // Set colorspace based on format and alpha presence (same logic as dwebp.c:330)
    if (format == NEXTIMAGE_FORMAT_RGBA) {
        // For PNG output, dwebp uses MODE_RGBA if has_alpha, MODE_RGB otherwise
        // But we always output RGBA if requested
        config.output.colorspace = MODE_RGBA;
    } else if (format == NEXTIMAGE_FORMAT_RGB) {
        config.output.colorspace = MODE_RGB;
    } else if (format == NEXTIMAGE_FORMAT_BGRA) {
        config.output.colorspace = MODE_BGRA;
    } else {
        nextimage_set_error("Unsupported output format: %d", format);
        return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    // Apply decode options (same as dwebp.c)
    if (options) {
        config.options.bypass_filtering = options->bypass_filtering;
        config.options.no_fancy_upsampling = options->no_fancy_upsampling;
        config.options.use_threads = options->use_threads;
    }

    // Decode (same as dwebp.c:382 - uses WebPDecode with config)
    status = WebPDecode(webp_data, webp_size, &config);
    if (status != VP8_STATUS_OK) {
        WebPFreeDecBuffer(&config.output);
        nextimage_set_error("WebP decoding failed: %d", status);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Extract decoded data from config.output
    WebPDecBuffer* dec_buffer = &config.output;
    output->width = dec_buffer->width;
    output->height = dec_buffer->height;
    output->bit_depth = 8;
    output->format = format;

    // Calculate size
    int bytes_per_pixel = (format == NEXTIMAGE_FORMAT_RGB) ? 3 : 4;
    output->stride = output->width * bytes_per_pixel;
    size_t buffer_size = output->stride * output->height;

    // Copy data from WebP decoder's internal buffer
    output->data = (uint8_t*)nextimage_malloc(buffer_size);
    if (!output->data) {
        WebPFreeDecBuffer(dec_buffer);
        nextimage_set_error("Failed to allocate output buffer");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // Copy from WebP's buffer to our buffer
    const uint8_t* src = dec_buffer->u.RGBA.rgba;
    uint8_t* dst = output->data;
    int src_stride = dec_buffer->u.RGBA.stride;
    for (int y = 0; y < output->height; y++) {
        memcpy(dst, src, output->stride);
        src += src_stride;
        dst += output->stride;
    }

    output->data_size = buffer_size;
    output->data_capacity = buffer_size;
    output->owns_data = 1;

    // Planar formats not supported for WebP
    output->u_plane = NULL;
    output->v_plane = NULL;
    output->u_stride = 0;
    output->v_stride = 0;

    // Free WebP's internal buffer
    WebPFreeDecBuffer(dec_buffer);

    return NEXTIMAGE_OK;
}

// WebPデコード（ユーザー提供バッファ）- dwebp.cの実装に基づく
NextImageStatus nextimage_webp_decode_into(
    const uint8_t* webp_data,
    size_t webp_size,
    const NextImageWebPDecodeOptions* options,
    NextImageDecodeBuffer* buffer
) {
    if (!webp_data || webp_size == 0 || !buffer || !buffer->data) {
        nextimage_set_error("Invalid parameters");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Initialize decoder config
    WebPDecoderConfig config;
    if (!WebPInitDecoderConfig(&config)) {
        nextimage_set_error("WebP library version mismatch");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Get bitstream features
    VP8StatusCode status = WebPGetFeatures(webp_data, webp_size, &config.input);
    if (status != VP8_STATUS_OK) {
        nextimage_set_error("Failed to get WebP features: %d", status);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    NextImagePixelFormat format = options ? options->format : NEXTIMAGE_FORMAT_RGBA;

    // Set colorspace (same logic as dwebp.c:330)
    if (format == NEXTIMAGE_FORMAT_RGBA) {
        config.output.colorspace = MODE_RGBA;
    } else if (format == NEXTIMAGE_FORMAT_RGB) {
        config.output.colorspace = MODE_RGB;
    } else if (format == NEXTIMAGE_FORMAT_BGRA) {
        config.output.colorspace = MODE_BGRA;
    } else {
        nextimage_set_error("Unsupported output format: %d", format);
        return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    // Apply decode options
    if (options) {
        config.options.bypass_filtering = options->bypass_filtering;
        config.options.no_fancy_upsampling = options->no_fancy_upsampling;
        config.options.use_threads = options->use_threads;
    }

    // Decode
    status = WebPDecode(webp_data, webp_size, &config);
    if (status != VP8_STATUS_OK) {
        WebPFreeDecBuffer(&config.output);
        nextimage_set_error("WebP decoding failed: %d", status);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Extract decoded data
    WebPDecBuffer* dec_buffer = &config.output;
    int bytes_per_pixel = (format == NEXTIMAGE_FORMAT_RGB) ? 3 : 4;
    int dst_stride = config.input.width * bytes_per_pixel;
    size_t required_size = dst_stride * config.input.height;

    if (buffer->data_capacity < required_size) {
        WebPFreeDecBuffer(dec_buffer);
        nextimage_set_error("Buffer too small: need %zu, have %zu", required_size, buffer->data_capacity);
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Copy from WebP's buffer to user buffer
    const uint8_t* src = dec_buffer->u.RGBA.rgba;
    uint8_t* dst = buffer->data;
    int src_stride = dec_buffer->u.RGBA.stride;
    for (int y = 0; y < config.input.height; y++) {
        memcpy(dst, src, dst_stride);
        src += src_stride;
        dst += dst_stride;
    }

    buffer->width = config.input.width;
    buffer->height = config.input.height;
    buffer->stride = dst_stride;
    buffer->bit_depth = 8;
    buffer->format = format;

    // Free WebP's internal buffer
    WebPFreeDecBuffer(dec_buffer);

    return NEXTIMAGE_OK;
}

// デコードサイズ取得
NextImageStatus nextimage_webp_decode_size(
    const uint8_t* webp_data,
    size_t webp_size,
    int* width,
    int* height,
    size_t* required_size
) {
    if (!webp_data || webp_size == 0 || !width || !height || !required_size) {
        nextimage_set_error("Invalid parameters");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    WebPBitstreamFeatures features;
    VP8StatusCode status = WebPGetFeatures(webp_data, webp_size, &features);
    if (status != VP8_STATUS_OK) {
        nextimage_set_error("Failed to get WebP features: %d", status);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    *width = features.width;
    *height = features.height;
    *required_size = features.width * features.height * 4; // RGBA

    return NEXTIMAGE_OK;
}

// GIF to WebP conversion
// ========================================
// WebP to GIF conversion helpers
// ========================================

// Memory writer for GIF output
typedef struct {
    uint8_t* data;
    size_t size;
    size_t capacity;
} GIFMemoryWriter;

static int gif_write_func(GifFileType* gif, const GifByteType* buf, int len) {
    GIFMemoryWriter* writer = (GIFMemoryWriter*)gif->UserData;

    // Expand buffer if needed
    size_t new_size = writer->size + len;
    if (new_size > writer->capacity) {
        size_t new_capacity = writer->capacity == 0 ? 4096 : writer->capacity * 2;
        while (new_capacity < new_size) {
            new_capacity *= 2;
        }

        uint8_t* new_data = (uint8_t*)nextimage_realloc(writer->data, new_capacity);
        if (!new_data) {
            return 0;
        }
        writer->data = new_data;
        writer->capacity = new_capacity;
    }

    memcpy(writer->data + writer->size, buf, len);
    writer->size += len;
    return len;
}

// Simple color quantization to 256 colors using 6x6x6 RGB cube
static void quantize_to_palette(const uint8_t* rgba_data, int width, int height,
                                ColorMapObject** out_colormap, uint8_t** out_indices,
                                int* out_transparent_index) {
    const int colors = 256;
    ColorMapObject* colormap = GifMakeMapObject(colors, NULL);
    if (!colormap) {
        *out_colormap = NULL;
        *out_indices = NULL;
        return;
    }

    // Build 6x6x6 RGB cube (216 colors)
    int idx = 0;
    for (int r = 0; r < 6; r++) {
        for (int g = 0; g < 6; g++) {
            for (int b = 0; b < 6; b++) {
                colormap->Colors[idx].Red = r * 51;
                colormap->Colors[idx].Green = g * 51;
                colormap->Colors[idx].Blue = b * 51;
                idx++;
            }
        }
    }

    // Add 40 grayscale levels
    for (int i = 0; i < 40; i++) {
        int gray = 6 + i * 6;
        colormap->Colors[idx].Red = gray;
        colormap->Colors[idx].Green = gray;
        colormap->Colors[idx].Blue = gray;
        idx++;
    }

    // Use last color as transparent
    *out_transparent_index = 255;
    colormap->Colors[255].Red = 0;
    colormap->Colors[255].Green = 0;
    colormap->Colors[255].Blue = 0;

    // Allocate index buffer
    size_t pixel_count = width * height;
    uint8_t* indices = (uint8_t*)nextimage_malloc(pixel_count);
    if (!indices) {
        GifFreeMapObject(colormap);
        *out_colormap = NULL;
        *out_indices = NULL;
        return;
    }

    // Map each pixel to nearest palette color
    for (size_t i = 0; i < pixel_count; i++) {
        uint8_t r = rgba_data[i * 4 + 0];
        uint8_t g = rgba_data[i * 4 + 1];
        uint8_t b = rgba_data[i * 4 + 2];
        uint8_t a = rgba_data[i * 4 + 3];

        // If transparent, use transparent index
        if (a < 128) {
            indices[i] = *out_transparent_index;
            continue;
        }

        // Find nearest color in 6x6x6 cube
        int ri = (r + 25) / 51;
        int gi = (g + 25) / 51;
        int bi = (b + 25) / 51;
        if (ri > 5) ri = 5;
        if (gi > 5) gi = 5;
        if (bi > 5) bi = 5;

        indices[i] = ri * 36 + gi * 6 + bi;
    }

    *out_colormap = colormap;
    *out_indices = indices;
}

// Memory-based GIF reading helper
typedef struct {
    const uint8_t* data;
    size_t size;
    size_t position;
} GIFMemoryReader;

static int gif_read_func(GifFileType* gif, GifByteType* buf, int size) {
    GIFMemoryReader* reader = (GIFMemoryReader*)gif->UserData;
    if (reader->position + size > reader->size) {
        size = (int)(reader->size - reader->position);
    }
    if (size > 0) {
        memcpy(buf, reader->data + reader->position, size);
        reader->position += size;
    }
    return size;
}

NextImageStatus nextimage_gif2webp_alloc(
    const uint8_t* gif_data,
    size_t gif_size,
    const NextImageWebPEncodeOptions* options,
    NextImageEncodeBuffer* output
) {
    if (!gif_data || gif_size == 0 || !output) {
        nextimage_set_error("Invalid parameters for GIF to WebP conversion");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    memset(output, 0, sizeof(NextImageEncodeBuffer));

    // Use default options if not provided
    NextImageWebPEncodeOptions default_opts;
    if (!options) {
        nextimage_webp_default_encode_options(&default_opts);
        options = &default_opts;
    }

    // Set up memory reader
    GIFMemoryReader reader = {gif_data, gif_size, 0};
    int error_code;
    GifFileType* gif = DGifOpen(&reader, gif_read_func, &error_code);
    if (!gif) {
        nextimage_set_error("Failed to open GIF from memory: error %d", error_code);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    NextImageStatus status = NEXTIMAGE_OK;
    WebPConfig config;
    WebPAnimEncoderOptions anim_options;
    WebPAnimEncoder* enc = NULL;
    WebPPicture frame, curr_canvas, prev_canvas;
    WebPData webp_data = {0};
    int frame_number = 0;
    int frame_timestamp = 0;
    int frame_duration = 0;
    int transparent_index = GIF_INDEX_INVALID;
    GIFDisposeMethod orig_dispose = GIF_DISPOSE_NONE;
    int loop_count = 0;
    int stored_loop_count = 0;
    int done = 0;

    memset(&frame, 0, sizeof(frame));
    memset(&curr_canvas, 0, sizeof(curr_canvas));
    memset(&prev_canvas, 0, sizeof(prev_canvas));

    // Initialize WebP config from options
    if (!setup_webp_config(&config, options)) {
        status = NEXTIMAGE_ERROR_ENCODE_FAILED;
        nextimage_set_error("Failed to setup WebP config");
        goto End;
    }

    // Initialize animation encoder options
    if (!WebPAnimEncoderOptionsInit(&anim_options)) {
        status = NEXTIMAGE_ERROR_ENCODE_FAILED;
        nextimage_set_error("Failed to initialize WebP animation encoder options");
        goto End;
    }

    // Apply animation options
    if (options->allow_mixed) anim_options.allow_mixed = 1;
    if (options->minimize_size) anim_options.minimize_size = 1;
    if (options->kmin >= 0) anim_options.kmin = options->kmin;
    if (options->kmax >= 0) anim_options.kmax = options->kmax;

    // Set default kmin/kmax if not specified
    if (anim_options.kmin < 0) {
        anim_options.kmin = config.lossless ? 9 : 3;
    }
    if (anim_options.kmax < 0) {
        anim_options.kmax = config.lossless ? 17 : 5;
    }

    // Loop over GIF images
    do {
        GifRecordType type;
        if (DGifGetRecordType(gif, &type) == GIF_ERROR) {
            status = NEXTIMAGE_ERROR_DECODE_FAILED;
            nextimage_set_error("Failed to get GIF record type");
            goto End;
        }

        switch (type) {
            case IMAGE_DESC_RECORD_TYPE: {
                GIFFrameRect gif_rect;
                GifImageDesc* image_desc = &gif->Image;

                if (!DGifGetImageDesc(gif)) {
                    status = NEXTIMAGE_ERROR_DECODE_FAILED;
                    nextimage_set_error("Failed to get GIF image descriptor");
                    goto End;
                }

                if (frame_number == 0) {
                    // Fix broken GIF global headers with 0x0 dimension
                    if (gif->SWidth == 0 || gif->SHeight == 0) {
                        image_desc->Left = 0;
                        image_desc->Top = 0;
                        gif->SWidth = image_desc->Width;
                        gif->SHeight = image_desc->Height;
                        if (gif->SWidth <= 0 || gif->SHeight <= 0) {
                            status = NEXTIMAGE_ERROR_DECODE_FAILED;
                            nextimage_set_error("Invalid GIF dimensions");
                            goto End;
                        }
                    }

                    // Allocate canvases
                    frame.width = gif->SWidth;
                    frame.height = gif->SHeight;
                    frame.use_argb = 1;
                    if (!WebPPictureAlloc(&frame)) {
                        status = NEXTIMAGE_ERROR_OUT_OF_MEMORY;
                        nextimage_set_error("Failed to allocate WebP frame");
                        goto End;
                    }
                    GIFClearPic(&frame, NULL);
                    if (!(WebPPictureCopy(&frame, &curr_canvas) &&
                          WebPPictureCopy(&frame, &prev_canvas))) {
                        status = NEXTIMAGE_ERROR_OUT_OF_MEMORY;
                        nextimage_set_error("Failed to allocate canvas");
                        goto End;
                    }

                    // Get background color
                    GIFGetBackgroundColor(gif->SColorMap, gif->SBackGroundColor,
                                        transparent_index, &anim_options.anim_params.bgcolor);

                    // Initialize encoder
                    enc = WebPAnimEncoderNew(curr_canvas.width, curr_canvas.height, &anim_options);
                    if (!enc) {
                        status = NEXTIMAGE_ERROR_ENCODE_FAILED;
                        nextimage_set_error("Failed to create WebP animation encoder");
                        goto End;
                    }
                }

                // Fix broken GIF sub-rect with zero width/height
                if (image_desc->Width == 0 || image_desc->Height == 0) {
                    image_desc->Width = gif->SWidth;
                    image_desc->Height = gif->SHeight;
                }

                if (!GIFReadFrame(gif, transparent_index, &gif_rect, &frame)) {
                    status = NEXTIMAGE_ERROR_DECODE_FAILED;
                    nextimage_set_error("Failed to read GIF frame");
                    goto End;
                }

                // Blend frame with canvas
                GIFBlendFrames(&frame, &gif_rect, &curr_canvas);

                if (!WebPAnimEncoderAdd(enc, &curr_canvas, frame_timestamp, &config)) {
                    status = NEXTIMAGE_ERROR_ENCODE_FAILED;
                    nextimage_set_error("Failed to add frame: %s", WebPAnimEncoderGetError(enc));
                    goto End;
                }
                ++frame_number;

                // Update canvases
                GIFDisposeFrame(orig_dispose, &gif_rect, &prev_canvas, &curr_canvas);
                GIFCopyPixels(&curr_canvas, &prev_canvas);

                // Force small durations to 100ms
                if (frame_duration <= 10) {
                    frame_duration = 100;
                }

                // Update timestamp for next frame
                frame_timestamp += frame_duration;

                // Reset frame properties for next frame
                orig_dispose = GIF_DISPOSE_NONE;
                frame_duration = 0;
                transparent_index = GIF_INDEX_INVALID;
                break;
            }
            case EXTENSION_RECORD_TYPE: {
                int extension;
                GifByteType* data = NULL;
                if (DGifGetExtension(gif, &extension, &data) == GIF_ERROR) {
                    status = NEXTIMAGE_ERROR_DECODE_FAILED;
                    nextimage_set_error("Failed to read GIF extension");
                    goto End;
                }
                if (data == NULL) continue;

                switch (extension) {
                    case GRAPHICS_EXT_FUNC_CODE: {
                        if (!GIFReadGraphicsExtension(data, &frame_duration, &orig_dispose,
                                                    &transparent_index)) {
                            status = NEXTIMAGE_ERROR_DECODE_FAILED;
                            nextimage_set_error("Failed to read graphics extension");
                            goto End;
                        }
                        break;
                    }
                    case APPLICATION_EXT_FUNC_CODE: {
                        if (data[0] == 11 && (!memcmp(data + 1, "NETSCAPE2.0", 11) ||
                                             !memcmp(data + 1, "ANIMEXTS1.0", 11))) {
                            if (!GIFReadLoopCount(gif, &data, &loop_count)) {
                                status = NEXTIMAGE_ERROR_DECODE_FAILED;
                                nextimage_set_error("Failed to read loop count");
                                goto End;
                            }
                            stored_loop_count = options->loop_compatibility ? (loop_count != 0) : 1;
                        }
                        break;
                    }
                    default:
                        break;
                }
                while (data != NULL) {
                    if (DGifGetExtensionNext(gif, &data) == GIF_ERROR) {
                        status = NEXTIMAGE_ERROR_DECODE_FAILED;
                        nextimage_set_error("Failed to read extension next");
                        goto End;
                    }
                }
                break;
            }
            case TERMINATE_RECORD_TYPE: {
                done = 1;
                break;
            }
            default:
                break;
        }
    } while (!done);

    // Add final NULL frame
    if (!WebPAnimEncoderAdd(enc, NULL, frame_timestamp, NULL)) {
        status = NEXTIMAGE_ERROR_ENCODE_FAILED;
        nextimage_set_error("Failed to flush WebP muxer: %s", WebPAnimEncoderGetError(enc));
        goto End;
    }

    // Assemble the animation
    if (!WebPAnimEncoderAssemble(enc, &webp_data)) {
        status = NEXTIMAGE_ERROR_ENCODE_FAILED;
        nextimage_set_error("Failed to assemble WebP animation: %s", WebPAnimEncoderGetError(enc));
        goto End;
    }

    // Handle loop count
    if (frame_number == 1) {
        loop_count = 0;
    } else if (!options->loop_compatibility) {
        if (!stored_loop_count && frame_number > 1) {
            stored_loop_count = 1;
            loop_count = 1;
        } else if (loop_count > 0 && loop_count < 65535) {
            loop_count += 1;
        }
    }

    // Apply user-specified loop count if set
    if (options->anim_loop_count >= 0) {
        loop_count = options->anim_loop_count;
    }

    // Copy output data
    output->data = nextimage_malloc(webp_data.size);
    if (!output->data) {
        status = NEXTIMAGE_ERROR_OUT_OF_MEMORY;
        nextimage_set_error("Failed to allocate output buffer");
        goto End;
    }
    memcpy(output->data, webp_data.bytes, webp_data.size);
    output->size = webp_data.size;

End:
    WebPDataClear(&webp_data);
    if (enc) WebPAnimEncoderDelete(enc);
    WebPPictureFree(&frame);
    WebPPictureFree(&curr_canvas);
    WebPPictureFree(&prev_canvas);
    if (gif) DGifCloseFile(gif, &error_code);

    return status;
}

// WebP to GIF conversion using giflib
NextImageStatus nextimage_webp2gif_alloc(
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageEncodeBuffer* output
) {
    if (!webp_data || webp_size == 0 || !output) {
        nextimage_set_error("Invalid parameters for WebP to GIF conversion");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    memset(output, 0, sizeof(NextImageEncodeBuffer));

    // Decode WebP to RGBA
    int width, height;
    uint8_t* rgba_data = WebPDecodeRGBA(webp_data, webp_size, &width, &height);
    if (!rgba_data) {
        nextimage_set_error("Failed to decode WebP data");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Quantize to 256 colors
    ColorMapObject* colormap = NULL;
    uint8_t* indices = NULL;
    int transparent_index = -1;
    quantize_to_palette(rgba_data, width, height, &colormap, &indices, &transparent_index);

    WebPFree(rgba_data);

    if (!colormap || !indices) {
        nextimage_set_error("Failed to quantize image to 256 colors");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // Create GIF in memory
    GIFMemoryWriter writer = {0};
    int error_code;

    GifFileType* gif = EGifOpen(&writer, gif_write_func, &error_code);
    if (!gif) {
        GifFreeMapObject(colormap);
        nextimage_free(indices);
        nextimage_set_error("Failed to create GIF: %d", error_code);
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Set GIF dimensions and color map
    if (EGifPutScreenDesc(gif, width, height, 8, 0, colormap) == GIF_ERROR) {
        EGifCloseFile(gif, &error_code);
        GifFreeMapObject(colormap);
        nextimage_free(indices);
        if (writer.data) nextimage_free(writer.data);
        nextimage_set_error("Failed to write GIF screen descriptor");
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Write graphics control extension for transparency
    if (transparent_index >= 0) {
        uint8_t ext_data[4] = {1, 0, 0, (uint8_t)transparent_index}; // transparent flag + index
        if (EGifPutExtension(gif, GRAPHICS_EXT_FUNC_CODE, 4, ext_data) == GIF_ERROR) {
            EGifCloseFile(gif, &error_code);
            GifFreeMapObject(colormap);
            nextimage_free(indices);
            if (writer.data) nextimage_free(writer.data);
            nextimage_set_error("Failed to write GIF graphics extension");
            return NEXTIMAGE_ERROR_ENCODE_FAILED;
        }
    }

    // Write image data
    if (EGifPutImageDesc(gif, 0, 0, width, height, 0, NULL) == GIF_ERROR) {
        EGifCloseFile(gif, &error_code);
        GifFreeMapObject(colormap);
        nextimage_free(indices);
        if (writer.data) nextimage_free(writer.data);
        nextimage_set_error("Failed to write GIF image descriptor");
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Write scanlines
    for (int y = 0; y < height; y++) {
        if (EGifPutLine(gif, indices + y * width, width) == GIF_ERROR) {
            EGifCloseFile(gif, &error_code);
            GifFreeMapObject(colormap);
            nextimage_free(indices);
            if (writer.data) nextimage_free(writer.data);
            nextimage_set_error("Failed to write GIF scanline");
            return NEXTIMAGE_ERROR_ENCODE_FAILED;
        }
    }

    // Close GIF file
    if (EGifCloseFile(gif, &error_code) == GIF_ERROR) {
        GifFreeMapObject(colormap);
        nextimage_free(indices);
        if (writer.data) nextimage_free(writer.data);
        nextimage_set_error("Failed to close GIF: %d", error_code);
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Cleanup
    GifFreeMapObject(colormap);
    nextimage_free(indices);

    // Set output
    output->data = writer.data;
    output->size = writer.size;

    return NEXTIMAGE_OK;
}

// ========================================
// インスタンスベースのエンコーダー/デコーダー
// ========================================

// エンコーダー構造体
struct NextImageWebPEncoder {
    WebPConfig config;
    NextImageWebPEncodeOptions options;
};

// デコーダー構造体
struct NextImageWebPDecoder {
    NextImageWebPDecodeOptions options;
};

// エンコーダーの作成
NextImageWebPEncoder* nextimage_webp_encoder_create(
    const NextImageWebPEncodeOptions* options
) {
    NextImageWebPEncoder* encoder = (NextImageWebPEncoder*)nextimage_malloc(sizeof(NextImageWebPEncoder));
    if (!encoder) {
        nextimage_set_error("Failed to allocate encoder");
        return NULL;
    }

    // オプションをコピー
    if (options) {
        encoder->options = *options;
    } else {
        nextimage_webp_default_encode_options(&encoder->options);
    }

    // WebPConfigを事前設定
    if (!setup_webp_config(&encoder->config, &encoder->options)) {
        nextimage_free(encoder);
        return NULL;
    }

    return encoder;
}

// エンコーダーでエンコード
NextImageStatus nextimage_webp_encoder_encode(
    NextImageWebPEncoder* encoder,
    const uint8_t* input_data,
    size_t input_size,
    NextImageEncodeBuffer* output
) {
    if (!encoder) {
        nextimage_set_error("Invalid encoder instance");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // エンコーダーのconfigを使って通常のエンコード処理
    // （configは既に設定済み）
    return nextimage_webp_encode_alloc(input_data, input_size, &encoder->options, output);
}

// エンコーダーの破棄
void nextimage_webp_encoder_destroy(NextImageWebPEncoder* encoder) {
    if (encoder) {
        nextimage_free(encoder);
    }
}

// デコーダーの作成
NextImageWebPDecoder* nextimage_webp_decoder_create(
    const NextImageWebPDecodeOptions* options
) {
    NextImageWebPDecoder* decoder = (NextImageWebPDecoder*)nextimage_malloc(sizeof(NextImageWebPDecoder));
    if (!decoder) {
        nextimage_set_error("Failed to allocate decoder");
        return NULL;
    }

    if (options) {
        decoder->options = *options;
    } else {
        nextimage_webp_default_decode_options(&decoder->options);
    }

    return decoder;
}

// デコーダーでデコード
NextImageStatus nextimage_webp_decoder_decode(
    NextImageWebPDecoder* decoder,
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageDecodeBuffer* output
) {
    if (!decoder) {
        nextimage_set_error("Invalid decoder instance");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    return nextimage_webp_decode_alloc(webp_data, webp_size, &decoder->options, output);
}

// デコーダーの破棄
void nextimage_webp_decoder_destroy(NextImageWebPDecoder* decoder) {
    if (decoder) {
        nextimage_free(decoder);
    }
}
