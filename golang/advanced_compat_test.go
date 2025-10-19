package libnextimage

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCompat_WebP_Preset tests the -preset option
func TestCompat_WebP_Preset(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name   string
		preset WebPPreset
		args   []string
	}{
		{
			name:   "preset-default",
			preset: PresetDefault,
			args:   []string{"-preset", "default"},
		},
		{
			name:   "preset-picture",
			preset: PresetPicture,
			args:   []string{"-preset", "picture"},
		},
		{
			name:   "preset-photo",
			preset: PresetPhoto,
			args:   []string{"-preset", "photo"},
		},
		{
			name:   "preset-drawing",
			preset: PresetDrawing,
			args:   []string{"-preset", "drawing"},
		},
		{
			name:   "preset-icon",
			preset: PresetIcon,
			args:   []string{"-preset", "icon"},
		},
		{
			name:   "preset-text",
			preset: PresetText,
			args:   []string{"-preset", "text"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
			outputFile := filepath.Join(tempDir, "cwebp-output.webp")

			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, inputFile, tc.args, outputFile)

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.Preset = tc.preset
			libOutput := webpEncodeWithLibrary(t, inputFile, opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_ImageHint tests the -hint option (for lossless encoding)
func TestCompat_WebP_ImageHint(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name string
		hint WebPImageHint
		args []string
	}{
		{
			name: "hint-picture",
			hint: HintPicture,
			args: []string{"-lossless", "-hint", "picture"},
		},
		{
			name: "hint-photo",
			hint: HintPhoto,
			args: []string{"-lossless", "-hint", "photo"},
		},
		{
			name: "hint-graph",
			hint: HintGraph,
			args: []string{"-lossless", "-hint", "graph"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), tc.args, filepath.Join(tempDir, "cwebp-output.webp"))

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.Lossless = true
			opts.ImageHint = tc.hint
			libOutput := webpEncodeWithLibrary(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_LosslessPreset tests the -z option (lossless preset)
func TestCompat_WebP_LosslessPreset(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name           string
		losslessPreset int
		args           []string
	}{
		{
			name:           "lossless-preset-0-fast",
			losslessPreset: 0,
			args:           []string{"-z", "0"},
		},
		{
			name:           "lossless-preset-3",
			losslessPreset: 3,
			args:           []string{"-z", "3"},
		},
		{
			name:           "lossless-preset-6",
			losslessPreset: 6,
			args:           []string{"-z", "6"},
		},
		{
			name:           "lossless-preset-9-best",
			losslessPreset: 9,
			args:           []string{"-z", "9"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, filepath.Join(testdataDir, "source/sizes/small-128x128.png"), tc.args, filepath.Join(tempDir, "cwebp-output.webp"))

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.LosslessPreset = tc.losslessPreset
			libOutput := webpEncodeWithLibrary(t, filepath.Join(testdataDir, "source/sizes/small-128x128.png"), opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_SNS tests the -sns option (spatial noise shaping)
func TestCompat_WebP_SNS(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name        string
		snsStrength int
		args        []string
	}{
		{
			name:        "sns-0",
			snsStrength: 0,
			args:        []string{"-sns", "0"},
		},
		{
			name:        "sns-50",
			snsStrength: 50,
			args:        []string{"-sns", "50"},
		},
		{
			name:        "sns-100",
			snsStrength: 100,
			args:        []string{"-sns", "100"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), tc.args, filepath.Join(tempDir, "cwebp-output.webp"))

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.SNSStrength = tc.snsStrength
			libOutput := webpEncodeWithLibrary(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_FilterOptions tests filter-related options
func TestCompat_WebP_FilterOptions(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name   string
		opts   func() WebPEncodeOptions
		args   []string
	}{
		{
			name: "filter-strength-0",
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.FilterStrength = 0
				return o
			},
			args: []string{"-f", "0"},
		},
		{
			name: "filter-strength-100",
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.FilterStrength = 100
				return o
			},
			args: []string{"-f", "100"},
		},
		{
			name: "sharpness-0",
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.FilterSharpness = 0
				return o
			},
			args: []string{"-sharpness", "0"},
		},
		{
			name: "sharpness-7",
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.FilterSharpness = 7
				return o
			},
			args: []string{"-sharpness", "7"},
		},
		{
			name: "strong-filter",
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.FilterType = FilterTypeStrong
				return o
			},
			args: []string{"-strong"},
		},
		{
			name: "autofilter",
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Autofilter = true
				return o
			},
			args: []string{"-af"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), tc.args, filepath.Join(tempDir, "cwebp-output.webp"))

			// ライブラリで変換
			libOutput := webpEncodeWithLibrary(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), tc.opts())

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_NearLossless tests the -near_lossless option
func TestCompat_WebP_NearLossless(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name         string
		nearLossless int
		args         []string
	}{
		{
			name:         "near-lossless-0",
			nearLossless: 0,
			args:         []string{"-near_lossless", "0"},
		},
		{
			name:         "near-lossless-50",
			nearLossless: 50,
			args:         []string{"-near_lossless", "50"},
		},
		{
			name:         "near-lossless-100",
			nearLossless: 100,
			args:         []string{"-near_lossless", "100"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), tc.args, filepath.Join(tempDir, "cwebp-output.webp"))

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.NearLossless = tc.nearLossless
			libOutput := webpEncodeWithLibrary(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_Segments tests the -segments option
func TestCompat_WebP_Segments(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name     string
		segments int
		args     []string
	}{
		{
			name:     "segments-1",
			segments: 1,
			args:     []string{"-segments", "1"},
		},
		{
			name:     "segments-4",
			segments: 4,
			args:     []string{"-segments", "4"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), tc.args, filepath.Join(tempDir, "cwebp-output.webp"))

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.Segments = tc.segments
			libOutput := webpEncodeWithLibrary(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_SharpYUV tests the -sharp_yuv option
func TestCompat_WebP_SharpYUV(t *testing.T) {
	setupCompatTest(t)

	// Check if cwebp supports -sharp_yuv option
	cmd := exec.Command("cwebp", "-h")
	output, _ := cmd.CombinedOutput()
	if !bytes.Contains(output, []byte("-sharp_yuv")) {
		t.Skip("cwebp doesn't support -sharp_yuv option")
	}

	testCases := []struct {
		name       string
		sharpYUV   bool
		args       []string
	}{
		{
			name:     "sharp-yuv",
			sharpYUV: true,
			args:     []string{"-sharp_yuv"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), tc.args, filepath.Join(tempDir, "cwebp-output.webp"))

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.UseSharpYUV = tc.sharpYUV
			libOutput := webpEncodeWithLibrary(t, filepath.Join(testdataDir, "source/sizes/medium-512x512.png"), opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_Metadata tests the -metadata option
func TestCompat_WebP_Metadata(t *testing.T) {
	setupCompatTest(t)

	// Create a test JPEG with EXIF data
	exifJPEG := filepath.Join(tempDir, "with-exif.jpg")

	// Use ImageMagick to create a JPEG with EXIF
	cmd := exec.Command("convert",
		filepath.Join(testdataDir, "source/sizes/medium-512x512.png"),
		"-set", "exif:DateTime", "2024:01:01 12:00:00",
		exifJPEG)
	if err := cmd.Run(); err != nil {
		t.Skip("ImageMagick not available, skipping metadata test")
	}

	testCases := []struct {
		name         string
		keepMetadata int
		args         []string
	}{
		{
			name:         "metadata-none",
			keepMetadata: 0,
			args:         []string{"-metadata", "none"},
		},
		{
			name:         "metadata-exif",
			keepMetadata: 1,
			args:         []string{"-metadata", "exif"},
		},
		{
			name:         "metadata-all",
			keepMetadata: 4,
			args:         []string{"-metadata", "all"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, exifJPEG, tc.args, filepath.Join(tempDir, "cwebp-output.webp"))

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.KeepMetadata = tc.keepMetadata
			libOutput := webpEncodeWithLibrary(t, exifJPEG, opts)

			// バイナリ完全一致チェック
			cmdSize := len(cmdOutput)
			libSize := len(libOutput)
			t.Logf("  cwebp:   %d bytes", cmdSize)
			t.Logf("  library: %d bytes", libSize)

			if bytes.Equal(cmdOutput, libOutput) {
				t.Logf("  ✓ PASSED: Binary exact match")
			} else {
				sizeDiff := cmdSize - libSize
				if sizeDiff < 0 {
					sizeDiff = -sizeDiff
				}
				sizeDiffPercent := float64(sizeDiff) * 100.0 / float64(cmdSize)
				t.Errorf("  ❌ FAILED: Binary mismatch (size difference: %d bytes, %.2f%%)", sizeDiff, sizeDiffPercent)
			}
		})
	}
}

// TestCompat_WebP_AlphaFiltering tests the -alpha_filter option
func TestCompat_WebP_AlphaFiltering(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name           string
		alphaFiltering WebPAlphaFilter
		args           []string
	}{
		{
			name:           "alpha-filter-none",
			alphaFiltering: AlphaFilterNone,
			args:           []string{"-alpha_filter", "none"},
		},
		{
			name:           "alpha-filter-fast",
			alphaFiltering: AlphaFilterFast,
			args:           []string{"-alpha_filter", "fast"},
		},
		{
			name:           "alpha-filter-best",
			alphaFiltering: AlphaFilterBest,
			args:           []string{"-alpha_filter", "best"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, "source/alpha/alpha-gradient.png")
			outputFile := filepath.Join(tempDir, "cwebp-output.webp")

			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, inputFile, tc.args, outputFile)

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.AlphaFiltering = tc.alphaFiltering
			libOutput := webpEncodeWithLibrary(t, inputFile, opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_Preprocessing tests the -pre option
func TestCompat_WebP_Preprocessing(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name          string
		preprocessing int
		args          []string
	}{
		{
			name:          "preprocessing-0",
			preprocessing: 0,
			args:          []string{"-pre", "0"},
		},
		{
			name:          "preprocessing-1",
			preprocessing: 1,
			args:          []string{"-pre", "1"},
		},
		{
			name:          "preprocessing-2",
			preprocessing: 2,
			args:          []string{"-pre", "2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
			outputFile := filepath.Join(tempDir, "cwebp-output.webp")

			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, inputFile, tc.args, outputFile)

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.Preprocessing = tc.preprocessing
			libOutput := webpEncodeWithLibrary(t, inputFile, opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_Partitions tests the -partition_limit option
func TestCompat_WebP_Partitions(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name           string
		partitionLimit int
		args           []string
	}{
		{
			name:           "partition-limit-0",
			partitionLimit: 0,
			args:           []string{"-partition_limit", "0"},
		},
		{
			name:           "partition-limit-50",
			partitionLimit: 50,
			args:           []string{"-partition_limit", "50"},
		},
		{
			name:           "partition-limit-100",
			partitionLimit: 100,
			args:           []string{"-partition_limit", "100"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
			outputFile := filepath.Join(tempDir, "cwebp-output.webp")

			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, inputFile, tc.args, outputFile)

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.PartitionLimit = tc.partitionLimit
			libOutput := webpEncodeWithLibrary(t, inputFile, opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_TargetSize tests the -size option
func TestCompat_WebP_TargetSize(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name       string
		targetSize int
		args       []string
	}{
		{
			name:       "target-size-1000",
			targetSize: 1000,
			args:       []string{"-size", "1000"},
		},
		{
			name:       "target-size-2000",
			targetSize: 2000,
			args:       []string{"-size", "2000"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
			outputFile := filepath.Join(tempDir, "cwebp-output.webp")

			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, inputFile, tc.args, outputFile)

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.TargetSize = tc.targetSize
			libOutput := webpEncodeWithLibrary(t, inputFile, opts)

			// ターゲットサイズの場合、完全一致は難しいのでサイズが近ければOK
			cmdSize := len(cmdOutput)
			libSize := len(libOutput)
			t.Logf("  cwebp:   %d bytes (target: %d)", cmdSize, tc.targetSize)
			t.Logf("  library: %d bytes (target: %d)", libSize, tc.targetSize)

			// バイナリ完全一致チェック
			if bytes.Equal(cmdOutput, libOutput) {
				t.Logf("  ✓ PASSED: Binary exact match")
			} else {
				sizeDiff := cmdSize - libSize
				if sizeDiff < 0 {
					sizeDiff = -sizeDiff
				}
				sizeDiffPercent := float64(sizeDiff) * 100.0 / float64(cmdSize)
				t.Errorf("  ❌ FAILED: Binary mismatch (size difference: %d bytes, %.2f%%)", sizeDiff, sizeDiffPercent)
			}
		})
	}
}

// TestCompat_WebP_TargetPSNR tests the -psnr option
func TestCompat_WebP_TargetPSNR(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name       string
		targetPSNR float32
		args       []string
	}{
		{
			name:       "target-psnr-40",
			targetPSNR: 40.0,
			args:       []string{"-psnr", "40"},
		},
		{
			name:       "target-psnr-45",
			targetPSNR: 45.0,
			args:       []string{"-psnr", "45"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
			outputFile := filepath.Join(tempDir, "cwebp-output.webp")

			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, inputFile, tc.args, outputFile)

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.TargetPSNR = tc.targetPSNR
			libOutput := webpEncodeWithLibrary(t, inputFile, opts)

			// PSNR目標の場合、サイズ差が大きくなる可能性があるので緩い基準
			cmdSize := len(cmdOutput)
			libSize := len(libOutput)
			t.Logf("  cwebp:   %d bytes (target PSNR: %.1f)", cmdSize, tc.targetPSNR)
			t.Logf("  library: %d bytes (target PSNR: %.1f)", libSize, tc.targetPSNR)

			// バイナリ完全一致チェック
			if bytes.Equal(cmdOutput, libOutput) {
				t.Logf("  ✓ PASSED: Binary exact match")
			} else {
				sizeDiff := cmdSize - libSize
				if sizeDiff < 0 {
					sizeDiff = -sizeDiff
				}
				sizeDiffPercent := float64(sizeDiff) * 100.0 / float64(cmdSize)
				t.Errorf("  ❌ FAILED: Binary mismatch (size difference: %d bytes, %.2f%%)", sizeDiff, sizeDiffPercent)
			}
		})
	}
}

// TestCompat_WebP_LowMemory tests the -low_memory option
func TestCompat_WebP_LowMemory(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name      string
		lowMemory bool
		args      []string
	}{
		{
			name:      "low-memory",
			lowMemory: true,
			args:      []string{"-low_memory"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, "source/sizes/large-2048x2048.png")
			outputFile := filepath.Join(tempDir, "cwebp-output.webp")

			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, inputFile, tc.args, outputFile)

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.LowMemory = tc.lowMemory
			libOutput := webpEncodeWithLibrary(t, inputFile, opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// TestCompat_WebP_QMinQMax tests the -qrange option
func TestCompat_WebP_QMinQMax(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name string
		qmin int
		qmax int
		args []string
	}{
		{
			name: "qrange-0-50",
			qmin: 0,
			qmax: 50,
			args: []string{"-qrange", "0", "50"},
		},
		{
			name: "qrange-50-100",
			qmin: 50,
			qmax: 100,
			args: []string{"-qrange", "50", "100"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
			outputFile := filepath.Join(tempDir, "cwebp-output.webp")

			// cwebpコマンドで変換
			cmdOutput := runCWebP(t, inputFile, tc.args, outputFile)

			// ライブラリで変換
			opts := DefaultWebPEncodeOptions()
			opts.QMin = tc.qmin
			opts.QMax = tc.qmax
			libOutput := webpEncodeWithLibrary(t, inputFile, opts)

			// 出力を比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// Helper: ライブラリでWebPエンコード
func webpEncodeWithLibrary(t *testing.T, inputPath string, opts WebPEncodeOptions) []byte {
	t.Helper()

	data, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	webpData, err := WebPEncodeBytes(data, opts)
	if err != nil {
		t.Fatalf("Library WebPEncodeBytes failed: %v", err)
	}

	return webpData
}
