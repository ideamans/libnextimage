package libnextimage

/*
#include "webp.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

// WebPEncodeOptions represents WebP encoding options
type WebPEncodeOptions struct {
	Quality          float32 // 0-100, default 75
	Lossless         bool    // default false
	Method           int     // 0-6, default 4 (quality/speed trade-off)
	TargetSize       int     // target size in bytes (0 = disabled)
	TargetPSNR       float32 // target PSNR (0 = disabled)
	Exact            bool    // preserve RGB values in transparent area
	AlphaCompression bool    // compress alpha channel (default true)
	AlphaQuality     int     // 0-100, transparency compression quality
	Pass             int     // number of entropy-analysis passes (1-10)
	Preprocessing    int     // 0=none, 1=segment-smooth, 2=pseudo-random dithering
	Partitions       int     // 0-3, log2(number of token partitions)
	PartitionLimit   int     // quality degradation allowed (0-100)
}

// WebPDecodeOptions represents WebP decoding options
type WebPDecodeOptions struct {
	UseThreads         bool        // enable multi-threading
	BypassFiltering    bool        // disable in-loop filtering
	NoFancyUpsampling  bool        // use faster pointwise upsampler
	Format             PixelFormat // desired pixel format (default: RGBA)
}

// DefaultWebPEncodeOptions returns default WebP encoding options
func DefaultWebPEncodeOptions() WebPEncodeOptions {
	return WebPEncodeOptions{
		Quality:          75.0,
		Lossless:         false,
		Method:           4,
		TargetSize:       0,
		TargetPSNR:       0.0,
		Exact:            false,
		AlphaCompression: true,
		AlphaQuality:     100,
		Pass:             1,
		Preprocessing:    0,
		Partitions:       0,
		PartitionLimit:   0,
	}
}

// DefaultWebPDecodeOptions returns default WebP decoding options
func DefaultWebPDecodeOptions() WebPDecodeOptions {
	return WebPDecodeOptions{
		UseThreads:        false,
		BypassFiltering:   false,
		NoFancyUpsampling: false,
		Format:            FormatRGBA,
	}
}

// convertEncodeOptions converts Go options to C options
func convertEncodeOptions(opts WebPEncodeOptions) C.NextImageWebPEncodeOptions {
	var cOpts C.NextImageWebPEncodeOptions
	C.nextimage_webp_default_encode_options(&cOpts)

	cOpts.quality = C.float(opts.Quality)
	if opts.Lossless {
		cOpts.lossless = 1
	} else {
		cOpts.lossless = 0
	}
	cOpts.method = C.int(opts.Method)
	cOpts.target_size = C.int(opts.TargetSize)
	cOpts.target_psnr = C.float(opts.TargetPSNR)
	if opts.Exact {
		cOpts.exact = 1
	} else {
		cOpts.exact = 0
	}
	if opts.AlphaCompression {
		cOpts.alpha_compression = 1
	} else {
		cOpts.alpha_compression = 0
	}
	cOpts.alpha_quality = C.int(opts.AlphaQuality)
	cOpts.pass = C.int(opts.Pass)
	cOpts.preprocessing = C.int(opts.Preprocessing)
	cOpts.partitions = C.int(opts.Partitions)
	cOpts.partition_limit = C.int(opts.PartitionLimit)

	return cOpts
}

// convertDecodeOptions converts Go options to C options
func convertDecodeOptions(opts WebPDecodeOptions) C.NextImageWebPDecodeOptions {
	var cOpts C.NextImageWebPDecodeOptions
	C.nextimage_webp_default_decode_options(&cOpts)

	if opts.UseThreads {
		cOpts.use_threads = 1
	} else {
		cOpts.use_threads = 0
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
	cOpts.format = C.NextImagePixelFormat(opts.Format)

	return cOpts
}

// WebPEncodeBytes encodes image file data (JPEG, PNG, GIF, etc.) to WebP format
// This is equivalent to the cwebp command-line tool.
// The input data should be a complete image file (JPEG, PNG, GIF, TIFF, WebP, etc.)
// not raw pixel data.
func WebPEncodeBytes(imageFileData []byte, opts WebPEncodeOptions) ([]byte, error) {
	clearError()

	if len(imageFileData) == 0 {
		return nil, fmt.Errorf("webp encode: empty input data")
	}

	cOpts := convertEncodeOptions(opts)
	var encoded C.NextImageEncodeBuffer

	status := C.nextimage_webp_encode_alloc(
		(*C.uint8_t)(unsafe.Pointer(&imageFileData[0])),
		C.size_t(len(imageFileData)),
		&cOpts,
		&encoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp encode")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(encoded.data), C.int(encoded.size))

	// Free C buffer
	freeEncodeBuffer(&encoded)

	return result, nil
}

// WebPEncodeFile encodes an image file to WebP format
// This reads the image file (JPEG, PNG, GIF, etc.) and converts it to WebP.
func WebPEncodeFile(inputPath string, opts WebPEncodeOptions) ([]byte, error) {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("webp encode file: %w", err)
	}

	return WebPEncodeBytes(data, opts)
}

// WebPDecodeBytes decodes WebP data to pixel data
func WebPDecodeBytes(webpData []byte, opts WebPDecodeOptions) (*DecodedImage, error) {
	clearError()

	if len(webpData) == 0 {
		return nil, fmt.Errorf("webp decode: empty input data")
	}

	cOpts := convertDecodeOptions(opts)
	var decoded C.NextImageDecodeBuffer

	status := C.nextimage_webp_decode_alloc(
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&cOpts,
		&decoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp decode")
	}

	// Convert to Go struct
	img := convertDecodeBuffer(&decoded)

	// Free C buffer
	freeDecodeBuffer(&decoded)

	return img, nil
}

// WebPDecodeFile decodes a WebP file to pixel data
func WebPDecodeFile(inputPath string, opts WebPDecodeOptions) (*DecodedImage, error) {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("webp decode file: %w", err)
	}

	return WebPDecodeBytes(data, opts)
}

// WebPDecodeToFile decodes WebP data and writes to a file
func WebPDecodeToFile(webpData []byte, outputPath string, opts WebPDecodeOptions) error {
	img, err := WebPDecodeBytes(webpData, opts)
	if err != nil {
		return err
	}

	// For now, just write raw pixel data
	// TODO: Implement proper image encoding based on output format
	err = os.WriteFile(outputPath, img.Data, 0644)
	if err != nil {
		return fmt.Errorf("webp decode to file: %w", err)
	}

	return nil
}

// WebPDecodeSize calculates required buffer size for decoding
func WebPDecodeSize(webpData []byte) (width, height int, requiredSize int, err error) {
	clearError()

	if len(webpData) == 0 {
		return 0, 0, 0, fmt.Errorf("webp decode size: empty input data")
	}

	var w, h C.int
	var size C.size_t

	status := C.nextimage_webp_decode_size(
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&w,
		&h,
		&size,
	)

	if status != C.NEXTIMAGE_OK {
		return 0, 0, 0, makeError(status, "webp decode size")
	}

	return int(w), int(h), int(size), nil
}

// WebPDecodeInto decodes WebP data into a user-provided buffer
func WebPDecodeInto(webpData []byte, buffer []byte, opts WebPDecodeOptions) (*DecodedImage, error) {
	clearError()

	if len(webpData) == 0 {
		return nil, fmt.Errorf("webp decode into: empty input data")
	}
	if len(buffer) == 0 {
		return nil, fmt.Errorf("webp decode into: empty buffer")
	}

	// Due to CGO pointer rules, we cannot pass Go buffer directly to C
	// So we allocate in C first, then copy to user buffer
	img, err := WebPDecodeBytes(webpData, opts)
	if err != nil {
		return nil, fmt.Errorf("webp decode into: %w", err)
	}

	// Check buffer size
	if len(buffer) < len(img.Data) {
		return nil, fmt.Errorf("webp decode into: buffer too small (need %d bytes, have %d bytes)", len(img.Data), len(buffer))
	}

	// Copy decoded data to user buffer
	copy(buffer, img.Data)

	// Return metadata with user buffer reference
	return &DecodedImage{
		Data:     buffer[:len(img.Data)],
		Stride:   img.Stride,
		UPlane:   nil,
		UStride:  0,
		VPlane:   nil,
		VStride:  0,
		Width:    img.Width,
		Height:   img.Height,
		BitDepth: img.BitDepth,
		Format:   img.Format,
	}, nil
}

// GIF2WebP converts GIF data to WebP format
func GIF2WebP(gifData []byte, opts WebPEncodeOptions) ([]byte, error) {
	clearError()

	if len(gifData) == 0 {
		return nil, fmt.Errorf("gif2webp: empty input data")
	}

	cOpts := convertEncodeOptions(opts)
	var encoded C.NextImageEncodeBuffer

	status := C.nextimage_gif2webp_alloc(
		(*C.uint8_t)(unsafe.Pointer(&gifData[0])),
		C.size_t(len(gifData)),
		&cOpts,
		&encoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "gif2webp")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(encoded.data), C.int(encoded.size))

	// Free C buffer
	freeEncodeBuffer(&encoded)

	return result, nil
}

// WebP2GIF converts WebP data to GIF format
func WebP2GIF(webpData []byte) ([]byte, error) {
	clearError()

	if len(webpData) == 0 {
		return nil, fmt.Errorf("webp2gif: empty input data")
	}

	var encoded C.NextImageEncodeBuffer

	status := C.nextimage_webp2gif_alloc(
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&encoded,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp2gif")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(encoded.data), C.int(encoded.size))

	// Free C buffer
	freeEncodeBuffer(&encoded)

	return result, nil
}
