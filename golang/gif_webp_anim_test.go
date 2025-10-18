package libnextimage

import (
	"os"
	"testing"
)

// TestGIF2WebPAnimationBasic tests basic animated GIF to WebP conversion
func TestGIF2WebPAnimationBasic(t *testing.T) {
	// Read animated GIF
	gifData, err := os.ReadFile("../testdata/gif-source/animated-3frames.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF: %v", err)
	}

	// Convert to WebP with default options
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert animated GIF to WebP: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output for animated GIF")
	}

	t.Logf("Animated GIF→WebP: input %d bytes → output %d bytes", len(gifData), len(webpData))
}

// TestGIF2WebPAnimationLossless tests lossless animated GIF to WebP conversion
func TestGIF2WebPAnimationLossless(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/animated-3frames.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF: %v", err)
	}

	// Use lossless encoding
	opts := DefaultWebPEncodeOptions()
	opts.Lossless = true
	opts.Quality = 100

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert to lossless WebP: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty lossless WebP output")
	}

	t.Logf("Lossless animated GIF→WebP: input %d bytes → output %d bytes", len(gifData), len(webpData))
}

// TestGIF2WebPAnimationWithAlpha tests animated GIF with alpha channel
func TestGIF2WebPAnimationWithAlpha(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/animated-alpha.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF with alpha: %v", err)
	}

	opts := DefaultWebPEncodeOptions()
	opts.Quality = 90
	opts.AlphaQuality = 100

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert animated alpha GIF: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output for animated alpha GIF")
	}

	t.Logf("Animated alpha GIF→WebP: input %d bytes → output %d bytes", len(gifData), len(webpData))
}

// TestGIF2WebPAnimationMixed tests mixed lossy/lossless encoding
func TestGIF2WebPAnimationMixed(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/animated-3frames.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF: %v", err)
	}

	// Enable mixed mode
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.AllowMixed = true

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert with mixed mode: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output with mixed mode")
	}

	t.Logf("Mixed mode GIF→WebP: input %d bytes → output %d bytes", len(gifData), len(webpData))
}

// TestGIF2WebPAnimationMinimizeSize tests minimize size option
func TestGIF2WebPAnimationMinimizeSize(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/animated-3frames.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF: %v", err)
	}

	// Test with minimize_size disabled (faster)
	optsNormal := DefaultWebPEncodeOptions()
	optsNormal.Quality = 80
	optsNormal.MinimizeSize = false

	webpNormal, err := GIF2WebP(gifData, optsNormal)
	if err != nil {
		t.Fatalf("Failed to convert without minimize_size: %v", err)
	}

	// Test with minimize_size enabled (slower but smaller)
	optsMin := DefaultWebPEncodeOptions()
	optsMin.Quality = 80
	optsMin.MinimizeSize = true

	webpMin, err := GIF2WebP(gifData, optsMin)
	if err != nil {
		t.Fatalf("Failed to convert with minimize_size: %v", err)
	}

	t.Logf("Normal mode: %d bytes, Minimize size mode: %d bytes", len(webpNormal), len(webpMin))
	if len(webpMin) > len(webpNormal) {
		t.Logf("Warning: minimize_size produced larger output (expected smaller)")
	}
}

// TestGIF2WebPAnimationKeyFrames tests keyframe distance options
func TestGIF2WebPAnimationKeyFrames(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/animated-3frames.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF: %v", err)
	}

	// Test with custom keyframe distances
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.Kmin = 5
	opts.Kmax = 10

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert with custom keyframes: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output with custom keyframes")
	}

	t.Logf("Custom keyframes (kmin=%d, kmax=%d): output %d bytes", opts.Kmin, opts.Kmax, len(webpData))
}

// TestGIF2WebPAnimationLoopCount tests animation loop count
func TestGIF2WebPAnimationLoopCount(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/animated-3frames.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF: %v", err)
	}

	// Test infinite loop (default)
	optsInfinite := DefaultWebPEncodeOptions()
	optsInfinite.Quality = 80
	optsInfinite.AnimLoopCount = 0 // 0 = infinite

	webpInfinite, err := GIF2WebP(gifData, optsInfinite)
	if err != nil {
		t.Fatalf("Failed to convert with infinite loop: %v", err)
	}

	// Test specific loop count
	optsOnce := DefaultWebPEncodeOptions()
	optsOnce.Quality = 80
	optsOnce.AnimLoopCount = 1 // play once

	webpOnce, err := GIF2WebP(gifData, optsOnce)
	if err != nil {
		t.Fatalf("Failed to convert with loop count 1: %v", err)
	}

	t.Logf("Infinite loop: %d bytes, Loop once: %d bytes", len(webpInfinite), len(webpOnce))
}

// TestGIF2WebPAnimationLoopCompatibility tests Chrome M62 loop compatibility mode
func TestGIF2WebPAnimationLoopCompatibility(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/animated-3frames.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF: %v", err)
	}

	// Test with loop compatibility mode
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 80
	opts.LoopCompatibility = true
	opts.AnimLoopCount = 5

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert with loop compatibility: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output with loop compatibility")
	}

	t.Logf("Loop compatibility mode: output %d bytes", len(webpData))
}

// TestGIF2WebPStaticGIF tests that static GIF still works
func TestGIF2WebPStaticGIF(t *testing.T) {
	// Read static GIF
	gifData, err := os.ReadFile("../testdata/gif-source/static-64x64.gif")
	if err != nil {
		t.Fatalf("Failed to read static GIF: %v", err)
	}

	// Convert to WebP
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 85

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert static GIF: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output for static GIF")
	}

	t.Logf("Static GIF→WebP: input %d bytes → output %d bytes", len(gifData), len(webpData))
}

// TestGIF2WebPLargeAnimation tests large animated GIF
func TestGIF2WebPLargeAnimation(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/gradient.gif")
	if err != nil {
		t.Fatalf("Failed to read large GIF: %v", err)
	}

	// Use moderate quality for large file
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 75
	opts.Method = 4

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert large GIF: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output for large GIF")
	}

	compressionRatio := float64(len(gifData)) / float64(len(webpData))
	t.Logf("Large GIF→WebP: input %d bytes → output %d bytes (compression ratio: %.2fx)",
		len(gifData), len(webpData), compressionRatio)
}

// TestGIF2WebPAllOptions tests combining all animation options
func TestGIF2WebPAllOptions(t *testing.T) {
	gifData, err := os.ReadFile("../testdata/gif-source/animated-3frames.gif")
	if err != nil {
		t.Fatalf("Failed to read animated GIF: %v", err)
	}

	// Use all animation options together
	opts := DefaultWebPEncodeOptions()
	opts.Quality = 85
	opts.AllowMixed = true
	opts.MinimizeSize = true
	opts.Kmin = 3
	opts.Kmax = 5
	opts.AnimLoopCount = 0 // infinite
	opts.LoopCompatibility = false
	opts.AlphaQuality = 90

	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Failed to convert with all options: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Empty WebP output with all options")
	}

	t.Logf("All options combined: output %d bytes", len(webpData))
}
