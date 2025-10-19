// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ideamans/libnextimage/golang/avifenc"
)

func main() {
	// Read PNG file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		fmt.Printf("Failed to read PNG: %v\n", err)
		return
	}

	// Create AVIF encoder command
	cmd, err := avifenc.NewCommand(nil)
	if err != nil {
		fmt.Printf("Failed to create avifenc command: %v\n", err)
		return
	}
	defer cmd.Close()

	// Encode to AVIF
	avifData, err := cmd.Run(pngData)
	if err != nil {
		fmt.Printf("Failed to encode: %v\n", err)
		return
	}

	// Save AVIF file
	avifDir := filepath.Join("..", "..", "testdata", "avif")
	os.MkdirAll(avifDir, 0755)

	avifPath := filepath.Join(avifDir, "red.avif")
	err = os.WriteFile(avifPath, avifData, 0644)
	if err != nil {
		fmt.Printf("Failed to write AVIF: %v\n", err)
		return
	}

	fmt.Printf("Created %s (%d bytes)\n", avifPath, len(avifData))
}
