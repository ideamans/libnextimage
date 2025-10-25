package libnextimage

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupAVIFCompatTest ensures avifenc/avifdec are available
func setupAVIFCompatTest(t *testing.T) {
	t.Helper()

	// Check if avifenc is available
	if _, err := exec.LookPath("avifenc"); err != nil {
		t.Skip("avifenc not found in PATH - skipping AVIF compatibility tests")
	}

	// Check if avifdec is available
	if _, err := exec.LookPath("avifdec"); err != nil {
		t.Skip("avifdec not found in PATH - skipping AVIF compatibility tests")
	}
}

// runAVIFEnc runs avifenc command and returns output bytes
func runAVIFEnc(t *testing.T, inputPath string, args []string) []byte {
	t.Helper()

	// Create temporary output file
	tmpFile, err := os.CreateTemp("", "test-*.avif")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Build command
	cmdArgs := append(args, inputPath, tmpFile.Name())
	cmd := exec.Command("avifenc", cmdArgs...)

	// Run command
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("avifenc failed: %v\nOutput: %s", err, string(output))
	}

	// Read output file
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read avifenc output: %v", err)
	}

	return data
}

// runAVIFDec runs avifdec command and returns output bytes
func runAVIFDec(t *testing.T, avifData []byte, args []string) []byte {
	t.Helper()

	// Create temporary input file
	tmpInput, err := os.CreateTemp("", "test-input-*.avif")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(tmpInput.Name())

	if _, err := tmpInput.Write(avifData); err != nil {
		t.Fatalf("failed to write temp input file: %v", err)
	}
	tmpInput.Close()

	// Create temporary output file
	tmpOutput, err := os.CreateTemp("", "test-output-*.png")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(tmpOutput.Name())
	tmpOutput.Close()

	// Build command
	cmdArgs := append(args, tmpInput.Name(), tmpOutput.Name())
	cmd := exec.Command("avifdec", cmdArgs...)

	// Run command
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("avifdec failed: %v\nOutput: %s", err, string(output))
	}

	// Read output file
	data, err := os.ReadFile(tmpOutput.Name())
	if err != nil {
		t.Fatalf("failed to read avifdec output: %v", err)
	}

	return data
}

// compareAVIFOutputs compares two AVIF outputs and requires binary exact match
func compareAVIFOutputs(t *testing.T, cmd, lib []byte) bool {
	t.Helper()

	// バイナリ完全一致チェック
	if bytes.Equal(cmd, lib) {
		t.Logf("  ✓ PASSED: Binary exact match (%d bytes)", len(cmd))
		return true
	}

	// バイナリ不一致の場合は失敗
	cmdHash := sha256.Sum256(cmd)
	libHash := sha256.Sum256(lib)
	sizeDiff := len(cmd) - len(lib)
	if sizeDiff < 0 {
		sizeDiff = -sizeDiff
	}
	sizePercent := float64(sizeDiff) / float64(len(cmd)) * 100.0

	t.Logf("  avifenc: %d bytes (sha256: %x...)", len(cmd), cmdHash[:4])
	t.Logf("  library: %d bytes (sha256: %x...)", len(lib), libHash[:4])
	t.Logf("  difference: %d bytes (%.2f%%)", sizeDiff, sizePercent)
	t.Errorf("  ❌ FAILED: Binary mismatch")
	return false
}

// compareDecodeAVIFOutputs compares decoded pixel data (currently unused but kept for future use)
/*
func compareDecodeAVIFOutputs(t *testing.T, cmdPNG, libPNG []byte) bool {
	t.Helper()

	// For decoded output, we expect pixel-exact match (PNG format should be identical)
	if bytes.Equal(cmdPNG, libPNG) {
		t.Logf("  ✓ PASSED: Pixel exact match (%d bytes)", len(cmdPNG))
		return true
	}

	// Calculate hashes
	cmdHash := sha256.Sum256(cmdPNG)
	libHash := sha256.Sum256(libPNG)

	t.Logf("  avifdec: %d bytes (sha256: %x...)", len(cmdPNG), cmdHash[:4])
	t.Logf("  library: %d bytes (sha256: %x...)", len(libPNG), libHash[:4])

	// For decoding, we want exact match
	t.Errorf("  ✗ FAILED: Decoded outputs differ")
	return false
}
*/

// TestCompat_AVIF_EncodeQuality tests AVIF encoding with different quality settings
func TestCompat_AVIF_EncodeQuality(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name    string
		quality int
		args    []string
	}{
		{
			name:    "quality-30",
			quality: 30,
			args:    []string{"-q", "30"},
		},
		{
			name:    "quality-50",
			quality: 50,
			args:    []string{"-q", "50"},
		},
		{
			name:    "quality-75",
			quality: 75,
			args:    []string{"-q", "75"},
		},
		{
			name:    "quality-90",
			quality: 90,
			args:    []string{"-q", "90"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.Quality = tc.quality
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_EncodeSpeed tests AVIF encoding with different speed settings
func TestCompat_AVIF_EncodeSpeed(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name  string
		speed int
		args  []string
	}{
		{
			name:  "speed-0",
			speed: 0,
			args:  []string{"-s", "0"},
		},
		{
			name:  "speed-4",
			speed: 4,
			args:  []string{"-s", "4"},
		},
		{
			name:  "speed-6",
			speed: 6,
			args:  []string{"-s", "6"},
		},
		{
			name:  "speed-10",
			speed: 10,
			args:  []string{"-s", "10"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.Speed = tc.speed
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_EncodeDepth tests AVIF encoding with different bit depths
func TestCompat_AVIF_EncodeDepth(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name     string
		bitDepth int
		args     []string
	}{
		{
			name:     "depth-8",
			bitDepth: 8,
			args:     []string{"-d", "8"},
		},
		{
			name:     "depth-10",
			bitDepth: 10,
			args:     []string{"-d", "10"},
		},
		{
			name:     "depth-12",
			bitDepth: 12,
			args:     []string{"-d", "12"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.BitDepth = tc.bitDepth
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_EncodeYUVFormat tests AVIF encoding with different YUV formats
func TestCompat_AVIF_EncodeYUVFormat(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name      string
		yuvFormat AVIFYUVFormat
		formatStr string
		args      []string
	}{
		{
			name:      "yuv-444",
			yuvFormat: YUVFormat444,
			formatStr: "444",
			args:      []string{"-y", "444"},
		},
		{
			name:      "yuv-422",
			yuvFormat: YUVFormat422,
			formatStr: "422",
			args:      []string{"-y", "422"},
		},
		{
			name:      "yuv-420",
			yuvFormat: YUVFormat420,
			formatStr: "420",
			args:      []string{"-y", "420"},
		},
		{
			name:      "yuv-400",
			yuvFormat: YUVFormat400,
			formatStr: "400",
			args:      []string{"-y", "400"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.YUVFormat = tc.yuvFormat
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_Decode tests AVIF decoding
func TestCompat_AVIF_Decode(t *testing.T) {
	setupAVIFCompatTest(t)

	// First create an AVIF file with metadata for testing
	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75

	// Add minimal metadata for testing ignore options
	opts.ExifData = []byte{
		'E', 'x', 'i', 'f', 0, 0, // Exif header
		'I', 'I', 42, 0, 8, 0, 0, 0, // TIFF header
		0, 0, // No IFD entries
	}
	opts.XMPData = []byte(`<?xml version="1.0"?><x:xmpmeta xmlns:x="adobe:ns:meta/"></x:xmpmeta>`)

	avifData, err := encodeAVIFWithLibrary(inputPath, opts)
	if err != nil {
		t.Fatalf("failed to create test AVIF: %v", err)
	}

	testCases := []struct {
		name       string
		ignoreExif bool
		ignoreXMP  bool
		ignoreICC  bool
	}{
		{
			name:       "decode-default",
			ignoreExif: false,
			ignoreXMP:  false,
			ignoreICC:  false,
		},
		{
			name:       "decode-ignore-exif",
			ignoreExif: true,
			ignoreXMP:  false,
			ignoreICC:  false,
		},
		{
			name:       "decode-ignore-xmp",
			ignoreExif: false,
			ignoreXMP:  true,
			ignoreICC:  false,
		},
		{
			name:       "decode-ignore-icc",
			ignoreExif: false,
			ignoreXMP:  false,
			ignoreICC:  true,
		},
		{
			name:       "decode-ignore-all",
			ignoreExif: true,
			ignoreXMP:  true,
			ignoreICC:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF decoding: %s", tc.name)

			// Run library decoding with options
			decOpts := DefaultAVIFDecodeOptions()
			decOpts.IgnoreExif = tc.ignoreExif
			decOpts.IgnoreXMP = tc.ignoreXMP
			decOpts.IgnoreICC = tc.ignoreICC

			decoded, err := AVIFDecodeBytes(avifData, decOpts)
			if err != nil {
				t.Fatalf("library decoding failed: %v", err)
			}

			t.Logf("  Decoded: %dx%d, %d bytes, format: %v",
				decoded.Width, decoded.Height, len(decoded.Data), decoded.Format)

			// Verify decoding succeeded with valid dimensions
			if decoded.Width <= 0 || decoded.Height <= 0 || len(decoded.Data) == 0 {
				t.Errorf("  ✗ FAILED: Invalid decoded image dimensions or data")
			} else {
				t.Logf("  ✓ PASSED: Decoding succeeded with ignore options")
			}
		})
	}
}

// TestCompat_AVIF_RoundTrip tests encoding and decoding round trip
func TestCompat_AVIF_RoundTrip(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name    string
		quality int
		speed   int
	}{
		{
			name:    "q50-s6",
			quality: 50,
			speed:   6,
		},
		{
			name:    "q75-s6",
			quality: 75,
			speed:   6,
		},
		{
			name:    "q90-s8",
			quality: 90,
			speed:   8,
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF round trip: %s", tc.name)

			// Encode
			opts := DefaultAVIFEncodeOptions()
			opts.Quality = tc.quality
			opts.Speed = tc.speed
			avifData, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("encoding failed: %v", err)
			}

			t.Logf("  Encoded: %d bytes", len(avifData))

			// Decode
			decOpts := DefaultAVIFDecodeOptions()
			decoded, err := AVIFDecodeBytes(avifData, decOpts)
			if err != nil {
				t.Fatalf("decoding failed: %v", err)
			}

			t.Logf("  Decoded: %dx%d, %d bytes", decoded.Width, decoded.Height, len(decoded.Data))

			if decoded.Width <= 0 || decoded.Height <= 0 || len(decoded.Data) == 0 {
				t.Errorf("  ✗ FAILED: Invalid decoded image")
			} else {
				t.Logf("  ✓ PASSED: Round trip successful")
			}
		})
	}
}

// TestCompat_AVIF_QualityAlpha tests AVIF encoding with different alpha quality settings
func TestCompat_AVIF_QualityAlpha(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name         string
		quality      int
		qualityAlpha int
		args         []string
	}{
		{
			name:         "qalpha-50",
			quality:      75,
			qualityAlpha: 50,
			args:         []string{"-q", "75", "--qalpha", "50"},
		},
		{
			name:         "qalpha-100",
			quality:      75,
			qualityAlpha: 100,
			args:         []string{"-q", "75", "--qalpha", "100"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.Quality = tc.quality
			opts.QualityAlpha = tc.qualityAlpha
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_YUVRange tests AVIF encoding with different YUV range settings
func TestCompat_AVIF_YUVRange(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name     string
		yuvRange AVIFYUVRange
		rangeStr string
		args     []string
	}{
		{
			name:     "range-full",
			yuvRange: YUVRangeFull,
			rangeStr: "full",
			args:     []string{"-r", "full"},
		},
		{
			name:     "range-limited",
			yuvRange: YUVRangeLimited,
			rangeStr: "limited",
			args:     []string{"-r", "limited"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.YUVRange = tc.yuvRange
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_CICP tests AVIF encoding with different CICP/nclx settings
func TestCompat_AVIF_CICP(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name                    string
		colorPrimaries          int
		transferCharacteristics int
		matrixCoefficients      int
		args                    []string
	}{
		{
			name:                    "cicp-bt709-srgb-bt601",
			colorPrimaries:          1,
			transferCharacteristics: 13,
			matrixCoefficients:      6,
			args:                    []string{"--cicp", "1/13/6"},
		},
		{
			name:                    "cicp-bt2020-pq-bt2020nc",
			colorPrimaries:          9,
			transferCharacteristics: 16,
			matrixCoefficients:      9,
			args:                    []string{"--cicp", "9/16/9"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.ColorPrimaries = tc.colorPrimaries
			opts.TransferCharacteristics = tc.transferCharacteristics
			opts.MatrixCoefficients = tc.matrixCoefficients
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}


// TestCompat_AVIF_Metadata tests AVIF encoding with metadata (EXIF, XMP, ICC)
func TestCompat_AVIF_Metadata(t *testing.T) {
	setupAVIFCompatTest(t)

	// Create test metadata files
	testdataMetaDir := filepath.Join(testdataDir, "metadata")
	err := os.MkdirAll(testdataMetaDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create metadata directory: %v", err)
	}

	// Minimal EXIF data (TIFF header with no IFDs)
	// Format: "Exif\0\0" + TIFF header (II for little-endian, 42, offset to first IFD)
	exifData := []byte{
		'E', 'x', 'i', 'f', 0, 0, // Exif header
		'I', 'I', 42, 0, 8, 0, 0, 0, // TIFF header (little-endian)
		0, 0, // No IFD entries
	}
	exifPath := filepath.Join(testdataMetaDir, "test.exif")
	err = os.WriteFile(exifPath, exifData, 0644)
	if err != nil {
		t.Fatalf("Failed to write EXIF file: %v", err)
	}

	// Minimal XMP data
	xmpData := []byte(`<?xml version="1.0"?><x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:format>image/avif</dc:format></rdf:Description></rdf:RDF></x:xmpmeta>`)
	xmpPath := filepath.Join(testdataMetaDir, "test.xmp")
	err = os.WriteFile(xmpPath, xmpData, 0644)
	if err != nil {
		t.Fatalf("Failed to write XMP file: %v", err)
	}

	// Minimal ICC profile (sRGB-like profile header)
	// This is a minimal valid ICC profile with required tags
	iccData := []byte{
		// Profile header (128 bytes)
		0, 0, 1, 44, // Profile size (300 bytes)
		0, 0, 0, 0, // CMM type
		2, 0, 0, 0, // Profile version 2.0
		'r', 'g', 'b', ' ', // Color space (RGB)
		'X', 'Y', 'Z', ' ', // PCS (XYZ)
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // Date/time
		'a', 'c', 's', 'p', // Profile file signature
		0, 0, 0, 0, // Platform
		0, 0, 0, 0, // Flags
		0, 0, 0, 0, // Device manufacturer
		0, 0, 0, 0, // Device model
		0, 0, 0, 0, 0, 0, 0, 0, // Device attributes
		0, 0, 0, 0, // Rendering intent
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // PCS illuminant
		0, 0, 0, 0, // Creator
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // Reserved (44 bytes)
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		// Tag table (minimal - just count and desc tag)
		0, 0, 0, 1, // Tag count: 1
		'd', 'e', 's', 'c', // desc tag signature
		0, 0, 0, 132, // Tag offset
		0, 0, 0, 168, // Tag size
		// desc tag data (168 bytes of zeros for simplicity)
	}
	// Pad to 300 bytes total
	for len(iccData) < 300 {
		iccData = append(iccData, 0)
	}
	iccPath := filepath.Join(testdataMetaDir, "test.icc")
	err = os.WriteFile(iccPath, iccData, 0644)
	if err != nil {
		t.Fatalf("Failed to write ICC file: %v", err)
	}

	testCases := []struct {
		name     string
		exif     bool
		xmp      bool
		icc      bool
		args     []string
		metadata map[string][]byte
	}{
		{
			name:     "metadata-exif",
			exif:     true,
			args:     []string{"--exif", exifPath},
			metadata: map[string][]byte{"exif": exifData},
		},
		{
			name:     "metadata-xmp",
			xmp:      true,
			args:     []string{"--xmp", xmpPath},
			metadata: map[string][]byte{"xmp": xmpData},
		},
		{
			name:     "metadata-icc",
			icc:      true,
			args:     []string{"--icc", iccPath},
			metadata: map[string][]byte{"icc": iccData},
		},
		{
			name: "metadata-all",
			exif: true,
			xmp:  true,
			icc:  true,
			args: []string{
				"--exif", exifPath,
				"--xmp", xmpPath,
				"--icc", iccPath,
			},
			metadata: map[string][]byte{
				"exif": exifData,
				"xmp":  xmpData,
				"icc":  iccData,
			},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			if exifData, ok := tc.metadata["exif"]; ok {
				opts.ExifData = exifData
			}
			if xmpData, ok := tc.metadata["xmp"]; ok {
				opts.XMPData = xmpData
			}
			if iccData, ok := tc.metadata["icc"]; ok {
				opts.ICCData = iccData
			}

			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_Transformations tests AVIF encoding with transformation properties
func TestCompat_AVIF_Transformations(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name      string
		irotAngle int
		imirAxis  AVIFMirrorAxis
		args      []string
	}{
		{
			name:      "irot-90",
			irotAngle: 1,
			imirAxis:  MirrorAxisNone,
			args:      []string{"--irot", "1"},
		},
		{
			name:      "irot-180",
			irotAngle: 2,
			imirAxis:  MirrorAxisNone,
			args:      []string{"--irot", "2"},
		},
		{
			name:      "irot-270",
			irotAngle: 3,
			imirAxis:  MirrorAxisNone,
			args:      []string{"--irot", "3"},
		},
		{
			name:      "imir-vertical",
			irotAngle: 0,
			imirAxis:  MirrorAxisVertical,
			args:      []string{"--imir", "0"},
		},
		{
			name:      "imir-horizontal",
			irotAngle: 0,
			imirAxis:  MirrorAxisHorizontal,
			args:      []string{"--imir", "1"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			if tc.irotAngle > 0 {
				opts.IRotAngle = tc.irotAngle
			}
			if tc.imirAxis != MirrorAxisNone {
				opts.IMirAxis = tc.imirAxis
			}

			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_PASP tests AVIF encoding with pixel aspect ratio
func TestCompat_AVIF_PASP(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name      string
		hSpacing  int
		vSpacing  int
		args      []string
	}{
		{
			name:     "pasp-1-1",
			hSpacing: 1,
			vSpacing: 1,
			args:     []string{"--pasp", "1,1"},
		},
		{
			name:     "pasp-4-3",
			hSpacing: 4,
			vSpacing: 3,
			args:     []string{"--pasp", "4,3"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.PASP = [2]int{tc.hSpacing, tc.vSpacing}

			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_CLLI tests AVIF encoding with content light level information
func TestCompat_AVIF_CLLI(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name    string
		maxCLL  int
		maxPALL int
		args    []string
	}{
		{
			name:    "clli-1000-400",
			maxCLL:  1000,
			maxPALL: 400,
			args:    []string{"--clli", "1000,400"},
		},
		{
			name:    "clli-4000-1000",
			maxCLL:  4000,
			maxPALL: 1000,
			args:    []string{"--clli", "4000,1000"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.CLLI = [2]int{tc.maxCLL, tc.maxPALL}

			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_Tiling tests AVIF encoding with tiling settings
func TestCompat_AVIF_Tiling(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name         string
		tileRowsLog2 int
		tileColsLog2 int
		args         []string
	}{
		{
			name:         "tiling-1x1",
			tileRowsLog2: 1,
			tileColsLog2: 1,
			args:         []string{"--tilerowslog2", "1", "--tilecolslog2", "1"},
		},
		{
			name:         "tiling-2x2",
			tileRowsLog2: 2,
			tileColsLog2: 2,
			args:         []string{"--tilerowslog2", "2", "--tilecolslog2", "2"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.TileRowsLog2 = tc.tileRowsLog2
			opts.TileColsLog2 = tc.tileColsLog2

			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_Lossless tests AVIF lossless encoding
func TestCompat_AVIF_Lossless(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name                string
		quality             int
		matrixCoefficients  int
		args                []string
	}{
		{
			name:               "lossless-flag",
			quality:            100,
			matrixCoefficients: 0, // Identity matrix for true lossless
			args:               []string{"-l"},
		},
		{
			name:               "lossless-quality-100-identity",
			quality:            100,
			matrixCoefficients: 0, // Identity matrix for true lossless
			args:               []string{"-q", "100", "--cicp", "1/2/0"},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.Quality = tc.quality
			opts.MatrixCoefficients = tc.matrixCoefficients
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_PremultiplyAlpha tests AVIF encoding with premultiply alpha
func TestCompat_AVIF_PremultiplyAlpha(t *testing.T) {
	setupAVIFCompatTest(t)

	testCases := []struct {
		name             string
		premultiplyAlpha bool
		args             []string
	}{
		{
			name:             "premultiply-enabled",
			premultiplyAlpha: true,
			args:             []string{"-p"},
		},
		{
			name:             "premultiply-disabled",
			premultiplyAlpha: false,
			args:             []string{},
		},
	}

	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF encoding: %s", tc.name)

			// Run avifenc command
			cmdOutput := runAVIFEnc(t, inputPath, tc.args)

			// Run library encoding
			opts := DefaultAVIFEncodeOptions()
			opts.PremultiplyAlpha = tc.premultiplyAlpha
			libOutput, err := encodeAVIFWithLibrary(inputPath, opts)
			if err != nil {
				t.Fatalf("library encoding failed: %v", err)
			}

			// Compare outputs
			compareAVIFOutputs(t, cmdOutput, libOutput)
		})
	}
}

// TestCompat_AVIF_DecodeSecurityLimits tests AVIF decoding with security limits
func TestCompat_AVIF_DecodeSecurityLimits(t *testing.T) {
	setupAVIFCompatTest(t)

	// Create a normal AVIF file for testing
	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75
	avifData, err := encodeAVIFWithLibrary(inputPath, opts)
	if err != nil {
		t.Fatalf("failed to create test AVIF: %v", err)
	}

	testCases := []struct {
		name                string
		imageSizeLimit      uint32
		imageDimensionLimit uint32
		shouldFail          bool
	}{
		{
			name:                "size-limit-normal",
			imageSizeLimit:      268435456, // default: 16384 x 16384
			imageDimensionLimit: 32768,     // default
			shouldFail:          false,
		},
		{
			name:                "size-limit-too-small",
			imageSizeLimit:      100000, // 512x512 = 262144 > 100000
			imageDimensionLimit: 32768,
			shouldFail:          true,
		},
		{
			name:                "dimension-limit-too-small",
			imageSizeLimit:      268435456,
			imageDimensionLimit: 256, // 512 > 256
			shouldFail:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF decoding: %s", tc.name)

			decOpts := DefaultAVIFDecodeOptions()
			decOpts.ImageSizeLimit = tc.imageSizeLimit
			decOpts.ImageDimensionLimit = tc.imageDimensionLimit

			decoded, err := AVIFDecodeBytes(avifData, decOpts)

			if tc.shouldFail {
				if err == nil {
					t.Errorf("  ✗ FAILED: Expected decode to fail with limit, but it succeeded")
				} else {
					t.Logf("  ✓ PASSED: Decode correctly failed with limit: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("  ✗ FAILED: Decode failed unexpectedly: %v", err)
				} else if decoded.Width <= 0 || decoded.Height <= 0 {
					t.Errorf("  ✗ FAILED: Invalid decoded image")
				} else {
					t.Logf("  ✓ PASSED: Decode succeeded with limits")
				}
			}
		})
	}
}

// TestCompat_AVIF_DecodeStrictFlags tests AVIF decoding with strict validation flags
func TestCompat_AVIF_DecodeStrictFlags(t *testing.T) {
	setupAVIFCompatTest(t)

	// Create a normal AVIF file for testing
	inputPath := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	opts := DefaultAVIFEncodeOptions()
	opts.Quality = 75
	avifData, err := encodeAVIFWithLibrary(inputPath, opts)
	if err != nil {
		t.Fatalf("failed to create test AVIF: %v", err)
	}

	testCases := []struct {
		name        string
		strictFlags int
	}{
		{
			name:        "strict-enabled",
			strictFlags: 1, // AVIF_STRICT_ENABLED
		},
		{
			name:        "strict-disabled",
			strictFlags: 0, // AVIF_STRICT_DISABLED
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing AVIF decoding: %s", tc.name)

			decOpts := DefaultAVIFDecodeOptions()
			decOpts.StrictFlags = tc.strictFlags

			decoded, err := AVIFDecodeBytes(avifData, decOpts)
			if err != nil {
				t.Fatalf("library decoding failed: %v", err)
			}

			t.Logf("  Decoded: %dx%d, %d bytes, strict_flags=%d",
				decoded.Width, decoded.Height, len(decoded.Data), tc.strictFlags)

			if decoded.Width <= 0 || decoded.Height <= 0 || len(decoded.Data) == 0 {
				t.Errorf("  ✗ FAILED: Invalid decoded image")
			} else {
				t.Logf("  ✓ PASSED: Decoding succeeded with strict_flags=%d", tc.strictFlags)
			}
		})
	}
}

// Note: SharpYUV and TargetSize tests are not included because:
// - SharpYUV requires libsharpyuv which may not be available in all avifenc builds
// - TargetSize is implemented in avifenc as a multi-pass encoding loop, not a direct libavif API
// These features are available in the library API but cannot be easily tested against avifenc

// convertToAVIFEncOptions converts AVIFEncodeOptions to AVIFEncOptions
func convertToAVIFEncOptions(opts AVIFEncodeOptions) AVIFEncOptions {
	return AVIFEncOptions{
		Quality:                 opts.Quality,
		QualityAlpha:            opts.QualityAlpha,
		Speed:                   opts.Speed,
		MinQuantizer:            opts.MinQuantizer,
		MaxQuantizer:            opts.MaxQuantizer,
		MinQuantizerAlpha:       opts.MinQuantizerAlpha,
		MaxQuantizerAlpha:       opts.MaxQuantizerAlpha,
		BitDepth:                opts.BitDepth,
		YUVFormat:               int(opts.YUVFormat),
		YUVRange:                int(opts.YUVRange),
		EnableAlpha:             opts.EnableAlpha,
		PremultiplyAlpha:        opts.PremultiplyAlpha,
		TileRowsLog2:            opts.TileRowsLog2,
		TileColsLog2:            opts.TileColsLog2,
		ColorPrimaries:          opts.ColorPrimaries,
		TransferCharacteristics: opts.TransferCharacteristics,
		MatrixCoefficients:      opts.MatrixCoefficients,
		SharpYUV:                opts.SharpYUV,
		TargetSize:              opts.TargetSize,
		EXIFData:                opts.ExifData,
		XMPData:                 opts.XMPData,
		ICCData:                 opts.ICCData,
		IrotAngle:               opts.IRotAngle,
		ImirAxis:                int(opts.IMirAxis),
		PASP:                    opts.PASP,
		Crop:                    opts.Crop,
		CLAP:                    opts.CLAP,
		CLLIMaxCLL:              opts.CLLI[0],
		CLLIMaxPALL:             opts.CLLI[1],
		Timescale:               opts.Timescale,
		KeyframeInterval:        opts.KeyframeInterval,
	}
}

// encodeAVIFWithLibrary encodes an image file to AVIF using the command API
func encodeAVIFWithLibrary(inputPath string, opts AVIFEncodeOptions) ([]byte, error) {
	// Convert options
	avifencOpts := convertToAVIFEncOptions(opts)

	// Create command
	cmd, err := NewAVIFEncCommand(&avifencOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create avifenc command: %w", err)
	}
	defer cmd.Close()

	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	// Encode
	avifData, err := cmd.Run(data)
	if err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return avifData, nil
}
