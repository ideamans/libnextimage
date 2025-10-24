package libnextimage

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWebPEncoder tests the instance-based WebPEncoder
func TestWebPEncoder(t *testing.T) {
	// Read test image
	pngData, err := os.ReadFile(filepath.Join("..", "testdata", "png", "red.png"))
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	// Create encoder with custom options
	encoder, err := NewWebPEncoder(func(opts *WebPEncodeOptions) {
		opts.Quality = 80
		opts.Method = 6
	})
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	// Encode first image
	webp1, err := encoder.Encode(pngData)
	if err != nil {
		t.Fatalf("Failed to encode first image: %v", err)
	}

	if len(webp1) == 0 {
		t.Fatal("Encoded WebP is empty")
	}

	// Verify WebP signature
	if string(webp1[0:4]) != "RIFF" {
		t.Fatal("Invalid WebP signature (RIFF)")
	}
	if string(webp1[8:12]) != "WEBP" {
		t.Fatal("Invalid WebP signature (WEBP)")
	}

	// Reuse encoder for second image
	pngData2, _ := os.ReadFile(filepath.Join("..", "testdata", "png", "blue.png"))
	webp2, err := encoder.Encode(pngData2)
	if err != nil {
		t.Fatalf("Failed to encode second image: %v", err)
	}

	if len(webp2) == 0 {
		t.Fatal("Second encoded WebP is empty")
	}

	t.Logf("✓ WebP encoder reused successfully: image1=%d bytes, image2=%d bytes", len(webp1), len(webp2))
}

// TestWebPEncoderWithDefaults tests encoder with default options
func TestWebPEncoderWithDefaults(t *testing.T) {
	pngData, err := os.ReadFile(filepath.Join("..", "testdata", "png", "red.png"))
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	// Create encoder with nil callback (use defaults)
	encoder, err := NewWebPEncoder(nil)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	webp, err := encoder.Encode(pngData)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	if len(webp) == 0 {
		t.Fatal("Encoded WebP is empty")
	}

	t.Logf("✓ WebP encoder with defaults: %d bytes", len(webp))
}

// TestWebPDecoder tests the instance-based WebPDecoder
func TestWebPDecoder(t *testing.T) {
	// First encode an image
	pngData, err := os.ReadFile(filepath.Join("..", "testdata", "png", "red.png"))
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	encoder, _ := NewWebPEncoder(func(opts *WebPEncodeOptions) {
		opts.Quality = 90
	})
	webpData, _ := encoder.Encode(pngData)
	encoder.Close()

	// Create decoder
	decoder, err := NewWebPDecoder(func(opts *WebPDecodeOptions) {
		opts.Format = FormatRGBA
		opts.UseThreads = true
	})
	if err != nil {
		t.Fatalf("Failed to create decoder: %v", err)
	}
	defer decoder.Close()

	// Decode
	decoded, err := decoder.Decode(webpData)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.Width == 0 || decoded.Height == 0 {
		t.Fatal("Decoded image has invalid dimensions")
	}

	if len(decoded.Data) == 0 {
		t.Fatal("Decoded image data is empty")
	}

	t.Logf("✓ WebP decoder: %dx%d, format=%d, %d bytes",
		decoded.Width, decoded.Height, decoded.Format, len(decoded.Data))
}

// TestAVIFEncoder tests the instance-based AVIFEncoder
func TestAVIFEncoder(t *testing.T) {
	pngData, err := os.ReadFile(filepath.Join("..", "testdata", "png", "red.png"))
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	// Create encoder with custom options
	encoder, err := NewAVIFEncoder(func(opts *AVIFEncodeOptions) {
		opts.Quality = 60
		opts.Speed = 8
		opts.BitDepth = 8
	})
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	// Encode first image
	avif1, err := encoder.Encode(pngData)
	if err != nil {
		t.Fatalf("Failed to encode first image: %v", err)
	}

	if len(avif1) == 0 {
		t.Fatal("Encoded AVIF is empty")
	}

	// Reuse encoder for second image
	pngData2, _ := os.ReadFile(filepath.Join("..", "testdata", "png", "blue.png"))
	avif2, err := encoder.Encode(pngData2)
	if err != nil {
		t.Fatalf("Failed to encode second image: %v", err)
	}

	if len(avif2) == 0 {
		t.Fatal("Second encoded AVIF is empty")
	}

	t.Logf("✓ AVIF encoder reused successfully: image1=%d bytes, image2=%d bytes", len(avif1), len(avif2))
}

// TestAVIFDecoder tests the instance-based AVIFDecoder
func TestAVIFDecoder(t *testing.T) {
	// First encode an image
	pngData, err := os.ReadFile(filepath.Join("..", "testdata", "png", "red.png"))
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	encoder, _ := NewAVIFEncoder(func(opts *AVIFEncodeOptions) {
		opts.Quality = 70
		opts.Speed = 6
	})
	avifData, _ := encoder.Encode(pngData)
	encoder.Close()

	// Create decoder
	decoder, err := NewAVIFDecoder(func(opts *AVIFDecodeOptions) {
		opts.Format = FormatRGBA
		opts.Jobs = -1
	})
	if err != nil {
		t.Fatalf("Failed to create decoder: %v", err)
	}
	defer decoder.Close()

	// Decode
	decoded, err := decoder.Decode(avifData)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.Width == 0 || decoded.Height == 0 {
		t.Fatal("Decoded image has invalid dimensions")
	}

	if len(decoded.Data) == 0 {
		t.Fatal("Decoded image data is empty")
	}

	t.Logf("✓ AVIF decoder: %dx%d, format=%d, %d bytes",
		decoded.Width, decoded.Height, decoded.Format, len(decoded.Data))
}

// TestEncoderBatchProcessing tests efficient batch processing with encoder reuse
func TestEncoderBatchProcessing(t *testing.T) {
	files := []string{"red.png", "blue.png"}

	encoder, err := NewWebPEncoder(func(opts *WebPEncodeOptions) {
		opts.Quality = 75
		opts.Method = 4
	})
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	var totalSize int
	for _, filename := range files {
		data, err := os.ReadFile(filepath.Join("..", "testdata", "png", filename))
		if err != nil {
			t.Fatalf("Failed to read %s: %v", filename, err)
		}

		webp, err := encoder.Encode(data)
		if err != nil {
			t.Fatalf("Failed to encode %s: %v", filename, err)
		}

		totalSize += len(webp)
		t.Logf("  %s: %d bytes → %d bytes", filename, len(data), len(webp))
	}

	t.Logf("✓ Batch processing complete: %d images, total %d bytes", len(files), totalSize)
}
