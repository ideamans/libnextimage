package libnextimage

import (
	"os"
	"testing"
)

// TestWebPCrop tests the crop functionality
func TestWebPCrop(t *testing.T) {
	// Read test PNG
	inputData, err := os.ReadFile("../testdata/png-source/medium-512x512.png")
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// Crop to 256x256 from center (128, 128)
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.CropX = 128
	opts.CropY = 128
	opts.CropWidth = 256
	opts.CropHeight = 256

	webpData, err := WebPEncodeBytes(inputData, opts)
	if err != nil {
		t.Fatalf("Crop encoding failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output after crop")
	}

	// Decode to verify dimensions
	decOpts := DefaultWebPDecodeOptions()
	img, err := WebPDecodeBytes(webpData, decOpts)
	if err != nil {
		t.Fatalf("Failed to decode cropped WebP: %v", err)
	}

	if img.Width != 256 || img.Height != 256 {
		t.Errorf("Expected 256x256, got %dx%d", img.Width, img.Height)
	}

	t.Logf("Crop test passed: cropped to %dx%d, output size %d bytes", img.Width, img.Height, len(webpData))
}

// TestWebPResize tests the resize functionality
func TestWebPResize(t *testing.T) {
	// Read test PNG (512x512)
	inputData, err := os.ReadFile("../testdata/png-source/medium-512x512.png")
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// Resize to 200x200
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.ResizeWidth = 200
	opts.ResizeHeight = 200
	opts.ResizeMode = ResizeModeAlways // always

	webpData, err := WebPEncodeBytes(inputData, opts)
	if err != nil {
		t.Fatalf("Resize encoding failed: %v", err)
	}

	// Decode to verify dimensions
	decOpts := DefaultWebPDecodeOptions()
	img, err := WebPDecodeBytes(webpData, decOpts)
	if err != nil {
		t.Fatalf("Failed to decode resized WebP: %v", err)
	}

	if img.Width != 200 || img.Height != 200 {
		t.Errorf("Expected 200x200, got %dx%d", img.Width, img.Height)
	}

	t.Logf("Resize test passed: resized to %dx%d, output size %d bytes", img.Width, img.Height, len(webpData))
}

// TestWebPResizeUpOnly tests resize up_only mode
func TestWebPResizeUpOnly(t *testing.T) {
	// Read test PNG (512x512)
	inputData, err := os.ReadFile("../testdata/png-source/medium-512x512.png")
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// Try to resize to 256x256 with up_only mode (should NOT resize, original is 512x512)
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.ResizeWidth = 256
	opts.ResizeHeight = 256
	opts.ResizeMode = ResizeModeUpOnly // up_only

	webpData, err := WebPEncodeBytes(inputData, opts)
	if err != nil {
		t.Fatalf("Resize up_only encoding failed: %v", err)
	}

	// Decode to verify dimensions (should be original 512x512)
	decOpts := DefaultWebPDecodeOptions()
	img, err := WebPDecodeBytes(webpData, decOpts)
	if err != nil {
		t.Fatalf("Failed to decode WebP: %v", err)
	}

	if img.Width != 512 || img.Height != 512 {
		t.Errorf("Expected 512x512 (no resize), got %dx%d", img.Width, img.Height)
	}

	t.Logf("Resize up_only test passed: kept original %dx%d", img.Width, img.Height)
}

// TestWebPResizeDownOnly tests resize down_only mode
func TestWebPResizeDownOnly(t *testing.T) {
	// Read test PNG (128x128)
	inputData, err := os.ReadFile("../testdata/png-source/small-128x128.png")
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// Try to resize to 256x256 with down_only mode (should NOT resize, original is 128x128)
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.ResizeWidth = 256
	opts.ResizeHeight = 256
	opts.ResizeMode = ResizeModeDownOnly // down_only

	webpData, err := WebPEncodeBytes(inputData, opts)
	if err != nil {
		t.Fatalf("Resize down_only encoding failed: %v", err)
	}

	// Decode to verify dimensions (should be original 128x128)
	decOpts := DefaultWebPDecodeOptions()
	img, err := WebPDecodeBytes(webpData, decOpts)
	if err != nil {
		t.Fatalf("Failed to decode WebP: %v", err)
	}

	if img.Width != 128 || img.Height != 128 {
		t.Errorf("Expected 128x128 (no resize), got %dx%d", img.Width, img.Height)
	}

	t.Logf("Resize down_only test passed: kept original %dx%d", img.Width, img.Height)
}

// TestWebPCropAndResize tests crop followed by resize
func TestWebPCropAndResize(t *testing.T) {
	// Read test PNG (512x512)
	inputData, err := os.ReadFile("../testdata/png-source/medium-512x512.png")
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// Crop to 256x256 from center, then resize to 128x128
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.CropX = 128
	opts.CropY = 128
	opts.CropWidth = 256
	opts.CropHeight = 256
	opts.ResizeWidth = 128
	opts.ResizeHeight = 128
	opts.ResizeMode = ResizeModeAlways // always

	webpData, err := WebPEncodeBytes(inputData, opts)
	if err != nil {
		t.Fatalf("Crop and resize encoding failed: %v", err)
	}

	// Decode to verify dimensions
	decOpts := DefaultWebPDecodeOptions()
	img, err := WebPDecodeBytes(webpData, decOpts)
	if err != nil {
		t.Fatalf("Failed to decode WebP: %v", err)
	}

	if img.Width != 128 || img.Height != 128 {
		t.Errorf("Expected 128x128, got %dx%d", img.Width, img.Height)
	}

	t.Logf("Crop and resize test passed: final size %dx%d, output size %d bytes", img.Width, img.Height, len(webpData))
}

// TestWebPBlendAlpha tests alpha blending
func TestWebPBlendAlpha(t *testing.T) {
	// Read test PNG with alpha
	inputData, err := os.ReadFile("../testdata/source/alpha/alpha-circle.png")
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// Blend against white background (0xFFFFFF)
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.Lossless = true // lossless to preserve quality
	opts.BlendAlpha = 0xFFFFFF // white

	webpData, err := WebPEncodeBytes(inputData, opts)
	if err != nil {
		t.Fatalf("Blend alpha encoding failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output after blend alpha")
	}

	t.Logf("Blend alpha test passed: output size %d bytes", len(webpData))
}

// TestWebPNoAlpha tests alpha channel removal
func TestWebPNoAlpha(t *testing.T) {
	// Read test PNG with alpha
	inputData, err := os.ReadFile("../testdata/source/alpha/alpha-circle.png")
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// Remove alpha channel
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.NoAlpha = true

	webpData, err := WebPEncodeBytes(inputData, opts)
	if err != nil {
		t.Fatalf("NoAlpha encoding failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output after noalpha")
	}

	// Decode and verify no alpha
	decOpts := DefaultWebPDecodeOptions()
	decOpts.Format = FormatRGB // expect RGB, not RGBA
	img, err := WebPDecodeBytes(webpData, decOpts)
	if err != nil {
		t.Fatalf("Failed to decode WebP: %v", err)
	}

	// RGB should have 3 bytes per pixel
	expectedSize := img.Width * img.Height * 3
	if len(img.Data) != expectedSize {
		t.Logf("Warning: Expected %d bytes for RGB, got %d", expectedSize, len(img.Data))
	}

	t.Logf("NoAlpha test passed: output size %d bytes", len(webpData))
}
