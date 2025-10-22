#ifndef NEXTIMAGE_WEBP_H
#define NEXTIMAGE_WEBP_H

#include "nextimage.h"

// SPEC.md準拠の新しいコマンドベースインターフェース
#include "nextimage/cwebp.h"
#include "nextimage/dwebp.h"
#include "nextimage/gif2webp.h"
#include "nextimage/webp2gif.h"

#ifdef __cplusplus
extern "C" {
#endif

// WebP preset types (matches WebPPreset)
typedef enum {
    NEXTIMAGE_WEBP_PRESET_DEFAULT = 0,  // default preset
    NEXTIMAGE_WEBP_PRESET_PICTURE = 1,  // digital picture, like portrait
    NEXTIMAGE_WEBP_PRESET_PHOTO = 2,    // outdoor photograph
    NEXTIMAGE_WEBP_PRESET_DRAWING = 3,  // hand or line drawing
    NEXTIMAGE_WEBP_PRESET_ICON = 4,     // small-sized colorful images
    NEXTIMAGE_WEBP_PRESET_TEXT = 5      // text-like
} NextImageWebPPreset;

// WebP image hint types (matches WebPImageHint)
typedef enum {
    NEXTIMAGE_WEBP_HINT_DEFAULT = 0,  // default hint
    NEXTIMAGE_WEBP_HINT_PICTURE = 1,  // digital picture, like portrait
    NEXTIMAGE_WEBP_HINT_PHOTO = 2,    // outdoor photograph
    NEXTIMAGE_WEBP_HINT_GRAPH = 3     // discrete tone image (graph, map-tile)
} NextImageWebPImageHint;

// WebP エンコードオプション (全WebPConfigフィールドに対応)
typedef struct {
    // 基本設定
    float quality;              // 0-100, default 75
    int lossless;              // 0 or 1, default 0
    int method;                // 0-6, default 4 (quality/speed trade-off)

    // プリセット
    NextImageWebPPreset preset;      // -1=none (default), or preset type
    NextImageWebPImageHint image_hint; // image type hint, default 0

    // ロスレスプリセット
    int lossless_preset;       // -1=don't use (default), 0-9=use preset (0=fast, 9=best)

    // ターゲット設定
    int target_size;           // target size in bytes (0 = disabled)
    float target_psnr;         // target PSNR (0 = disabled)

    // セグメント/フィルタ設定
    int segments;              // 1-4, number of segments, default 4
    int sns_strength;          // 0-100, spatial noise shaping, default 50
    int filter_strength;       // 0-100, filter strength, default 60
    int filter_sharpness;      // 0-7, filter sharpness, default 0
    int filter_type;           // 0=simple, 1=strong, default 1
    int autofilter;            // 0 or 1, auto-adjust filter strength, default 0

    // アルファチャンネル設定
    int alpha_compression;     // 0=none, 1=compressed (default)
    int alpha_filtering;       // 0=none, 1=fast, 2=best, default 1
    int alpha_quality;         // 0-100, alpha compression quality, default 100

    // エントロピー設定
    int pass;                  // 1-10, entropy-analysis passes, default 1

    // その他の設定
    int show_compressed;       // 0 or 1, export compressed picture, default 0
    int preprocessing;         // 0=none, 1=segment-smooth, 2=pseudo-random dithering
    int partitions;            // 0-3, log2(number of token partitions), default 0
    int partition_limit;       // 0-100, quality degradation allowed, default 0
    int emulate_jpeg_size;     // 0 or 1, match JPEG size, default 0
    int thread_level;          // 0 or 1, use multi-threading, default 0
    int low_memory;            // 0 or 1, reduce memory usage, default 0
    int near_lossless;         // -1=not set (default), 0-100=use near lossless (auto-enables lossless)
    int exact;                 // 0 or 1, preserve RGB in transparent area, default 0
    int use_delta_palette;     // reserved, default 0
    int use_sharp_yuv;         // 0 or 1, sharp RGB->YUV conversion, default 0
    int qmin;                  // 0-100, minimum permissible quality, default 0
    int qmax;                  // 0-100, maximum permissible quality, default 100

    // メタデータ設定 (cwebp -metadata)
    int keep_metadata;         // -1=default, 0=none, 1=exif, 2=icc, 3=xmp, 4=all

    // 画像変換設定 (cwebp -crop, -resize)
    int crop_x;                // crop rectangle x (-crop x y w h), -1=disabled
    int crop_y;                // crop rectangle y
    int crop_width;            // crop rectangle width
    int crop_height;           // crop rectangle height

    int resize_width;          // resize width (-resize w h), -1=disabled
    int resize_height;         // resize height
    int resize_mode;           // 0=always (default), 1=up_only, 2=down_only

    // アルファチャンネル特殊処理
    uint32_t blend_alpha;      // blend alpha against background color (0xRRGGBB), -1=disabled
    int noalpha;               // 0 or 1, discard alpha channel, default 0

    // アニメーション設定 (gif2webp, WebPAnimEncoder)
    int allow_mixed;           // 0 or 1, allow mixed lossy/lossless (default: 0)
    int minimize_size;         // 0 or 1, minimize output size (slow, default: 0)
    int kmin;                  // min distance between key frames (default: -1=auto)
    int kmax;                  // max distance between key frames (default: -1=auto)
    int anim_loop_count;       // animation loop count, 0=infinite (default: 0)
    int loop_compatibility;    // 0 or 1, Chrome M62 compatibility mode (default: 0)
} NextImageWebPEncodeOptions;

// Output format enum for decoder (must match dwebp.h)
typedef enum {
    NEXTIMAGE_WEBP_OUTPUT_PNG = 0,   // PNG output (default)
    NEXTIMAGE_WEBP_OUTPUT_JPEG = 1   // JPEG output
} NextImageWebPOutputFormat;

// WebP デコードオプション (dwebpの全オプションに対応)
typedef struct {
    // Output format options
    NextImageWebPOutputFormat output_format; // PNG or JPEG output (default: PNG)
    int jpeg_quality;                        // JPEG quality 0-100 (default: 90, only for JPEG output)

    // 基本設定
    int use_threads;            // 0 or 1, enable multi-threading (-mt)
    int bypass_filtering;       // 0 or 1, disable in-loop filtering (-nofilter)
    int no_fancy_upsampling;    // 0 or 1, use faster pointwise upsampler (-nofancy)
    NextImagePixelFormat format; // desired pixel format (default: RGBA)

    // ディザリング設定
    int no_dither;              // 0 or 1, disable dithering (-nodither)
    int dither_strength;        // 0-100, dithering strength (-dither), default 100
    int alpha_dither;           // 0 or 1, use alpha-plane dithering (-alpha_dither)

    // 画像操作
    int crop_x;                 // crop rectangle x (-crop x y w h)
    int crop_y;                 // crop rectangle y
    int crop_width;             // crop rectangle width
    int crop_height;            // crop rectangle height
    int use_crop;               // 0 or 1, enable cropping

    int resize_width;           // resize width (-resize w h)
    int resize_height;          // resize height
    int use_resize;             // 0 or 1, enable resizing

    int flip;                   // 0 or 1, flip vertically (-flip)

    // 特殊モード
    int alpha_only;             // 0 or 1, save only alpha plane (-alpha)
    int incremental;            // 0 or 1, use incremental decoding (-incremental)
} NextImageWebPDecodeOptions;

// デフォルトオプションの取得
void nextimage_webp_default_encode_options(NextImageWebPEncodeOptions* options);
void nextimage_webp_default_decode_options(NextImageWebPDecodeOptions* options);

// ========================================
// WebP エンコード
// ========================================

// エンコード（ライブラリがメモリを割り当て）
// input_data: 画像ファイルデータ（JPEG, PNG, GIF等のバイトデータ）
// input_size: データサイズ（バイト単位）
// options: エンコードオプション（NULLでデフォルト）
// output: 出力バッファ（成功時にdataとsizeが設定される）
// 注: 画像フォーマットは自動判定されます
NextImageStatus nextimage_webp_encode_alloc(
    const uint8_t* input_data,
    size_t input_size,
    const NextImageWebPEncodeOptions* options,
    NextImageBuffer* output
);

// ========================================
// WebP デコード
// ========================================

// デコード（ライブラリがメモリを割り当て）
// webp_data: WebPファイルデータ
// webp_size: データサイズ
// options: デコードオプション（NULLでデフォルト）
// output: 出力バッファ（成功時にピクセルデータとメタデータが設定される）
NextImageStatus nextimage_webp_decode_alloc(
    const uint8_t* webp_data,
    size_t webp_size,
    const NextImageWebPDecodeOptions* options,
    NextImageDecodeBuffer* output
);

// デコード（呼び出し側が用意したバッファを使用）
// buffer->data, buffer->data_capacity を事前に設定すること
// 必要なバッファサイズは nextimage_webp_decode_size() で取得可能
NextImageStatus nextimage_webp_decode_into(
    const uint8_t* webp_data,
    size_t webp_size,
    const NextImageWebPDecodeOptions* options,
    NextImageDecodeBuffer* buffer
);

// デコードに必要なバッファサイズを事前に計算
// width, height: 画像サイズが返される
// required_size: 必要なバッファサイズが返される
NextImageStatus nextimage_webp_decode_size(
    const uint8_t* webp_data,
    size_t webp_size,
    int* width,
    int* height,
    size_t* required_size
);

// ========================================
// GIF to WebP
// ========================================

// GIF to WebP（ライブラリがメモリを割り当て）
// gif_data: GIFファイルデータ
// gif_size: データサイズ
// options: エンコードオプション（NULLでデフォルト）
// output: 出力バッファ（成功時にWebPデータが設定される）
NextImageStatus nextimage_gif2webp_alloc(
    const uint8_t* gif_data,
    size_t gif_size,
    const NextImageWebPEncodeOptions* options,
    NextImageBuffer* output
);

// ========================================
// WebP to GIF (新機能)
// ========================================

// WebP to GIF（ライブラリがメモリを割り当て）
// webp_data: WebPファイルデータ
// webp_size: データサイズ
// output: 出力バッファ（成功時にGIFデータが設定される）
NextImageStatus nextimage_webp2gif_alloc(
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageBuffer* output
);

// ========================================
// インスタンスベースのエンコーダー/デコーダー
// ========================================

// WebPエンコーダーインスタンス（不透明な構造体）
typedef struct NextImageWebPEncoder NextImageWebPEncoder;

// エンコーダーの作成（libwebpの初期化を含む）
// options: エンコードオプション（NULLでデフォルト）
// 戻り値: エンコーダーインスタンス（失敗時はNULL）
NextImageWebPEncoder* nextimage_webp_encoder_create(
    const NextImageWebPEncodeOptions* options);

// エンコーダーでエンコード（繰り返し呼び出し可能）
// encoder: エンコーダーインスタンス
// input_data: 画像ファイルデータ（JPEG, PNG等）
// input_size: データサイズ
// output: 出力バッファ（成功時にWebPデータが設定される）
NextImageStatus nextimage_webp_encoder_encode(
    NextImageWebPEncoder* encoder,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output);

// エンコーダーの破棄（内部メモリの解放）
void nextimage_webp_encoder_destroy(NextImageWebPEncoder* encoder);

// WebPデコーダーインスタンス（不透明な構造体）
typedef struct NextImageWebPDecoder NextImageWebPDecoder;

// デコーダーの作成
// options: デコードオプション（NULLでデフォルト）
// 戻り値: デコーダーインスタンス（失敗時はNULL）
NextImageWebPDecoder* nextimage_webp_decoder_create(
    const NextImageWebPDecodeOptions* options);

// デコーダーでデコード（繰り返し呼び出し可能）
// decoder: デコーダーインスタンス
// webp_data: WebPファイルデータ
// webp_size: データサイズ
// output: 出力バッファ（成功時にピクセルデータとメタデータが設定される）
NextImageStatus nextimage_webp_decoder_decode(
    NextImageWebPDecoder* decoder,
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageDecodeBuffer* output);

// デコーダーの破棄（内部メモリの解放）
void nextimage_webp_decoder_destroy(NextImageWebPDecoder* decoder);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_WEBP_H
