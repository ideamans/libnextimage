# libnextimage

高性能WebP/AVIFエンコード/デコードライブラリ（Go用FFIインターフェース）

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/ideamans/libnextimage)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0--alpha-orange)](DEPENDENCIES.txt)

## 概要

`libnextimage`は、WebPとAVIFのエンコード/デコード機能への直接FFIアクセスを提供し、画像変換操作のために別プロセスを起動するオーバーヘッドを排除します。

### 主な機能

- **ゼロプロセスオーバーヘッド**: cwebp/dwebp/avifenc/avifdecプロセスを起動せず、ライブラリを直接呼び出し
- **マルチプレーンサポート**: YUVプレーナー形式の完全サポート（4:2:0、4:2:2、4:4:4）
- **高ビット深度**: 8ビット、10ビット、12ビットAVIFエンコードのサポート
- **メモリセーフ**: AddressSanitizerとUndefinedBehaviorSanitizerによる包括的なテスト
- **スレッドセーフ**: スレッドローカルエラーハンドリングと並行エンコード/デコード
- **Go統合**: バイト、ファイル、ストリーム用の明示的な関数を持つGoらしいAPI

## 使用例（Go）

### WebPエンコード/デコード

```go
package main

import (
    "fmt"
    "os"
    "github.com/ideamans/libnextimage/golang"
)

func main() {
    // PNGファイルを読み込み
    pngData, err := os.ReadFile("input.png")
    if err != nil {
        panic(err)
    }

    // PNG→WebP変換
    opts := libnextimage.DefaultWebPEncodeOptions()
    opts.Quality = 90.0
    opts.Method = 6  // 高品質

    webpData, err := libnextimage.WebPEncodeBytes(pngData, opts)
    if err != nil {
        panic(err)
    }

    // WebPファイルを保存
    os.WriteFile("output.webp", webpData, 0644)
    fmt.Printf("PNG→WebP変換完了: %d バイト\n", len(webpData))

    // WebP→PNGに戻す（メモリ版）
    pngDataOut, err := libnextimage.WebPDecodeToPNGBytes(
        webpData,
        libnextimage.DefaultWebPDecodeOptions(),
        9, // PNG圧縮レベル
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("WebP→PNG変換完了: %d バイト\n", len(pngDataOut))
    os.WriteFile("output.png", pngDataOut, 0644)
}
```

### AVIF高度なオプションでエンコード

```go
import "github.com/ideamans/libnextimage/golang"

// PNGを高品質でAVIFにエンコード
opts := libnextimage.DefaultAVIFEncodeOptions()
opts.Quality = 90
opts.Speed = 4
opts.YUVFormat = 0  // 4:4:4で最高品質

avifData, err := libnextimage.AVIFEncodeFile("input.png", opts)
if err != nil {
    panic(err)
}

os.WriteFile("output.avif", avifData, 0644)
```

### AVIF→PNG/JPEG変換

```go
import "github.com/ideamans/libnextimage/golang"

// AVIF→PNG変換（圧縮付き）
avifData, _ := os.ReadFile("input.avif")
decOpts := libnextimage.DefaultAVIFDecodeOptions()
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingBestQuality

err := libnextimage.AVIFDecodeToPNG(
    avifData,
    "output.png",
    decOpts,
    9,  // PNG圧縮レベル (0-9, -1=デフォルト)
)

// AVIF→JPEG変換
err = libnextimage.AVIFDecodeToJPEG(
    avifData,
    "output.jpg",
    decOpts,
    90,  // JPEG品質 (1-100)
)

// ファイルベースの変換
err = libnextimage.AVIFDecodeFileToPNG(
    "input.avif",
    "output.png",
    decOpts,
    -1,  // デフォルト圧縮
)
```

### クロマアップサンプリングオプション

```go
import "github.com/ideamans/libnextimage/golang"

decOpts := libnextimage.DefaultAVIFDecodeOptions()

// 利用可能なアップサンプリングモード:
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingAutomatic   // 0 (デフォルト)
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingFastest     // 1 (最速)
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingBestQuality // 2 (最高品質)
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingNearest     // 3 (最近傍)
decOpts.ChromaUpsampling = libnextimage.ChromaUpsamplingBilinear    // 4 (バイリニア)
```

### AVIFデコード時のセキュリティ制限

```go
import "github.com/ideamans/libnextimage/golang"

decOpts := libnextimage.DefaultAVIFDecodeOptions()
decOpts.ImageSizeLimit = 100_000_000      // 最大1億ピクセル
decOpts.ImageDimensionLimit = 16384       // 最大16384px幅/高さ
decOpts.StrictFlags = 1                   // 厳格な検証を有効化

decoded, err := libnextimage.AVIFDecodeBytes(avifData, decOpts)
```

## API リファレンス

### AVIF変換関数

#### AVIFDecodeToPNG

AVIFデータをデコードしてPNGファイルとして保存します。

```go
func AVIFDecodeToPNG(
    avifData []byte,
    outputPath string,
    options AVIFDecodeOptions,
    pngCompressionLevel int,
) error
```

**パラメータ:**
- `avifData`: AVIFファイルのバイナリデータ
- `outputPath`: 出力PNGファイルのパス
- `options`: デコードオプション
- `pngCompressionLevel`: PNG圧縮レベル
  - `0`: 無圧縮（最速、最大サイズ）
  - `9`: 最高圧縮（最遅、最小サイズ）
  - `-1`: デフォルト圧縮

#### AVIFDecodeToJPEG

AVIFデータをデコードしてJPEGファイルとして保存します。

```go
func AVIFDecodeToJPEG(
    avifData []byte,
    outputPath string,
    options AVIFDecodeOptions,
    jpegQuality int,
) error
```

**パラメータ:**
- `avifData`: AVIFファイルのバイナリデータ
- `outputPath`: 出力JPEGファイルのパス
- `options`: デコードオプション
- `jpegQuality`: JPEG品質 (1-100)
  - `1`: 最低品質（最小サイズ）
  - `100`: 最高品質（最大サイズ）
  - 範囲外の値は自動的に補正されます

#### AVIFDecodeFileToPNG

AVIFファイルをPNGファイルに変換します。

```go
func AVIFDecodeFileToPNG(
    avifPath string,
    pngPath string,
    options AVIFDecodeOptions,
    pngCompressionLevel int,
) error
```

#### AVIFDecodeFileToJPEG

AVIFファイルをJPEGファイルに変換します。

```go
func AVIFDecodeFileToJPEG(
    avifPath string,
    jpegPath string,
    options AVIFDecodeOptions,
    jpegQuality int,
) error
```

### AVIFDecodeOptions

```go
type AVIFDecodeOptions struct {
    UseThreads          bool             // マルチスレッド有効化
    Format              PixelFormat      // ピクセル形式 (RGBA, RGB, etc.)
    IgnoreExif          bool             // EXIFメタデータを無視
    IgnoreXMP           bool             // XMPメタデータを無視
    IgnoreICC           bool             // ICCプロファイルを無視
    ImageSizeLimit      uint32           // 最大画像サイズ（総ピクセル数）
    ImageDimensionLimit uint32           // 最大画像寸法（幅または高さ）
    StrictFlags         int              // 厳格な検証フラグ
    ChromaUpsampling    ChromaUpsampling // クロマアップサンプリングモード
}
```

**デフォルト値:**
- `ImageSizeLimit`: 268,435,456 ピクセル（16384 × 16384）
- `ImageDimensionLimit`: 32768
- `StrictFlags`: 1（厳格な検証有効）
- `ChromaUpsampling`: 0（自動）

### ChromaUpsampling型

```go
type ChromaUpsampling int

const (
    ChromaUpsamplingAutomatic   ChromaUpsampling = 0  // 自動選択（デフォルト）
    ChromaUpsamplingFastest     ChromaUpsampling = 1  // 最速
    ChromaUpsamplingBestQuality ChromaUpsampling = 2  // 最高品質
    ChromaUpsamplingNearest     ChromaUpsampling = 3  // 最近傍補間
    ChromaUpsamplingBilinear    ChromaUpsampling = 4  // バイリニア補間
)
```

## クイックスタート

### ソースからビルド

#### 前提条件

- CMake 3.15以降
- C11互換コンパイラ（GCC、Clang、またはMSVC）
- Git（サブモジュール管理用）

#### 基本的なビルド

```bash
# リポジトリをクローン
git clone --recursive https://github.com/ideamans/libnextimage.git
cd libnextimage

# Cライブラリをビルド
cd c
mkdir build && cd build
cmake ..
cmake --build .

# テストを実行
ctest --output-on-failure
```

## テスト

### Goテストの実行

```bash
cd golang

# 全テスト（詳細表示）
go test -v

# AVIFテストのみ実行
go test -v -run TestAVIF

# AVIF変換テストのみ実行
go test -v -run TestAVIF.*Convert

# レースディテクター付き
go test -race

# カバレッジ
go test -cover
```

### テスト結果

**AVIFテスト: 22グループ、65個のテストケース - 全てパス** ✅

- デコード機能: 18グループ、53テスト
- PNG/JPEG変換機能: 4グループ、12テスト

## avifdec互換性

avifdecコマンドライン�ールの**コア機能を完全サポート** ✅

### 対応機能

- ✅ **メタデータ無視オプション**: `ignore_exif`, `ignore_xmp`, `ignore_icc`
- ✅ **セキュリティ制限**: `image_size_limit`, `image_dimension_limit` 🔒
- ✅ **厳格な検証制御**: `strict_flags`
- ✅ **PNG/JPEG変換機能**: `-q`, `--png-compress` 🎨
- ✅ **クロマアップサンプリング**: `-u, --upsampling`

### avifdecオプション対応表

| avifdecオプション | libnextimage API | 説明 |
|------------------|------------------|------|
| `-q, --quality Q` | `AVIFDecodeToJPEG()` の `jpegQuality` | JPEG品質 (1-100) |
| `--png-compress L` | `AVIFDecodeToPNG()` の `pngCompressionLevel` | PNG圧縮レベル (0-9) |
| `-u, --upsampling U` | `options.ChromaUpsampling` | クロマアップサンプリングモード |
| `--no-strict` | `options.StrictFlags = 0` | 厳格な検証を無効化 |
| `--size-limit C` | `options.ImageSizeLimit` | 最大画像サイズ（ピクセル数） |
| `--dimension-limit C` | `options.ImageDimensionLimit` | 最大画像寸法 |
| `--ignore-exif` | `options.IgnoreExif = true` | EXIFメタデータを無視 |
| `--ignore-xmp` | `options.IgnoreXMP = true` | XMPメタデータを無視 |
| `--ignore-icc` | `options.IgnoreICC = true` | ICCプロファイルを無視 |

## プロジェクト構造

```
libnextimage/
├── c/                        # C FFI層
│   ├── include/              # 公開ヘッダー
│   │   ├── nextimage.h       # メインFFIインターフェース
│   │   ├── webp.h            # WebP API
│   │   └── avif.h            # AVIF API
│   ├── src/                  # 実装
│   │   ├── common.c          # メモリ・エラーハンドリング
│   │   ├── webp.c            # WebP実装
│   │   └── avif.c            # AVIF実装
│   └── CMakeLists.txt
├── deps/                     # 依存関係（gitサブモジュール）
│   ├── libwebp/              # WebPライブラリ
│   └── libavif/              # AVIFライブラリ
├── golang/                   # Goバインディング
│   ├── common.go             # 共通型・ユーティリティ
│   ├── webp.go               # WebP Go API
│   ├── avif.go               # AVIF Go API
│   ├── avif_convert.go       # AVIF変換機能
│   └── *_test.go             # テスト
├── SPEC.md                   # 詳細仕様
├── COMPAT.md                 # 互換性ドキュメント
├── DEPENDENCIES.txt          # 部品表
└── LICENSE                   # MITライセンス
```

## ドキュメント

- [README.md](README.md) - 英語版README
- [SPEC.md](SPEC.md) - 包括的な仕様と開発計画
- [COMPAT.md](COMPAT.md) - cwebp/dwebp/avifenc/avifdec互換性ドキュメント
- [DEPENDENCIES.txt](DEPENDENCIES.txt) - 全依存関係の部品表

## ライセンス

このプロジェクトはMITライセンスの下でライセンスされています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。

## クレジット

開発: [株式会社アイデアマンズ](https://www.ideamans.com/)

以下のライブラリを使用:
- [libwebp](https://github.com/webmproject/libwebp)
- [libavif](https://github.com/AOMediaCodec/libavif)
