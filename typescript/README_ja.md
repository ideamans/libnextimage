# @ideamans/libnextimage

Node.js向けの高性能WebP・AVIF画像処理ライブラリ、TypeScriptサポート付き。

## 機能

- ✅ **WebPエンコード/デコード** - すべてのエンコードオプションを備えた完全なWebPサポート
- ✅ **AVIFエンコード/デコード** - 次世代AVIFフォーマットサポート
- ✅ **GIF変換** - GIFとWebPフォーマット間の変換（アニメーションGIF含む）
- ✅ **ネイティブコンパイル不要** - プリビルドバイナリを自動ダウンロード
- ✅ **TypeScriptネイティブ** - 完全な型定義を含む
- ✅ **クロスプラットフォーム** - macOS (ARM64/Intel)、Linux (ARM64/x64)、Windows (x64)
- ✅ **高性能** - 最小限のオーバーヘッドで直接FFIバインディング
- ✅ **本番環境対応** - 自動リソースクリーンアップによるメモリセーフ

## インストール

```bash
npm install @ideamans/libnextimage
```

インストール時にお使いのプラットフォームに適したプリビルドネイティブライブラリが自動的にダウンロードされます。コンパイル不要！

### サポートプラットフォーム

- macOS (Apple Silicon M1/M2/M3 および Intel)
- Linux (ARM64 および x64)
- Windows (x64)

## クイックスタート

### WebPエンコード

```typescript
import { WebPEncoder } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

// オプションを指定してエンコーダーを作成
const encoder = new WebPEncoder({
  quality: 80,
  method: 6
})

// JPEGをWebPに変換
const jpegData = readFileSync('input.jpg')
const webpData = encoder.encode(jpegData)
writeFileSync('output.webp', webpData)

// リソースをクリーンアップ
encoder.close()

console.log(`変換完了: ${jpegData.length} バイト → ${webpData.length} バイト`)
```

### AVIFエンコード

```typescript
import { AVIFEncoder } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

// オプションを指定してエンコーダーを作成
const encoder = new AVIFEncoder({
  quality: 60,
  speed: 6
})

// JPEGをAVIFに変換
const jpegData = readFileSync('input.jpg')
const avifData = encoder.encode(jpegData)
writeFileSync('output.avif', avifData)

// リソースをクリーンアップ
encoder.close()

console.log(`変換完了: ${jpegData.length} バイト → ${avifData.length} バイト`)
```

### WebPデコード

```typescript
import { WebPDecoder } from '@ideamans/libnextimage'
import { readFileSync } from 'fs'

const decoder = new WebPDecoder({
  format: 'RGBA'
})

const webpData = readFileSync('input.webp')
const decoded = decoder.decode(webpData)

console.log(`デコード完了: ${decoded.width}x${decoded.height}、${decoded.data.length} バイト`)

decoder.close()
```

### AVIFデコード

```typescript
import { AVIFDecoder } from '@ideamans/libnextimage'
import { readFileSync } from 'fs'

const decoder = new AVIFDecoder({
  format: 'RGBA'
})

const avifData = readFileSync('input.avif')
const decoded = decoder.decode(avifData)

console.log(`デコード完了: ${decoded.width}x${decoded.height}、${decoded.data.length} バイト`)

decoder.close()
```

### GIFからWebPへの変換

```typescript
import { GIF2WebPConverter } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

const converter = new GIF2WebPConverter({
  quality: 80,
  method: 6  // アニメーション用の高品質設定
})

// アニメーションGIFをアニメーションWebPに変換
const gifData = readFileSync('animated.gif')
const webpData = converter.convert(gifData)

writeFileSync('animated.webp', webpData)
converter.close()

console.log(`GIF: ${gifData.length} バイト → WebP: ${webpData.length} バイト`)
```

### WebPからGIFへの変換

```typescript
import { WebP2GIFConverter } from '@ideamans/libnextimage'
import { readFileSync, writeFileSync } from 'fs'

const converter = new WebP2GIFConverter()

const webpData = readFileSync('image.webp')
const gifData = converter.convert(webpData)

writeFileSync('output.gif', gifData)
converter.close()

console.log(`WebP: ${webpData.length} バイト → GIF: ${gifData.length} バイト`)
```

## APIリファレンス

### WebPEncoder

画像をWebPフォーマットにエンコードします。

#### コンストラクタオプション

```typescript
interface WebPEncodeOptions {
  quality?: number          // 0-100、デフォルト: 75
  lossless?: boolean        // デフォルト: false
  method?: number           // 0-6、デフォルト: 4（品質/速度のトレードオフ）
  preset?: WebPPreset       // 'default', 'picture', 'photo', 'drawing', 'icon', 'text'

  // 高度なオプション
  targetSize?: number       // ターゲットファイルサイズ（バイト）
  targetPSNR?: number       // ターゲットPSNR
  segments?: number         // 1-4、セグメント数
  snsStrength?: number      // 0-100、空間ノイズシェーピング
  filterStrength?: number   // 0-100、フィルター強度
  autofilter?: boolean      // フィルター設定の自動調整

  // アルファチャンネル
  alphaQuality?: number     // 0-100、アルファ圧縮品質

  // メタデータ
  keepMetadata?: number     // MetadataEXIF | MetadataICC | MetadataXMP | MetadataAll

  // 変換
  cropX?: number
  cropY?: number
  cropWidth?: number
  cropHeight?: number
  resizeWidth?: number
  resizeHeight?: number
}
```

#### メソッド

```typescript
class WebPEncoder {
  constructor(options?: Partial<WebPEncodeOptions>)

  encode(data: Buffer): Buffer
  encodeFile(path: string): Buffer

  close(): void

  static getDefaultOptions(): WebPEncodeOptions
}
```

### AVIFEncoder

画像をAVIFフォーマットにエンコードします。

#### コンストラクタオプション

```typescript
interface AVIFEncodeOptions {
  quality?: number          // 0-100、デフォルト: 60
  qualityAlpha?: number     // 0-100、デフォルト: -1（qualityを使用）
  speed?: number            // 0-10、デフォルト: 6（0=最遅/最良、10=最速/最悪）

  bitDepth?: number         // 8、10、または12（デフォルト: 8）
  yuvFormat?: AVIFYUVFormat // 'YUV444', 'YUV422', 'YUV420', 'YUV400'

  // 高度なオプション
  lossless?: boolean
  sharpYUV?: boolean
  targetSize?: number

  // スレッディング
  jobs?: number             // -1=全コア、0=自動、>0=スレッド数

  // タイリング
  tileRowsLog2?: number     // 0-6
  tileColsLog2?: number     // 0-6
  autoTiling?: boolean

  // メタデータ
  exifData?: Buffer
  xmpData?: Buffer
  iccData?: Buffer
}
```

#### メソッド

```typescript
class AVIFEncoder {
  constructor(options?: Partial<AVIFEncodeOptions>)

  encode(data: Buffer): Buffer
  encodeFile(path: string): Buffer

  close(): void

  static getDefaultOptions(): AVIFEncodeOptions
}
```

### WebPDecoder

WebP画像を生のRGBAデータにデコードします。

```typescript
interface WebPDecodeOptions {
  format?: PixelFormat      // 'RGBA', 'BGRA', 'RGB', 'BGR'
  useThreads?: boolean
  bypassFiltering?: boolean
  noFancyUpsampling?: boolean

  cropX?: number
  cropY?: number
  cropWidth?: number
  cropHeight?: number

  scaleWidth?: number
  scaleHeight?: number
}

class WebPDecoder {
  constructor(options?: Partial<WebPDecodeOptions>)

  decode(data: Buffer): DecodedImage
  decodeFile(path: string): DecodedImage

  close(): void

  static getDefaultOptions(): WebPDecodeOptions
}

interface DecodedImage {
  width: number
  height: number
  data: Buffer              // YUVフォーマットの場合はYプレーン、RGBフォーマットの場合は全データ
  format: PixelFormat
  stride: number            // 1行あたりのバイト数（YUVの場合はYプレーンのストライド）
  bitDepth: number          // ビット深度（8、10、12）

  // YUVプレーンデータ（YUVフォーマットの場合のみ存在）
  uPlane?: Buffer           // U/Cbプレーンデータ
  vPlane?: Buffer           // V/Crプレーンデータ
  uStride?: number          // Uプレーンの1行あたりのバイト数
  vStride?: number          // Vプレーンの1行あたりのバイト数
}
```

### AVIFDecoder

AVIF画像を生のRGBAデータにデコードします。

```typescript
interface AVIFDecodeOptions {
  format?: PixelFormat      // 'RGBA', 'BGRA', 'RGB', 'BGR'
  jobs?: number             // -1=全コア、0=自動、>0=スレッド数

  chromaUpsampling?: ChromaUpsampling

  ignoreExif?: boolean
  ignoreXMP?: boolean
  ignoreICC?: boolean

  imageSizeLimit?: number
  imageDimensionLimit?: number
}

class AVIFDecoder {
  constructor(options?: Partial<AVIFDecodeOptions>)

  decode(data: Buffer): DecodedImage
  decodeFile(path: string): DecodedImage

  close(): void

  static getDefaultOptions(): AVIFDecodeOptions
}
```

### GIF2WebPConverter

GIF画像（アニメーションGIF含む）をWebPフォーマットに変換します。

#### コンストラクタオプション

```typescript
// GIF2WebPConverterはWebPEncoderと同じオプションを受け付けます
interface WebPEncodeOptions {
  quality?: number          // 0-100、デフォルト: 75
  lossless?: boolean        // デフォルト: false
  method?: number           // 0-6、デフォルト: 4
  // ... 上記のWebPEncoderオプションを参照
}
```

#### メソッド

```typescript
class GIF2WebPConverter {
  constructor(options?: Partial<WebPEncodeOptions>)

  convert(gifData: Buffer): Buffer  // GIFをWebPに変換（アニメーションを保持）

  close(): void
}
```

#### 例: アニメーションGIF変換

```typescript
import { GIF2WebPConverter } from '@ideamans/libnextimage'

const converter = new GIF2WebPConverter({
  quality: 80,
  method: 6  // アニメーション用の高品質
})

const gifData = readFileSync('animated.gif')
const webpData = converter.convert(gifData)

// アニメーションWebPはGIFよりもはるかに小さくなります
console.log(`圧縮率: ${((1 - webpData.length / gifData.length) * 100).toFixed(1)}%`)

writeFileSync('animated.webp', webpData)
converter.close()
```

### WebP2GIFConverter

WebP画像をGIFフォーマットに変換します。

#### コンストラクタオプション

```typescript
interface WebP2GIFOptions {
  reserved?: number  // 将来の使用のため予約
}
```

#### メソッド

```typescript
class WebP2GIFConverter {
  constructor(options?: WebP2GIFOptions)

  convert(webpData: Buffer): Buffer  // WebPをGIFに変換

  close(): void
}
```

#### 例: WebPからGIFへ

```typescript
import { WebP2GIFConverter } from '@ideamans/libnextimage'

const converter = new WebP2GIFConverter()

const webpData = readFileSync('image.webp')
const gifData = converter.convert(webpData)

writeFileSync('output.gif', gifData)
converter.close()
```

## バッチ処理の例

```typescript
import { WebPEncoder } from '@ideamans/libnextimage'
import { readdirSync, readFileSync, writeFileSync } from 'fs'
import { join } from 'path'

const encoder = new WebPEncoder({ quality: 80 })

const files = readdirSync('images')
  .filter(f => f.endsWith('.jpg') || f.endsWith('.png'))

for (const file of files) {
  const inputPath = join('images', file)
  const outputPath = join('output', file.replace(/\.(jpg|png)$/, '.webp'))

  const inputData = readFileSync(inputPath)
  const webpData = encoder.encode(inputData)
  writeFileSync(outputPath, webpData)

  console.log(`✓ ${file}: ${inputData.length} → ${webpData.length} バイト`)
}

encoder.close()
```

## メモリ管理

すべてのエンコーダー/デコーダー/コンバーターインスタンスは、ガベージコレクション時の自動クリーンアップにFinalizationRegistryを使用します。ただし、**決定的なリソース管理のため、明示的に `close()` を呼び出すことを強く推奨します**。

```typescript
// ベストプラクティス: try/finallyによる明示的クリーンアップ
const encoder = new WebPEncoder({ quality: 80 })
try {
  const result = encoder.encode(data)
  // ... 結果を使用
} finally {
  encoder.close()  // 明示的にリソースを解放
}

// 良い例: 複数ファイルでエンコーダーを再利用
const encoder = new WebPEncoder({ quality: 80 })
for (const file of files) {
  const result = encoder.encode(readFileSync(file))
  // ... 結果を処理
}
encoder.close()

// 自動クリーンアップ（本番環境では非推奨）
// リソースは最終的に解放されますが、タイミングは予測不可能です
const encoder = new WebPEncoder({ quality: 80 })
const result = encoder.encode(data)
// encoderは最終的にガベージコレクターによってクリーンアップされます
```

## バージョン管理

このパッケージはデュアルバージョンシステムを使用しています：

- **パッケージバージョン**（package.json内）: NPMパッケージバージョン
- **ネイティブライブラリバージョン**（library-version.json内）: プリビルドライブラリバージョン

これにより、ネイティブライブラリを再ビルドすることなく、TypeScriptコードのパッチリリースが可能になります。

```typescript
import { getLibraryVersion } from '@ideamans/libnextimage'

console.log(getLibraryVersion()) // 例: "0.4.0"
```

## トラブルシューティング

### "Cannot find libnextimage shared library"

インストール中にネイティブライブラリがダウンロードされませんでした。

**解決方法:**
```bash
npm install --force @ideamans/libnextimage
```

### "Unsupported platform"

お使いのプラットフォームはまだサポートされていません。サポートプラットフォーム:
- macOS (ARM64、x64)
- Linux (ARM64、x64)
- Windows (x64)

**解決方法:** ソースからビルド（メインリポジトリのREADMEを参照）

### メモリの問題

多数の画像を処理する場合は、以下を確認してください：
1. 新しいインスタンスを作成するのではなく、エンコーダー/デコーダーを再利用する
2. 終了時に `close()` を呼び出す
3. 必要に応じて画像をバッチで処理する

## 使用例

完全な動作例については、[examples/typescript/](../examples/typescript/) ディレクトリを参照してください：

- `jpeg-to-webp.ts` - JPEGからWebPへの変換
- `jpeg-to-avif.ts` - JPEGからAVIFへの変換
- `batch-convert.ts` - 進捗表示付きバッチ変換

## ランタイムサポート

現在サポート:
- ✅ **Node.js** 18+（完全サポート）

計画中:
- 🔄 **Deno**（近日公開）
- 🔄 **Bun**（近日公開）

## ライセンス

BSD-3-Clause

## リンク

- [GitHubリポジトリ](https://github.com/ideamans/libnextimage)
- [使用例](../examples/typescript/)
- [バージョン管理ガイド](./VERSION-MANAGEMENT.md)
- [イシュートラッカー](https://github.com/ideamans/libnextimage/issues)

## コントリビューション

コントリビューションを歓迎します！コントリビューションガイドラインについては、メインリポジトリを参照してください。
