// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

/*
#cgo CFLAGS: -I${SRCDIR}/../../include
#include <stdlib.h>
#include <string.h>
#include "nextimage.h"
#include "nextimage/cwebp.h"
*/
import "C"
import "unsafe"

func main() {
	// Read PNG file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		fmt.Printf("Failed to read PNG: %v\n", err)
		return
	}

	// Create WebP encoder command
	cmd := C.cwebp_new_command(nil)
	if cmd == nil {
		fmt.Println("Failed to create cwebp command")
		return
	}
	defer C.cwebp_free_command(cmd)

	// Encode to WebP
	var output C.NextImageBuffer
	C.memset(unsafe.Pointer(&output), 0, C.sizeof_NextImageBuffer)

	status := C.cwebp_run_command(
		cmd,
		(*C.uint8_t)(unsafe.Pointer(&pngData[0])),
		C.size_t(len(pngData)),
		&output,
	)

	if status != C.NEXTIMAGE_OK {
		fmt.Printf("Failed to encode: status %d\n", status)
		return
	}

	// Save WebP file
	webpData := C.GoBytes(unsafe.Pointer(output.data), C.int(output.size))
	C.nextimage_free_buffer(&output)

	webpDir := filepath.Join("..", "..", "testdata", "webp")
	os.MkdirAll(webpDir, 0755)

	webpPath := filepath.Join(webpDir, "gradient.webp")
	err = os.WriteFile(webpPath, webpData, 0644)
	if err != nil {
		fmt.Printf("Failed to write WebP: %v\n", err)
		return
	}

	fmt.Printf("Created %s (%d bytes)\n", webpPath, len(webpData))
}
