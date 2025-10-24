package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
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

var libraryVersion = "0.3.0" // This should match the tag version

func main() {
	version := flag.String("version", libraryVersion, "Version to install (default: current version)")
	force := flag.Bool("force", false, "Force download even if library exists")
	targetDir := flag.String("dir", ".", "Target directory to install library (default: current directory)")
	list := flag.Bool("list", false, "List available platforms")
	flag.Parse()

	if *list {
		fmt.Println("Available platforms:")
		platforms := []string{"darwin-arm64", "darwin-amd64", "linux-amd64", "linux-arm64", "windows-amd64"}
		for _, platform := range platforms {
			fmt.Printf("  - %s\n", platform)
		}
		return
	}

	fmt.Printf("libnextimage installer v%s\n", libraryVersion)
	fmt.Println()

	// Get target directory
	absTargetDir, err := filepath.Abs(*targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check if library already exists in target directory
	platform := runtime.GOOS + "-" + runtime.GOARCH
	libPath := filepath.Join(absTargetDir, "lib", platform, "libnextimage.a")

	if !*force {
		if _, err := os.Stat(libPath); err == nil {
			fmt.Printf("Pre-built library already exists at: %s\n", libPath)
			fmt.Println("Use -force flag to re-download.")
			return
		}
	}

	// Download the library
	fmt.Printf("Installing libnextimage v%s to %s...\n", *version, absTargetDir)
	fmt.Printf("Platform: %s\n", platform)

	archiveName := fmt.Sprintf("libnextimage-v%s-%s.tar.gz", *version, platform)
	url := fmt.Sprintf("%s/v%s/%s", baseURL, *version, archiveName)

	fmt.Printf("Downloading from: %s\n", url)

	// Download
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "The version %s may not be available yet.\n", *version)
		fmt.Fprintf(os.Stderr, "Check available releases at: https://github.com/%s/%s/releases\n", githubUser, githubRepo)
		os.Exit(1)
	}

	// Extract
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		target := filepath.Join(absTargetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			f.Close()
		}
	}

	fmt.Println()
	fmt.Println("✅ Installation completed successfully!")
	fmt.Printf("Library installed to: %s\n", libPath)
	fmt.Println()
	fmt.Println("You can now use libnextimage in your Go project.")
}
