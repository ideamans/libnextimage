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

// Platform-specific embedded static libraries (shared across all golang modules)
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
#include "nextimage/gif2webp.h"
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"
)

// Options represents GIF to WebP encoding options.
// This corresponds to Gif2WebPOptions (which is typedef of CWebPOptions) in C.
type Gif2WebPOptions struct {
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
}

// Command represents a gif2webp command instance that can be reused for multiple conversions.
type Gif2WebPCommand struct {
	cmd *C.Gif2WebPCommand
}

// NewDefaultOptions creates default GIF to WebP encoding options.
func NewDefaultGif2WebPOptions() Gif2WebPOptions {
	cOpts := C.gif2webp_create_default_options()
	if cOpts == nil {
		return Gif2WebPOptions{Quality: 75, Method: 4} // fallback defaults
	}
	defer C.gif2webp_free_options(cOpts)

	return Gif2WebPOptions{
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
	}
}

// optionsToCOptions converts Go Options to C Gif2WebPOptions
func gif2webpOptionsToCOptions(opts Gif2WebPOptions) *C.Gif2WebPOptions {
	cOpts := C.gif2webp_create_default_options()
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

	return cOpts
}

// NewCommand creates a new gif2webp command with the given options.
// If opts is nil, default options are used.
// The returned Command must be closed with Close() when done.
func NewGif2WebPCommand(opts *Gif2WebPOptions) (*Gif2WebPCommand, error) {
	var cOpts *C.Gif2WebPOptions
	if opts != nil {
		cOpts = gif2webpOptionsToCOptions(*opts)
		if cOpts == nil {
			return nil, fmt.Errorf("failed to create options")
		}
	}

	cCmd := C.gif2webp_new_command(cOpts)
	if cOpts != nil {
		C.gif2webp_free_options(cOpts)
	}

	if cCmd == nil {
		errMsg := C.nextimage_last_error_message()
		return nil, fmt.Errorf("failed to create gif2webp command: %s", C.GoString(errMsg))
	}

	cmd := &Gif2WebPCommand{cmd: cCmd}
	runtime.SetFinalizer(cmd, func(c *Gif2WebPCommand) {
		_ = c.Close()
	})
	return cmd, nil
}

// Run converts GIF data to WebP format.
// This is the core method that performs the conversion.
func (c *Gif2WebPCommand) Run(gifData []byte) ([]byte, error) {
	if c.cmd == nil {
		return nil, fmt.Errorf("command is closed")
	}
	if len(gifData) == 0 {
		return nil, fmt.Errorf("input data is empty")
	}

	var output C.NextImageBuffer
	C.memset(unsafe.Pointer(&output), 0, C.sizeof_NextImageBuffer)

	status := C.gif2webp_run_command(
		c.cmd,
		(*C.uint8_t)(unsafe.Pointer(&gifData[0])),
		C.size_t(len(gifData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		errMsg := C.nextimage_last_error_message()
		return nil, fmt.Errorf("gif2webp encoding failed (status %d): %s", status, C.GoString(errMsg))
	}

	if output.data == nil || output.size == 0 {
		return nil, fmt.Errorf("encoding produced empty output")
	}

	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))
	C.nextimage_free_buffer(&output)
	return result, nil
}

// RunFile reads a GIF file, converts it to WebP, and writes the result to outputPath.
// This is sugar syntax over Run().
func (c *Gif2WebPCommand) RunFile(inputPath, outputPath string) error {
	if c.cmd == nil {
		return fmt.Errorf("command is closed")
	}

	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	webpData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, webpData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// RunIO reads GIF data from input, converts it to WebP, and writes the result to output.
// This is sugar syntax over Run().
func (c *Gif2WebPCommand) RunIO(input io.Reader, output io.Writer) error {
	if c.cmd == nil {
		return fmt.Errorf("command is closed")
	}

	inputData, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	webpData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	if _, err := output.Write(webpData); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Close releases the resources associated with the command.
// After calling Close, the command cannot be used anymore.
func (c *Gif2WebPCommand) Close() error {
	if c.cmd != nil {
		C.gif2webp_free_command(c.cmd)
		c.cmd = nil
	}
	return nil
}
