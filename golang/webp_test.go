package libnextimage

import (
	"os"
	"testing"
)

// Generate test RGBA image data (gradient)
func generateTestImageRGBA(width, height int) []byte {
	data := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			data[idx+0] = uint8((x * 255) / width)      // R
			data[idx+1] = uint8((y * 255) / height)     // G
			data[idx+2] = 128                           // B
			data[idx+3] = 255                           // A
		}
	}
	return data
}

// Generate red RGB image
func generateRedImageRGB(width, height int) []byte {
	data := make([]byte, width*height*3)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 3
			data[idx+0] = 255 // R
			data[idx+1] = 0   // G
			data[idx+2] = 0   // B
		}
	}
	return data
}

func TestVersion(t *testing.T) {
	version := Version()
	if version == "" {
		t.Error("Version should not be empty")
	}
	t.Logf("Library version: %s", version)
}

func TestDefaultWebPOptions(t *testing.T) {
	encOpts := DefaultWebPEncodeOptions()
	if encOpts.Quality != 75.0 {
		t.Errorf("Default quality should be 75.0, got %f", encOpts.Quality)
	}
	if encOpts.Lossless != false {
		t.Error("Default lossless should be false")
	}
	if encOpts.Method != 4 {
		t.Errorf("Default method should be 4, got %d", encOpts.Method)
	}

	decOpts := DefaultWebPDecodeOptions()
	if decOpts.Format != FormatRGBA {
		t.Errorf("Default format should be RGBA, got %d", decOpts.Format)
	}
	if decOpts.UseThreads != false {
		t.Error("Default use_threads should be false")
	}
}

func TestWebPEncodeDecodeBytes_RGBA(t *testing.T) {
	width, height := 64, 64

	// Generate test image
	input := generateTestImageRGBA(width, height)
	t.Logf("Generated test image: %dx%d RGBA, %d bytes", width, height, len(input))

	// Encode
	encOpts := DefaultWebPEncodeOptions()
	encOpts.Quality = 90.0

	encoded, err := WebPEncodeBytes(input, width, height, FormatRGBA, encOpts)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}
	t.Logf("Encoded to WebP: %d bytes", len(encoded))

	// Decode
	decOpts := DefaultWebPDecodeOptions()
	decOpts.Format = FormatRGBA

	decoded, err := WebPDecodeBytes(encoded, decOpts)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify metadata
	if decoded.Width != width {
		t.Errorf("Width mismatch: expected %d, got %d", width, decoded.Width)
	}
	if decoded.Height != height {
		t.Errorf("Height mismatch: expected %d, got %d", height, decoded.Height)
	}
	if decoded.Format != FormatRGBA {
		t.Errorf("Format mismatch: expected RGBA, got %d", decoded.Format)
	}
	if decoded.BitDepth != 8 {
		t.Errorf("Bit depth mismatch: expected 8, got %d", decoded.BitDepth)
	}
	if len(decoded.Data) == 0 {
		t.Fatal("Decoded data is empty")
	}
	t.Logf("Decoded from WebP: %dx%d, %d bytes", decoded.Width, decoded.Height, len(decoded.Data))
}

func TestWebPEncodeDecodeBytes_RGB(t *testing.T) {
	width, height := 32, 32

	// Generate red image
	input := generateRedImageRGB(width, height)
	t.Logf("Generated red image: %dx%d RGB, %d bytes", width, height, len(input))

	// Encode with default options
	encoded, err := WebPEncodeBytes(input, width, height, FormatRGB, DefaultWebPEncodeOptions())
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	t.Logf("Encoded to WebP: %d bytes", len(encoded))

	// Decode as RGB
	decOpts := DefaultWebPDecodeOptions()
	decOpts.Format = FormatRGB

	decoded, err := WebPDecodeBytes(encoded, decOpts)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Format != FormatRGB {
		t.Errorf("Format mismatch: expected RGB, got %d", decoded.Format)
	}
	t.Logf("Decoded as RGB format: %dx%d", decoded.Width, decoded.Height)
}

func TestWebPDecodeSize(t *testing.T) {
	width, height := 48, 48

	// Encode test image
	input := generateTestImageRGBA(width, height)
	encoded, err := WebPEncodeBytes(input, width, height, FormatRGBA, DefaultWebPEncodeOptions())
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Get decode size
	w, h, size, err := WebPDecodeSize(encoded)
	if err != nil {
		t.Fatalf("WebPDecodeSize failed: %v", err)
	}

	if w != width {
		t.Errorf("Width mismatch: expected %d, got %d", width, w)
	}
	if h != height {
		t.Errorf("Height mismatch: expected %d, got %d", height, h)
	}

	expectedSize := width * height * 4 // RGBA
	if size != expectedSize {
		t.Errorf("Size mismatch: expected %d, got %d", expectedSize, size)
	}
	t.Logf("Decode size: %dx%d, %d bytes required", w, h, size)
}

func TestWebPDecodeInto(t *testing.T) {
	width, height := 40, 40

	// Encode test image
	input := generateTestImageRGBA(width, height)
	encoded, err := WebPEncodeBytes(input, width, height, FormatRGBA, DefaultWebPEncodeOptions())
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Prepare user buffer
	bufferSize := width * height * 4
	userBuffer := make([]byte, bufferSize)

	// Decode into buffer
	decoded, err := WebPDecodeInto(encoded, userBuffer, DefaultWebPDecodeOptions())
	if err != nil {
		t.Fatalf("WebPDecodeInto failed: %v", err)
	}

	if decoded.Width != width {
		t.Errorf("Width mismatch: expected %d, got %d", width, decoded.Width)
	}
	if decoded.Height != height {
		t.Errorf("Height mismatch: expected %d, got %d", height, decoded.Height)
	}

	// Verify data is in user buffer
	if len(decoded.Data) > len(userBuffer) {
		t.Error("Decoded data exceeds user buffer")
	}

	t.Logf("Decoded into user buffer: %dx%d", decoded.Width, decoded.Height)
}

func TestWebPLossless(t *testing.T) {
	width, height := 32, 32

	input := generateTestImageRGBA(width, height)

	// Lossless encode
	encOpts := DefaultWebPEncodeOptions()
	encOpts.Lossless = true
	encOpts.Quality = 100.0

	encoded, err := WebPEncodeBytes(input, width, height, FormatRGBA, encOpts)
	if err != nil {
		t.Fatalf("Lossless encode failed: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("Lossless encoded data is empty")
	}
	t.Logf("Lossless encode: %d bytes", len(encoded))

	// Decode
	decoded, err := WebPDecodeBytes(encoded, DefaultWebPDecodeOptions())
	if err != nil {
		t.Fatalf("Lossless decode failed: %v", err)
	}
	t.Logf("Lossless decode successful: %dx%d", decoded.Width, decoded.Height)
}

func TestWebPErrorHandling(t *testing.T) {
	// Empty input
	_, err := WebPEncodeBytes([]byte{}, 10, 10, FormatRGBA, DefaultWebPEncodeOptions())
	if err == nil {
		t.Error("Should fail with empty input")
	}
	t.Logf("Empty input error (expected): %v", err)

	// Invalid dimensions
	dummy := make([]byte, 100)
	_, err = WebPEncodeBytes(dummy, 0, 0, FormatRGBA, DefaultWebPEncodeOptions())
	if err == nil {
		t.Error("Should fail with invalid dimensions")
	}
	t.Logf("Invalid dimensions error (expected): %v", err)

	// Invalid WebP data
	invalidData := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	_, err = WebPDecodeBytes(invalidData, DefaultWebPDecodeOptions())
	if err == nil {
		t.Error("Should fail with invalid WebP data")
	}
	t.Logf("Invalid WebP data error (expected): %v", err)

	// Empty WebP data
	_, err = WebPDecodeBytes([]byte{}, DefaultWebPDecodeOptions())
	if err == nil {
		t.Error("Should fail with empty WebP data")
	}
	t.Logf("Empty WebP data error (expected): %v", err)
}

func TestWebPDecodeFile(t *testing.T) {
	// Create temporary WebP file
	width, height := 32, 32
	input := generateTestImageRGBA(width, height)
	encoded, err := WebPEncodeBytes(input, width, height, FormatRGBA, DefaultWebPEncodeOptions())
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	tmpFile := "/tmp/test_webp_decode_file.webp"
	err = os.WriteFile(tmpFile, encoded, 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Decode from file
	decoded, err := WebPDecodeFile(tmpFile, DefaultWebPDecodeOptions())
	if err != nil {
		t.Fatalf("WebPDecodeFile failed: %v", err)
	}

	if decoded.Width != width || decoded.Height != height {
		t.Errorf("Dimension mismatch: expected %dx%d, got %dx%d", width, height, decoded.Width, decoded.Height)
	}
	t.Logf("WebPDecodeFile successful: %dx%d", decoded.Width, decoded.Height)
}

func TestWebPDecodeToFile(t *testing.T) {
	// Encode test image
	width, height := 32, 32
	input := generateTestImageRGBA(width, height)
	encoded, err := WebPEncodeBytes(input, width, height, FormatRGBA, DefaultWebPEncodeOptions())
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode to file
	tmpFile := "/tmp/test_webp_decode_to_file.raw"
	err = WebPDecodeToFile(encoded, tmpFile, DefaultWebPDecodeOptions())
	if err != nil {
		t.Fatalf("WebPDecodeToFile failed: %v", err)
	}
	defer os.Remove(tmpFile)

	// Verify file was created
	stat, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Output file not created: %v", err)
	}
	if stat.Size() == 0 {
		t.Error("Output file is empty")
	}
	t.Logf("WebPDecodeToFile successful: %d bytes written", stat.Size())
}

func TestGIF2WebP_NotImplemented(t *testing.T) {
	// GIF2WebP should return error as it's not implemented yet (Phase 4)
	dummyGIF := []byte("GIF89a") // Minimal GIF header
	_, err := GIF2WebP(dummyGIF, DefaultWebPEncodeOptions())
	if err == nil {
		t.Error("GIF2WebP should return error (not yet implemented)")
	}
	t.Logf("GIF2WebP error (expected): %v", err)
}

func TestWebP2GIF_NotImplemented(t *testing.T) {
	// WebP2GIF should return error as it's not implemented yet (Phase 4)
	width, height := 32, 32
	input := generateTestImageRGBA(width, height)
	encoded, err := WebPEncodeBytes(input, width, height, FormatRGBA, DefaultWebPEncodeOptions())
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	_, err = WebP2GIF(encoded)
	if err == nil {
		t.Error("WebP2GIF should return error (not yet implemented)")
	}
	t.Logf("WebP2GIF error (expected): %v", err)
}

// Benchmark WebP encoding
func BenchmarkWebPEncode(b *testing.B) {
	width, height := 1024, 768
	input := generateTestImageRGBA(width, height)
	opts := DefaultWebPEncodeOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := WebPEncodeBytes(input, width, height, FormatRGBA, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark WebP decoding
func BenchmarkWebPDecode(b *testing.B) {
	width, height := 1024, 768
	input := generateTestImageRGBA(width, height)
	encoded, err := WebPEncodeBytes(input, width, height, FormatRGBA, DefaultWebPEncodeOptions())
	if err != nil {
		b.Fatal(err)
	}

	opts := DefaultWebPDecodeOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := WebPDecodeBytes(encoded, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
