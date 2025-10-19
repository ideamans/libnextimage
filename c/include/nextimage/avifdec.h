#ifndef NEXTIMAGE_AVIFDEC_H
#define NEXTIMAGE_AVIFDEC_H

#include "../nextimage.h"

#ifdef __cplusplus
extern "C" {
#endif

// Output format enum
typedef enum {
    AVIFDEC_OUTPUT_PNG = 0,   // PNG output (default)
    AVIFDEC_OUTPUT_JPEG = 1   // JPEG output
} AVIFDecOutputFormat;

// avifdec デコードオプション
typedef struct {
    // 出力設定
    AVIFDecOutputFormat output_format; // PNG or JPEG output (default: PNG)
    int jpeg_quality;                  // JPEG quality 0-100 (default: 90, only for JPEG output)

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
} AVIFDecOptions;

// デフォルトオプションの作成
AVIFDecOptions* avifdec_create_default_options(void);
void avifdec_free_options(AVIFDecOptions* options);

// ========================================
// コマンドインターフェース
// ========================================

// 不透明なコマンド構造体
typedef struct AVIFDecCommand AVIFDecCommand;

// コマンドの作成
AVIFDecCommand* avifdec_new_command(const AVIFDecOptions* options);

// バイト列の変換
NextImageStatus avifdec_run_command(
    AVIFDecCommand* cmd,
    const uint8_t* avif_data,
    size_t avif_size,
    NextImageBuffer* output
);

// コマンドの解放
void avifdec_free_command(AVIFDecCommand* cmd);

#ifdef __cplusplus
}
#endif

#endif // NEXTIMAGE_AVIFDEC_H
