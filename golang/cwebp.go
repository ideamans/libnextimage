// Package libnextimage provides Go bindings for libnextimage - WebP and AVIF encoding/decoding library.
//
// This package includes encoders and decoders for:
//   - WebP (NewCWebPCommand, NewDWebPCommand)
//   - AVIF (NewAVIFEncCommand, NewAVIFDecCommand)
//   - GIF<->WebP conversion (NewGif2WebPCommand, NewWebP2GifCommand)
package libnextimage

/*
#cgo CFLAGS: -I${SRCDIR}/shared/include

// libnextimage.a is a fully self-contained static library that includes:
// - webp, avif, aom (image codecs)
// - jpeg, png, gif (system image libraries)
//
// Only minimal system libraries are needed:
// - zlib: compression (required by PNG)
// - C++ standard library: libavif and libaom are written in C++
// - pthread: multi-threading support
// - math library: mathematical functions

// Platform-specific embedded static libraries
#cgo darwin,arm64 LDFLAGS: ${SRCDIR}/shared/lib/darwin-arm64/libnextimage.a
#cgo darwin,amd64 LDFLAGS: ${SRCDIR}/shared/lib/darwin-amd64/libnextimage.a
#cgo linux,amd64 LDFLAGS: ${SRCDIR}/shared/lib/linux-amd64/libnextimage.a
#cgo linux,arm64 LDFLAGS: ${SRCDIR}/shared/lib/linux-arm64/libnextimage.a
#cgo windows,amd64 LDFLAGS: ${SRCDIR}/shared/lib/windows-amd64/libnextimage.a

// macOS
#cgo darwin LDFLAGS: -lz -lc++ -lpthread -lm

// Linux
#cgo linux LDFLAGS: -lz -lstdc++ -lpthread -lm

// Windows (MSYS2/MinGW)
#cgo windows LDFLAGS: -lz -lstdc++ -lpthread -lm

#include <stdlib.h>
#include <string.h>
#include "nextimage.h"
#include "nextimage/cwebp.h"
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"
)

// CWebPOptions represents WebP encoding options.
// This corresponds to CWebPOptions in C.
type CWebPOptions struct {
	Quality          float32
	Lossless         bool
	Method           int
	Preset           int // Preset type: -1=none (default), or PresetDefault/PresetPicture/PresetPhoto/PresetDrawing/PresetIcon/PresetText
	ImageHint        int // Image type hint: HintDefault/HintPicture/HintPhoto/HintGraph
	LosslessPreset   int // Lossless preset: -1=don't use (default), 0-9=use preset (0=fast, 9=best)
	TargetSize       int
	TargetPSNR       float32
	Segments         int
	SNSStrength      int
	FilterStrength   int
	FilterSharpness  int
	FilterType       int
	Autofilter       bool
	AlphaCompression int
	AlphaFiltering   int
	AlphaQuality     int
	Pass             int
	ShowCompressed   bool
	Preprocessing    int
	Partitions       int
	PartitionLimit   int
	EmulateJPEGSize  bool
	ThreadLevel      int
	LowMemory        bool
	NearLossless     int
	Exact            bool
	UseDeltaPalette  bool
	UseSharpYUV      bool
	QMin             int // Minimum permissible quality (0-100), default 0
	QMax             int // Maximum permissible quality (0-100), default 100

	// Metadata settings
	KeepMetadata int // Bitwise OR of MetadataEXIF, MetadataICC, MetadataXMP (e.g., MetadataEXIF | MetadataXMP)
}

// Command represents a cwebp command instance that can be reused for multiple conversions.
type CWebPCommand struct {
	cmd *C.CWebPCommand
}

// NewDefaultOptions creates default WebP encoding options.
func NewDefaultCWebPOptions() CWebPOptions {
	cOpts := C.cwebp_create_default_options()
	if cOpts == nil {
		return CWebPOptions{Quality: 75, Method: 4, Preset: -1, LosslessPreset: -1, QMax: 100} // fallback defaults
	}
	defer C.cwebp_free_options(cOpts)

	return CWebPOptions{
		Quality:          float32(cOpts.quality),
		Lossless:         cOpts.lossless != 0,
		Method:           int(cOpts.method),
		Preset:           int(cOpts.preset),
		ImageHint:        int(cOpts.image_hint),
		LosslessPreset:   int(cOpts.lossless_preset),
		TargetSize:       int(cOpts.target_size),
		TargetPSNR:       float32(cOpts.target_psnr),
		Segments:         int(cOpts.segments),
		SNSStrength:      int(cOpts.sns_strength),
		FilterStrength:   int(cOpts.filter_strength),
		FilterSharpness:  int(cOpts.filter_sharpness),
		FilterType:       int(cOpts.filter_type),
		Autofilter:       cOpts.autofilter != 0,
		AlphaCompression: int(cOpts.alpha_compression),
		AlphaFiltering:   int(cOpts.alpha_filtering),
		AlphaQuality:     int(cOpts.alpha_quality),
		Pass:             int(cOpts.pass),
		ShowCompressed:   cOpts.show_compressed != 0,
		Preprocessing:    int(cOpts.preprocessing),
		Partitions:       int(cOpts.partitions),
		PartitionLimit:   int(cOpts.partition_limit),
		EmulateJPEGSize:  cOpts.emulate_jpeg_size != 0,
		ThreadLevel:      int(cOpts.thread_level),
		LowMemory:        cOpts.low_memory != 0,
		NearLossless:     int(cOpts.near_lossless),
		Exact:            cOpts.exact != 0,
		UseDeltaPalette:  cOpts.use_delta_palette != 0,
		UseSharpYUV:      cOpts.use_sharp_yuv != 0,
		QMin:             int(cOpts.qmin),
		QMax:             int(cOpts.qmax),
		KeepMetadata:     int(cOpts.keep_metadata),
	}
}

// cwebpOptionsToCOptions converts Go Options to C CWebPOptions
func cwebpOptionsToCOptions(opts CWebPOptions) *C.CWebPOptions {
	cOpts := C.cwebp_create_default_options()
	if cOpts == nil {
		return nil
	}

	cOpts.quality = C.float(opts.Quality)
	if opts.Lossless {
		cOpts.lossless = 1
	} else {
		cOpts.lossless = 0
	}
	cOpts.method = C.int(opts.Method)
	cOpts.preset = C.CWebPPreset(opts.Preset)
	cOpts.image_hint = C.CWebPImageHint(opts.ImageHint)
	cOpts.lossless_preset = C.int(opts.LosslessPreset)
	cOpts.target_size = C.int(opts.TargetSize)
	cOpts.target_psnr = C.float(opts.TargetPSNR)
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
	cOpts.alpha_compression = C.int(opts.AlphaCompression)
	cOpts.alpha_filtering = C.int(opts.AlphaFiltering)
	cOpts.alpha_quality = C.int(opts.AlphaQuality)
	cOpts.pass = C.int(opts.Pass)
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
	cOpts.thread_level = C.int(opts.ThreadLevel)
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

	// Metadata settings
	cOpts.keep_metadata = C.int(opts.KeepMetadata)

	return cOpts
}

// NewCommand creates a new cwebp command with the given options.
// If opts is nil, default options are used.
func NewCWebPCommand(opts *CWebPOptions) (*CWebPCommand, error) {
	var cOpts *C.CWebPOptions
	if opts != nil {
		cOpts = cwebpOptionsToCOptions(*opts)
	} else {
		cOpts = nil
	}

	cCmd := C.cwebp_new_command(cOpts)

	if cOpts != nil {
		C.cwebp_free_options(cOpts)
	}

	if cCmd == nil {
		errMsg := C.nextimage_last_error_message()
		if errMsg != nil {
			return nil, fmt.Errorf("failed to create cwebp command: %s", C.GoString(errMsg))
		}
		return nil, fmt.Errorf("failed to create cwebp command")
	}

	cmd := &CWebPCommand{cmd: cCmd}
	runtime.SetFinalizer(cmd, func(c *CWebPCommand) {
		_ = c.Close()
	})
	return cmd, nil
}

// Run converts image data (JPEG/PNG) to WebP format.
// This is the core method that operates on byte slices.
func (c *CWebPCommand) Run(imageData []byte) ([]byte, error) {
	if c.cmd == nil {
		return nil, fmt.Errorf("command is closed")
	}

	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty input data")
	}

	var output C.NextImageBuffer
	C.memset(unsafe.Pointer(&output), 0, C.sizeof_NextImageBuffer)

	status := C.cwebp_run_command(
		c.cmd,
		(*C.uint8_t)(unsafe.Pointer(&imageData[0])),
		C.size_t(len(imageData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		errMsg := C.nextimage_last_error_message()
		if errMsg != nil {
			return nil, fmt.Errorf("cwebp encoding failed: %s", C.GoString(errMsg))
		}
		return nil, fmt.Errorf("cwebp encoding failed with status %d", int(status))
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))

	// Free C buffer
	C.nextimage_free_buffer(&output)

	return result, nil
}

// RunFile converts an image file to WebP format and saves it to outputPath.
// This is a convenience method for file-based operations.
func (c *CWebPCommand) RunFile(inputPath, outputPath string) error {
	// Read input file
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Convert
	webpData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	// Write output file
	err = os.WriteFile(outputPath, webpData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// RunIO converts image data from a reader to WebP format and writes to a writer.
// This is a convenience method for stream-based operations.
func (c *CWebPCommand) RunIO(input io.Reader, output io.Writer) error {
	// Read all input
	inputData, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Convert
	webpData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	// Write output
	_, err = output.Write(webpData)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Close releases the command resources.
// After calling Close, the command cannot be used anymore.
func (c *CWebPCommand) Close() error {
	if c.cmd != nil {
		C.cwebp_free_command(c.cmd)
		c.cmd = nil
	}
	return nil
}
