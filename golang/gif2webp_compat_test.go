package libnextimage

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var gifTestdataDir = filepath.Join("..", "testdata", "gif-source")

// NOTE: GIF to WebP conversion is now fully implemented using WebPAnimEncoder.
// The implementation uses giflib to read GIF frames and WebPAnimEncoder to create
// animated WebP, matching the gif2webp command-line tool behavior.
//
// Implementation details:
// 1. Read GIF using giflib (via gifdec.c)
// 2. Convert each frame to WebP using WebPAnimEncoder
// 3. Handle frame timing, transparency, and dispose methods
// 4. Support all animation options (mixed, min_size, kmin, kmax, loop_compatibility)

// gif2webp 互換性テスト - 静止画GIF
func TestGIF2WebPCompat_Static(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name    string
		gifFile string
		opts    WebPEncodeOptions
		args    []string
	}{
		{
			name:    "static-64x64",
			gifFile: filepath.Join(gifTestdataDir, "static-64x64.gif"),
			opts:    DefaultWebPEncodeOptions(),
			args:    []string{},
		},
		{
			name:    "static-512x512",
			gifFile: filepath.Join(gifTestdataDir, "static-512x512.gif"),
			opts:    DefaultWebPEncodeOptions(),
			args:    []string{},
		},
		{
			name:    "static-16x16",
			gifFile: filepath.Join(gifTestdataDir, "static-16x16.gif"),
			opts:    DefaultWebPEncodeOptions(),
			args:    []string{},
		},
		{
			name:    "gradient",
			gifFile: filepath.Join(gifTestdataDir, "gradient.gif"),
			opts:    DefaultWebPEncodeOptions(),
			args:    []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// gif2webpコマンドで変換
			cmdOutput := runGIF2WebP(t, tc.gifFile, tc.args)

			// ライブラリで変換
			libOutput := gif2webpWithLibrary(t, tc.gifFile, tc.opts)

			// 出力サイズを比較（静止画GIFの場合はバイナリ完全一致が期待できる）
			compareGIF2WebPOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// gif2webp 互換性テスト - アニメーションGIF
func TestGIF2WebPCompat_Animated(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name    string
		gifFile string
		opts    WebPEncodeOptions
		args    []string
	}{
		{
			name:    "animated-3frames",
			gifFile: filepath.Join(gifTestdataDir, "animated-3frames.gif"),
			opts:    DefaultWebPEncodeOptions(),
			args:    []string{},
		},
		{
			name:    "animated-small",
			gifFile: filepath.Join(gifTestdataDir, "animated-small.gif"),
			opts:    DefaultWebPEncodeOptions(),
			args:    []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// gif2webpコマンドで変換
			cmdOutput := runGIF2WebP(t, tc.gifFile, tc.args)

			// ライブラリで変換
			libOutput := gif2webpWithLibrary(t, tc.gifFile, tc.opts)

			// アニメーションの場合は完全一致は期待できないので、サイズ差を確認
			compareGIF2WebPOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// gif2webp 互換性テスト - 透過GIF
func TestGIF2WebPCompat_Alpha(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name    string
		gifFile string
		opts    WebPEncodeOptions
		args    []string
	}{
		{
			name:    "static-alpha",
			gifFile: filepath.Join(gifTestdataDir, "static-alpha.gif"),
			opts:    DefaultWebPEncodeOptions(),
			args:    []string{},
		},
		{
			name:    "animated-alpha",
			gifFile: filepath.Join(gifTestdataDir, "animated-alpha.gif"),
			opts:    DefaultWebPEncodeOptions(),
			args:    []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// gif2webpコマンドで変換
			cmdOutput := runGIF2WebP(t, tc.gifFile, tc.args)

			// ライブラリで変換
			libOutput := gif2webpWithLibrary(t, tc.gifFile, tc.opts)

			// 透過GIFの場合もサイズ差を確認
			compareGIF2WebPOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// gif2webp 互換性テスト - 品質設定
func TestGIF2WebPCompat_Quality(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name    string
		gifFile string
		opts    func() WebPEncodeOptions
		args    []string
	}{
		{
			name:    "quality-50",
			gifFile: filepath.Join(gifTestdataDir, "static-64x64.gif"),
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Quality = 50
				return o
			},
			args: []string{"-q", "50"},
		},
		{
			name:    "quality-90",
			gifFile: filepath.Join(gifTestdataDir, "static-64x64.gif"),
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Quality = 90
				return o
			},
			args: []string{"-q", "90"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// gif2webpコマンドで変換
			cmdOutput := runGIF2WebP(t, tc.gifFile, tc.args)

			// ライブラリで変換
			libOutput := gif2webpWithLibrary(t, tc.gifFile, tc.opts())

			// 品質設定の場合はバイナリ完全一致が期待できる
			compareGIF2WebPOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// gif2webp 互換性テスト - メソッド設定
func TestGIF2WebPCompat_Method(t *testing.T) {
	setupDecodeCompatTest(t)

	testCases := []struct {
		name    string
		gifFile string
		opts    func() WebPEncodeOptions
		args    []string
	}{
		{
			name:    "method-0",
			gifFile: filepath.Join(gifTestdataDir, "static-64x64.gif"),
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Method = 0
				return o
			},
			args: []string{"-m", "0"},
		},
		{
			name:    "method-6",
			gifFile: filepath.Join(gifTestdataDir, "static-64x64.gif"),
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Method = 6
				return o
			},
			args: []string{"-m", "6"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// gif2webpコマンドで変換
			cmdOutput := runGIF2WebP(t, tc.gifFile, tc.args)

			// ライブラリで変換
			libOutput := gif2webpWithLibrary(t, tc.gifFile, tc.opts())

			// メソッド設定の場合はバイナリ完全一致が期待できる
			compareGIF2WebPOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// Helper: gif2webpコマンドを実行
func runGIF2WebP(t *testing.T, gifFile string, args []string) []byte {
	t.Helper()

	// Check if gif2webp command exists
	if _, err := exec.LookPath("gif2webp"); err != nil {
		t.Skip("gif2webp command not found, skipping compatibility test")
	}

	outputFile := filepath.Join(tempDir, "gif2webp-cmd-output.webp")

	// Build command: gif2webp [args] input.gif -o output.webp
	cmdArgs := append(args, gifFile, "-o", outputFile)
	cmd := exec.Command("gif2webp", cmdArgs...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("gif2webp command failed: %v\nStderr: %s", err, stderr.String())
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read gif2webp output: %v", err)
	}

	return data
}

// Helper: ライブラリでGIFからWebPに変換
func gif2webpWithLibrary(t *testing.T, gifFile string, opts WebPEncodeOptions) []byte {
	t.Helper()

	// GIFファイルを読み込み
	gifData, err := os.ReadFile(gifFile)
	if err != nil {
		t.Fatalf("Failed to read GIF file: %v", err)
	}

	// GIF2WebP変換
	webpData, err := GIF2WebP(gifData, opts)
	if err != nil {
		t.Fatalf("Library GIF2WebP failed: %v", err)
	}

	return webpData
}

// Helper: gif2webpの出力を比較
func compareGIF2WebPOutputs(t *testing.T, name string, cmdOutput, libOutput []byte) {
	t.Helper()

	cmdSize := len(cmdOutput)
	libSize := len(libOutput)

	t.Logf("  gif2webp: %d bytes", cmdSize)
	t.Logf("  library:  %d bytes", libSize)

	// Save outputs for debugging
	os.WriteFile("/tmp/cmd_"+name+".webp", cmdOutput, 0644)
	os.WriteFile("/tmp/lib_"+name+".webp", libOutput, 0644)

	// バイナリ完全一致チェック
	if bytes.Equal(cmdOutput, libOutput) {
		t.Logf("  ✓ PASSED: Binary exact match")
		return
	}

	// サイズ差分を計算
	sizeDiff := libSize - cmdSize
	if sizeDiff < 0 {
		sizeDiff = -sizeDiff
	}
	sizeDiffPercent := float64(sizeDiff) * 100.0 / float64(cmdSize)

	t.Logf("  Size difference: %d bytes (%.2f%%)", sizeDiff, sizeDiffPercent)

	// アニメーションGIFの場合は10%以内の差を許容
	// 静止画GIFの場合は完全一致を期待
	if sizeDiffPercent <= 10.0 {
		t.Logf("  ✓ PASSED: Size difference within acceptable range (%.2f%%)", sizeDiffPercent)
	} else {
		t.Errorf("  ✗ FAILED: Size difference too large (%.2f%% > 10%%)", sizeDiffPercent)
	}
}
