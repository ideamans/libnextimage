package avifenc

/*
#cgo CFLAGS: -I${SRCDIR}/../../include

// libnextimage.a is a fully self-contained static library that includes:
// - webp, avif, aom (image codecs)
// - jpeg, png, gif (system image libraries)
//
// Only minimal system libraries are needed:
// - zlib: compression (required by PNG)
// - C++ standard library: libavif and libaom are written in C++
// - pthread: multi-threading support
// - math library: mathematical functions

// Use full path to static library to avoid linking against shared library
#cgo LDFLAGS: ${SRCDIR}/../../lib/static/libnextimage.a

// macOS
#cgo darwin LDFLAGS: -lz -lc++ -lpthread -lm

// Linux
#cgo linux LDFLAGS: -lz -lstdc++ -lpthread -lm

// Windows (MSYS2/MinGW)
#cgo windows LDFLAGS: -lz -lstdc++ -lpthread -lm

#include <stdlib.h>
#include <string.h>
#include "nextimage.h"
#include "nextimage/avifenc.h"
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"
)

// Options represents AVIF encoding options
type Options struct {
	// Quality settings
	Quality      int // 0-100, default 60 (for color/YUV)
	QualityAlpha int // 0-100, default 100 (for alpha channel, -1=use quality)
	Speed        int // 0-10, default 6 (0=slowest/best, 10=fastest/worst)

	// Deprecated quantizer settings (for backward compatibility)
	MinQuantizer      int // 0-63, default -1 (use quality instead)
	MaxQuantizer      int // 0-63, default -1 (use quality instead)
	MinQuantizerAlpha int // 0-63, default -1 (use quality_alpha instead)
	MaxQuantizerAlpha int // 0-63, default -1 (use quality_alpha instead)

	// Format settings
	BitDepth  int // 8, 10, or 12 (default: 8)
	YUVFormat int // 0=444, 1=422, 2=420, 3=400 (default: 444)
	YUVRange  int // 0=limited, 1=full (default: 1=full for PNG/JPEG)

	// Alpha settings
	EnableAlpha       bool // default true
	PremultiplyAlpha  bool // default false (premultiply color by alpha)

	// Tiling settings
	TileRowsLog2 int // 0-6, default 0
	TileColsLog2 int // 0-6, default 0

	// CICP (nclx) color settings
	ColorPrimaries          int // CICP color primaries, -1=auto (default: 1=BT709)
	TransferCharacteristics int // CICP transfer, -1=auto (default: 13=sRGB)
	MatrixCoefficients      int // CICP matrix, -1=auto (default: 6=BT601)

	// Advanced settings
	SharpYUV   bool // use sharp RGB->YUV conversion (default: false)
	TargetSize int  // target file size in bytes, 0=disabled (default: 0)

	// Metadata settings
	EXIFData []byte // EXIF metadata bytes (nil=no EXIF)
	XMPData  []byte // XMP metadata bytes (nil=no XMP)
	ICCData  []byte // ICC profile bytes (nil=no ICC)

	// Transformation settings
	IrotAngle int // Image rotation: 0-3 (90 * angle degrees anti-clockwise), -1=disabled
	ImirAxis  int // Image mirror: 0=vertical, 1=horizontal, -1=disabled

	// Pixel aspect ratio (pasp) - [h_spacing, v_spacing]
	PASP [2]int // -1=disabled, otherwise [h_spacing, v_spacing]

	// Crop rectangle - [x, y, width, height]
	Crop [4]int // -1=disabled, otherwise [x, y, width, height]

	// Clean aperture (clap) - [widthN, widthD, heightN, heightD, horizOffN, horizOffD, vertOffN, vertOffD]
	CLAP [8]int // -1=disabled, otherwise [wN,wD, hN,hD, hOffN,hOffD, vOffN,vOffD]

	// Content light level information (clli)
	CLLIMaxCLL  int // Max content light level (0-65535), -1=disabled
	CLLIMaxPALL int // Max picture average light level (0-65535), -1=disabled

	// Animation settings (for future use)
	Timescale        int // timescale/fps for animations (default: 30)
	KeyframeInterval int // max keyframe interval (default: 0=disabled)
}

// Command represents an AVIF encoder command that can be reused for multiple conversions
type Command struct {
	cmd *C.AVIFEncCommand
}

// NewDefaultOptions creates a new Options struct with default values
func NewDefaultOptions() Options {
	cOpts := C.avifenc_create_default_options()
	if cOpts == nil {
		// Return hardcoded defaults if C function fails
		return Options{
			Quality:                 60,
			QualityAlpha:            -1,
			Speed:                   6,
			MinQuantizer:            -1,
			MaxQuantizer:            -1,
			MinQuantizerAlpha:       -1,
			MaxQuantizerAlpha:       -1,
			BitDepth:                8,
			YUVFormat:               0,
			YUVRange:                1,
			EnableAlpha:             true,
			PremultiplyAlpha:        false,
			TileRowsLog2:            0,
			TileColsLog2:            0,
			ColorPrimaries:          1,
			TransferCharacteristics: 13,
			MatrixCoefficients:      6,
			SharpYUV:                false,
			TargetSize:              0,
			IrotAngle:               -1,
			ImirAxis:                -1,
			PASP:                    [2]int{-1, -1},
			Crop:                    [4]int{-1, -1, -1, -1},
			CLAP:                    [8]int{-1, -1, -1, -1, -1, -1, -1, -1},
			CLLIMaxCLL:              -1,
			CLLIMaxPALL:             -1,
			Timescale:               30,
			KeyframeInterval:        0,
		}
	}
	defer C.avifenc_free_options(cOpts)

	return Options{
		Quality:                 int(cOpts.quality),
		QualityAlpha:            int(cOpts.quality_alpha),
		Speed:                   int(cOpts.speed),
		MinQuantizer:            int(cOpts.min_quantizer),
		MaxQuantizer:            int(cOpts.max_quantizer),
		MinQuantizerAlpha:       int(cOpts.min_quantizer_alpha),
		MaxQuantizerAlpha:       int(cOpts.max_quantizer_alpha),
		BitDepth:                int(cOpts.bit_depth),
		YUVFormat:               int(cOpts.yuv_format),
		YUVRange:                int(cOpts.yuv_range),
		EnableAlpha:             cOpts.enable_alpha != 0,
		PremultiplyAlpha:        cOpts.premultiply_alpha != 0,
		TileRowsLog2:            int(cOpts.tile_rows_log2),
		TileColsLog2:            int(cOpts.tile_cols_log2),
		ColorPrimaries:          int(cOpts.color_primaries),
		TransferCharacteristics: int(cOpts.transfer_characteristics),
		MatrixCoefficients:      int(cOpts.matrix_coefficients),
		SharpYUV:                cOpts.sharp_yuv != 0,
		TargetSize:              int(cOpts.target_size),
		IrotAngle:               int(cOpts.irot_angle),
		ImirAxis:                int(cOpts.imir_axis),
		PASP:                    [2]int{int(cOpts.pasp[0]), int(cOpts.pasp[1])},
		Crop:                    [4]int{int(cOpts.crop[0]), int(cOpts.crop[1]), int(cOpts.crop[2]), int(cOpts.crop[3])},
		CLAP:                    [8]int{int(cOpts.clap[0]), int(cOpts.clap[1]), int(cOpts.clap[2]), int(cOpts.clap[3]), int(cOpts.clap[4]), int(cOpts.clap[5]), int(cOpts.clap[6]), int(cOpts.clap[7])},
		CLLIMaxCLL:              int(cOpts.clli_max_cll),
		CLLIMaxPALL:             int(cOpts.clli_max_pall),
		Timescale:               int(cOpts.timescale),
		KeyframeInterval:        int(cOpts.keyframe_interval),
	}
}

// optionsToCOptions converts Go Options to C AVIFEncOptions
func optionsToCOptions(opts Options) *C.AVIFEncOptions {
	cOpts := C.avifenc_create_default_options()
	if cOpts == nil {
		return nil
	}

	cOpts.quality = C.int(opts.Quality)
	cOpts.quality_alpha = C.int(opts.QualityAlpha)
	cOpts.speed = C.int(opts.Speed)
	cOpts.min_quantizer = C.int(opts.MinQuantizer)
	cOpts.max_quantizer = C.int(opts.MaxQuantizer)
	cOpts.min_quantizer_alpha = C.int(opts.MinQuantizerAlpha)
	cOpts.max_quantizer_alpha = C.int(opts.MaxQuantizerAlpha)
	cOpts.bit_depth = C.int(opts.BitDepth)
	cOpts.yuv_format = C.int(opts.YUVFormat)
	cOpts.yuv_range = C.int(opts.YUVRange)
	if opts.EnableAlpha {
		cOpts.enable_alpha = 1
	} else {
		cOpts.enable_alpha = 0
	}
	if opts.PremultiplyAlpha {
		cOpts.premultiply_alpha = 1
	} else {
		cOpts.premultiply_alpha = 0
	}
	cOpts.tile_rows_log2 = C.int(opts.TileRowsLog2)
	cOpts.tile_cols_log2 = C.int(opts.TileColsLog2)
	cOpts.color_primaries = C.int(opts.ColorPrimaries)
	cOpts.transfer_characteristics = C.int(opts.TransferCharacteristics)
	cOpts.matrix_coefficients = C.int(opts.MatrixCoefficients)
	if opts.SharpYUV {
		cOpts.sharp_yuv = 1
	} else {
		cOpts.sharp_yuv = 0
	}
	cOpts.target_size = C.int(opts.TargetSize)

	// Metadata settings
	if len(opts.EXIFData) > 0 {
		cOpts.exif_data = (*C.uint8_t)(unsafe.Pointer(&opts.EXIFData[0]))
		cOpts.exif_size = C.size_t(len(opts.EXIFData))
	} else {
		cOpts.exif_data = nil
		cOpts.exif_size = 0
	}

	if len(opts.XMPData) > 0 {
		cOpts.xmp_data = (*C.uint8_t)(unsafe.Pointer(&opts.XMPData[0]))
		cOpts.xmp_size = C.size_t(len(opts.XMPData))
	} else {
		cOpts.xmp_data = nil
		cOpts.xmp_size = 0
	}

	if len(opts.ICCData) > 0 {
		cOpts.icc_data = (*C.uint8_t)(unsafe.Pointer(&opts.ICCData[0]))
		cOpts.icc_size = C.size_t(len(opts.ICCData))
	} else {
		cOpts.icc_data = nil
		cOpts.icc_size = 0
	}

	cOpts.irot_angle = C.int(opts.IrotAngle)
	cOpts.imir_axis = C.int(opts.ImirAxis)

	// Pixel aspect ratio (pasp)
	cOpts.pasp[0] = C.int(opts.PASP[0])
	cOpts.pasp[1] = C.int(opts.PASP[1])

	// Crop rectangle
	cOpts.crop[0] = C.int(opts.Crop[0])
	cOpts.crop[1] = C.int(opts.Crop[1])
	cOpts.crop[2] = C.int(opts.Crop[2])
	cOpts.crop[3] = C.int(opts.Crop[3])

	// Clean aperture (clap)
	cOpts.clap[0] = C.int(opts.CLAP[0])
	cOpts.clap[1] = C.int(opts.CLAP[1])
	cOpts.clap[2] = C.int(opts.CLAP[2])
	cOpts.clap[3] = C.int(opts.CLAP[3])
	cOpts.clap[4] = C.int(opts.CLAP[4])
	cOpts.clap[5] = C.int(opts.CLAP[5])
	cOpts.clap[6] = C.int(opts.CLAP[6])
	cOpts.clap[7] = C.int(opts.CLAP[7])

	// Content light level information (clli)
	cOpts.clli_max_cll = C.int(opts.CLLIMaxCLL)
	cOpts.clli_max_pall = C.int(opts.CLLIMaxPALL)

	cOpts.timescale = C.int(opts.Timescale)
	cOpts.keyframe_interval = C.int(opts.KeyframeInterval)

	return cOpts
}

// NewCommand creates a new AVIF encoder command with the given options.
// If opts is nil, default options are used.
// The returned Command must be closed with Close() when done.
func NewCommand(opts *Options) (*Command, error) {
	var cOpts *C.AVIFEncOptions
	if opts != nil {
		cOpts = optionsToCOptions(*opts)
		if cOpts == nil {
			return nil, fmt.Errorf("failed to create options")
		}
	}

	cCmd := C.avifenc_new_command(cOpts)
	if cOpts != nil {
		C.avifenc_free_options(cOpts)
	}

	if cCmd == nil {
		errMsg := C.nextimage_last_error_message()
		return nil, fmt.Errorf("failed to create avifenc command: %s", C.GoString(errMsg))
	}

	cmd := &Command{cmd: cCmd}
	runtime.SetFinalizer(cmd, func(c *Command) {
		_ = c.Close()
	})
	return cmd, nil
}

// Run converts image data (JPEG or PNG) to AVIF format.
// This is the core method that performs the conversion.
func (c *Command) Run(imageData []byte) ([]byte, error) {
	if c.cmd == nil {
		return nil, fmt.Errorf("command is closed")
	}
	if len(imageData) == 0 {
		return nil, fmt.Errorf("input data is empty")
	}

	var output C.NextImageBuffer
	C.memset(unsafe.Pointer(&output), 0, C.sizeof_NextImageBuffer)

	status := C.avifenc_run_command(
		c.cmd,
		(*C.uint8_t)(unsafe.Pointer(&imageData[0])),
		C.size_t(len(imageData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		errMsg := C.nextimage_last_error_message()
		return nil, fmt.Errorf("avifenc encoding failed (status %d): %s", status, C.GoString(errMsg))
	}

	if output.data == nil || output.size == 0 {
		return nil, fmt.Errorf("encoding produced empty output")
	}

	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))
	C.nextimage_free_buffer(&output)
	return result, nil
}

// RunFile reads an image file, converts it to AVIF, and writes the result to outputPath.
// This is sugar syntax over Run().
func (c *Command) RunFile(inputPath, outputPath string) error {
	if c.cmd == nil {
		return fmt.Errorf("command is closed")
	}

	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	avifData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, avifData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// RunIO reads image data from input, converts it to AVIF, and writes the result to output.
// This is sugar syntax over Run().
func (c *Command) RunIO(input io.Reader, output io.Writer) error {
	if c.cmd == nil {
		return fmt.Errorf("command is closed")
	}

	inputData, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	avifData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	if _, err := output.Write(avifData); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Close releases the resources associated with the command.
// After calling Close, the command cannot be used anymore.
func (c *Command) Close() error {
	if c.cmd != nil {
		C.avifenc_free_command(c.cmd)
		c.cmd = nil
	}
	return nil
}
