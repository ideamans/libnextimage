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
#include "nextimage/webp2gif.h"
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"
)

// Options represents WebP to GIF conversion options.
// Currently minimal options - reserved for future extensions.
type WebP2GifOptions struct {
	Reserved int // reserved for future use
}

// Command represents a webp2gif command instance that can be reused for multiple conversions.
type WebP2GifCommand struct {
	cmd *C.WebP2GifCommand
}

// NewDefaultOptions creates default WebP to GIF conversion options.
func NewDefaultWebP2GifOptions() WebP2GifOptions {
	cOpts := C.webp2gif_create_default_options()
	if cOpts == nil {
		return WebP2GifOptions{Reserved: 0}
	}
	defer C.webp2gif_free_options(cOpts)

	return WebP2GifOptions{
		Reserved: int(cOpts.reserved),
	}
}

// optionsToCOptions converts Go Options to C WebP2GifOptions
func webp2gifOptionsToCOptions(opts WebP2GifOptions) *C.WebP2GifOptions {
	cOpts := C.webp2gif_create_default_options()
	if cOpts == nil {
		return nil
	}

	cOpts.reserved = C.int(opts.Reserved)

	return cOpts
}

// NewCommand creates a new webp2gif command with the given options.
// If opts is nil, default options are used.
// The returned Command must be closed with Close() when done.
func NewWebP2GifCommand(opts *WebP2GifOptions) (*WebP2GifCommand, error) {
	var cOpts *C.WebP2GifOptions
	if opts != nil {
		cOpts = webp2gifOptionsToCOptions(*opts)
		if cOpts == nil {
			return nil, fmt.Errorf("failed to create options")
		}
	}

	cCmd := C.webp2gif_new_command(cOpts)
	if cOpts != nil {
		C.webp2gif_free_options(cOpts)
	}

	if cCmd == nil {
		errMsg := C.nextimage_last_error_message()
		return nil, fmt.Errorf("failed to create webp2gif command: %s", C.GoString(errMsg))
	}

	cmd := &WebP2GifCommand{cmd: cCmd}
	runtime.SetFinalizer(cmd, func(c *WebP2GifCommand) {
		_ = c.Close()
	})
	return cmd, nil
}

// Run converts WebP data to GIF format.
// This is the core method that performs the conversion.
func (c *WebP2GifCommand) Run(webpData []byte) ([]byte, error) {
	if c.cmd == nil {
		return nil, fmt.Errorf("command is closed")
	}
	if len(webpData) == 0 {
		return nil, fmt.Errorf("input data is empty")
	}

	var output C.NextImageBuffer
	C.memset(unsafe.Pointer(&output), 0, C.sizeof_NextImageBuffer)

	status := C.webp2gif_run_command(
		c.cmd,
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		errMsg := C.nextimage_last_error_message()
		return nil, fmt.Errorf("webp2gif conversion failed (status %d): %s", status, C.GoString(errMsg))
	}

	if output.data == nil || output.size == 0 {
		return nil, fmt.Errorf("conversion produced empty output")
	}

	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))
	C.nextimage_free_buffer(&output)
	return result, nil
}

// RunFile reads a WebP file, converts it to GIF, and writes the result to outputPath.
// This is sugar syntax over Run().
func (c *WebP2GifCommand) RunFile(inputPath, outputPath string) error {
	if c.cmd == nil {
		return fmt.Errorf("command is closed")
	}

	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	gifData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, gifData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// RunIO reads WebP data from input, converts it to GIF, and writes the result to output.
// This is sugar syntax over Run().
func (c *WebP2GifCommand) RunIO(input io.Reader, output io.Writer) error {
	if c.cmd == nil {
		return fmt.Errorf("command is closed")
	}

	inputData, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	gifData, err := c.Run(inputData)
	if err != nil {
		return err
	}

	if _, err := output.Write(gifData); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Close releases the resources associated with the command.
// After calling Close, the command cannot be used anymore.
func (c *WebP2GifCommand) Close() error {
	if c.cmd != nil {
		C.webp2gif_free_command(c.cmd)
		c.cmd = nil
	}
	return nil
}
