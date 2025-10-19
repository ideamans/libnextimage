package libnextimage

/*
#include "webp.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

// WebPPreset represents preset types for WebP encoding
type WebPPreset int

const (
	PresetDefault WebPPreset = 0 // default preset
	PresetPicture WebPPreset = 1 // digital picture, like portrait
	PresetPhoto   WebPPreset = 2 // outdoor photograph
	PresetDrawing WebPPreset = 3 // hand or line drawing
	PresetIcon    WebPPreset = 4 // small-sized colorful images
	PresetText    WebPPreset = 5 // text-like
)

// WebPImageHint represents image type hint for WebP encoding
type WebPImageHint int

const (
	HintDefault WebPImageHint = 0 // default hint
	HintPicture WebPImageHint = 1 // digital picture, like portrait
	HintPhoto   WebPImageHint = 2 // outdoor photograph
	HintGraph   WebPImageHint = 3 // discrete tone image (graph, map-tile)
)

// WebPEncodeOptions represents WebP encoding options (全cwebpオプションに対応)
type WebPEncodeOptions struct {
	// 基本設定
	Quality  float32 // 0-100, default 75
	Lossless bool    // default false
	Method   int     // 0-6, default 4 (quality/speed trade-off)

	// プリセット
	Preset         int           // -1=none (default), or preset type (0-5)
	ImageHint      WebPImageHint // image type hint, default 0
	LosslessPreset int           // -1=don't use (default), 0-9=use preset (0=fast, 9=best)

	// ターゲット設定
	TargetSize int     // target size in bytes (0 = disabled)
	TargetPSNR float32 // target PSNR (0 = disabled)

	// セグメント/フィルタ設定
	Segments        int  // 1-4, number of segments, default 4
	SNSStrength     int  // 0-100, spatial noise shaping, default 50
	FilterStrength  int  // 0-100, filter strength, default 60
	FilterSharpness int  // 0-7, filter sharpness, default 0
	FilterType      int  // 0=simple, 1=strong, default 1
	Autofilter      bool // auto-adjust filter strength, default false

	// アルファチャンネル設定
	AlphaCompression bool // compress alpha channel, default true
	AlphaFiltering   int  // 0=none, 1=fast, 2=best, default 1
	AlphaQuality     int  // 0-100, alpha compression quality, default 100

	// エントロピー設定
	Pass int // 1-10, entropy-analysis passes, default 1

	// その他の設定
	ShowCompressed   bool // export compressed picture, default false
	Preprocessing    int  // 0=none, 1=segment-smooth, 2=pseudo-random dithering
	Partitions       int  // 0-3, log2(number of token partitions), default 0
	PartitionLimit   int  // 0-100, quality degradation allowed, default 0
	EmulateJPEGSize  bool // match JPEG size, default false
	ThreadLevel      bool // use multi-threading, default false
	LowMemory        bool // reduce memory usage, default false
	NearLossless     int  // -1=not set (default), 0-100=use near lossless (auto-enables lossless mode)
	Exact            bool // preserve RGB in transparent area, default false
	UseDeltaPalette  bool // reserved, default false
	UseSharpYUV      bool // sharp RGB->YUV conversion, default false
	QMin             int  // 0-100, minimum permissible quality, default 0
	QMax             int  // 0-100, maximum permissible quality, default 100

	// メタデータ設定
	KeepMetadata int // -1=default, 0=none, 1=exif, 2=icc, 3=xmp, 4=all

	// 画像変換設定 (cwebp -crop, -resize)
	CropX      int // crop rectangle x (-crop x y w h), -1=disabled
	CropY      int // crop rectangle y
	CropWidth  int // crop rectangle width
	CropHeight int // crop rectangle height

	ResizeWidth  int // resize width (-resize w h), -1=disabled
	ResizeHeight int // resize height
	ResizeMode   int // 0=always (default), 1=up_only, 2=down_only

	// アルファチャンネル特殊処理
	BlendAlpha uint32 // blend alpha against background color (0xRRGGBB), 0xFFFFFFFF=disabled
	NoAlpha    bool   // discard alpha channel, default false

	// アニメーション設定 (gif2webp, WebPAnimEncoder)
	AllowMixed        bool // allow mixed lossy/lossless, default false
	MinimizeSize      bool // minimize output size (slow), default false
	Kmin              int  // min distance between key frames, -1=auto (default: lossless=9, lossy=3)
	Kmax              int  // max distance between key frames, -1=auto (default: lossless=17, lossy=5)
	AnimLoopCount     int  // animation loop count, 0=infinite (default: 0)
	LoopCompatibility bool // Chrome M62 compatibility mode, default false
}

// WebPDecodeOptions represents WebP decoding options (全dwebpオプションに対応)
type WebPDecodeOptions struct {
	// 基本設定
	UseThreads        bool        // enable multi-threading (-mt)
	BypassFiltering   bool        // disable in-loop filtering (-nofilter)
	NoFancyUpsampling bool        // use faster pointwise upsampler (-nofancy)
	Format            PixelFormat // desired pixel format (default: RGBA)

	// ディザリング設定
	NoDither       bool // disable dithering (-nodither)
	DitherStrength int  // 0-100, dithering strength (-dither), default 100
	AlphaDither    bool // use alpha-plane dithering (-alpha_dither)

	// 画像操作
	CropX      int  // crop rectangle x (-crop x y w h)
	CropY      int  // crop rectangle y
	CropWidth  int  // crop rectangle width
	CropHeight int  // crop rectangle height
	UseCrop    bool // enable cropping

	ResizeWidth  int  // resize width (-resize w h)
	ResizeHeight int  // resize height
	UseResize    bool // enable resizing

	Flip bool // flip vertically (-flip)

	// 特殊モード
	AlphaOnly   bool // save only alpha plane (-alpha)
	Incremental bool // use incremental decoding (-incremental)
}

// DefaultWebPEncodeOptions returns default WebP encoding options
func DefaultWebPEncodeOptions() WebPEncodeOptions {
	return WebPEncodeOptions{
		// 基本設定
		Quality:  75.0,
		Lossless: false,
		Method:   4,

		// プリセット
		Preset:         -1, // -1 = none (don't use preset)
		ImageHint:      HintDefault,
		LosslessPreset: -1, // -1 = don't use preset

		// ターゲット設定
		TargetSize: 0,
		TargetPSNR: 0.0,

		// セグメント/フィルタ設定
		Segments:        4,
		SNSStrength:     50,
		FilterStrength:  60,
		FilterSharpness: 0,
		FilterType:      1, // strong filter
		Autofilter:      false,

		// アルファチャンネル設定
		AlphaCompression: true,
		AlphaFiltering:   1, // fast
		AlphaQuality:     100,

		// エントロピー設定
		Pass: 1,

		// その他の設定
		ShowCompressed:   false,
		Preprocessing:    0,
		Partitions:       0,
		PartitionLimit:   0,
		EmulateJPEGSize:  false,
		ThreadLevel:      false,
		LowMemory:        false,
		NearLossless:     -1, // -1 = not set
		Exact:            false,
		UseDeltaPalette:  false,
		UseSharpYUV:      false,
		QMin:             0,
		QMax:             100,

		// メタデータ設定
		KeepMetadata: -1, // -1 = default behavior

		// 画像変換設定
		CropX:      -1, // -1 = disabled
		CropY:      -1,
		CropWidth:  -1,
		CropHeight: -1,
		ResizeWidth:  -1, // -1 = disabled
		ResizeHeight: -1,
		ResizeMode:   0, // 0 = always (default)

		// アルファチャンネル特殊処理
		BlendAlpha: 0xFFFFFFFF, // 0xFFFFFFFF = disabled
		NoAlpha:    false,

		// アニメーション設定
		AllowMixed:        false,
		MinimizeSize:      false,
		Kmin:              -1, // -1 = auto
		Kmax:              -1, // -1 = auto
		AnimLoopCount:     0,  // 0 = infinite loop
		LoopCompatibility: false,
	}
}

// DefaultWebPDecodeOptions returns default WebP decoding options
func DefaultWebPDecodeOptions() WebPDecodeOptions {
	return WebPDecodeOptions{
		// 基本設定
		UseThreads:        false,
		BypassFiltering:   false,
		NoFancyUpsampling: false,
		Format:            FormatRGBA,

		// ディザリング設定
		NoDither:       false,
		DitherStrength: 100,
		AlphaDither:    false,

		// 画像操作
		CropX:      0,
		CropY:      0,
		CropWidth:  0,
		CropHeight: 0,
		UseCrop:    false,

		ResizeWidth:  0,
		ResizeHeight: 0,
		UseResize:    false,

		Flip: false,

		// 特殊モード
		AlphaOnly:   false,
		Incremental: false,
	}
}

// convertEncodeOptions converts Go options to C options
func convertEncodeOptions(opts WebPEncodeOptions) C.NextImageWebPEncodeOptions {
	var cOpts C.NextImageWebPEncodeOptions
	C.nextimage_webp_default_encode_options(&cOpts)

	// 基本設定
	cOpts.quality = C.float(opts.Quality)
	if opts.Lossless {
		cOpts.lossless = 1
	} else {
		cOpts.lossless = 0
	}
	cOpts.method = C.int(opts.Method)

	// プリセット
	cOpts.preset = C.NextImageWebPPreset(opts.Preset)
	cOpts.image_hint = C.NextImageWebPImageHint(opts.ImageHint)
	cOpts.lossless_preset = C.int(opts.LosslessPreset)

	// ターゲット設定
	cOpts.target_size = C.int(opts.TargetSize)
	cOpts.target_psnr = C.float(opts.TargetPSNR)

	// セグメント/フィルタ設定
	cOpts.segments = C.int(opts.Segments)
	cOpts.sns_strength = C.int(opts.SNSStrength)
	cOpts.filter_strength = C.int(opts.FilterStrength)
	cOpts.filter_sharpness = C.int(opts.FilterSharpness)
	cOpts.filter_type = C.int(opts.FilterType)
	if opts.Autofilter {
		cOpts.autofilter = 1
	} else {
		cOpts.autofilter = 0
	}

	// アルファチャンネル設定
	if opts.AlphaCompression {
		cOpts.alpha_compression = 1
	} else {
		cOpts.alpha_compression = 0
	}
	cOpts.alpha_filtering = C.int(opts.AlphaFiltering)
	cOpts.alpha_quality = C.int(opts.AlphaQuality)

	// エントロピー設定
	cOpts.pass = C.int(opts.Pass)

	// その他の設定
	if opts.ShowCompressed {
		cOpts.show_compressed = 1
	} else {
		cOpts.show_compressed = 0
	}
	cOpts.preprocessing = C.int(opts.Preprocessing)
	cOpts.partitions = C.int(opts.Partitions)
	cOpts.partition_limit = C.int(opts.PartitionLimit)
	if opts.EmulateJPEGSize {
		cOpts.emulate_jpeg_size = 1
	} else {
		cOpts.emulate_jpeg_size = 0
	}
	if opts.ThreadLevel {
		cOpts.thread_level = 1
	} else {
		cOpts.thread_level = 0
	}
	if opts.LowMemory {
		cOpts.low_memory = 1
	} else {
		cOpts.low_memory = 0
	}
	cOpts.near_lossless = C.int(opts.NearLossless)
	if opts.Exact {
		cOpts.exact = 1
	} else {
		cOpts.exact = 0
	}
	if opts.UseDeltaPalette {
		cOpts.use_delta_palette = 1
	} else {
		cOpts.use_delta_palette = 0
	}
	if opts.UseSharpYUV {
		cOpts.use_sharp_yuv = 1
	} else {
		cOpts.use_sharp_yuv = 0
	}
	cOpts.qmin = C.int(opts.QMin)
	cOpts.qmax = C.int(opts.QMax)

	// メタデータ設定
	cOpts.keep_metadata = C.int(opts.KeepMetadata)

	// 画像変換設定
	cOpts.crop_x = C.int(opts.CropX)
	cOpts.crop_y = C.int(opts.CropY)
	cOpts.crop_width = C.int(opts.CropWidth)
	cOpts.crop_height = C.int(opts.CropHeight)
	cOpts.resize_width = C.int(opts.ResizeWidth)
	cOpts.resize_height = C.int(opts.ResizeHeight)
	cOpts.resize_mode = C.int(opts.ResizeMode)

	// アルファチャンネル特殊処理
	cOpts.blend_alpha = C.uint32_t(opts.BlendAlpha)
	if opts.NoAlpha {
		cOpts.noalpha = 1
	} else {
		cOpts.noalpha = 0
	}

	// アニメーション設定
	if opts.AllowMixed {
		cOpts.allow_mixed = 1
	} else {
		cOpts.allow_mixed = 0
	}
	if opts.MinimizeSize {
		cOpts.minimize_size = 1
	} else {
		cOpts.minimize_size = 0
	}
	cOpts.kmin = C.int(opts.Kmin)
	cOpts.kmax = C.int(opts.Kmax)
	cOpts.anim_loop_count = C.int(opts.AnimLoopCount)
	if opts.LoopCompatibility {
		cOpts.loop_compatibility = 1
	} else {
		cOpts.loop_compatibility = 0
	}

	return cOpts
}

// convertDecodeOptions converts Go options to C options
func convertDecodeOptions(opts WebPDecodeOptions) C.NextImageWebPDecodeOptions {
	var cOpts C.NextImageWebPDecodeOptions
	C.nextimage_webp_default_decode_options(&cOpts)

	// 基本設定
	if opts.UseThreads {
		cOpts.use_threads = 1
	} else {
		cOpts.use_threads = 0
	}
	if opts.BypassFiltering {
		cOpts.bypass_filtering = 1
	} else {
		cOpts.bypass_filtering = 0
	}
	if opts.NoFancyUpsampling {
		cOpts.no_fancy_upsampling = 1
	} else {
		cOpts.no_fancy_upsampling = 0
	}
	cOpts.format = C.NextImagePixelFormat(opts.Format)

	// ディザリング設定
	if opts.NoDither {
		cOpts.no_dither = 1
	} else {
		cOpts.no_dither = 0
	}
	cOpts.dither_strength = C.int(opts.DitherStrength)
	if opts.AlphaDither {
		cOpts.alpha_dither = 1
	} else {
		cOpts.alpha_dither = 0
	}

	// 画像操作
	cOpts.crop_x = C.int(opts.CropX)
	cOpts.crop_y = C.int(opts.CropY)
	cOpts.crop_width = C.int(opts.CropWidth)
	cOpts.crop_height = C.int(opts.CropHeight)
	if opts.UseCrop {
		cOpts.use_crop = 1
	} else {
		cOpts.use_crop = 0
	}

	cOpts.resize_width = C.int(opts.ResizeWidth)
	cOpts.resize_height = C.int(opts.ResizeHeight)
	if opts.UseResize {
		cOpts.use_resize = 1
	} else {
		cOpts.use_resize = 0
	}

	if opts.Flip {
		cOpts.flip = 1
	} else {
		cOpts.flip = 0
	}

	// 特殊モード
	if opts.AlphaOnly {
		cOpts.alpha_only = 1
	} else {
		cOpts.alpha_only = 0
	}
	if opts.Incremental {
		cOpts.incremental = 1
	} else {
		cOpts.incremental = 0
	}

	return cOpts
}

// WebPEncodeBytes encodes image file data (JPEG, PNG, GIF, etc.) to WebP format
// This is equivalent to the cwebp command-line tool.
// The input data should be a complete image file (JPEG, PNG, GIF, TIFF, WebP, etc.)
// not raw pixel data.
func WebPEncodeBytes(imageFileData []byte, opts WebPEncodeOptions) ([]byte, error) {
	clearError()

	if len(imageFileData) == 0 {
		return nil, fmt.Errorf("webp encode: empty input data")
	}

	cOpts := convertEncodeOptions(opts)
	var encoded C.NextImageBuffer

	status := C.nextimage_webp_encode_alloc(
		(*C.uint8_t)(unsafe.Pointer(&imageFileData[0])),
		C.size_t(len(imageFileData)),
		&cOpts,
		&encoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp encode")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(encoded.data), C.int(encoded.size))

	// Free C buffer
	freeEncodeBuffer(&encoded)

	return result, nil
}

// WebPEncodeFile encodes an image file to WebP format
// This reads the image file (JPEG, PNG, GIF, etc.) and converts it to WebP.
func WebPEncodeFile(inputPath string, opts WebPEncodeOptions) ([]byte, error) {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("webp encode file: %w", err)
	}

	return WebPEncodeBytes(data, opts)
}

// WebPDecodeBytes decodes WebP data to pixel data
func WebPDecodeBytes(webpData []byte, opts WebPDecodeOptions) (*DecodedImage, error) {
	clearError()

	if len(webpData) == 0 {
		return nil, fmt.Errorf("webp decode: empty input data")
	}

	cOpts := convertDecodeOptions(opts)
	var decoded C.NextImageDecodeBuffer

	status := C.nextimage_webp_decode_alloc(
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&cOpts,
		&decoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp decode")
	}

	// Convert to Go struct
	img := convertDecodeBuffer(&decoded)

	// Free C buffer
	freeDecodeBuffer(&decoded)

	return img, nil
}

// WebPDecodeFile decodes a WebP file to pixel data
func WebPDecodeFile(inputPath string, opts WebPDecodeOptions) (*DecodedImage, error) {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("webp decode file: %w", err)
	}

	return WebPDecodeBytes(data, opts)
}

// WebPDecodeSize calculates required buffer size for decoding
func WebPDecodeSize(webpData []byte) (width, height int, requiredSize int, err error) {
	clearError()

	if len(webpData) == 0 {
		return 0, 0, 0, fmt.Errorf("webp decode size: empty input data")
	}

	var w, h C.int
	var size C.size_t

	status := C.nextimage_webp_decode_size(
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&w,
		&h,
		&size,
	)

	if status != C.NEXTIMAGE_OK {
		return 0, 0, 0, makeError(status, "webp decode size")
	}

	return int(w), int(h), int(size), nil
}

// WebPDecodeInto decodes WebP data into a user-provided buffer
func WebPDecodeInto(webpData []byte, buffer []byte, opts WebPDecodeOptions) (*DecodedImage, error) {
	clearError()

	if len(webpData) == 0 {
		return nil, fmt.Errorf("webp decode into: empty input data")
	}
	if len(buffer) == 0 {
		return nil, fmt.Errorf("webp decode into: empty buffer")
	}

	// Due to CGO pointer rules, we cannot pass Go buffer directly to C
	// So we allocate in C first, then copy to user buffer
	img, err := WebPDecodeBytes(webpData, opts)
	if err != nil {
		return nil, fmt.Errorf("webp decode into: %w", err)
	}

	// Check buffer size
	if len(buffer) < len(img.Data) {
		return nil, fmt.Errorf("webp decode into: buffer too small (need %d bytes, have %d bytes)", len(img.Data), len(buffer))
	}

	// Copy decoded data to user buffer
	copy(buffer, img.Data)

	// Return metadata with user buffer reference
	return &DecodedImage{
		Data:     buffer[:len(img.Data)],
		Stride:   img.Stride,
		UPlane:   nil,
		UStride:  0,
		VPlane:   nil,
		VStride:  0,
		Width:    img.Width,
		Height:   img.Height,
		BitDepth: img.BitDepth,
		Format:   img.Format,
	}, nil
}

// GIF2WebP converts GIF data to WebP format
func GIF2WebP(gifData []byte, opts WebPEncodeOptions) ([]byte, error) {
	clearError()

	if len(gifData) == 0 {
		return nil, fmt.Errorf("gif2webp: empty input data")
	}

	cOpts := convertEncodeOptions(opts)
	var encoded C.NextImageBuffer

	status := C.nextimage_gif2webp_alloc(
		(*C.uint8_t)(unsafe.Pointer(&gifData[0])),
		C.size_t(len(gifData)),
		&cOpts,
		&encoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "gif2webp")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(encoded.data), C.int(encoded.size))

	// Free C buffer
	freeEncodeBuffer(&encoded)

	return result, nil
}

// WebP2GIF converts WebP data to GIF format
func WebP2GIF(webpData []byte) ([]byte, error) {
	clearError()

	if len(webpData) == 0 {
		return nil, fmt.Errorf("webp2gif: empty input data")
	}

	var encoded C.NextImageBuffer

	status := C.nextimage_webp2gif_alloc(
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&encoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp2gif")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(encoded.data), C.int(encoded.size))

	// Free C buffer
	freeEncodeBuffer(&encoded)

	return result, nil
}
