#ifndef NEXTIMAGE_WEBP2GIF_H
#define NEXTIMAGE_WEBP2GIF_H

#include "../nextimage.h"

#ifdef __cplusplus
extern "C" {
#endif

// webp2gif オプション（現在は特にオプションなし）
typedef struct {
    // 将来の拡張用に予約
    int reserved;
} WebP2GifOptions;

// デフォルトオプションの作成
WebP2GifOptions* webp2gif_create_default_options(void);
void webp2gif_free_options(WebP2GifOptions* options);

// ========================================
// コマンドインターフェース
// ========================================

// 不透明なコマンド構造体
typedef struct WebP2GifCommand WebP2GifCommand;

// コマンドの作成
WebP2GifCommand* webp2gif_new_command(const WebP2GifOptions* options);

// バイト列の変換
NextImageStatus webp2gif_run_command(
    WebP2GifCommand* cmd,
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageBuffer* output
);

// コマンドの解放
void webp2gif_free_command(WebP2GifCommand* cmd);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_WEBP2GIF_H
