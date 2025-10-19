// Package cwebp provides a Go interface to the cwebp command (JPEG/PNG to WebP conversion)
// following the SPEC.md specification.
package cwebp

/*
#cgo CFLAGS: -I../../c/include
#cgo LDFLAGS: -L../../c/build -L../../c/build/libwebp -L../../c/build/libavif -L/opt/homebrew/lib -lnextimage -limageenc -limagedec -limageioutil -lwebp -lwebpdemux -lwebpmux -lsharpyuv -lgif -ljpeg -lpng -lz -lm
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

// Metadata flags (can be combined with bitwise OR)
const (
	MetadataNone = 0 // No metadata
	MetadataEXIF = 1 // Keep EXIF metadata (1 << 0)
	MetadataICC  = 2 // Keep ICC profile (1 << 1)
	MetadataXMP  = 4 // Keep XMP metadata (1 << 2)
	MetadataAll  = 7 // Keep all metadata (EXIF | ICC | XMP)
)

// Options represents WebP encoding options.
// This corresponds to CWebPOptions in C.
type Options struct {
	Quality          float32
	Lossless         bool
	Method           int
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

	// Metadata settings
	KeepMetadata int // Bitwise OR of MetadataEXIF, MetadataICC, MetadataXMP (e.g., MetadataEXIF | MetadataXMP)
}

// Command represents a cwebp command instance that can be reused for multiple conversions.
type Command struct {
	cmd *C.CWebPCommand
}

// NewDefaultOptions creates default WebP encoding options.
func NewDefaultOptions() Options {
	cOpts := C.cwebp_create_default_options()
	if cOpts == nil {
		return Options{Quality: 75, Method: 4} // fallback defaults
	}
	defer C.cwebp_free_options(cOpts)

	return Options{
		Quality:          float32(cOpts.quality),
		Lossless:         cOpts.lossless != 0,
		Method:           int(cOpts.method),
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
		KeepMetadata:     int(cOpts.keep_metadata),
	}
}

// optionsToCOptions converts Go Options to C CWebPOptions
func optionsToCOptions(opts Options) *C.CWebPOptions {
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

	// Metadata settings
	cOpts.keep_metadata = C.int(opts.KeepMetadata)

	return cOpts
}

// NewCommand creates a new cwebp command with the given options.
// If opts is nil, default options are used.
func NewCommand(opts *Options) (*Command, error) {
	var cOpts *C.CWebPOptions
	if opts != nil {
		cOpts = optionsToCOptions(*opts)
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

	cmd := &Command{cmd: cCmd}
	runtime.SetFinalizer(cmd, (*Command).Close)
	return cmd, nil
}

// Run converts image data (JPEG/PNG) to WebP format.
// This is the core method that operates on byte slices.
func (c *Command) Run(imageData []byte) ([]byte, error) {
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
func (c *Command) RunFile(inputPath, outputPath string) error {
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
func (c *Command) RunIO(input io.Reader, output io.Writer) error {
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
func (c *Command) Close() error {
	if c.cmd != nil {
		C.cwebp_free_command(c.cmd)
		c.cmd = nil
	}
	return nil
}
