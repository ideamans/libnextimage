package libnextimage

/*
#cgo CFLAGS: -I${SRCDIR}/../c/include

// macOS ARM64: Link to combined static library in lib/darwin-arm64
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/../lib/darwin-arm64 -lnextimage
#cgo darwin,arm64 LDFLAGS: /opt/homebrew/lib/libjpeg.a /opt/homebrew/lib/libpng.a -lz
#cgo darwin,arm64 LDFLAGS: -framework CoreFoundation

// macOS Intel: Link to combined static library in lib/darwin-amd64
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/../lib/darwin-amd64 -lnextimage
#cgo darwin,amd64 LDFLAGS: /opt/homebrew/lib/libjpeg.a /opt/homebrew/lib/libpng.a -lz
#cgo darwin,amd64 LDFLAGS: -framework CoreFoundation

// Linux x64: Link to combined static library in lib/linux-amd64
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../lib/linux-amd64 -lnextimage
#cgo linux,amd64 LDFLAGS: -ljpeg -lpng -lz
#cgo linux,amd64 LDFLAGS: -lpthread -lm -ldl

// Linux ARM64: Link to combined static library in lib/linux-arm64
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../lib/linux-arm64 -lnextimage
#cgo linux,arm64 LDFLAGS: -ljpeg -lpng -lz
#cgo linux,arm64 LDFLAGS: -lpthread -lm -ldl

// Windows x64: Link to combined static library in lib/windows-amd64
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/../lib/windows-amd64 -lnextimage
#cgo windows,amd64 LDFLAGS: -ljpeg -lpng -lz
#cgo windows,amd64 LDFLAGS: -lws2_32 -lkernel32 -luser32

// Other platforms: Link to combined static library in lib/other
#cgo !darwin,!linux,!windows LDFLAGS: -L${SRCDIR}/../lib/other -lnextimage
#cgo !darwin,!linux,!windows LDFLAGS: -ljpeg -lpng -lz
#cgo !darwin,!linux,!windows LDFLAGS: -lpthread -lm

#include "nextimage.h"
#include "webp.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// PixelFormat represents the pixel format of an image
type PixelFormat int

const (
	FormatRGBA   PixelFormat = C.NEXTIMAGE_FORMAT_RGBA
	FormatRGB    PixelFormat = C.NEXTIMAGE_FORMAT_RGB
	FormatBGRA   PixelFormat = C.NEXTIMAGE_FORMAT_BGRA
	FormatYUV420 PixelFormat = C.NEXTIMAGE_FORMAT_YUV420
	FormatYUV422 PixelFormat = C.NEXTIMAGE_FORMAT_YUV422
	FormatYUV444 PixelFormat = C.NEXTIMAGE_FORMAT_YUV444
)

// DecodedImage represents a decoded image with pixel data
type DecodedImage struct {
	// Primary plane (full data for interleaved formats, Y plane for planar)
	Data   []byte
	Stride int

	// UV planes (for YUV planar formats only)
	UPlane  []byte
	UStride int
	VPlane  []byte
	VStride int

	// Metadata
	Width    int
	Height   int
	BitDepth int
	Format   PixelFormat
}

// IsPlanar returns true if the image uses planar format
func (img *DecodedImage) IsPlanar() bool {
	return img.UPlane != nil && img.VPlane != nil
}

// IsHighBitDepth returns true if the image uses more than 8 bits per channel
func (img *DecodedImage) IsHighBitDepth() bool {
	return img.BitDepth > 8
}

// getLastError retrieves the last error message from C
func getLastError() string {
	msg := C.nextimage_last_error_message()
	if msg == nil {
		return "unknown error"
	}
	return C.GoString(msg)
}

// clearError clears the last error message
func clearError() {
	C.nextimage_clear_error()
}

// makeError creates a Go error from C error status
func makeError(status C.NextImageStatus, operation string) error {
	if status == C.NEXTIMAGE_OK {
		return nil
	}

	errMsg := getLastError()
	if errMsg == "unknown error" {
		switch status {
		case C.NEXTIMAGE_ERROR_INVALID_PARAM:
			errMsg = "invalid parameter"
		case C.NEXTIMAGE_ERROR_ENCODE_FAILED:
			errMsg = "encoding failed"
		case C.NEXTIMAGE_ERROR_DECODE_FAILED:
			errMsg = "decoding failed"
		case C.NEXTIMAGE_ERROR_OUT_OF_MEMORY:
			errMsg = "out of memory"
		case C.NEXTIMAGE_ERROR_UNSUPPORTED:
			errMsg = "unsupported operation"
		case C.NEXTIMAGE_ERROR_BUFFER_TOO_SMALL:
			errMsg = "buffer too small"
		}
	}

	return fmt.Errorf("%s: %s", operation, errMsg)
}

// Version returns the library version
func Version() string {
	return C.GoString(C.nextimage_version())
}

// freeEncodeBuffer safely frees an encode buffer
func freeEncodeBuffer(buf *C.NextImageEncodeBuffer) {
	if buf != nil {
		C.nextimage_free_encode_buffer(buf)
	}
}

// freeDecodeBuffer safely frees a decode buffer
func freeDecodeBuffer(buf *C.NextImageDecodeBuffer) {
	if buf != nil {
		C.nextimage_free_decode_buffer(buf)
	}
}

// convertDecodeBuffer converts C decode buffer to Go DecodedImage
func convertDecodeBuffer(cbuf *C.NextImageDecodeBuffer) *DecodedImage {
	img := &DecodedImage{
		Width:    int(cbuf.width),
		Height:   int(cbuf.height),
		BitDepth: int(cbuf.bit_depth),
		Format:   PixelFormat(cbuf.format),
		Stride:   int(cbuf.stride),
	}

	// Copy primary data
	if cbuf.data != nil && cbuf.data_size > 0 {
		img.Data = C.GoBytes(unsafe.Pointer(cbuf.data), C.int(cbuf.data_size))
	}

	// Copy UV planes if present
	if cbuf.u_plane != nil && cbuf.u_size > 0 {
		img.UPlane = C.GoBytes(unsafe.Pointer(cbuf.u_plane), C.int(cbuf.u_size))
		img.UStride = int(cbuf.u_stride)
	}

	if cbuf.v_plane != nil && cbuf.v_size > 0 {
		img.VPlane = C.GoBytes(unsafe.Pointer(cbuf.v_plane), C.int(cbuf.v_size))
		img.VStride = int(cbuf.v_stride)
	}

	return img
}
