# libnextimage - C ライブラリ

C言語による高性能WebP、AVIF、GIF画像処理ライブラリ。

## 機能

- **WebPエンコード・デコード**: 品質とエンコードオプションを完全にコントロールできる高速WebP画像処理
- **AVIFエンコード・デコード**: 品質と速度プリセットを備えた最新のAVIFフォーマットサポート
- **GIF変換**: アニメーションGIFを含むGIFとWebPフォーマット間の変換
- **依存関係ゼロ**（リンク時）: すべての依存関係が組み込まれた単一の共有ライブラリ
- **クロスプラットフォーム**: macOS (ARM64/x64)、Linux (x64/ARM64)、Windows (x64)をサポート
- **メモリセーフ**: 明示的なクリーンアップ関数による適切なリソース管理
- **スレッドセーフ**: 複数のスレッドから安全に使用可能

## ビルド

### 前提条件

- CMake 3.15+
- Cコンパイラ（GCC、Clang、またはMSVC）
- Git（依存関係のダウンロード用）

### ビルドコマンド

```bash
# ビルドディレクトリを作成
mkdir -p build
cd build

# 設定
cmake ..

# ビルド
cmake --build .

# インストール（オプション）
sudo cmake --install .
```

### ビルドオプション

```bash
# デバッグビルド
cmake -DCMAKE_BUILD_TYPE=Debug ..

# 最適化付きリリースビルド
cmake -DCMAKE_BUILD_TYPE=Release ..

# 静的ライブラリでビルド
cmake -DBUILD_SHARED_LIBS=OFF ..
```

## インストール

### システム全体へのインストール

```bash
cd build
sudo cmake --install .
```

以下がインストールされます：
- ヘッダーファイル: `/usr/local/include/nextimage/`
- ライブラリ: `/usr/local/lib/`
- CMake設定ファイル

### プロジェクトでの使用

#### CMakeを使用

```cmake
find_package(libnextimage REQUIRED)

add_executable(myapp main.c)
target_link_libraries(myapp PRIVATE libnextimage::libnextimage)
```

#### pkg-configを使用

```bash
gcc main.c $(pkg-config --cflags --libs libnextimage) -o myapp
```

#### 手動リンク

```bash
gcc main.c -I/usr/local/include -L/usr/local/lib -lnextimage -o myapp
```

## クイックスタート

### WebPエンコード

```c
#include <stdio.h>
#include <stdlib.h>
#include <nextimage/webp_encoder.h>

int main() {
    // 入力画像を読み込む
    FILE* fp = fopen("input.jpg", "rb");
    fseek(fp, 0, SEEK_END);
    size_t size = ftell(fp);
    fseek(fp, 0, SEEK_SET);

    uint8_t* input_data = malloc(size);
    fread(input_data, 1, size, fp);
    fclose(fp);

    // オプション付きでエンコーダーを作成
    WebPEncoderOptions* opts = webp_encoder_create_default_options();
    opts->quality = 80;
    opts->method = 6;

    WebPEncoderCommand* encoder = webp_encoder_new_command(opts);
    webp_encoder_free_options(opts);

    if (!encoder) {
        fprintf(stderr, "エンコーダーの作成に失敗\n");
        return 1;
    }

    // WebPにエンコード
    NextImageBuffer output = {0};
    NextImageStatus status = webp_encoder_run_command(
        encoder, input_data, size, &output
    );

    if (status != NEXTIMAGE_STATUS_OK) {
        fprintf(stderr, "エンコード失敗: %s\n", nextimage_last_error_message());
        webp_encoder_free_command(encoder);
        return 1;
    }

    // 出力を保存
    fp = fopen("output.webp", "wb");
    fwrite(output.data, 1, output.size, fp);
    fclose(fp);

    printf("変換完了: %zu バイト → %zu バイト\n", size, output.size);

    // クリーンアップ
    nextimage_free_buffer(&output);
    webp_encoder_free_command(encoder);
    free(input_data);

    return 0;
}
```

### WebPデコード

```c
#include <stdio.h>
#include <stdlib.h>
#include <nextimage/webp_decoder.h>

int main() {
    // WebPファイルを読み込む
    FILE* fp = fopen("image.webp", "rb");
    fseek(fp, 0, SEEK_END);
    size_t size = ftell(fp);
    fseek(fp, 0, SEEK_SET);

    uint8_t* webp_data = malloc(size);
    fread(webp_data, 1, size, fp);
    fclose(fp);

    // デコーダーを作成
    WebPDecoderOptions* opts = webp_decoder_create_default_options();
    WebPDecoderCommand* decoder = webp_decoder_new_command(opts);
    webp_decoder_free_options(opts);

    // RGBAにデコード
    NextImageBuffer output = {0};
    int width, height;
    NextImageStatus status = webp_decoder_run_command(
        decoder, webp_data, size, &output, &width, &height
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        printf("幅: %d、高さ: %d\n", width, height);
        printf("データサイズ: %zu バイト\n", output.size);
    }

    // クリーンアップ
    nextimage_free_buffer(&output);
    webp_decoder_free_command(decoder);
    free(webp_data);

    return 0;
}
```

### AVIFエンコード

```c
#include <nextimage/avif_encoder.h>

int main() {
    // 入力を読み込む（WebPの例と同様）
    uint8_t* input_data;
    size_t size;
    // ... ファイル読み込み ...

    // オプション付きでエンコーダーを作成
    AVIFEncoderOptions* opts = avif_encoder_create_default_options();
    opts->quality = 60;
    opts->speed = 6;

    AVIFEncoderCommand* encoder = avif_encoder_new_command(opts);
    avif_encoder_free_options(opts);

    // AVIFにエンコード
    NextImageBuffer output = {0};
    NextImageStatus status = avif_encoder_run_command(
        encoder, input_data, size, &output
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        // 出力を保存
        FILE* fp = fopen("output.avif", "wb");
        fwrite(output.data, 1, output.size, fp);
        fclose(fp);
    }

    // クリーンアップ
    nextimage_free_buffer(&output);
    avif_encoder_free_command(encoder);

    return 0;
}
```

### AVIFデコード

```c
#include <nextimage/avif_decoder.h>

int main() {
    // AVIFファイルを読み込む（前の例と同様）
    uint8_t* avif_data;
    size_t size;
    // ... ファイル読み込み ...

    // デコーダーを作成
    AVIFDecoderOptions* opts = avif_decoder_create_default_options();
    AVIFDecoderCommand* decoder = avif_decoder_new_command(opts);
    avif_decoder_free_options(opts);

    // RGBAにデコード
    NextImageBuffer output = {0};
    int width, height;
    NextImageStatus status = avif_decoder_run_command(
        decoder, avif_data, size, &output, &width, &height
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        printf("%dx%d AVIF画像をデコード完了\n", width, height);
    }

    // クリーンアップ
    nextimage_free_buffer(&output);
    avif_decoder_free_command(decoder);

    return 0;
}
```

### GIFからWebPへの変換

```c
#include <nextimage/gif2webp.h>

int main() {
    // GIFファイルを読み込む
    uint8_t* gif_data;
    size_t size;
    // ... ファイル読み込み ...

    // オプション付きでコンバーターを作成
    Gif2WebPOptions* opts = gif2webp_create_default_options();
    opts->quality = 80;
    opts->method = 6;

    Gif2WebPCommand* cmd = gif2webp_new_command(opts);
    gif2webp_free_options(opts);

    // GIFをWebPに変換（アニメーションを保持）
    NextImageBuffer output = {0};
    NextImageStatus status = gif2webp_run_command(
        cmd, gif_data, size, &output
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        // 出力を保存
        FILE* fp = fopen("animated.webp", "wb");
        fwrite(output.data, 1, output.size, fp);
        fclose(fp);

        float compression = (1.0f - (float)output.size / size) * 100;
        printf("圧縮率: %.1f%%\n", compression);
    }

    // クリーンアップ
    nextimage_free_buffer(&output);
    gif2webp_free_command(cmd);

    return 0;
}
```

### WebPからGIFへの変換

```c
#include <nextimage/webp2gif.h>

int main() {
    // WebPファイルを読み込む
    uint8_t* webp_data;
    size_t size;
    // ... ファイル読み込み ...

    // コンバーターを作成
    WebP2GifOptions* opts = webp2gif_create_default_options();
    WebP2GifCommand* cmd = webp2gif_new_command(opts);
    webp2gif_free_options(opts);

    // WebPをGIFに変換
    NextImageBuffer output = {0};
    NextImageStatus status = webp2gif_run_command(
        cmd, webp_data, size, &output
    );

    if (status == NEXTIMAGE_STATUS_OK) {
        // 出力を保存
        FILE* fp = fopen("output.gif", "wb");
        fwrite(output.data, 1, output.size, fp);
        fclose(fp);
    }

    // クリーンアップ
    nextimage_free_buffer(&output);
    webp2gif_free_command(cmd);

    return 0;
}
```

## APIリファレンス

### 共通型

```c
// ステータスコード
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

// 出力データ用バッファ
typedef struct {
    uint8_t* data;
    size_t size;
} NextImageBuffer;

// ライブラリによって割り当てられたバッファを解放
void nextimage_free_buffer(NextImageBuffer* buffer);

// エラーハンドリング
const char* nextimage_last_error_message(void);
void nextimage_clear_error(void);
```

### WebPエンコーダー

```c
// ヘッダー: nextimage/webp_encoder.h

typedef struct {
    float quality;         // 0-100、デフォルト: 75
    int lossless;          // 0 または 1
    int method;            // 0-6、デフォルト: 4
    int preset;            // 0=default、1=picture、2=photo、3=drawing、4=icon、5=text
    int image_hint;        // 0=default、1=picture、2=photo、3=graph

    int target_size;       // ターゲットファイルサイズ（バイト）
    float target_psnr;     // ターゲットPSNR（dB）
    int segments;          // 1-4
    int sns_strength;      // 0-100
    int filter_strength;   // 0-100
    int filter_sharpness;  // 0-7
    int filter_type;       // 0=simple、1=strong
    int autofilter;        // 0 または 1

    int alpha_compression; // 0 または 1
    int alpha_filtering;   // 0=none、1=fast、2=best
    int alpha_quality;     // 0-100

    int pass;              // 1-10
    int show_compressed;   // 0 または 1
    int preprocessing;     // 0-7
    int partitions;        // 0-3
    int partition_limit;   // 0-100

    int emulate_jpeg_size; // 0 または 1
    int thread_level;      // 0 または 1
    int low_memory;        // 0 または 1
    int near_lossless;     // 0-100
    int exact;             // 0 または 1
    int use_delta_palette; // 0 または 1
    int use_sharp_yuv;     // 0 または 1

    int qmin;              // 0-100
    int qmax;              // 0-100

    // ... さらに多くのオプション
} WebPEncoderOptions;

// デフォルトオプションを作成
WebPEncoderOptions* webp_encoder_create_default_options(void);

// オプションを解放
void webp_encoder_free_options(WebPEncoderOptions* options);

// エンコーダーコマンドを作成
WebPEncoderCommand* webp_encoder_new_command(const WebPEncoderOptions* options);

// エンコードを実行
NextImageStatus webp_encoder_run_command(
    WebPEncoderCommand* cmd,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output
);

// エンコーダーを解放
void webp_encoder_free_command(WebPEncoderCommand* cmd);
```

### WebPデコーダー

```c
// ヘッダー: nextimage/webp_decoder.h

typedef struct {
    int format;            // ピクセルフォーマット（RGBA、BGRA、RGB、BGR）
    int use_threads;       // 0 または 1
    int bypass_filtering;  // 0 または 1
    int no_fancy_upsampling; // 0 または 1

    int crop_x;
    int crop_y;
    int crop_width;
    int crop_height;

    int scale_width;
    int scale_height;
} WebPDecoderOptions;

// エンコーダーと同様の関数
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

### AVIFエンコーダー

```c
// ヘッダー: nextimage/avif_encoder.h

typedef struct {
    int quality;           // 0-100、デフォルト: 60
    int quality_alpha;     // 0-100、デフォルト: -1
    int speed;             // 0-10、デフォルト: 6

    int bit_depth;         // 8、10、または12
    int yuv_format;        // YUV444、YUV422、YUV420、YUV400

    int lossless;          // 0 または 1
    int sharp_yuv;         // 0 または 1
    int target_size;       // ターゲットファイルサイズ

    int jobs;              // -1=全コア、0=自動、>0=スレッド数

    int tile_rows_log2;    // 0-6
    int tile_cols_log2;    // 0-6
    int auto_tiling;       // 0 または 1

    // ... さらに多くのオプション
} AVIFEncoderOptions;

// WebPエンコーダーと同様のAPI
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

### AVIFデコーダー

```c
// ヘッダー: nextimage/avif_decoder.h

typedef struct {
    int format;                  // ピクセルフォーマット
    int jobs;                    // スレッド数
    int chroma_upsampling;       // アップサンプリングモード

    int ignore_exif;             // 0 または 1
    int ignore_xmp;              // 0 または 1
    int ignore_icc;              // 0 または 1

    int image_size_limit;        // 最大画像サイズ
    int image_dimension_limit;   // 最大寸法
} AVIFDecoderOptions;

// WebPデコーダーと同様のAPI
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
// ヘッダー: nextimage/gif2webp.h

// WebPEncoderOptionsと同じオプション構造を使用
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
// ヘッダー: nextimage/webp2gif.h

typedef struct {
    int reserved;  // 将来の使用のため予約
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

## 高度な使い方

### エラーハンドリング

```c
NextImageStatus status = webp_encoder_run_command(encoder, data, size, &output);

if (status != NEXTIMAGE_STATUS_OK) {
    const char* error_msg = nextimage_last_error_message();

    switch (status) {
        case NEXTIMAGE_STATUS_ERROR_INVALID_PARAMETER:
            fprintf(stderr, "無効なパラメータ: %s\n", error_msg);
            break;
        case NEXTIMAGE_STATUS_ERROR_OUT_OF_MEMORY:
            fprintf(stderr, "メモリ不足: %s\n", error_msg);
            break;
        case NEXTIMAGE_STATUS_ERROR_ENCODE_FAILED:
            fprintf(stderr, "エンコード失敗: %s\n", error_msg);
            break;
        default:
            fprintf(stderr, "エラー: %s\n", error_msg);
    }

    nextimage_clear_error();
}
```

### バッチ処理

```c
#include <dirent.h>
#include <string.h>

void convert_directory(const char* dir_path) {
    DIR* dir = opendir(dir_path);
    struct dirent* entry;

    // エンコーダーを一度作成し、すべてのファイルで再利用
    WebPEncoderOptions* opts = webp_encoder_create_default_options();
    opts->quality = 80;
    WebPEncoderCommand* encoder = webp_encoder_new_command(opts);
    webp_encoder_free_options(opts);

    while ((entry = readdir(dir)) != NULL) {
        if (strstr(entry->d_name, ".jpg") || strstr(entry->d_name, ".png")) {
            // 入力ファイルを読み込む
            char input_path[1024];
            snprintf(input_path, sizeof(input_path), "%s/%s", dir_path, entry->d_name);

            // ... ファイル読み込み ...

            NextImageBuffer output = {0};
            NextImageStatus status = webp_encoder_run_command(
                encoder, input_data, input_size, &output
            );

            if (status == NEXTIMAGE_STATUS_OK) {
                // 出力を保存
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

### スレッドセーフ

ライブラリはスレッドセーフです。各スレッドに個別のエンコーダー/デコーダーインスタンスを作成できます：

```c
#include <pthread.h>

typedef struct {
    const char* input_file;
    const char* output_file;
} ConvertJob;

void* convert_thread(void* arg) {
    ConvertJob* job = (ConvertJob*)arg;

    // 各スレッドが独自のエンコーダーを取得
    WebPEncoderOptions* opts = webp_encoder_create_default_options();
    opts->quality = 80;
    WebPEncoderCommand* encoder = webp_encoder_new_command(opts);
    webp_encoder_free_options(opts);

    // ... ファイル変換 ...

    webp_encoder_free_command(encoder);
    return NULL;
}

void convert_parallel(ConvertJob* jobs, int count, int num_threads) {
    pthread_t* threads = malloc(sizeof(pthread_t) * num_threads);

    for (int i = 0; i < count; i++) {
        pthread_create(&threads[i % num_threads], NULL, convert_thread, &jobs[i]);

        if ((i + 1) % num_threads == 0 || i == count - 1) {
            // バッチの完了を待つ
            for (int j = 0; j <= (i % num_threads); j++) {
                pthread_join(threads[j], NULL);
            }
        }
    }

    free(threads);
}
```

## プラットフォームサポート

| プラットフォーム | アーキテクチャ | 状態 |
|-----------------|---------------|------|
| macOS           | ARM64 (M1/M2/M3) | ✅ |
| macOS           | x64           | ✅ |
| Linux           | x64           | ✅ |
| Linux           | ARM64         | ✅ |
| Windows         | x64           | ✅ |

## メモリ管理

**重要:** 使用後は必ずバッファとコマンドを解放してください：

1. **バッファ**: すべての出力バッファに対して `nextimage_free_buffer()` を呼び出す
2. **コマンド**: すべてのエンコーダー/デコーダーインスタンスに対して `*_free_command()` を呼び出す
3. **オプション**: コマンド作成後に `*_free_options()` を呼び出す

```c
// 良いパターン
WebPEncoderOptions* opts = webp_encoder_create_default_options();
WebPEncoderCommand* cmd = webp_encoder_new_command(opts);
webp_encoder_free_options(opts);  // 使用後すぐに解放

NextImageBuffer output = {0};
webp_encoder_run_command(cmd, data, size, &output);

// 出力を使用...

nextimage_free_buffer(&output);   // バッファを解放
webp_encoder_free_command(cmd);   // コマンドを解放
```

## テスト

テストをビルドして実行：

```bash
cd build
cmake .. -DBUILD_TESTING=ON
cmake --build .
ctest
```

## パフォーマンスのヒント

1. **コマンドインスタンスを再利用する** バッチ処理用
2. **適切な品質設定を選択する** ユースケースに応じて
3. **スレッディングを使用する** 並列処理用
4. **アプリケーションをプロファイルする** ボトルネックを見つけるため

## ライセンス

BSD-3-Clause

## リンク

- GitHub: https://github.com/ideamans/libnextimage
- イシュー: https://github.com/ideamans/libnextimage/issues

## コントリビューション

コントリビューションを歓迎します！コントリビューションガイドラインについては、メインリポジトリを参照してください。
