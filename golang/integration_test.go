package libnextimage

import (
	"os"
	"testing"
)

func TestWebPEncodeFromJPEG(t *testing.T) {
	// Read JPEG file
	jpegData, err := os.ReadFile("../testdata/jpeg/gradient.jpg")
	if err != nil {
		t.Fatalf("Failed to read JPEG file: %v", err)
	}
	t.Logf("Read JPEG file: %d bytes", len(jpegData))

	// Encode to WebP
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80

	webpData, err := WebPEncodeBytes(jpegData, opts)
	if err != nil {
		t.Fatalf("WebP encode failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Encoded WebP data is empty")
	}

	t.Logf("Encoded to WebP: %d bytes", len(webpData))

	// Optionally save for manual inspection
	_ = os.WriteFile("/tmp/test_go_output.webp", webpData, 0644)
}

func TestWebPEncodeFromFile(t *testing.T) {
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 85

	webpData, err := WebPEncodeFile("../testdata/png/red.png", opts)
	if err != nil {
		t.Fatalf("WebP encode file failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Encoded WebP data is empty")
	}

	t.Logf("Encoded PNG to WebP: %d bytes", len(webpData))
}

func TestAVIFEncodeFromPNG(t *testing.T) {
	// Read PNG file
	pngData, err := os.ReadFile("../testdata/png/blue.png")
	if err != nil {
		t.Fatalf("Failed to read PNG file: %v", err)
	}
	t.Logf("Read PNG file: %d bytes", len(pngData))

	// Encode to AVIF
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 60
	opts.Speed = 8

	avifData, err := AVIFEncodeBytes(pngData, opts)
	if err != nil {
		t.Fatalf("AVIF encode failed: %v", err)
	}

	if len(avifData) == 0 {
		t.Fatal("Encoded AVIF data is empty")
	}

	t.Logf("Encoded to AVIF: %d bytes", len(avifData))

	// Optionally save for manual inspection
	_ = os.WriteFile("/tmp/test_go_output.avif", avifData, 0644)
}

func TestAVIFEncodeFromFile(t *testing.T) {
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 50
	opts.Speed = 8

	avifData, err := AVIFEncodeFile("../testdata/jpeg/test.jpg", opts)
	if err != nil {
		t.Fatalf("AVIF encode file failed: %v", err)
	}

	if len(avifData) == 0 {
		t.Fatal("Encoded AVIF data is empty")
	}

	t.Logf("Encoded JPEG to AVIF: %d bytes", len(avifData))
}

func TestWebPRoundTrip(t *testing.T) {
	// Encode JPEG to WebP
	jpegData, err := os.ReadFile("../testdata/jpeg/test.jpg")
	if err != nil {
		t.Fatalf("Failed to read JPEG file: %v", err)
	}

	encOpts := DefaultWebPEncodeOptions()
	encOpts.Quality = 90

	webpData, err := WebPEncodeBytes(jpegData, encOpts)
	if err != nil {
		t.Fatalf("WebP encode failed: %v", err)
	}
	t.Logf("Encoded to WebP: %d bytes", len(webpData))

	// Decode WebP
	decOpts := DefaultWebPDecodeOptions()
	decOpts.Format = FormatRGBA

	decoded, err := WebPDecodeBytes(webpData, decOpts)
	if err != nil {
		t.Fatalf("WebP decode failed: %v", err)
	}

	if decoded.Width == 0 || decoded.Height == 0 {
		t.Fatal("Decoded image has invalid dimensions")
	}

	if len(decoded.Data) == 0 {
		t.Fatal("Decoded image data is empty")
	}

	t.Logf("Decoded WebP: %dx%d, %d bytes", decoded.Width, decoded.Height, len(decoded.Data))
}

func TestAVIFRoundTrip(t *testing.T) {
	// Encode PNG to AVIF
	pngData, err := os.ReadFile("../testdata/png/red.png")
	if err != nil {
		t.Fatalf("Failed to read PNG file: %v", err)
	}

	encOpts := DefaultAVIFEncodeOptions()
	encOpts.Quality = 60
	encOpts.Speed = 8

	avifData, err := AVIFEncodeBytes(pngData, encOpts)
	if err != nil {
		t.Fatalf("AVIF encode failed: %v", err)
	}
	t.Logf("Encoded to AVIF: %d bytes", len(avifData))

	// Decode AVIF
	decOpts := DefaultAVIFDecodeOptions()
	decOpts.Format = FormatRGBA

	decoded, err := AVIFDecodeBytes(avifData, decOpts)
	if err != nil {
		t.Fatalf("AVIF decode failed: %v", err)
	}

	if decoded.Width == 0 || decoded.Height == 0 {
		t.Fatal("Decoded image has invalid dimensions")
	}

	if len(decoded.Data) == 0 {
		t.Fatal("Decoded image data is empty")
	}

	t.Logf("Decoded AVIF: %dx%d, %d bytes", decoded.Width, decoded.Height, len(decoded.Data))
}

func TestWebPEncoderInstance(t *testing.T) {
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 85

	encoder, err := NewWebPEncoder(opts)
	if err != nil {
		t.Fatalf("Failed to create WebP encoder: %v", err)
	}
	defer encoder.Close()

	t.Log("Created WebP encoder")

	// Encode multiple images
	testFiles := []string{
		"../testdata/jpeg/gradient.jpg",
		"../testdata/png/red.png",
		"../testdata/png/blue.png",
	}

	for _, file := range testFiles {
		data, err := encoder.EncodeFile(file)
		if err != nil {
			t.Fatalf("Failed to encode %s: %v", file, err)
		}

		if len(data) == 0 {
			t.Fatalf("Encoded data is empty for %s", file)
		}

		t.Logf("Encoded %s: %d bytes", file, len(data))
	}
}

func TestAVIFEncoderInstance(t *testing.T) {
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 50
	opts.Speed = 8

	encoder, err := NewAVIFEncoder(opts)
	if err != nil {
		t.Fatalf("Failed to create AVIF encoder: %v", err)
	}
	defer encoder.Close()

	t.Log("Created AVIF encoder")

	// Encode multiple images
	testFiles := []string{
		"../testdata/jpeg/gradient.jpg",
		"../testdata/png/red.png",
	}

	for _, file := range testFiles {
		data, err := encoder.EncodeFile(file)
		if err != nil {
			t.Fatalf("Failed to encode %s: %v", file, err)
		}

		if len(data) == 0 {
			t.Fatalf("Encoded data is empty for %s", file)
		}

		t.Logf("Encoded %s: %d bytes", file, len(data))
	}
}

func TestWebPDecoderInstance(t *testing.T) {
	// First encode to get WebP data
	jpegData, _ := os.ReadFile("../testdata/jpeg/test.jpg")
	encOpts := DefaultWebPEncodeOptions()
	webpData, _ := WebPEncodeBytes(jpegData, encOpts)

	// Create decoder
	decOpts := DefaultWebPDecodeOptions()
	decOpts.Format = FormatRGBA

	decoder, err := NewWebPDecoder(decOpts)
	if err != nil {
		t.Fatalf("Failed to create WebP decoder: %v", err)
	}
	defer decoder.Close()

	t.Log("Created WebP decoder")

	// Decode
	decoded, err := decoder.Decode(webpData)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.Width == 0 || decoded.Height == 0 {
		t.Fatal("Decoded image has invalid dimensions")
	}

	t.Logf("Decoded WebP: %dx%d, %d bytes", decoded.Width, decoded.Height, len(decoded.Data))
}

func TestAVIFDecoderInstance(t *testing.T) {
	// First encode to get AVIF data
	pngData, _ := os.ReadFile("../testdata/png/red.png")
	encOpts := DefaultAVIFEncodeOptions()
	encOpts.Speed = 8
	avifData, _ := AVIFEncodeBytes(pngData, encOpts)

	// Create decoder
	decOpts := DefaultAVIFDecodeOptions()
	decOpts.Format = FormatRGBA

	decoder, err := NewAVIFDecoder(decOpts)
	if err != nil {
		t.Fatalf("Failed to create AVIF decoder: %v", err)
	}
	defer decoder.Close()

	t.Log("Created AVIF decoder")

	// Decode
	decoded, err := decoder.Decode(avifData)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.Width == 0 || decoded.Height == 0 {
		t.Fatal("Decoded image has invalid dimensions")
	}

	t.Logf("Decoded AVIF: %dx%d, %d bytes", decoded.Width, decoded.Height, len(decoded.Data))
}

func TestWebP2GIFIntegration(t *testing.T) {
	// Encode JPEG to WebP
	jpegData, err := os.ReadFile("../testdata/jpeg/test.jpg")
	if err != nil {
		t.Fatalf("Failed to read JPEG file: %v", err)
	}

	encOpts := DefaultWebPEncodeOptions()
	encOpts.Quality = 85

	webpData, err := WebPEncodeBytes(jpegData, encOpts)
	if err != nil {
		t.Fatalf("WebP encode failed: %v", err)
	}
	t.Logf("Encoded to WebP: %d bytes", len(webpData))

	// Convert WebP to GIF
	gifData, err := WebP2GIFConvertBytes(webpData)
	if err != nil {
		t.Fatalf("WebP to GIF conversion failed: %v", err)
	}

	if len(gifData) == 0 {
		t.Fatal("GIF data is empty")
	}

	t.Logf("Converted to GIF: %d bytes (256-color quantized)", len(gifData))

	// Optionally save for manual inspection
	_ = os.WriteFile("/tmp/test_integration_webp2gif.gif", gifData, 0644)
}
