package libnextimage

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
)

// AVIFDecodeToPNGBytes decodes AVIF data and returns PNG data as []byte
// pngCompressionLevel: 0-9 (0=no compression, 9=best compression, -1=default)
func AVIFDecodeToPNGBytes(avifData []byte, options AVIFDecodeOptions, pngCompressionLevel int) ([]byte, error) {
	// Decode AVIF to pixel data
	decoded, err := AVIFDecodeBytes(avifData, options)
	if err != nil {
		return nil, fmt.Errorf("avif decode to png bytes: %w", err)
	}

	// Convert to Go image.Image
	img, err := decodedImageToGoImage(decoded)
	if err != nil {
		return nil, fmt.Errorf("avif decode to png bytes: %w", err)
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
		return nil, fmt.Errorf("avif decode to png bytes: failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// AVIFDecodeToPNG decodes AVIF data and saves it as a PNG file
// pngCompressionLevel: 0-9 (0=no compression, 9=best compression, -1=default)
func AVIFDecodeToPNG(avifData []byte, outputPath string, options AVIFDecodeOptions, pngCompressionLevel int) error {
	pngData, err := AVIFDecodeToPNGBytes(avifData, options, pngCompressionLevel)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, pngData, 0644); err != nil {
		return fmt.Errorf("avif decode to png: failed to write file: %w", err)
	}

	return nil
}

// AVIFDecodeFileToPNG decodes an AVIF file and saves it as a PNG file
func AVIFDecodeFileToPNG(avifPath, pngPath string, options AVIFDecodeOptions, pngCompressionLevel int) error {
	data, err := os.ReadFile(avifPath)
	if err != nil {
		return fmt.Errorf("avif decode file to png: %w", err)
	}
	return AVIFDecodeToPNG(data, pngPath, options, pngCompressionLevel)
}

// AVIFDecodeToJPEGBytes decodes AVIF data and returns JPEG data as []byte
// jpegQuality: 1-100 (higher is better quality)
func AVIFDecodeToJPEGBytes(avifData []byte, options AVIFDecodeOptions, jpegQuality int) ([]byte, error) {
	// Decode AVIF to pixel data
	decoded, err := AVIFDecodeBytes(avifData, options)
	if err != nil {
		return nil, fmt.Errorf("avif decode to jpeg bytes: %w", err)
	}

	// Convert to Go image.Image
	img, err := decodedImageToGoImage(decoded)
	if err != nil {
		return nil, fmt.Errorf("avif decode to jpeg bytes: %w", err)
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
		return nil, fmt.Errorf("avif decode to jpeg bytes: failed to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

// AVIFDecodeToJPEG decodes AVIF data and saves it as a JPEG file
// jpegQuality: 1-100 (higher is better quality)
func AVIFDecodeToJPEG(avifData []byte, outputPath string, options AVIFDecodeOptions, jpegQuality int) error {
	jpegData, err := AVIFDecodeToJPEGBytes(avifData, options, jpegQuality)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, jpegData, 0644); err != nil {
		return fmt.Errorf("avif decode to jpeg: failed to write file: %w", err)
	}

	return nil
}

// AVIFDecodeFileToJPEG decodes an AVIF file and saves it as a JPEG file
func AVIFDecodeFileToJPEG(avifPath, jpegPath string, options AVIFDecodeOptions, jpegQuality int) error {
	data, err := os.ReadFile(avifPath)
	if err != nil {
		return fmt.Errorf("avif decode file to jpeg: %w", err)
	}
	return AVIFDecodeToJPEG(data, jpegPath, options, jpegQuality)
}

// decodedImageToGoImage converts DecodedImage to Go's image.Image
func decodedImageToGoImage(decoded *DecodedImage) (image.Image, error) {
	switch decoded.Format {
	case FormatRGBA:
		// RGBA format
		img := image.NewRGBA(image.Rect(0, 0, decoded.Width, decoded.Height))
		copy(img.Pix, decoded.Data)
		return img, nil

	case FormatRGB:
		// RGB format - convert to RGBA
		rgba := image.NewRGBA(image.Rect(0, 0, decoded.Width, decoded.Height))
		srcIdx := 0
		dstIdx := 0
		for y := 0; y < decoded.Height; y++ {
			for x := 0; x < decoded.Width; x++ {
				rgba.Pix[dstIdx+0] = decoded.Data[srcIdx+0] // R
				rgba.Pix[dstIdx+1] = decoded.Data[srcIdx+1] // G
				rgba.Pix[dstIdx+2] = decoded.Data[srcIdx+2] // B
				rgba.Pix[dstIdx+3] = 255                    // A (opaque)
				srcIdx += 3
				dstIdx += 4
			}
		}
		return rgba, nil

	case FormatBGRA:
		// BGRA format - convert to RGBA
		rgba := image.NewRGBA(image.Rect(0, 0, decoded.Width, decoded.Height))
		for i := 0; i < len(decoded.Data); i += 4 {
			rgba.Pix[i+0] = decoded.Data[i+2] // R (was B)
			rgba.Pix[i+1] = decoded.Data[i+1] // G
			rgba.Pix[i+2] = decoded.Data[i+0] // B (was R)
			rgba.Pix[i+3] = decoded.Data[i+3] // A
		}
		return rgba, nil

	default:
		return nil, fmt.Errorf("unsupported pixel format: %d", decoded.Format)
	}
}
