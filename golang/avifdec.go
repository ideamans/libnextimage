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
#include "nextimage/avifdec.h"
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

// Options represents AVIF decoding options
type AVIFDecOptions struct {
	OutputFormat         OutputFormat // PNG or JPEG output (default: PNG)
	JPEGQuality          int          // JPEG quality 0-100 (default: 90, only for JPEG output)
	UseThreads           bool         // enable multi-threading
	Format               string       // desired pixel format: "RGBA", "RGB", "BGRA" (default: "RGBA")
	IgnoreExif           bool         // ignore EXIF metadata
	IgnoreXMP            bool         // ignore XMP metadata
	IgnoreICC            bool         // ignore ICC profile (Note: ICC profile is not returned by decode, so this has no effect)
	ImageSizeLimit       uint32       // Maximum image size in total pixels (default: 268435456)
	ImageDimensionLimit  uint32       // Maximum image dimension (width or height), 0=ignore (default: 32768)
	StrictFlags          int          // Strict validation flags: 0=disabled, 1=enabled (default: 1)
	ChromaUpsampling     int          // 0=automatic (default), 1=fastest, 2=best_quality, 3=nearest, 4=bilinear

	// Image manipulation (for future implementation)
	CropX      int  // crop rectangle x
	CropY      int  // crop rectangle y
	CropWidth  int  // crop rectangle width
	CropHeight int  // crop rectangle height
	UseCrop    bool // enable cropping

	ResizeWidth  int  // resize width
	ResizeHeight int  // resize height
	UseResize    bool // enable resizing
}

// Command represents an AVIF decoder command that can be reused for multiple conversions
type AVIFDecCommand struct {
	cmd *C.AVIFDecCommand
}

// stringToPixelFormat converts a format string to NextImagePixelFormat
func stringToPixelFormat(format string) C.NextImagePixelFormat {
	switch format {
	case "RGB":
		return C.NEXTIMAGE_FORMAT_RGB
	case "BGRA":
		return C.NEXTIMAGE_FORMAT_BGRA
	case "RGBA", "":
		return C.NEXTIMAGE_FORMAT_RGBA
	default:
		return C.NEXTIMAGE_FORMAT_RGBA
	}
}

// pixelFormatToString converts NextImagePixelFormat to a string
func pixelFormatToString(format C.NextImagePixelFormat) string {
	switch format {
	case C.NEXTIMAGE_FORMAT_RGB:
		return "RGB"
	case C.NEXTIMAGE_FORMAT_BGRA:
		return "BGRA"
	case C.NEXTIMAGE_FORMAT_RGBA:
		return "RGBA"
	default:
		return "RGBA"
	}
}

// NewDefaultOptions creates a new Options struct with default values
func NewDefaultAVIFDecOptions() AVIFDecOptions {
	cOpts := C.avifdec_create_default_options()
	if cOpts == nil {
		// Return hardcoded defaults if C function fails
		return AVIFDecOptions{
			OutputFormat:        OutputPNG,
			JPEGQuality:         90,
			UseThreads:          false,
			Format:              "RGBA",
			IgnoreExif:          false,
			IgnoreXMP:           false,
			IgnoreICC:           false,
			ImageSizeLimit:      268435456,
			ImageDimensionLimit: 32768,
			StrictFlags:         1,
			ChromaUpsampling:    0,
		}
	}
	defer C.avifdec_free_options(cOpts)

	return AVIFDecOptions{
		OutputFormat:        OutputFormat(cOpts.output_format),
		JPEGQuality:         int(cOpts.jpeg_quality),
		UseThreads:          cOpts.use_threads != 0,
		Format:              pixelFormatToString(cOpts.format),
		IgnoreExif:          cOpts.ignore_exif != 0,
		IgnoreXMP:           cOpts.ignore_xmp != 0,
		IgnoreICC:           cOpts.ignore_icc != 0,
		ImageSizeLimit:      uint32(cOpts.image_size_limit),
		ImageDimensionLimit: uint32(cOpts.image_dimension_limit),
		StrictFlags:         int(cOpts.strict_flags),
		ChromaUpsampling:    int(cOpts.chroma_upsampling),
		CropX:               int(cOpts.crop_x),
		CropY:               int(cOpts.crop_y),
		CropWidth:           int(cOpts.crop_width),
		CropHeight:          int(cOpts.crop_height),
		UseCrop:             cOpts.use_crop != 0,
		ResizeWidth:         int(cOpts.resize_width),
		ResizeHeight:        int(cOpts.resize_height),
		UseResize:           cOpts.use_resize != 0,
	}
}

// optionsToCOptions converts Go Options to C AVIFDecOptions
func avifdecOptionsToCOptions(opts AVIFDecOptions) *C.AVIFDecOptions {
	cOpts := C.avifdec_create_default_options()
	if cOpts == nil {
		return nil
	}

	// Set output format options
	cOpts.output_format = C.AVIFDecOutputFormat(opts.OutputFormat)
	cOpts.jpeg_quality = C.int(opts.JPEGQuality)

	if opts.UseThreads {
		cOpts.use_threads = 1
	} else {
		cOpts.use_threads = 0
	}
	cOpts.format = stringToPixelFormat(opts.Format)
	if opts.IgnoreExif {
		cOpts.ignore_exif = 1
	} else {
		cOpts.ignore_exif = 0
	}
	if opts.IgnoreXMP {
		cOpts.ignore_xmp = 1
	} else {
		cOpts.ignore_xmp = 0
	}
	if opts.IgnoreICC {
		cOpts.ignore_icc = 1
	} else {
		cOpts.ignore_icc = 0
	}
	cOpts.image_size_limit = C.uint32_t(opts.ImageSizeLimit)
	cOpts.image_dimension_limit = C.uint32_t(opts.ImageDimensionLimit)
	cOpts.strict_flags = C.int(opts.StrictFlags)
	cOpts.chroma_upsampling = C.int(opts.ChromaUpsampling)

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

	return cOpts
}

// NewCommand creates a new AVIF decoder command with the given options.
// If opts is nil, default options are used.
// The returned Command must be closed with Close() when done.
func NewAVIFDecCommand(opts *AVIFDecOptions) (*AVIFDecCommand, error) {
	var cOpts *C.AVIFDecOptions
	if opts != nil {
		cOpts = avifdecOptionsToCOptions(*opts)
		if cOpts == nil {
			return nil, fmt.Errorf("failed to create options")
		}
	}

	cCmd := C.avifdec_new_command(cOpts)
	if cOpts != nil {
		C.avifdec_free_options(cOpts)
	}

	if cCmd == nil {
		errMsg := C.nextimage_last_error_message()
		return nil, fmt.Errorf("failed to create avifdec command: %s", C.GoString(errMsg))
	}

	cmd := &AVIFDecCommand{cmd: cCmd}
	runtime.SetFinalizer(cmd, func(c *AVIFDecCommand) {
		_ = c.Close()
	})
	return cmd, nil
}

// Run converts AVIF data to PNG format.
// This is the core method that performs the conversion.
func (c *AVIFDecCommand) Run(avifData []byte) ([]byte, error) {
	if c.cmd == nil {
		return nil, fmt.Errorf("command is closed")
	}
	if len(avifData) == 0 {
		return nil, fmt.Errorf("input data is empty")
	}

	var output C.NextImageBuffer
	C.memset(unsafe.Pointer(&output), 0, C.sizeof_NextImageBuffer)

	status := C.avifdec_run_command(
		c.cmd,
		(*C.uint8_t)(unsafe.Pointer(&avifData[0])),
		C.size_t(len(avifData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		errMsg := C.nextimage_last_error_message()
		return nil, fmt.Errorf("avifdec decoding failed (status %d): %s", status, C.GoString(errMsg))
	}

	if output.data == nil || output.size == 0 {
		return nil, fmt.Errorf("decoding produced empty output")
	}

	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))
	C.nextimage_free_buffer(&output)
	return result, nil
}

// RunFile reads an AVIF file, converts it to PNG, and writes the result to outputPath.
// This is sugar syntax over Run().
func (c *AVIFDecCommand) RunFile(inputPath, outputPath string) error {
	if c.cmd == nil {
		return fmt.Errorf("command is closed")
	}

	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	pngData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, pngData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// RunIO reads AVIF data from input, converts it to PNG, and writes the result to output.
// This is sugar syntax over Run().
func (c *AVIFDecCommand) RunIO(input io.Reader, output io.Writer) error {
	if c.cmd == nil {
		return fmt.Errorf("command is closed")
	}

	inputData, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	pngData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	if _, err := output.Write(pngData); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Close releases the resources associated with the command.
// After calling Close, the command cannot be used anymore.
func (c *AVIFDecCommand) Close() error {
	if c.cmd != nil {
		C.avifdec_free_command(c.cmd)
		c.cmd = nil
	}
	return nil
}
