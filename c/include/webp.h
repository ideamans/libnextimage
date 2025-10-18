#ifndef NEXTIMAGE_WEBP_H
#define NEXTIMAGE_WEBP_H

#include "nextimage.h"

#ifdef __cplusplus
extern "C" {
#endif

// WebP エンコードオプション
typedef struct {
    float quality;           // 0-100, default 75
    int lossless;           // 0 or 1, default 0
    int method;             // 0-6, default 4 (quality/speed trade-off)
    int target_size;        // target size in bytes (0 = disabled)
    float target_psnr;      // target PSNR (0 = disabled)
    int exact;              // preserve RGB values in transparent area
    int alpha_compression;  // 0=none, 1=compressed (default)
    int alpha_quality;      // 0-100, transparency compression quality
    int pass;               // number of entropy-analysis passes (1-10)
    int preprocessing;      // 0=none, 1=segment-smooth, 2=pseudo-random dithering
    int partitions;         // 0-3, log2(number of token partitions)
    int partition_limit;    // quality degradation allowed (0-100)
} NextImageWebPEncodeOptions;

// WebP デコードオプション
typedef struct {
    int use_threads;            // 0 or 1, enable multi-threading
    int bypass_filtering;       // 0 or 1, disable in-loop filtering
    int no_fancy_upsampling;    // 0 or 1, use faster pointwise upsampler
    NextImagePixelFormat format; // desired pixel format (default: RGBA)
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
    NextImageEncodeBuffer* output
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
    NextImageEncodeBuffer* output
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
    NextImageEncodeBuffer* output
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
    NextImageEncodeBuffer* output);

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
