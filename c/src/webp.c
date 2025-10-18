#include "webp.h"
#include "internal.h"
#include <string.h>

// libwebp headers
#include "webp/encode.h"
#include "webp/decode.h"
#include "webp/demux.h"
#include "webp/mux.h"

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

// エンコード実装
NextImageStatus nextimage_webp_encode_alloc(
    const uint8_t* input_data,
    size_t input_size,
    int width,
    int height,
    NextImagePixelFormat input_format,
    const NextImageWebPEncodeOptions* options,
    NextImageEncodeBuffer* output
) {
    // input_size is for future validation
    (void)input_size;

    if (!input_data || !output) {
        nextimage_set_error("Invalid parameters: NULL input or output");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    if (width <= 0 || height <= 0) {
        nextimage_set_error("Invalid image dimensions: %dx%d", width, height);
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Clear output
    memset(output, 0, sizeof(NextImageEncodeBuffer));

    // Setup WebP config
    WebPConfig config;
    if (!setup_webp_config(&config, options)) {
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Setup picture
    WebPPicture picture;
    if (!WebPPictureInit(&picture)) {
        nextimage_set_error("Failed to initialize WebP picture");
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    picture.width = width;
    picture.height = height;
    picture.use_argb = 1; // Use ARGB format internally

    // Import pixels based on format
    int import_result = 0;
    int stride = 0;

    switch (input_format) {
        case NEXTIMAGE_FORMAT_RGBA:
            stride = width * 4;
            import_result = WebPPictureImportRGBA(&picture, input_data, stride);
            break;
        case NEXTIMAGE_FORMAT_RGB:
            stride = width * 3;
            import_result = WebPPictureImportRGB(&picture, input_data, stride);
            break;
        case NEXTIMAGE_FORMAT_BGRA:
            stride = width * 4;
            import_result = WebPPictureImportBGRA(&picture, input_data, stride);
            break;
        default:
            WebPPictureFree(&picture);
            nextimage_set_error("Unsupported input format: %d", input_format);
            return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    if (!import_result) {
        WebPPictureFree(&picture);
        nextimage_set_error("Failed to import pixels into WebP picture");
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Setup memory writer
    WebPMemoryWriter writer;
    WebPMemoryWriterInit(&writer);
    picture.writer = WebPMemoryWrite;
    picture.custom_ptr = &writer;

    // Encode
    if (!WebPEncode(&config, &picture)) {
        WebPPictureFree(&picture);
        WebPMemoryWriterClear(&writer);
        nextimage_set_error("WebP encoding failed: %d", picture.error_code);
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Copy output data using our tracked allocation
    output->data = nextimage_malloc(writer.size);
    if (!output->data) {
        WebPPictureFree(&picture);
        WebPMemoryWriterClear(&writer);
        nextimage_set_error("Failed to allocate output buffer");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    memcpy(output->data, writer.mem, writer.size);
    output->size = writer.size;

    // Cleanup
    WebPPictureFree(&picture);
    WebPMemoryWriterClear(&writer);

    return NEXTIMAGE_OK;
}

// デコード実装（alloc版）
NextImageStatus nextimage_webp_decode_alloc(
    const uint8_t* webp_data,
    size_t webp_size,
    const NextImageWebPDecodeOptions* options,
    NextImageDecodeBuffer* output
) {
    if (!webp_data || !output) {
        nextimage_set_error("Invalid parameters: NULL input or output");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Clear output
    memset(output, 0, sizeof(NextImageDecodeBuffer));

    // Get image dimensions
    int width, height;
    if (!WebPGetInfo(webp_data, webp_size, &width, &height)) {
        nextimage_set_error("Failed to get WebP image info");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Decode based on requested format
    NextImagePixelFormat format = options ? options->format : NEXTIMAGE_FORMAT_RGBA;
    uint8_t* decoded_data = NULL;
    int bytes_per_pixel = 0;

    switch (format) {
        case NEXTIMAGE_FORMAT_RGBA:
            bytes_per_pixel = 4;
            decoded_data = WebPDecodeRGBA(webp_data, webp_size, &width, &height);
            break;
        case NEXTIMAGE_FORMAT_RGB:
            bytes_per_pixel = 3;
            decoded_data = WebPDecodeRGB(webp_data, webp_size, &width, &height);
            break;
        case NEXTIMAGE_FORMAT_BGRA:
            bytes_per_pixel = 4;
            decoded_data = WebPDecodeBGRA(webp_data, webp_size, &width, &height);
            break;
        default:
            nextimage_set_error("Unsupported output format: %d", format);
            return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    if (!decoded_data) {
        nextimage_set_error("WebP decoding failed");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Calculate size and copy to tracked allocation
    size_t data_size = (size_t)width * height * bytes_per_pixel;
    output->data = nextimage_malloc(data_size);
    if (!output->data) {
        WebPFree(decoded_data);
        nextimage_set_error("Failed to allocate output buffer");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    memcpy(output->data, decoded_data, data_size);
    WebPFree(decoded_data);

    // Set output metadata
    output->data_size = data_size;
    output->data_capacity = data_size;
    output->stride = width * bytes_per_pixel;
    output->width = width;
    output->height = height;
    output->bit_depth = 8;
    output->format = format;
    output->owns_data = 1;

    // Planar data not used for interleaved formats
    output->u_plane = NULL;
    output->v_plane = NULL;

    return NEXTIMAGE_OK;
}

// デコードサイズ計算
NextImageStatus nextimage_webp_decode_size(
    const uint8_t* webp_data,
    size_t webp_size,
    int* width,
    int* height,
    size_t* required_size
) {
    if (!webp_data || !width || !height || !required_size) {
        nextimage_set_error("Invalid parameters: NULL pointer");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    if (!WebPGetInfo(webp_data, webp_size, width, height)) {
        nextimage_set_error("Failed to get WebP image info");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Assume RGBA format (4 bytes per pixel)
    *required_size = (size_t)(*width) * (*height) * 4;

    return NEXTIMAGE_OK;
}

// デコード（into版）- 簡易実装（allocしてコピー）
NextImageStatus nextimage_webp_decode_into(
    const uint8_t* webp_data,
    size_t webp_size,
    const NextImageWebPDecodeOptions* options,
    NextImageDecodeBuffer* buffer
) {
    if (!buffer || !buffer->data || buffer->data_capacity == 0) {
        nextimage_set_error("Invalid buffer: data or capacity not set");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Decode to temporary buffer first
    NextImageDecodeBuffer temp;
    NextImageStatus status = nextimage_webp_decode_alloc(webp_data, webp_size, options, &temp);
    if (status != NEXTIMAGE_OK) {
        return status;
    }

    // Check buffer size
    if (buffer->data_capacity < temp.data_size) {
        nextimage_free_decode_buffer(&temp);
        nextimage_set_error("Buffer too small: need %zu bytes, have %zu bytes",
                          temp.data_size, buffer->data_capacity);
        return NEXTIMAGE_ERROR_BUFFER_TOO_SMALL;
    }

    // Copy to user buffer
    memcpy(buffer->data, temp.data, temp.data_size);
    buffer->data_size = temp.data_size;
    buffer->stride = temp.stride;
    buffer->width = temp.width;
    buffer->height = temp.height;
    buffer->bit_depth = temp.bit_depth;
    buffer->format = temp.format;
    // owns_data remains as set by caller

    nextimage_free_decode_buffer(&temp);

    return NEXTIMAGE_OK;
}

// GIF to WebP と WebP to GIF は後で実装（Phase 4）
NextImageStatus nextimage_gif2webp_alloc(
    const uint8_t* gif_data,
    size_t gif_size,
    const NextImageWebPEncodeOptions* options,
    NextImageEncodeBuffer* output
) {
    (void)gif_data;
    (void)gif_size;
    (void)options;
    (void)output;
    nextimage_set_error("GIF to WebP conversion not yet implemented");
    return NEXTIMAGE_ERROR_UNSUPPORTED;
}

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
