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

// AVIFYUVFormat represents YUV format for AVIF encoding
type AVIFYUVFormat int

const (
	YUVFormat444  AVIFYUVFormat = 0  // 4:4:4 (no chroma subsampling)
	YUVFormat422  AVIFYUVFormat = 1  // 4:2:2 (horizontal subsampling)
	YUVFormat420  AVIFYUVFormat = 2  // 4:2:0 (both horizontal and vertical subsampling)
	YUVFormat400  AVIFYUVFormat = 3  // 4:0:0 (grayscale, no chroma)
	YUVFormatAuto AVIFYUVFormat = -1 // Auto-detect from input (default)
)

// AVIFYUVRange represents YUV range for AVIF encoding
type AVIFYUVRange int

const (
	YUVRangeLimited AVIFYUVRange = 0 // Limited range (16-235 for 8-bit)
	YUVRangeFull    AVIFYUVRange = 1 // Full range (0-255 for 8-bit, default)
)

// AVIFMirrorAxis represents mirror axis for AVIF encoding
type AVIFMirrorAxis int

const (
	MirrorAxisNone       AVIFMirrorAxis = -1 // No mirroring (disabled, default)
	MirrorAxisVertical   AVIFMirrorAxis = 0  // Top-to-bottom mirroring
	MirrorAxisHorizontal AVIFMirrorAxis = 1  // Left-to-right mirroring
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
	BitDepth  int           // 8, 10, or 12 (default: 8)
	YUVFormat AVIFYUVFormat // YUV format: 444/422/420/400 (default: 444)
	YUVRange  AVIFYUVRange  // YUV range: limited/full (default: full)

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
	Lossless   bool // Lossless mode (sets Quality=100), default false

	// Threading and tiling
	Jobs       int  // -1=all cores, 0=auto, >0=specific thread count (default: -1)
	AutoTiling bool // Enable automatic tiling (default: true)

	// Metadata settings
	ExifData []byte // EXIF metadata bytes (nil=no EXIF)
	XMPData  []byte // XMP metadata bytes (nil=no XMP)
	ICCData  []byte // ICC profile bytes (nil=no ICC)

	// Transformation settings
	IRotAngle int            // Image rotation: 0-3 (90 * angle degrees anti-clockwise), -1=disabled
	IMirAxis  AVIFMirrorAxis // Image mirror: vertical/horizontal/none (default: none)

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
	// Threading
	Jobs int // -1=all cores (default), 0=auto, >0=specific thread count

	// Output format
	Format PixelFormat // desired pixel format (default: RGBA)

	// Metadata handling
	IgnoreExif bool // Ignore embedded EXIF metadata
	IgnoreXMP  bool // Ignore embedded XMP metadata
	IgnoreICC  bool // Ignore embedded ICC profile

	// Security limits
	ImageSizeLimit      uint32 // Maximum image size in total pixels (default: 268435456)
	ImageDimensionLimit uint32 // Maximum image dimension (width or height), 0=ignore (default: 32768)

	// Validation flags
	StrictFlags int // Strict validation flags: 0=disabled, 1=enabled (default: 1)

	// Chroma upsampling (for YUV to RGB conversion)
	ChromaUpsampling ChromaUpsampling // 0=automatic (default), 1=fastest, 2=best_quality, 3=nearest, 4=bilinear

	// Image manipulation
	CropX      int  // Crop rectangle X coordinate
	CropY      int  // Crop rectangle Y coordinate
	CropWidth  int  // Crop rectangle width
	CropHeight int  // Crop rectangle height
	UseCrop    bool // Enable cropping

	ResizeWidth  int  // Resize target width
	ResizeHeight int  // Resize target height
	UseResize    bool // Enable resizing
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
		YUVFormat:               AVIFYUVFormat(opts.yuv_format),
		YUVRange:                AVIFYUVRange(opts.yuv_range),
		EnableAlpha:             opts.enable_alpha != 0,
		PremultiplyAlpha:        opts.premultiply_alpha != 0,
		TileRowsLog2:            int(opts.tile_rows_log2),
		TileColsLog2:            int(opts.tile_cols_log2),
		ColorPrimaries:          int(opts.color_primaries),
		TransferCharacteristics: int(opts.transfer_characteristics),
		MatrixCoefficients:      int(opts.matrix_coefficients),
		SharpYUV:     opts.sharp_yuv != 0,
		TargetSize:   int(opts.target_size),
		Lossless:     false,
		Jobs:         -1,   // -1 = all cores (default)
		AutoTiling:   true, // automatic tiling enabled by default
		IRotAngle:    int(opts.irot_angle),
		IMirAxis:     AVIFMirrorAxis(opts.imir_axis),
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
		// Threading
		Jobs: -1, // -1 = all cores (default)

		// Output format
		Format: PixelFormat(opts.format),

		// Metadata handling
		IgnoreExif: opts.ignore_exif != 0,
		IgnoreXMP:  opts.ignore_xmp != 0,
		IgnoreICC:  opts.ignore_icc != 0,

		// Security limits
		ImageSizeLimit:      uint32(opts.image_size_limit),
		ImageDimensionLimit: uint32(opts.image_dimension_limit),

		// Validation
		StrictFlags: int(opts.strict_flags),

		// Chroma upsampling
		ChromaUpsampling: ChromaUpsampling(opts.chroma_upsampling),

		// Image manipulation
		CropX:        int(opts.crop_x),
		CropY:        int(opts.crop_y),
		CropWidth:    int(opts.crop_width),
		CropHeight:   int(opts.crop_height),
		UseCrop:      opts.use_crop != 0,
		ResizeWidth:  int(opts.resize_width),
		ResizeHeight: int(opts.resize_height),
		UseResize:    opts.use_resize != 0,
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

	// Threading: Jobs field maps to use_threads (bool in C)
	// -1 or >0 means use threads, 0 means don't use threads
	if opts.Jobs != 0 {
		copts.use_threads = 1
	} else {
		copts.use_threads = 0
	}

	copts.format = C.NextImagePixelFormat(opts.Format)

	// Metadata handling
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

	// Image manipulation - Cropping
	copts.crop_x = C.int(opts.CropX)
	copts.crop_y = C.int(opts.CropY)
	copts.crop_width = C.int(opts.CropWidth)
	copts.crop_height = C.int(opts.CropHeight)
	if opts.UseCrop {
		copts.use_crop = 1
	} else {
		copts.use_crop = 0
	}

	// Image manipulation - Resizing
	copts.resize_width = C.int(opts.ResizeWidth)
	copts.resize_height = C.int(opts.ResizeHeight)
	if opts.UseResize {
		copts.use_resize = 1
	} else {
		copts.use_resize = 0
	}

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

// ========================================
// AVIF Encoder/Decoder (Instance-based API)
// ========================================

// AVIFEncoder represents an AVIF encoder instance that can be reused for multiple images
type AVIFEncoder struct {
	encoderPtr *C.NextImageAVIFEncoder
}

// NewAVIFEncoder creates a new AVIF encoder with the given options
// Options can be customized using the provided callback function
//
// Example:
//
//	encoder, err := NewAVIFEncoder(func(opts *AVIFEncodeOptions) {
//	    opts.Quality = 60
//	    opts.Speed = 6
//	    opts.BitDepth = 10
//	})
func NewAVIFEncoder(optsFn func(*AVIFEncodeOptions)) (*AVIFEncoder, error) {
	clearError()

	// Get default options
	opts := DefaultAVIFEncodeOptions()

	// Apply user customizations
	if optsFn != nil {
		optsFn(&opts)
	}

	// Convert to C struct
	cOpts := opts.toCEncodeOptions()

	// Create encoder
	encoderPtr := C.nextimage_avif_encoder_create(&cOpts)
	if encoderPtr == nil {
		return nil, fmt.Errorf("avif encoder: failed to create encoder: %s", getLastError())
	}

	return &AVIFEncoder{encoderPtr: encoderPtr}, nil
}

// Encode encodes image file data (JPEG, PNG, etc.) to AVIF format
// The encoder instance can be reused for multiple images, reducing initialization overhead
func (e *AVIFEncoder) Encode(imageFileData []byte) ([]byte, error) {
	if e.encoderPtr == nil {
		return nil, fmt.Errorf("avif encoder: encoder is closed")
	}

	if len(imageFileData) == 0 {
		return nil, fmt.Errorf("avif encoder: empty input data")
	}

	var encoded C.NextImageBuffer
	status := C.nextimage_avif_encoder_encode(
		e.encoderPtr,
		(*C.uint8_t)(unsafe.Pointer(&imageFileData[0])),
		C.size_t(len(imageFileData)),
		&encoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "avif encoder encode")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(encoded.data), C.int(encoded.size))

	// Free C buffer
	freeEncodeBuffer(&encoded)

	return result, nil
}

// Close releases resources associated with the encoder
// Must be called when done using the encoder
func (e *AVIFEncoder) Close() {
	if e.encoderPtr != nil {
		C.nextimage_avif_encoder_destroy(e.encoderPtr)
		e.encoderPtr = nil
	}
}

// AVIFDecoder represents an AVIF decoder instance that can be reused for multiple images
type AVIFDecoder struct {
	decoderPtr *C.NextImageAVIFDecoder
}

// NewAVIFDecoder creates a new AVIF decoder with the given options
// Options can be customized using the provided callback function
//
// Example:
//
//	decoder, err := NewAVIFDecoder(func(opts *AVIFDecodeOptions) {
//	    opts.Format = PixelFormatRGBA
//	    opts.Jobs = -1  // use all cores
//	})
func NewAVIFDecoder(optsFn func(*AVIFDecodeOptions)) (*AVIFDecoder, error) {
	clearError()

	// Get default options
	opts := DefaultAVIFDecodeOptions()

	// Apply user customizations
	if optsFn != nil {
		optsFn(&opts)
	}

	// Convert to C struct
	cOpts := opts.toCDecodeOptions()

	// Create decoder
	decoderPtr := C.nextimage_avif_decoder_create(&cOpts)
	if decoderPtr == nil {
		return nil, fmt.Errorf("avif decoder: failed to create decoder: %s", getLastError())
	}

	return &AVIFDecoder{decoderPtr: decoderPtr}, nil
}

// Decode decodes AVIF data to pixel data
// The decoder instance can be reused for multiple images, reducing initialization overhead
func (d *AVIFDecoder) Decode(avifData []byte) (*DecodedImage, error) {
	if d.decoderPtr == nil {
		return nil, fmt.Errorf("avif decoder: decoder is closed")
	}

	if len(avifData) == 0 {
		return nil, fmt.Errorf("avif decoder: empty input data")
	}

	var decoded C.NextImageDecodeBuffer
	status := C.nextimage_avif_decoder_decode(
		d.decoderPtr,
		(*C.uint8_t)(unsafe.Pointer(&avifData[0])),
		C.size_t(len(avifData)),
		&decoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "avif decoder decode")
	}

	// Convert to Go structure
	img := convertDecodeBuffer(&decoded)

	// Free C buffer
	freeDecodeBuffer(&decoded)

	return img, nil
}

// Close releases resources associated with the decoder
// Must be called when done using the decoder
func (d *AVIFDecoder) Close() {
	if d.decoderPtr != nil {
		C.nextimage_avif_decoder_destroy(d.decoderPtr)
		d.decoderPtr = nil
	}
}
