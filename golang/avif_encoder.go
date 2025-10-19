package libnextimage

/*
#include "avif.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"runtime"
	"unsafe"
)

// AVIFEncoder is an instance-based AVIF encoder
// It can be reused for encoding multiple images with the same options.
type AVIFEncoder struct {
	encoder *C.NextImageAVIFEncoder
}

// NewAVIFEncoder creates a new AVIF encoder with the specified options.
// The encoder can be reused for encoding multiple images.
// It must be closed using Close() when no longer needed.
func NewAVIFEncoder(opts AVIFEncodeOptions) (*AVIFEncoder, error) {
	clearError()

	cOpts := opts.toCEncodeOptions()
	encoder := C.nextimage_avif_encoder_create(&cOpts)
	if encoder == nil {
		return nil, makeError(C.NEXTIMAGE_ERROR_OUT_OF_MEMORY, "avif encoder create")
	}

	e := &AVIFEncoder{encoder: encoder}

	// Set finalizer to ensure cleanup
	runtime.SetFinalizer(e, (*AVIFEncoder).finalize)

	return e, nil
}

// Encode encodes image file data (JPEG, PNG, etc.) to AVIF format.
// This method can be called multiple times with different input data.
func (e *AVIFEncoder) Encode(imageFileData []byte) ([]byte, error) {
	if e.encoder == nil {
		return nil, fmt.Errorf("avif encoder: encoder is closed")
	}

	clearError()

	if len(imageFileData) == 0 {
		return nil, fmt.Errorf("avif encoder encode: empty input data")
	}

	var output C.NextImageBuffer

	status := C.nextimage_avif_encoder_encode(
		e.encoder,
		(*C.uint8_t)(unsafe.Pointer(&imageFileData[0])),
		C.size_t(len(imageFileData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "avif encoder encode")
	}

	// Copy data to Go slice
	result := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))

	// Free C buffer
	freeEncodeBuffer(&output)

	return result, nil
}

// EncodeFile encodes an image file to AVIF format.
// This is a convenience method that reads the file and calls Encode.
func (e *AVIFEncoder) EncodeFile(inputPath string) ([]byte, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("avif encoder encode file: %w", err)
	}

	return e.Encode(data)
}

// Close releases the resources associated with the encoder.
// After calling Close, the encoder cannot be used anymore.
func (e *AVIFEncoder) Close() {
	if e.encoder != nil {
		C.nextimage_avif_encoder_destroy(e.encoder)
		e.encoder = nil
		runtime.SetFinalizer(e, nil)
	}
}

// finalize is called by the garbage collector if Close was not called
func (e *AVIFEncoder) finalize() {
	e.Close()
}

// AVIFDecoder is an instance-based AVIF decoder
// It can be reused for decoding multiple images with the same options.
type AVIFDecoder struct {
	decoder *C.NextImageAVIFDecoder
}

// NewAVIFDecoder creates a new AVIF decoder with the specified options.
// The decoder can be reused for decoding multiple images.
// It must be closed using Close() when no longer needed.
func NewAVIFDecoder(opts AVIFDecodeOptions) (*AVIFDecoder, error) {
	clearError()

	cOpts := opts.toCDecodeOptions()
	decoder := C.nextimage_avif_decoder_create(&cOpts)
	if decoder == nil {
		return nil, makeError(C.NEXTIMAGE_ERROR_OUT_OF_MEMORY, "avif decoder create")
	}

	d := &AVIFDecoder{decoder: decoder}

	// Set finalizer to ensure cleanup
	runtime.SetFinalizer(d, (*AVIFDecoder).finalize)

	return d, nil
}

// Decode decodes AVIF data to pixel data.
// This method can be called multiple times with different input data.
func (d *AVIFDecoder) Decode(avifData []byte) (*DecodedImage, error) {
	if d.decoder == nil {
		return nil, fmt.Errorf("avif decoder: decoder is closed")
	}

	clearError()

	if len(avifData) == 0 {
		return nil, fmt.Errorf("avif decoder decode: empty input data")
	}

	var output C.NextImageDecodeBuffer

	status := C.nextimage_avif_decoder_decode(
		d.decoder,
		(*C.uint8_t)(unsafe.Pointer(&avifData[0])),
		C.size_t(len(avifData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		return nil, makeError(status, "avif decoder decode")
	}

	// Convert to Go struct
	img := convertDecodeBuffer(&output)

	// Free C buffer
	freeDecodeBuffer(&output)

	return img, nil
}

// DecodeFile decodes an AVIF file to pixel data.
// This is a convenience method that reads the file and calls Decode.
func (d *AVIFDecoder) DecodeFile(inputPath string) (*DecodedImage, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("avif decoder decode file: %w", err)
	}

	return d.Decode(data)
}

// Close releases the resources associated with the decoder.
// After calling Close, the decoder cannot be used anymore.
func (d *AVIFDecoder) Close() {
	if d.decoder != nil {
		C.nextimage_avif_decoder_destroy(d.decoder)
		d.decoder = nil
		runtime.SetFinalizer(d, nil)
	}
}

// finalize is called by the garbage collector if Close was not called
func (d *AVIFDecoder) finalize() {
	d.Close()
}
