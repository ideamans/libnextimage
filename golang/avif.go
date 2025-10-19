package libnextimage

/*
#include "avif.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

// AVIFEncodeOptions represents AVIF encoding options
type AVIFEncodeOptions struct {
	// Quality settings
	Quality      int // 0-100, default 60 (for color/YUV)
	QualityAlpha int // 0-100, default -1 (use Quality if -1)
	Speed        int // 0-10, default 6 (0=slowest/best, 10=fastest/worst)

	// Deprecated quantizer settings (for backward compatibility)
	MinQuantizer      int // 0-63, default -1 (use Quality instead)
	MaxQuantizer      int // 0-63, default -1 (use Quality instead)
	MinQuantizerAlpha int // 0-63, default -1 (use QualityAlpha instead)
	MaxQuantizerAlpha int // 0-63, default -1 (use QualityAlpha instead)

	// Format settings
	BitDepth  int // 8, 10, or 12 (default: 8)
	YUVFormat int // 0=444, 1=422, 2=420, 3=400 (default: 444)
	YUVRange  int // 0=limited, 1=full (default: 1=full)

	// Alpha settings
	EnableAlpha       bool
	PremultiplyAlpha  bool // Premultiply color by alpha

	// Tiling settings
	TileRowsLog2 int // 0-6, default 0
	TileColsLog2 int // 0-6, default 0

	// CICP (nclx) color settings
	ColorPrimaries          int // CICP color primaries, -1=auto
	TransferCharacteristics int // CICP transfer, -1=auto
	MatrixCoefficients      int // CICP matrix, -1=auto

	// Advanced settings
	SharpYUV   bool // Use sharp RGB->YUV conversion
	TargetSize int  // Target file size in bytes, 0=disabled

	// Metadata settings
	ExifData []byte // EXIF metadata bytes (nil=no EXIF)
	XMPData  []byte // XMP metadata bytes (nil=no XMP)
	ICCData  []byte // ICC profile bytes (nil=no ICC)

	// Transformation settings
	IRotAngle int // Image rotation: 0-3 (90 * angle degrees anti-clockwise), -1=disabled
	IMirAxis  int // Image mirror: 0=vertical, 1=horizontal, -1=disabled

	// Pixel aspect ratio (pasp) - array[2]: [h_spacing, v_spacing]
	PASP [2]int // -1=disabled, otherwise [h_spacing, v_spacing]

	// Crop rectangle (simpler interface) - array[4]: [x, y, width, height]
	// This will be converted to clap using avifCleanApertureBoxFromCropRect
	Crop [4]int // -1=disabled, otherwise [x, y, width, height]

	// Clean aperture (clap) - array[8]: [wN,wD, hN,hD, hOffN,hOffD, vOffN,vOffD]
	// Use this for direct clap values, or use Crop[] for simpler interface
	CLAP [8]int // -1=disabled, otherwise [widthN,widthD, heightN,heightD, horizOffN,horizOffD, vertOffN,vertOffD]

	// Content light level information (clli) - array[2]: [maxCLL, maxPALL]
	CLLI [2]int // -1=disabled, otherwise [maxCLL, maxPALL]

	// Animation settings (for future use)
	Timescale        int // Timescale/fps for animations (default: 30)
	KeyframeInterval int // Max keyframe interval (default: 0=disabled)
}

// ChromaUpsampling represents chroma upsampling mode for YUV to RGB conversion
type ChromaUpsampling int

const (
	ChromaUpsamplingAutomatic  ChromaUpsampling = 0 // Automatic (default)
	ChromaUpsamplingFastest    ChromaUpsampling = 1 // Fastest (nearest neighbor)
	ChromaUpsamplingBestQuality ChromaUpsampling = 2 // Best quality (bilinear)
	ChromaUpsamplingNearest    ChromaUpsampling = 3 // Nearest neighbor
	ChromaUpsamplingBilinear   ChromaUpsampling = 4 // Bilinear
)

// AVIFDecodeOptions represents AVIF decoding options
type AVIFDecodeOptions struct {
	UseThreads bool
	Format     PixelFormat
	IgnoreExif bool
	IgnoreXMP  bool
	IgnoreICC  bool

	// Security limits
	ImageSizeLimit      uint32 // Maximum image size in total pixels (default: 268435456)
	ImageDimensionLimit uint32 // Maximum image dimension (width or height), 0=ignore (default: 32768)

	// Validation flags
	StrictFlags int // Strict validation flags: 0=disabled, 1=enabled (default: 1)

	// Chroma upsampling (for YUV to RGB conversion)
	ChromaUpsampling ChromaUpsampling // 0=automatic (default), 1=fastest, 2=best_quality, 3=nearest, 4=bilinear
}

// DefaultAVIFEncodeOptions returns default AVIF encoding options
func DefaultAVIFEncodeOptions() AVIFEncodeOptions {
	var opts C.NextImageAVIFEncodeOptions
	C.nextimage_avif_default_encode_options(&opts)

	return AVIFEncodeOptions{
		Quality:                 int(opts.quality),
		QualityAlpha:            int(opts.quality_alpha),
		Speed:                   int(opts.speed),
		MinQuantizer:            int(opts.min_quantizer),
		MaxQuantizer:            int(opts.max_quantizer),
		MinQuantizerAlpha:       int(opts.min_quantizer_alpha),
		MaxQuantizerAlpha:       int(opts.max_quantizer_alpha),
		BitDepth:                int(opts.bit_depth),
		YUVFormat:               int(opts.yuv_format),
		YUVRange:                int(opts.yuv_range),
		EnableAlpha:             opts.enable_alpha != 0,
		PremultiplyAlpha:        opts.premultiply_alpha != 0,
		TileRowsLog2:            int(opts.tile_rows_log2),
		TileColsLog2:            int(opts.tile_cols_log2),
		ColorPrimaries:          int(opts.color_primaries),
		TransferCharacteristics: int(opts.transfer_characteristics),
		MatrixCoefficients:      int(opts.matrix_coefficients),
		SharpYUV:     opts.sharp_yuv != 0,
		TargetSize:   int(opts.target_size),
		IRotAngle:    int(opts.irot_angle),
		IMirAxis:     int(opts.imir_axis),
		PASP:         [2]int{int(opts.pasp[0]), int(opts.pasp[1])},
		Crop:         [4]int{int(opts.crop[0]), int(opts.crop[1]), int(opts.crop[2]), int(opts.crop[3])},
		CLAP:         [8]int{int(opts.clap[0]), int(opts.clap[1]), int(opts.clap[2]), int(opts.clap[3]), int(opts.clap[4]), int(opts.clap[5]), int(opts.clap[6]), int(opts.clap[7])},
		CLLI:         [2]int{int(opts.clli_max_cll), int(opts.clli_max_pall)},
		Timescale:    int(opts.timescale),
		KeyframeInterval: int(opts.keyframe_interval),
	}
}

// DefaultAVIFDecodeOptions returns default AVIF decoding options
func DefaultAVIFDecodeOptions() AVIFDecodeOptions {
	var opts C.NextImageAVIFDecodeOptions
	C.nextimage_avif_default_decode_options(&opts)

	return AVIFDecodeOptions{
		UseThreads:          opts.use_threads != 0,
		Format:              PixelFormat(opts.format),
		IgnoreExif:          opts.ignore_exif != 0,
		IgnoreXMP:           opts.ignore_xmp != 0,
		IgnoreICC:           opts.ignore_icc != 0,
		ImageSizeLimit:      uint32(opts.image_size_limit),
		ImageDimensionLimit: uint32(opts.image_dimension_limit),
		StrictFlags:         int(opts.strict_flags),
		ChromaUpsampling:    ChromaUpsampling(opts.chroma_upsampling),
	}
}

// toCEncodeOptions converts Go options to C options
func (opts *AVIFEncodeOptions) toCEncodeOptions() C.NextImageAVIFEncodeOptions {
	var copts C.NextImageAVIFEncodeOptions

	// Quality settings
	copts.quality = C.int(opts.Quality)
	copts.quality_alpha = C.int(opts.QualityAlpha)
	copts.speed = C.int(opts.Speed)

	// Deprecated quantizer settings
	copts.min_quantizer = C.int(opts.MinQuantizer)
	copts.max_quantizer = C.int(opts.MaxQuantizer)
	copts.min_quantizer_alpha = C.int(opts.MinQuantizerAlpha)
	copts.max_quantizer_alpha = C.int(opts.MaxQuantizerAlpha)

	// Format settings
	copts.bit_depth = C.int(opts.BitDepth)
	copts.yuv_format = C.int(opts.YUVFormat)
	copts.yuv_range = C.int(opts.YUVRange)

	// Alpha settings
	if opts.EnableAlpha {
		copts.enable_alpha = 1
	}
	if opts.PremultiplyAlpha {
		copts.premultiply_alpha = 1
	}

	// Tiling settings
	copts.tile_rows_log2 = C.int(opts.TileRowsLog2)
	copts.tile_cols_log2 = C.int(opts.TileColsLog2)

	// CICP color settings
	copts.color_primaries = C.int(opts.ColorPrimaries)
	copts.transfer_characteristics = C.int(opts.TransferCharacteristics)
	copts.matrix_coefficients = C.int(opts.MatrixCoefficients)

	// Advanced settings
	if opts.SharpYUV {
		copts.sharp_yuv = 1
	}
	copts.target_size = C.int(opts.TargetSize)

	// Metadata settings - will be set in the caller to avoid Go pointer issues
	// The caller must set these pointers and manage their lifetime

	// Transformation settings
	copts.irot_angle = C.int(opts.IRotAngle)
	copts.imir_axis = C.int(opts.IMirAxis)

	// Pixel aspect ratio (pasp)
	copts.pasp[0] = C.int(opts.PASP[0])
	copts.pasp[1] = C.int(opts.PASP[1])

	// Crop rectangle
	copts.crop[0] = C.int(opts.Crop[0])
	copts.crop[1] = C.int(opts.Crop[1])
	copts.crop[2] = C.int(opts.Crop[2])
	copts.crop[3] = C.int(opts.Crop[3])

	// Clean aperture (clap)
	copts.clap[0] = C.int(opts.CLAP[0])
	copts.clap[1] = C.int(opts.CLAP[1])
	copts.clap[2] = C.int(opts.CLAP[2])
	copts.clap[3] = C.int(opts.CLAP[3])
	copts.clap[4] = C.int(opts.CLAP[4])
	copts.clap[5] = C.int(opts.CLAP[5])
	copts.clap[6] = C.int(opts.CLAP[6])
	copts.clap[7] = C.int(opts.CLAP[7])

	// Content light level information (clli)
	copts.clli_max_cll = C.int(opts.CLLI[0])
	copts.clli_max_pall = C.int(opts.CLLI[1])

	// Animation settings
	copts.timescale = C.int(opts.Timescale)
	copts.keyframe_interval = C.int(opts.KeyframeInterval)

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
	if opts.IgnoreICC {
		copts.ignore_icc = 1
	} else {
		copts.ignore_icc = 0
	}

	// Security limits
	copts.image_size_limit = C.uint32_t(opts.ImageSizeLimit)
	copts.image_dimension_limit = C.uint32_t(opts.ImageDimensionLimit)

	// Validation flags
	copts.strict_flags = C.int(opts.StrictFlags)

	// Chroma upsampling
	copts.chroma_upsampling = C.int(opts.ChromaUpsampling)

	return copts
}

// AVIFEncodeBytes encodes image file data (JPEG, PNG, etc.) to AVIF format
// This is equivalent to the avifenc command-line tool.
// The input data should be a complete image file (JPEG, PNG, etc.) not raw pixel data.
func AVIFEncodeBytes(
	imageFileData []byte,
	options AVIFEncodeOptions,
) ([]byte, error) {
	clearError()

	if len(imageFileData) == 0 {
		return nil, fmt.Errorf("avif encode: empty input data")
	}

	// Convert options
	copts := options.toCEncodeOptions()

	// Copy metadata to C memory to avoid Go pointer issues
	// We need to free these after the C call
	var exifPtr, xmpPtr, iccPtr unsafe.Pointer
	if len(options.ExifData) > 0 {
		exifPtr = C.CBytes(options.ExifData)
		defer C.free(exifPtr)
		copts.exif_data = (*C.uint8_t)(exifPtr)
		copts.exif_size = C.size_t(len(options.ExifData))
	}
	if len(options.XMPData) > 0 {
		xmpPtr = C.CBytes(options.XMPData)
		defer C.free(xmpPtr)
		copts.xmp_data = (*C.uint8_t)(xmpPtr)
		copts.xmp_size = C.size_t(len(options.XMPData))
	}
	if len(options.ICCData) > 0 {
		iccPtr = C.CBytes(options.ICCData)
		defer C.free(iccPtr)
		copts.icc_data = (*C.uint8_t)(iccPtr)
		copts.icc_size = C.size_t(len(options.ICCData))
	}

	// Encode
	var output C.NextImageBuffer
	status := C.nextimage_avif_encode_alloc(
		(*C.uint8_t)(unsafe.Pointer(&imageFileData[0])),
		C.size_t(len(imageFileData)),
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

// AVIFEncodeFile encodes an image file to AVIF format
// This reads the image file (JPEG, PNG, etc.) and converts it to AVIF.
func AVIFEncodeFile(inputPath string, options AVIFEncodeOptions) ([]byte, error) {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("avif encode file: %w", err)
	}

	return AVIFEncodeBytes(data, options)
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
