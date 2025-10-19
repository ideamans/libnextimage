package dwebp

import (
	"os"
	"path/filepath"
	"testing"
)

// isJPEG checks if the data starts with a JPEG file signature
func isJPEG(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	// JPEG signature: FF D8
	return data[0] == 0xFF && data[1] == 0xD8
}

func TestJPEGOutput(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Create command with JPEG output
	opts := NewDefaultOptions()
	opts.OutputFormat = OutputJPEG
	opts.JPEGQuality = 85

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	jpegData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to decode WebP to JPEG: %v", err)
	}

	if len(jpegData) == 0 {
		t.Error("Decoded JPEG data is empty")
	}

	// Verify JPEG signature
	if !isJPEG(jpegData) {
		t.Errorf("Output does not appear to be valid JPEG (first 2 bytes: %v)", jpegData[:2])
	}

	t.Logf("Successfully decoded WebP (%d bytes) to JPEG (%d bytes, quality=%d)",
		len(webpData), len(jpegData), opts.JPEGQuality)
}

func TestJPEGQuality(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	qualities := []int{50, 75, 90}
	sizes := make(map[int]int)

	for _, quality := range qualities {
		opts := NewDefaultOptions()
		opts.OutputFormat = OutputJPEG
		opts.JPEGQuality = quality

		cmd, err := NewCommand(&opts)
		if err != nil {
			t.Fatalf("Failed to create command: %v", err)
		}

		jpegData, err := cmd.Run(webpData)
		cmd.Close()

		if err != nil {
			t.Fatalf("Failed to decode WebP to JPEG (quality=%d): %v", quality, err)
		}

		if !isJPEG(jpegData) {
			t.Errorf("Output is not valid JPEG (quality=%d)", quality)
		}

		sizes[quality] = len(jpegData)
		t.Logf("Quality %d: %d bytes", quality, len(jpegData))
	}

	// Higher quality should generally produce larger files
	if sizes[90] < sizes[50] {
		t.Logf("Note: Quality 90 (%d bytes) is smaller than quality 50 (%d bytes) - this can happen with simple images",
			sizes[90], sizes[50])
	}
}

func TestPNGvsJPEGOutput(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	// Test PNG output
	pngOpts := NewDefaultOptions()
	pngOpts.OutputFormat = OutputPNG

	pngCmd, err := NewCommand(&pngOpts)
	if err != nil {
		t.Fatalf("Failed to create PNG command: %v", err)
	}
	defer pngCmd.Close()

	pngData, err := pngCmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to decode to PNG: %v", err)
	}

	// Test JPEG output
	jpegOpts := NewDefaultOptions()
	jpegOpts.OutputFormat = OutputJPEG
	jpegOpts.JPEGQuality = 90

	jpegCmd, err := NewCommand(&jpegOpts)
	if err != nil {
		t.Fatalf("Failed to create JPEG command: %v", err)
	}
	defer jpegCmd.Close()

	jpegData, err := jpegCmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to decode to JPEG: %v", err)
	}

	// Verify signatures
	if !isPNG(pngData) {
		t.Error("PNG output is not valid PNG")
	}
	if !isJPEG(jpegData) {
		t.Error("JPEG output is not valid JPEG")
	}

	t.Logf("PNG output: %d bytes", len(pngData))
	t.Logf("JPEG output: %d bytes (quality=90)", len(jpegData))
}

func TestJPEGOutputFile(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	inputPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	outputPath := filepath.Join(tmpDir, "output.jpg")

	opts := NewDefaultOptions()
	opts.OutputFormat = OutputJPEG
	opts.JPEGQuality = 85

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	err = cmd.RunFile(inputPath, outputPath)
	if err != nil {
		t.Fatalf("RunFile failed: %v", err)
	}

	// Verify output file exists and is valid JPEG
	jpegData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !isJPEG(jpegData) {
		t.Error("Output file is not valid JPEG")
	}

	t.Logf("RunFile successful: created %s (%d bytes)", outputPath, len(jpegData))
}
