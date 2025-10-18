package libnextimage

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
)

// WebPDecodeToPNGBytes decodes WebP data and returns PNG data as []byte
// pngCompressionLevel: 0-9 (0=no compression, 9=best compression, -1=default)
func WebPDecodeToPNGBytes(webpData []byte, options WebPDecodeOptions, pngCompressionLevel int) ([]byte, error) {
	// Decode WebP to pixel data
	decoded, err := WebPDecodeBytes(webpData, options)
	if err != nil {
		return nil, fmt.Errorf("webp decode to png bytes: %w", err)
	}

	// Convert to Go image.Image
	img, err := decodedImageToGoImage(decoded)
	if err != nil {
		return nil, fmt.Errorf("webp decode to png bytes: %w", err)
	}

	// Create buffer for PNG data
	var buf bytes.Buffer

	// Set PNG encoder options
	encoder := &png.Encoder{}
	if pngCompressionLevel >= 0 && pngCompressionLevel <= 9 {
		encoder.CompressionLevel = png.CompressionLevel(pngCompressionLevel)
	} else {
		encoder.CompressionLevel = png.DefaultCompression
	}

	// Encode to PNG
	if err := encoder.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("webp decode to png bytes: failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// WebPDecodeToPNG decodes WebP data and saves it as a PNG file
// pngCompressionLevel: 0-9 (0=no compression, 9=best compression, -1=default)
func WebPDecodeToPNG(webpData []byte, outputPath string, options WebPDecodeOptions, pngCompressionLevel int) error {
	pngData, err := WebPDecodeToPNGBytes(webpData, options, pngCompressionLevel)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, pngData, 0644); err != nil {
		return fmt.Errorf("webp decode to png: failed to write file: %w", err)
	}

	return nil
}

// WebPDecodeFileToPNG decodes a WebP file and saves it as a PNG file
func WebPDecodeFileToPNG(webpPath, pngPath string, options WebPDecodeOptions, pngCompressionLevel int) error {
	data, err := os.ReadFile(webpPath)
	if err != nil {
		return fmt.Errorf("webp decode file to png: %w", err)
	}
	return WebPDecodeToPNG(data, pngPath, options, pngCompressionLevel)
}

// WebPDecodeToJPEGBytes decodes WebP data and returns JPEG data as []byte
// jpegQuality: 1-100 (higher is better quality)
func WebPDecodeToJPEGBytes(webpData []byte, options WebPDecodeOptions, jpegQuality int) ([]byte, error) {
	// Decode WebP to pixel data
	decoded, err := WebPDecodeBytes(webpData, options)
	if err != nil {
		return nil, fmt.Errorf("webp decode to jpeg bytes: %w", err)
	}

	// Convert to Go image.Image
	img, err := decodedImageToGoImage(decoded)
	if err != nil {
		return nil, fmt.Errorf("webp decode to jpeg bytes: %w", err)
	}

	// Set JPEG quality
	quality := jpegQuality
	if quality < 1 {
		quality = 1
	} else if quality > 100 {
		quality = 100
	}

	// Create buffer for JPEG data
	var buf bytes.Buffer

	// Encode to JPEG
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("webp decode to jpeg bytes: failed to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

// WebPDecodeToJPEG decodes WebP data and saves it as a JPEG file
// jpegQuality: 1-100 (higher is better quality)
func WebPDecodeToJPEG(webpData []byte, outputPath string, options WebPDecodeOptions, jpegQuality int) error {
	jpegData, err := WebPDecodeToJPEGBytes(webpData, options, jpegQuality)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, jpegData, 0644); err != nil {
		return fmt.Errorf("webp decode to jpeg: failed to write file: %w", err)
	}

	return nil
}

// WebPDecodeFileToJPEG decodes a WebP file and saves it as a JPEG file
func WebPDecodeFileToJPEG(webpPath, jpegPath string, options WebPDecodeOptions, jpegQuality int) error {
	data, err := os.ReadFile(webpPath)
	if err != nil {
		return fmt.Errorf("webp decode file to jpeg: %w", err)
	}
	return WebPDecodeToJPEG(data, jpegPath, options, jpegQuality)
}
