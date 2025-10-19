package cwebp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDefaultOptions(t *testing.T) {
	opts := NewDefaultOptions()
	if opts.Quality <= 0 || opts.Quality > 100 {
		t.Errorf("Invalid default quality: %f", opts.Quality)
	}
	if opts.Method < 0 || opts.Method > 6 {
		t.Errorf("Invalid default method: %d", opts.Method)
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
	opts.Quality = 80
	opts.Method = 4

	cmd2, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command with custom options: %v", err)
	}
	defer cmd2.Close()
}

func TestRunWithPNG(t *testing.T) {
	// Read test PNG file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
		return
	}

	// Create command
	opts := NewDefaultOptions()
	opts.Quality = 80
	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert PNG to WebP
	webpData, err := cmd.Run(pngData)
	if err != nil {
		t.Fatalf("Failed to convert PNG to WebP: %v", err)
	}

	// Verify output
	if len(webpData) == 0 {
		t.Error("WebP output is empty")
	}

	// Verify WebP signature (RIFF...WEBP)
	if len(webpData) < 12 {
		t.Error("WebP output too short")
	}
	if string(webpData[0:4]) != "RIFF" {
		t.Errorf("Invalid WebP signature (RIFF): got %s", string(webpData[0:4]))
	}
	if string(webpData[8:12]) != "WEBP" {
		t.Errorf("Invalid WebP signature (WEBP): got %s", string(webpData[8:12]))
	}

	t.Logf("Successfully converted PNG (%d bytes) to WebP (%d bytes)", len(pngData), len(webpData))
}

func TestRunWithJPEG(t *testing.T) {
	// Read test JPEG file
	jpegPath := filepath.Join("..", "..", "testdata", "jpeg", "gradient.jpg")
	jpegData, err := os.ReadFile(jpegPath)
	if err != nil {
		t.Skipf("Test JPEG file not found: %v", err)
		return
	}

	// Create command
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert JPEG to WebP
	webpData, err := cmd.Run(jpegData)
	if err != nil {
		t.Fatalf("Failed to convert JPEG to WebP: %v", err)
	}

	// Verify output
	if len(webpData) == 0 {
		t.Error("WebP output is empty")
	}

	// Verify WebP signature
	if len(webpData) >= 12 {
		if string(webpData[0:4]) != "RIFF" || string(webpData[8:12]) != "WEBP" {
			t.Error("Invalid WebP signature")
		}
	}

	t.Logf("Successfully converted JPEG (%d bytes) to WebP (%d bytes)", len(jpegData), len(webpData))
}

func TestCommandReuse(t *testing.T) {
	// Create command
	opts := NewDefaultOptions()
	opts.Quality = 75
	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Read test files
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
		return
	}

	// Convert same image multiple times
	for i := 0; i < 3; i++ {
		webpData, err := cmd.Run(pngData)
		if err != nil {
			t.Fatalf("Failed to convert on iteration %d: %v", i, err)
		}
		if len(webpData) == 0 {
			t.Errorf("Empty output on iteration %d", i)
		}
	}

	t.Log("Successfully reused command for multiple conversions")
}

func TestRunFile(t *testing.T) {
	// Create temporary output file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.webp")

	// Create command
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert file
	inputPath := filepath.Join("..", "..", "testdata", "png", "red.png")
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

	// Verify WebP signature
	webpData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if len(webpData) >= 12 {
		if string(webpData[0:4]) != "RIFF" || string(webpData[8:12]) != "WEBP" {
			t.Error("Invalid WebP signature in output file")
		}
	}

	t.Logf("Successfully converted file to %s (%d bytes)", outputPath, info.Size())
}

func TestRunIO(t *testing.T) {
	// Read test file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
		return
	}

	// Create command
	cmd, err := NewCommand(nil)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert using io.Reader and io.Writer
	input := bytes.NewReader(pngData)
	var output bytes.Buffer

	err = cmd.RunIO(input, &output)
	if err != nil {
		t.Fatalf("Failed to convert using RunIO: %v", err)
	}

	// Verify output
	webpData := output.Bytes()
	if len(webpData) == 0 {
		t.Error("WebP output is empty")
	}

	// Verify WebP signature
	if len(webpData) >= 12 {
		if string(webpData[0:4]) != "RIFF" || string(webpData[8:12]) != "WEBP" {
			t.Error("Invalid WebP signature")
		}
	}

	t.Logf("Successfully converted using RunIO (%d bytes)", len(webpData))
}

func TestLosslessMode(t *testing.T) {
	// Read test PNG
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
		return
	}

	// Create command with lossless mode
	opts := NewDefaultOptions()
	opts.Lossless = true
	cmd, err := NewCommand(&opts)
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}
	defer cmd.Close()

	// Convert
	webpData, err := cmd.Run(pngData)
	if err != nil {
		t.Fatalf("Failed to convert in lossless mode: %v", err)
	}

	if len(webpData) == 0 {
		t.Error("Lossless WebP output is empty")
	}

	t.Logf("Successfully converted in lossless mode (%d bytes)", len(webpData))
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
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, _ := os.ReadFile(pngPath)
	if pngData != nil {
		_, err = cmd.Run(pngData)
		if err == nil {
			t.Error("Expected error when using closed command")
		}
	}
}

func BenchmarkCWebPConversion(b *testing.B) {
	// Read test PNG
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		b.Skipf("Test PNG file not found: %v", err)
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
		_, err := cmd.Run(pngData)
		if err != nil {
			b.Fatalf("Conversion failed: %v", err)
		}
	}
}
