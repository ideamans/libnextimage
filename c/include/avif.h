#ifndef NEXTIMAGE_AVIF_H
#define NEXTIMAGE_AVIF_H

#include "nextimage.h"

// SPEC.md準拠の新しいコマンドベースインターフェース
#include "nextimage/avifenc.h"
#include "nextimage/avifdec.h"

#ifdef __cplusplus
extern "C" {
#endif

// AVIF エンコードオプション
typedef struct {
    // Quality settings
    int quality;            // 0-100, default 60 (for color/YUV)
    int quality_alpha;      // 0-100, default 100 (for alpha channel, -1=use quality)
    int speed;              // 0-10, default 6 (0=slowest/best, 10=fastest/worst)

    // Deprecated quantizer settings (for backward compatibility)
    int min_quantizer;      // 0-63, default -1 (use quality instead)
    int max_quantizer;      // 0-63, default -1 (use quality instead)
    int min_quantizer_alpha;// 0-63, default -1 (use quality_alpha instead)
    int max_quantizer_alpha;// 0-63, default -1 (use quality_alpha instead)

    // Format settings
    int bit_depth;          // 8, 10, or 12 (default: 8)
    int yuv_format;         // 0=444, 1=422, 2=420, 3=400 (default: 444)
    int yuv_range;          // 0=limited, 1=full (default: 1=full for PNG/JPEG)

    // Alpha settings
    int enable_alpha;       // 0 or 1, default 1
    int premultiply_alpha;  // 0 or 1, default 0 (premultiply color by alpha)

    // Tiling settings
    int tile_rows_log2;     // 0-6, default 0
    int tile_cols_log2;     // 0-6, default 0

    // CICP (nclx) color settings
    int color_primaries;    // CICP color primaries, -1=auto (default: 1=BT709)
    int transfer_characteristics; // CICP transfer, -1=auto (default: 13=sRGB)
    int matrix_coefficients;// CICP matrix, -1=auto (default: 6=BT601)

    // Advanced settings
    int sharp_yuv;          // 0 or 1, use sharp RGB->YUV conversion (default: 0)
    int target_size;        // target file size in bytes, 0=disabled (default: 0)

    // Metadata settings
    const uint8_t* exif_data;   // EXIF metadata bytes (NULL=no EXIF)
    size_t exif_size;           // EXIF data size in bytes
    const uint8_t* xmp_data;    // XMP metadata bytes (NULL=no XMP)
    size_t xmp_size;            // XMP data size in bytes
    const uint8_t* icc_data;    // ICC profile bytes (NULL=no ICC)
    size_t icc_size;            // ICC profile size in bytes

    // Transformation settings
    int irot_angle;         // Image rotation: 0-3 (90 * angle degrees anti-clockwise), -1=disabled
    int imir_axis;          // Image mirror: 0=vertical, 1=horizontal, -1=disabled

    // Pixel aspect ratio (pasp) - array[2]: [h_spacing, v_spacing]
    int pasp[2];            // -1=disabled, otherwise [h_spacing, v_spacing]

    // Crop rectangle (simpler interface) - array[4]: [x, y, width, height]
    // This will be converted to clap using avifCleanApertureBoxFromCropRect
    int crop[4];            // -1=disabled, otherwise [x, y, width, height]

    // Clean aperture (clap) - array[8]: [wN,wD, hN,hD, hOffN,hOffD, vOffN,vOffD]
    // Use this for direct clap values, or use crop[] for simpler interface
    int clap[8];            // -1=disabled, otherwise [widthN,widthD, heightN,heightD, horizOffN,horizOffD, vertOffN,vertOffD]

    // Content light level information (clli)
    int clli_max_cll;       // Max content light level (0-65535), -1=disabled
    int clli_max_pall;      // Max picture average light level (0-65535), -1=disabled

    // Animation settings (for future use)
    int timescale;          // timescale/fps for animations (default: 30)
    int keyframe_interval;  // max keyframe interval (default: 0=disabled)
} NextImageAVIFEncodeOptions;

// AVIF デコードオプション
typedef struct {
    int use_threads;            // 0 or 1, enable multi-threading
    NextImagePixelFormat format; // desired pixel format (default: RGBA)
    int ignore_exif;            // 0 or 1, ignore EXIF metadata
    int ignore_xmp;             // 0 or 1, ignore XMP metadata
    int ignore_icc;             // 0 or 1, ignore ICC profile (Note: ICC profile is not returned by decode, so this has no effect)

    // Security limits
    uint32_t image_size_limit;      // Maximum image size in total pixels (default: AVIF_DEFAULT_IMAGE_SIZE_LIMIT = 268435456)
    uint32_t image_dimension_limit; // Maximum image dimension (width or height), 0=ignore (default: AVIF_DEFAULT_IMAGE_DIMENSION_LIMIT = 32768)

    // Validation flags
    int strict_flags;           // Strict validation flags: 0=disabled, 1=enabled (default: 1=AVIF_STRICT_ENABLED)

    // Chroma upsampling (for YUV to RGB conversion)
    int chroma_upsampling;      // 0=automatic (default), 1=fastest, 2=best_quality, 3=nearest, 4=bilinear
} NextImageAVIFDecodeOptions;

// デフォルトオプションの取得
void nextimage_avif_default_encode_options(NextImageAVIFEncodeOptions* options);
void nextimage_avif_default_decode_options(NextImageAVIFDecodeOptions* options);

// ========================================
// AVIF エンコード
// ========================================

// エンコード（ライブラリがメモリを割り当て）
// input_data: 画像ファイルデータ（JPEG, PNG等のバイトデータ）
// input_size: データサイズ（バイト単位）
// options: エンコードオプション（NULLでデフォルト）
// output: 出力バッファ（成功時にdataとsizeが設定される）
// 注: 画像フォーマットは自動判定されます
NextImageStatus nextimage_avif_encode_alloc(
    const uint8_t* input_data,
    size_t input_size,
    const NextImageAVIFEncodeOptions* options,
    NextImageBuffer* output
);

// ========================================
// AVIF デコード
// ========================================

// デコード（ライブラリがメモリを割り当て）
// avif_data: AVIFファイルデータ
// avif_size: データサイズ
// options: デコードオプション（NULLでデフォルト）
// output: 出力バッファ（成功時にピクセルデータとメタデータが設定される）
NextImageStatus nextimage_avif_decode_alloc(
    const uint8_t* avif_data,
    size_t avif_size,
    const NextImageAVIFDecodeOptions* options,
    NextImageDecodeBuffer* output
);

// デコード（呼び出し側が用意したバッファを使用）
// buffer->data, buffer->data_capacity を事前に設定すること
// 必要なバッファサイズは nextimage_avif_decode_size() で取得可能
NextImageStatus nextimage_avif_decode_into(
    const uint8_t* avif_data,
    size_t avif_size,
    const NextImageAVIFDecodeOptions* options,
    NextImageDecodeBuffer* buffer
);

// デコードに必要なバッファサイズを事前に計算
// width, height, bit_depth: 画像情報が返される
// required_size: 必要なバッファサイズが返される
NextImageStatus nextimage_avif_decode_size(
    const uint8_t* avif_data,
    size_t avif_size,
    int* width,
    int* height,
    int* bit_depth,
    size_t* required_size
);

// ========================================
// インスタンスベースのエンコーダー/デコーダー
// ========================================

// AVIFエンコーダーインスタンス（不透明な構造体）
typedef struct NextImageAVIFEncoder NextImageAVIFEncoder;

// エンコーダーの作成（libavifの初期化を含む）
// options: エンコードオプション（NULLでデフォルト）
// 戻り値: エンコーダーインスタンス（失敗時はNULL）
NextImageAVIFEncoder* nextimage_avif_encoder_create(
    const NextImageAVIFEncodeOptions* options);

// エンコーダーでエンコード（繰り返し呼び出し可能）
// encoder: エンコーダーインスタンス
// input_data: 画像ファイルデータ（JPEG, PNG等）
// input_size: データサイズ
// output: 出力バッファ（成功時にAVIFデータが設定される）
NextImageStatus nextimage_avif_encoder_encode(
    NextImageAVIFEncoder* encoder,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output);

// エンコーダーの破棄（内部メモリの解放）
void nextimage_avif_encoder_destroy(NextImageAVIFEncoder* encoder);

// AVIFデコーダーインスタンス（不透明な構造体）
typedef struct NextImageAVIFDecoder NextImageAVIFDecoder;

// デコーダーの作成
// options: デコードオプション（NULLでデフォルト）
// 戻り値: デコーダーインスタンス（失敗時はNULL）
NextImageAVIFDecoder* nextimage_avif_decoder_create(
    const NextImageAVIFDecodeOptions* options);

// デコーダーでデコード（繰り返し呼び出し可能）
// decoder: デコーダーインスタンス
// avif_data: AVIFファイルデータ
// avif_size: データサイズ
// output: 出力バッファ（成功時にピクセルデータとメタデータが設定される）
NextImageStatus nextimage_avif_decoder_decode(
    NextImageAVIFDecoder* decoder,
    const uint8_t* avif_data,
    size_t avif_size,
    NextImageDecodeBuffer* output);

// デコーダーの破棄（内部メモリの解放）
void nextimage_avif_decoder_destroy(NextImageAVIFDecoder* decoder);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_AVIF_H
