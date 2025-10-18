#include "webp.h"
#include "internal.h"
#include <string.h>

// libwebp headers
#include "webp/encode.h"
#include "webp/decode.h"
#include "webp/demux.h"
#include "webp/mux.h"

// imageio headers for reading JPEG/PNG/etc
#include "image_dec.h"

// デフォルトエンコードオプション
void nextimage_webp_default_encode_options(NextImageWebPEncodeOptions* options) {
    if (!options) return;

    memset(options, 0, sizeof(NextImageWebPEncodeOptions));
    options->quality = 75.0f;
    options->lossless = 0;
    options->method = 4;
    options->target_size = 0;
    options->target_psnr = 0.0f;
    options->exact = 0;
    options->alpha_compression = 1;
    options->alpha_quality = 100;
    options->pass = 1;
    options->preprocessing = 0;
    options->partitions = 0;
    options->partition_limit = 0;
}

// デフォルトデコードオプション
void nextimage_webp_default_decode_options(NextImageWebPDecodeOptions* options) {
    if (!options) return;

    memset(options, 0, sizeof(NextImageWebPDecodeOptions));
    options->use_threads = 0;
    options->bypass_filtering = 0;
    options->no_fancy_upsampling = 0;
    options->format = NEXTIMAGE_FORMAT_RGBA;
}

// WebP config を NextImageWebPEncodeOptions から設定
static int setup_webp_config(WebPConfig* config, const NextImageWebPEncodeOptions* options) {
    if (!WebPConfigInit(config)) {
        nextimage_set_error("Failed to initialize WebP config");
        return 0;
    }

    if (options) {
        config->quality = options->quality;
        config->lossless = options->lossless;
        config->method = options->method;
        config->target_size = options->target_size;
        config->target_PSNR = options->target_psnr;
        config->exact = options->exact;
        config->alpha_compression = options->alpha_compression;
        config->alpha_quality = options->alpha_quality;
        config->pass = options->pass;
        config->preprocessing = options->preprocessing;
        config->partitions = options->partitions;
        config->partition_limit = options->partition_limit;
    }

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

    // 画像を読み込む（keep_alpha=1, metadata=NULL）
    if (!reader(input_data, input_size, &picture, 1, NULL)) {
        WebPPictureFree(&picture);
        nextimage_set_error("Failed to read input image");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // WebP設定
    WebPConfig config;
    if (!setup_webp_config(&config, options)) {
        WebPPictureFree(&picture);
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
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

// WebPデコード実装
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

    // Get image info
    WebPBitstreamFeatures features;
    VP8StatusCode status = WebPGetFeatures(webp_data, webp_size, &features);
    if (status != VP8_STATUS_OK) {
        nextimage_set_error("Failed to get WebP features: %d", status);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    output->width = features.width;
    output->height = features.height;
    output->bit_depth = 8;
    output->format = options ? options->format : NEXTIMAGE_FORMAT_RGBA;

    // Calculate buffer size
    int bytes_per_pixel = (output->format == NEXTIMAGE_FORMAT_RGB) ? 3 : 4;
    output->stride = output->width * bytes_per_pixel;
    size_t buffer_size = output->stride * output->height;

    // Allocate buffer
    output->data = (uint8_t*)nextimage_malloc(buffer_size);
    if (!output->data) {
        nextimage_set_error("Failed to allocate decode buffer");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }
    output->data_capacity = buffer_size;

    // Decode
    uint8_t* result = NULL;
    if (output->format == NEXTIMAGE_FORMAT_RGBA) {
        result = WebPDecodeRGBAInto(webp_data, webp_size, output->data, buffer_size, output->stride);
    } else if (output->format == NEXTIMAGE_FORMAT_RGB) {
        result = WebPDecodeRGBInto(webp_data, webp_size, output->data, buffer_size, output->stride);
    } else if (output->format == NEXTIMAGE_FORMAT_BGRA) {
        result = WebPDecodeBGRAInto(webp_data, webp_size, output->data, buffer_size, output->stride);
    } else {
        nextimage_free(output->data);
        output->data = NULL;
        nextimage_set_error("Unsupported output format: %d", output->format);
        return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    if (!result) {
        nextimage_free(output->data);
        output->data = NULL;
        nextimage_set_error("WebP decoding failed");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Set data_size
    output->data_size = buffer_size;
    output->owns_data = 1;

    // Planar formats not supported for WebP
    output->u_plane = NULL;
    output->v_plane = NULL;
    output->u_stride = 0;
    output->v_stride = 0;

    return NEXTIMAGE_OK;
}

// WebPデコード（ユーザー提供バッファ）
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

    // Get image info
    WebPBitstreamFeatures features;
    VP8StatusCode status = WebPGetFeatures(webp_data, webp_size, &features);
    if (status != VP8_STATUS_OK) {
        nextimage_set_error("Failed to get WebP features: %d", status);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    NextImagePixelFormat format = options ? options->format : NEXTIMAGE_FORMAT_RGBA;
    int bytes_per_pixel = (format == NEXTIMAGE_FORMAT_RGB) ? 3 : 4;
    int stride = features.width * bytes_per_pixel;
    size_t required_size = stride * features.height;

    if (buffer->data_capacity < required_size) {
        nextimage_set_error("Buffer too small: need %zu, have %zu", required_size, buffer->data_capacity);
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Decode into user buffer
    uint8_t* result = NULL;
    if (format == NEXTIMAGE_FORMAT_RGBA) {
        result = WebPDecodeRGBAInto(webp_data, webp_size, buffer->data, buffer->data_capacity, stride);
    } else if (format == NEXTIMAGE_FORMAT_RGB) {
        result = WebPDecodeRGBInto(webp_data, webp_size, buffer->data, buffer->data_capacity, stride);
    } else if (format == NEXTIMAGE_FORMAT_BGRA) {
        result = WebPDecodeBGRAInto(webp_data, webp_size, buffer->data, buffer->data_capacity, stride);
    } else {
        nextimage_set_error("Unsupported output format: %d", format);
        return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    if (!result) {
        nextimage_set_error("WebP decoding failed");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    buffer->width = features.width;
    buffer->height = features.height;
    buffer->stride = stride;
    buffer->bit_depth = 8;
    buffer->format = format;

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
NextImageStatus nextimage_gif2webp_alloc(
    const uint8_t* gif_data,
    size_t gif_size,
    const NextImageWebPEncodeOptions* options,
    NextImageEncodeBuffer* output
) {
    // GIF is also handled by image_dec.h
    // Just use the same encode_alloc function
    return nextimage_webp_encode_alloc(gif_data, gif_size, options, output);
}

// WebP to GIF conversion (placeholder - requires GIF encoder)
NextImageStatus nextimage_webp2gif_alloc(
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageEncodeBuffer* output
) {
    (void)webp_data;
    (void)webp_size;
    (void)output;

    nextimage_set_error("WebP to GIF conversion not yet implemented");
    return NEXTIMAGE_ERROR_UNSUPPORTED;
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
