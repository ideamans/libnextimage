package libnextimage

/*
#include "webp.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"runtime"
	"unsafe"
)

// WebPEncoder is an instance-based WebP encoder
// It can be reused for encoding multiple images with the same options.
type WebPEncoder struct {
	encoder *C.NextImageWebPEncoder
}

// NewWebPEncoder creates a new WebP encoder with the specified options.
// The encoder can be reused for encoding multiple images.
// It must be closed using Close() when no longer needed.
func NewWebPEncoder(opts WebPEncodeOptions) (*WebPEncoder, error) {
	clearError()

	cOpts := convertEncodeOptions(opts)
	encoder := C.nextimage_webp_encoder_create(&cOpts)
	if encoder == nil {
		return nil, makeError(C.NEXTIMAGE_ERROR_OUT_OF_MEMORY, "webp encoder create")
	}

	e := &WebPEncoder{encoder: encoder}

	// Set finalizer to ensure cleanup
	runtime.SetFinalizer(e, (*WebPEncoder).finalize)

	return e, nil
}

// Encode encodes image file data (JPEG, PNG, GIF, etc.) to WebP format.
// This method can be called multiple times with different input data.
func (e *WebPEncoder) Encode(imageFileData []byte) ([]byte, error) {
	if e.encoder == nil {
		return nil, fmt.Errorf("webp encoder: encoder is closed")
	}

	clearError()

	if len(imageFileData) == 0 {
		return nil, fmt.Errorf("webp encoder encode: empty input data")
	}

	var output C.NextImageBuffer

	status := C.nextimage_webp_encoder_encode(
		e.encoder,
		(*C.uint8_t)(unsafe.Pointer(&imageFileData[0])),
		C.size_t(len(imageFileData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp encoder encode")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))

	// Free C buffer
	freeEncodeBuffer(&output)

	return result, nil
}

// EncodeFile encodes an image file to WebP format.
// This is a convenience method that reads the file and calls Encode.
func (e *WebPEncoder) EncodeFile(inputPath string) ([]byte, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("webp encoder encode file: %w", err)
	}

	return e.Encode(data)
}

// Close releases the resources associated with the encoder.
// After calling Close, the encoder cannot be used anymore.
func (e *WebPEncoder) Close() {
	if e.encoder != nil {
		C.nextimage_webp_encoder_destroy(e.encoder)
		e.encoder = nil
		runtime.SetFinalizer(e, nil)
	}
}

// finalize is called by the garbage collector if Close was not called
func (e *WebPEncoder) finalize() {
	e.Close()
}

// WebPDecoder is an instance-based WebP decoder
// It can be reused for decoding multiple images with the same options.
type WebPDecoder struct {
	decoder *C.NextImageWebPDecoder
}

// NewWebPDecoder creates a new WebP decoder with the specified options.
// The decoder can be reused for decoding multiple images.
// It must be closed using Close() when no longer needed.
func NewWebPDecoder(opts WebPDecodeOptions) (*WebPDecoder, error) {
	clearError()

	cOpts := convertDecodeOptions(opts)
	decoder := C.nextimage_webp_decoder_create(&cOpts)
	if decoder == nil {
		return nil, makeError(C.NEXTIMAGE_ERROR_OUT_OF_MEMORY, "webp decoder create")
	}

	d := &WebPDecoder{decoder: decoder}

	// Set finalizer to ensure cleanup
	runtime.SetFinalizer(d, (*WebPDecoder).finalize)

	return d, nil
}

// Decode decodes WebP data to pixel data.
// This method can be called multiple times with different input data.
func (d *WebPDecoder) Decode(webpData []byte) (*DecodedImage, error) {
	if d.decoder == nil {
		return nil, fmt.Errorf("webp decoder: decoder is closed")
	}

	clearError()

	if len(webpData) == 0 {
		return nil, fmt.Errorf("webp decoder decode: empty input data")
	}

	var output C.NextImageDecodeBuffer

	status := C.nextimage_webp_decoder_decode(
		d.decoder,
		(*C.uint8_t)(unsafe.Pointer(&webpData[0])),
		C.size_t(len(webpData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "webp decoder decode")
	}

	// Convert to Go struct
	img := convertDecodeBuffer(&output)

	// Free C buffer
	freeDecodeBuffer(&output)

	return img, nil
}

// DecodeFile decodes a WebP file to pixel data.
// This is a convenience method that reads the file and calls Decode.
func (d *WebPDecoder) DecodeFile(inputPath string) (*DecodedImage, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("webp decoder decode file: %w", err)
	}

	return d.Decode(data)
}

// Close releases the resources associated with the decoder.
// After calling Close, the decoder cannot be used anymore.
func (d *WebPDecoder) Close() {
	if d.decoder != nil {
		C.nextimage_webp_decoder_destroy(d.decoder)
		d.decoder = nil
		runtime.SetFinalizer(d, nil)
	}
}

// finalize is called by the garbage collector if Close was not called
func (d *WebPDecoder) finalize() {
	d.Close()
}
