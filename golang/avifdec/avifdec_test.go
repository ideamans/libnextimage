package avifdec

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// isPNG checks if the data starts with a PNG file signature
func isPNG(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	// PNG signature: 137 80 78 71 13 10 26 10
	pngSig := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	return bytes.Equal(data[:8], pngSig)
}

func TestNewDefaultOptions(t *testing.T) {
	opts := NewDefaultOptions()

	// Check some key default values
	if opts.Format != "RGBA" {
		t.Errorf("Expected default format RGBA, got %s", opts.Format)
	}
	if opts.UseThreads {
		t.Error("Expected use_threads to be false by default")
	}
	if opts.StrictFlags != 1 {
		t.Errorf("Expected default strict_flags 1, got %d", opts.StrictFlags)
	}
	if opts.ImageSizeLimit != 268435456 {
		t.Errorf("Expected default image_size_limit 268435456, got %d", opts.ImageSizeLimit)
	}
	if opts.ImageDimensionLimit != 32768 {
		t.Errorf("Expected default image_dimension_limit 32768, got %d", opts.ImageDimensionLimit)
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

func TestRunWithAVIF(t *testing.T) {
	// Read test AVIF file
	avifPath := filepath.Join("..", "..", "testdata", "avif", "red.avif")
	avifData, err := os.ReadFile(avifPath)
	if err != nil {
		t.Skipf("Test AVIF file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	pngData, err := cmd.Run(avifData)
	if err != nil {
		t.Fatalf("Failed to decode AVIF to PNG: %v", err)
	}

	if len(pngData) == 0 {
		t.Error("Decoded PNG data is empty")
	}

	// Verify PNG signature
	if !isPNG(pngData) {
		t.Errorf("Output does not appear to be valid PNG (first 8 bytes: %v)", pngData[:8])
	}

	t.Logf("Successfully decoded AVIF (%d bytes) to PNG (%d bytes)", len(avifData), len(pngData))
}

func TestCommandReuse(t *testing.T) {
	// Read test AVIF file
	avifPath := filepath.Join("..", "..", "testdata", "avif", "red.avif")
	avifData, err := os.ReadFile(avifPath)
	if err != nil {
		t.Skipf("Test AVIF file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// First conversion
	pngData1, err := cmd.Run(avifData)
	if err != nil {
		t.Fatalf("First conversion failed: %v", err)
	}

	// Second conversion with same command
	pngData2, err := cmd.Run(avifData)
	if err != nil {
		t.Fatalf("Second conversion failed: %v", err)
	}

	// Both should produce PNG data
	if !isPNG(pngData1) {
		t.Error("First conversion did not produce valid PNG")
	}
	if !isPNG(pngData2) {
		t.Error("Second conversion did not produce valid PNG")
	}

	t.Logf("Command reuse successful: both conversions produced valid PNG")
}

func TestRunFile(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	inputPath := filepath.Join("..", "..", "testdata", "avif", "red.avif")
	outputPath := filepath.Join(tmpDir, "output.png")

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	err = cmd.RunFile(inputPath, outputPath)
	if err != nil {
		t.Fatalf("RunFile failed: %v", err)
	}

	// Verify output file exists and is valid PNG
	pngData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !isPNG(pngData) {
		t.Error("Output file is not valid PNG")
	}

	t.Logf("RunFile successful: created %s (%d bytes)", outputPath, len(pngData))
}

func TestRunIO(t *testing.T) {
	// Read test AVIF file
	avifPath := filepath.Join("..", "..", "testdata", "avif", "red.avif")
	avifData, err := os.ReadFile(avifPath)
	if err != nil {
		t.Skipf("Test AVIF file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Use bytes.Buffer for input and output
	input := bytes.NewReader(avifData)
	var output bytes.Buffer

	err = cmd.RunIO(input, &output)
	if err != nil {
		t.Fatalf("RunIO failed: %v", err)
	}

	pngData := output.Bytes()
	if !isPNG(pngData) {
		t.Error("RunIO output is not valid PNG")
	}

	t.Logf("RunIO successful: converted %d bytes to %d bytes", len(avifData), len(pngData))
}

func TestCustomOptions(t *testing.T) {
	// Read test AVIF file
	avifPath := filepath.Join("..", "..", "testdata", "avif", "red.avif")
	avifData, err := os.ReadFile(avifPath)
	if err != nil {
		t.Skipf("Test AVIF file not found: %v", err)
	}

	// Create custom options
	opts := NewDefaultOptions()
	opts.Format = "RGB"
	opts.UseThreads = true

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command with custom options: %v", err)
	}
	defer cmd.Close()

	pngData, err := cmd.Run(avifData)
	if err != nil {
		t.Fatalf("Failed to decode with custom options: %v", err)
	}

	if !isPNG(pngData) {
		t.Error("Output with custom options is not valid PNG")
	}

	t.Logf("Custom options successful: format=RGB, use_threads=true, output=%d bytes", len(pngData))
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
