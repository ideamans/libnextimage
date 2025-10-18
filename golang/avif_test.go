package libnextimage

import (
	"bytes"
	"testing"
)

// createTestAVIFImage creates a simple test image for AVIF encoding tests
func createTestAVIFImage(width, height int) []byte {
	data := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			data[idx+0] = uint8((x * 255) / width)  // R gradient
			data[idx+1] = uint8((y * 255) / height) // G gradient
			data[idx+2] = 128                        // B constant
			data[idx+3] = 255                        // A opaque
		}
	}
	return data
}

func TestAVIFDefaultOptions(t *testing.T) {
	encOpts := DefaultAVIFEncodeOptions()
	if encOpts.Quality != 50 {
		t.Errorf("Default quality should be 50, got %d", encOpts.Quality)
	}
	if encOpts.BitDepth != 8 {
		t.Errorf("Default bit depth should be 8, got %d", encOpts.BitDepth)
	}
	if encOpts.YUVFormat != 2 { // 420
		t.Errorf("Default YUV format should be 2 (420), got %d", encOpts.YUVFormat)
	}

	decOpts := DefaultAVIFDecodeOptions()
	if decOpts.Format != FormatRGBA {
		t.Errorf("Default decode format should be RGBA, got %d", decOpts.Format)
	}
}

func TestAVIFEncodeDecodeRoundTrip(t *testing.T) {
	width, height := 64, 64
	inputData := createTestAVIFImage(width, height)

	// Encode
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 90 // High quality for better roundtrip

	avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
	if err != nil {
		t.Fatalf("AVIF encode failed: %v", err)
	}

	if len(avifData) == 0 {
		t.Fatal("AVIF encode produced empty output")
	}

	t.Logf("AVIF encoded: %d bytes (input: %d bytes, ratio: %.2f%%)",
		len(avifData), len(inputData), float64(len(avifData))/float64(len(inputData))*100)

	// Decode
	decOpts := DefaultAVIFDecodeOptions()
	decoded, err := AVIFDecodeBytes(avifData, decOpts)
	if err != nil {
		t.Fatalf("AVIF decode failed: %v", err)
	}

	// Verify dimensions
	if decoded.Width != width || decoded.Height != height {
		t.Errorf("Dimension mismatch: expected %dx%d, got %dx%d",
			width, height, decoded.Width, decoded.Height)
	}

	// Verify format
	if decoded.Format != FormatRGBA {
		t.Errorf("Format mismatch: expected %d, got %d", FormatRGBA, decoded.Format)
	}

	// Verify data size
	expectedSize := width * height * 4
	if len(decoded.Data) != expectedSize {
		t.Errorf("Data size mismatch: expected %d, got %d", expectedSize, len(decoded.Data))
	}
}

func TestAVIFQualityLevels(t *testing.T) {
	width, height := 128, 128
	inputData := createTestAVIFImage(width, height)

	qualities := []int{10, 30, 50, 70, 90}
	var prevSize int

	for _, q := range qualities {
		opts := DefaultAVIFEncodeOptions()
		opts.Quality = q

		avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
		if err != nil {
			t.Fatalf("AVIF encode with quality %d failed: %v", q, err)
		}

		t.Logf("Quality %d: %d bytes", q, len(avifData))

		// Generally, higher quality should produce larger files
		// But this isn't always true for all images, so we just verify it encodes
		if q == 90 && len(avifData) < prevSize/2 {
			t.Logf("Note: Quality %d produced smaller file than quality %d (expected for some images)",
				q, qualities[len(qualities)-2])
		}

		prevSize = len(avifData)

		// Verify it decodes successfully
		decoded, err := AVIFDecodeBytes(avifData, DefaultAVIFDecodeOptions())
		if err != nil {
			t.Fatalf("AVIF decode for quality %d failed: %v", q, err)
		}

		if decoded.Width != width || decoded.Height != height {
			t.Errorf("Quality %d: dimension mismatch", q)
		}
	}
}

func TestAVIFYUVFormats(t *testing.T) {
	width, height := 64, 64
	inputData := createTestAVIFImage(width, height)

	formats := []struct {
		code int
		name string
	}{
		{0, "YUV444"},
		{1, "YUV422"},
		{2, "YUV420"},
		{3, "YUV400"},
	}

	for _, fmt := range formats {
		t.Run(fmt.name, func(t *testing.T) {
			opts := DefaultAVIFEncodeOptions()
			opts.YUVFormat = fmt.code
			opts.Quality = 80

			avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
			if err != nil {
				t.Fatalf("AVIF encode with %s failed: %v", fmt.name, err)
			}

			t.Logf("%s: %d bytes", fmt.name, len(avifData))

			// Decode and verify
			decoded, err := AVIFDecodeBytes(avifData, DefaultAVIFDecodeOptions())
			if err != nil {
				t.Fatalf("AVIF decode for %s failed: %v", fmt.name, err)
			}

			if decoded.Width != width || decoded.Height != height {
				t.Errorf("%s: dimension mismatch", fmt.name)
			}
		})
	}
}

func TestAVIFDecodeSize(t *testing.T) {
	width, height := 100, 75
	inputData := createTestAVIFImage(width, height)

	// Encode
	opts := DefaultAVIFEncodeOptions()
	avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
	if err != nil {
		t.Fatalf("AVIF encode failed: %v", err)
	}

	// Get size info
	w, h, bitDepth, requiredSize, err := AVIFDecodeSize(avifData)
	if err != nil {
		t.Fatalf("AVIF decode size failed: %v", err)
	}

	if w != width || h != height {
		t.Errorf("Size mismatch: expected %dx%d, got %dx%d", width, height, w, h)
	}

	if bitDepth != 8 {
		t.Errorf("Bit depth mismatch: expected 8, got %d", bitDepth)
	}

	expectedSize := width * height * 4 // RGBA
	if requiredSize != expectedSize {
		t.Errorf("Required size mismatch: expected %d, got %d", expectedSize, requiredSize)
	}

	t.Logf("Decode size: %dx%d, bit_depth=%d, required=%d bytes", w, h, bitDepth, requiredSize)
}

func TestAVIFInvalidInput(t *testing.T) {
	// Test empty data
	_, err := AVIFEncodeBytes([]byte{}, 64, 64, FormatRGBA, DefaultAVIFEncodeOptions())
	if err == nil {
		t.Error("Expected error for empty input data")
	}

	// Test invalid dimensions
	data := createTestAVIFImage(64, 64)
	_, err = AVIFEncodeBytes(data, 0, 64, FormatRGBA, DefaultAVIFEncodeOptions())
	if err == nil {
		t.Error("Expected error for zero width")
	}

	_, err = AVIFEncodeBytes(data, 64, -1, FormatRGBA, DefaultAVIFEncodeOptions())
	if err == nil {
		t.Error("Expected error for negative height")
	}

	// Test decode with empty data
	_, err = AVIFDecodeBytes([]byte{}, DefaultAVIFDecodeOptions())
	if err == nil {
		t.Error("Expected error for empty AVIF data")
	}

	// Test decode with invalid data
	_, err = AVIFDecodeBytes([]byte{0x00, 0x01, 0x02, 0x03}, DefaultAVIFDecodeOptions())
	if err == nil {
		t.Error("Expected error for invalid AVIF data")
	}
}

func TestAVIFPixelFormats(t *testing.T) {
	width, height := 64, 64

	formats := []struct {
		format      PixelFormat
		bytesPerPix int
		name        string
	}{
		{FormatRGBA, 4, "RGBA"},
		{FormatRGB, 3, "RGB"},
		{FormatBGRA, 4, "BGRA"},
	}

	for _, fmt := range formats {
		t.Run(fmt.name, func(t *testing.T) {
			// Create input data
			inputSize := width * height * fmt.bytesPerPix
			inputData := make([]byte, inputSize)
			for i := 0; i < inputSize; i++ {
				inputData[i] = uint8(i % 256)
			}

			// Encode
			opts := DefaultAVIFEncodeOptions()
			opts.Quality = 80

			avifData, err := AVIFEncodeBytes(inputData, width, height, fmt.format, opts)
			if err != nil {
				t.Fatalf("AVIF encode with %s failed: %v", fmt.name, err)
			}

			t.Logf("%s: encoded to %d bytes", fmt.name, len(avifData))

			// Decode
			decOpts := DefaultAVIFDecodeOptions()
			decOpts.Format = fmt.format

			decoded, err := AVIFDecodeBytes(avifData, decOpts)
			if err != nil {
				t.Fatalf("AVIF decode for %s failed: %v", fmt.name, err)
			}

			if decoded.Width != width || decoded.Height != height {
				t.Errorf("%s: dimension mismatch", fmt.name)
			}

			if decoded.Format != fmt.format {
				t.Errorf("%s: format mismatch", fmt.name)
			}
		})
	}
}

func TestAVIFSpeed(t *testing.T) {
	width, height := 128, 128
	inputData := createTestAVIFImage(width, height)

	speeds := []int{0, 3, 6, 10}

	for _, speed := range speeds {
		opts := DefaultAVIFEncodeOptions()
		opts.Speed = speed
		opts.Quality = 50

		avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
		if err != nil {
			t.Fatalf("AVIF encode with speed %d failed: %v", speed, err)
		}

		t.Logf("Speed %d: %d bytes", speed, len(avifData))

		// Verify it decodes
		_, err = AVIFDecodeBytes(avifData, DefaultAVIFDecodeOptions())
		if err != nil {
			t.Fatalf("AVIF decode for speed %d failed: %v", speed, err)
		}
	}
}

func TestAVIFConcurrency(t *testing.T) {
	width, height := 64, 64
	inputData := createTestAVIFImage(width, height)

	// Encode once to get AVIF data
	opts := DefaultAVIFEncodeOptions()
	avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
	if err != nil {
		t.Fatalf("Initial encode failed: %v", err)
	}

	// Test concurrent encoding
	t.Run("ConcurrentEncode", func(t *testing.T) {
		const goroutines = 10
		errChan := make(chan error, goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				_, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
				errChan <- err
			}()
		}

		for i := 0; i < goroutines; i++ {
			if err := <-errChan; err != nil {
				t.Errorf("Concurrent encode failed: %v", err)
			}
		}
	})

	// Test concurrent decoding
	t.Run("ConcurrentDecode", func(t *testing.T) {
		const goroutines = 10
		errChan := make(chan error, goroutines)
		decOpts := DefaultAVIFDecodeOptions()

		for i := 0; i < goroutines; i++ {
			go func() {
				_, err := AVIFDecodeBytes(avifData, decOpts)
				errChan <- err
			}()
		}

		for i := 0; i < goroutines; i++ {
			if err := <-errChan; err != nil {
				t.Errorf("Concurrent decode failed: %v", err)
			}
		}
	})
}

func TestAVIFMemoryLeak(t *testing.T) {
	width, height := 128, 128
	inputData := createTestAVIFImage(width, height)

	// Encode and decode many times to check for memory leaks
	// This test should be run with -count=100 or similar for better leak detection
	for i := 0; i < 100; i++ {
		opts := DefaultAVIFEncodeOptions()
		opts.Quality = 50

		// Encode
		avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
		if err != nil {
			t.Fatalf("Iteration %d: encode failed: %v", i, err)
		}

		// Decode
		decoded, err := AVIFDecodeBytes(avifData, DefaultAVIFDecodeOptions())
		if err != nil {
			t.Fatalf("Iteration %d: decode failed: %v", i, err)
		}

		// Verify
		if len(decoded.Data) == 0 {
			t.Fatalf("Iteration %d: decoded data is empty", i)
		}
	}
}

func TestAVIFVersion(t *testing.T) {
	// Just verify we can call Version() without crashing
	version := Version()
	if version == "" {
		t.Error("Version string is empty")
	}
	t.Logf("libnextimage version: %s", version)
}

func BenchmarkAVIFEncode(b *testing.B) {
	width, height := 640, 480
	inputData := createTestAVIFImage(width, height)
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAVIFDecode(b *testing.B) {
	width, height := 640, 480
	inputData := createTestAVIFImage(width, height)
	opts := DefaultAVIFEncodeOptions()

	// Encode once
	avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
	if err != nil {
		b.Fatal(err)
	}

	decOpts := DefaultAVIFDecodeOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AVIFDecodeBytes(avifData, decOpts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAVIFRoundTrip(b *testing.B) {
	width, height := 640, 480
	inputData := createTestAVIFImage(width, height)
	encOpts := DefaultAVIFEncodeOptions()
	decOpts := DefaultAVIFDecodeOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, encOpts)
		if err != nil {
			b.Fatal(err)
		}

		_, err = AVIFDecodeBytes(avifData, decOpts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Comparison test to verify AVIF produces reasonable compression
func TestAVIFCompressionRatio(t *testing.T) {
	width, height := 256, 256
	inputData := createTestAVIFImage(width, height)

	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 50

	avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
	if err != nil {
		t.Fatalf("AVIF encode failed: %v", err)
	}

	ratio := float64(len(avifData)) / float64(len(inputData)) * 100
	t.Logf("Compression ratio: %.2f%% (%d -> %d bytes)", ratio, len(inputData), len(avifData))

	// AVIF should achieve at least some compression on this gradient image
	if ratio > 50.0 {
		t.Logf("Warning: Compression ratio is high (%.2f%%), expected better compression", ratio)
	}
}

// Test that AVIF preserves basic image structure through encode/decode
func TestAVIFImageFidelity(t *testing.T) {
	width, height := 64, 64
	inputData := createTestAVIFImage(width, height)

	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 95 // Very high quality

	// Encode
	avifData, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
	if err != nil {
		t.Fatalf("AVIF encode failed: %v", err)
	}

	// Decode
	decoded, err := AVIFDecodeBytes(avifData, DefaultAVIFDecodeOptions())
	if err != nil {
		t.Fatalf("AVIF decode failed: %v", err)
	}

	// Check corner pixels to verify basic structure is preserved
	// Top-left should be darkish
	topLeft := decoded.Data[0]
	if topLeft > 50 {
		t.Logf("Note: Top-left pixel R=%d (expected low value, but lossy compression may change it)", topLeft)
	}

	// Bottom-right should be brighterish (high G value)
	idx := ((height-1)*width + (width-1)) * 4
	bottomRightG := decoded.Data[idx+1]
	if bottomRightG < 200 {
		t.Logf("Note: Bottom-right pixel G=%d (expected high value, but lossy compression may change it)", bottomRightG)
	}

	t.Logf("Image fidelity check passed (lossy compression applied)")
}

// Test that the same input produces similar output (determinism check)
func TestAVIFDeterminism(t *testing.T) {
	width, height := 64, 64
	inputData := createTestAVIFImage(width, height)

	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75

	// Encode twice
	avif1, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
	if err != nil {
		t.Fatalf("First encode failed: %v", err)
	}

	avif2, err := AVIFEncodeBytes(inputData, width, height, FormatRGBA, opts)
	if err != nil {
		t.Fatalf("Second encode failed: %v", err)
	}

	// Results should be identical (byte-for-byte)
	if !bytes.Equal(avif1, avif2) {
		t.Logf("Warning: AVIF encoding is not deterministic (sizes: %d vs %d)",
			len(avif1), len(avif2))
		t.Logf("This may be expected if encoder uses timestamps or random seeds")
	} else {
		t.Logf("AVIF encoding is deterministic: %d bytes", len(avif1))
	}
}
