package webp2gif

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
type Options struct {
	Reserved int // reserved for future use
}

// Command represents a webp2gif command instance that can be reused for multiple conversions.
type Command struct {
	cmd *C.WebP2GifCommand
}

// NewDefaultOptions creates default WebP to GIF conversion options.
func NewDefaultOptions() Options {
	cOpts := C.webp2gif_create_default_options()
	if cOpts == nil {
		return Options{Reserved: 0}
	}
	defer C.webp2gif_free_options(cOpts)

	return Options{
		Reserved: int(cOpts.reserved),
	}
}

// optionsToCOptions converts Go Options to C WebP2GifOptions
func optionsToCOptions(opts Options) *C.WebP2GifOptions {
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
func NewCommand(opts *Options) (*Command, error) {
	var cOpts *C.WebP2GifOptions
	if opts != nil {
		cOpts = optionsToCOptions(*opts)
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

	cmd := &Command{cmd: cCmd}
	runtime.SetFinalizer(cmd, (*Command).Close)
	return cmd, nil
}

// Run converts WebP data to GIF format.
// This is the core method that performs the conversion.
func (c *Command) Run(webpData []byte) ([]byte, error) {
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
func (c *Command) RunFile(inputPath, outputPath string) error {
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
func (c *Command) RunIO(input io.Reader, output io.Writer) error {
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
func (c *Command) Close() error {
	if c.cmd != nil {
		C.webp2gif_free_command(c.cmd)
		c.cmd = nil
	}
	return nil
}
