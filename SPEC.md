# このプロジェクトは

libwebp, libavifに含まれるエンコード・デコードコマンドである、

- cwebp / dwebp / gif2webp
- avifenc / avifdec

これらをFFIおよびGoのラッパーとして利用できるようにするものです。

また、アニメーションwebpからアニメーションGIFへの変換を行う`webp2gif`コマンドも新設します。

通常、これらのコマンドはバイナリファイルとして同梱し、プロセスを生成して実行しますが、このプロジェクトではそれらのコマンドを直接呼び出すことで、プロセス生成のオーバーヘッドを削減し、パフォーマンスの向上を図ります。

# ライセンス

このプロジェクトはMITライセンスのもとで公開されています。

# プロジェクト構造

```
libnextimage/
├── deps/                      # 依存ライブラリ (git submodules)
│   ├── libwebp/              # WebPエンコーダー/デコーダー
│   └── libavif/              # AVIFエンコーダー/デコーダー
├── c/                        # C言語FFIレイヤー
│   ├── include/              # ヘッダーファイル
│   │   ├── nextimage.h      # 共通インターフェース定義
│   │   ├── webp.h           # WebP関連のFFI
│   │   └── avif.h           # AVIF関連のFFI
│   ├── src/                  # C実装
│   │   ├── webp.c           # WebPエンコード/デコード実装
│   │   ├── avif.c           # AVIFエンコード/デコード実装
│   │   └── webp2gif.c       # WebP→GIF変換実装
│   ├── test/                 # 最小限のテストプログラム
│   │   └── basic_test.c     # 基本動作確認用テスト
│   ├── CMakeLists.txt        # CMakeビルド設定
│   └── Makefile              # 簡易ビルド用Makefile
├── golang/                   # Go言語バインディング
│   ├── cwebp.go             # cwebpラッパー
│   ├── dwebp.go             # dwebpラッパー
│   ├── gif2webp.go          # gif2webpラッパー
│   ├── avifenc.go           # avifencラッパー
│   ├── avifdec.go           # avifdecラッパー
│   ├── webp2gif.go          # webp2gifラッパー
│   ├── options.go           # オプション構造体定義
│   ├── common.go            # 共通ユーティリティ
│   └── *_test.go            # 各機能のテスト
├── lib/                      # プリコンパイル済み静的ライブラリ
│   ├── darwin-arm64/        # macOS Apple Silicon
│   ├── darwin-amd64/        # macOS Intel
│   ├── linux-amd64/         # Linux x64
│   ├── linux-arm64/         # Linux ARM64
│   ├── windows-amd64/       # Windows x64
│   └── other/               # その他の環境（ソースビルド必須）
├── scripts/                  # ビルド・テストスクリプト
│   ├── build.sh             # ビルドスクリプト
│   ├── build-all.sh         # 全プラットフォームビルド
│   ├── test.sh              # テスト実行
│   └── download-libs.sh     # プリコンパイル済みライブラリダウンロード
├── testdata/                 # テスト用画像ファイル
│   ├── jpeg/
│   ├── png/
│   ├── gif/
│   ├── webp/
│   └── avif/
├── .github/
│   └── workflows/
│       ├── build.yml        # ビルドワークフロー
│       └── release.yml      # リリースワークフロー
├── CLAUDE.md                 # このファイル
├── README.md                 # プロジェクト説明
└── LICENSE                   # MITライセンス
```

# C言語FFIインターフェース

## 設計方針

- メモリ管理は明示的に分離（`*_alloc`関数は自動割り当て、`*_into`関数は呼び出し側バッファ使用）
- エラーハンドリングは戻り値、エラーコード、および詳細なエラーメッセージで行う
- スレッドセーフな実装（スレッドローカルストレージでエラーメッセージを管理）
- デコード時のピクセルフォーマットを明示的に指定・取得
- AVIFのマルチプレーンデータに対応

## 基本インターフェース

```c
// nextimage.h - 共通定義

typedef enum {
    NEXTIMAGE_OK = 0,
    NEXTIMAGE_ERROR_INVALID_PARAM = -1,
    NEXTIMAGE_ERROR_ENCODE_FAILED = -2,
    NEXTIMAGE_ERROR_DECODE_FAILED = -3,
    NEXTIMAGE_ERROR_OUT_OF_MEMORY = -4,
    NEXTIMAGE_ERROR_UNSUPPORTED = -5,
    NEXTIMAGE_ERROR_BUFFER_TOO_SMALL = -6,
} NextImageStatus;

// ピクセルフォーマット定義
typedef enum {
    NEXTIMAGE_FORMAT_RGBA = 0,      // RGBA 8bit/channel
    NEXTIMAGE_FORMAT_RGB = 1,       // RGB 8bit/channel
    NEXTIMAGE_FORMAT_BGRA = 2,      // BGRA 8bit/channel
    NEXTIMAGE_FORMAT_YUV420 = 3,    // YUV 4:2:0 planar
    NEXTIMAGE_FORMAT_YUV422 = 4,    // YUV 4:2:2 planar
    NEXTIMAGE_FORMAT_YUV444 = 5,    // YUV 4:4:4 planar
} NextImagePixelFormat;

// エンコード用バッファ（常にライブラリが割り当て）
typedef struct {
    uint8_t* data;
    size_t size;
} NextImageEncodeBuffer;

// デコード用バッファ情報（プレーン別の詳細情報を含む）
typedef struct {
    // プライマリプレーン（インターリーブ形式の場合は全データ、planarの場合はYプレーン）
    uint8_t* data;
    size_t data_capacity;       // dataバッファの容量（*_into関数用、バイト単位）
    size_t data_size;           // 実際のデータサイズ（バイト単位）
    size_t stride;              // Y/プライマリプレーンの行ごとのバイト数

    // Uプレーン（YUV planarの場合のみ使用）
    uint8_t* u_plane;
    size_t u_capacity;          // Uプレーンバッファの容量（*_into関数用）
    size_t u_size;              // Uプレーンの実際のサイズ
    size_t u_stride;            // Uプレーンの行ごとのバイト数

    // Vプレーン（YUV planarの場合のみ使用）
    uint8_t* v_plane;
    size_t v_capacity;          // Vプレーンバッファの容量（*_into関数用）
    size_t v_size;              // Vプレーンの実際のサイズ
    size_t v_stride;            // Vプレーンの行ごとのバイト数

    // メタデータ
    int width;                  // 画像幅（ピクセル単位）
    int height;                 // 画像高さ（ピクセル単位）
    int bit_depth;              // ビット深度（8, 10, 12）
    NextImagePixelFormat format; // ピクセルフォーマット
    int owns_data;              // 1ならライブラリがメモリを所有
} NextImageDecodeBuffer;

// メモリ解放（owns_data == 1の場合のみ解放される）
void nextimage_free_encode_buffer(NextImageEncodeBuffer* buffer);
void nextimage_free_decode_buffer(NextImageDecodeBuffer* buffer);

// エラーメッセージ取得
// - スレッドローカルストレージに保存された最後のエラーメッセージを返す
// - 返される文字列は次のFFI呼び出しまで有効（コピー不要だがスレッドローカル）
// - 成功した呼び出しでは自動的にクリアされない（明示的なクリアが必要）
// - NULLが返された場合はエラーメッセージが設定されていない
const char* nextimage_last_error_message(void);

// エラーメッセージのクリア
// - 次のエラーまでnextimage_last_error_message()がNULLを返すようにする
void nextimage_clear_error(void);
```

## WebP FFI

```c
// webp.h

typedef struct {
    float quality;           // 0-100, default 75
    int lossless;           // 0 or 1, default 0
    int method;             // 0-6, default 4
    int target_size;        // target size in bytes
    float target_psnr;      // target PSNR
    int exact;              // preserve RGB values in transparent area
    // ... その他のオプション
} NextImageWebPEncodeOptions;

typedef struct {
    int use_threads;            // 0 or 1
    int bypass_filtering;       // 0 or 1
    int no_fancy_upsampling;    // 0 or 1
    NextImagePixelFormat format; // 希望するピクセルフォーマット（デフォルト: RGBA）
    // ... その他のオプション
} NextImageWebPDecodeOptions;

// エンコード（ライブラリがメモリを割り当て）
NextImageStatus nextimage_webp_encode_alloc(
    const uint8_t* input_data,
    size_t input_size,
    int width,
    int height,
    NextImagePixelFormat input_format,
    const NextImageWebPEncodeOptions* options,
    NextImageEncodeBuffer* output
);

// デコード（ライブラリがメモリを割り当て）
NextImageStatus nextimage_webp_decode_alloc(
    const uint8_t* webp_data,
    size_t webp_size,
    const NextImageWebPDecodeOptions* options,
    NextImageDecodeBuffer* output
);

// デコード（呼び出し側が用意したバッファを使用）
// buffer->capacity, buffer->data を事前に設定すること
// 必要なバッファサイズは nextimage_webp_decode_size() で取得可能
NextImageStatus nextimage_webp_decode_into(
    const uint8_t* webp_data,
    size_t webp_size,
    const NextImageWebPDecodeOptions* options,
    NextImageDecodeBuffer* buffer
);

// デコードに必要なバッファサイズを事前に計算
NextImageStatus nextimage_webp_decode_size(
    const uint8_t* webp_data,
    size_t webp_size,
    int* width,
    int* height,
    size_t* required_size
);

// GIF to WebP（ライブラリがメモリを割り当て）
NextImageStatus nextimage_gif2webp_alloc(
    const uint8_t* gif_data,
    size_t gif_size,
    const NextImageWebPEncodeOptions* options,
    NextImageEncodeBuffer* output
);

// WebP to GIF（新機能、ライブラリがメモリを割り当て）
NextImageStatus nextimage_webp2gif_alloc(
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageEncodeBuffer* output
);
```

## AVIF FFI

```c
// avif.h

typedef struct {
    int quality;            // 0-100, default 50
    int speed;              // 0-10, default 6
    int min_quantizer;      // 0-63
    int max_quantizer;      // 0-63
    int enable_alpha;       // 0 or 1
    int bit_depth;          // 8, 10, or 12 (default: 8)
    // ... その他のオプション
} NextImageAVIFEncodeOptions;

typedef struct {
    int use_threads;            // 0 or 1
    NextImagePixelFormat format; // 希望するピクセルフォーマット（デフォルト: RGBA）
    // ... その他のオプション
} NextImageAVIFDecodeOptions;

// エンコード（ライブラリがメモリを割り当て）
NextImageStatus nextimage_avif_encode_alloc(
    const uint8_t* input_data,
    size_t input_size,
    int width,
    int height,
    NextImagePixelFormat input_format,
    const NextImageAVIFEncodeOptions* options,
    NextImageEncodeBuffer* output
);

// デコード（ライブラリがメモリを割り当て）
NextImageStatus nextimage_avif_decode_alloc(
    const uint8_t* avif_data,
    size_t avif_size,
    const NextImageAVIFDecodeOptions* options,
    NextImageDecodeBuffer* output
);

// デコード（呼び出し側が用意したバッファを使用）
NextImageStatus nextimage_avif_decode_into(
    const uint8_t* avif_data,
    size_t avif_size,
    const NextImageAVIFDecodeOptions* options,
    NextImageDecodeBuffer* buffer
);

// デコードに必要なバッファサイズを事前に計算
NextImageStatus nextimage_avif_decode_size(
    const uint8_t* avif_data,
    size_t avif_size,
    int* width,
    int* height,
    int* bit_depth,
    size_t* required_size
);
```

# Go言語インターフェース

## インストール

```bash
go get github.com/ideamans/libnextimage/golang
```

## 設計方針

- 明示的な関数名による型安全なインターフェース（`EncodeBytes`, `EncodeFile`, `EncodeStream`など）
- 入出力は `[]byte`, `io.Reader/io.Writer`, `string`(ファイルパス) に対応
- オプションはCLI互換の構造体で提供
- エラーハンドリングはGoの標準的な方法に従う
- 詳細なエラーメッセージをGoのerror型にラップして提供

## 使用例

```go
package main

import (
    "os"
    "github.com/ideamans/libnextimage/golang"
)

func main() {
    // 例1: WebP エンコード - ファイル同士
    err := libnextimage.WebPEncodeFile("input.jpg", "output.webp",
        libnextimage.WebPEncodeOptions{
            Quality: 80,
            Method: 4,
        })
    if err != nil {
        panic(err)
    }

    // 例2: WebP デコード - ファイル同士
    err = libnextimage.WebPDecodeFile("input.webp", "output.png",
        libnextimage.WebPDecodeOptions{
            UseThreads: true,
            Format: libnextimage.FormatRGBA,
        })
    if err != nil {
        panic(err)
    }

    // 例3: AVIF エンコード - ストリーム同士
    inFile, _ := os.Open("input.jpg")
    defer inFile.Close()

    outFile, _ := os.Create("output.avif")
    defer outFile.Close()

    err = libnextimage.AVIFEncodeStream(inFile, outFile,
        libnextimage.AVIFEncodeOptions{
            Quality: 50,
            Speed: 6,
        })
    if err != nil {
        panic(err)
    }

    // 例4: AVIF エンコード - バイト列同士
    imageData := make([]byte, width*height*4) // RGBA data
    // ... imageDataを準備 ...

    avifBytes, err := libnextimage.AVIFEncodeBytes(
        imageData, width, height,
        libnextimage.FormatRGBA,
        libnextimage.AVIFEncodeOptions{Quality: 75})
    if err != nil {
        panic(err)
    }

    // 例5: AVIF デコード - バイト列から構造体へ
    decoded, err := libnextimage.AVIFDecodeBytes(avifBytes,
        libnextimage.AVIFDecodeOptions{Format: libnextimage.FormatRGBA})
    if err != nil {
        panic(err)
    }
    // decoded.Data, decoded.Width, decoded.Height などを使用
}
```

## API設計

シンプルさと明確さを重視し、入出力の組み合わせを3つのパターンに限定:
1. **バイト配列同士** (`*Bytes`): メモリ上のデータを直接変換
2. **ファイル同士** (`*File`): ファイルパスを指定して変換
3. **ストリーム同士** (`*Stream`): io.Reader/io.Writerで変換

各フォーマットに対して以下の関数を提供:

### WebP

```go
// エンコード - []byte入力 → []byte出力
func WebPEncodeBytes(data []byte, width, height int, format PixelFormat, opts WebPEncodeOptions) ([]byte, error)

// エンコード - ファイル入力 → ファイル出力
func WebPEncodeFile(inputPath string, outputPath string, opts WebPEncodeOptions) error

// エンコード - ストリーム入力 → ストリーム出力
func WebPEncodeStream(input io.Reader, output io.Writer, opts WebPEncodeOptions) error

// デコード - []byte入力 → DecodedImage出力
func WebPDecodeBytes(webpData []byte, opts WebPDecodeOptions) (*DecodedImage, error)

// デコード - ファイル入力 → ファイル出力
func WebPDecodeFile(inputPath string, outputPath string, opts WebPDecodeOptions) error

// デコード - ストリーム入力 → ストリーム出力
func WebPDecodeStream(input io.Reader, output io.Writer, opts WebPDecodeOptions) error

// GIF to WebP - []byte入力 → []byte出力
func GIF2WebPBytes(gifData []byte, opts WebPEncodeOptions) ([]byte, error)

// GIF to WebP - ファイル入力 → ファイル出力
func GIF2WebPFile(inputPath string, outputPath string, opts WebPEncodeOptions) error

// WebP to GIF - []byte入力 → []byte出力
func WebP2GIFBytes(webpData []byte) ([]byte, error)

// WebP to GIF - ファイル入力 → ファイル出力
func WebP2GIFFile(inputPath string, outputPath string) error
```

### AVIF

```go
// エンコード - []byte入力 → []byte出力
func AVIFEncodeBytes(data []byte, width, height int, format PixelFormat, opts AVIFEncodeOptions) ([]byte, error)

// エンコード - ファイル入力 → ファイル出力
func AVIFEncodeFile(inputPath string, outputPath string, opts AVIFEncodeOptions) error

// エンコード - ストリーム入力 → ストリーム出力
func AVIFEncodeStream(input io.Reader, output io.Writer, opts AVIFEncodeOptions) error

// デコード - []byte入力 → DecodedImage出力
func AVIFDecodeBytes(avifData []byte, opts AVIFDecodeOptions) (*DecodedImage, error)

// デコード - ファイル入力 → ファイル出力
func AVIFDecodeFile(inputPath string, outputPath string, opts AVIFDecodeOptions) error

// デコード - ストリーム入力 → ストリーム出力
func AVIFDecodeStream(input io.Reader, output io.Writer, opts AVIFDecodeOptions) error
```

### API設計の原則

1. **入出力の一貫性**: 入力と出力の型を統一（Bytes同士、File同士、Stream同士）
2. **関数名の明確さ**: 関数名のサフィックスで入出力の型が分かる
3. **ジェネリクス不使用**: 型安全性とシンプルさを両立
4. **エラーハンドリング**: すべての関数がerrorを返す
5. **オプション構造体**: 各フォーマット固有のオプションを型安全に扱う

### 共通型定義

```go
type PixelFormat int

const (
    FormatRGBA PixelFormat = iota
    FormatRGB
    FormatBGRA
    FormatYUV420
    FormatYUV422
    FormatYUV444
)

type DecodedImage struct {
    // プライマリプレーン（インターリーブ形式の場合は全データ、planarの場合はYプレーン）
    Data   []byte
    Stride int

    // UVプレーン（YUV planarの場合のみ）
    UPlane   []byte  // nil if not planar
    UStride  int
    VPlane   []byte  // nil if not planar
    VStride  int

    // メタデータ
    Width    int
    Height   int
    BitDepth int          // 8, 10, or 12
    Format   PixelFormat
}

// ヘルパーメソッド
func (img *DecodedImage) IsPlanar() bool {
    return img.UPlane != nil && img.VPlane != nil
}

func (img *DecodedImage) IsHighBitDepth() bool {
    return img.BitDepth > 8
}
```

## CGOビルドタグとリンク設定

**重要**: libnextimage.aは全ての依存ライブラリを含む完全なクロージャです。
libtoolまたはar MRIスクリプトにより、以下のライブラリが統合されています:
- libnextimage (本体)
- libwebp, libwebpdemux, libsharpyuv (WebP関連)
- libavif_internal (AVIF本体)
- libaom (AV1コーデック)

そのため、Goからは**libnextimage.aのみをリンク**すれば動作します。

```go
package libnextimage

/*
#cgo CFLAGS: -I${SRCDIR}/../c/include

// macOS ARM64: libnextimage.aのみ + システムライブラリ
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/../lib/darwin-arm64 -lnextimage
#cgo darwin,arm64 LDFLAGS: -framework CoreFoundation

// macOS Intel: libnextimage.aのみ + システムライブラリ
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/../lib/darwin-amd64 -lnextimage
#cgo darwin,amd64 LDFLAGS: -framework CoreFoundation

// Linux x64: libnextimage.aのみ + システムライブラリ
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../lib/linux-amd64 -lnextimage
#cgo linux,amd64 LDFLAGS: -lpthread -lm -ldl

// Linux ARM64: libnextimage.aのみ + システムライブラリ
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../lib/linux-arm64 -lnextimage
#cgo linux,arm64 LDFLAGS: -lpthread -lm -ldl

// Windows x64: libnextimage.aのみ + システムライブラリ
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/../lib/windows-amd64 -lnextimage
#cgo windows,amd64 LDFLAGS: -lws2_32 -lkernel32 -luser32

// その他のプラットフォーム: ソースからビルドされたライブラリを使用
#cgo !darwin,!linux,!windows LDFLAGS: -L${SRCDIR}/../lib/other -lnextimage
#cgo !darwin,!linux,!windows LDFLAGS: -lpthread -lm

#include "nextimage.h"
#include "webp.h"
#include "avif.h"
*/
import "C"
```

### ライブラリ統合の仕組み

**macOS/BSD**: `libtool -static`を使用
```bash
libtool -static -o libnextimage.a \
  libnextimage.a libwebp.a libwebpdemux.a libsharpyuv.a \
  libavif_internal.a libaom.a
```

**Linux**: `ar` MRIスクリプトを使用
```bash
cat > combine.mri <<EOF
CREATE libnextimage.a
ADDLIB libnextimage.a
ADDLIB libwebp.a
ADDLIB libwebpdemux.a
ADDLIB libsharpyuv.a
ADDLIB libavif_internal.a
ADDLIB libaom.a
SAVE
END
EOF
ar -M < combine.mri
ranlib libnextimage.a
```

この方式により、重複するオブジェクトファイル名（例: scale.c.o）の問題を回避し、
全てのシンボルが正しく含まれることを保証します。

## 依存関係のBOM（Bill of Materials）

各プラットフォーム用のプリコンパイルライブラリには、以下のすべての依存関係が含まれます：

### コア依存関係

- **libnextimage**: このプロジェクトのFFIレイヤー
- **libwebp**: WebPエンコーダー/デコーダー
  - libwebpdemux: WebP demuxer
  - libwebpmux: WebP muxer
  - libsharpyuv: YUV変換
- **libavif**: AVIFエンコーダー/デコーダー
  - libaom: AV1エンコーダー/デコーダー
  - libdav1d: 高速AV1デコーダー
  - libyuv: YUV/RGB変換

### システム依存関係

- **zlib**: 圧縮ライブラリ
- **pthread** (Unix系): スレッドサポート
- **libm** (Unix系): 数学ライブラリ
- **libdl** (Linux): 動的リンクサポート
- **ws2_32, kernel32, user32** (Windows): Windowsシステムライブラリ

### ライセンス情報

各リリースには以下のライセンスファイルが含まれます：

- `LICENSE` - プロジェクト自体のMITライセンス
- `LICENSES/` ディレクトリ:
  - `libwebp-LICENSE` - BSD-3-Clause
  - `libavif-LICENSE` - BSD-2-Clause
  - `libaom-LICENSE` - BSD-2-Clause with Patent Grant
  - `libdav1d-LICENSE` - BSD-2-Clause
  - `libyuv-LICENSE` - BSD-3-Clause
  - `zlib-LICENSE` - zlib License

### バージョン追跡

各リリースには`DEPENDENCIES.txt`ファイルを含め、使用している各ライブラリのバージョンを記録：

```
libnextimage: 1.0.0
libwebp: 1.3.2
libavif: 1.0.3
libaom: 3.8.0
libdav1d: 1.3.0
libyuv: 1862
zlib: 1.3
```

# ビルド方法

## 依存関係の初期化

```bash
# git submodulesの初期化
git submodule update --init --recursive
```

## C言語ライブラリのビルド

```bash
# 現在のプラットフォーム用にビルド
cd c
mkdir build && cd build
cmake ..
make

# ライブラリを適切な場所にコピー
# 例: macOS ARM64の場合
cp libnextimage.a ../../lib/darwin-arm64/
```

または簡易スクリプトを使用:

```bash
./scripts/build.sh
```

## Go言語パッケージの利用

プリコンパイル済みライブラリを使用する場合:

```bash
go get github.com/ideamans/libnextimage/golang
```

ソースからビルドする場合:

```bash
# C言語ライブラリをビルド後
cd golang
go build
go test
```

# 利用方法

利用時にはふたつの方法をサポートします。

## 1. プリコンパイル済みライブラリのダウンロード

CI/CDにより、v*タグのついた静的ライブラリをGitHub Actionsでビルドし、Releaseとして公開します。`go get`などをすると、自動的にそれらの静的ライブラリも一式ダウンロードされ、即座に利用できます。

対応プラットフォーム:
- darwin/arm64 (macOS Apple Silicon)
- darwin/amd64 (macOS Intel)
- linux/amd64 (Linux x64)
- linux/arm64 (Linux ARM64)
- windows/amd64 (Windows x64)

## 2. ソースコードのビルド

上記以外の環境や、最新のソースコードを使用したい場合は、ソースからビルドします。

```bash
git clone --recursive https://github.com/ideamans/libnextimage.git
cd libnextimage
./scripts/build.sh
```

ビルドされたライブラリは `lib/other/` ディレクトリに配置され、CGOがそれを参照します。

# テスト

## テスト方針

### C言語レイヤー: Sanitizerベースのテスト

C言語FFIレイヤーでは、Sanitizerを活用した徹底的なメモリ・動作検証を行います。

**テスト対象:**
- 各関数が正常にコンパイルできること
- 基本的なエンコード/デコードが動作すること
- エラーコードが正しく返されること
- メモリ解放関数が正常に動作すること
- バッファオーバーラン・アンダーランの検出
- Use-after-freeの検出
- 未定義動作の検出

**テストツール:**
- **AddressSanitizer (ASan)**: メモリエラー検出
  ```bash
  cmake -DCMAKE_C_FLAGS="-fsanitize=address -fno-omit-frame-pointer -g" ..
  ```
- **UndefinedBehaviorSanitizer (UBSan)**: 未定義動作検出
  ```bash
  cmake -DCMAKE_C_FLAGS="-fsanitize=undefined -g" ..
  ```
- **Valgrind**: 詳細なメモリリーク検出（CI用の軽量ハーネス）
  ```bash
  valgrind --leak-check=full --show-leak-kinds=all ./c_basic_test
  ```
- **手動リークカウンター**: FFI内部で割り当て/解放をカウント
  ```c
  int64_t nextimage_allocation_counter(void);  // デバッグビルドのみ
  ```

**実装場所:**
- `c/test/` ディレクトリに基本テストプログラム
- `c/test/sanitizer/` ディレクトリにSanitizer専用テスト
- CMakeのテストターゲットとして定義（通常ビルド、ASanビルド、UBSanビルド）

### Go言語レイヤー: 詳細なテスト

Go言語バインディングでは、包括的なテストを実施します。すべてのパターン、エラーケース、メモリ管理を網羅的にテストします。

**テスト対象:**
- 全ての入出力パターン (`[]byte`, `io.Reader/Writer`, `string`)
- 全てのオプションの組み合わせ
- エラーハンドリング
- 並行処理の安全性
- **メモリリークの検出（最重要）**
- パフォーマンス測定

## テストデータの準備

`testdata/` ディレクトリに以下のパターンの画像を用意:

**JPEG:**
- `baseline.jpg` - ベースラインJPEG
- `progressive.jpg` - プログレッシブJPEG
- `exif.jpg` - EXIF情報付き
- `large.jpg` - 大サイズ (4000x3000以上)
- `small.jpg` - 小サイズ (100x100以下)
- `grayscale.jpg` - グレースケール
- `corrupted-header.jpg` - 破損したヘッダー（エラーテスト用）
- `truncated.jpg` - 不完全なファイル（エラーテスト用）

**PNG:**
- `rgb.png` - RGB
- `rgba.png` - RGBA (透過あり)
- `grayscale.png` - グレースケール
- `indexed.png` - インデックスカラー
- `transparent.png` - 透過PNG
- `16bit.png` - 16ビット深度
- `corrupted.png` - 破損したPNG（エラーテスト用）

**GIF:**
- `static.gif` - 静止画GIF
- `animated.gif` - アニメーションGIF
- `transparent.gif` - 透過GIF
- `animated-transparent.gif` - 透過アニメーションGIF
- `corrupted.gif` - 破損したGIF（エラーテスト用）

**WebP:**
- `lossless.webp` - ロスレスWebP
- `lossy.webp` - ロッシーWebP
- `animated.webp` - アニメーションWebP
- `alpha.webp` - アルファチャンネル付き
- `animated-alpha.webp` - アルファ付きアニメーション
- `corrupted.webp` - 破損したWebP（エラーテスト用）
- `truncated.webp` - 不完全なWebP（エラーテスト用）

**AVIF:**
- `quality-high.avif` - 高品質
- `quality-low.avif` - 低品質
- `alpha.avif` - アルファチャンネル付き
- `yuv420.avif` - YUV 4:2:0
- `yuv422.avif` - YUV 4:2:2
- `yuv444.avif` - YUV 4:4:4
- `10bit.avif` - 10ビット深度
- `12bit.avif` - 12ビット深度
- `animated.avif` - アニメーションAVIF
- `animated-alpha.avif` - アルファ付きアニメーションAVIF
- `corrupted.avif` - 破損したAVIF（エラーテスト用）
- `truncated.avif` - 不完全なAVIF（エラーテスト用）

**エッジケース:**
- `empty.bin` - 空ファイル
- `random.bin` - ランダムデータ（画像ではない）
- `1x1.png` - 最小サイズ画像
- `10000x10000.png` - 超大型画像（メモリテスト用）

## Goテストの実行

### 基本テスト

```bash
cd golang
go test -v ./...
```

### メモリリークテスト（最重要）

複数の方法でメモリリークを検出します。CヒープとGoヒープの両方を監視します。

#### 1. C層リークカウンターテスト（最優先）

FFI内部のカウンターを使ってCヒープのリークを直接検出:

```go
func TestMemoryLeak_WebPEncode_CCounter(t *testing.T) {
    // デバッグビルドのみで有効
    if !libnextimage.IsDebugBuild() {
        t.Skip("Debug build required for leak counter test")
    }

    initialCount := libnextimage.AllocationCounter()

    // 100回エンコード/デコードを実行
    for i := 0; i < 100; i++ {
        data, err := libnextimage.WebPEncodeFile("testdata/jpeg/baseline.jpg",
            libnextimage.WebPEncodeOptions{Quality: 80})
        if err != nil {
            t.Fatal(err)
        }
        runtime.KeepAlive(data) // 最適化防止
    }

    // すべてのメモリが解放されているはず
    finalCount := libnextimage.AllocationCounter()
    if finalCount != initialCount {
        t.Errorf("C heap leak detected: %d allocations not freed", finalCount-initialCount)
    }
}
```

#### 2. Go標準のメモリプロファイリング（開発用）

```bash
cd golang
go test -v -memprofile=mem.prof ./...
go tool pprof -alloc_space mem.prof
```

#### 3. 軽量Valgrindテスト（CI用）

重いループではなく、軽量なハーネスでValgrindを使用:

```bash
# 専用の軽量テストバイナリをビルド
CGO_ENABLED=1 go test -c -tags=valgrind -o test.valgrind.bin ./...
# 10回程度の実行で検証
valgrind --leak-check=full --error-exitcode=1 ./test.valgrind.bin -test.run=ValgrindLeak
```

#### 4. 夜間ジョブでの長時間テスト

CI上では軽量テスト、夜間ジョブで重いテストを実施:

```yaml
# Nightly job only
- name: Long-running memory leak test
  if: github.event_name == 'schedule'
  run: go test -v -count=1000 -timeout=2h -run=TestMemoryLeak ./...
```

#### 5. Go+C混合メモリ監視テスト

```go
func TestMemoryLeak_Mixed(t *testing.T) {
    var m1, m2 runtime.MemStats
    c1 := libnextimage.AllocationCounter()

    runtime.GC()
    runtime.ReadMemStats(&m1)

    // 適度な回数（100回）で実行
    for i := 0; i < 100; i++ {
        data, _ := libnextimage.WebPEncodeBytes(testImageData, 640, 480,
            libnextimage.FormatRGBA, libnextimage.WebPEncodeOptions{Quality: 80})
        runtime.KeepAlive(data)
    }

    runtime.GC()
    runtime.ReadMemStats(&m2)
    c2 := libnextimage.AllocationCounter()

    // Cヒープチェック
    if c2 != c1 {
        t.Errorf("C heap leak: %d", c2-c1)
    }

    // Goヒープチェック（許容範囲: 5MB）
    allocDiff := m2.Alloc - m1.Alloc
    if allocDiff > 5*1024*1024 {
        t.Errorf("Possible Go heap leak: %d bytes", allocDiff)
    }
}
```

### カバレッジテスト

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### ベンチマークテスト

```bash
go test -v -bench=. -benchmem ./...
```

## Goテストの詳細内容

### 1. エンコード/デコード往復テスト

全ての画像フォーマット、全ての入出力パターンで往復変換をテスト:

```go
func TestRoundTrip_WebP_AllPatterns(t *testing.T) {
    patterns := []struct{
        name string
        input string
    }{
        {"baseline", "testdata/jpeg/baseline.jpg"},
        {"progressive", "testdata/jpeg/progressive.jpg"},
        // ... 全パターン
    }

    for _, p := range patterns {
        t.Run(p.name, func(t *testing.T) {
            // []byte入力 -> []byte出力
            testRoundTripBytes(t, p.input)
            // string入力 -> string出力
            testRoundTripFiles(t, p.input)
            // io.Reader入力 -> io.Writer出力
            testRoundTripStreams(t, p.input)
        })
    }
}
```

### 2. オプション網羅テスト

全てのオプションの組み合わせをテスト:

```go
func TestCWebP_AllOptions(t *testing.T) {
    qualities := []float32{10, 50, 75, 90, 100}
    methods := []int{0, 2, 4, 6}
    lossless := []int{0, 1}

    for _, q := range qualities {
        for _, m := range methods {
            for _, l := range lossless {
                // テスト実行
            }
        }
    }
}
```

### 3. エラーハンドリングテスト

異常系のテストを徹底的に実施:

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct{
        name string
        input []byte
        expectError bool
    }{
        {"empty data", []byte{}, true},
        {"invalid data", []byte{0x00, 0x01, 0x02}, true},
        {"truncated data", truncatedJPEG, true},
        {"corrupted header", corruptedJPEG, true},
        {"valid data", validJPEG, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := libnextimage.WebPEncodeBytes(tt.input, 640, 480,
                libnextimage.FormatRGBA, libnextimage.WebPEncodeOptions{})
            if (err != nil) != tt.expectError {
                t.Errorf("expected error: %v, got: %v", tt.expectError, err)
            }
        })
    }
}
```

### 4. 並行処理テスト

複数のgoroutineから同時にエンコード/デコードを実行:

```go
func TestConcurrency(t *testing.T) {
    const goroutines = 100
    const iterations = 10

    var wg sync.WaitGroup
    errors := make(chan error, goroutines*iterations)

    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < iterations; j++ {
                _, err := libnextimage.WebPEncodeBytes(testImageData, 640, 480,
                    libnextimage.FormatRGBA, libnextimage.WebPEncodeOptions{Quality: 80})
                if err != nil {
                    errors <- err
                }
            }
        }()
    }

    wg.Wait()
    close(errors)

    if len(errors) > 0 {
        t.Errorf("concurrent execution failed with %d errors", len(errors))
    }
}
```

### 5. メモリリークテスト（最重要）

上記の詳細なメモリリークテストを全ての関数に対して実施:

- エンコード/デコード繰り返し実行
- ファイナライザーの動作確認
- CGO境界でのメモリ管理確認
- エラー時のメモリ解放確認

### 6. パフォーマンステスト

プロセス生成版との速度比較:

```go
func BenchmarkWebP_Library_vs_Process(b *testing.B) {
    b.Run("Library", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            libnextimage.WebPEncodeBytes(testImageData, 640, 480,
                libnextimage.FormatRGBA, libnextimage.WebPEncodeOptions{Quality: 80})
        }
    })

    b.Run("Process", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            cmd := exec.Command("cwebp", "-q", "80", "-o", "/tmp/out.webp", "/tmp/in.jpg")
            cmd.Run()
        }
    })
}
```

### 7. CLIコマンド互換性テスト

実際のCLIコマンドとの出力比較:

```go
func TestCLICompatibility_WebP(t *testing.T) {
    // ライブラリでエンコード
    libOutput, err := libnextimage.WebPEncodeFile("testdata/jpeg/baseline.jpg",
        libnextimage.WebPEncodeOptions{Quality: 80})
    if err != nil {
        t.Fatal(err)
    }

    // CLIコマンドでエンコード
    cmd := exec.Command("cwebp", "-q", "80", "testdata/jpeg/baseline.jpg", "-o", "/tmp/cli.webp")
    cmd.Run()
    cliOutput, _ := os.ReadFile("/tmp/cli.webp")

    // バイト完全一致は難しいため、デコード後の画像を比較
    // または、ファイルサイズの差が5%以内であることを確認
    sizeDiff := math.Abs(float64(len(libOutput)-len(cliOutput))) / float64(len(cliOutput))
    if sizeDiff > 0.05 {
        t.Errorf("output size differs too much: %f%%", sizeDiff*100)
    }
}
```

## CI/CDでの自動テスト

GitHub Actionsで以下のテストを自動実行:

```yaml
# C層のSanitizerテスト
- name: Build and test with AddressSanitizer
  run: |
    cd c
    mkdir build-asan && cd build-asan
    cmake -DCMAKE_C_FLAGS="-fsanitize=address -fno-omit-frame-pointer -g" ..
    make
    ctest --output-on-failure

- name: Build and test with UndefinedBehaviorSanitizer
  run: |
    cd c
    mkdir build-ubsan && cd build-ubsan
    cmake -DCMAKE_C_FLAGS="-fsanitize=undefined -g" ..
    make
    ctest --output-on-failure

# Go層の基本テスト
- name: Run Go tests with race detector
  run: |
    cd golang
    go test -v -race ./...

# C層リークカウンターテスト（軽量）
- name: Run C heap leak counter tests
  run: |
    cd golang
    go test -v -tags=debug -run=TestMemoryLeak_.*_CCounter ./...

# 軽量Valgrindテスト（Linux限定）
- name: Valgrind leak check (Linux, lightweight)
  if: runner.os == 'Linux'
  run: |
    sudo apt-get install -y valgrind
    cd golang
    CGO_ENABLED=1 go test -c -tags=valgrind -o test.valgrind.bin ./...
    valgrind --leak-check=full --error-exitcode=1 ./test.valgrind.bin -test.run=ValgrindLeak

# カバレッジレポート
- name: Generate coverage report
  run: |
    cd golang
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# ベンチマーク（main pushのみ）
- name: Run benchmarks
  if: github.ref == 'refs/heads/main'
  run: |
    cd golang
    go test -v -bench=. -benchmem ./... | tee benchmark.txt

# 夜間ジョブ: 長時間メモリリークテスト
- name: Long-running memory leak test
  if: github.event_name == 'schedule'
  run: |
    cd golang
    go test -v -count=1000 -timeout=2h -run=TestMemoryLeak ./...
```

# 開発計画

## Phase 1: 基盤構築とFFI設計 (Week 1-2)

- [x] プロジェクト構造の確立
- [x] git submodulesの設定 (libwebp, libavif)
- [ ] C言語FFI基本インターフェースの設計と実装
  - [ ] `nextimage.h`: 共通定義（ピクセルフォーマット、バッファ構造）
  - [ ] メモリ管理の実装（_alloc / _into分離、リークカウンター）
  - [ ] エラーハンドリングの実装（スレッドローカルエラーメッセージ）
- [ ] CMakeビルドシステムの構築
  - [ ] 通常ビルド、ASanビルド、UBSanビルドの3種類
  - [ ] 依存関係の完全な静的リンク設定
- [ ] CI/CDワークフローの基本設定
- [ ] 依存関係BOMの初期バージョン作成

## Phase 2: WebP実装 (Week 3-4)

- [ ] C言語WebP FFIの実装
  - [ ] webp_encode_alloc機能の実装
  - [ ] webp_decode_alloc / webp_decode_into機能の実装
  - [ ] gif2webp機能の実装
  - [ ] ASan/UBSanでの基本テスト
- [ ] Go言語WebPバインディングの実装
  - [ ] 明示的関数API（EncodeBytes, EncodeFile, EncodeStream等）
  - [ ] オプション構造体の定義
  - [ ] CGO統合（完全な依存関係リンク）
- [ ] WebPユニットテストの作成
  - [ ] C層リークカウンターテスト
  - [ ] 全入出力パターンテスト
  - [ ] エラーハンドリングテスト
- [ ] WebPテストデータの準備（破損ファイルを含む）

## Phase 3: AVIF実装 (Week 5-6)

- [ ] C言語AVIF FFIの実装
  - [ ] avif_encode_alloc機能の実装（10bit/12bit対応）
  - [ ] avif_decode_alloc / avif_decode_into機能の実装
  - [ ] YUV 4:2:2/4:4:4サポート
  - [ ] ASan/UBSanでの基本テスト
- [ ] Go言語AVIFバインディングの実装
- [ ] AVIFユニットテストの作成
- [ ] AVIFテストデータの準備（YUV各種、10bit/12bit、アニメーションを含む）

## Phase 4: 新機能実装 (Week 7)

- [ ] webp2gifの実装
  - [ ] C言語FFI
  - [ ] Go言語バインディング
  - [ ] テスト

## Phase 5: セキュリティとファジング (Week 8-9)

- [ ] ファジングの実装
  - [ ] go-fuzzまたはOSS-Fuzz統合
  - [ ] 破損データに対するロバストネステスト
  - [ ] クラッシュしないことの確認
- [ ] セキュリティレビュー
  - [ ] バッファオーバーフロー可能性のレビュー
  - [ ] 整数オーバーフロー可能性のレビュー
  - [ ] メモリ安全性の最終確認
- [ ] ライセンスコンプライアンス
  - [ ] 全依存ライブラリのライセンス確認
  - [ ] LICENSES/ディレクトリの整備
  - [ ] DEPENDENCIES.txtの完成
- [ ] パフォーマンステスト
  - [ ] プロセス生成版との速度比較
  - [ ] ベンチマーク結果のドキュメント化

## Phase 6: 最適化とプラットフォーム検証 (Week 10-11)

- [ ] 各種プラットフォームでの動作確認
  - [ ] macOS ARM64/Intel
  - [ ] Linux x64/ARM64
  - [ ] Windows x64
- [ ] メモリリークチェックの最終確認
  - [ ] Valgrind完全テスト
  - [ ] 長時間実行テスト（夜間ジョブ）
- [ ] 並行処理の安全性確認
  - [ ] Race detector完全テスト
  - [ ] 高負荷並行テスト
- [ ] ドキュメント整備
  - [ ] API リファレンス
  - [ ] 使用例集
  - [ ] トラブルシューティングガイド

## Phase 7: リリース準備 (Week 12)

- [ ] プリコンパイル済みライブラリのビルド自動化
  - [ ] 全プラットフォームのクロスビルド
  - [ ] 依存関係の完全なバンドル確認
  - [ ] ライセンスファイルの同梱
- [ ] GitHub Releaseワークフローの構築
  - [ ] 自動ビルド・テスト・リリース
  - [ ] バージョンタグからのリリースノート生成
- [ ] README、ドキュメントの完成
- [ ] セキュリティ監査の最終レビュー
- [ ] v1.0.0リリース

# バージョニング

セマンティックバージョニングを使用し、依存ライブラリのバージョンをメタデータとして付与します。

形式: `MAJOR.MINOR.PATCH+libwebpX.Y.Z-libavifA.B.C`

例: `1.0.0+libwebp1.6.0-libavif1.3.0`

- MAJOR: 破壊的変更
- MINOR: 後方互換性のある機能追加
- PATCH: 後方互換性のあるバグフィックス
- メタデータ: 依存ライブラリのバージョン

# CI/CDワークフロー

## ビルドワークフロー (.github/workflows/build.yml)

- トリガー: push, pull_request
- ジョブ:
  - コードフォーマットチェック
  - 各プラットフォームでのビルド (darwin-arm64, darwin-amd64, linux-amd64, windows-amd64)
  - ユニットテストの実行
  - カバレッジレポート

## リリースワークフロー (.github/workflows/release.yml)

- トリガー: v*タグのpush
- ジョブ:
  - 全プラットフォームでのビルド
  - 静的ライブラリのアーカイブ作成
  - GitHub Releaseへのアップロード
  - Go ModuleのProxy更新通知

# 開発時の注意事項

## CGOの制約

- CGOを使用するため、クロスコンパイルには制限がある
- 各プラットフォーム用のビルド環境が必要
- 静的リンクを使用してデプロイメントを簡素化

## メモリ管理

- C言語で確保したメモリは必ず適切な解放関数で解放
  - エンコードバッファ: `nextimage_free_encode_buffer()`
  - デコードバッファ: `nextimage_free_decode_buffer()`
- `owns_data`フラグが1の場合のみライブラリが解放の責任を持つ
- Go側でのファイナライザーの適切な設定
- メモリリークのチェックをCIで自動化（ASan、リークカウンター、Valgrind）

## スレッドセーフ

- WebP/AVIFライブラリのスレッドセーフ性を確認
- 必要に応じてmutexによる保護を実装

## エラーハンドリング

- C言語のエラーコードをGoのerrorに適切に変換
- エラーメッセージは英語で、詳細な情報を含める
