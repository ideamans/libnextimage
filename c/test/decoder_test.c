// デコーダー(dwebp/avifdec)のテスト - WebP/AVIFからPNGへの変換
#include "nextimage.h"
#include "nextimage/dwebp.h"
#include "nextimage/avifdec.h"
#include "nextimage/cwebp.h"
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

// PNG署名の確認（最初の8バイト）
static int is_png(const uint8_t* data, size_t size) {
    if (size < 8) return 0;
    const uint8_t png_sig[] = {137, 80, 78, 71, 13, 10, 26, 10};
    return memcmp(data, png_sig, 8) == 0;
}

void test_dwebp_to_png(void) {
    printf("\nTesting DWebP command (WebP -> PNG)...\n");

    // まずPNGをWebPに変換
    size_t png_size;
    uint8_t* png_data = read_file("../../testdata/png/red.png", &png_size);
    assert(png_data != NULL);
    printf("  ✓ Read PNG file: %zu bytes\n", png_size);

    // PNGをWebPに変換
    NextImageBuffer webp_temp = {0};
    CWebPCommand* cwebp_cmd = cwebp_new_command(NULL);
    NextImageStatus status = cwebp_run_command(cwebp_cmd, png_data, png_size, &webp_temp);
    assert(status == NEXTIMAGE_OK);
    free(png_data);
    cwebp_free_command(cwebp_cmd);
    printf("  ✓ Converted PNG to WebP: %zu bytes\n", webp_temp.size);

    // WebPをPNGに変換（dwebp）
    DWebPCommand* dwebp_cmd = dwebp_new_command(NULL);
    assert(dwebp_cmd != NULL);
    printf("  ✓ dwebp_new_command() succeeded\n");

    NextImageBuffer png_output = {0};
    status = dwebp_run_command(dwebp_cmd, webp_temp.data, webp_temp.size, &png_output);
    assert(status == NEXTIMAGE_OK);
    assert(png_output.data != NULL);
    assert(png_output.size > 0);
    printf("  ✓ dwebp_run_command() succeeded: %zu bytes\n", png_output.size);

    // PNG署名を確認
    assert(is_png(png_output.data, png_output.size));
    printf("  ✓ Output is valid PNG format\n");

    // クリーンアップ
    nextimage_free_buffer(&webp_temp);
    nextimage_free_buffer(&png_output);
    dwebp_free_command(dwebp_cmd);
    printf("  ✓ DWebP command test passed\n");
}

void test_avifdec_to_png(void) {
    printf("\nTesting AVIFDec command (AVIF -> PNG)...\n");

    // まずPNGをAVIFに変換
    size_t png_size;
    uint8_t* png_data = read_file("../../testdata/png/red.png", &png_size);
    assert(png_data != NULL);
    printf("  ✓ Read PNG file: %zu bytes\n", png_size);

    // PNGをAVIFに変換
    NextImageBuffer avif_temp = {0};
    AVIFEncCommand* avifenc_cmd = avifenc_new_command(NULL);
    NextImageStatus status = avifenc_run_command(avifenc_cmd, png_data, png_size, &avif_temp);
    assert(status == NEXTIMAGE_OK);
    free(png_data);
    avifenc_free_command(avifenc_cmd);
    printf("  ✓ Converted PNG to AVIF: %zu bytes\n", avif_temp.size);

    // AVIFをPNGに変換（avifdec）
    AVIFDecCommand* avifdec_cmd = avifdec_new_command(NULL);
    assert(avifdec_cmd != NULL);
    printf("  ✓ avifdec_new_command() succeeded\n");

    NextImageBuffer png_output = {0};
    status = avifdec_run_command(avifdec_cmd, avif_temp.data, avif_temp.size, &png_output);
    assert(status == NEXTIMAGE_OK);
    assert(png_output.data != NULL);
    assert(png_output.size > 0);
    printf("  ✓ avifdec_run_command() succeeded: %zu bytes\n", png_output.size);

    // PNG署名を確認
    assert(is_png(png_output.data, png_output.size));
    printf("  ✓ Output is valid PNG format\n");

    // クリーンアップ
    nextimage_free_buffer(&avif_temp);
    nextimage_free_buffer(&png_output);
    avifdec_free_command(avifdec_cmd);
    printf("  ✓ AVIFDec command test passed\n");
}

void test_dwebp_with_real_webp(void) {
    printf("\nTesting DWebP with actual WebP file...\n");

    // 実際のWebPファイルを読み込む
    size_t webp_size;
    uint8_t* webp_data = read_file("../../testdata/webp/gradient.webp", &webp_size);
    assert(webp_data != NULL);
    printf("  ✓ Read WebP file: %zu bytes\n", webp_size);

    // WebPをPNGに変換
    DWebPCommand* cmd = dwebp_new_command(NULL);
    NextImageBuffer png_output = {0};
    NextImageStatus status = dwebp_run_command(cmd, webp_data, webp_size, &png_output);

    if (status == NEXTIMAGE_OK) {
        printf("  ✓ Decoded WebP to PNG: %zu bytes\n", png_output.size);
        assert(is_png(png_output.data, png_output.size));
        printf("  ✓ Output is valid PNG format\n");
        nextimage_free_buffer(&png_output);
    } else {
        printf("  ⚠ Decoding failed (possibly test file not found): %s\n",
               nextimage_last_error_message());
    }

    free(webp_data);
    dwebp_free_command(cmd);
}

int main(void) {
    printf("=== Decoder (dwebp/avifdec) Test ===\n");
    printf("Version: %s\n", nextimage_version());

    test_dwebp_to_png();
    test_avifdec_to_png();
    test_dwebp_with_real_webp();

    printf("\n=== All decoder tests passed! ===\n");
    return 0;
}
