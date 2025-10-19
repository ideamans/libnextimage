// SPEC.md準拠のコマンドベースインターフェースのテスト
#include "nextimage.h"
#include "nextimage/cwebp.h"
#include "nextimage/gif2webp.h"
#include "nextimage/webp2gif.h"
#include "nextimage/avifenc.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>

// ファイル読み込みヘルパー
static uint8_t* read_file(const char* path, size_t* size) {
    FILE* f = fopen(path, "rb");
    if (!f) {
        return NULL;
    }

    fseek(f, 0, SEEK_END);
    *size = ftell(f);
    rewind(f);

    uint8_t* data = malloc(*size);
    if (data) {
        fread(data, 1, *size, f);
    }
    fclose(f);
    return data;
}

void test_cwebp_command(void) {
    printf("\nTesting CWebP command interface...\n");

    // デフォルトオプションを作成
    CWebPOptions* opts = cwebp_create_default_options();
    assert(opts != NULL);
    printf("  ✓ cwebp_create_default_options() succeeded\n");

    // オプションをカスタマイズ
    opts->quality = 80;
    opts->method = 4;

    // コマンドを作成
    CWebPCommand* cmd = cwebp_new_command(opts);
    assert(cmd != NULL);
    printf("  ✓ cwebp_new_command() succeeded\n");

    // 画像を読み込んでエンコード
    size_t jpeg_size;
    uint8_t* jpeg_data = read_file("../../testdata/jpeg/gradient.jpg", &jpeg_size);
    assert(jpeg_data != NULL);
    printf("  ✓ Read JPEG file: %zu bytes\n", jpeg_size);

    NextImageBuffer webp_output = {0};
    NextImageStatus status = cwebp_run_command(cmd, jpeg_data, jpeg_size, &webp_output);
    assert(status == NEXTIMAGE_OK);
    assert(webp_output.data != NULL);
    assert(webp_output.size > 0);
    printf("  ✓ cwebp_run_command() succeeded: %zu bytes\n", webp_output.size);

    // 2枚目の画像も同じコマンドで処理
    free(jpeg_data);
    jpeg_data = read_file("../../testdata/png/red.png", &jpeg_size);
    assert(jpeg_data != NULL);

    nextimage_free_buffer(&webp_output);
    memset(&webp_output, 0, sizeof(webp_output));

    status = cwebp_run_command(cmd, jpeg_data, jpeg_size, &webp_output);
    assert(status == NEXTIMAGE_OK);
    assert(webp_output.data != NULL);
    printf("  ✓ Second cwebp_run_command() succeeded: %zu bytes\n", webp_output.size);

    // クリーンアップ
    nextimage_free_buffer(&webp_output);
    free(jpeg_data);
    cwebp_free_command(cmd);
    cwebp_free_options(opts);
    printf("  ✓ CWebP command interface test passed\n");
}

void test_gif2webp_command(void) {
    printf("\nTesting Gif2WebP command interface...\n");

    // デフォルトオプションを作成
    Gif2WebPOptions* opts = gif2webp_create_default_options();
    assert(opts != NULL);
    printf("  ✓ gif2webp_create_default_options() succeeded\n");

    // コマンドを作成
    Gif2WebPCommand* cmd = gif2webp_new_command(opts);
    assert(cmd != NULL);
    printf("  ✓ gif2webp_new_command() succeeded\n");

    // GIFを読み込んで変換
    size_t gif_size;
    uint8_t* gif_data = read_file("../../testdata/gif/static-64x64.gif", &gif_size);
    assert(gif_data != NULL);
    printf("  ✓ Read GIF file: %zu bytes\n", gif_size);

    NextImageBuffer webp_output = {0};
    NextImageStatus status = gif2webp_run_command(cmd, gif_data, gif_size, &webp_output);
    assert(status == NEXTIMAGE_OK);
    assert(webp_output.data != NULL);
    printf("  ✓ gif2webp_run_command() succeeded: %zu bytes\n", webp_output.size);

    // クリーンアップ
    nextimage_free_buffer(&webp_output);
    free(gif_data);
    gif2webp_free_command(cmd);
    gif2webp_free_options(opts);
    printf("  ✓ Gif2WebP command interface test passed\n");
}

void test_webp2gif_command(void) {
    printf("\nTesting WebP2Gif command interface...\n");

    // デフォルトオプションを作成
    WebP2GifOptions* opts = webp2gif_create_default_options();
    assert(opts != NULL);
    printf("  ✓ webp2gif_create_default_options() succeeded\n");

    // コマンドを作成
    WebP2GifCommand* cmd = webp2gif_new_command(opts);
    assert(cmd != NULL);
    printf("  ✓ webp2gif_new_command() succeeded\n");

    // まずPNGをWebPに変換
    size_t png_size;
    uint8_t* png_data = read_file("../../testdata/png/red.png", &png_size);
    assert(png_data != NULL);

    NextImageBuffer webp_temp = {0};
    CWebPCommand* cwebp_cmd = cwebp_new_command(NULL);
    NextImageStatus status = cwebp_run_command(cwebp_cmd, png_data, png_size, &webp_temp);
    assert(status == NEXTIMAGE_OK);
    free(png_data);
    cwebp_free_command(cwebp_cmd);
    printf("  ✓ Converted PNG to WebP: %zu bytes\n", webp_temp.size);

    // WebPをGIFに変換
    NextImageBuffer gif_output = {0};
    status = webp2gif_run_command(cmd, webp_temp.data, webp_temp.size, &gif_output);
    assert(status == NEXTIMAGE_OK);
    assert(gif_output.data != NULL);
    printf("  ✓ webp2gif_run_command() succeeded: %zu bytes\n", gif_output.size);

    // クリーンアップ
    nextimage_free_buffer(&webp_temp);
    nextimage_free_buffer(&gif_output);
    webp2gif_free_command(cmd);
    webp2gif_free_options(opts);
    printf("  ✓ WebP2Gif command interface test passed\n");
}

void test_avifenc_command(void) {
    printf("\nTesting AVIFEnc command interface...\n");

    // デフォルトオプションを作成
    AVIFEncOptions* opts = avifenc_create_default_options();
    assert(opts != NULL);
    printf("  ✓ avifenc_create_default_options() succeeded\n");

    // オプションをカスタマイズ
    opts->quality = 75;
    opts->speed = 6;

    // コマンドを作成
    AVIFEncCommand* cmd = avifenc_new_command(opts);
    assert(cmd != NULL);
    printf("  ✓ avifenc_new_command() succeeded\n");

    // PNGを読み込んでエンコード
    size_t png_size;
    uint8_t* png_data = read_file("../../testdata/png/red.png", &png_size);
    assert(png_data != NULL);
    printf("  ✓ Read PNG file: %zu bytes\n", png_size);

    NextImageBuffer avif_output = {0};
    NextImageStatus status = avifenc_run_command(cmd, png_data, png_size, &avif_output);
    assert(status == NEXTIMAGE_OK);
    assert(avif_output.data != NULL);
    printf("  ✓ avifenc_run_command() succeeded: %zu bytes\n", avif_output.size);

    // クリーンアップ
    nextimage_free_buffer(&avif_output);
    free(png_data);
    avifenc_free_command(cmd);
    avifenc_free_options(opts);
    printf("  ✓ AVIFEnc command interface test passed\n");
}

int main(void) {
    printf("=== SPEC.md Command Interface Test ===\n");
    printf("Version: %s\n", nextimage_version());

    test_cwebp_command();
    test_gif2webp_command();
    test_webp2gif_command();
    test_avifenc_command();

    printf("\n=== All command interface tests passed! ===\n");
    return 0;
}
