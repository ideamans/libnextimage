package gif2webp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// isWebP checks if the data starts with a WebP file signature
func isWebP(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	// WebP signature: "RIFF" + 4 bytes size + "WEBP"
	return bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP"))
}

func TestNewDefaultOptions(t *testing.T) {
	opts := NewDefaultOptions()

	// Check some key default values
	if opts.Quality != 75 {
		t.Errorf("Expected default quality 75, got %f", opts.Quality)
	}
	if opts.Method != 4 {
		t.Errorf("Expected default method 4, got %d", opts.Method)
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

func TestRunWithStaticGIF(t *testing.T) {
	// Read test GIF file
	gifPath := filepath.Join("..", "..", "testdata", "gif", "static.gif")
	gifData, err := os.ReadFile(gifPath)
	if err != nil {
		t.Skipf("Test GIF file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	webpData, err := cmd.Run(gifData)
	if err != nil {
		t.Fatalf("Failed to encode GIF to WebP: %v", err)
	}

	if len(webpData) == 0 {
		t.Error("Encoded WebP data is empty")
	}

	// Verify WebP signature
	if !isWebP(webpData) {
		t.Errorf("Output does not appear to be valid WebP (first 12 bytes: %v)", webpData[:12])
	}

	t.Logf("Successfully encoded static GIF (%d bytes) to WebP (%d bytes)", len(gifData), len(webpData))
}

func TestRunWithAnimatedGIF(t *testing.T) {
	// Read test animated GIF file
	gifPath := filepath.Join("..", "..", "testdata", "gif", "animated.gif")
	gifData, err := os.ReadFile(gifPath)
	if err != nil {
		t.Skipf("Test animated GIF file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	webpData, err := cmd.Run(gifData)
	if err != nil {
		t.Fatalf("Failed to encode animated GIF to WebP: %v", err)
	}

	if len(webpData) == 0 {
		t.Error("Encoded WebP data is empty")
	}

	// Verify WebP signature
	if !isWebP(webpData) {
		t.Errorf("Output does not appear to be valid WebP (first 12 bytes: %v)", webpData[:12])
	}

	t.Logf("Successfully encoded animated GIF (%d bytes) to WebP (%d bytes)", len(gifData), len(webpData))
}

func TestCommandReuse(t *testing.T) {
	// Read test GIF file
	gifPath := filepath.Join("..", "..", "testdata", "gif", "static.gif")
	gifData, err := os.ReadFile(gifPath)
	if err != nil {
		t.Skipf("Test GIF file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// First conversion
	webpData1, err := cmd.Run(gifData)
	if err != nil {
		t.Fatalf("First conversion failed: %v", err)
	}

	// Second conversion with same command
	webpData2, err := cmd.Run(gifData)
	if err != nil {
		t.Fatalf("Second conversion failed: %v", err)
	}

	// Both should produce WebP data
	if !isWebP(webpData1) {
		t.Error("First conversion did not produce valid WebP")
	}
	if !isWebP(webpData2) {
		t.Error("Second conversion did not produce valid WebP")
	}

	t.Logf("Command reuse successful: both conversions produced valid WebP")
}

func TestRunFile(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	inputPath := filepath.Join("..", "..", "testdata", "gif", "static.gif")
	outputPath := filepath.Join(tmpDir, "output.webp")

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	err = cmd.RunFile(inputPath, outputPath)
	if err != nil {
		t.Fatalf("RunFile failed: %v", err)
	}

	// Verify output file exists and is valid WebP
	webpData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !isWebP(webpData) {
		t.Error("Output file is not valid WebP")
	}

	t.Logf("RunFile successful: created %s (%d bytes)", outputPath, len(webpData))
}

func TestRunIO(t *testing.T) {
	// Read test GIF file
	gifPath := filepath.Join("..", "..", "testdata", "gif", "static.gif")
	gifData, err := os.ReadFile(gifPath)
	if err != nil {
		t.Skipf("Test GIF file not found: %v", err)
	}

	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Use bytes.Buffer for input and output
	input := bytes.NewReader(gifData)
	var output bytes.Buffer

	err = cmd.RunIO(input, &output)
	if err != nil {
		t.Fatalf("RunIO failed: %v", err)
	}

	webpData := output.Bytes()
	if !isWebP(webpData) {
		t.Error("RunIO output is not valid WebP")
	}

	t.Logf("RunIO successful: converted %d bytes to %d bytes", len(gifData), len(webpData))
}

func TestCustomOptions(t *testing.T) {
	// Read test GIF file
	gifPath := filepath.Join("..", "..", "testdata", "gif", "static.gif")
	gifData, err := os.ReadFile(gifPath)
	if err != nil {
		t.Skipf("Test GIF file not found: %v", err)
	}

	// Create custom options
	opts := NewDefaultOptions()
	opts.Quality = 80
	opts.Method = 4

	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command with custom options: %v", err)
	}
	defer cmd.Close()

	webpData, err := cmd.Run(gifData)
	if err != nil {
		t.Fatalf("Failed to encode with custom options: %v", err)
	}

	if !isWebP(webpData) {
		t.Error("Output with custom options is not valid WebP")
	}

	t.Logf("Custom options successful: quality=80, method=4, output=%d bytes", len(webpData))
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
