#include "avif.h"
#include "internal.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

// libavif headers
#include "avif/avif.h"

// imageio headers for reading JPEG/PNG/etc (from libwebp)
#include "image_dec.h"

// libwebp Picture for intermediate conversion
#include "webp/encode.h"
#include "webp/decode.h"

// Platform-specific headers for CPU count query
#if defined(_WIN32)
#include <windows.h>
#elif defined(__APPLE__)
#include <sys/sysctl.h>
#else
#include <unistd.h>
#endif

// Query CPU count (based on avifutil.c from libavif apps)
static int queryCPUCount(void) {
#if defined(_WIN32)
    SYSTEM_INFO sysinfo;
    GetSystemInfo(&sysinfo);
    return (int)sysinfo.dwNumberOfProcessors;
#elif defined(__APPLE__)
    int mib[2];
    int numCPU;
    size_t len = sizeof(numCPU);

    mib[0] = CTL_HW;
    mib[1] = HW_AVAILCPU;

    sysctl(mib, 2, &numCPU, &len, NULL, 0);

    if (numCPU < 1) {
        mib[1] = HW_NCPU;
        sysctl(mib, 2, &numCPU, &len, NULL, 0);
        if (numCPU < 1)
            numCPU = 1;
    }
    return numCPU;
#else
    // POSIX
    int numCPU = (int)sysconf(_SC_NPROCESSORS_ONLN);
    return (numCPU > 0) ? numCPU : 1;
#endif
}

// デフォルトエンコードオプション
void nextimage_avif_default_encode_options(NextImageAVIFEncodeOptions* options) {
    if (!options) return;

    memset(options, 0, sizeof(NextImageAVIFEncodeOptions));

    // Quality settings (avifenc defaults)
    options->quality = 60;        // avifenc default quality (line 855 in avifenc.c)
    options->quality_alpha = -1;  // -1 means use quality value
    options->speed = 6;           // avifenc default speed (line 1466) - NOT AVIF_SPEED_DEFAULT!

    // Deprecated quantizer settings (for backward compatibility)
    options->min_quantizer = -1;       // -1 means "use quality setting"
    options->max_quantizer = -1;       // -1 means "use quality setting"
    options->min_quantizer_alpha = -1; // -1 means "use quality_alpha setting"
    options->max_quantizer_alpha = -1; // -1 means "use quality_alpha setting"

    // Format settings
    options->bit_depth = 8;
    options->yuv_format = 0;  // 444 (avifenc default for PNG, line 1467)
    options->yuv_range = 1;   // full range (avifenc default for PNG/JPEG, line 1468)

    // Alpha settings
    options->enable_alpha = 1;
    options->premultiply_alpha = 0;

    // Tiling settings
    options->tile_rows_log2 = 0;
    options->tile_cols_log2 = 0;

    // CICP (nclx) color settings - avifenc defaults (lines 1469-1471)
    options->color_primaries = 1;        // BT709 (avifenc default)
    options->transfer_characteristics = 13; // sRGB (avifenc default)
    options->matrix_coefficients = 6;    // BT601 (avifenc default)

    // Advanced settings
    options->sharp_yuv = 0;
    options->target_size = 0;  // disabled

    // Transformation settings
    options->irot_angle = -1;      // disabled
    options->imir_axis = -1;       // disabled

    // Pixel aspect ratio (pasp)
    options->pasp[0] = -1;  // disabled
    options->pasp[1] = -1;

    // Crop rectangle
    options->crop[0] = -1;  // disabled
    options->crop[1] = -1;
    options->crop[2] = -1;
    options->crop[3] = -1;

    // Clean aperture (clap)
    options->clap[0] = -1;  // disabled (widthN)
    options->clap[1] = 1;   // widthD
    options->clap[2] = -1;  // heightN
    options->clap[3] = 1;   // heightD
    options->clap[4] = 0;   // horizOffN
    options->clap[5] = 1;   // horizOffD
    options->clap[6] = 0;   // vertOffN
    options->clap[7] = 1;   // vertOffD

    // Content light level information (clli)
    options->clli_max_cll = -1;    // disabled
    options->clli_max_pall = -1;   // disabled

    // Animation settings
    options->timescale = 30;
    options->keyframe_interval = 0;  // disabled
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
    // CRITICAL: Zero-initialize to prevent stack memory pollution between calls
    memset(&picture, 0, sizeof(WebPPicture));
    if (!WebPPictureInit(&picture)) {
        nextimage_set_error("Failed to initialize WebPPicture");
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    // CRITICAL: Set use_argb BEFORE calling reader to request ARGB format
    // The reader checks this flag and preserves ARGB if set
    picture.use_argb = 1;

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

    // Set color properties from options (CICP/nclx)
    // Use options values if specified, otherwise use defaults
    if (options->color_primaries >= 0) {
        image->colorPrimaries = (avifColorPrimaries)options->color_primaries;
    } else {
        image->colorPrimaries = AVIF_COLOR_PRIMARIES_BT709;  // default
    }

    if (options->transfer_characteristics >= 0) {
        image->transferCharacteristics = (avifTransferCharacteristics)options->transfer_characteristics;
    } else {
        image->transferCharacteristics = AVIF_TRANSFER_CHARACTERISTICS_SRGB;  // default: sRGB
    }

    if (options->matrix_coefficients >= 0) {
        image->matrixCoefficients = (avifMatrixCoefficients)options->matrix_coefficients;
    } else {
        image->matrixCoefficients = AVIF_MATRIX_COEFFICIENTS_BT601;  // default
    }

    // Set YUV range (limited vs full)
    image->yuvRange = (options->yuv_range == 0) ? AVIF_RANGE_LIMITED : AVIF_RANGE_FULL;

    // Set alpha premultiplication flag
    if (options->premultiply_alpha) {
        image->alphaPremultiplied = AVIF_TRUE;
    }

    // avifRGBImageを設定
    avifRGBImage rgb;
    // CRITICAL: Zero-initialize to match avifenc behavior
    memset(&rgb, 0, sizeof(avifRGBImage));

    avifRGBImageSetDefaults(&rgb, image);
    rgb.format = AVIF_RGB_FORMAT_RGBA;
    rgb.depth = 8;

    // Set chroma downsampling method (SharpYUV if requested)
    if (options->sharp_yuv && options->yuv_format == 2) {  // YUV420 only
        rgb.chromaDownsampling = AVIF_CHROMA_DOWNSAMPLING_SHARP_YUV;
    } else {
        rgb.chromaDownsampling = AVIF_CHROMA_DOWNSAMPLING_AUTOMATIC;
    }

    // RGBAバッファを割り当て
    avifResult allocResult = avifRGBImageAllocatePixels(&rgb);
    if (allocResult != AVIF_RESULT_OK) {
        avifImageDestroy(image);
        WebPPictureFree(&picture);
        nextimage_set_error("Failed to allocate RGB buffer: %s", avifResultToString(allocResult));
        return NEXTIMAGE_ERROR_OUT_OF_MEMORY;
    }

    // WebPPictureのARGBデータをavifRGBImageにコピー
    // picture.use_argb=1を事前設定しているため、readerはARGBフォーマットで読み込むはず
    if (!picture.use_argb || !picture.argb) {
        avifRGBImageFreePixels(&rgb);
        avifImageDestroy(image);
        WebPPictureFree(&picture);
        nextimage_set_error("WebPPicture is not in ARGB format (use_argb=%d, argb=%p)",
                           picture.use_argb, (void*)picture.argb);
        return NEXTIMAGE_ERROR_DECODE_FAILED;
    }

    // ARGB (uint32_t) -> RGBA (uint8_t) conversion
    const uint32_t* src = picture.argb;
    uint8_t* dst = rgb.pixels;
    for (int y = 0; y < picture.height; y++) {
        for (int x = 0; x < picture.width; x++) {
            const uint32_t argb_val = src[y * picture.argb_stride + x];
            const uint8_t a = (argb_val >> 24) & 0xFF;
            const uint8_t r = (argb_val >> 16) & 0xFF;
            const uint8_t g = (argb_val >> 8) & 0xFF;
            const uint8_t b = argb_val & 0xFF;

            dst[y * rgb.rowBytes + x * 4 + 0] = r;
            dst[y * rgb.rowBytes + x * 4 + 1] = g;
            dst[y * rgb.rowBytes + x * 4 + 2] = b;
            dst[y * rgb.rowBytes + x * 4 + 3] = a;
        }
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

    // Set metadata (EXIF, XMP, ICC) if provided
    if (options->exif_data && options->exif_size > 0) {
        result = avifImageSetMetadataExif(image, options->exif_data, options->exif_size);
        if (result != AVIF_RESULT_OK) {
            avifImageDestroy(image);
            nextimage_set_error("Failed to set EXIF metadata: %s", avifResultToString(result));
            return NEXTIMAGE_ERROR_ENCODE_FAILED;
        }
    }

    if (options->xmp_data && options->xmp_size > 0) {
        result = avifImageSetMetadataXMP(image, options->xmp_data, options->xmp_size);
        if (result != AVIF_RESULT_OK) {
            avifImageDestroy(image);
            nextimage_set_error("Failed to set XMP metadata: %s", avifResultToString(result));
            return NEXTIMAGE_ERROR_ENCODE_FAILED;
        }
    }

    if (options->icc_data && options->icc_size > 0) {
        result = avifImageSetProfileICC(image, options->icc_data, options->icc_size);
        if (result != AVIF_RESULT_OK) {
            avifImageDestroy(image);
            nextimage_set_error("Failed to set ICC profile: %s", avifResultToString(result));
            return NEXTIMAGE_ERROR_ENCODE_FAILED;
        }
    }

    // Set transformation properties
    // Note: transformFlags must be set to enable encoding of these properties
    image->transformFlags = 0;

    // Image rotation (irot)
    if (options->irot_angle >= 0 && options->irot_angle <= 3) {
        image->irot.angle = (uint8_t)options->irot_angle;
        image->transformFlags |= AVIF_TRANSFORM_IROT;
    }

    // Image mirror (imir)
    if (options->imir_axis >= 0 && options->imir_axis <= 1) {
        image->imir.axis = (uint8_t)options->imir_axis;
        image->transformFlags |= AVIF_TRANSFORM_IMIR;
    }

    // Pixel aspect ratio (pasp)
    if (options->pasp[0] >= 0 && options->pasp[1] >= 0) {
        image->pasp.hSpacing = (uint32_t)options->pasp[0];
        image->pasp.vSpacing = (uint32_t)options->pasp[1];
        image->transformFlags |= AVIF_TRANSFORM_PASP;
    }

    // Crop rectangle - convert to clap
    if (options->crop[0] >= 0) {
        avifCropRect cropRect;
        cropRect.x = (uint32_t)options->crop[0];
        cropRect.y = (uint32_t)options->crop[1];
        cropRect.width = (uint32_t)options->crop[2];
        cropRect.height = (uint32_t)options->crop[3];

        avifDiagnostics diag;
        avifDiagnosticsClearError(&diag);

        avifBool convertResult = avifCleanApertureBoxFromCropRect(
            &image->clap,
            &cropRect,
            picture.width,
            picture.height,
            &diag
        );

        if (!convertResult) {
            avifRGBImageFreePixels(&rgb);
            avifImageDestroy(image);
            WebPPictureFree(&picture);
            nextimage_set_error("Failed to convert crop rect to clap: %s", diag.error);
            return NEXTIMAGE_ERROR_ENCODE_FAILED;
        }
        image->transformFlags |= AVIF_TRANSFORM_CLAP;
    }
    // Clean aperture (clap) - direct values
    else if (options->clap[0] >= 0) {
        image->clap.widthN = (uint32_t)options->clap[0];
        image->clap.widthD = (uint32_t)options->clap[1];
        image->clap.heightN = (uint32_t)options->clap[2];
        image->clap.heightD = (uint32_t)options->clap[3];
        image->clap.horizOffN = (uint32_t)options->clap[4];
        image->clap.horizOffD = (uint32_t)options->clap[5];
        image->clap.vertOffN = (uint32_t)options->clap[6];
        image->clap.vertOffD = (uint32_t)options->clap[7];
        image->transformFlags |= AVIF_TRANSFORM_CLAP;
    }

    // Content light level information (clli)
    if (options->clli_max_cll >= 0 || options->clli_max_pall >= 0) {
        image->clli.maxCLL = (options->clli_max_cll >= 0) ? (uint16_t)options->clli_max_cll : 0;
        image->clli.maxPALL = (options->clli_max_pall >= 0) ? (uint16_t)options->clli_max_pall : 0;
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

    // Set maxThreads to match avifenc default (CPU count for -j all)
    // avifenc v1.3.0 uses avifQueryCPUCount() when jobs == -1 (default)
    encoder->maxThreads = queryCPUCount();

    // Quality settings (color/YUV)
    if (options->min_quantizer >= 0 && options->max_quantizer >= 0) {
        // Deprecated: use quantizers directly for backward compatibility
        encoder->minQuantizer = options->min_quantizer;
        encoder->maxQuantizer = options->max_quantizer;
    } else {
        // Use quality setting (0-100) - this is the recommended approach
        encoder->quality = options->quality;
    }

    // Quality settings (alpha)
    int quality_alpha = (options->quality_alpha >= 0) ? options->quality_alpha : options->quality;
    if (options->min_quantizer_alpha >= 0 && options->max_quantizer_alpha >= 0) {
        // Deprecated: use quantizers directly for backward compatibility
        encoder->minQuantizerAlpha = options->min_quantizer_alpha;
        encoder->maxQuantizerAlpha = options->max_quantizer_alpha;
    } else {
        // Use qualityAlpha setting (0-100)
        encoder->qualityAlpha = quality_alpha;
    }

    // Tiling settings
    encoder->tileRowsLog2 = options->tile_rows_log2;
    encoder->tileColsLog2 = options->tile_cols_log2;

    // avifenc v1.3.0 starts in automatic tiling mode by default (commit 1894a21d, 2025-04-11)
    // This automatically determines optimal tile configuration based on image dimensions
    // Disable autoTiling if manual tiling is specified
    if (options->tile_rows_log2 > 0 || options->tile_cols_log2 > 0) {
        encoder->autoTiling = AVIF_FALSE;
    } else {
        encoder->autoTiling = AVIF_TRUE;
    }

    // Animation settings (for future use)
    if (options->timescale > 0) {
        encoder->timescale = options->timescale;
    }
    if (options->keyframe_interval > 0) {
        encoder->keyframeInterval = options->keyframe_interval;
    }

    // Encode using avifEncoderAddImage + avifEncoderFinish
    // (matching avifenc.c implementation, lines 1244-1287)
    result = avifEncoderAddImage(encoder, image, 1, AVIF_ADD_IMAGE_FLAG_SINGLE);
    if (result != AVIF_RESULT_OK) {
        avifEncoderDestroy(encoder);
        avifImageDestroy(image);
        nextimage_set_error("AVIF add image failed: %s", avifResultToString(result));
        return NEXTIMAGE_ERROR_ENCODE_FAILED;
    }

    avifRWData raw = AVIF_DATA_EMPTY;
    result = avifEncoderFinish(encoder, &raw);
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

// ========================================
// インスタンスベースのエンコーダー/デコーダー
// ========================================

// エンコーダー構造体
struct NextImageAVIFEncoder {
    NextImageAVIFEncodeOptions options;
};

// デコーダー構造体
struct NextImageAVIFDecoder {
    NextImageAVIFDecodeOptions options;
};

// エンコーダーの作成
NextImageAVIFEncoder* nextimage_avif_encoder_create(
    const NextImageAVIFEncodeOptions* options
) {
    NextImageAVIFEncoder* encoder = (NextImageAVIFEncoder*)nextimage_malloc(sizeof(NextImageAVIFEncoder));
    if (!encoder) {
        nextimage_set_error("Failed to allocate encoder");
        return NULL;
    }

    // オプションをコピー
    if (options) {
        encoder->options = *options;
    } else {
        nextimage_avif_default_encode_options(&encoder->options);
    }

    return encoder;
}

// エンコーダーでエンコード
NextImageStatus nextimage_avif_encoder_encode(
    NextImageAVIFEncoder* encoder,
    const uint8_t* input_data,
    size_t input_size,
    NextImageEncodeBuffer* output
) {
    if (!encoder) {
        nextimage_set_error("Invalid encoder instance");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    return nextimage_avif_encode_alloc(input_data, input_size, &encoder->options, output);
}

// エンコーダーの破棄
void nextimage_avif_encoder_destroy(NextImageAVIFEncoder* encoder) {
    if (encoder) {
        nextimage_free(encoder);
    }
}

// デコーダーの作成
NextImageAVIFDecoder* nextimage_avif_decoder_create(
    const NextImageAVIFDecodeOptions* options
) {
    NextImageAVIFDecoder* decoder = (NextImageAVIFDecoder*)nextimage_malloc(sizeof(NextImageAVIFDecoder));
    if (!decoder) {
        nextimage_set_error("Failed to allocate decoder");
        return NULL;
    }

    if (options) {
        decoder->options = *options;
    } else {
        nextimage_avif_default_decode_options(&decoder->options);
    }

    return decoder;
}

// デコーダーでデコード
NextImageStatus nextimage_avif_decoder_decode(
    NextImageAVIFDecoder* decoder,
    const uint8_t* avif_data,
    size_t avif_size,
    NextImageDecodeBuffer* output
) {
    if (!decoder) {
        nextimage_set_error("Invalid decoder instance");
        return NEXTIMAGE_ERROR_INVALID_PARAM;
    }

    return nextimage_avif_decode_alloc(avif_data, avif_size, &decoder->options, output);
}

// デコーダーの破棄
void nextimage_avif_decoder_destroy(NextImageAVIFDecoder* decoder) {
    if (decoder) {
        nextimage_free(decoder);
    }
}
