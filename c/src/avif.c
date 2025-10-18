#include "avif.h"
#include "internal.h"
#include <string.h>

// libavif headers
#include "avif/avif.h"

// imageio headers for reading JPEG/PNG/etc (from libwebp)
#include "image_dec.h"

// libwebp Picture for intermediate conversion
#include "webp/encode.h"

// デフォルトエンコードオプション
void nextimage_avif_default_encode_options(NextImageAVIFEncodeOptions* options) {
    if (!options) return;

    memset(options, 0, sizeof(NextImageAVIFEncodeOptions));
    options->quality = 50;
    options->speed = AVIF_SPEED_DEFAULT;  // 6
    options->min_quantizer = AVIF_QUANTIZER_BEST_QUALITY;  // 0
    options->max_quantizer = AVIF_QUANTIZER_WORST_QUALITY; // 63
    options->min_quantizer_alpha = AVIF_QUANTIZER_BEST_QUALITY;
    options->max_quantizer_alpha = AVIF_QUANTIZER_WORST_QUALITY;
    options->enable_alpha = 1;
    options->bit_depth = 8;
    options->yuv_format = 2; // 420
    options->tile_rows_log2 = 0;
    options->tile_cols_log2 = 0;
}

// デフォルトデコードオプション
void nextimage_avif_default_decode_options(NextImageAVIFDecodeOptions* options) {
    if (!options) return;

    memset(options, 0, sizeof(NextImageAVIFDecodeOptions));
    options->use_threads = 0;
    options->format = NEXTIMAGE_FORMAT_RGBA;
    options->ignore_exif = 0;
    options->ignore_xmp = 0;
}

// Quality (0-100) をAVIF quantizer (0-63)に変換
// quality 100 -> quantizer 0 (best)
// quality 0   -> quantizer 63 (worst)
static int quality_to_quantizer(int quality) {
    if (quality < 0) quality = 0;
    if (quality > 100) quality = 100;
    return (int)(AVIF_QUANTIZER_WORST_QUALITY - (quality * AVIF_QUANTIZER_WORST_QUALITY / 100.0));
}

// YUV format を avifPixelFormat に変換
static avifPixelFormat yuv_format_to_avif(int yuv_format) {
    switch (yuv_format) {
        case 0: return AVIF_PIXEL_FORMAT_YUV444;
        case 1: return AVIF_PIXEL_FORMAT_YUV422;
        case 2: return AVIF_PIXEL_FORMAT_YUV420;
        case 3: return AVIF_PIXEL_FORMAT_YUV400;
        default: return AVIF_PIXEL_FORMAT_YUV420;
    }
}

// NextImagePixelFormat を avifRGBFormat に変換
static avifRGBFormat pixel_format_to_avif_rgb(NextImagePixelFormat format) {
    switch (format) {
        case NEXTIMAGE_FORMAT_RGBA:
            return AVIF_RGB_FORMAT_RGBA;
        case NEXTIMAGE_FORMAT_RGB:
            return AVIF_RGB_FORMAT_RGB;
        case NEXTIMAGE_FORMAT_BGRA:
            return AVIF_RGB_FORMAT_BGRA;
        default:
            return AVIF_RGB_FORMAT_RGBA;
    }
}

// エンコード実装（画像ファイルデータから）
NextImageStatus nextimage_avif_encode_alloc(
    const uint8_t* input_data,
    size_t input_size,
    const NextImageAVIFEncodeOptions* options,
    NextImageEncodeBuffer* output
) {
    if (!input_data || input_size == 0 || !output) {
        nextimage_set_error("Invalid parameters: NULL input or output");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    memset(output, 0, sizeof(NextImageEncodeBuffer));

    // デフォルトオプション
    NextImageAVIFEncodeOptions default_opts;
    if (!options) {
        nextimage_avif_default_encode_options(&default_opts);
        options = &default_opts;
    }

    // 画像フォーマットを推測（libwebpのimageioを使用）
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

    // WebPPictureに一旦読み込む（imageioを使うため）
    WebPPicture picture;
    if (!WebPPictureInit(&picture)) {
        nextimage_set_error("Failed to initialize WebPPicture");
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    if (!reader(input_data, input_size, &picture, 1, NULL)) {
        WebPPictureFree(&picture);
        nextimage_set_error("Failed to read input image");
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // avifImageを作成
    avifImage* image = avifImageCreate(
        picture.width,
        picture.height,
        options->bit_depth,
        yuv_format_to_avif(options->yuv_format)
    );
    if (!image) {
        WebPPictureFree(&picture);
        nextimage_set_error("Failed to create avifImage");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // avifRGBImageを設定
    avifRGBImage rgb;
    avifRGBImageSetDefaults(&rgb, image);
    rgb.format = AVIF_RGB_FORMAT_RGBA;
    rgb.depth = 8;

    // RGBAバッファを割り当て
    avifResult allocResult = avifRGBImageAllocatePixels(&rgb);
    if (allocResult != AVIF_RESULT_OK) {
        avifImageDestroy(image);
        WebPPictureFree(&picture);
        nextimage_set_error("Failed to allocate RGB buffer: %s", avifResultToString(allocResult));
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // WebPPictureのピクセルをRGBバッファにコピー
    if (!WebPPictureImportRGBA(&picture, rgb.pixels, rgb.rowBytes)) {
        avifRGBImageFreePixels(&rgb);
        avifImageDestroy(image);
        WebPPictureFree(&picture);
        nextimage_set_error("Failed to import RGBA data");
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // RGBからYUVに変換
    avifResult result = avifImageRGBToYUV(image, &rgb);
    avifRGBImageFreePixels(&rgb);
    WebPPictureFree(&picture);

    if (result != AVIF_RESULT_OK) {
        avifImageDestroy(image);
        nextimage_set_error("Failed to convert RGB to YUV: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Create encoder
    avifEncoder* encoder = avifEncoderCreate();
    if (!encoder) {
        avifImageDestroy(image);
        nextimage_set_error("Failed to create AVIF encoder");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // Set encoder options
    encoder->speed = options->speed;
    encoder->minQuantizer = options->quality > 0 ? quality_to_quantizer(options->quality) : options->min_quantizer;
    encoder->maxQuantizer = options->quality > 0 ? quality_to_quantizer(options->quality) : options->max_quantizer;
    encoder->minQuantizerAlpha = options->min_quantizer_alpha;
    encoder->maxQuantizerAlpha = options->max_quantizer_alpha;
    encoder->tileRowsLog2 = options->tile_rows_log2;
    encoder->tileColsLog2 = options->tile_cols_log2;

    // Encode
    avifRWData raw = AVIF_DATA_EMPTY;
    result = avifEncoderWrite(encoder, image, &raw);
    if (result != AVIF_RESULT_OK) {
        avifEncoderDestroy(encoder);
        avifImageDestroy(image);
        nextimage_set_error("AVIF encoding failed: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // Copy output data using our tracked allocation
    output->data = nextimage_malloc(raw.size);
    if (!output->data) {
        avifRWDataFree(&raw);
        avifEncoderDestroy(encoder);
        avifImageDestroy(image);
        nextimage_set_error("Failed to allocate output buffer");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    memcpy(output->data, raw.data, raw.size);
    output->size = raw.size;

    // Cleanup
    avifRWDataFree(&raw);
    avifEncoderDestroy(encoder);
    avifImageDestroy(image);

    return NEXTIMAGE_OK;
}

// デコード実装（alloc版）
NextImageStatus nextimage_avif_decode_alloc(
    const uint8_t* avif_data,
    size_t avif_size,
    const NextImageAVIFDecodeOptions* options,
    NextImageDecodeBuffer* output
) {
    if (!avif_data || !output) {
        nextimage_set_error("Invalid parameters: NULL input or output");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Clear output
    memset(output, 0, sizeof(NextImageDecodeBuffer));

    // Get options or use defaults
    NextImageAVIFDecodeOptions default_opts;
    if (!options) {
        nextimage_avif_default_decode_options(&default_opts);
        options = &default_opts;
    }

    // Create decoder
    avifDecoder* decoder = avifDecoderCreate();
    if (!decoder) {
        nextimage_set_error("Failed to create AVIF decoder");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // Set decoder options
    decoder->ignoreExif = options->ignore_exif ? AVIF_TRUE : AVIF_FALSE;
    decoder->ignoreXMP = options->ignore_xmp ? AVIF_TRUE : AVIF_FALSE;

    // Parse input
    avifResult result = avifDecoderSetIOMemory(decoder, avif_data, avif_size);
    if (result != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        nextimage_set_error("Failed to set AVIF decoder input: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Parse image
    result = avifDecoderParse(decoder);
    if (result != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        nextimage_set_error("Failed to parse AVIF: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Get next image (first frame)
    result = avifDecoderNextImage(decoder);
    if (result != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        nextimage_set_error("Failed to decode AVIF image: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    avifImage* image = decoder->image;

    // Setup RGB output
    avifRGBImage rgb;
    avifRGBImageSetDefaults(&rgb, image);
    rgb.format = pixel_format_to_avif_rgb(options->format);
    rgb.depth = 8; // Always output 8-bit for now

    // Allocate RGB buffer
    avifResult allocResult = avifRGBImageAllocatePixels(&rgb);
    if (allocResult != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        nextimage_set_error("Failed to allocate RGB buffer: %s", avifResultToString(allocResult));
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // Convert YUV to RGB
    result = avifImageYUVToRGB(image, &rgb);
    if (result != AVIF_RESULT_OK) {
        avifRGBImageFreePixels(&rgb);
        avifDecoderDestroy(decoder);
        nextimage_set_error("Failed to convert YUV to RGB: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Calculate output size
    int bytes_per_pixel;
    switch (options->format) {
        case NEXTIMAGE_FORMAT_RGBA:
        case NEXTIMAGE_FORMAT_BGRA:
            bytes_per_pixel = 4;
            break;
        case NEXTIMAGE_FORMAT_RGB:
            bytes_per_pixel = 3;
            break;
        default:
            avifRGBImageFreePixels(&rgb);
            avifDecoderDestroy(decoder);
            nextimage_set_error("Unsupported output format: %d", options->format);
            return NEXTIMAGE_ERROR_UNSUPPORTED;
    }

    size_t data_size = (size_t)rgb.width * rgb.height * bytes_per_pixel;

    // Copy to tracked allocation
    output->data = nextimage_malloc(data_size);
    if (!output->data) {
        avifRGBImageFreePixels(&rgb);
        avifDecoderDestroy(decoder);
        nextimage_set_error("Failed to allocate output buffer");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    memcpy(output->data, rgb.pixels, data_size);

    // Set output metadata
    output->data_size = data_size;
    output->data_capacity = data_size;
    output->stride = rgb.width * bytes_per_pixel;
    output->width = rgb.width;
    output->height = rgb.height;
    output->bit_depth = 8; // Always 8-bit output for now
    output->format = options->format;
    output->owns_data = 1;

    // Planar data not used for RGB formats
    output->u_plane = NULL;
    output->v_plane = NULL;

    // Cleanup
    avifRGBImageFreePixels(&rgb);
    avifDecoderDestroy(decoder);

    return NEXTIMAGE_OK;
}

// デコードサイズ計算
NextImageStatus nextimage_avif_decode_size(
    const uint8_t* avif_data,
    size_t avif_size,
    int* width,
    int* height,
    int* bit_depth,
    size_t* required_size
) {
    if (!avif_data || !width || !height || !bit_depth || !required_size) {
        nextimage_set_error("Invalid parameters: NULL pointer");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Create decoder
    avifDecoder* decoder = avifDecoderCreate();
    if (!decoder) {
        nextimage_set_error("Failed to create AVIF decoder");
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // Parse input
    avifResult result = avifDecoderSetIOMemory(decoder, avif_data, avif_size);
    if (result != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        nextimage_set_error("Failed to set AVIF decoder input: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Parse image (just header)
    result = avifDecoderParse(decoder);
    if (result != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        nextimage_set_error("Failed to parse AVIF: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // Get image info
    *width = decoder->image->width;
    *height = decoder->image->height;
    *bit_depth = decoder->image->depth;

    // Assume RGBA format (4 bytes per pixel)
    *required_size = (size_t)(*width) * (*height) * 4;

    avifDecoderDestroy(decoder);

    return NEXTIMAGE_OK;
}

// デコード（into版）
NextImageStatus nextimage_avif_decode_into(
    const uint8_t* avif_data,
    size_t avif_size,
    const NextImageAVIFDecodeOptions* options,
    NextImageDecodeBuffer* buffer
) {
    if (!buffer || !buffer->data || buffer->data_capacity == 0) {
        nextimage_set_error("Invalid buffer: data or capacity not set");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    // Decode to temporary buffer first
    NextImageDecodeBuffer temp;
    NextImageStatus status = nextimage_avif_decode_alloc(avif_data, avif_size, options, &temp);
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
