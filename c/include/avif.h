#ifndef NEXTIMAGE_AVIF_H
#define NEXTIMAGE_AVIF_H

#include "nextimage.h"

#ifdef __cplusplus
extern "C" {
#endif

// AVIF エンコードオプション
typedef struct {
    int quality;            // 0-100, default 50 (AVIF_QUANTIZER_BEST_QUALITY=0, WORST=63 mapped to 0-100)
    int speed;              // 0-10, default 6 (0=slowest/best, 10=fastest/worst)
    int min_quantizer;      // 0-63, default 0 (best quality)
    int max_quantizer;      // 0-63, default determined by quality
    int min_quantizer_alpha;// 0-63, default 0
    int max_quantizer_alpha;// 0-63, default determined by quality
    int enable_alpha;       // 0 or 1, default 1
    int bit_depth;          // 8, 10, or 12 (default: 8)
    int yuv_format;         // 0=444, 1=422, 2=420, 3=400 (default: 420)
    int tile_rows_log2;     // 0-6, default 0
    int tile_cols_log2;     // 0-6, default 0
} NextImageAVIFEncodeOptions;

// AVIF デコードオプション
typedef struct {
    int use_threads;            // 0 or 1, enable multi-threading
    NextImagePixelFormat format; // desired pixel format (default: RGBA)
    int ignore_exif;            // 0 or 1, ignore EXIF metadata
    int ignore_xmp;             // 0 or 1, ignore XMP metadata
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
    NextImageEncodeBuffer* output
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
    NextImageEncodeBuffer* output);

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
