package cwebp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMetadataFlags(t *testing.T) {
	// Test individual flags
	tests := []struct {
		name  string
		flags int
		desc  string
	}{
		{"None", MetadataNone, "No metadata"},
		{"EXIF only", MetadataEXIF, "EXIF only"},
		{"ICC only", MetadataICC, "ICC only"},
		{"XMP only", MetadataXMP, "XMP only"},
		{"EXIF and ICC", MetadataEXIF | MetadataICC, "EXIF and ICC"},
		{"EXIF and XMP", MetadataEXIF | MetadataXMP, "EXIF and XMP"},
		{"ICC and XMP", MetadataICC | MetadataXMP, "ICC and XMP"},
		{"All", MetadataAll, "All metadata"},
	}

	// Read test PNG file
	pngPath := filepath.Join("..", "..", "testdata", "png", "red.png")
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewDefaultOptions()
			opts.Quality = 80
			opts.KeepMetadata = tt.flags

			cmd, err := NewCommand(&opts)
			if err != nil {
				t.Fatalf("Failed to create command: %v", err)
			}
			defer cmd.Close()

			webpData, err := cmd.Run(pngData)
			if err != nil {
				t.Fatalf("Failed to encode PNG to WebP: %v", err)
			}

			if len(webpData) == 0 {
				t.Error("Encoded WebP data is empty")
			}

			t.Logf("%s: PNG (%d bytes) -> WebP (%d bytes, metadata=%d)",
				tt.desc, len(pngData), len(webpData), tt.flags)
		})
	}
}

func TestMetadataFlagCombinations(t *testing.T) {
	// Verify flag values
	if MetadataNone != 0 {
		t.Errorf("MetadataNone should be 0, got %d", MetadataNone)
	}
	if MetadataEXIF != 1 {
		t.Errorf("MetadataEXIF should be 1, got %d", MetadataEXIF)
	}
	if MetadataICC != 2 {
		t.Errorf("MetadataICC should be 2, got %d", MetadataICC)
	}
	if MetadataXMP != 4 {
		t.Errorf("MetadataXMP should be 4, got %d", MetadataXMP)
	}
	if MetadataAll != 7 {
		t.Errorf("MetadataAll should be 7, got %d", MetadataAll)
	}

	// Verify bitwise combinations
	if (MetadataEXIF | MetadataICC) != 3 {
		t.Errorf("EXIF | ICC should be 3, got %d", MetadataEXIF|MetadataICC)
	}
	if (MetadataEXIF | MetadataXMP) != 5 {
		t.Errorf("EXIF | XMP should be 5, got %d", MetadataEXIF|MetadataXMP)
	}
	if (MetadataICC | MetadataXMP) != 6 {
		t.Errorf("ICC | XMP should be 6, got %d", MetadataICC|MetadataXMP)
	}
	if (MetadataEXIF | MetadataICC | MetadataXMP) != MetadataAll {
		t.Errorf("EXIF | ICC | XMP should equal MetadataAll")
	}

	t.Logf("All metadata flag combinations are correct")
}

func TestMetadataDefaultValue(t *testing.T) {
	opts := NewDefaultOptions()

	// Default should be -1 (use cwebp's default behavior) or a reasonable default
	t.Logf("Default KeepMetadata value: %d", opts.KeepMetadata)

	// Ensure it's a valid value (-1 means use cwebp default, 0-7 are explicit values)
	if opts.KeepMetadata < -1 || opts.KeepMetadata > MetadataAll {
		t.Errorf("Invalid default KeepMetadata value: %d (should be -1 to 7)", opts.KeepMetadata)
	}
}
