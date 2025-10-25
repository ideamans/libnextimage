#ifndef NEXTIMAGE_CWEBP_H
#define NEXTIMAGE_CWEBP_H

#include "../nextimage.h"

#ifdef __cplusplus
extern "C" {
#endif

// WebP preset types (matches WebPPreset)
typedef enum {
    CWEBP_PRESET_DEFAULT = 0,  // default preset
    CWEBP_PRESET_PICTURE = 1,  // digital picture, like portrait
    CWEBP_PRESET_PHOTO = 2,    // outdoor photograph
    CWEBP_PRESET_DRAWING = 3,  // hand or line drawing
    CWEBP_PRESET_ICON = 4,     // small-sized colorful images
    CWEBP_PRESET_TEXT = 5      // text-like
} CWebPPreset;

// WebP image hint types (matches WebPImageHint)
typedef enum {
    CWEBP_HINT_DEFAULT = 0,  // default hint
    CWEBP_HINT_PICTURE = 1,  // digital picture, like portrait
    CWEBP_HINT_PHOTO = 2,    // outdoor photograph
    CWEBP_HINT_GRAPH = 3     // discrete tone image (graph, map-tile)
} CWebPImageHint;

// Metadata flags (can be combined with bitwise OR)
typedef enum {
    CWEBP_METADATA_NONE = 0,     // No metadata (0)
    CWEBP_METADATA_EXIF = 1,     // Keep EXIF metadata (1 << 0)
    CWEBP_METADATA_ICC  = 2,     // Keep ICC profile (1 << 1)
    CWEBP_METADATA_XMP  = 4,     // Keep XMP metadata (1 << 2)
    CWEBP_METADATA_ALL  = 7      // Keep all metadata (EXIF | ICC | XMP)
} CWebPMetadataFlag;

// cwebp エンコードオプション (全WebPConfigフィールドに対応)
typedef struct {
    // 基本設定
    float quality;              // 0-100, default 75
    int lossless;              // 0 or 1, default 0
    int method;                // 0-6, default 4 (quality/speed trade-off)

    // プリセット
    CWebPPreset preset;             // -1=none (default), or preset type
    CWebPImageHint image_hint;      // image type hint, default 0

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
} CWebPOptions;

// デフォルトオプションの作成
CWebPOptions* cwebp_create_default_options(void);
void cwebp_free_options(CWebPOptions* options);

// ========================================
// コマンドインターフェース
// ========================================

// 不透明なコマンド構造体
typedef struct CWebPCommand CWebPCommand;

// コマンドの作成
CWebPCommand* cwebp_new_command(const CWebPOptions* options);

// バイト列の変換
NextImageStatus cwebp_run_command(
    CWebPCommand* cmd,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output
);

// コマンドの解放
void cwebp_free_command(CWebPCommand* cmd);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_CWEBP_H
