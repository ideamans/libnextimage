// ヘッダーファイルのコンパイルテスト
// このファイルは新しいインターフェースがコンパイルできることを確認するためのもの

#include "nextimage.h"
#include "nextimage/cwebp.h"
#include "nextimage/dwebp.h"
#include "nextimage/gif2webp.h"
#include "nextimage/webp2gif.h"
#include "nextimage/avifenc.h"
#include "nextimage/avifdec.h"

#include <stdio.h>

int main(void) {
    printf("=== Header Compilation Test ===\n");

    // 型が定義されていることを確認
    printf("Testing type definitions...\n");

    // CWebP
    CWebPOptions cwebp_opts;
    CWebPCommand* cwebp_cmd = NULL;
    printf("  ✓ CWebPOptions and CWebPCommand defined\n");

    // DWebP
    DWebPOptions dwebp_opts;
    DWebPCommand* dwebp_cmd = NULL;
    printf("  ✓ DWebPOptions and DWebPCommand defined\n");

    // Gif2WebP
    Gif2WebPOptions gif2webp_opts;
    Gif2WebPCommand* gif2webp_cmd = NULL;
    printf("  ✓ Gif2WebPOptions and Gif2WebPCommand defined\n");

    // WebP2Gif
    WebP2GifOptions webp2gif_opts;
    WebP2GifCommand* webp2gif_cmd = NULL;
    printf("  ✓ WebP2GifOptions and WebP2GifCommand defined\n");

    // AVIFEnc
    AVIFEncOptions avifenc_opts;
    AVIFEncCommand* avifenc_cmd = NULL;
    printf("  ✓ AVIFEncOptions and AVIFEncCommand defined\n");

    // AVIFDec
    AVIFDecOptions avifdec_opts;
    AVIFDecCommand* avifdec_cmd = NULL;
    printf("  ✓ AVIFDecOptions and AVIFDecCommand defined\n");

    // NextImageBuffer
    NextImageBuffer buffer;
    printf("  ✓ NextImageBuffer defined\n");

    printf("\n=== All header tests passed! ===\n");
    printf("Note: This test only checks compilation, not runtime behavior.\n");

    return 0;
}
