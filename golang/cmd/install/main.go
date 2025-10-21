package main

import (
	"flag"
	"fmt"
	"os"

	libnextimage "github.com/ideamans/libnextimage/golang"
)

func main() {
	version := flag.String("version", libnextimage.LibraryVersion, "Version to install (default: current version)")
	force := flag.Bool("force", false, "Force download even if library exists")
	list := flag.Bool("list", false, "List available platforms")
	flag.Parse()

	if *list {
		fmt.Println("Available platforms:")
		for _, platform := range libnextimage.LibraryPlatforms {
			fmt.Printf("  - %s\n", platform)
		}
		return
	}

	fmt.Printf("libnextimage installer v%s\n", libnextimage.LibraryVersion)
	fmt.Println()

	// Check if library already exists
	if !*force && libnextimage.CheckLibraryExists() {
		fmt.Println("Pre-built library already exists.")
		fmt.Println("Use -force flag to re-download.")
		return
	}

	// Download the library
	fmt.Printf("Installing libnextimage v%s...\n", *version)
	if err := libnextimage.DownloadLibrary(*version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Installation completed successfully!")
}
