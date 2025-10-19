#include "nextimage.h"
#include "avif.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Create a simple test image (red gradient)
static void create_test_image(uint8_t* data, int width, int height) {
    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            int idx = (y * width + x) * 4;
            data[idx + 0] = (uint8_t)((x * 255) / width);  // R
            data[idx + 1] = (uint8_t)((y * 255) / height); // G
            data[idx + 2] = 128;                            // B
            data[idx + 3] = 255;                            // A
        }
    }
}

int main(void) {
    printf("=== AVIF Test ===\n\n");

    // Test parameters
    const int width = 64;
    const int height = 64;
    const size_t rgba_size = width * height * 4;

    // Create test image
    uint8_t* rgba_data = (uint8_t*)malloc(rgba_size);
    if (!rgba_data) {
        fprintf(stderr, "Failed to allocate test image\n");
        return 1;
    }
    create_test_image(rgba_data, width, height);
    printf("Created test image: %dx%d RGBA\n", width, height);

    // Test 1: Encode with default options
    printf("\n--- Test 1: Encode with default options ---\n");
    NextImageAVIFEncodeOptions enc_opts;
    nextimage_avif_default_encode_options(&enc_opts);
    printf("Default encode options: quality=%d, speed=%d, bit_depth=%d\n",
           enc_opts.quality, enc_opts.speed, enc_opts.bit_depth);

    NextImageBuffer avif_buf;
    NextImageStatus status = nextimage_avif_encode_alloc(
        rgba_data, rgba_size,
        width, height,
        NEXTIMAGE_FORMAT_RGBA,
        &enc_opts,
        &avif_buf
    );

    if (status != NEXTIMAGE_OK) {
        fprintf(stderr, "AVIF encoding failed: %s\n", nextimage_last_error_message());
        free(rgba_data);
        return 1;
    }
    printf("✓ Encoded to AVIF: %zu bytes\n", avif_buf.size);

    // Test 2: Decode the encoded image
    printf("\n--- Test 2: Decode AVIF ---\n");
    NextImageAVIFDecodeOptions dec_opts;
    nextimage_avif_default_decode_options(&dec_opts);
    dec_opts.format = NEXTIMAGE_FORMAT_RGBA;

    NextImageDecodeBuffer decoded;
    status = nextimage_avif_decode_alloc(
        avif_buf.data, avif_buf.size,
        &dec_opts,
        &decoded
    );

    if (status != NEXTIMAGE_OK) {
        fprintf(stderr, "AVIF decoding failed: %s\n", nextimage_last_error_message());
        nextimage_free_buffer(&avif_buf);
        free(rgba_data);
        return 1;
    }
    printf("✓ Decoded AVIF: %dx%d, format=%d, bit_depth=%d\n",
           decoded.width, decoded.height, decoded.format, decoded.bit_depth);
    printf("  Data size: %zu bytes\n", decoded.data_size);

    // Test 3: Verify dimensions
    printf("\n--- Test 3: Verify dimensions ---\n");
    if (decoded.width != width || decoded.height != height) {
        fprintf(stderr, "✗ Dimension mismatch: expected %dx%d, got %dx%d\n",
                width, height, decoded.width, decoded.height);
        nextimage_free_decode_buffer(&decoded);
        nextimage_free_buffer(&avif_buf);
        free(rgba_data);
        return 1;
    }
    printf("✓ Dimensions match: %dx%d\n", decoded.width, decoded.height);

    // Test 4: Test decode_size function
    printf("\n--- Test 4: Test decode_size ---\n");
    int w, h, bit_depth;
    size_t required_size;
    status = nextimage_avif_decode_size(
        avif_buf.data, avif_buf.size,
        &w, &h, &bit_depth, &required_size
    );

    if (status != NEXTIMAGE_OK) {
        fprintf(stderr, "✗ decode_size failed: %s\n", nextimage_last_error_message());
        nextimage_free_decode_buffer(&decoded);
        nextimage_free_buffer(&avif_buf);
        free(rgba_data);
        return 1;
    }
    printf("✓ decode_size: %dx%d, bit_depth=%d, required_size=%zu\n",
           w, h, bit_depth, required_size);

    if (w != width || h != height) {
        fprintf(stderr, "✗ decode_size dimension mismatch\n");
        nextimage_free_decode_buffer(&decoded);
        nextimage_free_buffer(&avif_buf);
        free(rgba_data);
        return 1;
    }

    // Test 5: Test decode_into with user buffer
    printf("\n--- Test 5: Test decode_into ---\n");
    NextImageDecodeBuffer user_buf;
    memset(&user_buf, 0, sizeof(user_buf));
    user_buf.data = (uint8_t*)malloc(required_size);
    user_buf.data_capacity = required_size;
    user_buf.owns_data = 1;

    if (!user_buf.data) {
        fprintf(stderr, "✗ Failed to allocate user buffer\n");
        nextimage_free_decode_buffer(&decoded);
        nextimage_free_buffer(&avif_buf);
        free(rgba_data);
        return 1;
    }

    status = nextimage_avif_decode_into(
        avif_buf.data, avif_buf.size,
        &dec_opts,
        &user_buf
    );

    if (status != NEXTIMAGE_OK) {
        fprintf(stderr, "✗ decode_into failed: %s\n", nextimage_last_error_message());
        nextimage_free_decode_buffer(&user_buf);
        nextimage_free_decode_buffer(&decoded);
        nextimage_free_buffer(&avif_buf);
        free(rgba_data);
        return 1;
    }
    printf("✓ decode_into successful: %dx%d\n", user_buf.width, user_buf.height);

    // Test 6: Test different quality levels
    printf("\n--- Test 6: Test different quality levels ---\n");
    int qualities[] = {10, 50, 90};
    for (int i = 0; i < 3; i++) {
        NextImageAVIFEncodeOptions opts;
        nextimage_avif_default_encode_options(&opts);
        opts.quality = qualities[i];

        NextImageBuffer test_buf;
        status = nextimage_avif_encode_alloc(
            rgba_data, rgba_size,
            width, height,
            NEXTIMAGE_FORMAT_RGBA,
            &opts,
            &test_buf
        );

        if (status != NEXTIMAGE_OK) {
            fprintf(stderr, "✗ Failed to encode with quality %d\n", qualities[i]);
            nextimage_free_decode_buffer(&user_buf);
            nextimage_free_decode_buffer(&decoded);
            nextimage_free_buffer(&avif_buf);
            free(rgba_data);
            return 1;
        }

        printf("  Quality %d: %zu bytes\n", qualities[i], test_buf.size);
        nextimage_free_buffer(&test_buf);
    }
    printf("✓ All quality levels work\n");

    // Test 7: Test different YUV formats
    printf("\n--- Test 7: Test different YUV formats ---\n");
    const char* format_names[] = {"YUV444", "YUV422", "YUV420", "YUV400"};
    for (int fmt = 0; fmt < 4; fmt++) {
        NextImageAVIFEncodeOptions opts;
        nextimage_avif_default_encode_options(&opts);
        opts.yuv_format = fmt;
        opts.quality = 80;

        NextImageBuffer test_buf;
        status = nextimage_avif_encode_alloc(
            rgba_data, rgba_size,
            width, height,
            NEXTIMAGE_FORMAT_RGBA,
            &opts,
            &test_buf
        );

        if (status != NEXTIMAGE_OK) {
            fprintf(stderr, "✗ Failed to encode with %s\n", format_names[fmt]);
            nextimage_free_decode_buffer(&user_buf);
            nextimage_free_decode_buffer(&decoded);
            nextimage_free_buffer(&avif_buf);
            free(rgba_data);
            return 1;
        }

        printf("  %s: %zu bytes\n", format_names[fmt], test_buf.size);
        nextimage_free_buffer(&test_buf);
    }
    printf("✓ All YUV formats work\n");

    // Cleanup
    nextimage_free_decode_buffer(&user_buf);
    nextimage_free_decode_buffer(&decoded);
    nextimage_free_buffer(&avif_buf);
    free(rgba_data);

    printf("\n=== All AVIF tests passed! ===\n");
    return 0;
}
