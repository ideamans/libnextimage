package libnextimage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWebP2GIF(t *testing.T) {
	// Read PNG file
	pngData, err := os.ReadFile("../testdata/png/red.png")
	if err != nil {
		t.Fatalf("Failed to read PNG file: %v", err)
	}
	t.Logf("Read PNG file: %d bytes", len(pngData))

	// Encode to WebP first
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 90

	webpData, err := WebPEncodeBytes(pngData, opts)
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

	t.Logf("Converted to GIF: %d bytes", len(gifData))

	// Optionally save for manual inspection
	tempGIF := filepath.Join(os.TempDir(), "test_go_webp2gif.gif")
	_ = os.WriteFile(tempGIF, gifData, 0644)
	t.Logf("Saved to %s", tempGIF)
}

func TestWebP2GIF_FromFile(t *testing.T) {
	// First create a WebP file
	pngData, _ := os.ReadFile("../testdata/png/blue.png")
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 85

	webpData, _ := WebPEncodeBytes(pngData, opts)
	tempWebP := filepath.Join(os.TempDir(), "test_input.webp")
	_ = os.WriteFile(tempWebP, webpData, 0644)
	defer os.Remove(tempWebP)

	// Convert WebP file to GIF
	gifData, err := WebP2GIFConvertFile(tempWebP)
	if err != nil {
		t.Fatalf("WebP to GIF conversion from file failed: %v", err)
	}

	if len(gifData) == 0 {
		t.Fatal("GIF data is empty")
	}

	t.Logf("Converted WebP file to GIF: %d bytes", len(gifData))
}

func TestWebP2GIF_Transparency(t *testing.T) {
	// Test with a PNG that has transparency
	pngData, err := os.ReadFile("../testdata/png/red.png")
	if err != nil {
		t.Fatalf("Failed to read PNG file: %v", err)
	}

	// Encode to WebP with lossless to preserve alpha
	opts := DefaultWebPEncodeOptions()
	opts.Lossless = true

	webpData, err := WebPEncodeBytes(pngData, opts)
	if err != nil {
		t.Fatalf("WebP encode failed: %v", err)
	}

	// Convert to GIF (transparency will be quantized)
	gifData, err := WebP2GIFConvertBytes(webpData)
	if err != nil {
		t.Fatalf("WebP to GIF conversion failed: %v", err)
	}

	if len(gifData) == 0 {
		t.Fatal("GIF data is empty")
	}

	t.Logf("Converted transparent WebP to GIF: %d bytes", len(gifData))

	// Save for manual inspection
	tempGIF := filepath.Join(os.TempDir(), "test_go_webp2gif_alpha.gif")
	_ = os.WriteFile(tempGIF, gifData, 0644)
	t.Logf("Saved to %s", tempGIF)
}

// GIF to WebP conversion is now supported via the new command-based interface
func TestGIF2WebP(t *testing.T) {
	// This test verifies that GIF2WebP conversion works
	gifData, err := os.ReadFile("../testdata/gif/static.gif")
	if err != nil {
		t.Skip("GIF test file not available")
	}

	opts := DefaultWebPEncodeOptions()
	webpData, err := GIF2WebPEncodeBytes(gifData, opts)

	// GIF2WebP is now supported via the new interface
	if err != nil {
		t.Fatalf("GIF2WebP conversion failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("WebP data is empty")
	}

	t.Logf("GIF2WebP conversion successful: %d bytes GIF -> %d bytes WebP", len(gifData), len(webpData))
}
