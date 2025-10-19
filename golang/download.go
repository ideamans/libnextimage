package libnextimage

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	githubUser = "ideamans"
	githubRepo = "libnextimage"
	baseURL    = "https://github.com/" + githubUser + "/" + githubRepo + "/releases/download"
)

// getPlatform returns the current platform string (e.g., "darwin-arm64")
func getPlatform() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

// getLibraryPath returns the expected path to the combined library
func getLibraryPath() string {
	platform := getPlatform()
	return filepath.Join("lib", platform, "libnextimage.a")
}

// checkLibraryExists checks if the combined library exists
func checkLibraryExists() bool {
	// Get the directory where this Go file is located
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return false
	}
	packageDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(packageDir)
	
	libPath := filepath.Join(projectRoot, getLibraryPath())
	_, err := os.Stat(libPath)
	return err == nil
}

// downloadLibrary downloads the pre-built library from GitHub Releases
func downloadLibrary() error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get caller information")
	}
	packageDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(packageDir)
	
	platform := getPlatform()
	archiveName := fmt.Sprintf("libnextimage-v%s-%s.tar.gz", LibraryVersion, platform)
	url := fmt.Sprintf("%s/v%s/%s", baseURL, LibraryVersion, archiveName)
	
	fmt.Printf("Downloading libnextimage library for %s...\n", platform)
	fmt.Printf("URL: %s\n", url)
	
	// Download the tar.gz archive
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download library: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download library: HTTP %d", resp.StatusCode)
	}
	
	// Extract the tar.gz archive
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()
	
	tr := tar.NewReader(gzr)
	
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}
		
		target := filepath.Join(projectRoot, header.Name)
		
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			f.Close()
		}
	}
	
	fmt.Printf("Successfully downloaded and extracted library to %s\n", projectRoot)
	return nil
}

// init is called automatically when the package is imported
// It checks if the library exists and downloads it if necessary
func init() {
	// Skip download check in certain cases:
	// 1. When running go mod download
	// 2. When the library already exists
	// 3. When building from source (lib/ directory exists in git)
	
	if checkLibraryExists() {
		// Library already exists, no need to download
		return
	}
	
	// Library doesn't exist - this might be a go get scenario
	// Try to download from GitHub Releases
	fmt.Println("Pre-built library not found. Attempting to download from GitHub Releases...")
	
	if err := downloadLibrary(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to download pre-built library: %v\n", err)
		fmt.Fprintf(os.Stderr, "You may need to build the C library manually using:\n")
		fmt.Fprintf(os.Stderr, "  bash scripts/build-c-library.sh\n")
		// Don't panic here - let the build fail with a more helpful error message
	}
}
