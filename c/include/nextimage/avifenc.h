#ifndef NEXTIMAGE_AVIFENC_H
#define NEXTIMAGE_AVIFENC_H

#include "../nextimage.h"

#ifdef __cplusplus
extern "C" {
#endif

// avifenc エンコードオプション
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
} AVIFEncOptions;

// デフォルトオプションの作成
AVIFEncOptions* avifenc_create_default_options(void);
void avifenc_free_options(AVIFEncOptions* options);

// ========================================
// コマンドインターフェース
// ========================================

// 不透明なコマンド構造体
typedef struct AVIFEncCommand AVIFEncCommand;

// コマンドの作成
AVIFEncCommand* avifenc_new_command(const AVIFEncOptions* options);

// バイト列の変換
NextImageStatus avifenc_run_command(
    AVIFEncCommand* cmd,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output
);

// コマンドの解放
void avifenc_free_command(AVIFEncCommand* cmd);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_AVIFENC_H
