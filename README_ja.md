# libnextimage

WebPとAVIFの高性能画像エンコード/デコードライブラリ - Go、TypeScript/Node.js、Bun、Denoに対応

## 概要

`libnextimage`は、libwebpとlibavifへの統一されたインターフェースを提供し、公式CLIツール（`cwebp`、`dwebp`、`avifenc`、`avifdec`）と同じ動作をしながら、複数の言語から便利にアクセスできるようにします。

## 主な機能

- **WebPとAVIFのサポート**: 完全なエンコード/デコード機能
- **複数言語対応**: Go、TypeScript/Node.js、Bun、Deno
- **CLIツールとの互換性**: 公式ツールと同一の出力を生成
- **クロスプラットフォーム**: macOS（Intel/ARM）、Linux（x64/ARM64）、Windows
- **簡単なインストール**: パッケージマネージャーによる自動ライブラリダウンロード
- **依存関係なし**: 必要なライブラリをすべてバンドル

## インストール

### Go

```bash
go get github.com/ideamans/libnextimage/golang
```

ライブラリは初回実行時に自動的にダウンロードされます。明示的にダウンロードすることもできます：

```go
import "github.com/ideamans/libnextimage/golang"

func main() {
    // ライブラリを確保（必要に応じてダウンロード）
    if err := libnextimage.EnsureLibrary(); err != nil {
        panic(err)
    }
    // ... 使用する
}
```

**環境変数:**
- `LIBNEXTIMAGE_CACHE_DIR`: カスタムキャッシュディレクトリ（デフォルト: `~/.cache/libnextimage`）
- `XDG_CACHE_HOME`: 標準XDGキャッシュディレクトリ

### TypeScript/Node.js

npmでインストール - ネイティブライブラリはインストール時に自動ダウンロードされます：

```bash
npm install @ideamans/libnextimage
```

パッケージは自動的にプラットフォーム（macOS、Linux、Windows）に適したネイティブライブラリをGitHub Releasesからダウンロードします。

**クイックスタート:**

```typescript
import { WebPEncoder, AVIFEncoder } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

// PNGをWebPにエンコード
const pngData = readFileSync('input.png')
const webpEncoder = new WebPEncoder({ quality: 90 })
const webpData = webpEncoder.encode(pngData)
webpEncoder.close()
writeFileSync('output.webp', webpData)

// AVIFにエンコード
const avifEncoder = new AVIFEncoder({ quality: 60, speed: 6 })
const avifData = avifEncoder.encode(pngData)
avifEncoder.close()
writeFileSync('output.avif', avifData)
```

詳細は[typescript/README.md](typescript/README.md)をご覧ください。

### C/C++

[GitHub Releases](https://github.com/ideamans/libnextimage/releases)からビルド済みバイナリをダウンロードするか、インストールスクリプトを使用：

```bash
# リポジトリをクローン
git clone https://github.com/ideamans/libnextimage.git
cd libnextimage

# プラットフォームに合わせてインストール（GitHub Releasesからダウンロード）
bash scripts/install.sh

# または特定バージョンをインストール
bash scripts/install.sh v0.4.0
```

インストールされるもの:
- `lib/<platform>/libnextimage.a` - 静的ライブラリ（C/C++とGo用）
- `include/*.h` - ヘッダーファイル

詳細は[ソースからのビルド](#ソースからのビルド)セクションをご覧ください。

## 使用例

### Go

```go
import (
    "os"
    "github.com/ideamans/libnextimage/golang"
)

// ライブラリが利用可能か確認（必要に応じてダウンロード）
if err := libnextimage.EnsureLibrary(); err != nil {
    panic(err)
}

// PNGをWebPにエンコード
data, _ := os.ReadFile("input.png")
opts := libnextimage.DefaultWebPEncodeOptions()
opts.Quality = 90

webpData, _ := libnextimage.WebPEncode(data, opts)
os.WriteFile("output.webp", webpData, 0644)

// AVIFにエンコード
avifOpts := libnextimage.DefaultAVIFEncodeOptions()
avifOpts.Quality = 60
avifOpts.Speed = 6

avifData, _ := libnextimage.AVIFEncode(data, avifOpts)
os.WriteFile("output.avif", avifData, 0644)
```

### TypeScript/Node.js

```typescript
import { WebPEncoder, AVIFEncoder } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

const imageData = readFileSync('input.png')

// WebPエンコード
const webpEncoder = new WebPEncoder({ quality: 90 })
const webpData = webpEncoder.encode(imageData)
webpEncoder.close()
writeFileSync('output.webp', webpData)

// AVIFエンコード
const avifEncoder = new AVIFEncoder({ quality: 60, speed: 6 })
const avifData = avifEncoder.encode(imageData)
avifEncoder.close()
writeFileSync('output.avif', avifData)
```

### C

```c
#include "nextimage.h"
#include "webp.h"

NextImageWebPEncodeOptions opts;
nextimage_webp_default_encode_options(&opts);
opts.quality = 90;

NextImageBuffer input, output;
// ... 入力データを読み込み ...

NextImageStatus status = nextimage_webp_encode(
    input.data, input.size, &opts, &output
);

if (status == NEXTIMAGE_STATUS_OK) {
    // output.data, output.sizeを使用
    nextimage_buffer_free(&output);
}
```

## サポートプラットフォーム

- **macOS**: Intel（x64）、Apple Silicon（ARM64）
- **Linux**: x64、ARM64
- **Windows**: x64

すべてのプラットフォーム向けのビルド済みバイナリが[GitHub Releases](https://github.com/ideamans/libnextimage/releases)で提供されています。

## ソースからのビルド

### 必要なもの

- CMake 3.15以降
- C11対応コンパイラ（GCC、Clang、MSVC）
- システムライブラリ: libjpeg、libpng、libgif

### macOS

```bash
brew install cmake jpeg libpng giflib
bash scripts/build-c-library.sh
```

### Linux

```bash
sudo apt-get install cmake build-essential libjpeg-dev libpng-dev libgif-dev
bash scripts/build-c-library.sh
```

### ビルド出力

ビルドスクリプトは以下を生成します：
- `lib/<platform>/libnextimage.a` - 統合静的ライブラリ
- `include/*.h` - ヘッダーファイル

## テスト

### Go

```bash
cd golang
go test -v
```

### TypeScript/Node.js

```bash
cd typescript
npm install
npm test
```

すべてのテストは公式CLIツールとバイト単位で完全一致することを検証します。

## CLIツール互換性

このライブラリは公式CLIツールと**バイト単位で完全一致**する出力を生成します：

- ✅ `cwebp` / `dwebp` - WebPエンコード/デコード
- ✅ `avifenc` / `avifdec` - AVIFエンコード/デコード
- ✅ `gif2webp` / `webp2gif` - GIF変換

## ライセンス

このプロジェクトはBSD 3-Clause Licenseでライセンスされています。

- libwebp: BSD License
- libavif: BSD License
- libaom: BSD License

## ドキュメント

- [TypeScript/Node.jsドキュメント](typescript/README.md)
- [サンプル](examples/)
- [英語版README](README.md)

## コントリビューション

コントリビューションを歓迎します！プルリクエストを送る前にすべてのテストが通ることを確認してください。

## サポート

- Issues: https://github.com/ideamans/libnextimage/issues
- Releases: https://github.com/ideamans/libnextimage/releases

## クレジット

開発: [株式会社アイデアマンズ](https://www.ideamans.com/)
