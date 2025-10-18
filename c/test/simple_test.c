#include "nextimage.h"
#include "webp.h"
#include "avif.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>

// ա�뒭��
static uint8_t* read_file(const char* path, size_t* size) {
    FILE* f = fopen(path, "rb");
    if (!f) {
        fprintf(stderr, "Failed to open file: %s\n", path);
        return NULL;
    }

    fseek(f, 0, SEEK_END);
    *size = ftell(f);
    fseek(f, 0, SEEK_SET);

    uint8_t* data = malloc(*size);
    if (!data) {
        fclose(f);
        return NULL;
    }

    size_t read = fread(data, 1, *size, f);
    fclose(f);

    if (read != *size) {
        free(data);
        return NULL;
    }

    return data;
}

// ա��k�M��
static int write_file(const char* path, const uint8_t* data, size_t size) {
    FILE* f = fopen(path, "wb");
    if (!f) {
        fprintf(stderr, "Failed to open file for writing: %s\n", path);
        return 0;
    }

    size_t written = fwrite(data, 1, size, f);
    fclose(f);

    return written == size;
}

void test_webp_encode_jpeg(void) {
    printf("\nTesting WebP encode from JPEG...\n");

    // JPEGա�뒭��
    size_t input_size;
    uint8_t* input_data = read_file("../../testdata/jpeg/gradient.jpg", &input_size);
    assert(input_data != NULL);
    printf("   Read JPEG file: %zu bytes\n", input_size);

    // WebPk����
    NextImageWebPEncodeOptions opts;
    nextimage_webp_default_encode_options(&opts);
    opts.quality = 80.0f;

    NextImageEncodeBuffer encoded = {0};
    NextImageStatus status = nextimage_webp_encode_alloc(
        input_data,
        input_size,
        &opts,
        &encoded
    );

    free(input_data);

    assert(status == NEXTIMAGE_OK);
    assert(encoded.data != NULL);
    assert(encoded.size > 0);
    printf("   Encoded to WebP: %zu bytes\n", encoded.size);

    // P���X�׷��	
    write_file("/tmp/test_output.webp", encoded.data, encoded.size);
    printf("   Saved to /tmp/test_output.webp\n");

    nextimage_free_encode_buffer(&encoded);
    printf("   WebP encode from JPEG test passed\n");
}

void test_avif_encode_png(void) {
    printf("\nTesting AVIF encode from PNG...\n");

    // PNGա�뒭��
    size_t input_size;
    uint8_t* input_data = read_file("../../testdata/png/red.png", &input_size);
    assert(input_data != NULL);
    printf("   Read PNG file: %zu bytes\n", input_size);

    // AVIFk����
    NextImageAVIFEncodeOptions opts;
    nextimage_avif_default_encode_options(&opts);
    opts.quality = 60;
    opts.speed = 8;

    NextImageEncodeBuffer encoded = {0};
    NextImageStatus status = nextimage_avif_encode_alloc(
        input_data,
        input_size,
        &opts,
        &encoded
    );

    free(input_data);

    assert(status == NEXTIMAGE_OK);
    assert(encoded.data != NULL);
    assert(encoded.size > 0);
    printf("   Encoded to AVIF: %zu bytes\n", encoded.size);

    // P���X�׷��	
    write_file("/tmp/test_output.avif", encoded.data, encoded.size);
    printf("   Saved to /tmp/test_output.avif\n");

    nextimage_free_encode_buffer(&encoded);
    printf("   AVIF encode from PNG test passed\n");
}

void test_webp_decode(void) {
    printf("\nTesting WebP decode...\n");

    // ~ZJPEG�WebPk����
    size_t jpeg_size;
    uint8_t* jpeg_data = read_file("../../testdata/jpeg/test.jpg", &jpeg_size);
    assert(jpeg_data != NULL);

    NextImageWebPEncodeOptions enc_opts;
    nextimage_webp_default_encode_options(&enc_opts);

    NextImageEncodeBuffer webp_encoded = {0};
    NextImageStatus status = nextimage_webp_encode_alloc(
        jpeg_data,
        jpeg_size,
        &enc_opts,
        &webp_encoded
    );
    free(jpeg_data);

    assert(status == NEXTIMAGE_OK);
    printf("   Encoded JPEG to WebP: %zu bytes\n", webp_encoded.size);

    // WebP�ǳ��
    NextImageWebPDecodeOptions dec_opts;
    nextimage_webp_default_decode_options(&dec_opts);
    dec_opts.format = NEXTIMAGE_FORMAT_RGBA;

    NextImageDecodeBuffer decoded = {0};
    status = nextimage_webp_decode_alloc(
        webp_encoded.data,
        webp_encoded.size,
        &dec_opts,
        &decoded
    );

    nextimage_free_encode_buffer(&webp_encoded);

    assert(status == NEXTIMAGE_OK);
    assert(decoded.data != NULL);
    assert(decoded.width > 0);
    assert(decoded.height > 0);
    printf("   Decoded WebP: %dx%d, %zu bytes\n",
           decoded.width, decoded.height, decoded.data_size);

    nextimage_free_decode_buffer(&decoded);
    printf("   WebP decode test passed\n");
}

void test_instance_based_webp_encoder(void) {
    printf("\nTesting instance-based WebP encoder...\n");

    // ������\
    NextImageWebPEncodeOptions opts;
    nextimage_webp_default_encode_options(&opts);
    opts.quality = 85.0f;

    NextImageWebPEncoder* encoder = nextimage_webp_encoder_create(&opts);
    assert(encoder != NULL);
    printf("   Created WebP encoder\n");

    // pn;ϒ����
    const char* test_files[] = {
        "../../testdata/jpeg/gradient.jpg",
        "../../testdata/png/red.png"
    };

    for (int i = 0; i < 2; i++) {
        size_t input_size;
        uint8_t* input_data = read_file(test_files[i], &input_size);
        assert(input_data != NULL);

        NextImageEncodeBuffer encoded = {0};
        NextImageStatus status = nextimage_webp_encoder_encode(
            encoder,
            input_data,
            input_size,
            &encoded
        );

        free(input_data);

        assert(status == NEXTIMAGE_OK);
        assert(encoded.size > 0);
        printf("   Encoded %s: %zu bytes\n", test_files[i], encoded.size);

        nextimage_free_encode_buffer(&encoded);
    }

    // ������4�
    nextimage_webp_encoder_destroy(encoder);
    printf("   Instance-based WebP encoder test passed\n");
}

void test_instance_based_avif_encoder(void) {
    printf("\nTesting instance-based AVIF encoder...\n");

    // ������\
    NextImageAVIFEncodeOptions opts;
    nextimage_avif_default_encode_options(&opts);
    opts.quality = 50;
    opts.speed = 8;

    NextImageAVIFEncoder* encoder = nextimage_avif_encoder_create(&opts);
    assert(encoder != NULL);
    printf("   Created AVIF encoder\n");

    // ;ϒ����
    size_t input_size;
    uint8_t* input_data = read_file("../../testdata/png/blue.png", &input_size);
    assert(input_data != NULL);

    NextImageEncodeBuffer encoded = {0};
    NextImageStatus status = nextimage_avif_encoder_encode(
        encoder,
        input_data,
        input_size,
        &encoded
    );

    free(input_data);

    assert(status == NEXTIMAGE_OK);
    assert(encoded.size > 0);
    printf("   Encoded to AVIF: %zu bytes\n", encoded.size);

    nextimage_free_encode_buffer(&encoded);

    // ������4�
    nextimage_avif_encoder_destroy(encoder);
    printf("   Instance-based AVIF encoder test passed\n");
}

void test_gif2webp(void) {
    printf("\nTesting GIF to WebP conversion...\n");

    // Read static GIF
    size_t gif_size;
    uint8_t* gif_data = read_file("../../testdata/gif/static.gif", &gif_size);
    assert(gif_data != NULL);
    printf("   Read static GIF file: %zu bytes\n", gif_size);

    // Encode to WebP
    NextImageWebPEncodeOptions opts;
    nextimage_webp_default_encode_options(&opts);
    opts.quality = 80.0f;

    NextImageEncodeBuffer encoded = {0};
    NextImageStatus status = nextimage_gif2webp_alloc(
        gif_data,
        gif_size,
        &opts,
        &encoded
    );

    free(gif_data);

    if (status != NEXTIMAGE_OK) {
        printf("   ERROR: Conversion failed with status %d\n", status);
        printf("   Error message: %s\n", nextimage_last_error_message());
        return;
    }

    assert(status == NEXTIMAGE_OK);
    assert(encoded.data != NULL);
    assert(encoded.size > 0);
    printf("   Converted GIF to WebP: %zu bytes\n", encoded.size);

    write_file("/tmp/test_gif2webp.webp", encoded.data, encoded.size);
    printf("   Saved to /tmp/test_gif2webp.webp\n");

    nextimage_free_encode_buffer(&encoded);
    printf("   GIF to WebP conversion test passed\n");
}

void test_gif2webp_animated(void) {
    printf("\nTesting animated GIF to WebP conversion...\n");

    size_t gif_size;
    uint8_t* gif_data = read_file("../../testdata/gif/animated.gif", &gif_size);
    assert(gif_data != NULL);
    printf("   Read animated GIF file: %zu bytes\n", gif_size);

    NextImageWebPEncodeOptions opts;
    nextimage_webp_default_encode_options(&opts);
    opts.quality = 80.0f;

    NextImageEncodeBuffer encoded = {0};
    NextImageStatus status = nextimage_gif2webp_alloc(
        gif_data,
        gif_size,
        &opts,
        &encoded
    );

    free(gif_data);

    if (status != NEXTIMAGE_OK) {
        printf("   ERROR: Conversion failed with status %d\n", status);
        printf("   Error message: %s\n", nextimage_last_error_message());
        return;
    }

    assert(status == NEXTIMAGE_OK);
    assert(encoded.data != NULL);
    assert(encoded.size > 0);
    printf("   Converted animated GIF to WebP: %zu bytes\n", encoded.size);

    write_file("/tmp/test_gif2webp_animated.webp", encoded.data, encoded.size);
    printf("   Saved to /tmp/test_gif2webp_animated.webp\n");

    nextimage_free_encode_buffer(&encoded);
    printf("   Animated GIF to WebP conversion test passed\n");
}

void test_webp2gif(void) {
    printf("\nTesting WebP to GIF conversion...\n");

    // First encode a PNG to WebP
    size_t png_size;
    uint8_t* png_data = read_file("../../testdata/png/red.png", &png_size);
    assert(png_data != NULL);
    printf("   Read PNG file: %zu bytes\n", png_size);

    NextImageWebPEncodeOptions webp_opts;
    nextimage_webp_default_encode_options(&webp_opts);
    webp_opts.quality = 90.0f;

    NextImageEncodeBuffer webp_encoded = {0};
    NextImageStatus status = nextimage_webp_encode_alloc(
        png_data,
        png_size,
        &webp_opts,
        &webp_encoded
    );
    free(png_data);

    assert(status == NEXTIMAGE_OK);
    printf("   Encoded to WebP: %zu bytes\n", webp_encoded.size);

    // Convert WebP to GIF
    NextImageEncodeBuffer gif_encoded = {0};
    status = nextimage_webp2gif_alloc(
        webp_encoded.data,
        webp_encoded.size,
        &gif_encoded
    );

    nextimage_free_encode_buffer(&webp_encoded);

    if (status != NEXTIMAGE_OK) {
        printf("   ERROR: WebP to GIF conversion failed with status %d\n", status);
        printf("   Error message: %s\n", nextimage_last_error_message());
        return;
    }

    assert(status == NEXTIMAGE_OK);
    assert(gif_encoded.data != NULL);
    assert(gif_encoded.size > 0);
    printf("   Converted WebP to GIF: %zu bytes\n", gif_encoded.size);

    write_file("/tmp/test_webp2gif.gif", gif_encoded.data, gif_encoded.size);
    printf("   Saved to /tmp/test_webp2gif.gif\n");

    nextimage_free_encode_buffer(&gif_encoded);
    printf("   WebP to GIF conversion test passed\n");
}

int main(void) {
    printf("=== NextImage Simple API Test ===\n");
    printf("Version: %s\n", nextimage_version());

    test_webp_encode_jpeg();
    test_avif_encode_png();
    test_webp_decode();
    test_instance_based_webp_encoder();
    test_instance_based_avif_encoder();
    test_gif2webp();
    test_gif2webp_animated();
    test_webp2gif();

    printf("\n=== All tests passed! ===\n");
    return 0;
}
