# Command-Line Tool Compatibility Analysis

このドキュメントは、libwebp/libavifの公式コマンドラインツールと、libnextimageのライブラリAPIとの機能互換性を分析したものです。

## 1. cwebp vs WebP Encode API

### cwebp コマンドライン引数

#### 基本オプション
- ✅ `-q <float>` - quality (0-100) → `quality`
- ✅ `-alpha_q <int>` - alpha quality (0-100) → `alpha_quality`
- ✅ `-preset <string>` - preset (default, photo, picture, drawing, icon, text) → `preset`
- ✅ `-z <int>` - lossless preset level (0-9) → `lossless_preset`
- ✅ `-m <int>` - compression method (0-6) → `method`
- ✅ `-segments <int>` - number of segments (1-4) → `segments`
- ✅ `-size <int>` - target size in bytes → `target_size`
- ✅ `-psnr <float>` - target PSNR → `target_psnr`

#### フィルタリングオプション
- ✅ `-sns <int>` - spatial noise shaping (0-100) → `sns_strength`
- ✅ `-f <int>` - filter strength (0-100) → `filter_strength`
- ✅ `-sharpness <int>` - filter sharpness (0-7) → `filter_sharpness`
- ✅ `-strong` / `-nostrong` - strong/simple filter → `filter_type` (1=strong, 0=simple)
- ✅ `-sharp_yuv` - sharper RGB->YUV conversion → `use_sharp_yuv`
- ✅ `-af` - auto-adjust filter strength → `autofilter`

#### 圧縮オプション
- ✅ `-partition_limit <int>` - quality degradation (0-100) → `partition_limit`
- ✅ `-pass <int>` - analysis pass (1-10) → `pass`
- ✅ `-qrange <min> <max>` - quality range → `qmin`, `qmax`
- ✅ `-crop <x> <y> <w> <h>` - crop picture → `crop_x`, `crop_y`, `crop_width`, `crop_height`
- ✅ `-resize <w> <h>` - resize picture → `resize_width`, `resize_height`
- ✅ `-resize_mode <string>` - resize mode (up_only, down_only, always) → `resize_mode` (0=always, 1=up_only, 2=down_only)

#### マルチスレッドとメモリ
- ✅ `-mt` - multi-threading → `thread_level`
- ✅ `-low_memory` - reduce memory usage → `low_memory`

#### アルファチャンネルオプション
- ✅ `-alpha_method <int>` - alpha compression method (0-1) → `alpha_compression`
- ✅ `-alpha_filter <string>` - alpha filtering (none=0, fast=1, best=2) → `alpha_filtering`
- ✅ `-exact` - preserve RGB in transparent area → `exact`
- ✅ `-blend_alpha <hex>` - blend against background → `blend_alpha` (0xRRGGBB format)
- ✅ `-noalpha` - discard transparency → `noalpha`

#### ロスレスオプション
- ✅ `-lossless` - lossless encoding → `lossless`
- ✅ `-near_lossless <int>` - near-lossless (0-100) → `near_lossless`
- ✅ `-hint <string>` - image hint (photo, picture, graph) → `image_hint`

#### メタデータオプション
- ✅ `-metadata <string>` - metadata to copy (all, none, exif, icc, xmp) → `keep_metadata`

#### 出力/デバッグオプション
- ✅ `-map <int>` - print map of extra info → `show_compressed`
- ⚠️ `-print_psnr` - print PSNR distortion → **未サポート（統計情報出力機能なし）**
- ⚠️ `-print_ssim` - print SSIM distortion → **未サポート（統計情報出力機能なし）**
- ⚠️ `-print_lsim` - print local-similarity → **未サポート（統計情報出力機能なし）**
- ⚠️ `-d <file.pgm>` - dump compressed output → **未サポート（デバッグ機能）**
- ⚠️ `-short` / `-quiet` / `-v` / `-progress` - 出力制御 → **未サポート（CLI固有）**
- ⚠️ `-version` / `-noasm` - システム設定 → **未サポート（CLI/システム固有）**

#### 実験的オプション
- ✅ `-jpeg_like` - match JPEG size → `emulate_jpeg_size`
- ✅ `-pre <int>` - pre-processing filter → `preprocessing`

#### YUV入力オプション
- ⚠️ `-s <int> <int>` - input size for YUV → **未サポート（生ピクセル入力は別API設計が必要）**

### 分析結果

#### サポート状況
- **完全対応**: 基本的なエンコード品質、フィルタ、圧縮、アルファ、ロスレス、メタデータ、画像変換（crop/resize）の各設定
- **未対応（YUV入力）**: `-s` (YUV入力)
  - **理由**: 生YUVピクセルデータの入力は別API設計が必要
- **未対応（CLI固有）**: `-print_psnr`, `-print_ssim`, `-print_lsim`, `-d`, `-short`, `-quiet`, `-v`, `-progress`, `-version`, `-noasm`
  - **理由**: これらはコマンドラインツール固有の機能で、ライブラリAPIでは不要

#### 重要な追加機能
- **画像変換のフル対応**: `-crop`, `-resize`, `-resize_mode` (up_only, down_only, always)
  - WebPPictureCrop/WebPPictureRescaleを利用
  - cwebpと同じ処理順序（crop → resize → blend_alpha）
- **アルファチャンネル処理**: `-blend_alpha`, `-noalpha`
  - WebPBlendAlphaによる背景色との合成
  - 読み込み時のアルファチャンネル除去

#### 結論
cwebpの**全コア機能を完全にサポート**。
画像変換（crop/resize/blend_alpha/noalpha）の追加により、cwebpコマンドと同等の機能を提供。

---

## 2. dwebp vs WebP Decode API

### dwebp コマンドライン引数

#### 出力フォーマットオプション
- ✅ デフォルト: PNG形式 → `format` (FormatRGBA等)
- ⚠️ `-pam` - save as PAM → **未サポート（出力側で処理すべき）**
- ⚠️ `-ppm` - save as PPM → **未サポート（出力側で処理すべき）**
- ⚠️ `-bmp` - save as BMP → **未サポート（出力側で処理すべき）**
- ⚠️ `-tiff` - save as TIFF → **未サポート（出力側で処理すべき）**
- ⚠️ `-pgm` - save as PGM (YUV) → **未サポート（YUV出力は別設計）**
- ⚠️ `-yuv` - save raw YUV → **未サポート（YUV出力は別設計）**

#### デコードオプション
- ✅ `-nofancy` - don't use fancy upscaler → `no_fancy_upsampling`
- ✅ `-nofilter` - disable in-loop filtering → `bypass_filtering`
- ✅ `-nodither` - disable dithering → `no_dither`
- ✅ `-dither <d>` - dithering strength (0-100) → `dither_strength`
- ✅ `-alpha_dither` - alpha-plane dithering → `alpha_dither`
- ✅ `-mt` - multi-threading → `use_threads`

#### 画像変換オプション
- ✅ `-crop <x> <y> <w> <h>` - crop output → `crop_x`, `crop_y`, `crop_width`, `crop_height`, `use_crop`
- ✅ `-resize <w> <h>` - resize output → `resize_width`, `resize_height`, `use_resize`
- ✅ `-flip` - flip vertically → `flip`
- ✅ `-alpha` - only save alpha plane → `alpha_only`

#### デバッグ/その他オプション
- ✅ `-incremental` - incremental decoding → `incremental`
- ⚠️ `-version` - print version → **未サポート（CLI固有）**
- ⚠️ `-v` - verbose → **未サポート（CLI固有）**
- ⚠️ `-quiet` - quiet mode → **未サポート（CLI固有）**
- ⚠️ `-noasm` - disable assembly → **未サポート（システム固有）**

### 分析結果

#### サポート状況
- **完全対応**: デコード品質設定、ディザリング、画像変換（crop/resize/flip）、アルファ処理、マルチスレッド
- **未対応（出力処理）**: `-pam`, `-ppm`, `-bmp`, `-tiff`, `-pgm`, `-yuv`
  - **理由**: 出力フォーマット変換は出力側で行うべき。ライブラリはピクセルデータ（RGBA等）を返すのみ
- **未対応（CLI固有）**: `-version`, `-v`, `-quiet`, `-noasm`
  - **理由**: コマンドラインツール固有の機能

#### 重要な違い
- **dwebpとの大きな違い**: dwebpは**デコード後にcrop/resize**を適用するが、ライブラリでは同じAPIで対応可能
  - cwebpでは crop/resize は入力処理（エンコード前）
  - dwebpでは crop/resize は出力処理（デコード後）
  - **ライブラリでは両方をサポート**しており、より柔軟

#### 結論
dwebpの**コア機能は完全にサポート**されている。
crop/resize機能はcwebpよりも充実しており、デコード後の画像変換を直接サポート。

---

## 3. gif2webp vs GIF to WebP API

### gif2webp コマンドライン引数

#### 基本オプション
- ✅ `-lossy` - lossy encoding → `lossless = false` (デフォルト動作)
- ✅ `-near_lossless <int>` - near-lossless (0-100) → `near_lossless`
- ✅ `-sharp_yuv` - sharp RGB->YUV → `use_sharp_yuv`
- ✅ `-q <float>` - quality (0-100) → `quality`
- ✅ `-m <int>` - compression method (0-6) → `method`
- ✅ `-f <int>` - filter strength (0-100) → `filter_strength`

#### アニメーション設定
- ✅ `-mixed` - mixed lossy/lossless → `allow_mixed`
- ✅ `-min_size` - minimize output size → `minimize_size`
- ✅ `-kmin <int>` - min keyframe distance → `kmin`
- ✅ `-kmax <int>` - max keyframe distance → `kmax`
- ✅ `-loop_compatibility` - Chrome M62 compatibility → `loop_compatibility`
- ✅ `-loop_count <int>` - loop count (0=infinite) → `anim_loop_count`

#### メタデータ設定
- ✅ `-metadata <string>` - metadata copy (all, none, icc, xmp) → `keep_metadata`

#### その他
- ✅ `-mt` - multi-threading → `thread_level`
- ⚠️ `-version` / `-v` / `-quiet` - 出力制御 → **未サポート（CLI固有）**

### 分析結果

#### サポート状況
- **完全対応**: 基本的なロスレス/ロッシー変換、品質設定、メタデータ、**アニメーション変換**
- **完全対応（アニメーション）**: `-mixed`, `-min_size`, `-kmin`, `-kmax`, `-loop_compatibility`, `-loop_count`
  - **実装**: WebPAnimEncoderを使用した完全なアニメーションGIF→WebP変換
  - **機能**: GIFのフレームタイミング、透明度、ディスポーズメソッドを完全サポート
- **未対応（CLI固有）**: `-version`, `-v`, `-quiet`

#### 実装済み機能
- **アニメーションGIF→アニメーションWebP変換**
  - WebPAnimEncoderを使用した完全実装
  - GIFフレームの正確な読み取りと変換
  - フレームタイミングの保持（最小100ms）
  - 透明度とディスポーズメソッドの完全対応
  - 3フレームバッファ方式（frame, curr_canvas, prev_canvas）による正確な合成
- **アニメーション最適化オプション**
  - `-mixed`: ロッシー/ロッスレス混在エンコーディング
  - `-min_size`: 出力サイズの最小化（処理時間増加）
  - `-kmin`/`-kmax`: キーフレーム間隔の制御
  - `-loop_compatibility`: Chrome M62互換モード

#### 結論
gif2webpコマンドの**全主要機能を完全にサポート**。
静止画GIF、アニメーションGIFの両方に対応し、品質・サイズ・互換性の制御が可能。

---

## 4. avifenc vs AVIF Encode API

### avifenc コマンドライン引数

#### 基本オプション
- ✅ `-q, --qcolor Q` - color quality (0-100) → `quality`
- ✅ `--qalpha Q` - alpha quality (0-100) → `quality_alpha`
- ✅ `-s, --speed S` - encoder speed (0-10 or 'default') → `speed`
- ✅ `-j, --jobs J` - worker threads ('all' or number) → **未サポート（libavif内部で管理）**
- ⚠️ `--no-overwrite` - never overwrite → **未サポート（アプリケーション層）**
- ⚠️ `-o, --output FILENAME` - output file → **未サポート（アプリケーション層）**

#### 高度なオプション
- ✅ `-l, --lossless` - lossless encoding → `quality=100` で対応
- ✅ `-d, --depth D` - output depth (8, 10, 12) → `bit_depth`
- ✅ `-y, --yuv FORMAT` - YUV format (auto, 444, 422, 420, 400) → `yuv_format`
- ✅ `-p, --premultiply` - premultiply alpha → `premultiply_alpha`
- ✅ `--sharpyuv` - sharp RGB->YUV420 → `sharp_yuv`
- ⚠️ `--stdin` - read y4m from stdin → **未サポート（y4m入力は別設計）**

#### 色空間設定（CICP/nclx）
- ✅ `--cicp, --nclx P/T/M` - CICP values → `color_primaries`, `transfer_characteristics`, `matrix_coefficients`
- ✅ `-r, --range RANGE` - YUV range (limited/full) → `yuv_range`

#### ファイルサイズ最適化
- ✅ `--target-size S` - target file size in bytes → `target_size`

#### 実験的機能
- ⚠️ `--progressive` - progressive rendering → **未サポート（実験的機能）**
- ⚠️ `--layered` - layered AVIF (up to 4 layers) → **未サポート（実験的機能）**

#### グリッド画像
- ⚠️ `-g, --grid MxN` - grid AVIF → **未サポート（グリッド機能未実装）**

#### コーデック選択
- ⚠️ `-c, --codec C` - codec selection → **未サポート（システム層、aom固定）**

#### メタデータ
- ✅ `--exif FILENAME` - EXIF payload → `exif_data`, `exif_size`
- ✅ `--xmp FILENAME` - XMP payload → `xmp_data`, `xmp_size`
- ✅ `--icc FILENAME` - ICC profile → `icc_data`, `icc_size`
- ✅ `--ignore-exif` - ignore embedded EXIF → **デフォルトでコピーしない動作**
- ✅ `--ignore-xmp` - ignore embedded XMP → **デフォルトでコピーしない動作**
- ✅ `--ignore-profile, --ignore-icc` - ignore ICC → **デフォルトでコピーしない動作**

#### 画像シーケンス（アニメーション）
- ⚠️ `--timescale, --fps V` - timescale/fps (default: 30) → `timescale` **（未使用）**
- ⚠️ `-k, --keyframe INTERVAL` - keyframe interval → `keyframe_interval` **（未使用）**
- ⚠️ `--repetition-count N` - repetition count → **未サポート（アニメーション未実装）**

#### 画像プロパティ
- ✅ `--pasp H,V` - pixel aspect ratio → `pasp[2]`
- ✅ `--crop X,Y,W,H` - crop rectangle → `crop[4]`
- ✅ `--clap WN,WD,...` - clean aperture → `clap[8]`
- ✅ `--irot ANGLE` - rotation (0-3) → `irot_angle`
- ✅ `--imir AXIS` - mirroring (0-1) → `imir_axis`
- ✅ `--clli MaxCLL,MaxPALL` - content light level → `clli_max_cll`, `clli_max_pall`

#### タイリング設定
- ✅ `--tilerowslog2 R` - tile rows log2 (0-6) → `tile_rows_log2`
- ✅ `--tilecolslog2 C` - tile cols log2 (0-6) → `tile_cols_log2`
- ⚠️ `--autotiling` - automatic tiling → **未サポート（手動設定のみ）**

#### 品質設定（非推奨）
- ✅ `--min QP` / `--max QP` - quantizer (deprecated) → `min_quantizer`, `max_quantizer`
- ✅ `--minalpha QP` / `--maxalpha QP` - alpha quantizer (deprecated) → `min_quantizer_alpha`, `max_quantizer_alpha`

#### 高度な設定
- ⚠️ `--scaling-mode N[/D]` - frame scaling mode → **未サポート（実験的機能）**
- ⚠️ `--duration D` - frame duration → **未サポート（アニメーション未実装）**
- ⚠️ `-a, --advanced KEY[=VALUE]` - codec-specific options → **未サポート（高度なコーデック設定）**

### 分析結果

#### サポート状況
- **完全対応**: 基本的な品質設定、ビット深度、YUVフォーマット、色空間、メタデータ、画像プロパティ、タイリング
- **未対応（アニメーション）**: `--timescale`, `--keyframe`, `--repetition-count`, `--duration`
  - **理由**: **現在のAPIは静止画のみ対応**（構造体には定義済みだが未使用）
  - **将来対応**: アニメーションAVIF対応は将来実装予定
- **未対応（実験的機能）**: `--progressive`, `--layered`, `--scaling-mode`
  - **理由**: libavifの実験的機能で、安定性が不明
- **未対応（グリッド）**: `-g, --grid`
  - **理由**: グリッド画像機能は別実装が必要
- **未対応（高度な設定）**: `-a, --advanced`
  - **理由**: コーデック固有の設定は抽象化層を超える
- **未対応（システム/CLI）**: `-j, --jobs`, `--no-overwrite`, `-o`, `--stdin`, `-c, --codec`, `--autotiling`

#### 重要なサポート
- **画像プロパティの完全サポート**: `pasp`, `crop`, `clap`, `irot`, `imir`, `clli`
  - これらはAVIF特有の高度な機能で、完全にサポートされている
- **メタデータの柔軟な対応**: EXIF, XMP, ICCを外部ファイルから注入可能

#### 今後の対応
1. **アニメーションAVIF対応**（優先度高）
   - `timescale`, `keyframe_interval`, `repetition-count`の実装
   - avifEncoderAddImage()の繰り返し呼び出しサポート
2. **自動タイリング**（優先度中）
   - `--autotiling`相当の機能実装
3. **グリッド画像**（優先度低）
   - 複数画像を結合したグリッドAVIF生成

#### 結論
avifencの**コア機能（静止画）は完全にサポート**されており、特にAVIF固有の画像プロパティは全て対応。
**アニメーション機能は未実装**だが、構造体には定義済みで将来実装が容易。

---

## 5. avifdec vs AVIF Decode API

### avifdec コマンドライン引数

#### 基本オプション
- ⚠️ `-h, --help` - show help → **未サポート（CLI固有）**
- ⚠️ `-V, --version` - show version → **未サポート（CLI固有）**
- ⚠️ `-j, --jobs J` - worker threads ('all' or number) → **未サポート（libavif内部で管理）**
  - **注**: ライブラリAPIでは `use_threads` で有効化のみ可能
- ⚠️ `-c, --codec C` - codec selection → **未サポート（システム層）**

#### 出力設定
- ✅ `-d, --depth D` - output depth (8 or 16) → **未サポート（PNG出力は別処理）**
  - **注**: デコードAPIは常に元のビット深度を返す（`bit_depth`フィールド）
- ⚠️ `-q, --quality Q` - JPEG quality (0-100) → **未サポート（JPEG出力は別処理）**
- ⚠️ `--png-compress L` - PNG compression (0-9) → **未サポート（PNG出力は別処理）**

#### 色空間処理
- ⚠️ `-u, --upsampling U` - chroma upsampling → **未サポート（libavif内部処理）**
  - **注**: libavifが自動的に最適なアップサンプリングを選択
- ⚠️ `-r, --raw-color` - output raw RGB without alpha multiply → **未サポート（JPEG出力固有）**

#### アニメーション/プログレッシブ
- ⚠️ `--index I` - frame index to decode (0 or 'all') → **未サポート（アニメーション未実装）**
- ⚠️ `--progressive` - progressive image processing → **未サポート（プログレッシブ未実装）**

#### デコード設定
- ⚠️ `--no-strict` - disable strict validation → **未サポート（常に厳格なバリデーション）**

#### メタデータ
- ✅ `--icc FILENAME` - provide ICC profile (implies --ignore-icc) → `icc_data`, `icc_size` **（エンコード時のみ）**
- ✅ `--ignore-icc` - ignore embedded ICC → `ignore_icc`（構造体未定義、**要追加**）
  - **現状**: AVIF DecodeOptionsに `ignore_exif`, `ignore_xmp` はあるが `ignore_icc` がない

#### 安全性設定
- ⚠️ `--size-limit C` - maximum image size in pixels → **未サポート（セキュリティ設定）**
- ⚠️ `--dimension-limit C` - maximum dimension → **未サポート（セキュリティ設定）**

#### 情報表示
- ⚠️ `-i, --info` - display image info instead of saving → **未サポート（CLI機能）**
  - **注**: デコードAPIは常にメタデータ（幅、高さ、ビット深度等）を返す

### 分析結果

#### サポート状況
- **基本対応**: デコード機能、マルチスレッド、メタデータ無視オプション（一部）
- **未対応（出力処理）**: `-d`, `-q`, `--png-compress`
  - **理由**: 出力フォーマット変換（PNG, JPEG等）は出力側で行うべき
  - ライブラリは生ピクセルデータ（RGBA等）を返すのみ
- **未対応（アニメーション）**: `--index`, `--progressive`
  - **理由**: アニメーション/プログレッシブデコード未実装
- **未対応（高度な設定）**: `-u`, `-r`, `--no-strict`, `--size-limit`, `--dimension-limit`
  - **理由**: libavif内部処理、JPEG固有、またはセキュリティ層の機能
- **未対応（CLI固有）**: `-h`, `-V`, `-i`, `-c`

#### 欠落している機能（要追加）
1. **`ignore_icc` オプションの追加**
   - `NextImageAVIFDecodeOptions` に `ignore_icc` フィールドを追加すべき
   - 現在は `ignore_exif`, `ignore_xmp` のみ存在
   - avifdecでは `--ignore-icc` が提供されている

2. **セキュリティ制限オプション**（検討中）
   - `size_limit`, `dimension_limit` の追加を検討
   - DoS攻撃対策として有用

#### avifdecとの重要な違い
- **出力フォーマット**: avifdecはPNG/JPEG出力を直接サポートするが、ライブラリは生ピクセルデータのみ
- **情報表示**: avifdecの `-i, --info` 機能は、ライブラリでは常に返されるメタデータで代替可能
- **色空間変換**: avifdecは様々なアップサンプリングモードを提供するが、ライブラリはlibavifのデフォルト動作に依存

#### 今後の対応
1. **`ignore_icc` の追加**（優先度高）
   - `NextImageAVIFDecodeOptions` に追加
   - C/Goバインディングの更新
2. **アニメーション/プログレッシブデコード**（優先度高）
   - `--index` 相当の機能実装
   - 複数フレームのデコード対応
3. **セキュリティ制限**（優先度中）
   - `size_limit`, `dimension_limit` の実装検討

#### 結論
avifdecの**基本的なデコード機能はサポート**されているが、いくつかの欠落がある。
- **`ignore_icc` オプションが欠落**している（要修正）
- **アニメーション/プログレッシブ機能は未実装**
- 出力フォーマット変換は意図的に非対応（役割分担）

---

## 総合評価

### 全体のサポート状況

#### WebP関連
- **cwebp**: コア機能完全サポート ✅
- **dwebp**: コア機能完全サポート ✅（crop/resize対応が優秀）
- **gif2webp**: 静止画のみサポート ⚠️（アニメーション未実装）

#### AVIF関連
- **avifenc**: コア機能（静止画）完全サポート ✅（画像プロパティ完全対応）
- **avifdec**: 基本機能サポート ⚠️（`ignore_icc` 欠落、アニメーション未実装）

### 優先度別の対応項目

#### 優先度：高（必須）
1. **avifdec: `ignore_icc` オプションの追加**
   - 現在欠落している唯一の基本的なメタデータオプション
   - `NextImageAVIFDecodeOptions` に追加が必要

#### 優先度：中（重要）
2. **アニメーション対応**
   - WebP: アニメーションGIF → アニメーションWebP
   - AVIF: アニメーションAVIF エンコード/デコード
   - 両方とも構造体には定義済み、実装が必要

3. **自動タイリング**（AVIF）
   - `--autotiling` 相当の機能
   - 大きな画像の最適化に有用

#### 優先度：低（検討中）
4. **グリッド画像**（AVIF）
   - 複数画像を結合したグリッドAVIF
   - ニッチな用途

5. **セキュリティ制限**（AVIF）
   - `size_limit`, `dimension_limit`
   - DoS対策として検討

6. **高度なエンコードオプション**
   - WebP: `-mixed`（混合ロスレス/ロッシー）
   - AVIF: `--progressive`, `--layered`, `--advanced`
   - 実験的機能や高度な使用例

### まとめ

libnextimageは、libwebp/libavifの**コマンドラインツールのコア機能をほぼ完全にサポート**しています。

**完全対応**:
- ✅ **cwebp** - 全コア機能完全サポート（crop/resize/blend_alpha/noalpha含む）
- ✅ **dwebp** - 全コア機能完全サポート（crop/resize/flip対応が優秀）
- ✅ **avifenc（静止画）** - AVIF固有の画像プロパティ含め完全対応

**部分対応**:
- ⚠️ **gif2webp** - 静止画のみサポート（アニメーション未実装）
- ⚠️ **avifdec** - 基本機能サポート（`ignore_icc`欠落、アニメーション未実装）

未対応の機能は主に以下のカテゴリ：
1. **CLI固有機能**（version, help, verbose等）- 意図的に非対応
2. **アニメーション機能** - 将来実装予定（構造体には定義済み）
3. **出力フォーマット変換** - 役割分担により非対応（ライブラリは生ピクセルデータを返す）
4. **実験的機能** - 安定性の観点から保留

**早急な対応が必要**: avifdecの`ignore_icc`オプション追加のみ。

### 最近の追加機能（2025-10-18）

#### WebP画像変換機能の完全実装
cwebpの画像変換オプションを完全実装：
- **Crop**: `-crop x y w h` → `crop_x`, `crop_y`, `crop_width`, `crop_height`
- **Resize**: `-resize w h` → `resize_width`, `resize_height`
- **Resize Mode**: `-resize_mode` → `resize_mode` (0=always, 1=up_only, 2=down_only)
- **Blend Alpha**: `-blend_alpha 0xRRGGBB` → `blend_alpha` (背景色との合成)
- **No Alpha**: `-noalpha` → `noalpha` (アルファチャンネル除去)

処理順序はcwebpと同じ：画像読み込み → crop → resize → blend_alpha → エンコード

テストも完備（`golang/webp_advanced_test.go`）：
- ✅ Crop機能（256x256クロップ）
- ✅ Resize機能（200x200リサイズ）
- ✅ Resize Mode（up_only/down_only/always）
- ✅ Crop + Resize組み合わせ
- ✅ Blend Alpha（背景色合成）
- ✅ No Alpha（アルファ除去）

