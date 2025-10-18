package libnextimage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAVIFDecodeToPNG(t *testing.T) {
	// Create a test AVIF file
	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75
	avifData, err := AVIFEncodeFile(inputPath, opts)
	if err != nil {
		t.Fatalf("failed to create test AVIF: %v", err)
	}

	// Create temporary output file
	outputPath := filepath.Join(os.TempDir(), "test-avif-decode.png")
	defer os.Remove(outputPath)

	// Test PNG conversion with different compression levels
	testCases := []struct {
		name              string
		compressionLevel  int
		chromaUpsampling  ChromaUpsampling
	}{
		{
			name:             "default-compression",
			compressionLevel: -1,
			chromaUpsampling: ChromaUpsamplingAutomatic,
		},
		{
			name:             "no-compression",
			compressionLevel: 0,
			chromaUpsampling: ChromaUpsamplingAutomatic,
		},
		{
			name:             "best-compression",
			compressionLevel: 9,
			chromaUpsampling: ChromaUpsamplingAutomatic,
		},
		{
			name:             "best-quality-upsampling",
			compressionLevel: -1,
			chromaUpsampling: ChromaUpsamplingBestQuality,
		},
		{
			name:             "fastest-upsampling",
			compressionLevel: -1,
			chromaUpsampling: ChromaUpsamplingFastest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decOpts := DefaultAVIFDecodeOptions()
			decOpts.ChromaUpsampling = tc.chromaUpsampling

			err := AVIFDecodeToPNG(avifData, outputPath, decOpts, tc.compressionLevel)
			if err != nil {
				t.Fatalf("AVIFDecodeToPNG failed: %v", err)
			}

			// Verify PNG file was created
			stat, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("PNG file not created: %v", err)
			}

			if stat.Size() == 0 {
				t.Fatal("PNG file is empty")
			}

			t.Logf("PNG file created: %d bytes (compression=%d, upsampling=%d)",
				stat.Size(), tc.compressionLevel, tc.chromaUpsampling)
		})
	}
}

func TestAVIFDecodeToJPEG(t *testing.T) {
	// Create a test AVIF file
	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75
	avifData, err := AVIFEncodeFile(inputPath, opts)
	if err != nil {
		t.Fatalf("failed to create test AVIF: %v", err)
	}

	// Create temporary output file
	outputPath := filepath.Join(os.TempDir(), "test-avif-decode.jpg")
	defer os.Remove(outputPath)

	// Test JPEG conversion with different quality levels
	testCases := []struct {
		name             string
		quality          int
		chromaUpsampling ChromaUpsampling
	}{
		{
			name:             "quality-50",
			quality:          50,
			chromaUpsampling: ChromaUpsamplingAutomatic,
		},
		{
			name:             "quality-75",
			quality:          75,
			chromaUpsampling: ChromaUpsamplingAutomatic,
		},
		{
			name:             "quality-90",
			quality:          90,
			chromaUpsampling: ChromaUpsamplingAutomatic,
		},
		{
			name:             "quality-90-bilinear",
			quality:          90,
			chromaUpsampling: ChromaUpsamplingBilinear,
		},
		{
			name:             "quality-90-nearest",
			quality:          90,
			chromaUpsampling: ChromaUpsamplingNearest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decOpts := DefaultAVIFDecodeOptions()
			decOpts.ChromaUpsampling = tc.chromaUpsampling

			err := AVIFDecodeToJPEG(avifData, outputPath, decOpts, tc.quality)
			if err != nil {
				t.Fatalf("AVIFDecodeToJPEG failed: %v", err)
			}

			// Verify JPEG file was created
			stat, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("JPEG file not created: %v", err)
			}

			if stat.Size() == 0 {
				t.Fatal("JPEG file is empty")
			}

			t.Logf("JPEG file created: %d bytes (quality=%d, upsampling=%d)",
				stat.Size(), tc.quality, tc.chromaUpsampling)
		})
	}
}

func TestAVIFDecodeFileToPNG(t *testing.T) {
	// Create a test AVIF file
	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75

	avifPath := filepath.Join(os.TempDir(), "test-input.avif")
	defer os.Remove(avifPath)

	avifData, err := AVIFEncodeFile(inputPath, opts)
	if err != nil {
		t.Fatalf("failed to create test AVIF: %v", err)
	}

	err = os.WriteFile(avifPath, avifData, 0644)
	if err != nil {
		t.Fatalf("failed to write AVIF file: %v", err)
	}

	// Convert AVIF file to PNG
	outputPath := filepath.Join(os.TempDir(), "test-output.png")
	defer os.Remove(outputPath)

	decOpts := DefaultAVIFDecodeOptions()
	err = AVIFDecodeFileToPNG(avifPath, outputPath, decOpts, -1)
	if err != nil {
		t.Fatalf("AVIFDecodeFileToPNG failed: %v", err)
	}

	// Verify PNG file was created
	stat, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("PNG file not created: %v", err)
	}

	if stat.Size() == 0 {
		t.Fatal("PNG file is empty")
	}

	t.Logf("PNG file created: %d bytes", stat.Size())
}

func TestAVIFDecodeFileToJPEG(t *testing.T) {
	// Create a test AVIF file
	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75

	avifPath := filepath.Join(os.TempDir(), "test-input.avif")
	defer os.Remove(avifPath)

	avifData, err := AVIFEncodeFile(inputPath, opts)
	if err != nil {
		t.Fatalf("failed to create test AVIF: %v", err)
	}

	err = os.WriteFile(avifPath, avifData, 0644)
	if err != nil {
		t.Fatalf("failed to write AVIF file: %v", err)
	}

	// Convert AVIF file to JPEG
	outputPath := filepath.Join(os.TempDir(), "test-output.jpg")
	defer os.Remove(outputPath)

	decOpts := DefaultAVIFDecodeOptions()
	err = AVIFDecodeFileToJPEG(avifPath, outputPath, decOpts, 90)
	if err != nil {
		t.Fatalf("AVIFDecodeFileToJPEG failed: %v", err)
	}

	// Verify JPEG file was created
	stat, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("JPEG file not created: %v", err)
	}

	if stat.Size() == 0 {
		t.Fatal("JPEG file is empty")
	}

	t.Logf("JPEG file created: %d bytes", stat.Size())
}
