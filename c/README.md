# libnextimage - C Library

High-performance WebP, AVIF, and GIF image processing library in C.

## Features

- **WebP Encoding & Decoding**: Fast WebP image processing with full control over quality and encoding options
- **AVIF Encoding & Decoding**: Modern AVIF format support with quality and speed presets
- **GIF Conversion**: Convert between GIF and WebP formats, including animated GIFs
- **Zero Dependencies** (for linking): Single shared library with all dependencies bundled
- **Cross-Platform**: Supports macOS (ARM64/x64), Linux (x64/ARM64), and Windows (x64)
- **Memory Safe**: Proper resource management with explicit cleanup functions
- **Thread Safe**: Can be used safely from multiple threads

## Building

### Prerequisites

- CMake 3.15+
- C compiler (GCC, Clang, or MSVC)
- Git (for downloading dependencies)

### Build Commands

```bash
# Create build directory
mkdir -p build
cd build

# Configure
cmake ..

# Build
cmake --build .

# Install (optional)
sudo cmake --install .
```

### Build Options

```bash
# Debug build
cmake -DCMAKE_BUILD_TYPE=Debug ..

# Release build with optimizations
cmake -DCMAKE_BUILD_TYPE=Release ..

# Build with static libraries
cmake -DBUILD_SHARED_LIBS=OFF ..
```

## Installation

### System-wide Installation

```bash
cd build
sudo cmake --install .
```

This installs:
- Headers to `/usr/local/include/nextimage/`
- Library to `/usr/local/lib/`
- CMake config files

### Using in Your Project

#### With CMake

```cmake
find_package(libnextimage REQUIRED)

add_executable(myapp main.c)
target_link_libraries(myapp PRIVATE libnextimage::libnextimage)
```

#### With pkg-config

```bash
gcc main.c $(pkg-config --cflags --libs libnextimage) -o myapp
```

#### Manual Linking

```bash
gcc main.c -I/usr/local/include -L/usr/local/lib -lnextimage -o myapp
```

## Quick Start

### WebP Encoding

```c
#include <stdio.h>
#include <stdlib.h>
#include <nextimage/webp_encoder.h>

int main() {
    // Read input image
    FILE* fp = fopen("input.jpg", "rb");
    fseek(fp, 0, SEEK_END);
    size_t size = ftell(fp);
    fseek(fp, 0, SEEK_SET);

    uint8_t* input_data = malloc(size);
    fread(input_data, 1, size, fp);
    fclose(fp);

    // Create encoder with options
    WebPEncoderOptions* opts = webp_encoder_create_default_options();
    opts->quality = 80;
    opts->method = 6;

    WebPEncoderCommand* encoder = webp_encoder_new_command(opts);
    webp_encoder_free_options(opts);

    if (!encoder) {
        fprintf(stderr, "Failed to create encoder\n");
        return 1;
    }

    // Encode to WebP
    NextImageBuffer output = {0};
    NextImageStatus status = webp_encoder_run_command(
        encoder, input_data, size, &output
    );

    if (status != NEXTIMAGE_STATUS_OK) {
        fprintf(stderr, "Encoding failed: %s\n", nextimage_last_error_message());
        webp_encoder_free_command(encoder);
        return 1;
    }

    // Save output
    fp = fopen("output.webp", "wb");
    fwrite(output.data, 1, output.size, fp);
    fclose(fp);

    printf("Converted: %zu bytes → %zu bytes\n", size, output.size);

    // Cleanup
    nextimage_free_buffer(&output);
    webp_encoder_free_command(encoder);
    free(input_data);

    return 0;
}
```

### WebP Decoding

```c
#include <stdio.h>
#include <stdlib.h>
#include <nextimage/webp_decoder.h>

int main() {
    // Read WebP file
    FILE* fp = fopen("image.webp", "rb");
    fseek(fp, 0, SEEK_END);
    size_t size = ftell(fp);
    fseek(fp, 0, SEEK_SET);

    uint8_t* webp_data = malloc(size);
    fread(webp_data, 1, size, fp);
    fclose(fp);

    // Create decoder
    WebPDecoderOptions* opts = webp_decoder_create_default_options();
    WebPDecoderCommand* decoder = webp_decoder_new_command(opts);
    webp_decoder_free_options(opts);

    // Decode to RGBA
    NextImageBuffer output = {0};
    int width, height;
    NextImageStatus status = webp_decoder_run_command(
        decoder, webp_data, size, &output, &width, &height
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        printf("Width: %d, Height: %d\n", width, height);
        printf("Data size: %zu bytes\n", output.size);
    }

    // Cleanup
    nextimage_free_buffer(&output);
    webp_decoder_free_command(decoder);
    free(webp_data);

    return 0;
}
```

### AVIF Encoding

```c
#include <nextimage/avif_encoder.h>

int main() {
    // Read input (similar to WebP example)
    uint8_t* input_data;
    size_t size;
    // ... read file ...

    // Create encoder with options
    AVIFEncoderOptions* opts = avif_encoder_create_default_options();
    opts->quality = 60;
    opts->speed = 6;

    AVIFEncoderCommand* encoder = avif_encoder_new_command(opts);
    avif_encoder_free_options(opts);

    // Encode to AVIF
    NextImageBuffer output = {0};
    NextImageStatus status = avif_encoder_run_command(
        encoder, input_data, size, &output
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        // Save output
        FILE* fp = fopen("output.avif", "wb");
        fwrite(output.data, 1, output.size, fp);
        fclose(fp);
    }

    // Cleanup
    nextimage_free_buffer(&output);
    avif_encoder_free_command(encoder);

    return 0;
}
```

### AVIF Decoding

```c
#include <nextimage/avif_decoder.h>

int main() {
    // Read AVIF file (similar to previous examples)
    uint8_t* avif_data;
    size_t size;
    // ... read file ...

    // Create decoder
    AVIFDecoderOptions* opts = avif_decoder_create_default_options();
    AVIFDecoderCommand* decoder = avif_decoder_new_command(opts);
    avif_decoder_free_options(opts);

    // Decode to RGBA
    NextImageBuffer output = {0};
    int width, height;
    NextImageStatus status = avif_decoder_run_command(
        decoder, avif_data, size, &output, &width, &height
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        printf("Decoded %dx%d AVIF image\n", width, height);
    }

    // Cleanup
    nextimage_free_buffer(&output);
    avif_decoder_free_command(decoder);

    return 0;
}
```

### GIF to WebP Conversion

```c
#include <nextimage/gif2webp.h>

int main() {
    // Read GIF file
    uint8_t* gif_data;
    size_t size;
    // ... read file ...

    // Create converter with options
    Gif2WebPOptions* opts = gif2webp_create_default_options();
    opts->quality = 80;
    opts->method = 6;

    Gif2WebPCommand* cmd = gif2webp_new_command(opts);
    gif2webp_free_options(opts);

    // Convert GIF to WebP (preserves animation)
    NextImageBuffer output = {0};
    NextImageStatus status = gif2webp_run_command(
        cmd, gif_data, size, &output
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        // Save output
        FILE* fp = fopen("animated.webp", "wb");
        fwrite(output.data, 1, output.size, fp);
        fclose(fp);

        float compression = (1.0f - (float)output.size / size) * 100;
        printf("Compression: %.1f%%\n", compression);
    }

    // Cleanup
    nextimage_free_buffer(&output);
    gif2webp_free_command(cmd);

    return 0;
}
```

### WebP to GIF Conversion

```c
#include <nextimage/webp2gif.h>

int main() {
    // Read WebP file
    uint8_t* webp_data;
    size_t size;
    // ... read file ...

    // Create converter
    WebP2GifOptions* opts = webp2gif_create_default_options();
    WebP2GifCommand* cmd = webp2gif_new_command(opts);
    webp2gif_free_options(opts);

    // Convert WebP to GIF
    NextImageBuffer output = {0};
    NextImageStatus status = webp2gif_run_command(
        cmd, webp_data, size, &output
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        // Save output
        FILE* fp = fopen("output.gif", "wb");
        fwrite(output.data, 1, output.size, fp);
        fclose(fp);
    }

    // Cleanup
    nextimage_free_buffer(&output);
    webp2gif_free_command(cmd);

    return 0;
}
```

## API Reference

### Common Types

```c
// Status codes
typedef enum {
    NEXTIMAGE_STATUS_OK = 0,
    NEXTIMAGE_STATUS_ERROR_INVALID_PARAMETER = 1,
    NEXTIMAGE_STATUS_ERROR_OUT_OF_MEMORY = 2,
    NEXTIMAGE_STATUS_ERROR_DECODE_FAILED = 3,
    NEXTIMAGE_STATUS_ERROR_ENCODE_FAILED = 4,
    NEXTIMAGE_STATUS_ERROR_UNSUPPORTED_FORMAT = 5,
    NEXTIMAGE_STATUS_ERROR_FILE_NOT_FOUND = 6,
    NEXTIMAGE_STATUS_ERROR_INTERNAL = 99
} NextImageStatus;

// Buffer for output data
typedef struct {
    uint8_t* data;
    size_t size;
} NextImageBuffer;

// Free buffer allocated by library
void nextimage_free_buffer(NextImageBuffer* buffer);

// Error handling
const char* nextimage_last_error_message(void);
void nextimage_clear_error(void);
```

### WebP Encoder

```c
// Header: nextimage/webp_encoder.h

typedef struct {
    float quality;         // 0-100, default: 75
    int lossless;          // 0 or 1
    int method;            // 0-6, default: 4
    int preset;            // 0=default, 1=picture, 2=photo, 3=drawing, 4=icon, 5=text
    int image_hint;        // 0=default, 1=picture, 2=photo, 3=graph

    int target_size;       // Target file size in bytes
    float target_psnr;     // Target PSNR in dB
    int segments;          // 1-4
    int sns_strength;      // 0-100
    int filter_strength;   // 0-100
    int filter_sharpness;  // 0-7
    int filter_type;       // 0=simple, 1=strong
    int autofilter;        // 0 or 1

    int alpha_compression; // 0 or 1
    int alpha_filtering;   // 0=none, 1=fast, 2=best
    int alpha_quality;     // 0-100

    int pass;              // 1-10
    int show_compressed;   // 0 or 1
    int preprocessing;     // 0-7
    int partitions;        // 0-3
    int partition_limit;   // 0-100

    int emulate_jpeg_size; // 0 or 1
    int thread_level;      // 0 or 1
    int low_memory;        // 0 or 1
    int near_lossless;     // 0-100
    int exact;             // 0 or 1
    int use_delta_palette; // 0 or 1
    int use_sharp_yuv;     // 0 or 1

    int qmin;              // 0-100
    int qmax;              // 0-100

    // ... more options
} WebPEncoderOptions;

// Create default options
WebPEncoderOptions* webp_encoder_create_default_options(void);

// Free options
void webp_encoder_free_options(WebPEncoderOptions* options);

// Create encoder command
WebPEncoderCommand* webp_encoder_new_command(const WebPEncoderOptions* options);

// Run encoding
NextImageStatus webp_encoder_run_command(
    WebPEncoderCommand* cmd,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output
);

// Free encoder
void webp_encoder_free_command(WebPEncoderCommand* cmd);
```

### WebP Decoder

```c
// Header: nextimage/webp_decoder.h

typedef struct {
    int format;            // Pixel format (RGBA, BGRA, RGB, BGR)
    int use_threads;       // 0 or 1
    int bypass_filtering;  // 0 or 1
    int no_fancy_upsampling; // 0 or 1

    int crop_x;
    int crop_y;
    int crop_width;
    int crop_height;

    int scale_width;
    int scale_height;
} WebPDecoderOptions;

// Similar functions as encoder
WebPDecoderOptions* webp_decoder_create_default_options(void);
void webp_decoder_free_options(WebPDecoderOptions* options);
WebPDecoderCommand* webp_decoder_new_command(const WebPDecoderOptions* options);

NextImageStatus webp_decoder_run_command(
    WebPDecoderCommand* cmd,
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageBuffer* output,
    int* width,
    int* height
);

void webp_decoder_free_command(WebPDecoderCommand* cmd);
```

### AVIF Encoder

```c
// Header: nextimage/avif_encoder.h

typedef struct {
    int quality;           // 0-100, default: 60
    int quality_alpha;     // 0-100, default: -1
    int speed;             // 0-10, default: 6

    int bit_depth;         // 8, 10, or 12
    int yuv_format;        // YUV444, YUV422, YUV420, YUV400

    int lossless;          // 0 or 1
    int sharp_yuv;         // 0 or 1
    int target_size;       // Target file size

    int jobs;              // -1=all cores, 0=auto, >0=thread count

    int tile_rows_log2;    // 0-6
    int tile_cols_log2;    // 0-6
    int auto_tiling;       // 0 or 1

    // ... more options
} AVIFEncoderOptions;

// Similar API as WebP encoder
AVIFEncoderOptions* avif_encoder_create_default_options(void);
void avif_encoder_free_options(AVIFEncoderOptions* options);
AVIFEncoderCommand* avif_encoder_new_command(const AVIFEncoderOptions* options);

NextImageStatus avif_encoder_run_command(
    AVIFEncoderCommand* cmd,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output
);

void avif_encoder_free_command(AVIFEncoderCommand* cmd);
```

### AVIF Decoder

```c
// Header: nextimage/avif_decoder.h

typedef struct {
    int format;                  // Pixel format
    int jobs;                    // Thread count
    int chroma_upsampling;       // Upsampling mode

    int ignore_exif;             // 0 or 1
    int ignore_xmp;              // 0 or 1
    int ignore_icc;              // 0 or 1

    int image_size_limit;        // Max image size
    int image_dimension_limit;   // Max dimension
} AVIFDecoderOptions;

// Similar API as WebP decoder
AVIFDecoderOptions* avif_decoder_create_default_options(void);
void avif_decoder_free_options(AVIFDecoderOptions* options);
AVIFDecoderCommand* avif_decoder_new_command(const AVIFDecoderOptions* options);

NextImageStatus avif_decoder_run_command(
    AVIFDecoderCommand* cmd,
    const uint8_t* avif_data,
    size_t avif_size,
    NextImageBuffer* output,
    int* width,
    int* height
);

void avif_decoder_free_command(AVIFDecoderCommand* cmd);
```

### GIF2WebP

```c
// Header: nextimage/gif2webp.h

// Uses same options structure as WebPEncoderOptions
typedef WebPEncoderOptions Gif2WebPOptions;

Gif2WebPOptions* gif2webp_create_default_options(void);
void gif2webp_free_options(Gif2WebPOptions* options);
Gif2WebPCommand* gif2webp_new_command(const Gif2WebPOptions* options);

NextImageStatus gif2webp_run_command(
    Gif2WebPCommand* cmd,
    const uint8_t* gif_data,
    size_t gif_size,
    NextImageBuffer* output
);

void gif2webp_free_command(Gif2WebPCommand* cmd);
```

### WebP2GIF

```c
// Header: nextimage/webp2gif.h

typedef struct {
    int reserved;  // Reserved for future use
} WebP2GifOptions;

WebP2GifOptions* webp2gif_create_default_options(void);
void webp2gif_free_options(WebP2GifOptions* options);
WebP2GifCommand* webp2gif_new_command(const WebP2GifOptions* options);

NextImageStatus webp2gif_run_command(
    WebP2GifCommand* cmd,
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageBuffer* output
);

void webp2gif_free_command(WebP2GifCommand* cmd);
```

## Advanced Usage

### Error Handling

```c
NextImageStatus status = webp_encoder_run_command(encoder, data, size, &output);

if (status != NEXTIMAGE_STATUS_OK) {
    const char* error_msg = nextimage_last_error_message();

    switch (status) {
        case NEXTIMAGE_STATUS_ERROR_INVALID_PARAMETER:
            fprintf(stderr, "Invalid parameter: %s\n", error_msg);
            break;
        case NEXTIMAGE_STATUS_ERROR_OUT_OF_MEMORY:
            fprintf(stderr, "Out of memory: %s\n", error_msg);
            break;
        case NEXTIMAGE_STATUS_ERROR_ENCODE_FAILED:
            fprintf(stderr, "Encoding failed: %s\n", error_msg);
            break;
        default:
            fprintf(stderr, "Error: %s\n", error_msg);
    }

    nextimage_clear_error();
}
```

### Batch Processing

```c
#include <dirent.h>
#include <string.h>

void convert_directory(const char* dir_path) {
    DIR* dir = opendir(dir_path);
    struct dirent* entry;

    // Create encoder once, reuse for all files
    WebPEncoderOptions* opts = webp_encoder_create_default_options();
    opts->quality = 80;
    WebPEncoderCommand* encoder = webp_encoder_new_command(opts);
    webp_encoder_free_options(opts);

    while ((entry = readdir(dir)) != NULL) {
        if (strstr(entry->d_name, ".jpg") || strstr(entry->d_name, ".png")) {
            // Read input file
            char input_path[1024];
            snprintf(input_path, sizeof(input_path), "%s/%s", dir_path, entry->d_name);

            // ... read file ...

            NextImageBuffer output = {0};
            NextImageStatus status = webp_encoder_run_command(
                encoder, input_data, input_size, &output
            );

            if (status == NEXTIMAGE_STATUS_OK) {
                // Save output
                char output_path[1024];
                snprintf(output_path, sizeof(output_path), "%s/%s.webp",
                         dir_path, entry->d_name);

                FILE* fp = fopen(output_path, "wb");
                fwrite(output.data, 1, output.size, fp);
                fclose(fp);

                nextimage_free_buffer(&output);
                printf("✓ %s\n", entry->d_name);
            }
        }
    }

    closedir(dir);
    webp_encoder_free_command(encoder);
}
```

### Thread Safety

The library is thread-safe. You can create separate encoder/decoder instances for each thread:

```c
#include <pthread.h>

typedef struct {
    const char* input_file;
    const char* output_file;
} ConvertJob;

void* convert_thread(void* arg) {
    ConvertJob* job = (ConvertJob*)arg;

    // Each thread gets its own encoder
    WebPEncoderOptions* opts = webp_encoder_create_default_options();
    opts->quality = 80;
    WebPEncoderCommand* encoder = webp_encoder_new_command(opts);
    webp_encoder_free_options(opts);

    // ... convert file ...

    webp_encoder_free_command(encoder);
    return NULL;
}

void convert_parallel(ConvertJob* jobs, int count, int num_threads) {
    pthread_t* threads = malloc(sizeof(pthread_t) * num_threads);

    for (int i = 0; i < count; i++) {
        pthread_create(&threads[i % num_threads], NULL, convert_thread, &jobs[i]);

        if ((i + 1) % num_threads == 0 || i == count - 1) {
            // Wait for batch to complete
            for (int j = 0; j <= (i % num_threads); j++) {
                pthread_join(threads[j], NULL);
            }
        }
    }

    free(threads);
}
```

## Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| macOS    | ARM64 (M1/M2/M3) | ✅ |
| macOS    | x64          | ✅ |
| Linux    | x64          | ✅ |
| Linux    | ARM64        | ✅ |
| Windows  | x64          | ✅ |

## Memory Management

**Important:** Always free buffers and commands when done:

1. **Buffers**: Call `nextimage_free_buffer()` on all output buffers
2. **Commands**: Call `*_free_command()` on all encoder/decoder instances
3. **Options**: Call `*_free_options()` after creating commands

```c
// Good pattern
WebPEncoderOptions* opts = webp_encoder_create_default_options();
WebPEncoderCommand* cmd = webp_encoder_new_command(opts);
webp_encoder_free_options(opts);  // Free immediately after use

NextImageBuffer output = {0};
webp_encoder_run_command(cmd, data, size, &output);

// Use output...

nextimage_free_buffer(&output);   // Free buffer
webp_encoder_free_command(cmd);   // Free command
```

## Testing

Build and run tests:

```bash
cd build
cmake .. -DBUILD_TESTING=ON
cmake --build .
ctest
```

## Performance Tips

1. **Reuse command instances** for batch processing
2. **Choose appropriate quality settings** for your use case
3. **Use threading** for parallel processing
4. **Profile your application** to find bottlenecks

## License

BSD-3-Clause

## Links

- GitHub: https://github.com/ideamans/libnextimage
- Issues: https://github.com/ideamans/libnextimage/issues

## Contributing

Contributions are welcome! Please see the main repository for contribution guidelines.
