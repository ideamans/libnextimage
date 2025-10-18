package libnextimage

import (
	"os"
	"testing"
)

func TestWebPDecodeToPNGBytes(t *testing.T) {
	// Read WebP file
	webpData, err := os.ReadFile("../testdata/webp/test.webp")
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Convert WebP to PNG (memory)
	opts := DefaultWebPDecodeOptions()
	pngData, err := WebPDecodeToPNGBytes(webpData, opts, 9)
	if err != nil {
		t.Fatalf("WebP to PNG conversion failed: %v", err)
	}

	if len(pngData) == 0 {
		t.Fatal("PNG data is empty")
	}

	t.Logf("WebP→PNG conversion: %d → %d bytes", len(webpData), len(pngData))
}

func TestWebPDecodeToJPEGBytes(t *testing.T) {
	// Read WebP file
	webpData, err := os.ReadFile("../testdata/webp/test.webp")
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Convert WebP to JPEG (memory)
	opts := DefaultWebPDecodeOptions()
	jpegData, err := WebPDecodeToJPEGBytes(webpData, opts, 90)
	if err != nil {
		t.Fatalf("WebP to JPEG conversion failed: %v", err)
	}

	if len(jpegData) == 0 {
		t.Fatal("JPEG data is empty")
	}

	t.Logf("WebP→JPEG conversion: %d → %d bytes", len(webpData), len(jpegData))
}

func TestWebPConversionRoundTrip(t *testing.T) {
	// Read JPEG file
	jpegData, err := os.ReadFile("../testdata/jpeg/test.jpg")
	if err != nil {
		t.Fatalf("Failed to read JPEG file: %v", err)
	}
	t.Logf("Read JPEG: %d bytes", len(jpegData))

	// JPEG → WebP
	encOpts := DefaultWebPEncodeOptions()
	encOpts.Quality = 85
	webpData, err := WebPEncodeBytes(jpegData, encOpts)
	if err != nil {
		t.Fatalf("JPEG→WebP failed: %v", err)
	}
	t.Logf("JPEG→WebP: %d bytes", len(webpData))

	// WebP → PNG (memory)
	decOpts := DefaultWebPDecodeOptions()
	pngData, err := WebPDecodeToPNGBytes(webpData, decOpts, 9)
	if err != nil {
		t.Fatalf("WebP→PNG failed: %v", err)
	}
	t.Logf("WebP→PNG: %d bytes", len(pngData))

	// WebP → JPEG (memory)
	jpegData2, err := WebPDecodeToJPEGBytes(webpData, decOpts, 90)
	if err != nil {
		t.Fatalf("WebP→JPEG failed: %v", err)
	}
	t.Logf("WebP→JPEG: %d bytes", len(jpegData2))

	if len(pngData) == 0 || len(jpegData2) == 0 {
		t.Fatal("Converted data is empty")
	}
}
