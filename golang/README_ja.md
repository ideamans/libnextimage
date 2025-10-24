# libnextimage - Go バインディング

Go向けの高性能WebP、AVIF、GIF画像処理ライブラリ。

## 機能

- **WebPエンコード・デコード**: 品質とエンコードオプションを完全にコントロールできる高速WebP画像処理
- **AVIFエンコード・デコード**: 品質と速度プリセットを備えた最新のAVIFフォーマットサポート
- **GIF変換**: アニメーションGIFを含むGIFとWebPフォーマット間の変換
- **CGOベースのパフォーマンス**: 最高のパフォーマンスを実現する直接Cライブラリバインディング
- **自動リソース管理**: `runtime.SetFinalizer`による自動クリーンアップ
- **慣用的なGo API**: エラーを返すクリーンなGoインターフェース
- **クロスプラットフォーム**: macOS (ARM64/x64)、Linux (x64/ARM64)、Windows (x64)をサポート

## インストール

```bash
go get github.com/ideamans/libnextimage/golang
```

## クイックスタート

### WebPエンコード

```go
package main

import (
    "fmt"
    "os"

    "github.com/ideamans/libnextimage/golang/webp"
)

func main() {
    // 入力画像を読み込む（JPEG、PNGなど）
    inputData, err := os.ReadFile("input.jpg")
    if err != nil {
        panic(err)
    }

    // オプション付きでエンコーダーを作成
    opts := webp.NewEncoderOptions()
    opts.Quality = 80
    opts.Method = 6

    encoder, err := webp.NewEncoder(opts)
    if err != nil {
        panic(err)
    }
    defer encoder.Close()

    // WebPにエンコード
    webpData, err := encoder.Encode(inputData)
    if err != nil {
        panic(err)
    }

    // 出力を保存
    if err := os.WriteFile("output.webp", webpData, 0644); err != nil {
        panic(err)
    }

    fmt.Printf("変換完了: %d バイト → %d バイト\n", len(inputData), len(webpData))
}
```

### WebPデコード

```go
package main

import (
    "fmt"
    "os"

    "github.com/ideamans/libnextimage/golang/webp"
)

func main() {
    // WebPファイルを読み込む
    webpData, err := os.ReadFile("image.webp")
    if err != nil {
        panic(err)
    }

    // デコーダーを作成
    decoder, err := webp.NewDecoder()
    if err != nil {
        panic(err)
    }
    defer decoder.Close()

    // RGBAにデコード
    decoded, err := decoder.Decode(webpData)
    if err != nil {
        panic(err)
    }

    fmt.Printf("幅: %d、高さ: %d\n", decoded.Width, decoded.Height)
    fmt.Printf("データサイズ: %d バイト\n", len(decoded.Data))
}
```

### AVIFエンコード

```go
package main

import (
    "os"

    "github.com/ideamans/libnextimage/golang/avif"
)

func main() {
    inputData, err := os.ReadFile("input.jpg")
    if err != nil {
        panic(err)
    }

    // オプション付きでエンコーダーを作成
    opts := avif.NewEncoderOptions()
    opts.Quality = 60
    opts.Speed = 6

    encoder, err := avif.NewEncoder(opts)
    if err != nil {
        panic(err)
    }
    defer encoder.Close()

    // AVIFにエンコード
    avifData, err := encoder.Encode(inputData)
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("output.avif", avifData, 0644); err != nil {
        panic(err)
    }
}
```

### AVIFデコード

```go
package main

import (
    "fmt"
    "os"

    "github.com/ideamans/libnextimage/golang/avif"
)

func main() {
    avifData, err := os.ReadFile("image.avif")
    if err != nil {
        panic(err)
    }

    decoder, err := avif.NewDecoder()
    if err != nil {
        panic(err)
    }
    defer decoder.Close()

    decoded, err := decoder.Decode(avifData)
    if err != nil {
        panic(err)
    }

    fmt.Printf("%dx%d AVIF画像をデコード完了\n", decoded.Width, decoded.Height)
}
```

### GIFからWebPへの変換

```go
package main

import (
    "fmt"
    "os"

    "github.com/ideamans/libnextimage/golang/gif2webp"
)

func main() {
    gifData, err := os.ReadFile("animated.gif")
    if err != nil {
        panic(err)
    }

    // オプション付きでコンバーターを作成
    opts := gif2webp.NewOptions()
    opts.Quality = 80
    opts.Method = 6

    cmd, err := gif2webp.NewCommand(opts)
    if err != nil {
        panic(err)
    }
    defer cmd.Close()

    // GIFをWebPに変換（アニメーションを保持）
    webpData, err := cmd.Run(gifData)
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("animated.webp", webpData, 0644); err != nil {
        panic(err)
    }

    compression := (1.0 - float64(len(webpData))/float64(len(gifData))) * 100
    fmt.Printf("圧縮率: %.1f%%\n", compression)
}
```

### WebPからGIFへの変換

```go
package main

import (
    "os"

    "github.com/ideamans/libnextimage/golang/webp2gif"
)

func main() {
    webpData, err := os.ReadFile("image.webp")
    if err != nil {
        panic(err)
    }

    cmd, err := webp2gif.NewCommand(nil)
    if err != nil {
        panic(err)
    }
    defer cmd.Close()

    gifData, err := cmd.Run(webpData)
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("output.gif", gifData, 0644); err != nil {
        panic(err)
    }
}
```

## APIリファレンス

### WebPパッケージ

#### Encoder

```go
package webp

type EncoderOptions struct {
    Quality         float32
    Lossless        bool
    Method          int
    Preset          Preset
    ImageHint       ImageHint
    TargetSize      int
    TargetPSNR      float32
    Segments        int
    SNSStrength     int
    FilterStrength  int
    FilterSharpness int
    FilterType      int
    Autofilter      bool
    AlphaQuality    int
    Pass            int
    Exact           bool
    // ... 他にも多数のオプション
}

func NewEncoderOptions() *EncoderOptions
func NewEncoder(opts *EncoderOptions) (*Encoder, error)
func (e *Encoder) Encode(data []byte) ([]byte, error)
func (e *Encoder) Close() error
```

**主なオプション:**
- `Quality` (0-100): 品質レベル、デフォルト75
- `Lossless` (bool): ロスレスエンコードを使用
- `Method` (0-6): 圧縮方法、大きいほど遅いが高品質
- `Preset`: 事前定義された設定（Default、Picture、Photo、Drawing、Icon、Text）

#### Decoder

```go
type DecoderOptions struct {
    Format            PixelFormat
    UseThreads        bool
    BypassFiltering   bool
    NoFancyUpsampling bool
    CropX, CropY      int
    CropWidth         int
    CropHeight        int
    ScaleWidth        int
    ScaleHeight       int
}

type DecodedImage struct {
    Width  int
    Height int
    Data   []byte
    Format PixelFormat
}

func NewDecoderOptions() *DecoderOptions
func NewDecoder(opts *DecoderOptions) (*Decoder, error)
func (d *Decoder) Decode(data []byte) (*DecodedImage, error)
func (d *Decoder) Close() error
```

### AVIFパッケージ

#### Encoder

```go
package avif

type EncoderOptions struct {
    Quality       int
    QualityAlpha  int
    Speed         int
    BitDepth      int
    YUVFormat     YUVFormat
    Lossless      bool
    Jobs          int
    AutoTiling    bool
    // ... さらに多くのオプション
}

func NewEncoderOptions() *EncoderOptions
func NewEncoder(opts *EncoderOptions) (*Encoder, error)
func (e *Encoder) Encode(data []byte) ([]byte, error)
func (e *Encoder) Close() error
```

**主なオプション:**
- `Quality` (0-100): 品質レベル、デフォルト60
- `Speed` (0-10): エンコード速度、大きいほど速い（0=最遅/最良、10=最速/最悪）
- `BitDepth` (8/10/12): チャンネルあたりのビット深度
- `YUVFormat`: 色フォーマット（YUV444、YUV422、YUV420、YUV400）

#### Decoder

```go
type DecoderOptions struct {
    Format              PixelFormat
    Jobs                int
    ChromaUpsampling    ChromaUpsampling
    IgnoreExif          bool
    IgnoreXMP           bool
    IgnoreICC           bool
    ImageSizeLimit      int
    ImageDimensionLimit int
}

func NewDecoderOptions() *DecoderOptions
func NewDecoder(opts *DecoderOptions) (*Decoder, error)
func (d *Decoder) Decode(data []byte) (*DecodedImage, error)
func (d *Decoder) Close() error
```

### GIF2WebPパッケージ

```go
package gif2webp

type Options struct {
    // WebPエンコーダーオプションと同じ
    Quality  float32
    Lossless bool
    Method   int
    // ... など
}

type Command struct {
    // 内部実装
}

func NewOptions() *Options
func NewCommand(opts *Options) (*Command, error)
func (c *Command) Run(gifData []byte) ([]byte, error)
func (c *Command) Close() error
```

**例: バッチGIF変換**

```go
func convertGIFs(files []string) error {
    opts := gif2webp.NewOptions()
    opts.Quality = 80
    opts.Method = 6

    cmd, err := gif2webp.NewCommand(opts)
    if err != nil {
        return err
    }
    defer cmd.Close()

    for _, file := range files {
        gifData, err := os.ReadFile(file)
        if err != nil {
            return err
        }

        webpData, err := cmd.Run(gifData)
        if err != nil {
            log.Printf("変換失敗 %s: %v", file, err)
            continue
        }

        outFile := strings.TrimSuffix(file, ".gif") + ".webp"
        if err := os.WriteFile(outFile, webpData, 0644); err != nil {
            return err
        }

        fmt.Printf("✓ %s: %d → %d バイト\n", file, len(gifData), len(webpData))
    }

    return nil
}
```

### WebP2GIFパッケージ

```go
package webp2gif

type Options struct {
    Reserved int  // 将来の使用のため予約
}

type Command struct {
    // 内部実装
}

func NewOptions() *Options
func NewCommand(opts *Options) (*Command, error)
func (c *Command) Run(webpData []byte) ([]byte, error)
func (c *Command) Close() error
```

## 高度な使い方

### エンコーダー/デコーダーインスタンスの再利用

複数の画像を処理する際のパフォーマンス向上のため、インスタンスを再利用してください：

```go
func convertBatch(files []string) error {
    encoder, err := webp.NewEncoder(&webp.EncoderOptions{
        Quality: 80,
        Method:  4,
    })
    if err != nil {
        return err
    }
    defer encoder.Close()

    for _, file := range files {
        inputData, err := os.ReadFile(file)
        if err != nil {
            log.Printf("読み込み失敗 %s: %v", file, err)
            continue
        }

        webpData, err := encoder.Encode(inputData)
        if err != nil {
            log.Printf("エンコード失敗 %s: %v", file, err)
            continue
        }

        outFile := strings.TrimSuffix(file, filepath.Ext(file)) + ".webp"
        if err := os.WriteFile(outFile, webpData, 0644); err != nil {
            log.Printf("書き込み失敗 %s: %v", outFile, err)
            continue
        }

        fmt.Printf("✓ %s\n", file)
    }

    return nil
}
```

### 品質とファイルサイズのトレードオフ

```go
// 高品質（ファイルサイズ大）
hqOpts := webp.NewEncoderOptions()
hqOpts.Quality = 95
hqOpts.Method = 6    // 最も遅いが最良の圧縮
hqOpts.Pass = 10     // より多くの解析パス

// バランス型（ほとんどのユースケースに適している）
balancedOpts := webp.NewEncoderOptions()
balancedOpts.Quality = 80
balancedOpts.Method = 4

// 小ファイル（品質低）
smallOpts := webp.NewEncoderOptions()
smallOpts.Quality = 60
smallOpts.Method = 2
smallOpts.Preprocessing = 2
```

### ロスレスエンコード

```go
opts := webp.NewEncoderOptions()
opts.Lossless = true
opts.Quality = 100
opts.Exact = true  // 正確なRGB値を保持

encoder, err := webp.NewEncoder(opts)
if err != nil {
    return err
}
defer encoder.Close()

// 完璧な品質、グラフィックス/図表に適している
webpData, err := encoder.Encode(inputData)
```

### エラーハンドリング

```go
import (
    "errors"
    "github.com/ideamans/libnextimage/golang/webp"
    "github.com/ideamans/libnextimage/golang/types"
)

func encodeImage(data []byte) ([]byte, error) {
    encoder, err := webp.NewEncoder(&webp.EncoderOptions{
        Quality: 80,
    })
    if err != nil {
        return nil, fmt.Errorf("エンコーダー作成失敗: %w", err)
    }
    defer encoder.Close()

    webpData, err := encoder.Encode(data)
    if err != nil {
        var nextErr *types.NextImageError
        if errors.As(err, &nextErr) {
            switch nextErr.Status {
            case types.StatusInvalidParameter:
                return nil, fmt.Errorf("無効なエンコードパラメータ: %w", err)
            case types.StatusOutOfMemory:
                return nil, fmt.Errorf("メモリ不足: %w", err)
            case types.StatusEncodeFailed:
                return nil, fmt.Errorf("エンコード失敗: %w", err)
            default:
                return nil, fmt.Errorf("エンコードエラー: %w", err)
            }
        }
        return nil, err
    }

    return webpData, nil
}
```

### 並行処理

ゴルーチンを使用して複数の画像を並行処理：

```go
func convertConcurrent(files []string, workers int) error {
    jobs := make(chan string, len(files))
    results := make(chan error, len(files))

    // ワーカーを起動
    for w := 0; w < workers; w++ {
        go func() {
            encoder, err := webp.NewEncoder(&webp.EncoderOptions{
                Quality: 80,
            })
            if err != nil {
                results <- err
                return
            }
            defer encoder.Close()

            for file := range jobs {
                inputData, err := os.ReadFile(file)
                if err != nil {
                    results <- fmt.Errorf("%s: %w", file, err)
                    continue
                }

                webpData, err := encoder.Encode(inputData)
                if err != nil {
                    results <- fmt.Errorf("%s: %w", file, err)
                    continue
                }

                outFile := strings.TrimSuffix(file, filepath.Ext(file)) + ".webp"
                if err := os.WriteFile(outFile, webpData, 0644); err != nil {
                    results <- fmt.Errorf("%s: %w", outFile, err)
                    continue
                }

                results <- nil
                fmt.Printf("✓ %s\n", file)
            }
        }()
    }

    // ジョブを送信
    for _, file := range files {
        jobs <- file
    }
    close(jobs)

    // 結果を収集
    var errs []error
    for range files {
        if err := <-results; err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("変換中に %d 件のエラーが発生", len(errs))
    }

    return nil
}
```

### io.Readerからのストリーミング

```go
func encodeFromReader(r io.Reader) ([]byte, error) {
    // すべてのデータをメモリに読み込む
    inputData, err := io.ReadAll(r)
    if err != nil {
        return nil, fmt.Errorf("入力の読み込み失敗: %w", err)
    }

    encoder, err := webp.NewEncoder(&webp.EncoderOptions{
        Quality: 80,
    })
    if err != nil {
        return nil, err
    }
    defer encoder.Close()

    return encoder.Encode(inputData)
}

func encodeFile(inputPath, outputPath string) error {
    f, err := os.Open(inputPath)
    if err != nil {
        return err
    }
    defer f.Close()

    webpData, err := encodeFromReader(f)
    if err != nil {
        return err
    }

    return os.WriteFile(outputPath, webpData, 0644)
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

## パフォーマンスのヒント

1. **エンコーダー/デコーダーインスタンスを再利用する** - 新しいインスタンスの作成にはオーバーヘッドがあります
2. **適切な品質設定を選択する** - 高品質 = 大きなファイル + 遅いエンコード
3. **適切なmethod値を使用する** - より大きなmethod = より良い圧縮だが遅い
4. **必要な場合のみロスレスを検討する** - ロスレスははるかに大きなファイルを生成します
5. **バッチ処理する** - 複数の画像を処理する際はインスタンスを再利用
6. **並行処理にゴルーチンを使用する** - CPUコア間で作業を分散

## リソース管理

すべてのエンコーダー/デコーダー/コマンドインスタンスは、ガベージコレクション時の自動クリーンアップのために `runtime.SetFinalizer` を使用します。ただし、明示的に `Close()` を呼び出すのがベストプラクティスです：

```go
encoder, err := webp.NewEncoder(opts)
if err != nil {
    return err
}
defer encoder.Close()  // 明示的にリソースを解放

// encoderを使用...
```

## テスト

テストスイートを実行：

```bash
cd golang
go test ./...
```

レース検出器付きでテストを実行：

```bash
go test -race ./...
```

ベンチマークを実行：

```bash
go test -bench=. ./...
```

## 使用例

完全な動作例については `examples/golang/` ディレクトリを参照してください：
- `jpeg-to-webp/` - JPEGからWebPへの変換
- `jpeg-to-avif/` - JPEGからAVIFへの変換
- `batch-convert/` - 進捗表示付きバッチ変換
- `gif-conversion/` - GIFからWebPとWebPからGIFの例

## CGO要件

このパッケージはCGOを有効にする必要があります：

```bash
CGO_ENABLED=1 go build
```

ネイティブライブラリ（libnextimage）はシステムのライブラリパスで利用可能である必要があります。または、ライブラリパスを設定できます：

```bash
# macOS
export DYLD_LIBRARY_PATH=/path/to/libnextimage

# Linux
export LD_LIBRARY_PATH=/path/to/libnextimage

# またはCGOフラグを設定
export CGO_LDFLAGS="-L/path/to/libnextimage"
```

## ライセンス

BSD-3-Clause

## リンク

- GitHub: https://github.com/ideamans/libnextimage
- イシュー: https://github.com/ideamans/libnextimage/issues
- ドキュメント: https://pkg.go.dev/github.com/ideamans/libnextimage/golang

## コントリビューション

コントリビューションを歓迎します！コントリビューションガイドラインについては、メインリポジトリを参照してください。
