# TypeScript/JavaScriptバインディング設計書

## 概要

本ドキュメントでは、libnextimageのTypeScript/JavaScriptバインディングの設計について説明します。以下の2つの利用パターンに対応します:

1. **npmパッケージインストール** - `npm install`時にプリビルド済み共有ライブラリをダウンロード
2. **ローカル開発** - ローカルでビルドした共有ライブラリを使用して開発・テスト

バインディングは3つのJavaScriptランタイムをサポートします: **Node.js**、**Deno**、**Bun**

## ディレクトリ構造

```
libnextimage/
├── typescript/                 # TypeScriptバインディングソース
│   ├── src/
│   │   ├── index.ts           # メインエントリポイント
│   │   ├── library.ts         # ライブラリパス解決
│   │   ├── ffi.ts             # FFIバインディング（Node.js用Koffi）
│   │   ├── webp.ts            # WebP API
│   │   └── avif.ts            # AVIF API（将来対応）
│   ├── dist/                  # コンパイル済みJavaScript（tscで生成）
│   ├── lib/                   # ダウンロードされた共有ライブラリ（npm install）
│   │   ├── darwin-arm64/
│   │   │   └── libnextimage.dylib
│   │   ├── darwin-amd64/
│   │   │   └── libnextimage.dylib
│   │   ├── linux-arm64/
│   │   │   └── libnextimage.so
│   │   ├── linux-amd64/
│   │   │   └── libnextimage.so
│   │   └── windows-amd64/
│   │       └── libnextimage.dll
│   ├── examples/
│   ├── test/
│   ├── package.json
│   ├── tsconfig.json
│   └── README.md
│
├── lib/                       # ビルド出力（ローカル開発）
│   ├── shared/                # 共有ライブラリ（.so/.dylib/.dll）
│   │   └── libnextimage.{so,dylib,dll}
│   ├── static/                # 静的ライブラリ（.a）
│   │   └── libnextimage.a
│   ├── darwin-arm64/          # プラットフォーム固有ビルド
│   │   └── libnextimage.a
│   ├── linux-amd64/
│   └── include/               # ヘッダーファイル
│       ├── nextimage.h
│       └── nextimage_types.h
│
└── c/                         # Cライブラリソース
```

## 2つの利用パターン

### パターン1: npmパッケージインストール（エンドユーザー）

ユーザーがnpmからパッケージをインストールする場合:

```bash
npm install @ideamans/libnextimage
```

**インストールフロー:**

1. npmレジストリからパッケージがダウンロードされる
2. `postinstall`スクリプトが自動実行される
3. スクリプトがプラットフォームを検出（darwin-arm64、linux-amd64など）
4. GitHub Releasesから対応するプリビルド済み共有ライブラリをダウンロード
5. `typescript/lib/<platform>/`に展開
6. ライブラリが使用可能に - コンパイル不要

**ライブラリ解決:**
- 検索パス: `typescript/lib/<platform>/libnextimage.{dylib,so,dll}`
- このパスはインストール済みパッケージの場所からの相対パス

**使用例:**
```typescript
import { encodeWebP } from '@ideamans/libnextimage';

const jpegData = fs.readFileSync('input.jpg');
const webpData = encodeWebP(jpegData);
fs.writeFileSync('output.webp', webpData);
```

### パターン2: ローカル開発

libnextimageの開発、またはローカル依存として使用する場合:

```bash
# Cライブラリのビルド
make install-c

# TypeScriptバインディングのビルド
cd typescript
npm install
npm run build

# 他のプロジェクトで使用
cd ../my-project
npm link ../libnextimage/typescript
```

**ビルドフロー:**

1. `make install-c`でCライブラリをコンパイル
2. 共有ライブラリが`lib/shared/libnextimage.{dylib,so,dll}`に出力される
3. TypeScriptコードが開発モードを検出
4. ローカルビルドのライブラリ（`../lib/shared/`）を使用

**ライブラリ解決:**
- 検索パス1: `../lib/shared/libnextimage.{dylib,so,dll}`（開発モード・共有ビルド）
- 検索パス2: `../lib/<platform>/libnextimage.{dylib,so,dll}`（開発モード・プラットフォーム固有）
- 検索パス3: `typescript/lib/<platform>/libnextimage.{dylib,so,dll}`（インストール済み）

**使用例:**
```typescript
// ローカルにリンクされている場合、ローカルビルドのライブラリを使用
import { encodeWebP } from '@ideamans/libnextimage';

const jpegData = fs.readFileSync('input.jpg');
const webpData = encodeWebP(jpegData);
```

## ランタイムサポート

### Node.js（主要）

- **FFIライブラリ:** [Koffi](https://github.com/Koromix/koffi)（推奨）
  - ネイティブパフォーマンス、最小限のオーバーヘッド
  - Bufferを直接扱える
  - クロスプラットフォーム（macOS、Linux、Windows）

- **モジュールシステム:** CommonJS（デフォルト）
  - 出力: `dist/index.js`（CommonJS）
  - package.jsonで`"type": "commonjs"`
  - TypeScriptターゲット: ES2020

**現在の実装状況:**
- ✅ ライブラリパス解決（`src/library.ts`）
- ✅ KoffiによるFFIバインディング（`src/ffi.ts`）
- ✅ WebPエンコードAPI（`src/webp.ts`）
- ✅ テストスイート（`test/webp-encode.test.ts`）

### Deno（将来対応）

DenoはネイティブFFIサポート（`Deno.dlopen()`）を持っています。

**ランタイム固有の課題:**
- `__dirname`, `require()`が使用不可
- `import.meta.url`ベースのパス解決が必要
- npm specifier (`npm:`) 経由でのインストール時、postinstallが実行されない

**実装計画:**

1. **ライブラリパス解決（Deno専用）:**
   ```typescript
   // deno/library.ts
   export function getLibraryPath(): string {
     const platform = getPlatform();
     const libFileName = getLibraryFileName();

     // import.meta.url を使用してパス解決
     const moduleDir = new URL('.', import.meta.url).pathname;

     // Deno向けバイナリ配布: deno.land/x/libnextimage/lib/<platform>/
     const libPath = new URL(`../lib/${platform}/${libFileName}`, import.meta.url).pathname;

     if (existsSync(libPath)) {
       return libPath;
     }

     throw new Error(`Cannot find library for ${platform}`);
   }
   ```

2. **FFIバインディング（Deno専用）:**
   ```typescript
   // deno/ffi.ts - Deno専用FFI
   const lib = Deno.dlopen(getLibraryPath(), {
     nextimage_webp_encode_alloc: {
       parameters: ["buffer", "usize", "pointer", "pointer"],
       result: "i32",
     },
     nextimage_free_buffer: {
       parameters: ["pointer"],
       result: "void",
     }
   });
   ```

3. **バイナリ配布:**
   - deno.land/x/libnextimage に公開
   - `lib/<platform>/` に全プラットフォームのバイナリを含める
   - GitHubリリースからダウンロードして同梱

**モジュールエクスポート:**
- エントリポイント: `deno/mod.ts`
- インポート: `import { WebPEncoder } from "https://deno.land/x/libnextimage@v0.4.0/deno/mod.ts"`
- または: `import { WebPEncoder } from "npm:@ideamans/libnextimage/deno"`

### Bun（将来対応）

BunはビルトインFFIサポート（`bun:ffi`）を持っています。

**ランタイム固有の課題:**
- Node.jsとの互換性は高いが、FFI APIが異なる
- `import.meta.url`と`__dirname`の両方が使用可能
- npm installは動作するが、FFI部分は書き換えが必要

**実装計画:**

1. **ライブラリパス解決（Bun専用）:**
   ```typescript
   // bun/library.ts
   // Node.jsと同様のロジックだが、Bun FFI用に調整
   export function getLibraryPath(): string {
     // Bunは__dirnameをサポート
     // Node.js版と同じパス解決ロジックを使用可能
     return findLibraryPath();
   }
   ```

2. **FFIバインディング（Bun専用）:**
   ```typescript
   // bun/ffi.ts - Bun専用FFI
   import { dlopen, FFIType, ptr } from "bun:ffi";

   const lib = dlopen(getLibraryPath(), {
     nextimage_webp_encode_alloc: {
       args: [FFIType.ptr, FFIType.u64, FFIType.ptr, FFIType.ptr],
       returns: FFIType.i32,
     },
     nextimage_free_buffer: {
       args: [FFIType.ptr],
       returns: FFIType.void,
     }
   });
   ```

3. **バイナリ配布:**
   - npm経由でのインストールをサポート（postinstall動作）
   - または、Bun専用パッケージとして公開

**モジュールエクスポート:**
- エントリポイント: `bun/mod.ts`
- インポート: `import { WebPEncoder } from "@ideamans/libnextimage/bun"`
- package.jsonのexportsフィールドで明示的にマッピング

**package.json exports設定例:**
```json
{
  "exports": {
    ".": {
      "node": "./dist/index.js",
      "bun": "./bun/mod.js",
      "deno": "./deno/mod.ts",
      "types": "./dist/index.d.ts"
    }
  }
}
```

## プラットフォーム検出

### サポートプラットフォーム

| プラットフォーム | アーキテクチャ | Node.js | Deno | Bun | ライブラリファイル |
|----------|-------------|---------|------|-----|--------------|
| macOS    | ARM64 (M1/M2/M3) | ✅ | 🔄 | 🔄 | libnextimage.dylib |
| macOS    | Intel (x64) | ✅ | 🔄 | 🔄 | libnextimage.dylib |
| Linux    | ARM64 (aarch64) | ✅ | 🔄 | 🔄 | libnextimage.so |
| Linux    | x64 (amd64) | ✅ | 🔄 | 🔄 | libnextimage.so |
| Windows  | x64 | ✅ | 🔄 | 🔄 | libnextimage.dll |

凡例: ✅ 実装済み | 🔄 計画中 | ❌ 非サポート

### プラットフォーム命名規則

```typescript
function getPlatform(): string {
  const platform = process.platform;  // 'darwin', 'linux', 'win32'
  const arch = process.arch;          // 'arm64', 'x64'

  // 返り値: 'darwin-arm64', 'linux-amd64', 'windows-amd64'など
}
```

**ディレクトリ命名:**
- `darwin-arm64`（macOS Apple Silicon）
- `darwin-amd64`（macOS Intel）
- `linux-arm64`（Linux ARM64）
- `linux-amd64`（Linux x64）
- `windows-amd64`（Windows x64）

## ライブラリパス解決戦略

`library.ts`モジュールは共有ライブラリを見つけるためのフォールバックチェーンを実装しています:

```typescript
export function findLibraryPath(): string {
  const platform = getPlatform();
  const libFileName = getLibraryFileName();

  // 優先順位:
  // 1. ローカル開発: ../../lib/shared/
  // __dirname = typescript/dist/ なので、2つ上がるとプロジェクトルート
  const devSharedPath = path.join(__dirname, '..', '..', 'lib', 'shared', libFileName);
  if (fs.existsSync(devSharedPath)) return devSharedPath;

  // 2. ローカル開発（プラットフォーム固有）: ../../lib/<platform>/
  const devPlatformPath = path.join(__dirname, '..', '..', 'lib', platform, libFileName);
  if (fs.existsSync(devPlatformPath)) return devPlatformPath;

  // 3. インストール済みパッケージ: ../lib/<platform>/
  // __dirname = node_modules/@ideamans/libnextimage/dist/ の場合
  const installedPath = path.join(__dirname, '..', 'lib', platform, libFileName);
  if (fs.existsSync(installedPath)) return installedPath;

  throw new Error(`${platform}用のlibnextimage共有ライブラリが見つかりません`);
}
```

**コンパイル済みコード（`dist/`）からのパス解決:**

TypeScriptがコンパイルされると、`dist/library.js`内の`__dirname`は以下を指します:
- ローカル開発: `<project-root>/typescript/dist/`
- インストール済み: `node_modules/@ideamans/libnextimage/dist/`

**パス計算（修正版）:**

| モード | `__dirname` | ターゲットパス | 相対パス |
|------|------------|---------------|---------|
| 開発（shared） | `typescript/dist/` | `lib/shared/libnextimage.dylib` | `../../lib/shared/` |
| 開発（platform） | `typescript/dist/` | `lib/darwin-arm64/libnextimage.dylib` | `../../lib/darwin-arm64/` |
| インストール済み | `node_modules/@ideamans/libnextimage/dist/` | `node_modules/@ideamans/libnextimage/lib/darwin-arm64/libnextimage.dylib` | `../lib/darwin-arm64/` |

**重要:** パス解決のテストを必ず実施すること（フェーズ1のタスクに含む）

## npmパッケージ構造

### package.json設定

```json
{
  "name": "@ideamans/libnextimage",
  "version": "0.4.0",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "files": [
    "dist",
    "lib",
    "README.md"
  ],
  "scripts": {
    "build": "tsc",
    "test": "npm run build && node --test dist/test/*.test.js",
    "postinstall": "node scripts/download-library.js",
    "prepublishOnly": "npm run build"
  },
  "dependencies": {
    "koffi": "^2.9.0"
  }
}
```

### postinstallスクリプト

`scripts/download-library.js`スクリプトは`npm install`後に実行されます:

```javascript
// scripts/download-library.js
const https = require('https');
const fs = require('fs');
const path = require('path');

const VERSION = '0.4.0';
const RELEASE_URL = `https://github.com/ideamans/libnextimage/releases/download/v${VERSION}`;

function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;

  if (platform === 'darwin') {
    return arch === 'arm64' ? 'darwin-arm64' : 'darwin-amd64';
  } else if (platform === 'linux') {
    return arch === 'arm64' ? 'linux-arm64' : 'linux-amd64';
  } else if (platform === 'win32') {
    return 'windows-amd64';
  }

  throw new Error(`サポートされていないプラットフォーム: ${platform}-${arch}`);
}

function getLibraryFileName(platformId) {
  if (platformId.startsWith('darwin')) return 'libnextimage.dylib';
  if (platformId.startsWith('linux')) return 'libnextimage.so';
  if (platformId.startsWith('windows')) return 'libnextimage.dll';
  throw new Error(`不明なプラットフォーム: ${platformId}`);
}

async function downloadLibrary() {
  const platformId = getPlatform();
  const libFileName = getLibraryFileName(platformId);

  // ダウンロードURL: https://github.com/ideamans/libnextimage/releases/download/v0.4.0/libnextimage-darwin-arm64.dylib
  const assetName = `libnextimage-${platformId}.${libFileName.split('.').pop()}`;
  const downloadUrl = `${RELEASE_URL}/${assetName}`;

  // 保存先: lib/<platform>/libnextimage.dylib
  const targetDir = path.join(__dirname, '..', 'lib', platformId);
  const targetPath = path.join(targetDir, libFileName);

  console.log(`GitHub releasesから${assetName}をダウンロード中...`);
  console.log(`  URL: ${downloadUrl}`);
  console.log(`  保存先: ${targetPath}`);

  // ディレクトリが存在しない場合は作成
  fs.mkdirSync(targetDir, { recursive: true });

  // ファイルをダウンロード
  await download(downloadUrl, targetPath);

  // Unix系システムで実行権限を設定
  if (process.platform !== 'win32') {
    fs.chmodSync(targetPath, 0o755);
  }

  console.log(`✓ ライブラリのインストールに成功しました`);
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https.get(url, (response) => {
      if (response.statusCode !== 200) {
        reject(new Error(`ダウンロード失敗: ${response.statusCode}`));
        return;
      }
      response.pipe(file);
      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest);
      reject(err);
    });
  });
}

// 直接呼び出された場合に実行
if (require.main === module) {
  downloadLibrary().catch((err) => {
    console.error('ライブラリのダウンロードに失敗しました:', err.message);
    console.error('ローカルでビルドが必要な場合があります: make install-c');
    process.exit(0); // npm installを失敗させない
  });
}
```

### GitHub Releaseアセット

各リリース（例: v0.4.0）に対して、以下のアセットをアップロードする必要があります:

```
libnextimage-darwin-arm64.dylib
libnextimage-darwin-amd64.dylib
libnextimage-linux-arm64.so
libnextimage-linux-amd64.so
libnextimage-windows-amd64.dll
```

**ビルドとアップロードのワークフロー:**

```bash
# 全プラットフォーム向けビルド（クロスコンパイル環境が必要）
make build-all-platforms

# GitHub releaseにアップロード
gh release create v0.4.0 \
  lib/darwin-arm64/libnextimage.dylib#libnextimage-darwin-arm64.dylib \
  lib/darwin-amd64/libnextimage.dylib#libnextimage-darwin-amd64.dylib \
  lib/linux-arm64/libnextimage.so#libnextimage-linux-arm64.so \
  lib/linux-amd64/libnextimage.so#libnextimage-linux-amd64.so \
  lib/windows-amd64/libnextimage.dll#libnextimage-windows-amd64.dll
```

## API設計

TypeScript APIは、Golang版と同様にエンコーダー/デコーダーのインスタンスベースの設計を採用します。
これにより、初期化オーバーヘッドを削減し、複数の画像ファイルに同じ設定を効率的に適用できます。

### WebP エンコーダー

```typescript
// エンコーダーを作成（設定を初期化）
const encoder = new WebPEncoder({
  quality: 80,
  method: 6,
  lossless: false,
  keepMetadata: MetadataAll
});

// 複数のファイルをエンコード（設定を再利用）
const webp1 = encoder.encode(jpegData1);
const webp2 = encoder.encode(jpegData2);
const webp3 = encoder.encode(pngData);

// リソースを解放
encoder.close();

// またはファイルから直接
const webpData = encoder.encodeFile('input.jpg');
```

### WebP デコーダー

```typescript
// デコーダーを作成（設定を初期化）
const decoder = new WebPDecoder({
  format: PixelFormat.RGBA,
  useThreads: true
});

// 複数のファイルをデコード（設定を再利用）
const decoded1 = decoder.decode(webpData1);
const decoded2 = decoder.decode(webpData2);

// リソースを解放
decoder.close();

// DecodedImageの構造
interface DecodedImage {
  width: number;
  height: number;
  data: Buffer;           // 生ピクセルデータ
  format: PixelFormat;    // RGBA, RGB, BGRA など
}
```

### AVIF エンコーダー

```typescript
// エンコーダーを作成
const encoder = new AVIFEncoder({
  quality: 65,
  speed: 6,
  yuvFormat: YUVFormat.YUV420,
  bitDepth: 10
});

// エンコード
const avifData = encoder.encode(jpegData);
encoder.close();
```

### AVIF デコーダー

```typescript
// デコーダーを作成
const decoder = new AVIFDecoder({
  format: PixelFormat.RGBA,
  jobs: -1  // すべてのコアを使用
});

// デコード
const decoded = decoder.decode(avifData);
decoder.close();
```

### パーシャルオプションのマージ

TypeScriptはオブジェクトのマージが容易なため、デフォルト設定を変更するのではなく、
パーシャルなオプションを渡してマージできる機能を提供します。

```typescript
// デフォルトオプションを取得
const defaultOptions = WebPEncoder.getDefaultOptions();

// 一部のオプションだけを変更
const encoder = new WebPEncoder({
  ...defaultOptions,
  quality: 90,
  method: 6
});

// または、部分的なオプションのみ指定（残りはデフォルト）
const encoder2 = new WebPEncoder({
  quality: 85  // その他のオプションはデフォルト
});
```

### 非同期API（バッチ処理用）

**問題:** 同期APIではイベントループがブロックされ、大量の画像処理時にスループットが低下します。

**解決策:** Worker Threadsまたはタスクプールを使用した非同期API

```typescript
// 非同期エンコーダー（Worker Thread使用）
class WebPEncoderAsync {
  constructor(options: Partial<WebPEncodeOptions>, workerCount?: number);

  async encode(data: Buffer): Promise<Buffer>;
  async encodeFile(path: string): Promise<Buffer>;
  async encodeBatch(files: string[]): Promise<Buffer[]>;

  async close(): Promise<void>;
}

// 使用例
const encoder = new WebPEncoderAsync({ quality: 80 }, 4); // 4 workers

// 並列処理
const results = await Promise.all([
  encoder.encode(data1),
  encoder.encode(data2),
  encoder.encode(data3)
]);

// バッチ処理（内部で並列化）
const files = ['img1.jpg', 'img2.jpg', 'img3.jpg'];
const webpFiles = await encoder.encodeBatch(files);

await encoder.close();
```

**実装方針:**
- フェーズ2で同期API実装後、フェーズ3-4で非同期版を追加
- Worker Threadsで同期エンコーダーをラップ
- タスクキューで負荷分散
- メモリ管理はWorker側で実施

**トレードオフ:**
- メリット: イベントループをブロックしない、CPU並列化
- デメリット: Worker起動オーバーヘッド、メモリ使用量増加
- 推奨: 10枚以上の画像処理時に非同期APIを使用

### オプション型定義

#### WebPEncodeOptions

```typescript
export interface WebPEncodeOptions {
  // 基本設定
  quality?: number;       // 0-100、デフォルト75
  lossless?: boolean;     // デフォルトfalse
  method?: number;        // 0-6、デフォルト4（品質/速度のトレードオフ）

  // プリセット
  preset?: WebPPreset;           // プリセットタイプ（デフォルト、写真、描画など）
  imageHint?: WebPImageHint;     // 画像タイプヒント
  losslessPreset?: number;       // 0-9、ロスレスプリセット

  // ターゲット設定
  targetSize?: number;     // ターゲットサイズ（バイト）、0=無効
  targetPSNR?: number;     // ターゲットPSNR、0=無効

  // セグメント/フィルタ設定
  segments?: number;              // 1-4、セグメント数、デフォルト4
  snsStrength?: number;           // 0-100、空間ノイズ整形
  filterStrength?: number;        // 0-100、フィルタ強度
  filterSharpness?: number;       // 0-7、フィルタ鮮鋭度
  filterType?: WebPFilterType;    // シンプル/強力
  autofilter?: boolean;           // 自動フィルタ調整

  // アルファチャンネル設定
  alphaMethod?: number;           // 0または1、透明度圧縮方法
  alphaFiltering?: WebPAlphaFilter; // アルファフィルタリング
  alphaQuality?: number;          // 0-100、アルファ圧縮品質

  // メタデータ設定
  keepMetadata?: number;   // MetadataNone, MetadataEXIF, MetadataICC, MetadataXMP, MetadataAll

  // 画像変換設定
  cropX?: number;          // クロップ矩形x
  cropY?: number;          // クロップ矩形y
  cropWidth?: number;      // クロップ幅
  cropHeight?: number;     // クロップ高さ
  resizeWidth?: number;    // リサイズ幅
  resizeHeight?: number;   // リサイズ高さ
  resizeMode?: WebPResizeMode; // リサイズモード

  // その他の設定
  nearLossless?: number;   // 0-100、ニアロスレス
  useSharpYUV?: boolean;   // シャープなRGB→YUV変換
  threadLevel?: boolean;   // マルチスレッド使用
  // ... 他のオプション（Golang版を参照）
}

// 列挙型
export enum WebPPreset {
  Default = 0,
  Picture = 1,
  Photo = 2,
  Drawing = 3,
  Icon = 4,
  Text = 5
}

export enum WebPImageHint {
  Default = 0,
  Picture = 1,
  Photo = 2,
  Graph = 3
}

export enum WebPFilterType {
  Simple = 0,
  Strong = 1
}

export enum WebPAlphaFilter {
  None = 0,
  Fast = 1,
  Best = 2
}

export enum WebPResizeMode {
  Always = 0,
  UpOnly = 1,
  DownOnly = 2
}

// メタデータフラグ（ビットフラグ）
export const MetadataNone = 0;
export const MetadataEXIF = 1 << 0;  // 1
export const MetadataICC = 1 << 1;   // 2
export const MetadataXMP = 1 << 2;   // 4
export const MetadataAll = MetadataEXIF | MetadataICC | MetadataXMP; // 7
```

#### WebPDecodeOptions

```typescript
export interface WebPDecodeOptions {
  // 基本設定
  useThreads?: boolean;          // マルチスレッド有効化
  bypassFiltering?: boolean;     // インループフィルタリング無効化
  noFancyUpsampling?: boolean;   // 高速ポイントワイズアップサンプラー使用
  format?: PixelFormat;          // ピクセルフォーマット（デフォルト: RGBA）

  // クロップ設定
  cropX?: number;
  cropY?: number;
  cropWidth?: number;
  cropHeight?: number;

  // スケール設定
  scaleWidth?: number;
  scaleHeight?: number;

  // ディザ設定
  ditheringStrength?: number;    // 0-100、ディザリング強度
}

export enum PixelFormat {
  RGBA = 0,
  BGRA = 1,
  RGB = 2,
  BGR = 3
}
```

#### AVIFEncodeOptions

```typescript
export interface AVIFEncodeOptions {
  // 品質設定
  quality?: number;        // 0-100、デフォルト60（色/YUV用）
  qualityAlpha?: number;   // 0-100、デフォルト-1（qualityを使用）
  speed?: number;          // 0-10、デフォルト6（0=最遅/最高品質、10=最速/最低品質）

  // フォーマット設定
  bitDepth?: number;       // 8、10、または12（デフォルト: 8）
  yuvFormat?: AVIFYUVFormat; // YUVフォーマット: 444/422/420/400
  yuvRange?: AVIFYUVRange;   // YUV範囲: limited/full

  // アルファ設定
  enableAlpha?: boolean;
  premultiplyAlpha?: boolean;

  // タイリング設定
  tileRowsLog2?: number;   // 0-6、デフォルト0
  tileColsLog2?: number;   // 0-6、デフォルト0

  // 詳細設定
  sharpYUV?: boolean;      // シャープなRGB→YUV変換
  targetSize?: number;     // ターゲットファイルサイズ（バイト）
  lossless?: boolean;      // ロスレスモード

  // スレッディングとタイリング
  jobs?: number;           // -1=すべてのコア、0=自動、>0=スレッド数
  autoTiling?: boolean;    // 自動タイリング有効化

  // メタデータ設定
  exifData?: Buffer;       // EXIFメタデータ
  xmpData?: Buffer;        // XMPメタデータ
  iccData?: Buffer;        // ICCプロファイル

  // 変換設定
  iRotAngle?: number;      // 画像回転: 0-3（90度×角度、反時計回り）
  iMirAxis?: AVIFMirrorAxis; // 画像ミラー

  // その他
  crop?: [number, number, number, number]; // [x, y, width, height]
  // ... 他のオプション（Golang版を参照）
}

export enum AVIFYUVFormat {
  YUV444 = 0,
  YUV422 = 1,
  YUV420 = 2,
  YUV400 = 3,
  Auto = -1
}

export enum AVIFYUVRange {
  Limited = 0,
  Full = 1
}

export enum AVIFMirrorAxis {
  None = -1,
  Vertical = 0,
  Horizontal = 1
}
```

#### AVIFDecodeOptions

```typescript
export interface AVIFDecodeOptions {
  // スレッディング
  jobs?: number;           // -1=すべてのコア、0=自動、>0=スレッド数

  // 出力フォーマット
  format?: PixelFormat;    // ピクセルフォーマット（デフォルト: RGBA）

  // 出力品質設定
  outputDepth?: number;    // 8または16ビット深度（PNGのみ）
  jpegQuality?: number;    // JPEG品質0-100（JPEGのみ）
  pngCompressLevel?: number; // PNG圧縮0-9（PNGのみ）

  // カラー処理
  rawColor?: boolean;      // アルファ乗算なしの生RGB出力

  // メタデータ処理
  ignoreExif?: boolean;
  ignoreXMP?: boolean;
  ignoreICC?: boolean;
  iccData?: Buffer;        // ICCプロファイルのオーバーライド

  // セキュリティ制限
  imageSizeLimit?: number;      // 最大画像サイズ（ピクセル数）
  imageDimensionLimit?: number; // 最大画像寸法（幅または高さ）

  // クロマアップサンプリング
  chromaUpsampling?: ChromaUpsampling;
}

export enum ChromaUpsampling {
  Automatic = 0,
  Fastest = 1,
  BestQuality = 2,
  Nearest = 3,
  Bilinear = 4
}
```

### ユーティリティ関数

```typescript
// ライブラリ情報
export function getLibraryPath(): string;
export function getPlatform(): string;
export function getLibraryFileName(): string;
export const VERSION: string;
```

### エラーハンドリング

```typescript
// エラーは説明的なメッセージを持つErrorオブジェクトとしてスローされます
try {
  const webpData = encodeWebP(jpegData);
} catch (error) {
  console.error('エンコード失敗:', error.message);
  // 例: "WebPエンコード失敗: 無効なパラメータ"
}

// Cライブラリからのステータスコード
export enum NextImageStatus {
  OK = 0,
  ERROR_INVALID_PARAM = -1,
  ERROR_ENCODE_FAILED = -2,
  ERROR_DECODE_FAILED = -3,
  ERROR_OUT_OF_MEMORY = -4,
  ERROR_UNSUPPORTED = -5,
  ERROR_BUFFER_TOO_SMALL = -6,
}
```

## FFI実装詳細

### Koffi（Node.js）

**型マッピング:**

| C型 | Koffi型 | TypeScript型 |
|--------|------------|-----------------|
| `uint8_t*` | `koffi.pointer('uint8_t')` | `Buffer` |
| `size_t` | `'size_t'` | `number` |
| `int` | `'int'` | `number` |
| `void*` | `koffi.pointer('void')` | `any` |
| `struct NextImageBuffer` | `koffi.struct(...)` | `interface` |

**バインディング例:**

```typescript
import koffi from 'koffi';

// C構造体を定義
const NextImageBufferStruct = koffi.struct('NextImageBuffer', {
  data: koffi.pointer('uint8_t'),
  size: 'size_t',
});

// ライブラリをロード
const lib = koffi.load(getLibraryPath());

// 関数をバインド
const encode = lib.func('nextimage_webp_encode_alloc', 'int', [
  koffi.pointer('uint8_t'),  // input_data
  'size_t',                  // input_size
  koffi.pointer('void'),     // options
  koffi.pointer(NextImageBufferStruct), // output
]);
```

**メモリ管理:**

```typescript
// 出力バッファ構造体を割り当て（JavaScriptヒープ上）
const outputPtr = koffi.alloc(NextImageBufferStruct, 1);

// 関数を呼び出し（C側でdata pointerを割り当て）
const status = encode(inputBuffer, inputBuffer.length, null, outputPtr);

if (status !== NextImageStatus.OK) {
  throw new Error(`Encoding failed: ${status}`);
}

// 結果をデコード
const output = koffi.decode(outputPtr, NextImageBufferStruct);

// 重要: CメモリからJavaScriptにデータをコピー
// output.dataはC側で割り当てられたポインタなので、必ずコピーしてから解放
const dataSize = Number(output.size);
const data = Buffer.from(koffi.decode(output.data, koffi.array('uint8_t', dataSize)));

// C割り当てメモリを解放（data pointerを解放）
// nextimage_free_buffer()は内部でfree(buffer->data)を実行
const lib = getLibrary();
const freeBuffer = lib.func('nextimage_free_buffer', 'void', [
  koffi.pointer(NextImageBufferStruct)
]);
freeBuffer(outputPtr);

// outputPtrはJavaScriptヒープ上の構造体なので自動的に回収される
```

**メモリ管理の注意点:**

1. **C側で割り当てられたメモリ（`output.data`）は必ず解放する**
   - `nextimage_free_buffer()`を呼ぶ前にデータをコピー
   - コピー後に`nextimage_free_buffer()`を呼んでC側メモリを解放

2. **JavaScript側で割り当てたメモリ（`outputPtr`）は自動回収される**
   - Koffiの`alloc()`で割り当てた構造体は自動的にGCされる
   - ただし、その中のポインタ（`data`）はC側の管理なので手動解放が必要

3. **FFIラッパー関数での実装例:**
   ```typescript
   function encodeWebPAlloc(inputData: Buffer): Buffer {
     const outputPtr = koffi.alloc(NextImageBufferStruct, 1);
     const status = encode(inputData, inputData.length, null, outputPtr);

     if (status !== NextImageStatus.OK) {
       throw new Error(`Encoding failed: ${status}`);
     }

     const output = koffi.decode(outputPtr, NextImageBufferStruct);
     const data = Buffer.from(koffi.decode(output.data, koffi.array('uint8_t', output.size)));

     // C側メモリを解放（dataをコピーした後なので安全）
     freeBuffer(outputPtr);

     return data;
   }
   ```

### Deno FFI（将来対応）

```typescript
const lib = Deno.dlopen(getLibraryPath(), {
  nextimage_webp_encode_alloc: {
    parameters: ["buffer", "usize", "pointer", "pointer"],
    result: "i32",
  },
  nextimage_free_buffer: {
    parameters: ["pointer"],
    result: "void",
  },
});

// 類似の使用パターン
const output = new Uint8Array(1024);
const status = lib.symbols.nextimage_webp_encode_alloc(
  inputData,
  inputData.length,
  null,
  output
);
```

### Bun FFI（将来対応）

```typescript
import { dlopen, FFIType, ptr } from "bun:ffi";

const lib = dlopen(getLibraryPath(), {
  nextimage_webp_encode_alloc: {
    args: [FFIType.ptr, FFIType.u64, FFIType.ptr, FFIType.ptr],
    returns: FFIType.i32,
  },
});

// 類似の使用パターン
```

## テスト戦略

### テスト構造

```
typescript/test/
├── webp-encode.test.ts    # WebPエンコードテスト
├── avif-encode.test.ts    # AVIFエンコードテスト（将来対応）
├── decode.test.ts         # デコードテスト（将来対応）
└── library.test.ts        # ライブラリパス解決テスト
```

### テストデータ

テストは`testdata/jpeg-source/`の画像を使用します:
- `landscape-like.jpg` - 実写真
- `gradient-horizontal.jpg` - グラデーションテスト
- `gradient-radial.jpg` - 放射状グラデーション
- `solid-black.jpg` - 単色
- `edges.jpg` - エッジ検出テスト

### テスト実行

```bash
# ビルドとテスト
npm test

# 出力:
#   dist/test/*.test.js      - コンパイル済みテスト
#   test-output/*.webp       - 目視確認用の生成画像
```

### CI/CDテスト

```yaml
# .github/workflows/test-typescript.yml
name: Test TypeScript Bindings

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [18, 20]

    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node }}

      - name: Build C library
        run: make install-c

      - name: Install TypeScript dependencies
        working-directory: typescript
        run: npm install

      - name: Run TypeScript tests
        working-directory: typescript
        run: npm test

      - name: Upload test outputs
        uses: actions/upload-artifact@v3
        with:
          name: test-outputs-${{ matrix.os }}-node${{ matrix.node }}
          path: typescript/test-output/
```

## 公開ワークフロー

### 前提条件

1. 全ターゲットプラットフォーム向けのCライブラリビルド
2. 共有ライブラリのGitHub releaseへのアップロード
3. TypeScriptコードのコンパイルとテスト

### リリースチェックリスト

- [ ] `package.json`のバージョン更新
- [ ] `VERSION`定数の更新（存在する場合）
- [ ] 全プラットフォーム向けCライブラリのビルド
- [ ] 共有ライブラリのGitHub releaseへのアップロード
- [ ] TypeScriptビルド: `npm run build`
- [ ] テスト実行: `npm test`
- [ ] コミットとタグ: `git tag v0.4.0`
- [ ] タグのプッシュ: `git push origin v0.4.0`
- [ ] npmへの公開: `npm publish`

### 公開コマンド

```bash
# 1. リリース準備
cd typescript
npm version patch  # または minor, major

# 2. ビルド
npm run build

# 3. テスト
npm test

# 4. 公開（ドライラン）
npm publish --dry-run

# 5. 公開
npm publish --access public

# 6. 確認
npm info @ideamans/libnextimage
```

## 現在の構造からの移行

現在の構造では、プラットフォーム固有のディレクトリがトップレベルにあります（`lib/darwin-arm64/`など）。
新しい構造では、開発用に`lib/shared/`、配布用に`typescript/lib/<platform>/`に統合します。

### 移行ステップ

1. **既存の`lib/`構造を維持** - 後方互換性のため
2. **`lib/shared/`を追加** - `make install-c`の主要な出力先として
3. **CMakeを更新** - 両方の場所にインストール:
   - `lib/shared/libnextimage.{so,dylib,dll}`（主要）
   - `lib/<platform>/`（プラットフォーム固有、互換性のため）
4. **`library.ts`を更新** - `lib/shared/`を優先
5. **`scripts/download-library.js`を追加** - npmインストール用
6. **`package.json`を更新** - postinstallフックを追加

### CMakeLists.txt変更

```cmake
# 共有ライブラリをlib/shared/にインストール（主要）
install(TARGETS nextimage_shared
  LIBRARY DESTINATION ${PROJECT_SOURCE_DIR}/../lib/shared
  RUNTIME DESTINATION ${PROJECT_SOURCE_DIR}/../lib/shared
)

# プラットフォーム固有のディレクトリにもインストール（互換性）
install(TARGETS nextimage_shared
  LIBRARY DESTINATION ${PROJECT_SOURCE_DIR}/../lib/${PLATFORM_ID}
  RUNTIME DESTINATION ${PROJECT_SOURCE_DIR}/../lib/${PLATFORM_ID}
)
```

## 実装計画

### フェーズ1: 基盤整備（Phase 1: Foundation）

**目標:** 基本的なFFIバインディングとライブラリロード機能を確立

**タスク:**

1. **プロジェクトセットアップ**
   - [ ] `typescript/`ディレクトリ構造の整備
   - [ ] `package.json`の作成・更新
   - [ ] `tsconfig.json`の設定
   - [ ] 依存関係のインストール（Koffi）
   - [ ] `.gitignore`の設定（`dist/`, `node_modules/`, `lib/`）

2. **ライブラリパス解決**
   - [ ] `src/library.ts`の実装
     - `getPlatform()` - プラットフォーム検出
     - `getLibraryFileName()` - ライブラリファイル名取得
     - `findLibraryPath()` - 3段階のフォールバック検索（修正版: 2段階上がる）
     - `getLibraryPath()` - キャッシュ機能付き
   - [ ] **重要:** パス解決のリグレッションテストを作成
     - `test/library-path.test.ts`の作成
     - コンパイル済みコード（`dist/`）からのパス解決テスト
     - 開発モード（`typescript/dist/` → `lib/shared/`）のテスト
     - インストールモード（`node_modules/@ideamans/libnextimage/dist/` → `../lib/<platform>/`）のテスト
     - モックファイルシステムでの検証

3. **基本的なFFI型定義**
   - [ ] `src/types.ts`の作成
     - `NextImageStatus` enum
     - `PixelFormat` enum
     - `NextImageBuffer` interface
     - `NextImageDecodeBuffer` interface
     - `DecodedImage` interface
   - [ ] エラーハンドリング用の型定義

4. **FFI基盤実装**
   - [ ] `src/ffi.ts`の実装
     - Koffiによるライブラリロード
     - 基本的な構造体定義
     - エラーメッセージ取得関数
   - [ ] ライブラリロードのテスト

**成果物:**
- ライブラリが正しくロードされる
- プラットフォーム検出が動作する
- 基本的な型定義が揃う

**所要時間:** 1-2日

---

### フェーズ2: WebPエンコーダー実装（Phase 2: WebP Encoder）

**目標:** WebPエンコーダーの完全実装

**タスク:**

1. **WebPオプション型定義**
   - [ ] `src/webp-types.ts`の作成
     - `WebPEncodeOptions` interface（全フィールド）
     - `WebPDecodeOptions` interface
     - 列挙型（`WebPPreset`, `WebPImageHint`, `WebPFilterType`など）
     - メタデータフラグ定数
   - [ ] デフォルトオプションの定義

2. **WebPエンコーダーFFIバインディング**
   - [ ] `src/ffi-webp.ts`の実装
     - `NextImageWebPEncoder`ポインタ型定義
     - C構造体のマッピング
     - エンコーダー作成/破棄関数のバインド
     - エンコード関数のバインド
   - [ ] オプション変換関数（JS → C構造体）

3. **WebPEncoderクラス実装**
   - [ ] `src/webp-encoder.ts`の作成
     - `constructor(options: Partial<WebPEncodeOptions>)`
     - `encode(data: Buffer): Buffer`
     - `encodeFile(path: string): Buffer`
     - `close(): void`
     - `static getDefaultOptions(): WebPEncodeOptions`
   - [ ] ファイナライザーの実装（自動リソース解放）
   - [ ] オプションマージ機能

4. **テストとドキュメント**
   - [ ] `test/webp-encoder.test.ts`の作成
     - 基本的なエンコードテスト
     - 品質設定テスト
     - ロスレスモードテスト
     - 複数ファイルのエンコードテスト
     - クロップ/リサイズテスト
     - メタデータ保持テスト
   - [ ] 使用例の作成（`examples/webp-encode.ts`）

**成果物:**
- WebPエンコーダーが完全に動作
- 全オプションがサポートされる
- テストカバレッジ80%以上

**所要時間:** 3-4日

---

### フェーズ3: WebPデコーダー実装（Phase 3: WebP Decoder）

**目標:** WebPデコーダーの完全実装

**タスク:**

1. **WebPデコーダーFFIバインディング**
   - [ ] `src/ffi-webp.ts`への追加
     - `NextImageWebPDecoder`ポインタ型定義
     - デコーダー作成/破棄関数のバインド
     - デコード関数のバインド
   - [ ] デコードバッファの変換関数

2. **WebPDecoderクラス実装**
   - [ ] `src/webp-decoder.ts`の作成
     - `constructor(options: Partial<WebPDecodeOptions>)`
     - `decode(data: Buffer): DecodedImage`
     - `decodeFile(path: string): DecodedImage`
     - `close(): void`
     - `static getDefaultOptions(): WebPDecodeOptions`

3. **テストとドキュメント**
   - [ ] `test/webp-decoder.test.ts`の作成
     - 基本的なデコードテスト
     - ピクセルフォーマット変換テスト
     - クロップ/スケールテスト
     - マルチスレッドテスト
   - [ ] 使用例の作成（`examples/webp-decode.ts`）

**成果物:**
- WebPデコーダーが完全に動作
- 様々なピクセルフォーマットに対応
- エンコーダーとのラウンドトリップテスト成功

**所要時間:** 2-3日

---

### フェーズ4: AVIFエンコーダー実装（Phase 4: AVIF Encoder）

**目標:** AVIFエンコーダーの完全実装

**タスク:**

1. **AVIFオプション型定義**
   - [ ] `src/avif-types.ts`の作成
     - `AVIFEncodeOptions` interface（全フィールド）
     - `AVIFDecodeOptions` interface
     - 列挙型（`AVIFYUVFormat`, `AVIFYUVRange`, `AVIFMirrorAxis`など）

2. **AVIFエンコーダーFFIバインディング**
   - [ ] `src/ffi-avif.ts`の実装
     - `NextImageAVIFEncoder`ポインタ型定義
     - C構造体のマッピング（YUVフォーマット、ビット深度など）
     - エンコーダー作成/破棄関数のバインド
     - エンコード関数のバインド
   - [ ] オプション変換関数（特にメタデータバイト配列）

3. **AVIFEncoderクラス実装**
   - [ ] `src/avif-encoder.ts`の作成
     - WebPEncoderと同様の構造
     - AVIF固有のオプション処理

4. **テストとドキュメント**
   - [ ] `test/avif-encoder.test.ts`の作成
     - 品質/速度設定テスト
     - ビット深度テスト（8/10/12bit）
     - YUVフォーマットテスト
     - メタデータ埋め込みテスト
   - [ ] 使用例の作成（`examples/avif-encode.ts`）

**成果物:**
- AVIFエンコーダーが完全に動作
- 10bit/12bit対応
- メタデータ埋め込み機能

**所要時間:** 3-4日

---

### フェーズ5: AVIFデコーダー実装（Phase 5: AVIF Decoder）

**目標:** AVIFデコーダーの完全実装

**タスク:**

1. **AVIFデコーダーFFIバインディング**
   - [ ] `src/ffi-avif.ts`への追加
     - `NextImageAVIFDecoder`ポインタ型定義
     - デコーダー作成/破棄関数のバインド
     - デコード関数のバインド

2. **AVIFDecoderクラス実装**
   - [ ] `src/avif-decoder.ts`の作成
     - WebPDecoderと同様の構造
     - クロマアップサンプリング設定
     - セキュリティ制限の実装

3. **テストとドキュメント**
   - [ ] `test/avif-decoder.test.ts`の作成
     - 基本的なデコードテスト
     - 高ビット深度デコードテスト
     - クロマアップサンプリングテスト
     - メタデータ抽出テスト
   - [ ] 使用例の作成（`examples/avif-decode.ts`）

**成果物:**
- AVIFデコーダーが完全に動作
- 高ビット深度対応
- メタデータ抽出機能

**所要時間:** 2-3日

---

### フェーズ6: パッケージング準備（Phase 6: Packaging）

**目標:** npm公開の準備

**タスク:**

1. **postinstallスクリプト実装**
   - [ ] `scripts/download-library.js`の作成
     - プラットフォーム検出
     - GitHub Releasesからのダウンロード
     - ファイル展開と権限設定
     - エラーハンドリング（ダウンロード失敗時の対応）

2. **ビルドとテスト自動化**
   - [ ] `package.json`のscripts更新
     - `build`: TypeScriptコンパイル
     - `test`: 全テスト実行
     - `test:watch`: 継続的テスト
     - `lint`: コードスタイルチェック
     - `prepublishOnly`: 公開前チェック

3. **ドキュメント整備**
   - [ ] `typescript/README.md`の作成
     - インストール方法
     - 基本的な使用例
     - API リファレンス
     - トラブルシューティング
   - [ ] API型定義ドキュメントの生成（TypeDoc）

4. **統合テスト**
   - [ ] 実際のnpmインストールフローのテスト
   - [ ] 各プラットフォームでの動作確認
   - [ ] メモリリークテスト
   - [ ] パフォーマンスベンチマーク

**成果物:**
- npm公開可能なパッケージ
- 完全なドキュメント
- CI/CD設定

**所要時間:** 2-3日

---

### フェーズ7: メインエクスポートと統合（Phase 7: Integration）

**目標:** 統一されたAPIの提供

**タスク:**

1. **メインエントリポイント実装**
   - [ ] `src/index.ts`の作成
     - すべてのエンコーダー/デコーダーのエクスポート
     - すべての型定義のエクスポート
     - すべての列挙型のエクスポート
     - ユーティリティ関数のエクスポート
     - バージョン情報のエクスポート

2. **統合テスト**
   - [ ] `test/integration.test.ts`の作成
     - エンコード → デコード → 再エンコードのラウンドトリップ
     - WebP ⇔ AVIF 相互変換
     - 複数エンコーダーの同時使用
     - メモリ管理の検証

3. **使用例とベストプラクティス**
   - [ ] `examples/batch-convert.ts` - バッチ変換
   - [ ] `examples/compare-formats.ts` - フォーマット比較
   - [ ] `examples/metadata-preservation.ts` - メタデータ保持
   - [ ] `examples/advanced-options.ts` - 詳細設定

4. **ランタイム別E2Eテスト例**
   - [ ] `examples/nodejs/` - Node.js向けE2Eテスト
     - `package.json` - 公開されたlibnextimageを依存関係として記述
     - `basic-encode.js` - 基本的なエンコードテスト
     - `batch-process.js` - バッチ処理テスト
     - `async-worker.js` - Worker Threads使用例
     - `README.md` - セットアップと実行手順
   - [ ] `examples/deno/` - Deno向けE2Eテスト
     - `deno.json` - import mapとパーミッション設定
     - `basic-encode.ts` - 基本的なエンコードテスト
     - `remote-import.ts` - deno.land/x からのインポート
     - `npm-specifier.ts` - npm: specifier使用例
     - `README.md` - セットアップと実行手順
   - [ ] `examples/bun/` - Bun向けE2Eテスト
     - `package.json` - 公開されたlibnextimageを依存関係として記述
     - `basic-encode.ts` - 基本的なエンコードテスト
     - `benchmark.ts` - パフォーマンスベンチマーク
     - `README.md` - セットアップと実行手順

**成果物:**
- 統一されたAPI
- 包括的な使用例
- ベストプラクティスドキュメント
- **ランタイム別のE2Eテスト（公開パッケージを使用）**

**所要時間:** 2-3日

---

### フェーズ8: CI/CD と公開（Phase 8: CI/CD & Release）

**目標:** 自動化とnpm公開

**タスク:**

1. **GitHub Actions CI/CD設定**
   - [ ] `.github/workflows/test-typescript.yml`の作成
     - 複数プラットフォームでのテスト（Ubuntu, macOS, Windows）
     - 複数Node.jsバージョンでのテスト（18, 20, 22）
     - テストカバレッジレポート
   - [ ] `.github/workflows/release.yml`の作成
     - タグプッシュ時の自動ビルド
     - GitHub Releaseの自動作成
     - 共有ライブラリの自動アップロード

2. **npm公開準備**
   - [ ] バージョン管理戦略の確立
   - [ ] CHANGELOG.mdの作成
   - [ ] LICENSE ファイルの確認
   - [ ] `.npmignore`の設定
   - [ ] npm dry-runでの検証

3. **初回リリース**
   - [ ] v0.4.0タグの作成
   - [ ] 全プラットフォーム向け共有ライブラリのビルド
   - [ ] GitHub Releaseへのアップロード
   - [ ] npm公開（`npm publish --access public`）
   - [ ] リリースノートの作成

**成果物:**
- 自動化されたCI/CD
- npm公開パッケージ
- 公開ドキュメント

**所要時間:** 2-3日

---

## 実装の優先順位とマイルストーン

### マイルストーン1: MVP（Minimum Viable Product）
- フェーズ1 + フェーズ2 + フェーズ6（部分）
- WebPエンコーダーのみ
- 基本的なnpmインストール対応
- **目標期間:** 1-2週間

### マイルストーン2: WebP完全対応
- マイルストーン1 + フェーズ3
- WebPエンコーダー + デコーダー
- 完全なテストカバレッジ
- **目標期間:** 2-3週間

### マイルストーン3: AVIF対応
- マイルストーン2 + フェーズ4 + フェーズ5
- WebP + AVIF 完全対応
- 相互変換機能
- **目標期間:** 3-4週間

### マイルストーン4: 公開リリース
- マイルストーン3 + フェーズ7 + フェーズ8
- 完全な機能とドキュメント
- CI/CD自動化
- npm公開
- **目標期間:** 4-5週間

---

## 開発上の注意点

### メモリ管理
- **必須:** Cメモリを解放する前に必ずデータをコピー
- **必須:** `close()`メソッドの実装と呼び出し
- **推奨:** ファイナライザーの実装（ガベージコレクション時の自動解放）
- **テスト:** メモリリークテストの実施

### エラーハンドリング
- すべてのFFI呼び出しでステータスコードをチェック
- わかりやすいエラーメッセージを提供
- エラー発生時のリソース解放を保証

### パフォーマンス
- 初期化オーバーヘッドの最小化（エンコーダー再利用）
- 大きなバッファのコピーを最小限に
- 必要に応じてマルチスレッド対応

### テスト
- 実際の画像ファイルを使用
- 各プラットフォームでのテスト
- エッジケースのカバー（空データ、巨大ファイルなど）
- メモリリークテスト

### ドキュメント
- TypeScriptのJSDocコメントを完備
- 使用例を豊富に提供
- トラブルシューティングガイド
- Golang版との差異を明記

---

## ランタイム別E2Eテストの詳細設計

### examples/nodejs/ の構造

**目的:** 公開されたnpmパッケージを実際にインストールして使用する完全なE2Eテスト

**ディレクトリ構造:**
```
examples/nodejs/
├── package.json           # 公開されたlibnextimageを依存関係に
├── .gitignore            # node_modules/, output/ を除外
├── README.md             # セットアップと実行手順
├── input/                # テスト用画像
│   ├── test.jpg
│   ├── test.png
│   └── test-alpha.png
├── output/               # 出力先（生成される）
└── scripts/
    ├── basic-encode.js   # 基本的なエンコード
    ├── batch-process.js  # バッチ処理
    ├── async-worker.js   # Worker Threads
    └── all-features.js   # 全機能テスト
```

**package.json:**
```json
{
  "name": "libnextimage-nodejs-e2e",
  "version": "1.0.0",
  "private": true,
  "description": "End-to-end test for @ideamans/libnextimage on Node.js",
  "type": "module",
  "dependencies": {
    "@ideamans/libnextimage": "^0.4.0"
  },
  "scripts": {
    "test:basic": "node scripts/basic-encode.js",
    "test:batch": "node scripts/batch-process.js",
    "test:async": "node scripts/async-worker.js",
    "test:all": "node scripts/all-features.js",
    "test": "npm run test:basic && npm run test:batch && npm run test:all"
  }
}
```

**basic-encode.js の例:**
```javascript
import fs from 'fs';
import path from 'path';
import { WebPEncoder, AVIFEncoder } from '@ideamans/libnextimage';

console.log('=== libnextimage Node.js E2E Test: Basic Encode ===\n');

// WebPエンコード
const webpEncoder = new WebPEncoder({ quality: 80 });
const jpegData = fs.readFileSync(path.join('input', 'test.jpg'));
const webpData = webpEncoder.encode(jpegData);
fs.writeFileSync(path.join('output', 'test.webp'), webpData);
console.log(`✓ WebP: ${jpegData.length} bytes → ${webpData.length} bytes`);
webpEncoder.close();

// AVIFエンコード
const avifEncoder = new AVIFEncoder({ quality: 65, speed: 6 });
const avifData = avifEncoder.encode(jpegData);
fs.writeFileSync(path.join('output', 'test.avif'), avifData);
console.log(`✓ AVIF: ${jpegData.length} bytes → ${avifData.length} bytes`);
avifEncoder.close();

console.log('\nAll tests passed!');
```

### examples/deno/ の構造

**目的:** Deno環境での公開パッケージ使用テスト

**ディレクトリ構造:**
```
examples/deno/
├── deno.json             # import map とパーミッション
├── .gitignore
├── README.md
├── input/
│   └── test.jpg
├── output/
└── scripts/
    ├── basic-encode.ts       # 基本的なエンコード
    ├── remote-import.ts      # deno.land/x からインポート
    └── npm-specifier.ts      # npm: specifier使用
```

**deno.json:**
```json
{
  "imports": {
    "@ideamans/libnextimage": "npm:@ideamans/libnextimage@^0.4.0",
    "libnextimage-deno": "https://deno.land/x/libnextimage@v0.4.0/deno/mod.ts"
  },
  "tasks": {
    "test:basic": "deno run --allow-read --allow-write --allow-ffi scripts/basic-encode.ts",
    "test:remote": "deno run --allow-read --allow-write --allow-ffi --allow-net scripts/remote-import.ts",
    "test:npm": "deno run --allow-read --allow-write --allow-ffi --allow-env scripts/npm-specifier.ts"
  }
}
```

**basic-encode.ts の例:**
```typescript
import { WebPEncoder } from "@ideamans/libnextimage";

console.log('=== libnextimage Deno E2E Test: Basic Encode ===\n');

const jpegData = await Deno.readFile('input/test.jpg');

const encoder = new WebPEncoder({ quality: 80 });
const webpData = encoder.encode(jpegData);
await Deno.writeFile('output/test.webp', webpData);

console.log(`✓ WebP: ${jpegData.length} bytes → ${webpData.length} bytes`);
encoder.close();

console.log('\nTest passed!');
```

**remote-import.ts の例:**
```typescript
import { WebPEncoder } from "libnextimage-deno";

console.log('=== libnextimage Deno E2E Test: Remote Import ===\n');
console.log('Using: https://deno.land/x/libnextimage/deno/mod.ts\n');

// 同じテストロジック
```

### examples/bun/ の構造

**目的:** Bun環境での公開パッケージ使用テスト

**ディレクトリ構造:**
```
examples/bun/
├── package.json
├── bunfig.toml           # Bun設定
├── .gitignore
├── README.md
├── input/
│   └── test.jpg
├── output/
└── scripts/
    ├── basic-encode.ts
    ├── benchmark.ts      # パフォーマンス比較
    └── memory-test.ts    # メモリ使用量測定
```

**package.json:**
```json
{
  "name": "libnextimage-bun-e2e",
  "version": "1.0.0",
  "private": true,
  "description": "End-to-end test for @ideamans/libnextimage on Bun",
  "type": "module",
  "dependencies": {
    "@ideamans/libnextimage": "^0.4.0"
  },
  "scripts": {
    "test:basic": "bun run scripts/basic-encode.ts",
    "test:benchmark": "bun run scripts/benchmark.ts",
    "test:memory": "bun run scripts/memory-test.ts",
    "test": "bun run test:basic && bun run test:benchmark"
  }
}
```

**benchmark.ts の例:**
```typescript
import { WebPEncoder, AVIFEncoder } from '@ideamans/libnextimage/bun';
import { readFileSync, writeFileSync } from 'fs';

console.log('=== libnextimage Bun E2E Test: Benchmark ===\n');

const jpegData = readFileSync('input/test.jpg');
const iterations = 100;

// WebP benchmark
const webpEncoder = new WebPEncoder({ quality: 80 });
const webpStart = Bun.nanoseconds();
for (let i = 0; i < iterations; i++) {
  const webpData = webpEncoder.encode(jpegData);
}
const webpTime = (Bun.nanoseconds() - webpStart) / 1_000_000;
webpEncoder.close();

// AVIF benchmark
const avifEncoder = new AVIFEncoder({ quality: 65, speed: 6 });
const avifStart = Bun.nanoseconds();
for (let i = 0; i < iterations; i++) {
  const avifData = avifEncoder.encode(jpegData);
}
const avifTime = (Bun.nanoseconds() - avifStart) / 1_000_000;
avifEncoder.close();

console.log(`WebP: ${(webpTime / iterations).toFixed(2)} ms/image`);
console.log(`AVIF: ${(avifTime / iterations).toFixed(2)} ms/image`);
console.log('\nBenchmark completed!');
```

### 各ランタイムのREADME.mdに含めるべき内容

1. **前提条件**
   - ランタイムのインストール方法
   - 必要なバージョン

2. **セットアップ**
   ```bash
   # Node.js
   cd examples/nodejs
   npm install

   # Deno
   cd examples/deno
   # 依存関係は自動ダウンロード

   # Bun
   cd examples/bun
   bun install
   ```

3. **実行方法**
   ```bash
   # Node.js
   npm test

   # Deno
   deno task test:basic

   # Bun
   bun test
   ```

4. **トラブルシューティング**
   - プラットフォーム固有の問題
   - パーミッションエラーの対処
   - バイナリが見つからない場合の対処

5. **期待される出力**
   - 成功時のログ例
   - 生成されるファイル一覧

### CI/CDでのE2Eテスト実行

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e-nodejs:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [18, 20, 22]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node }}
      - name: Run Node.js E2E tests
        working-directory: examples/nodejs
        run: |
          npm install
          npm test

  e2e-deno:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v3
      - uses: denoland/setup-deno@v1
        with:
          deno-version: v1.x
      - name: Run Deno E2E tests
        working-directory: examples/deno
        run: |
          deno task test:basic
          deno task test:npm

  e2e-bun:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
    steps:
      - uses: actions/checkout@v3
      - uses: oven-sh/setup-bun@v1
      - name: Run Bun E2E tests
        working-directory: examples/bun
        run: |
          bun install
          bun test
```

---

## 将来の機能強化（フェーズ9以降）

### 短期
- [ ] Deno FFI対応（`deno/`ディレクトリ）
- [ ] Bun FFI対応（`bun/`ディレクトリ）
- [ ] ストリーミングAPI（大きなファイル対応）
- [ ] プログレスコールバック（長時間操作）

### 中期
- [ ] WebAssembly版（ブラウザ対応）
- [ ] GIFアニメーション対応
- [ ] WebPアニメーション対応
- [ ] バッチ処理ユーティリティ

### 長期
- [ ] 画像操作API（リサイズ、クロップ、回転）
- [ ] メタデータ抽出・編集API
- [ ] 色空間変換
- [ ] AIベースの最適化ヒント

## 参考資料

- [Koffiドキュメント](https://github.com/Koromix/koffi)
- [Deno FFIドキュメント](https://deno.land/manual/runtime/ffi_api)
- [Bun FFIドキュメント](https://bun.sh/docs/api/ffi)
- [WebP仕様](https://developers.google.com/speed/webp)
- [AVIF仕様](https://aomediacodec.github.io/av1-avif/)

## 備考

- この設計は**インストールの容易さ**（プリビルドバイナリ）を**ビルドの柔軟性**よりも優先します
- **プラットフォーム検出**は自動 - 設定不要
- **開発ワークフロー**は最適化 - 手動ファイルコピー不要
- **エラーメッセージ**はユーザーを解決策へ導きます（例: "make install-cを実行してください"）
- **FFIアプローチ**はネイティブアドオンのコンパイル（node-gyp、cmake-js）を回避
- **メモリ安全性**は重要 - Cメモリを解放する前に常にデータをコピー
- **テスト**は単なるユニットテストではなく、実際の画像を使用した検証
