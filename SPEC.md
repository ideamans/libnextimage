# このプロジェクトは

libwebp, libavifに含まれるエンコード・デコードコマンドである、

- cwebp / dwebp / gif2webp
- avifenc / avifdec

これらをFFIおよびGoのラッパーとして利用できるようにするものです。

また、アニメーションwebpからアニメーションGIFへの変換を行う`webp2gif`コマンドも新設します。

通常、これらのコマンドはバイナリファイルとして同梱し、プロセスを生成して実行しますが、このプロジェクトではそれらのコマンドを直接呼び出すことで、プロセス生成のオーバーヘッドを削減し、パフォーマンスの向上を図ります。

## プロジェクトの使命

**このライブラリの最も重要な使命は、対応するコマンドラインツールの動作を可能な限り高い精度で再現することです。**

単なる画像変換ライブラリではなく、公式コマンドラインツール（cwebp, dwebp, gif2webp, avifenc, avifdec）の**完全互換実装**を目指します。これにより、既存のコマンドラインツールからライブラリAPIへの移行を安心して行えます。

### CLI Clone Philosophy - 設計の大原則

このライブラリは**CLIツールのクローン**として設計されています。以下の5つの大原則を厳守します：

1. **Complete Option Parity（完全なオプション対応）**
   - CLI固有オプション（`-v`, `--version`, `-h`, `--help`, `-quiet`, `-progress`など）を除き、すべてのCLIオプションに対応するライブラリオプションを提供する
   - 例外なく、すべての機能オプションをライブラリで利用可能にする

2. **Identical Behavior（同一動作保証）**
   - 同じオプション値を指定した場合、CLIツールとライブラリは**バイト完全一致**の出力を生成する
   - エンコーダの変動がある場合でも、SHA256ハッシュまたは規定の許容範囲内（WebP: ±2%, AVIF: ±10%）で一致する
   - エラー条件も同一（同じ入力で同じエラーを返す）

3. **Type Safety over Magic Numbers（型安全性優先）**
   - マジックナンバーを避け、必ず定数、enum、または明確な型を使用する
   - 例: `FilterType int` ではなく `FilterType WebPFilterType` enum
   - ユーザーが誤った値を設定できないようにする

4. **Flexible Input（柔軟な入力形式）**
   - CLIが複数値を受け取る場合（例: `-qrange <min> <max>`）、ライブラリは構造化された入力を提供する
   - 例: `QMin int` と `QMax int` の2つのフィールド、または配列/スライス
   - CLIが複数の組み合わせを許す場合（例: `-metadata exif,xmp,icc`）、ビットフラグまたは複数boolで対応する

5. **No Feature Gaps（機能の欠落なし）**
   - CLIでサポートされている機能は、すべてライブラリAPIでサポートする
   - 「CLIでできるがライブラリではできない」という状況を作らない
   - 非対応機能は明示的に文書化し、理由を説明する

### 検証基準

- **Binary Identity**: バイナリ完全一致（理想）
- **Hash Identity**: SHA256ハッシュ一致（許容）
- **Size Tolerance**: サイズが許容範囲内（WebP: ±2%, AVIF: ±10%）
- **Pixel Identity**: デコーダーではピクセルデータが一致または許容範囲内

これらの原則からの逸脱は、バグとして扱い、修正を追跡します。

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

# アーキテクチャ

## レイヤー構造

```
┌─────────────────────────────────────────────────────────┐
│ Go言語パッケージ (cwebp, dwebp, avifenc, avifdec)        │
│                                                         │
│  Run([]byte) ([]byte, error)                           │ ← コアメソッド (CGOラッパー)
│    ↓ CGO呼び出し                                         │
│  cwebp_run_command()                                   │
│                                                         │
│  RunFile(string, string) error                         │ ← シュガーシンタックス
│    ↓ 内部実装                                            │
│  os.ReadFile → Run() → os.WriteFile                   │
│                                                         │
│  RunIO(io.Reader, io.Writer) error                     │ ← シュガーシンタックス
│    ↓ 内部実装                                            │
│  io.ReadAll → Run() → io.Write                        │
└─────────────────────────────────────────────────────────┘
                      ↓ CGO
┌─────────────────────────────────────────────────────────┐
│ C言語FFI (cwebp.h, dwebp.h, avifenc.h, avifdec.h)       │
│                                                         │
│  cwebp_run_command()                                   │ ← 実装の本体
│    ↓                                                    │
│  libwebp API呼び出し (WebPEncode等)                     │
└─────────────────────────────────────────────────────────┘
```

## 設計原則

- **C言語がコア実装**: libwebp/libavifを直接呼び出し、バイト列変換を実装
- **Go言語はCGOラッパー**: C言語のコア実装を薄くラップし、Go的なインターフェースを提供
- **コマンドオブジェクトによる連続使用**: 初期化オーバーヘッドを削減し、同じ設定で複数の変換を効率的に実行
- **バイト列ベースのインターフェース**: 画像ファイル形式（JPEG/PNG/WebP/AVIF）のバイト列のみを扱う
- **シュガーシンタックス（Go言語のみ）**: ファイル/IOはコアメソッド`Run()`を使う薄いラッパー
- **コマンド名との一致**: パッケージ名、関数プレフィックス、型名をコマンド名（cwebp, dwebp等）に合わせる
- **デフォルト設定の作成**: C言語は`*_create_default_options()`、Go言語は`NewDefaultOptions()`
- **明示的なリソース解放**: C言語は`*_free_*()`、Go言語は`Close()`
- **エラーハンドリング**: C言語はステータスコード+エラーメッセージ、Go言語はerror型
- **スレッドセーフ**: スレッドローカルストレージでエラーメッセージを管理

# 互換性保証とテスト基準

## コマンドライン互換性の要件

このライブラリは、公式コマンドラインツールとの**高精度な互換性**を保証します。リファクタリング後も以下の基準を維持します。

### 1. gif2webp - バイナリ完全一致（Binary Exact Match）

**達成基準**: 公式gif2webpコマンドと**バイト単位で完全一致**する出力

**現在の達成状況**: ✅ 全12テストケースでバイナリ完全一致

**必須テストケース**:
- ✅ 静止画GIF（4テスト）: static-64x64, static-512x512, static-16x16, gradient
- ✅ アニメーションGIF（2テスト）: animated-3frames, animated-small
- ✅ 透過GIF（2テスト）: static-alpha, animated-alpha
- ✅ 品質設定（2テスト）: quality-50, quality-90
- ✅ メソッド設定（2テスト）: method-0, method-6

**技術的要件**:
- 単一フレームGIFは静的VP8L WebPエンコード（`WebPEncode()`使用）
- アニメーションGIFはWebPAnimEncoderによる変換
- デフォルトlosslessエンコード（gif2webpの動作に準拠）
- GIFフレームタイミング、透明度、ディスポーズメソッドの完全保持
- 3フレームバッファ方式（frame, curr_canvas, prev_canvas）による正確な合成

### 2. cwebp - 全コア機能の完全サポート

**達成基準**: cwebpの全オプションを正確に実装し、同等の出力を保証

**必須対応オプション**:
- ✅ 基本オプション: `-q`, `-alpha_q`, `-preset`, `-z`, `-m`, `-segments`, `-size`, `-psnr`
- ✅ フィルタリング: `-sns`, `-f`, `-sharpness`, `-strong`, `-sharp_yuv`, `-af`
- ✅ 圧縮: `-partition_limit`, `-pass`, `-qrange`, `-mt`, `-low_memory`
- ✅ アルファ: `-alpha_method`, `-alpha_filter`, `-exact`, `-blend_alpha`, `-noalpha`
- ✅ ロスレス: `-lossless`, `-near_lossless`, `-hint`
- ✅ メタデータ: `-metadata` (all, none, exif, icc, xmp)
- ✅ **画像変換**: `-crop`, `-resize`, `-resize_mode` (up_only, down_only, always)
- ✅ 実験的: `-jpeg_like`, `-pre`

**処理順序の厳守**:
```
画像読み込み → crop → resize → blend_alpha → エンコード
```

**必須テスト項目**:
- Crop機能（256x256クロップ）
- Resize機能（200x200リサイズ）
- Resize Mode（up_only/down_only/always）
- Crop + Resize組み合わせ
- Blend Alpha（背景色合成）
- No Alpha（アルファ除去）

### 3. dwebp - デコード機能の完全サポート

**達成基準**: dwebpの全デコードオプションを正確に実装

**必須対応オプション**:
- ✅ デコード品質: `-nofancy`, `-nofilter`, `-nodither`, `-dither`, `-alpha_dither`, `-mt`
- ✅ 画像変換: `-crop`, `-resize`, `-flip`, `-alpha`
- ✅ インクリメンタルデコード: `-incremental`

**重要な差異**:
- dwebpはデコード後にcrop/resizeを適用
- cwebpはエンコード前にcrop/resizeを適用
- ライブラリは両方の処理順序をサポート

### 4. avifenc - 静止画AVIF完全サポート

**達成基準**: avifencの静止画機能を完全実装（アニメーション機能は明示的に非対応）

**必須対応オプション**:
- ✅ 基本: `-q`, `--qalpha`, `-s`, `-l`, `-d`, `-y`, `-p`, `--sharpyuv`
- ✅ 色空間: `--cicp`, `--nclx`, `-r` (YUV range)
- ✅ ファイルサイズ: `--target-size`
- ✅ メタデータ: `--exif`, `--xmp`, `--icc`, `--ignore-*`
- ✅ **画像プロパティ**: `--pasp`, `--crop`, `--clap`, `--irot`, `--imir`, `--clli`
- ✅ タイリング: `--tilerowslog2`, `--tilecolslog2`
- ✅ 品質設定: `--min`, `--max`, `--minalpha`, `--maxalpha`

**非対応（明示的）**:
- ❌ アニメーション: `--timescale`, `--keyframe`, `--repetition-count`, `--duration`
- ❌ 実験的機能: `--progressive`, `--layered`, `--scaling-mode`
- ❌ グリッド: `-g`, `--grid`

### 5. avifdec - デコード・変換機能の完全サポート

**達成基準**: avifdecの全デコード・変換機能を正確に実装

**必須対応オプション**:
- ✅ **PNG/JPEG変換**: `-q` (JPEG品質 1-100), `--png-compress` (0-9)
- ✅ **クロマアップサンプリング**: `-u` (0=auto, 1=fastest, 2=best_quality, 3=nearest, 4=bilinear)
- ✅ **セキュリティ**: `--size-limit`, `--dimension-limit` 🔒
- ✅ **厳格な検証**: `--no-strict` (strict_flags)
- ✅ **メタデータ**: `--icc`, `--ignore-icc`, `--ignore-exif`, `--ignore-xmp`

**互換性テスト結果（必須維持）**:
- ✅ デコード機能: 18テストグループ、53個のテストケース
- ✅ PNG/JPEG変換: 4テストグループ、12個のテストケース
- ✅ セキュリティ制限: 正常サイズ/制限超過の動作確認
- ✅ クロマアップサンプリング: 5モード全ての動作確認

**セキュリティ要件**:
- デフォルト制限: 268,435,456ピクセル（16384 × 16384）
- 寸法制限: 32768（デフォルト）
- 制限超過時は適切なエラーを返す

## 非対応機能の明確化

以下の機能は**意図的に非対応**とし、実装しません:

### CLI固有機能（全コマンド共通）
- `-h`, `--help`, `-V`, `--version`, `-v`, `--verbose`, `-q`, `--quiet`, `--progress`
- **理由**: コマンドラインツール固有の機能で、ライブラリAPIでは不要

### デバッグ・統計出力（cwebp）
- `-print_psnr`, `-print_ssim`, `-print_lsim`, `-d <file.pgm>`
- **理由**: デバッグ・統計情報出力はライブラリAPIでは不要

### 生ピクセル入力（cwebp/avifenc）
- `-s <int> <int>` (YUV入力), `--stdin` (y4m入力)
- **理由**: 生ピクセルデータ入力は別API設計が必要。本ライブラリは画像ファイル形式のみを扱う

### 出力フォーマット変換（dwebp）
- `-pam`, `-ppm`, `-bmp`, `-tiff`, `-pgm`, `-yuv`
- **理由**: 出力フォーマット変換は出力側で処理すべき。ライブラリはRGBAピクセルデータを返す

### AVIFアニメーション機能
- `--timescale`, `--fps`, `--keyframe`, `--repetition-count`, `--duration`, `--index`
- **理由**: アニメーション機能は明示的に非対応（静止画のみ対応）

### 実験的機能
- `--progressive`, `--layered`, `--scaling-mode`, `-a, --advanced`
- **理由**: 実験的機能であり不要

### システム固有設定
- `-j, --jobs` (スレッド数指定), `-c, --codec` (コーデック選択), `--noasm`, `--autotiling`
- **理由**: libwebp/libavif内部で管理、またはコマンド専用機能

## テスト基準の維持

リファクタリング後も以下のテスト基準を**必ず維持**します:

### 1. バイナリ完全一致テスト（gif2webp）
```bash
# 全12テストケースでバイト単位の完全一致を確認
go test -v -run TestGif2WebPBinaryCompatibility
```

### 2. 機能互換性テスト（全コマンド）
```bash
# cwebp: 画像変換機能のテスト
go test -v -run TestCWebPAdvanced

# dwebp: デコード機能のテスト
go test -v -run TestDWebP

# avifenc: エンコード機能のテスト
go test -v -run TestAVIFEnc

# avifdec: デコード・変換・セキュリティのテスト（65個のテストケース）
go test -v -run TestAVIFDec
```

### 3. オプション網羅性テスト
各コマンドの全オプションが正しく動作することを確認:
- 全オプションの組み合わせテスト
- デフォルト値の検証
- 境界値テスト（最小値・最大値）
- エラーハンドリングの検証

### 4. メモリ安全性テスト
```bash
# メモリリーク検出
go test -v -run . -memprofile=mem.prof

# 競合状態の検出
go test -race -v -run .
```

## 互換性保証のレベル

| コマンド | 互換性レベル | 説明 |
|---------|------------|------|
| **gif2webp** | **バイナリ完全一致** ✅ | 公式コマンドと1バイトも違わない出力 |
| **cwebp** | **機能完全互換** ✅ | 全コアオプションを正確に実装 |
| **dwebp** | **機能完全互換** ✅ | 全デコードオプションを正確に実装 |
| **avifenc** | **静止画完全互換** ✅ | 静止画機能を完全実装（アニメーション除く） |
| **avifdec** | **機能完全互換** ✅ | デコード・変換・セキュリティ全機能実装 |

## リファクタリング時の注意事項

**インターフェース変更を行う場合でも、以下を必ず維持してください**:

1. **gif2webpのバイナリ完全一致** - これは最優先の互換性保証です
2. **全テストケースの成功** - 既存の全テストが必ず成功すること
3. **コマンドオプションの網羅性** - COMPAT.mdに記載された全オプションのサポート
4. **処理順序の厳守** - crop → resize → blend_alpha などの処理順序
5. **エラーハンドリングの一貫性** - 公式コマンドと同じエラー条件で失敗すること

**新しいインターフェースでも、以下の機能は必ず実装してください**:

- コマンドオブジェクトの再利用（初期化オーバーヘッド削減）
- バイト列ベースのコア実装
- デフォルト設定からの部分的変更
- 明示的なリソース解放
- 詳細なエラーメッセージ

これらの基準を下げることなく、より良いインターフェース設計を目指してください。

# C言語FFIインターフェース

## C言語使用例

```c
#include "nextimage/cwebp.h"
#include <stdio.h>
#include <stdlib.h>

int main() {
    // デフォルト設定を作成し、部分的に変更
    CWebPOptions* options = cwebp_create_default_options();
    options->quality = 80;
    options->method = 4;

    // コマンドを作成（この時点で初期化完了）
    CWebPCommand* cmd = cwebp_new_command(options);
    if (!cmd) {
        fprintf(stderr, "Failed to create command\n");
        cwebp_free_options(options);
        return 1;
    }

    // 同じコマンドで複数の画像を連続変換
    // 画像1
    FILE* f1 = fopen("image1.jpg", "rb");
    fseek(f1, 0, SEEK_END);
    size_t size1 = ftell(f1);
    rewind(f1);
    uint8_t* jpeg1 = malloc(size1);
    fread(jpeg1, 1, size1, f1);
    fclose(f1);

    NextImageBuffer webp1;
    NextImageStatus status = cwebp_run_command(cmd, jpeg1, size1, &webp1);
    if (status == NEXTIMAGE_OK) {
        FILE* out1 = fopen("image1.webp", "wb");
        fwrite(webp1.data, 1, webp1.size, out1);
        fclose(out1);
        nextimage_free_buffer(&webp1);
    }
    free(jpeg1);

    // 画像2（同じコマンドを再利用）
    FILE* f2 = fopen("image2.jpg", "rb");
    fseek(f2, 0, SEEK_END);
    size_t size2 = ftell(f2);
    rewind(f2);
    uint8_t* jpeg2 = malloc(size2);
    fread(jpeg2, 1, size2, f2);
    fclose(f2);

    NextImageBuffer webp2;
    status = cwebp_run_command(cmd, jpeg2, size2, &webp2);
    if (status == NEXTIMAGE_OK) {
        FILE* out2 = fopen("image2.webp", "wb");
        fwrite(webp2.data, 1, webp2.size, out2);
        fclose(out2);
        nextimage_free_buffer(&webp2);
    }
    free(jpeg2);

    // リソース解放
    cwebp_free_command(cmd);
    cwebp_free_options(options);

    return 0;
}
```

## 基本インターフェース

```c
// nextimage.h - 共通定義

// ステータスコード
typedef enum {
    NEXTIMAGE_OK = 0,
    NEXTIMAGE_ERROR_INVALID_PARAM = -1,
    NEXTIMAGE_ERROR_ENCODE_FAILED = -2,
    NEXTIMAGE_ERROR_DECODE_FAILED = -3,
    NEXTIMAGE_ERROR_OUT_OF_MEMORY = -4,
    NEXTIMAGE_ERROR_UNSUPPORTED = -5,
    NEXTIMAGE_ERROR_IO_FAILED = -6,
} NextImageStatus;

// 出力バッファ（画像ファイル形式のバイト列）
typedef struct {
    uint8_t* data;
    size_t size;
} NextImageBuffer;

// バッファのメモリ解放
void nextimage_free_buffer(NextImageBuffer* buffer);

// エラーメッセージ取得
// - スレッドローカルストレージに保存された最後のエラーメッセージを返す
// - 返される文字列は次のFFI呼び出しまで有効（コピー不要だがスレッドローカル）
// - NULLが返された場合はエラーメッセージが設定されていない
const char* nextimage_last_error_message(void);

// エラーメッセージのクリア
void nextimage_clear_error(void);
```

## WebP FFI

### オプション管理

```c
// cwebp.h

typedef struct {
    float quality;           // 0-100, default 75
    int lossless;           // 0 or 1, default 0
    int method;             // 0-6, default 4
    int target_size;        // target size in bytes
    float target_psnr;      // target PSNR
    int exact;              // preserve RGB values in transparent area
    // ... その他のオプション
} CWebPOptions;

// デフォルト設定の作成
CWebPOptions* cwebp_create_default_options(void);
void cwebp_free_options(CWebPOptions* options);
```

```c
// dwebp.h

typedef struct {
    int use_threads;            // 0 or 1
    int bypass_filtering;       // 0 or 1
    int no_fancy_upsampling;    // 0 or 1
    int output_format;          // PNG, PPM, etc.
    // ... その他のオプション
} DWebPOptions;

// デフォルト設定の作成
DWebPOptions* dwebp_create_default_options(void);
void dwebp_free_options(DWebPOptions* options);
```

### コマンドインターフェース（cwebp）

```c
// 不透明なコマンド構造体
typedef struct CWebPCommand CWebPCommand;

// コマンドの作成
CWebPCommand* cwebp_new_command(const CWebPOptions* options);

// バイト列の変換
NextImageStatus cwebp_run_command(
    CWebPCommand* cmd,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output
);

// コマンドの解放
void cwebp_free_command(CWebPCommand* cmd);
```

### コマンドインターフェース（dwebp）

```c
// 不透明なコマンド構造体
typedef struct DWebPCommand DWebPCommand;

// コマンドの作成
DWebPCommand* dwebp_new_command(const DWebPOptions* options);

// バイト列の変換
NextImageStatus dwebp_run_command(
    DWebPCommand* cmd,
    const uint8_t* webp_data,
    size_t webp_size,
    NextImageBuffer* output  // PNG/JPEGなどのフォーマットで出力
);

// コマンドの解放
void dwebp_free_command(DWebPCommand* cmd);
```

## AVIF FFI

### オプション管理

```c
// avifenc.h

typedef struct {
    int quality;            // 0-100, default 50
    int speed;              // 0-10, default 6
    int min_quantizer;      // 0-63
    int max_quantizer;      // 0-63
    // ... その他のオプション
} AVIFEncOptions;

// デフォルト設定の作成
AVIFEncOptions* avifenc_create_default_options(void);
void avifenc_free_options(AVIFEncOptions* options);
```

```c
// avifdec.h

typedef struct {
    int use_threads;        // 0 or 1
    int output_format;      // PNG, JPEG, etc.
    // ... その他のオプション
} AVIFDecOptions;

// デフォルト設定の作成
AVIFDecOptions* avifdec_create_default_options(void);
void avifdec_free_options(AVIFDecOptions* options);
```

### コマンドインターフェース（avifenc）

```c
// 不透明なコマンド構造体
typedef struct AVIFEncCommand AVIFEncCommand;

// コマンドの作成
AVIFEncCommand* avifenc_new_command(const AVIFEncOptions* options);

// バイト列の変換
NextImageStatus avifenc_run_command(
    AVIFEncCommand* cmd,
    const uint8_t* input_data,
    size_t input_size,
    NextImageBuffer* output
);

// コマンドの解放
void avifenc_free_command(AVIFEncCommand* cmd);
```

### コマンドインターフェース（avifdec）

```c
// 不透明なコマンド構造体
typedef struct AVIFDecCommand AVIFDecCommand;

// コマンドの作成
AVIFDecCommand* avifdec_new_command(const AVIFDecOptions* options);

// バイト列の変換
NextImageStatus avifdec_run_command(
    AVIFDecCommand* cmd,
    const uint8_t* avif_data,
    size_t avif_size,
    NextImageBuffer* output
);

// コマンドの解放
void avifdec_free_command(AVIFDecCommand* cmd);
```

# Go言語インターフェース

## インストール

```bash
go get github.com/ideamans/libnextimage/golang
```

## 設計方針

- **連続使用を前提とした初期化オーバーヘッドの削減**: コマンドオブジェクトを一度作成し、繰り返し実行可能
- **パッケージ名はコマンド名に準拠**: `cwebp`, `dwebp`, `avifenc`, `avifdec` など
- **コアはバイト列、ファイル/IOはシュガーシンタックス**:
  - `Run([]byte)` - コアメソッド、バイト列変換
  - `RunFile(string, string)` - シュガーシンタックス、内部でファイル読み書き
  - `RunIO(io.Reader, io.Writer)` - シュガーシンタックス、内部でストリーム読み書き
- **設定値の部分的アレンジ**: `DefaultOptions()` でデフォルト値を取得し、必要な部分のみ変更
- **明示的なリソース解放**: `Close()` メソッドで確実にリソースを解放（Go言語の慣習に従う）
- エラーハンドリングはGoの標準的な方法に従う
- 詳細なエラーメッセージをGoのerror型にラップして提供

## 使用例

```go
package main

import (
    "os"
    "github.com/ideamans/libnextimage/golang/cwebp"
    "github.com/ideamans/libnextimage/golang/dwebp"
    "github.com/ideamans/libnextimage/golang/avifenc"
)

func main() {
    // デフォルト設定を作成し、部分的に変更
    options := cwebp.NewDefaultOptions()
    options.Quality = 80
    options.Method = 4

    // コマンドを作成（この時点で初期化完了）
    cmd, err := cwebp.NewCommand(options)
    if err != nil {
        panic(err)
    }
    defer cmd.Close()

    // 例1: バイト列変換（コアメソッド）
    jpeg1, _ := os.ReadFile("image1.jpg")
    webp1, err := cmd.Run(jpeg1)
    if err != nil {
        panic(err)
    }
    os.WriteFile("image1.webp", webp1, 0644)

    // 同じコマンドで2枚目も変換
    jpeg2, _ := os.ReadFile("image2.jpg")
    webp2, err := cmd.Run(jpeg2)
    if err != nil {
        panic(err)
    }
    os.WriteFile("image2.webp", webp2, 0644)

    // 例2: ファイル変換（シュガーシンタックス）
    err = cmd.RunFile("image3.jpg", "image3.webp")
    if err != nil {
        panic(err)
    }

    // 例3: IO変換（シュガーシンタックス）
    reader, _ := os.Open("image4.jpg")
    writer, _ := os.Create("image4.webp")
    err = cmd.RunIO(reader, writer)
    reader.Close()
    writer.Close()

    // 例4: デコーダー（WebP → PNG）
    decOptions := dwebp.NewDefaultOptions()
    decOptions.Format = dwebp.FormatPNG

    decCmd, _ := dwebp.NewCommand(decOptions)
    defer decCmd.Close()

    // ファイル変換のシュガーシンタックス
    err = decCmd.RunFile("input.webp", "output.png")

    // 例5: AVIF エンコーダー
    avifOpts := avifenc.NewDefaultOptions()
    avifOpts.Quality = 75
    avifOpts.Speed = 6

    avifCmd, _ := avifenc.NewCommand(avifOpts)
    defer avifCmd.Close()

    jpegData, _ := os.ReadFile("input.jpg")
    avifData, _ := avifCmd.Run(jpegData)
    os.WriteFile("output.avif", avifData, 0644)
}
```

## API設計

### 設計方針

1. **コマンドオブジェクトの連続使用**
   - 同じ設定で複数の画像を処理する場合、初期化オーバーヘッドを削減
   - コマンドを事前にセットアップし、繰り返し実行可能
   - 各コマンドは独立したパッケージとして提供（`cwebp`, `dwebp`, `avifenc`, `avifdec`など）

2. **3種類のインターフェース**
   - **Bytes**: バイト配列間の変換（メモリ間）
   - **File**: ファイルパス指定の変換（ファイル間）
   - **IO**: io.Reader/io.Writer の変換（ストリーム間）

3. **設定値の管理**
   - `DefaultOptions()` でコマンドのデフォルト設定を取得
   - 必要な部分のみ変更して使用
   - CLI コマンドとの互換性を重視

### cwebp パッケージ（WebPエンコーダー）

```go
package cwebp

// オプション構造体
type Options struct {
    Quality float32  // 0-100, default 75
    Lossless int     // 0 or 1, default 0
    Method int       // 0-6, default 4
    // ... その他のオプション
}

// デフォルト設定を作成
func NewDefaultOptions() Options

// コマンド構造体
type Command struct { /* 内部実装 */ }

// コマンドの作成
func NewCommand(opts Options) (*Command, error)

// コアメソッド: バイト列変換
func (c *Command) Run(imageData []byte) ([]byte, error)

// シュガーシンタックス: ファイル変換
// 内部で os.ReadFile → Run() → os.WriteFile を実行
func (c *Command) RunFile(inputPath string, outputPath string) error

// シュガーシンタックス: IO変換
// 内部で io.ReadAll → Run() → io.Writer.Write を実行
func (c *Command) RunIO(input io.Reader, output io.Writer) error

// リソース解放
func (c *Command) Close() error
```

### dwebp パッケージ（WebPデコーダー）

```go
package dwebp

type OutputFormat int

const (
    FormatPNG OutputFormat = iota
    FormatPPM
    FormatPGM
    FormatYUV
)

type Options struct {
    Format OutputFormat  // 出力フォーマット
    UseThreads bool      // マルチスレッド使用
    // ... その他のオプション
}

func NewDefaultOptions() Options

type Command struct { /* 内部実装 */ }

func NewCommand(opts Options) (*Command, error)
func (c *Command) Run(webpData []byte) ([]byte, error)           // コアメソッド
func (c *Command) RunFile(inputPath, outputPath string) error    // シュガーシンタックス
func (c *Command) RunIO(input io.Reader, output io.Writer) error // シュガーシンタックス
func (c *Command) Close() error
```

### avifenc パッケージ（AVIFエンコーダー）

```go
package avifenc

type Options struct {
    Quality int      // 0-100, default 50
    Speed int        // 0-10, default 6
    MinQuantizer int // 0-63
    MaxQuantizer int // 0-63
    BitDepth int     // 8, 10, or 12
    // ... その他のオプション
}

func NewDefaultOptions() Options

type Command struct { /* 内部実装 */ }

func NewCommand(opts Options) (*Command, error)
func (c *Command) Run(imageData []byte) ([]byte, error)          // コアメソッド
func (c *Command) RunFile(inputPath, outputPath string) error    // シュガーシンタックス
func (c *Command) RunIO(input io.Reader, output io.Writer) error // シュガーシンタックス
func (c *Command) Close() error
```

### avifdec パッケージ（AVIFデコーダー）

```go
package avifdec

type Options struct {
    Format OutputFormat
    UseThreads bool
    // ... その他のオプション
}

func NewDefaultOptions() Options

type Command struct { /* 内部実装 */ }

func NewCommand(opts Options) (*Command, error)
func (c *Command) Run(avifData []byte) ([]byte, error)           // コアメソッド
func (c *Command) RunFile(inputPath, outputPath string) error    // シュガーシンタックス
func (c *Command) RunIO(input io.Reader, output io.Writer) error // シュガーシンタックス
func (c *Command) Close() error
```

### gif2webp パッケージ（GIF → WebP変換）

```go
package gif2webp

type Options struct {
    Quality float32
    Method int
    // WebPエンコード関連のオプション
}

func NewDefaultOptions() Options

type Command struct { /* 内部実装 */ }

func NewCommand(opts Options) (*Command, error)
func (c *Command) Run(gifData []byte) ([]byte, error)            // コアメソッド
func (c *Command) RunFile(inputPath, outputPath string) error    // シュガーシンタックス
func (c *Command) RunIO(input io.Reader, output io.Writer) error // シュガーシンタックス
func (c *Command) Close() error
```

### webp2gif パッケージ（WebP → GIF変換）

```go
package webp2gif

type Options struct {
    // GIFエンコード関連のオプション
}

func NewDefaultOptions() Options

type Command struct { /* 内部実装 */ }

func NewCommand(opts Options) (*Command, error)
func (c *Command) Run(webpData []byte) ([]byte, error)           // コアメソッド
func (c *Command) RunFile(inputPath, outputPath string) error    // シュガーシンタックス
func (c *Command) RunIO(input io.Reader, output io.Writer) error // シュガーシンタックス
func (c *Command) Close() error
```

### API設計の原則

1. **コマンドの再利用**: 同じ設定で複数ファイルを処理する際の初期化オーバーヘッド削減
2. **コアはバイト列**: `Run()`メソッドが基本、画像ファイル形式のバイト列を入出力
3. **シュガーシンタックス**: `RunFile()`と`RunIO()`は便利な薄いラッパー
4. **画像フォーマットの自動判定**: JPEG/PNGなどは内部で自動判定
5. **エラーハンドリング**: すべての関数/メソッドがerrorを返す
6. **リソース管理**: C言語は`*_free_*()`、Go言語は`Close()`で解放
7. **コマンド名との一致**: パッケージ名、関数プレフィックス、型名をコマンド名に合わせる

