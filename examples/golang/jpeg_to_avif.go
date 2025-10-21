package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ideamans/libnextimage/golang"
)

func main() {
	// Parse command line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run jpeg_to_avif.go <input.jpg> <output.avif> [quality] [speed]")
		fmt.Println("\nArguments:")
		fmt.Println("  quality: 0-100 (default: 60, lower = smaller file)")
		fmt.Println("  speed:   0-10 (default: 6, higher = faster encoding)")
		fmt.Println("\nExamples:")
		fmt.Println("  go run jpeg_to_avif.go photo.jpg photo.avif")
		fmt.Println("  go run jpeg_to_avif.go photo.jpg photo.avif 50")
		fmt.Println("  go run jpeg_to_avif.go photo.jpg photo.avif 70 8")
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

	// Configure AVIF encoding options
	opts := libnextimage.DefaultAVIFEncodeOptions()

	// Parse quality setting
	if len(os.Args) > 3 {
		quality, err := strconv.Atoi(os.Args[3])
		if err != nil || quality < 0 || quality > 100 {
			fmt.Fprintf(os.Stderr, "Invalid quality value. Use 0-100\n")
			os.Exit(1)
		}
		opts.Quality = quality
	}

	// Parse speed setting
	if len(os.Args) > 4 {
		speed, err := strconv.Atoi(os.Args[4])
		if err != nil || speed < 0 || speed > 10 {
			fmt.Fprintf(os.Stderr, "Invalid speed value. Use 0-10\n")
			os.Exit(1)
		}
		opts.Speed = speed
	}

	fmt.Printf("Encoding settings:\n")
	fmt.Printf("  Quality: %d (0=smallest, 100=best)\n", opts.Quality)
	fmt.Printf("  Speed: %d (0=slowest, 10=fastest)\n", opts.Speed)

	// Convert JPEG to AVIF
	fmt.Println("\nConverting to AVIF...")
	fmt.Println("(This may take a moment for large images...)")
	avifData, err := libnextimage.AVIFEncodeBytes(inputData, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding AVIF: %v\n", err)
		os.Exit(1)
	}

	// Write output AVIF file
	fmt.Printf("Writing AVIF file: %s\n", outputPath)
	if err := os.WriteFile(outputPath, avifData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	// Show results
	fmt.Printf("Output file size: %d bytes (%.2f KB)\n", len(avifData), float64(len(avifData))/1024)
	compression := (1.0 - float64(len(avifData))/float64(len(inputData))) * 100
	fmt.Printf("\n✅ Conversion successful!\n")
	fmt.Printf("Compression: %.1f%% smaller\n", compression)
	fmt.Printf("Original: %.2f KB → AVIF: %.2f KB\n",
		float64(len(inputData))/1024,
		float64(len(avifData))/1024)
	fmt.Println("\nNote: AVIF provides excellent compression but requires modern browser support.")
}
