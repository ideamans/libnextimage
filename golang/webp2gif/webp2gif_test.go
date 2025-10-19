package webp2gif

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// isGIF checks if the data starts with a GIF file signature
func isGIF(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	// GIF signature: "GIF87a" or "GIF89a"
	return (bytes.Equal(data[0:6], []byte("GIF87a")) || bytes.Equal(data[0:6], []byte("GIF89a")))
}

func TestNewDefaultOptions(t *testing.T) {
	opts := NewDefaultOptions()

	// Just verify it doesn't panic - minimal options
	if opts.Reserved != 0 {
		t.Errorf("Expected reserved to be 0, got %d", opts.Reserved)
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

func TestRunWithWebP(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	gifData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to convert WebP to GIF: %v", err)
	}

	if len(gifData) == 0 {
		t.Error("Converted GIF data is empty")
	}

	// Verify GIF signature
	if !isGIF(gifData) {
		t.Errorf("Output does not appear to be valid GIF (first 6 bytes: %v)", gifData[:6])
	}

	t.Logf("Successfully converted WebP (%d bytes) to GIF (%d bytes)", len(webpData), len(gifData))
}

func TestCommandReuse(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// First conversion
	gifData1, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("First conversion failed: %v", err)
	}

	// Second conversion with same command
	gifData2, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Second conversion failed: %v", err)
	}

	// Both should produce GIF data
	if !isGIF(gifData1) {
		t.Error("First conversion did not produce valid GIF")
	}
	if !isGIF(gifData2) {
		t.Error("Second conversion did not produce valid GIF")
	}

	t.Logf("Command reuse successful: both conversions produced valid GIF")
}

func TestRunFile(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	inputPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	outputPath := filepath.Join(tmpDir, "output.gif")

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	err = cmd.RunFile(inputPath, outputPath)
	if err != nil {
		t.Fatalf("RunFile failed: %v", err)
	}

	// Verify output file exists and is valid GIF
	gifData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !isGIF(gifData) {
		t.Error("Output file is not valid GIF")
	}

	t.Logf("RunFile successful: created %s (%d bytes)", outputPath, len(gifData))
}

func TestRunIO(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Use bytes.Buffer for input and output
	input := bytes.NewReader(webpData)
	var output bytes.Buffer

	err = cmd.RunIO(input, &output)
	if err != nil {
		t.Fatalf("RunIO failed: %v", err)
	}

	gifData := output.Bytes()
	if !isGIF(gifData) {
		t.Error("RunIO output is not valid GIF")
	}

	t.Logf("RunIO successful: converted %d bytes to %d bytes", len(webpData), len(gifData))
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
