#include "nextimage.h"
#include "webp.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>

// テスト用の単純な画像データを生成（グラデーション）
static uint8_t* generate_test_image_rgba(int width, int height) {
    size_t size = (size_t)width * height * 4;
    uint8_t* data = malloc(size);
    if (!data) return NULL;

    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            int idx = (y * width + x) * 4;
            data[idx + 0] = (uint8_t)((x * 255) / width);      // R
            data[idx + 1] = (uint8_t)((y * 255) / height);     // G
            data[idx + 2] = 128;                                // B
            data[idx + 3] = 255;                                // A
        }
    }

    return data;
}

// テスト用の赤い画像を生成
static uint8_t* generate_red_image_rgb(int width, int height) {
    size_t size = (size_t)width * height * 3;
    uint8_t* data = malloc(size);
    if (!data) return NULL;

    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            int idx = (y * width + x) * 3;
            data[idx + 0] = 255;  // R
            data[idx + 1] = 0;    // G
            data[idx + 2] = 0;    // B
        }
    }

    return data;
}

void test_webp_default_options(void) {
    printf("\nTesting WebP default options...\n");

    NextImageWebPEncodeOptions enc_opts;
    nextimage_webp_default_encode_options(&enc_opts);

    assert(enc_opts.quality == 75.0f);
    assert(enc_opts.lossless == 0);
    assert(enc_opts.method == 4);
    printf("  ✓ Default encode options correct\n");

    NextImageWebPDecodeOptions dec_opts;
    nextimage_webp_default_decode_options(&dec_opts);

    assert(dec_opts.format == NEXTIMAGE_FORMAT_RGBA);
    assert(dec_opts.use_threads == 0);
    printf("  ✓ Default decode options correct\n");

    printf("  ✓ WebP default options test passed\n");
}

void test_webp_encode_decode_rgba(void) {
    printf("\nTesting WebP encode/decode (RGBA)...\n");

    const int width = 64;
    const int height = 64;

    // テスト画像を生成
    uint8_t* input = generate_test_image_rgba(width, height);
    assert(input != NULL);
    printf("  ✓ Generated test image: %dx%d RGBA\n", width, height);

    // エンコード
    NextImageWebPEncodeOptions enc_opts;
    nextimage_webp_default_encode_options(&enc_opts);
    enc_opts.quality = 90.0f;

    NextImageEncodeBuffer encoded = {0};
    NextImageStatus status = nextimage_webp_encode_alloc(
        input, width * height * 4,
        width, height,
        NEXTIMAGE_FORMAT_RGBA,
        &enc_opts,
        &encoded
    );

    assert(status == NEXTIMAGE_OK);
    assert(encoded.data != NULL);
    assert(encoded.size > 0);
    printf("  ✓ Encoded to WebP: %zu bytes\n", encoded.size);

    // デコード
    NextImageWebPDecodeOptions dec_opts;
    nextimage_webp_default_decode_options(&dec_opts);
    dec_opts.format = NEXTIMAGE_FORMAT_RGBA;

    NextImageDecodeBuffer decoded = {0};
    status = nextimage_webp_decode_alloc(
        encoded.data, encoded.size,
        &dec_opts,
        &decoded
    );

    assert(status == NEXTIMAGE_OK);
    assert(decoded.data != NULL);
    assert(decoded.width == width);
    assert(decoded.height == height);
    assert(decoded.format == NEXTIMAGE_FORMAT_RGBA);
    assert(decoded.bit_depth == 8);
    printf("  ✓ Decoded from WebP: %dx%d, %zu bytes\n",
           decoded.width, decoded.height, decoded.data_size);

    // クリーンアップ
    free(input);
    nextimage_free_encode_buffer(&encoded);
    nextimage_free_decode_buffer(&decoded);

    printf("  ✓ WebP encode/decode (RGBA) test passed\n");
}

void test_webp_encode_decode_rgb(void) {
    printf("\nTesting WebP encode/decode (RGB)...\n");

    const int width = 32;
    const int height = 32;

    // 赤い画像を生成
    uint8_t* input = generate_red_image_rgb(width, height);
    assert(input != NULL);
    printf("  ✓ Generated red image: %dx%d RGB\n", width, height);

    // エンコード
    NextImageEncodeBuffer encoded = {0};
    NextImageStatus status = nextimage_webp_encode_alloc(
        input, width * height * 3,
        width, height,
        NEXTIMAGE_FORMAT_RGB,
        NULL,  // デフォルトオプション
        &encoded
    );

    assert(status == NEXTIMAGE_OK);
    assert(encoded.data != NULL);
    printf("  ✓ Encoded to WebP: %zu bytes\n", encoded.size);

    // デコード (RGB形式で)
    NextImageWebPDecodeOptions dec_opts;
    nextimage_webp_default_decode_options(&dec_opts);
    dec_opts.format = NEXTIMAGE_FORMAT_RGB;

    NextImageDecodeBuffer decoded = {0};
    status = nextimage_webp_decode_alloc(
        encoded.data, encoded.size,
        &dec_opts,
        &decoded
    );

    assert(status == NEXTIMAGE_OK);
    assert(decoded.format == NEXTIMAGE_FORMAT_RGB);
    printf("  ✓ Decoded as RGB format\n");

    // クリーンアップ
    free(input);
    nextimage_free_encode_buffer(&encoded);
    nextimage_free_decode_buffer(&decoded);

    printf("  ✓ WebP encode/decode (RGB) test passed\n");
}

void test_webp_decode_size(void) {
    printf("\nTesting WebP decode size calculation...\n");

    const int width = 48;
    const int height = 48;

    // テスト画像をエンコード
    uint8_t* input = generate_test_image_rgba(width, height);
    NextImageEncodeBuffer encoded = {0};

    nextimage_webp_encode_alloc(
        input, width * height * 4,
        width, height,
        NEXTIMAGE_FORMAT_RGBA,
        NULL,
        &encoded
    );

    free(input);

    // サイズ計算
    int w = 0, h = 0;
    size_t required_size = 0;
    NextImageStatus status = nextimage_webp_decode_size(
        encoded.data, encoded.size,
        &w, &h,
        &required_size
    );

    assert(status == NEXTIMAGE_OK);
    assert(w == width);
    assert(h == height);
    assert(required_size == (size_t)width * height * 4);
    printf("  ✓ Size calculation: %dx%d, %zu bytes required\n", w, h, required_size);

    nextimage_free_encode_buffer(&encoded);

    printf("  ✓ WebP decode size test passed\n");
}

void test_webp_decode_into(void) {
    printf("\nTesting WebP decode into user buffer...\n");

    const int width = 40;
    const int height = 40;

    // エンコード
    uint8_t* input = generate_test_image_rgba(width, height);
    NextImageEncodeBuffer encoded = {0};

    nextimage_webp_encode_alloc(
        input, width * height * 4,
        width, height,
        NEXTIMAGE_FORMAT_RGBA,
        NULL,
        &encoded
    );

    free(input);

    // ユーザーバッファを用意
    size_t buffer_size = (size_t)width * height * 4;
    uint8_t* user_buffer = malloc(buffer_size);
    assert(user_buffer != NULL);

    NextImageDecodeBuffer decoded = {0};
    decoded.data = user_buffer;
    decoded.data_capacity = buffer_size;
    decoded.owns_data = 0;  // ユーザーが所有

    // デコード
    NextImageStatus status = nextimage_webp_decode_into(
        encoded.data, encoded.size,
        NULL,
        &decoded
    );

    assert(status == NEXTIMAGE_OK);
    assert(decoded.data == user_buffer);
    assert(decoded.width == width);
    assert(decoded.height == height);
    printf("  ✓ Decoded into user buffer: %dx%d\n", decoded.width, decoded.height);

    // クリーンアップ（user_bufferは手動で解放）
    nextimage_free_encode_buffer(&encoded);
    free(user_buffer);

    printf("  ✓ WebP decode into buffer test passed\n");
}

void test_webp_lossless(void) {
    printf("\nTesting WebP lossless encoding...\n");

    const int width = 32;
    const int height = 32;

    uint8_t* input = generate_test_image_rgba(width, height);

    // ロスレスエンコード
    NextImageWebPEncodeOptions opts;
    nextimage_webp_default_encode_options(&opts);
    opts.lossless = 1;
    opts.quality = 100.0f;

    NextImageEncodeBuffer encoded = {0};
    NextImageStatus status = nextimage_webp_encode_alloc(
        input, width * height * 4,
        width, height,
        NEXTIMAGE_FORMAT_RGBA,
        &opts,
        &encoded
    );

    assert(status == NEXTIMAGE_OK);
    assert(encoded.size > 0);
    printf("  ✓ Lossless encode: %zu bytes\n", encoded.size);

    // デコード
    NextImageDecodeBuffer decoded = {0};
    status = nextimage_webp_decode_alloc(
        encoded.data, encoded.size,
        NULL,
        &decoded
    );

    assert(status == NEXTIMAGE_OK);
    printf("  ✓ Lossless decode successful\n");

    // クリーンアップ
    free(input);
    nextimage_free_encode_buffer(&encoded);
    nextimage_free_decode_buffer(&decoded);

    printf("  ✓ WebP lossless test passed\n");
}

void test_webp_error_handling(void) {
    printf("\nTesting WebP error handling...\n");

    NextImageEncodeBuffer encoded = {0};
    NextImageDecodeBuffer decoded = {0};

    // NULL入力
    NextImageStatus status = nextimage_webp_encode_alloc(
        NULL, 100, 10, 10,
        NEXTIMAGE_FORMAT_RGBA,
        NULL,
        &encoded
    );
    assert(status == NEXTIMAGE_ERROR_INVALID_PARAM);
    printf("  ✓ NULL input handled\n");

    // 無効な寸法
    uint8_t dummy[100];
    status = nextimage_webp_encode_alloc(
        dummy, 100, 0, 0,
        NEXTIMAGE_FORMAT_RGBA,
        NULL,
        &encoded
    );
    assert(status == NEXTIMAGE_ERROR_INVALID_PARAM);
    printf("  ✓ Invalid dimensions handled\n");

    // 無効なWebPデータ
    uint8_t invalid_data[10] = {0};
    status = nextimage_webp_decode_alloc(
        invalid_data, 10,
        NULL,
        &decoded
    );
    assert(status == NEXTIMAGE_ERROR_DECODE_FAILED);
    printf("  ✓ Invalid WebP data handled\n");

    printf("  ✓ WebP error handling test passed\n");
}

int main(void) {
    printf("=== libnextimage WebP Test ===\n");

    test_webp_default_options();
    test_webp_encode_decode_rgba();
    test_webp_encode_decode_rgb();
    test_webp_decode_size();
    test_webp_decode_into();
    test_webp_lossless();
    test_webp_error_handling();

    printf("\n=== All WebP tests passed! ===\n");
    return 0;
}
