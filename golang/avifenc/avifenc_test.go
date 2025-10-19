package avifenc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// isAVIF checks if the data starts with an AVIF file signature
func isAVIF(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	// AVIF files start with: ftyp box
	// Bytes 4-7: "ftyp"
	// Bytes 8-11: brand, typically "avif" or "avis"
	return bytes.Equal(data[4:8], []byte("ftyp")) &&
		(bytes.Equal(data[8:12], []byte("avif")) ||
			bytes.Equal(data[8:12], []byte("avis")) ||
			bytes.Equal(data[8:12], []byte("MA1A")) ||
			bytes.Equal(data[8:12], []byte("MA1B")))
}

func TestNewDefaultOptions(t *testing.T) {
	opts := NewDefaultOptions()

	// Check some key default values
	if opts.Quality != 60 {
		t.Errorf("Expected default quality 60, got %d", opts.Quality)
	}
	if opts.QualityAlpha != -1 {
		t.Errorf("Expected default quality_alpha -1 (use quality value), got %d", opts.QualityAlpha)
	}
	if opts.Speed != 6 {
		t.Errorf("Expected default speed 6, got %d", opts.Speed)
	}
	if opts.BitDepth != 8 {
		t.Errorf("Expected default bit_depth 8, got %d", opts.BitDepth)
	}
	if !opts.EnableAlpha {
		t.Error("Expected enable_alpha to be true by default")
	}
}

func TestNewCommand(t *testing.T) {
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	if cmd.cmd == nil {
		t.Error("Command should not be nil")
	}
}

func TestRunWithPNG(t *testing.T) {
	// Read test PNG file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	avifData, err := cmd.Run(pngData)
	if err != nil {
		t.Fatalf("Failed to encode PNG to AVIF: %v", err)
	}

	if len(avifData) == 0 {
		t.Error("Encoded AVIF data is empty")
	}

	// Verify AVIF signature
	if !isAVIF(avifData) {
		t.Errorf("Output does not appear to be valid AVIF (first 12 bytes: %v)", avifData[:12])
	}

	t.Logf("Successfully encoded PNG (%d bytes) to AVIF (%d bytes)", len(pngData), len(avifData))
}

func TestRunWithJPEG(t *testing.T) {
	// Read test JPEG file
	jpegPath := filepath.Join("..", "..", "testdata", "jpeg", "red.jpg")
	jpegData, err := os.ReadFile(jpegPath)
	if err != nil {
		t.Skipf("Test JPEG file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	avifData, err := cmd.Run(jpegData)
	if err != nil {
		t.Fatalf("Failed to encode JPEG to AVIF: %v", err)
	}

	if len(avifData) == 0 {
		t.Error("Encoded AVIF data is empty")
	}

	// Verify AVIF signature
	if !isAVIF(avifData) {
		t.Errorf("Output does not appear to be valid AVIF (first 12 bytes: %v)", avifData[:12])
	}

	t.Logf("Successfully encoded JPEG (%d bytes) to AVIF (%d bytes)", len(jpegData), len(avifData))
}

func TestCommandReuse(t *testing.T) {
	// Read test PNG file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// First conversion
	avifData1, err := cmd.Run(pngData)
	if err != nil {
		t.Fatalf("First conversion failed: %v", err)
	}

	// Second conversion with same command
	avifData2, err := cmd.Run(pngData)
	if err != nil {
		t.Fatalf("Second conversion failed: %v", err)
	}

	// Both should produce AVIF data
	if !isAVIF(avifData1) {
		t.Error("First conversion did not produce valid AVIF")
	}
	if !isAVIF(avifData2) {
		t.Error("Second conversion did not produce valid AVIF")
	}

	t.Logf("Command reuse successful: both conversions produced valid AVIF")
}

func TestRunFile(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	inputPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	outputPath := filepath.Join(tmpDir, "output.avif")

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	err = cmd.RunFile(inputPath, outputPath)
	if err != nil {
		t.Fatalf("RunFile failed: %v", err)
	}

	// Verify output file exists and is valid AVIF
	avifData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !isAVIF(avifData) {
		t.Error("Output file is not valid AVIF")
	}

	t.Logf("RunFile successful: created %s (%d bytes)", outputPath, len(avifData))
}

func TestRunIO(t *testing.T) {
	// Read test PNG file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Use bytes.Buffer for input and output
	input := bytes.NewReader(pngData)
	var output bytes.Buffer

	err = cmd.RunIO(input, &output)
	if err != nil {
		t.Fatalf("RunIO failed: %v", err)
	}

	avifData := output.Bytes()
	if !isAVIF(avifData) {
		t.Error("RunIO output is not valid AVIF")
	}

	t.Logf("RunIO successful: converted %d bytes to %d bytes", len(pngData), len(avifData))
}

func TestCustomOptions(t *testing.T) {
	// Read test PNG file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
	}

	// Create custom options
	opts := NewDefaultOptions()
	opts.Quality = 80
	opts.Speed = 4

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command with custom options: %v", err)
	}
	defer cmd.Close()

	avifData, err := cmd.Run(pngData)
	if err != nil {
		t.Fatalf("Failed to encode with custom options: %v", err)
	}

	if !isAVIF(avifData) {
		t.Error("Output with custom options is not valid AVIF")
	}

	t.Logf("Custom options successful: quality=80, speed=4, output=%d bytes", len(avifData))
}

func TestCloseCommand(t *testing.T) {
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Close the command
	err = cmd.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Verify command is closed
	if cmd.cmd != nil {
		t.Error("Command should be nil after Close")
	}

	// Try to use closed command - should fail gracefully
	_, err = cmd.Run([]byte("test"))
	if err == nil {
		t.Error("Expected error when using closed command")
	}
}
