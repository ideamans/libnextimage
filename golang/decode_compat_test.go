package libnextimage

import (
	"bytes"
	"path/filepath"
	"testing"
)

// デコード結果のピクセルデータ比較
func compareDecodeOutputs(t *testing.T, testName string, cmdPNGFile, libPNGFile string) {
	// 両方のPNGファイルをデコード
	cmdPixels, cmdW, cmdH, err := decodePNGToRGBA(cmdPNGFile)
	if err != nil {
		t.Fatalf("Failed to decode command PNG: %v", err)
	}

	libPixels, libW, libH, err := decodePNGToRGBA(libPNGFile)
	if err != nil {
		t.Fatalf("Failed to decode library PNG: %v", err)
	}

	t.Logf("  dwebp:   %dx%d, %d bytes", cmdW, cmdH, len(cmdPixels))
	t.Logf("  library: %dx%d, %d bytes", libW, libH, len(libPixels))

	// サイズチェック
	if cmdW != libW || cmdH != libH {
		t.Errorf("  ❌ FAILED: Dimension mismatch (cmd: %dx%d, lib: %dx%d)", cmdW, cmdH, libW, libH)
		return
	}

	// ピクセルデータの完全一致チェック
	if bytes.Equal(cmdPixels, libPixels) {
		t.Logf("  ✓ PASSED: Pixel data exact match")
		return
	}

	// ピクセル単位の差分を計算（デバッグ情報用）
	diffCount := 0
	maxDiff := 0
	for i := 0; i < len(cmdPixels); i++ {
		diff := int(cmdPixels[i]) - int(libPixels[i])
		if diff < 0 {
			diff = -diff
		}
		if diff > 0 {
			diffCount++
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}

	diffPercent := float64(diffCount) * 100.0 / float64(len(cmdPixels))
	t.Logf("  Pixel differences: %d/%d (%.2f%%), max diff: %d", diffCount, len(cmdPixels), diffPercent, maxDiff)
	t.Errorf("  ❌ FAILED: Pixel data mismatch")
}

// dwebp デコード互換性テスト - 事前準備したWebPファイルを使用
func TestDecodeCompat_WebP_Default(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name       string
		webpFile   string
		decodeOpts WebPDecodeOptions
		dwebpArgs  []string
	}{
		{
			name:       "lossy-q75",
			webpFile:   filepath.Join(testdataDir, "webp-samples/lossy-q75.webp"),
			decodeOpts: DefaultWebPDecodeOptions(),
			dwebpArgs:  []string{}, // デフォルトオプション
		},
		{
			name:       "lossy-q90",
			webpFile:   filepath.Join(testdataDir, "webp-samples/lossy-q90.webp"),
			decodeOpts: DefaultWebPDecodeOptions(),
			dwebpArgs:  []string{},
		},
		{
			name:       "lossless",
			webpFile:   filepath.Join(testdataDir, "webp-samples/lossless.webp"),
			decodeOpts: DefaultWebPDecodeOptions(),
			dwebpArgs:  []string{},
		},
		{
			name:       "alpha-gradient",
			webpFile:   filepath.Join(testdataDir, "webp-samples/alpha-gradient.webp"),
			decodeOpts: DefaultWebPDecodeOptions(),
			dwebpArgs:  []string{},
		},
		{
			name:       "alpha-lossless",
			webpFile:   filepath.Join(testdataDir, "webp-samples/alpha-lossless.webp"),
			decodeOpts: DefaultWebPDecodeOptions(),
			dwebpArgs:  []string{},
		},
		{
			name:       "small-128x128",
			webpFile:   filepath.Join(testdataDir, "webp-samples/small-128x128.webp"),
			decodeOpts: DefaultWebPDecodeOptions(),
			dwebpArgs:  []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// dwebpコマンドでデコード
			cmdPNGFile := filepath.Join(tempDir, "decode-cmd-output", tc.name+".png")
			runDWebP(t, tc.webpFile, tc.dwebpArgs, cmdPNGFile)

			// ライブラリでデコード
			libPNGFile := filepath.Join(tempDir, "decode-lib-output", tc.name+".png")
			decodeWithLibraryToFile(t, tc.webpFile, libPNGFile, tc.decodeOpts)

			// ピクセルデータを比較
			compareDecodeOutputs(t, tc.name, cmdPNGFile, libPNGFile)
		})
	}
}

// dwebp デコード互換性テスト - NoFancyUpsampling
func TestDecodeCompat_WebP_NoFancy(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name       string
		webpFile   string
		decodeOpts WebPDecodeOptions
		dwebpArgs  []string
	}{
		{
			name:     "lossy-q75-nofancy",
			webpFile: filepath.Join(testdataDir, "webp-samples/lossy-q75.webp"),
			decodeOpts: func() WebPDecodeOptions {
				o := DefaultWebPDecodeOptions()
				o.NoFancyUpsampling = true
				return o
			}(),
			dwebpArgs: []string{"-nofancy"},
		},
		{
			name:     "alpha-gradient-nofancy",
			webpFile: filepath.Join(testdataDir, "webp-samples/alpha-gradient.webp"),
			decodeOpts: func() WebPDecodeOptions {
				o := DefaultWebPDecodeOptions()
				o.NoFancyUpsampling = true
				return o
			}(),
			dwebpArgs: []string{"-nofancy"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// dwebpコマンドでデコード
			cmdPNGFile := filepath.Join(tempDir, "decode-cmd-output", tc.name+".png")
			runDWebP(t, tc.webpFile, tc.dwebpArgs, cmdPNGFile)

			// ライブラリでデコード
			libPNGFile := filepath.Join(tempDir, "decode-lib-output", tc.name+".png")
			decodeWithLibraryToFile(t, tc.webpFile, libPNGFile, tc.decodeOpts)

			// ピクセルデータを比較
			compareDecodeOutputs(t, tc.name, cmdPNGFile, libPNGFile)
		})
	}
}

// dwebp デコード互換性テスト - NoFilter
func TestDecodeCompat_WebP_NoFilter(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name       string
		webpFile   string
		decodeOpts WebPDecodeOptions
		dwebpArgs  []string
	}{
		{
			name:     "lossy-q75-nofilter",
			webpFile: filepath.Join(testdataDir, "webp-samples/lossy-q75.webp"),
			decodeOpts: func() WebPDecodeOptions {
				o := DefaultWebPDecodeOptions()
				o.BypassFiltering = true
				return o
			}(),
			dwebpArgs: []string{"-nofilter"},
		},
		{
			name:     "alpha-gradient-nofilter",
			webpFile: filepath.Join(testdataDir, "webp-samples/alpha-gradient.webp"),
			decodeOpts: func() WebPDecodeOptions {
				o := DefaultWebPDecodeOptions()
				o.BypassFiltering = true
				return o
			}(),
			dwebpArgs: []string{"-nofilter"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// dwebpコマンドでデコード
			cmdPNGFile := filepath.Join(tempDir, "decode-cmd-output", tc.name+".png")
			runDWebP(t, tc.webpFile, tc.dwebpArgs, cmdPNGFile)

			// ライブラリでデコード
			libPNGFile := filepath.Join(tempDir, "decode-lib-output", tc.name+".png")
			decodeWithLibraryToFile(t, tc.webpFile, libPNGFile, tc.decodeOpts)

			// ピクセルデータを比較
			compareDecodeOutputs(t, tc.name, cmdPNGFile, libPNGFile)
		})
	}
}

// dwebp デコード互換性テスト - UseThreads
func TestDecodeCompat_WebP_MT(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name       string
		webpFile   string
		decodeOpts WebPDecodeOptions
		dwebpArgs  []string
	}{
		{
			name:     "large-2048x2048-mt",
			webpFile: filepath.Join(testdataDir, "webp-samples/large-2048x2048.webp"),
			decodeOpts: func() WebPDecodeOptions {
				o := DefaultWebPDecodeOptions()
				o.UseThreads = true
				return o
			}(),
			dwebpArgs: []string{"-mt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// dwebpコマンドでデコード
			cmdPNGFile := filepath.Join(tempDir, "decode-cmd-output", tc.name+".png")
			runDWebP(t, tc.webpFile, tc.dwebpArgs, cmdPNGFile)

			// ライブラリでデコード
			libPNGFile := filepath.Join(tempDir, "decode-lib-output", tc.name+".png")
			decodeWithLibraryToFile(t, tc.webpFile, libPNGFile, tc.decodeOpts)

			// ピクセルデータを比較
			compareDecodeOutputs(t, tc.name, cmdPNGFile, libPNGFile)
		})
	}
}

// dwebp デコード互換性テスト - オプション組み合わせ
func TestDecodeCompat_WebP_OptionCombinations(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name       string
		webpFile   string
		decodeOpts WebPDecodeOptions
		dwebpArgs  []string
	}{
		{
			name:     "lossy-q75-nofancy-nofilter",
			webpFile: filepath.Join(testdataDir, "webp-samples/lossy-q75.webp"),
			decodeOpts: func() WebPDecodeOptions {
				o := DefaultWebPDecodeOptions()
				o.NoFancyUpsampling = true
				o.BypassFiltering = true
				return o
			}(),
			dwebpArgs: []string{"-nofancy", "-nofilter"},
		},
		{
			name:     "large-2048x2048-nofancy-mt",
			webpFile: filepath.Join(testdataDir, "webp-samples/large-2048x2048.webp"),
			decodeOpts: func() WebPDecodeOptions {
				o := DefaultWebPDecodeOptions()
				o.NoFancyUpsampling = true
				o.UseThreads = true
				return o
			}(),
			dwebpArgs: []string{"-nofancy", "-mt"},
		},
		{
			name:     "alpha-gradient-nofancy-nofilter",
			webpFile: filepath.Join(testdataDir, "webp-samples/alpha-gradient.webp"),
			decodeOpts: func() WebPDecodeOptions {
				o := DefaultWebPDecodeOptions()
				o.NoFancyUpsampling = true
				o.BypassFiltering = true
				return o
			}(),
			dwebpArgs: []string{"-nofancy", "-nofilter"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// dwebpコマンドでデコード
			cmdPNGFile := filepath.Join(tempDir, "decode-cmd-output", tc.name+".png")
			runDWebP(t, tc.webpFile, tc.dwebpArgs, cmdPNGFile)

			// ライブラリでデコード
			libPNGFile := filepath.Join(tempDir, "decode-lib-output", tc.name+".png")
			decodeWithLibraryToFile(t, tc.webpFile, libPNGFile, tc.decodeOpts)

			// ピクセルデータを比較
			compareDecodeOutputs(t, tc.name, cmdPNGFile, libPNGFile)
		})
	}
}
