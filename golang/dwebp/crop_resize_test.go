package dwebp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCropFunctionality(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Create command with crop options
	opts := NewDefaultOptions()
	opts.CropX = 5
	opts.CropY = 5
	opts.CropWidth = 10
	opts.CropHeight = 10
	opts.UseCrop = true

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	pngData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to decode WebP with crop: %v", err)
	}

	if len(pngData) == 0 {
		t.Error("Cropped PNG data is empty")
	}

	// Verify PNG signature
	if !isPNG(pngData) {
		t.Error("Output is not valid PNG")
	}

	t.Logf("Successfully cropped WebP (%d bytes) to PNG (%d bytes, crop: %dx%d at %d,%d)",
		len(webpData), len(pngData), opts.CropWidth, opts.CropHeight, opts.CropX, opts.CropY)
}

func TestResizeFunctionality(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Create command with resize options
	opts := NewDefaultOptions()
	opts.ResizeWidth = 50
	opts.ResizeHeight = 50
	opts.UseResize = true

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	pngData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to decode WebP with resize: %v", err)
	}

	if len(pngData) == 0 {
		t.Error("Resized PNG data is empty")
	}

	// Verify PNG signature
	if !isPNG(pngData) {
		t.Error("Output is not valid PNG")
	}

	t.Logf("Successfully resized WebP (%d bytes) to PNG (%d bytes, resize: %dx%d)",
		len(webpData), len(pngData), opts.ResizeWidth, opts.ResizeHeight)
}

func TestCropAndResize(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Create command with both crop and resize options
	opts := NewDefaultOptions()
	opts.CropX = 2
	opts.CropY = 2
	opts.CropWidth = 16
	opts.CropHeight = 16
	opts.UseCrop = true
	opts.ResizeWidth = 32
	opts.ResizeHeight = 32
	opts.UseResize = true

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	pngData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to decode WebP with crop and resize: %v", err)
	}

	if len(pngData) == 0 {
		t.Error("Cropped and resized PNG data is empty")
	}

	// Verify PNG signature
	if !isPNG(pngData) {
		t.Error("Output is not valid PNG")
	}

	t.Logf("Successfully cropped and resized WebP (%d bytes) to PNG (%d bytes)",
		len(webpData), len(pngData))
}

func TestFlipFunctionality(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Create command with flip option
	opts := NewDefaultOptions()
	opts.Flip = true

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	pngData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to decode WebP with flip: %v", err)
	}

	if len(pngData) == 0 {
		t.Error("Flipped PNG data is empty")
	}

	// Verify PNG signature
	if !isPNG(pngData) {
		t.Error("Output is not valid PNG")
	}

	t.Logf("Successfully flipped WebP (%d bytes) to PNG (%d bytes)",
		len(webpData), len(pngData))
}

func TestCropWithJPEGOutput(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Create command with crop and JPEG output
	opts := NewDefaultOptions()
	opts.OutputFormat = OutputJPEG
	opts.JPEGQuality = 85
	opts.CropX = 4
	opts.CropY = 4
	opts.CropWidth = 12
	opts.CropHeight = 12
	opts.UseCrop = true

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	jpegData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to decode WebP with crop to JPEG: %v", err)
	}

	if len(jpegData) == 0 {
		t.Error("Cropped JPEG data is empty")
	}

	// Verify JPEG signature
	if !isJPEG(jpegData) {
		t.Error("Output is not valid JPEG")
	}

	t.Logf("Successfully cropped WebP to JPEG (%d bytes, crop: %dx%d)",
		len(jpegData), opts.CropWidth, opts.CropHeight)
}
