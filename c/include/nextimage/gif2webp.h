#ifndef NEXTIMAGE_GIF2WEBP_H
#define NEXTIMAGE_GIF2WEBP_H

#include "../nextimage.h"
#include "cwebp.h"

#ifdef __cplusplus
extern "C" {
#endif

// gif2webp は cwebp のオプションをそのまま使用
typedef CWebPOptions Gif2WebPOptions;

// デフォルトオプションの作成
Gif2WebPOptions* gif2webp_create_default_options(void);
void gif2webp_free_options(Gif2WebPOptions* options);

// ========================================
// コマンドインターフェース
// ========================================

// 不透明なコマンド構造体
typedef struct Gif2WebPCommand Gif2WebPCommand;

// コマンドの作成
Gif2WebPCommand* gif2webp_new_command(const Gif2WebPOptions* options);

// バイト列の変換
NextImageStatus gif2webp_run_command(
    Gif2WebPCommand* cmd,
    const uint8_t* gif_data,
    size_t gif_size,
    NextImageBuffer* output
);

// コマンドの解放
void gif2webp_free_command(Gif2WebPCommand* cmd);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_GIF2WEBP_H
