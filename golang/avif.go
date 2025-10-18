package libnextimage

/*
#include "avif.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// AVIFEncodeOptions represents AVIF encoding options
type AVIFEncodeOptions struct {
	Quality           int // 0-100, default 50 (mapped to AVIF quantizer 0-63)
	Speed             int // 0-10, default 6 (0=slowest/best, 10=fastest/worst)
	MinQuantizer      int // 0-63, default 0 (best quality)
	MaxQuantizer      int // 0-63, default 63 (worst quality)
	MinQuantizerAlpha int // 0-63, default 0
	MaxQuantizerAlpha int // 0-63, default 63
	EnableAlpha       bool
	BitDepth          int // 8, 10, or 12 (default: 8)
	YUVFormat         int // 0=444, 1=422, 2=420, 3=400 (default: 420)
	TileRowsLog2      int // 0-6, default 0
	TileColsLog2      int // 0-6, default 0
}

// AVIFDecodeOptions represents AVIF decoding options
type AVIFDecodeOptions struct {
	UseThreads bool
	Format     PixelFormat
	IgnoreExif bool
	IgnoreXMP  bool
}

// DefaultAVIFEncodeOptions returns default AVIF encoding options
func DefaultAVIFEncodeOptions() AVIFEncodeOptions {
	var opts C.NextImageAVIFEncodeOptions
	C.nextimage_avif_default_encode_options(&opts)

	return AVIFEncodeOptions{
		Quality:           int(opts.quality),
		Speed:             int(opts.speed),
		MinQuantizer:      int(opts.min_quantizer),
		MaxQuantizer:      int(opts.max_quantizer),
		MinQuantizerAlpha: int(opts.min_quantizer_alpha),
		MaxQuantizerAlpha: int(opts.max_quantizer_alpha),
		EnableAlpha:       opts.enable_alpha != 0,
		BitDepth:          int(opts.bit_depth),
		YUVFormat:         int(opts.yuv_format),
		TileRowsLog2:      int(opts.tile_rows_log2),
		TileColsLog2:      int(opts.tile_cols_log2),
	}
}

// DefaultAVIFDecodeOptions returns default AVIF decoding options
func DefaultAVIFDecodeOptions() AVIFDecodeOptions {
	var opts C.NextImageAVIFDecodeOptions
	C.nextimage_avif_default_decode_options(&opts)

	return AVIFDecodeOptions{
		UseThreads: opts.use_threads != 0,
		Format:     PixelFormat(opts.format),
		IgnoreExif: opts.ignore_exif != 0,
		IgnoreXMP:  opts.ignore_xmp != 0,
	}
}

// toCEncodeOptions converts Go options to C options
func (opts *AVIFEncodeOptions) toCEncodeOptions() C.NextImageAVIFEncodeOptions {
	var copts C.NextImageAVIFEncodeOptions
	copts.quality = C.int(opts.Quality)
	copts.speed = C.int(opts.Speed)
	copts.min_quantizer = C.int(opts.MinQuantizer)
	copts.max_quantizer = C.int(opts.MaxQuantizer)
	copts.min_quantizer_alpha = C.int(opts.MinQuantizerAlpha)
	copts.max_quantizer_alpha = C.int(opts.MaxQuantizerAlpha)
	if opts.EnableAlpha {
		copts.enable_alpha = 1
	} else {
		copts.enable_alpha = 0
	}
	copts.bit_depth = C.int(opts.BitDepth)
	copts.yuv_format = C.int(opts.YUVFormat)
	copts.tile_rows_log2 = C.int(opts.TileRowsLog2)
	copts.tile_cols_log2 = C.int(opts.TileColsLog2)
	return copts
}

// toCDecodeOptions converts Go options to C options
func (opts *AVIFDecodeOptions) toCDecodeOptions() C.NextImageAVIFDecodeOptions {
	var copts C.NextImageAVIFDecodeOptions
	if opts.UseThreads {
		copts.use_threads = 1
	} else {
		copts.use_threads = 0
	}
	copts.format = C.NextImagePixelFormat(opts.Format)
	if opts.IgnoreExif {
		copts.ignore_exif = 1
	} else {
		copts.ignore_exif = 0
	}
	if opts.IgnoreXMP {
		copts.ignore_xmp = 1
	} else {
		copts.ignore_xmp = 0
	}
	return copts
}

// AVIFEncodeBytes encodes RGBA/RGB/BGRA image data to AVIF format
func AVIFEncodeBytes(
	inputData []byte,
	width, height int,
	format PixelFormat,
	options AVIFEncodeOptions,
) ([]byte, error) {
	clearError()

	if len(inputData) == 0 {
		return nil, fmt.Errorf("avif encode: empty input data")
	}

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("avif encode: invalid dimensions %dx%d", width, height)
	}

	// Convert options
	copts := options.toCEncodeOptions()

	// Encode
	var output C.NextImageEncodeBuffer
	status := C.nextimage_avif_encode_alloc(
		(*C.uint8_t)(unsafe.Pointer(&inputData[0])),
		C.size_t(len(inputData)),
		C.int(width),
		C.int(height),
		C.NextImagePixelFormat(format),
		&copts,
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "avif encode")
	}

	// Copy output data to Go slice
	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))

	// Free C buffer
	freeEncodeBuffer(&output)

	return result, nil
}

// AVIFDecodeBytes decodes AVIF data to RGBA/RGB/BGRA format
func AVIFDecodeBytes(
	avifData []byte,
	options AVIFDecodeOptions,
) (*DecodedImage, error) {
	clearError()

	if len(avifData) == 0 {
		return nil, fmt.Errorf("avif decode: empty input data")
	}

	// Convert options
	copts := options.toCDecodeOptions()

	// Decode
	var output C.NextImageDecodeBuffer
	status := C.nextimage_avif_decode_alloc(
		(*C.uint8_t)(unsafe.Pointer(&avifData[0])),
		C.size_t(len(avifData)),
		&copts,
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "avif decode")
	}

	// Convert to Go structure
	img := convertDecodeBuffer(&output)

	// Free C buffer
	freeDecodeBuffer(&output)

	return img, nil
}

// AVIFDecodeSize returns the dimensions and required buffer size for decoding an AVIF image
func AVIFDecodeSize(avifData []byte) (width, height, bitDepth int, requiredSize int, err error) {
	clearError()

	if len(avifData) == 0 {
		return 0, 0, 0, 0, fmt.Errorf("avif decode size: empty input data")
	}

	var w, h, depth C.int
	var size C.size_t

	status := C.nextimage_avif_decode_size(
		(*C.uint8_t)(unsafe.Pointer(&avifData[0])),
		C.size_t(len(avifData)),
		&w,
		&h,
		&depth,
		&size,
	)

	if status != C.NEXTIMAGE_OK {
		return 0, 0, 0, 0, makeError(status, "avif decode size")
	}

	return int(w), int(h), int(depth), int(size), nil
}
