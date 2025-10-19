package dwebp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Helper to check PNG signature
func isPNG(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	pngSig := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	return bytes.Equal(data[0:8], pngSig)
}

func TestNewDefaultOptions(t *testing.T) {
	opts := NewDefaultOptions()
	if opts.Format == "" {
		t.Error("Default format is empty")
	}
}

func TestNewCommand(t *testing.T) {
	// Test with nil options (use defaults)
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command with nil options: %v", err)
	}
	defer cmd.Close()

	// Test with custom options
	opts := NewDefaultOptions()
	opts.Format = "RGB"

	cmd2, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command with custom options: %v", err)
	}
	defer cmd2.Close()
}

func TestRunWithWebP(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
		return
	}

	// Create command
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert WebP to PNG
	pngData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Failed to convert WebP to PNG: %v", err)
	}

	// Verify output
	if len(pngData) == 0 {
		t.Error("PNG output is empty")
	}

	// Verify PNG signature
	if !isPNG(pngData) {
		t.Error("Output is not a valid PNG file")
	}

	t.Logf("Successfully converted WebP (%d bytes) to PNG (%d bytes)", len(webpData), len(pngData))
}

func TestCommandReuse(t *testing.T) {
	// Read test WebP file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
		return
	}

	// Create command
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert same image multiple times
	for i := 0; i < 3; i++ {
		pngData, err := cmd.Run(webpData)
		if err != nil {
			t.Fatalf("Failed to convert on iteration %d: %v", i, err)
		}
		if len(pngData) == 0 {
			t.Errorf("Empty output on iteration %d", i)
		}
		if !isPNG(pngData) {
			t.Errorf("Invalid PNG on iteration %d", i)
		}
	}

	t.Log("Successfully reused command for multiple conversions")
}

func TestRunFile(t *testing.T) {
	// Create temporary output file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.png")

	// Create command
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert file
	inputPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	err = cmd.RunFile(inputPath, outputPath)
	if err != nil {
		t.Fatalf("Failed to convert file: %v", err)
	}

	// Verify output file exists
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Verify PNG signature
	pngData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if !isPNG(pngData) {
		t.Error("Output file is not a valid PNG")
	}

	t.Logf("Successfully converted file to %s (%d bytes)", outputPath, info.Size())
}

func TestRunIO(t *testing.T) {
	// Read test file
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		t.Skipf("Test WebP file not found: %v", err)
		return
	}

	// Create command
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert using io.Reader and io.Writer
	input := bytes.NewReader(webpData)
	var output bytes.Buffer

	err = cmd.RunIO(input, &output)
	if err != nil {
		t.Fatalf("Failed to convert using RunIO: %v", err)
	}

	// Verify output
	pngData := output.Bytes()
	if len(pngData) == 0 {
		t.Error("PNG output is empty")
	}

	// Verify PNG signature
	if !isPNG(pngData) {
		t.Error("Output is not a valid PNG")
	}

	t.Logf("Successfully converted using RunIO (%d bytes)", len(pngData))
}

func TestCloseCommand(t *testing.T) {
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Close the command
	err = cmd.Close()
	if err != nil {
		t.Errorf("Failed to close command: %v", err)
	}

	// Verify command is closed
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, _ := os.ReadFile(webpPath)
	if webpData != nil {
		_, err = cmd.Run(webpData)
		if err == nil {
			t.Error("Expected error when using closed command")
		}
	}
}

func BenchmarkDWebPConversion(b *testing.B) {
	// Read test WebP
	webpPath := filepath.Join("..", "..", "testdata", "webp", "gradient.webp")
	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		b.Skipf("Test WebP file not found: %v", err)
		return
	}

	// Create command
	cmd, err := NewCommand(nil)
	if err != nil {
		b.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cmd.Run(webpData)
		if err != nil {
			b.Fatalf("Conversion failed: %v", err)
		}
	}
}
