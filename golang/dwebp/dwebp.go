// Package dwebp provides a Go interface to the dwebp command (WebP to PNG conversion)
// following the SPEC.md specification.
package dwebp

/*
#cgo CFLAGS: -I${SRCDIR}/../../include
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/../../lib/darwin-arm64 -lnextimage
#cgo darwin,arm64 LDFLAGS: /opt/homebrew/lib/libjpeg.a /opt/homebrew/lib/libpng.a /opt/homebrew/lib/libgif.a -lz
#cgo darwin,arm64 LDFLAGS: -lc++ -framework CoreFoundation
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/../../lib/darwin-amd64 -lnextimage
#cgo darwin,amd64 LDFLAGS: /usr/local/lib/libjpeg.a /usr/local/lib/libpng.a /usr/local/lib/libgif.a -lz
#cgo darwin,amd64 LDFLAGS: -lc++ -framework CoreFoundation
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../../lib/linux-amd64 -lnextimage
#cgo linux,amd64 LDFLAGS: -ljpeg -lpng -lgif -lz -lpthread -lm -ldl
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../../lib/linux-arm64 -lnextimage
#cgo linux,arm64 LDFLAGS: -ljpeg -lpng -lgif -lz -lpthread -lm -ldl
#include <stdlib.h>
#include <string.h>
#include "nextimage.h"
#include "nextimage/dwebp.h"
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"
)

// OutputFormat represents the output format for decoded images
type OutputFormat int

const (
	OutputPNG  OutputFormat = 0 // PNG output (default)
	OutputJPEG OutputFormat = 1 // JPEG output
)

// Options represents WebP decoding options.
// This corresponds to DWebPOptions in C.
type Options struct {
	OutputFormat      OutputFormat // PNG or JPEG output (default: PNG)
	JPEGQuality       int          // JPEG quality 0-100 (default: 90, only for JPEG output)
	Format            string       // "RGBA", "RGB", "BGRA"
	BypassFiltering   bool
	NoFancyUpsampling bool
	UseThreads        bool

	// Image manipulation
	CropX      int  // crop rectangle x
	CropY      int  // crop rectangle y
	CropWidth  int  // crop rectangle width
	CropHeight int  // crop rectangle height
	UseCrop    bool // enable cropping

	ResizeWidth  int  // resize width
	ResizeHeight int  // resize height
	UseResize    bool // enable resizing

	Flip bool // flip vertically
}

// Command represents a dwebp command instance that can be reused for multiple conversions.
type Command struct {
	cmd *C.DWebPCommand
}

// NewDefaultOptions creates default WebP decoding options.
func NewDefaultOptions() Options {
	cOpts := C.dwebp_create_default_options()
	if cOpts == nil {
		return Options{
			OutputFormat: OutputPNG,
			JPEGQuality:  90,
			Format:       "RGBA",
		} // fallback defaults
	}
	defer C.dwebp_free_options(cOpts)

	format := "RGBA"
	switch cOpts.format {
	case C.NEXTIMAGE_FORMAT_RGB:
		format = "RGB"
	case C.NEXTIMAGE_FORMAT_BGRA:
		format = "BGRA"
	}

	return Options{
		OutputFormat:      OutputFormat(cOpts.output_format),
		JPEGQuality:       int(cOpts.jpeg_quality),
		Format:            format,
		BypassFiltering:   cOpts.bypass_filtering != 0,
		NoFancyUpsampling: cOpts.no_fancy_upsampling != 0,
		UseThreads:        cOpts.use_threads != 0,
		CropX:             int(cOpts.crop_x),
		CropY:             int(cOpts.crop_y),
		CropWidth:         int(cOpts.crop_width),
		CropHeight:        int(cOpts.crop_height),
		UseCrop:           cOpts.use_crop != 0,
		ResizeWidth:       int(cOpts.resize_width),
		ResizeHeight:      int(cOpts.resize_height),
		UseResize:         cOpts.use_resize != 0,
		Flip:              cOpts.flip != 0,
	}
}

// optionsToCOptions converts Go Options to C DWebPOptions
func optionsToCOptions(opts Options) *C.DWebPOptions {
	cOpts := C.dwebp_create_default_options()
	if cOpts == nil {
		return nil
	}

	// Set output format options
	cOpts.output_format = C.DWebPOutputFormat(opts.OutputFormat)
	cOpts.jpeg_quality = C.int(opts.JPEGQuality)

	switch opts.Format {
	case "RGB":
		cOpts.format = C.NEXTIMAGE_FORMAT_RGB
	case "BGRA":
		cOpts.format = C.NEXTIMAGE_FORMAT_BGRA
	default:
		cOpts.format = C.NEXTIMAGE_FORMAT_RGBA
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

	if opts.UseThreads {
		cOpts.use_threads = 1
	} else {
		cOpts.use_threads = 0
	}

	// Image manipulation options
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

	return cOpts
}

// NewCommand creates a new dwebp command with the given options.
// If opts is nil, default options are used.
func NewCommand(opts *Options) (*Command, error) {
	var cOpts *C.DWebPOptions
	if opts != nil {
		cOpts = optionsToCOptions(*opts)
	} else {
		cOpts = nil
	}

	cCmd := C.dwebp_new_command(cOpts)

	if cOpts != nil {
		C.dwebp_free_options(cOpts)
	}

	if cCmd == nil {
		errMsg := C.nextimage_last_error_message()
		if errMsg != nil {
			return nil, fmt.Errorf("failed to create dwebp command: %s", C.GoString(errMsg))
		}
		return nil, fmt.Errorf("failed to create dwebp command")
	}

	cmd := &Command{cmd: cCmd}
	runtime.SetFinalizer(cmd, (*Command).Close)
	return cmd, nil
}

// Run converts WebP data to PNG format.
// This is the core method that operates on byte slices.
func (c *Command) Run(webpData []byte) ([]byte, error) {
	if c.cmd == nil {
		return nil, fmt.Errorf("command is closed")
	}

	if len(webpData) == 0 {
		return nil, fmt.Errorf("empty input data")
	}

	var output C.NextImageBuffer
	C.memset(unsafe.Pointer(&output), 0, C.sizeof_NextImageBuffer)

	status := C.dwebp_run_command(
		c.cmd,
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		errMsg := C.nextimage_last_error_message()
		if errMsg != nil {
			return nil, fmt.Errorf("dwebp decoding failed: %s", C.GoString(errMsg))
		}
		return nil, fmt.Errorf("dwebp decoding failed with status %d", int(status))
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))

	// Free C buffer
	C.nextimage_free_buffer(&output)

	return result, nil
}

// RunFile converts a WebP file to PNG format and saves it to outputPath.
// This is a convenience method for file-based operations.
func (c *Command) RunFile(inputPath, outputPath string) error {
	// Read input file
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Convert
	pngData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	// Write output file
	err = os.WriteFile(outputPath, pngData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// RunIO converts WebP data from a reader to PNG format and writes to a writer.
// This is a convenience method for stream-based operations.
func (c *Command) RunIO(input io.Reader, output io.Writer) error {
	// Read all input
	inputData, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Convert
	pngData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	// Write output
	_, err = output.Write(pngData)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Close releases the command resources.
// After calling Close, the command cannot be used anymore.
func (c *Command) Close() error {
	if c.cmd != nil {
		C.dwebp_free_command(c.cmd)
		c.cmd = nil
	}
	return nil
}
