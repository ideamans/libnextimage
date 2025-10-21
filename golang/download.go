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

// CheckLibraryExists is a public wrapper for checking if the library exists
func CheckLibraryExists() bool {
	return checkLibraryExists()
}

// downloadLibrary downloads the pre-built library from GitHub Releases
func downloadLibrary() error {
	return downloadLibraryVersion(LibraryVersion)
}

// getPreviousVersion returns a reasonable fallback version
// This helps during the window when a new version is tagged but binaries aren't ready
func getPreviousVersion(currentVersion string) string {
	// Simple heuristic: if current is 0.X.0, try 0.X-1.0
	// This is a basic implementation - could be made more sophisticated
	if currentVersion == "0.3.0" {
		return "0.2.0"
	}
	if currentVersion == "0.4.0" {
		return "0.3.0"
	}
	// Add more mappings as needed, or implement proper semver parsing
	return ""
}

// downloadLibraryVersion downloads a specific version of the pre-built library
func downloadLibraryVersion(version string) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get caller information")
	}
	packageDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(packageDir)

	platform := getPlatform()
	archiveName := fmt.Sprintf("libnextimage-v%s-%s.tar.gz", version, platform)
	url := fmt.Sprintf("%s/v%s/%s", baseURL, version, archiveName)

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

// DownloadLibrary is a public wrapper for downloading the library with the default version
func DownloadLibrary(version string) error {
	if version == "" {
		return downloadLibrary()
	}
	return downloadLibraryVersion(version)
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
		fmt.Fprintf(os.Stderr, "Warning: Failed to download pre-built library v%s: %v\n", LibraryVersion, err)

		// Try to fall back to previous version (handles timing issues during release)
		previousVersion := getPreviousVersion(LibraryVersion)
		if previousVersion != "" {
			fmt.Fprintf(os.Stderr, "\nAttempting to download previous stable version v%s as fallback...\n", previousVersion)
			if fallbackErr := downloadLibraryVersion(previousVersion); fallbackErr == nil {
				fmt.Printf("Successfully downloaded fallback version v%s\n", previousVersion)
				fmt.Printf("Note: Using v%s library binaries with v%s code.\n", previousVersion, LibraryVersion)
				fmt.Printf("The library is backwards compatible. Update to v%s binaries when available.\n", LibraryVersion)
				return
			} else {
				fmt.Fprintf(os.Stderr, "Fallback download also failed: %v\n", fallbackErr)
			}
		}

		// If the exact version is not available, this might be a timing issue
		// where the code was released but binaries are still building
		fmt.Fprintf(os.Stderr, "\nPossible solutions:\n")
		fmt.Fprintf(os.Stderr, "1. Wait a few minutes for CI to build the release binaries\n")
		fmt.Fprintf(os.Stderr, "2. Use a previous version temporarily:\n")
		fmt.Fprintf(os.Stderr, "     go get github.com/ideamans/libnextimage/golang@v0.2.0\n")
		fmt.Fprintf(os.Stderr, "3. Build the C library manually:\n")
		fmt.Fprintf(os.Stderr, "     git clone --recursive https://github.com/ideamans/libnextimage.git\n")
		fmt.Fprintf(os.Stderr, "     cd libnextimage && bash scripts/build-c-library.sh\n")
		fmt.Fprintf(os.Stderr, "4. Check available releases at:\n")
		fmt.Fprintf(os.Stderr, "     https://github.com/ideamans/libnextimage/releases\n")
		// Don't panic here - let the build fail with a more helpful error message
	}
}
