package libnextimage

/*
#include "nextimage.h"
#include "webp.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

// GIF2WebPEncodeBytes converts GIF image data to WebP format
// Note: libwebp's imageio does not support GIF format, so this function will return an error
func GIF2WebPEncodeBytes(gifData []byte, opts WebPEncodeOptions) ([]byte, error) {
	if len(gifData) == 0 {
		return nil, fmt.Errorf("gif2webp: empty input data")
	}

	clearError()

	// Setup options
	var cOpts C.NextImageWebPEncodeOptions
	C.nextimage_webp_default_encode_options(&cOpts)
	cOpts.quality = C.float(opts.Quality)
	if opts.Lossless {
		cOpts.lossless = 1
	} else {
		cOpts.lossless = 0
	}
	cOpts.method = C.int(opts.Method)

	// Call C function
	var output C.NextImageEncodeBuffer
	status := C.nextimage_gif2webp_alloc(
		(*C.uint8_t)(unsafe.Pointer(&gifData[0])),
		C.size_t(len(gifData)),
		&cOpts,
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "gif2webp encode")
	}

	// Copy result to Go slice
	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))
	freeEncodeBuffer(&output)

	return result, nil
}

// GIF2WebPEncodeFile reads a GIF file and converts it to WebP format
// Note: libwebp's imageio does not support GIF format, so this function will return an error
func GIF2WebPEncodeFile(gifPath string, opts WebPEncodeOptions) ([]byte, error) {
	data, err := os.ReadFile(gifPath)
	if err != nil {
		return nil, fmt.Errorf("gif2webp: failed to read file: %w", err)
	}

	return GIF2WebPEncodeBytes(data, opts)
}

// WebP2GIFConvertBytes converts WebP image data to GIF format
// Uses 256-color quantization with 6x6x6 RGB cube + grayscale
// Supports transparency
func WebP2GIFConvertBytes(webpData []byte) ([]byte, error) {
	if len(webpData) == 0 {
		return nil, fmt.Errorf("webp2gif: empty input data")
	}

	clearError()

	// Call C function
	var output C.NextImageEncodeBuffer
	status := C.nextimage_webp2gif_alloc(
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp2gif convert")
	}

	// Copy result to Go slice
	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))
	freeEncodeBuffer(&output)

	return result, nil
}

// WebP2GIFConvertFile reads a WebP file and converts it to GIF format
func WebP2GIFConvertFile(webpPath string) ([]byte, error) {
	data, err := os.ReadFile(webpPath)
	if err != nil {
		return nil, fmt.Errorf("webp2gif: failed to read file: %w", err)
	}

	return WebP2GIFConvertBytes(data)
}
