package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ideamans/libnextimage/golang/cwebp"
)

func main() {
	// Parse command line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run jpeg_to_webp.go <input.jpg> <output.webp> [quality|lossless]")
		fmt.Println("\nExamples:")
		fmt.Println("  go run jpeg_to_webp.go photo.jpg photo.webp")
		fmt.Println("  go run jpeg_to_webp.go photo.jpg photo.webp 85")
		fmt.Println("  go run jpeg_to_webp.go photo.jpg photo.webp lossless")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// Read input JPEG file
	fmt.Printf("Reading JPEG file: %s\n", inputPath)
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Input file size: %d bytes (%.2f KB)\n", len(inputData), float64(len(inputData))/1024)

	// Configure WebP encoding options
	opts := cwebp.NewDefaultOptions()

	// Parse quality setting from command line
	if len(os.Args) > 3 {
		qualityArg := os.Args[3]
		if qualityArg == "lossless" {
			opts.Lossless = true
			fmt.Println("Encoding mode: Lossless")
		} else {
			quality, err := strconv.Atoi(qualityArg)
			if err != nil || quality < 0 || quality > 100 {
				fmt.Fprintf(os.Stderr, "Invalid quality value. Use 0-100 or 'lossless'\n")
				os.Exit(1)
			}
			opts.Quality = float32(quality)
			fmt.Printf("Encoding mode: Lossy (quality: %d)\n", quality)
		}
	} else {
		fmt.Printf("Encoding mode: Lossy (quality: %.0f, default)\n", opts.Quality)
	}

	// Create cwebp command
	cmd, err := cwebp.NewCommand(&opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating cwebp command: %v\n", err)
		os.Exit(1)
	}
	defer cmd.Close()

	// Convert JPEG to WebP
	fmt.Println("\nConverting to WebP...")
	webpData, err := cmd.Run(inputData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding WebP: %v\n", err)
		os.Exit(1)
	}

	// Write output WebP file
	fmt.Printf("Writing WebP file: %s\n", outputPath)
	if err := os.WriteFile(outputPath, webpData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	// Show results
	fmt.Printf("Output file size: %d bytes (%.2f KB)\n", len(webpData), float64(len(webpData))/1024)
	compression := (1.0 - float64(len(webpData))/float64(len(inputData))) * 100
	fmt.Printf("\n✅ Conversion successful!\n")
	fmt.Printf("Compression: %.1f%% smaller\n", compression)
	fmt.Printf("Original: %.2f KB → WebP: %.2f KB\n",
		float64(len(inputData))/1024,
		float64(len(webpData))/1024)
}
