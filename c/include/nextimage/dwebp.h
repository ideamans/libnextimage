#ifndef NEXTIMAGE_DWEBP_H
#define NEXTIMAGE_DWEBP_H

#include "../nextimage.h"

#ifdef __cplusplus
extern "C" {
#endif

// dwebp デコードオプション (dwebpの全オプションに対応)
typedef struct {
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
} DWebPOptions;

// デフォルトオプションの作成
DWebPOptions* dwebp_create_default_options(void);
void dwebp_free_options(DWebPOptions* options);

// ========================================
// コマンドインターフェース
// ========================================

// 不透明なコマンド構造体
typedef struct DWebPCommand DWebPCommand;

// コマンドの作成
DWebPCommand* dwebp_new_command(const DWebPOptions* options);

// バイト列の変換
NextImageStatus dwebp_run_command(
    DWebPCommand* cmd,
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageBuffer* output  // PNG/JPEGなどのフォーマットで出力
);

// コマンドの解放
void dwebp_free_command(DWebPCommand* cmd);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_DWEBP_H
