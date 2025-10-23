package libnextimage

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ideamans/libnextimage/golang/cwebp"
	"github.com/ideamans/libnextimage/golang/dwebp"
)

// テスト用のベースディレクトリ
var (
	projectRoot = filepath.Join("..")
	binDir      = filepath.Join(projectRoot, "bin")
	testdataDir = filepath.Join(projectRoot, "testdata")
	tempDir     = filepath.Join(projectRoot, "build-test-compat")
)

// テストのセットアップ
func setupCompatTest(t *testing.T) {
	// 一時ディレクトリを作成
	if err := os.MkdirAll(filepath.Join(tempDir, "cmd-output"), 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "lib-output"), 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// cwebpコマンドの存在確認
	cwebpPath := filepath.Join(binDir, "cwebp")
	if _, err := os.Stat(cwebpPath); os.IsNotExist(err) {
		t.Skipf("cwebp not found at %s, run scripts/build-cli-tools.sh first", cwebpPath)
	}
}

// デコードテストのセットアップ
func setupDecodeCompatTest(t *testing.T) {
	// 一時ディレクトリを作成
	if err := os.MkdirAll(filepath.Join(tempDir, "decode-cmd-output"), 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "decode-lib-output"), 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "webp-samples"), 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// dwebpコマンドの存在確認
	dwebpPath := filepath.Join(binDir, "dwebp")
	if _, err := os.Stat(dwebpPath); os.IsNotExist(err) {
		t.Skipf("dwebp not found at %s, run scripts/build-cli-tools.sh first", dwebpPath)
	}
}

// cwebpコマンドを実行してWebPエンコード
func runCWebP(t *testing.T, inputFile string, args []string, outputFile string) []byte {
	cwebpPath := filepath.Join(binDir, "cwebp")

	cmdArgs := append(args, inputFile, "-o", outputFile)
	cmd := exec.Command(cwebpPath, cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("cwebp output: %s", string(output))
		t.Fatalf("cwebp command failed: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read cwebp output: %v", err)
	}

	return data
}

// WebPEncodeOptionsをcwebp.Optionsに変換
func convertToCWebPOptions(opts WebPEncodeOptions) cwebp.Options {
	return cwebp.Options{
		Quality:          opts.Quality,
		Lossless:         opts.Lossless,
		Method:           opts.Method,
		Preset:           int(opts.Preset),
		ImageHint:        int(opts.ImageHint),
		LosslessPreset:   opts.LosslessPreset,
		TargetSize:       opts.TargetSize,
		TargetPSNR:       opts.TargetPSNR,
		Segments:         opts.Segments,
		SNSStrength:      opts.SNSStrength,
		FilterStrength:   opts.FilterStrength,
		FilterSharpness:  opts.FilterSharpness,
		FilterType:       int(opts.FilterType),
		Autofilter:       opts.Autofilter,
		AlphaCompression: opts.AlphaMethod,
		AlphaFiltering:   int(opts.AlphaFiltering),
		AlphaQuality:     opts.AlphaQuality,
		Pass:             opts.Pass,
		ShowCompressed:   opts.ShowCompressed,
		Preprocessing:    opts.Preprocessing,
		Partitions:       opts.Partitions,
		PartitionLimit:   opts.PartitionLimit,
		EmulateJPEGSize:  opts.EmulateJPEGSize,
		ThreadLevel:      boolToInt(opts.ThreadLevel),
		LowMemory:        opts.LowMemory,
		NearLossless:     opts.NearLossless,
		Exact:            opts.Exact,
		UseDeltaPalette:  opts.UseDeltaPalette,
		UseSharpYUV:      opts.UseSharpYUV,
		QMin:             opts.QMin,
		QMax:             opts.QMax,
		KeepMetadata:     opts.KeepMetadata,
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ライブラリでWebPエンコード（コマンドAPIを使用）
func encodeWithLibrary(t *testing.T, inputFile string, opts WebPEncodeOptions) []byte {
	t.Helper()

	// WebPEncodeOptionsをcwebp.Optionsに変換
	cwebpOpts := convertToCWebPOptions(opts)

	// コマンドを作成
	cmd, err := cwebp.NewCommand(&cwebpOpts)
	if err != nil {
		t.Fatalf("Failed to create cwebp command: %v", err)
	}
	defer cmd.Close()

	// ファイルを読み込み
	data, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// エンコード実行
	webpData, err := cmd.Run(data)
	if err != nil {
		t.Fatalf("Library encode failed: %v", err)
	}

	return webpData
}

// dwebpコマンドを実行してPNGデコード
func runDWebP(t *testing.T, inputFile string, args []string, outputFile string) []byte {
	dwebpPath := filepath.Join(binDir, "dwebp")

	cmdArgs := append(args, inputFile, "-o", outputFile)
	cmd := exec.Command(dwebpPath, cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("dwebp output: %s", string(output))
		t.Fatalf("dwebp command failed: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read dwebp output: %v", err)
	}

	return data
}

// PNG書き込みヘルパー
func writePNGHelper(filePath string, data []byte, width, height int, format PixelFormat) error {
	// NRGBAイメージを作成（非プリマルチプライアルファ）
	// WebPのMODE_RGBAは非プリマルチプライなので、NRGBAを使用する
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// ピクセルデータをコピー
	if format == FormatRGBA {
		copy(img.Pix, data)
	} else {
		return fmt.Errorf("unsupported format: %d", format)
	}

	// ファイルに書き込み
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// デフォルトのPNGエンコード（圧縮レベルはDefaultCompression）
	return png.Encode(f, img)
}

// PNGをデコードしてRGBAデータとして読み込む
func decodePNGToRGBA(filePath string) ([]byte, int, int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, 0, 0, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// RGBAに変換（原点を0,0にする）
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	return rgba.Pix, width, height, nil
}

// WebPDecodeOptionsをdwebp.Optionsに変換
func convertToDWebPOptions(opts WebPDecodeOptions) dwebp.Options {
	return dwebp.Options{
		Format:            formatToString(opts.Format),
		BypassFiltering:   opts.BypassFiltering,
		NoFancyUpsampling: opts.NoFancyUpsampling,
		UseThreads:        opts.UseThreads,
		CropX:             opts.CropX,
		CropY:             opts.CropY,
		CropWidth:         opts.CropWidth,
		CropHeight:        opts.CropHeight,
		UseCrop:           opts.UseCrop,
		ResizeWidth:       opts.ResizeWidth,
		ResizeHeight:      opts.ResizeHeight,
		UseResize:         opts.UseResize,
		Flip:              opts.Flip,
	}
}

func formatToString(format PixelFormat) string {
	switch format {
	case FormatRGB:
		return "RGB"
	case FormatBGRA:
		return "BGRA"
	default:
		return "RGBA"
	}
}

// ライブラリでWebPデコード（PNG出力、コマンドAPIを使用）
func decodeWithLibraryToFile(t *testing.T, webpFile string, pngFile string, opts WebPDecodeOptions) {
	t.Helper()

	// WebPファイルを読み込み
	webpData, err := os.ReadFile(webpFile)
	if err != nil {
		t.Fatalf("Failed to read WebP file: %v", err)
	}

	// WebPDecodeOptionsをdwebp.Optionsに変換
	dwebpOpts := convertToDWebPOptions(opts)

	// コマンドを作成
	cmd, err := dwebp.NewCommand(&dwebpOpts)
	if err != nil {
		t.Fatalf("Failed to create dwebp command: %v", err)
	}
	defer cmd.Close()

	// デコード実行（PNG形式で出力）
	pngData, err := cmd.Run(webpData)
	if err != nil {
		t.Fatalf("Library decode failed: %v", err)
	}

	// PNGファイルに書き込み
	err = os.WriteFile(pngFile, pngData, 0644)
	if err != nil {
		t.Fatalf("Failed to write PNG file: %v", err)
	}
}

// バイナリ完全一致比較（CLI Clone Philosophyに基づき、バイナリ完全一致のみを合格とする）
func compareOutputs(t *testing.T, testName string, cmdOutput, libOutput []byte) {
	cmdSize := len(cmdOutput)
	libSize := len(libOutput)

	t.Logf("  cwebp:   %d bytes", cmdSize)
	t.Logf("  library: %d bytes", libSize)

	// バイナリ完全一致チェック
	if bytes.Equal(cmdOutput, libOutput) {
		t.Logf("  ✓ PASSED: Binary exact match")
		return
	}

	// バイナリ不一致の場合は失敗
	sizeDiff := libSize - cmdSize
	if sizeDiff < 0 {
		sizeDiff = -sizeDiff
	}
	sizeDiffPercent := float64(sizeDiff) * 100.0 / float64(cmdSize)

	t.Errorf("  ❌ FAILED: Binary mismatch (size difference: %d bytes, %.2f%%)", sizeDiff, sizeDiffPercent)
}

// Quality オプションテスト
func TestCompat_WebP_Quality(t *testing.T) {
	setupCompatTest(t)

	inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	qualities := []float32{0, 25, 50, 75, 90, 100}

	for _, q := range qualities {
		t.Run(fmt.Sprintf("quality-%.0f", q), func(t *testing.T) {
			testName := fmt.Sprintf("quality-%.0f", q)

			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				[]string{"-q", fmt.Sprintf("%.0f", q)},
				filepath.Join(tempDir, "cmd-output", testName+".webp"))

			// ライブラリ実行
			opts := DefaultWebPEncodeOptions()
			opts.Quality = q
			libOutput := encodeWithLibrary(t, inputFile, opts)

			// 比較
			compareOutputs(t, testName, cmdOutput, libOutput)
		})
	}
}

// Lossless モードテスト
func TestCompat_WebP_Lossless(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name      string
		inputFile string
	}{
		{"lossless-medium", "source/sizes/medium-512x512.png"},
		{"lossless-small", "source/sizes/small-128x128.png"},
		{"lossless-alpha", "source/alpha/alpha-gradient.png"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, tc.inputFile)

			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				[]string{"-lossless"},
				filepath.Join(tempDir, "cmd-output", tc.name+".webp"))

			// ライブラリ実行
			opts := DefaultWebPEncodeOptions()
			opts.Lossless = true
			libOutput := encodeWithLibrary(t, inputFile, opts)

			// 比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// Method オプションテスト
func TestCompat_WebP_Method(t *testing.T) {
	setupCompatTest(t)

	inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	methods := []int{0, 2, 4, 6}

	for _, m := range methods {
		t.Run(fmt.Sprintf("method-%d", m), func(t *testing.T) {
			testName := fmt.Sprintf("method-%d", m)

			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				[]string{"-q", "75", "-m", fmt.Sprintf("%d", m)},
				filepath.Join(tempDir, "cmd-output", testName+".webp"))

			// ライブラリ実行
			opts := DefaultWebPEncodeOptions()
			opts.Quality = 75
			opts.Method = m
			libOutput := encodeWithLibrary(t, inputFile, opts)

			// 比較
			compareOutputs(t, testName, cmdOutput, libOutput)
		})
	}
}

// サイズバリエーションテスト
func TestCompat_WebP_Sizes(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name      string
		inputFile string
	}{
		{"tiny-16x16", "source/sizes/tiny-16x16.png"},
		{"small-128x128", "source/sizes/small-128x128.png"},
		{"medium-512x512", "source/sizes/medium-512x512.png"},
		{"large-2048x2048", "source/sizes/large-2048x2048.png"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, tc.inputFile)

			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				[]string{"-q", "80"},
				filepath.Join(tempDir, "cmd-output", tc.name+".webp"))

			// ライブラリ実行
			opts := DefaultWebPEncodeOptions()
			opts.Quality = 80
			libOutput := encodeWithLibrary(t, inputFile, opts)

			// 比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// 透明度テスト
func TestCompat_WebP_Alpha(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name      string
		inputFile string
	}{
		{"alpha-opaque", "source/alpha/opaque.png"},
		{"alpha-transparent", "source/alpha/transparent.png"},
		{"alpha-gradient", "source/alpha/alpha-gradient.png"},
		{"alpha-complex", "source/alpha/alpha-complex.png"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, tc.inputFile)

			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				[]string{"-q", "75"},
				filepath.Join(tempDir, "cmd-output", tc.name+".webp"))

			// ライブラリ実行
			opts := DefaultWebPEncodeOptions()
			opts.Quality = 75
			libOutput := encodeWithLibrary(t, inputFile, opts)

			// 比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// 圧縮特性テスト
func TestCompat_WebP_Compression(t *testing.T) {
	setupCompatTest(t)

	testCases := []struct {
		name      string
		inputFile string
	}{
		{"flat-color", "source/compression/flat-color.png"},
		{"noisy", "source/compression/noisy.png"},
		{"edges", "source/compression/edges.png"},
		{"text", "source/compression/text.png"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := filepath.Join(testdataDir, tc.inputFile)

			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				[]string{"-q", "75"},
				filepath.Join(tempDir, "cmd-output", tc.name+".webp"))

			// ライブラリ実行
			opts := DefaultWebPEncodeOptions()
			opts.Quality = 75
			libOutput := encodeWithLibrary(t, inputFile, opts)

			// 比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}

// Alpha Quality オプションテスト
func TestCompat_WebP_AlphaQuality(t *testing.T) {
	setupCompatTest(t)

	inputFile := filepath.Join(testdataDir, "source/alpha/alpha-gradient.png")
	alphaQualities := []int{0, 50, 100}

	for _, aq := range alphaQualities {
		t.Run(fmt.Sprintf("alpha-q-%d", aq), func(t *testing.T) {
			testName := fmt.Sprintf("alpha-q-%d", aq)

			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				[]string{"-q", "75", "-alpha_q", fmt.Sprintf("%d", aq)},
				filepath.Join(tempDir, "cmd-output", testName+".webp"))

			// ライブラリ実行
			opts := DefaultWebPEncodeOptions()
			opts.Quality = 75
			opts.AlphaQuality = aq
			libOutput := encodeWithLibrary(t, inputFile, opts)

			// 比較
			compareOutputs(t, testName, cmdOutput, libOutput)
		})
	}
}

// Exact モードテスト
func TestCompat_WebP_Exact(t *testing.T) {
	setupCompatTest(t)

	inputFile := filepath.Join(testdataDir, "source/alpha/alpha-gradient.png")

	t.Run("exact-mode", func(t *testing.T) {
		// cwebp実行
		cmdOutput := runCWebP(t, inputFile,
			[]string{"-q", "75", "-exact"},
			filepath.Join(tempDir, "cmd-output", "exact.webp"))

		// ライブラリ実行
		opts := DefaultWebPEncodeOptions()
		opts.Quality = 75
		opts.Exact = true
		libOutput := encodeWithLibrary(t, inputFile, opts)

		// 比較
		compareOutputs(t, "exact-mode", cmdOutput, libOutput)
	})
}

// Pass (エントロピー解析パス数) テスト
func TestCompat_WebP_Pass(t *testing.T) {
	setupCompatTest(t)

	inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")
	passes := []int{1, 5, 10}

	for _, p := range passes {
		t.Run(fmt.Sprintf("pass-%d", p), func(t *testing.T) {
			testName := fmt.Sprintf("pass-%d", p)

			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				[]string{"-q", "75", "-pass", fmt.Sprintf("%d", p)},
				filepath.Join(tempDir, "cmd-output", testName+".webp"))

			// ライブラリ実行
			opts := DefaultWebPEncodeOptions()
			opts.Quality = 75
			opts.Pass = p
			libOutput := encodeWithLibrary(t, inputFile, opts)

			// 比較
			compareOutputs(t, testName, cmdOutput, libOutput)
		})
	}
}

// オプション組み合わせテスト
func TestCompat_WebP_OptionCombinations(t *testing.T) {
	setupCompatTest(t)

	inputFile := filepath.Join(testdataDir, "source/sizes/medium-512x512.png")

	testCases := []struct {
		name     string
		args     []string
		opts     func() WebPEncodeOptions
	}{
		{
			name: "q90-m6",
			args: []string{"-q", "90", "-m", "6"},
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Quality = 90
				o.Method = 6
				return o
			},
		},
		{
			name: "q75-m4-pass10",
			args: []string{"-q", "75", "-m", "4", "-pass", "10"},
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Quality = 75
				o.Method = 4
				o.Pass = 10
				return o
			},
		},
		{
			name: "lossless-m4",
			args: []string{"-lossless", "-m", "4"},
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Lossless = true
				o.Method = 4
				return o
			},
		},
		{
			name: "q80-alpha_q50",
			args: []string{"-q", "80", "-alpha_q", "50"},
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Quality = 80
				o.AlphaQuality = 50
				return o
			},
		},
		{
			name: "q85-m6-pass5",
			args: []string{"-q", "85", "-m", "6", "-pass", "5"},
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Quality = 85
				o.Method = 6
				o.Pass = 5
				return o
			},
		},
		{
			name: "q90-m6-alpha_q100",
			args: []string{"-q", "90", "-m", "6", "-alpha_q", "100"},
			opts: func() WebPEncodeOptions {
				o := DefaultWebPEncodeOptions()
				o.Quality = 90
				o.Method = 6
				o.AlphaQuality = 100
				return o
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// cwebp実行
			cmdOutput := runCWebP(t, inputFile,
				tc.args,
				filepath.Join(tempDir, "cmd-output", tc.name+".webp"))

			// ライブラリ実行
			libOutput := encodeWithLibrary(t, inputFile, tc.opts())

			// 比較
			compareOutputs(t, tc.name, cmdOutput, libOutput)
		})
	}
}
