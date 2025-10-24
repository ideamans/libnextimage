package libnextimage

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	githubUser = "ideamans"
	githubRepo = "libnextimage"
	baseURL    = "https://github.com/" + githubUser + "/" + githubRepo + "/releases/download"

	// Default download timeout
	defaultDownloadTimeout = 5 * time.Minute
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

// getCacheDir returns a writable cache directory for downloaded libraries
// Priority: LIBNEXTIMAGE_CACHE_DIR > XDG_CACHE_HOME/libnextimage > ~/.cache/libnextimage
func getCacheDir() (string, error) {
	// Check environment variable first
	if cacheDir := os.Getenv("LIBNEXTIMAGE_CACHE_DIR"); cacheDir != "" {
		return cacheDir, nil
	}

	// Try XDG_CACHE_HOME
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "libnextimage"), nil
	}

	// Fall back to user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".cache", "libnextimage"), nil
}

// getProjectRoot returns the project root directory
func getProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	packageDir := filepath.Dir(filename)
	return filepath.Dir(packageDir), nil
}

// checkLibraryExists checks if the combined library exists in the project or cache
func checkLibraryExists() bool {
	// Check project directory first
	projectRoot, err := getProjectRoot()
	if err == nil {
		libPath := filepath.Join(projectRoot, getLibraryPath())
		if _, err := os.Stat(libPath); err == nil {
			return true
		}
	}

	// Check cache directory
	cacheDir, err := getCacheDir()
	if err == nil {
		libPath := filepath.Join(cacheDir, getLibraryPath())
		if _, err := os.Stat(libPath); err == nil {
			return true
		}
	}

	return false
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
	// Try to download to cache directory first
	cacheDir, err := getCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}

	// Try cache directory first
	targetDir := cacheDir
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		// If cache directory creation fails, try project directory
		projectRoot, projErr := getProjectRoot()
		if projErr != nil {
			return fmt.Errorf("failed to get target directory: cache=%w, project=%w", err, projErr)
		}
		targetDir = projectRoot
		fmt.Fprintf(os.Stderr, "Warning: Could not create cache directory, using project directory: %v\n", err)
	}

	platform := getPlatform()
	archiveName := fmt.Sprintf("libnextimage-v%s-%s.tar.gz", version, platform)
	url := fmt.Sprintf("%s/v%s/%s", baseURL, version, archiveName)

	fmt.Printf("Downloading libnextimage library for %s...\n", platform)
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Target: %s\n", targetDir)

	// Create HTTP client with timeout and context
	ctx, cancel := context.WithTimeout(context.Background(), defaultDownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: defaultDownloadTimeout,
	}

	// Download the tar.gz archive
	resp, err := client.Do(req)
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

		target := filepath.Join(targetDir, header.Name)

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

	fmt.Printf("Successfully downloaded and extracted library to %s\n", targetDir)
	return nil
}

// DownloadLibrary is a public wrapper for downloading the library with the default version
func DownloadLibrary(version string) error {
	if version == "" {
		return downloadLibrary()
	}
	return downloadLibraryVersion(version)
}

// EnsureLibrary ensures the library is available, downloading if necessary.
// This is the recommended way to initialize the library.
// Returns an error if the library cannot be found or downloaded.
func EnsureLibrary() error {
	if checkLibraryExists() {
		return nil
	}

	fmt.Println("Pre-built library not found. Attempting to download from GitHub Releases...")

	if err := downloadLibrary(); err != nil {
		// Try to fall back to previous version (handles timing issues during release)
		previousVersion := getPreviousVersion(LibraryVersion)
		if previousVersion != "" {
			fmt.Fprintf(os.Stderr, "\nAttempting to download previous stable version v%s as fallback...\n", previousVersion)
			if fallbackErr := downloadLibraryVersion(previousVersion); fallbackErr == nil {
				fmt.Printf("Successfully downloaded fallback version v%s\n", previousVersion)
				fmt.Printf("Note: Using v%s library binaries with v%s code.\n", previousVersion, LibraryVersion)
				fmt.Printf("The library is backwards compatible. Update to v%s binaries when available.\n", LibraryVersion)
				return nil
			}
		}

		return fmt.Errorf("failed to download library v%s: %w", LibraryVersion, err)
	}

	return nil
}

// init is intentionally left empty to avoid any side effects on package import.
// Users must explicitly call EnsureLibrary() to download the library if needed.
func init() {
	// No automatic operations on import
}
