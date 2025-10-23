package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ideamans/libnextimage/golang/avifenc"
	"github.com/ideamans/libnextimage/golang/cwebp"
)

func main() {
	// Parse command line arguments
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run batch_convert.go <input_dir> <output_dir> <format>")
		fmt.Println("\nFormats:")
		fmt.Println("  webp - Convert to WebP (quality: 80)")
		fmt.Println("  avif - Convert to AVIF (quality: 60)")
		fmt.Println("\nExample:")
		fmt.Println("  go run batch_convert.go ./photos ./output webp")
		os.Exit(1)
	}

	inputDir := os.Args[1]
	outputDir := os.Args[2]
	format := strings.ToLower(os.Args[3])

	if format != "webp" && format != "avif" {
		fmt.Fprintf(os.Stderr, "Invalid format. Use 'webp' or 'avif'\n")
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Find all JPEG files
	var jpegFiles []string
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" {
				jpegFiles = append(jpegFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(jpegFiles) == 0 {
		fmt.Println("No JPEG files found in", inputDir)
		os.Exit(0)
	}

	fmt.Printf("Found %d JPEG files\n", len(jpegFiles))
	fmt.Printf("Converting to %s...\n\n", strings.ToUpper(format))

	// Process files concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 4) // Limit to 4 concurrent conversions
	results := make(chan ConversionResult, len(jpegFiles))

	for i, inputPath := range jpegFiles {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			result := convertFile(idx+1, len(jpegFiles), path, outputDir, format)
			results <- result
		}(i, inputPath)
	}

	// Wait for all conversions to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var totalInput, totalOutput int64
	successCount := 0
	failCount := 0

	for result := range results {
		if result.Error != nil {
			failCount++
			fmt.Printf("❌ [%d/%d] %s: %v\n", result.Index, result.Total, result.InputPath, result.Error)
		} else {
			successCount++
			totalInput += result.InputSize
			totalOutput += result.OutputSize
			compression := (1.0 - float64(result.OutputSize)/float64(result.InputSize)) * 100
			fmt.Printf("✅ [%d/%d] %s → %s (%.1f%% smaller)\n",
				result.Index, result.Total,
				filepath.Base(result.InputPath),
				filepath.Base(result.OutputPath),
				compression)
		}
	}

	// Summary
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("Conversion complete!\n")
	fmt.Printf("  Success: %d files\n", successCount)
	fmt.Printf("  Failed: %d files\n", failCount)
	if successCount > 0 {
		overallCompression := (1.0 - float64(totalOutput)/float64(totalInput)) * 100
		fmt.Printf("  Total size: %.2f MB → %.2f MB\n",
			float64(totalInput)/(1024*1024),
			float64(totalOutput)/(1024*1024))
		fmt.Printf("  Overall compression: %.1f%%\n", overallCompression)
	}
	fmt.Println(strings.Repeat("=", 60))
}

type ConversionResult struct {
	Index      int
	Total      int
	InputPath  string
	OutputPath string
	InputSize  int64
	OutputSize int64
	Error      error
}

func convertFile(index, total int, inputPath, outputDir, format string) ConversionResult {
	result := ConversionResult{
		Index:     index,
		Total:     total,
		InputPath: inputPath,
	}

	// Read input file
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		result.Error = err
		return result
	}
	result.InputSize = int64(len(inputData))

	// Determine output path
	baseName := filepath.Base(inputPath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outputPath := filepath.Join(outputDir, nameWithoutExt+"."+format)
	result.OutputPath = outputPath

	// Convert based on format
	var outputData []byte
	switch format {
	case "webp":
		opts := cwebp.NewDefaultOptions()
		opts.Quality = 80
		cmd, cmdErr := cwebp.NewCommand(&opts)
		if cmdErr != nil {
			result.Error = cmdErr
			return result
		}
		defer cmd.Close()
		outputData, err = cmd.Run(inputData)
	case "avif":
		opts := avifenc.NewDefaultOptions()
		opts.Quality = 60
		opts.Speed = 6
		cmd, cmdErr := avifenc.NewCommand(&opts)
		if cmdErr != nil {
			result.Error = cmdErr
			return result
		}
		defer cmd.Close()
		outputData, err = cmd.Run(inputData)
	}

	if err != nil {
		result.Error = err
		return result
	}
	result.OutputSize = int64(len(outputData))

	// Write output file
	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		result.Error = err
		return result
	}

	return result
}
