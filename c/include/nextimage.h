#ifndef NEXTIMAGE_H
#define NEXTIMAGE_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// バージョン情報
#define NEXTIMAGE_VERSION_MAJOR 1
#define NEXTIMAGE_VERSION_MINOR 0
#define NEXTIMAGE_VERSION_PATCH 0

// ステータスコード
typedef enum {
    NEXTIMAGE_OK = 0,
    NEXTIMAGE_ERROR_INVALID_PARAM = -1,
    NEXTIMAGE_ERROR_ENCODE_FAILED = -2,
    NEXTIMAGE_ERROR_DECODE_FAILED = -3,
    NEXTIMAGE_ERROR_OUT_OF_MEMORY = -4,
    NEXTIMAGE_ERROR_UNSUPPORTED = -5,
    NEXTIMAGE_ERROR_BUFFER_TOO_SMALL = -6,
} NextImageStatus;

// ピクセルフォーマット定義
typedef enum {
    NEXTIMAGE_FORMAT_RGBA = 0,      // RGBA 8bit/channel
    NEXTIMAGE_FORMAT_RGB = 1,       // RGB 8bit/channel
    NEXTIMAGE_FORMAT_BGRA = 2,      // BGRA 8bit/channel
    NEXTIMAGE_FORMAT_YUV420 = 3,    // YUV 4:2:0 planar
    NEXTIMAGE_FORMAT_YUV422 = 4,    // YUV 4:2:2 planar
    NEXTIMAGE_FORMAT_YUV444 = 5,    // YUV 4:4:4 planar
} NextImagePixelFormat;

// 出力バッファ（画像ファイル形式のバイト列）
// エンコード結果など、常にライブラリが割り当てる
typedef struct {
    uint8_t* data;
    size_t size;
} NextImageBuffer;

// デコード用バッファ情報（プレーン別の詳細情報を含む）
typedef struct {
    // プライマリプレーン（インターリーブ形式の場合は全データ、planarの場合はYプレーン）
    uint8_t* data;
    size_t data_capacity;       // dataバッファの容量（*_into関数用、バイト単位）
    size_t data_size;           // 実際のデータサイズ（バイト単位）
    size_t stride;              // Y/プライマリプレーンの行ごとのバイト数

    // Uプレーン（YUV planarの場合のみ使用）
    uint8_t* u_plane;
    size_t u_capacity;          // Uプレーンバッファの容量（*_into関数用）
    size_t u_size;              // Uプレーンの実際のサイズ
    size_t u_stride;            // Uプレーンの行ごとのバイト数

    // Vプレーン（YUV planarの場合のみ使用）
    uint8_t* v_plane;
    size_t v_capacity;          // Vプレーンバッファの容量（*_into関数用）
    size_t v_size;              // Vプレーンの実際のサイズ
    size_t v_stride;            // Vプレーンの行ごとのバイト数

    // メタデータ
    int width;                  // 画像幅（ピクセル単位）
    int height;                 // 画像高さ（ピクセル単位）
    int bit_depth;              // ビット深度（8, 10, 12）
    NextImagePixelFormat format; // ピクセルフォーマット
    int owns_data;              // 1ならライブラリがメモリを所有
} NextImageDecodeBuffer;

// バッファのメモリ解放
void nextimage_free_buffer(NextImageBuffer* buffer);
void nextimage_free_decode_buffer(NextImageDecodeBuffer* buffer);

// エラーメッセージ取得
// - スレッドローカルストレージに保存された最後のエラーメッセージを返す
// - 返される文字列は次のFFI呼び出しまで有効（コピー不要だがスレッドローカル）
// - 成功した呼び出しでは自動的にクリアされない（明示的なクリアが必要）
// - NULLが返された場合はエラーメッセージが設定されていない
const char* nextimage_last_error_message(void);

// エラーメッセージのクリア
// - 次のエラーまでnextimage_last_error_message()がNULLを返すようにする
void nextimage_clear_error(void);

// デバッグビルド専用: メモリリークカウンター
// - 現在の割り当てカウント - 解放カウントを返す
// - リリースビルドでは常に0を返す
#ifdef NEXTIMAGE_DEBUG
int64_t nextimage_allocation_counter(void);
#endif

// バージョン取得
const char* nextimage_version(void);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_H
