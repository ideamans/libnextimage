//! High-performance WebP and AVIF image encoding/decoding library.
//!
//! This library provides Rust bindings for libnextimage, offering efficient
//! WebP and AVIF image processing with a simple, safe API.
//!
//! # Features
//!
//! - **WebP encoding/decoding** with full option support
//! - **AVIF encoding/decoding** with AOM codec
//! - **GIF to WebP conversion** for animated images
//! - **WebP to GIF conversion**
//! - **Automatic library download** (optional)
//! - **Cross-platform support** (macOS, Linux, Windows)
//!
//! # Quick Start
//!
//! ```no_run
//! use nextimage::{webp, avif};
//!
//! // Ensure the native library is available
//! nextimage::ensure_library()?;
//!
//! // Encode PNG to WebP
//! let png_data = std::fs::read("input.png")?;
//! let webp_data = webp::encode(&png_data)?;
//! std::fs::write("output.webp", webp_data)?;
//!
//! // Encode PNG to AVIF
//! let avif_data = avif::encode(&png_data)?;
//! std::fs::write("output.avif", avif_data)?;
//! # Ok::<(), Box<dyn std::error::Error>>(())
//! ```
//!
//! # Encoder/Decoder Reuse
//!
//! For better performance when processing multiple images, reuse encoder/decoder instances:
//!
//! ```no_run
//! use nextimage::webp::{WebPEncoder, WebPEncodeOptions};
//!
//! let encoder = WebPEncoder::new(WebPEncodeOptions::lossy(85.0))?;
//!
//! for path in &["image1.png", "image2.png", "image3.png"] {
//!     let input = std::fs::read(path)?;
//!     let webp = encoder.encode(&input)?;
//!     std::fs::write(format!("{}.webp", path), webp)?;
//! }
//! # Ok::<(), Box<dyn std::error::Error>>(())
//! ```
//!
//! # Library Installation
//!
//! The native library can be installed in several ways:
//!
//! 1. **Automatic download** (recommended): Call [`ensure_library()`] at startup
//! 2. **Environment variable**: Set `LIBNEXTIMAGE_LIB_DIR` to the library path
//! 3. **Manual installation**: Place `libnextimage.a` in the cache directory
//!
//! The library is cached at:
//! - `$LIBNEXTIMAGE_CACHE_DIR` if set
//! - `$XDG_CACHE_HOME/libnextimage` if set
//! - `~/.cache/libnextimage` on Unix
//! - `%LOCALAPPDATA%/libnextimage` on Windows

#![warn(missing_docs)]
#![warn(rust_2018_idioms)]

mod error;
mod ffi;

pub mod avif;
pub mod webp;

pub use error::{Error, Result};

use once_cell::sync::OnceCell;
use std::path::PathBuf;
use std::sync::Mutex;

/// Library version
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

/// Supported platforms for pre-built libraries
pub const PLATFORMS: &[&str] = &[
    "darwin-arm64",
    "darwin-amd64",
    "linux-amd64",
    "linux-arm64",
    "windows-amd64",
];

static LIBRARY_ENSURED: OnceCell<std::result::Result<(), String>> = OnceCell::new();
static DOWNLOAD_VERBOSE: Mutex<bool> = Mutex::new(false);

/// Set whether to show download progress messages.
///
/// By default, download messages are suppressed. Set to `true` to see download progress.
/// This must be called before [`ensure_library()`] to take effect.
pub fn set_download_verbose(verbose: bool) {
    if let Ok(mut v) = DOWNLOAD_VERBOSE.lock() {
        *v = verbose;
    }
}

#[cfg(feature = "auto-download")]
fn is_verbose() -> bool {
    std::env::var("LIBNEXTIMAGE_VERBOSE").is_ok()
        || DOWNLOAD_VERBOSE.lock().map(|v| *v).unwrap_or(false)
}

/// Get the current platform string (e.g., "darwin-arm64").
pub fn get_platform() -> String {
    let os = std::env::consts::OS;
    let arch = match std::env::consts::ARCH {
        "x86_64" => "amd64",
        "aarch64" => "arm64",
        arch => arch,
    };

    // Map OS names to libnextimage convention
    let os = match os {
        "macos" => "darwin",
        os => os,
    };

    format!("{}-{}", os, arch)
}

/// Get the cache directory for downloaded libraries.
pub fn get_cache_dir() -> Option<PathBuf> {
    // Check LIBNEXTIMAGE_CACHE_DIR first
    if let Ok(cache_dir) = std::env::var("LIBNEXTIMAGE_CACHE_DIR") {
        return Some(PathBuf::from(cache_dir));
    }

    // Check XDG_CACHE_HOME
    if let Ok(xdg_cache) = std::env::var("XDG_CACHE_HOME") {
        return Some(PathBuf::from(xdg_cache).join("libnextimage"));
    }

    // Fall back to platform-specific cache directory
    dirs::cache_dir().map(|d| d.join("libnextimage"))
}

/// Get the expected library path.
fn get_library_path(base: &PathBuf) -> PathBuf {
    let platform = get_platform();
    base.join("lib").join(&platform).join("libnextimage.a")
}

/// Check if the library exists at the given base path.
fn library_exists(base: &PathBuf) -> bool {
    get_library_path(base).exists()
}

/// Ensure the native library is available.
///
/// This function checks for the library in the following locations:
/// 1. `LIBNEXTIMAGE_LIB_DIR` environment variable
/// 2. Cache directory (downloads if not present)
/// 3. Relative to the crate (for development)
///
/// If the library is not found and the `auto-download` feature is enabled,
/// it will be downloaded from GitHub Releases.
///
/// # Errors
///
/// Returns an error if the library cannot be found or downloaded.
///
/// # Example
///
/// ```no_run
/// fn main() -> nextimage::Result<()> {
///     nextimage::ensure_library()?;
///
///     // Now you can use the library
///     let encoder = nextimage::webp::WebPEncoder::new(Default::default())?;
///     Ok(())
/// }
/// ```
pub fn ensure_library() -> Result<()> {
    let result = LIBRARY_ENSURED.get_or_init(|| {
        // Check LIBNEXTIMAGE_LIB_DIR
        if let Ok(lib_dir) = std::env::var("LIBNEXTIMAGE_LIB_DIR") {
            let path = PathBuf::from(&lib_dir);
            if path.join("libnextimage.a").exists() {
                return Ok(());
            }
        }

        // Check cache directory
        if let Some(cache_dir) = get_cache_dir() {
            if library_exists(&cache_dir) {
                return Ok(());
            }

            // Try to download
            #[cfg(feature = "auto-download")]
            {
                if let Err(e) = download_library(&cache_dir) {
                    return Err(format!("failed to download library: {}", e));
                }
                return Ok(());
            }
        }

        Err("library not found and auto-download is disabled".to_string())
    });

    result.clone().map_err(Error::LibraryNotAvailable)
}

/// Download the native library from GitHub Releases.
///
/// The release asset `libnextimage-v<version>-<platform>.tar.gz` contains `lib/libnextimage.a`
/// and `include/`. They are placed in `<cache>/lib/<platform>/`, where `build.rs` looks for them.
#[cfg(feature = "auto-download")]
fn download_library(cache_dir: &PathBuf) -> Result<()> {
    use flate2::read::GzDecoder;
    use std::io::Read;
    use tar::Archive;

    let platform = get_platform();
    let url = format!(
        "https://github.com/ideamans/libnextimage/releases/download/v{0}/libnextimage-v{0}-{1}.tar.gz",
        VERSION, platform
    );

    if is_verbose() {
        eprintln!("Downloading libnextimage from {}...", url);
    }

    // Download
    let response = ureq::get(&url)
        .call()
        .map_err(|e| Error::DownloadFailed(format!("{}: {}", url, e)))?;

    let mut reader = response.into_reader();
    let mut data = Vec::new();
    reader.read_to_end(&mut data)?;

    // Extract into a temporary directory, then move the library into lib/<platform>/
    let dest = get_library_path(cache_dir)
        .parent()
        .map(|p| p.to_path_buf())
        .ok_or_else(|| Error::DownloadFailed("invalid cache directory".to_string()))?;
    let tmp = dest.with_extension("download");
    let _ = std::fs::remove_dir_all(&tmp);
    std::fs::create_dir_all(&tmp)?;
    Archive::new(GzDecoder::new(&data[..])).unpack(&tmp)?;

    let lib = tmp.join("lib").join("libnextimage.a");
    if !lib.exists() {
        let _ = std::fs::remove_dir_all(&tmp);
        return Err(Error::DownloadFailed(format!(
            "{} does not contain lib/libnextimage.a",
            url
        )));
    }
    std::fs::create_dir_all(&dest)?;
    std::fs::copy(&lib, dest.join("libnextimage.a"))?;
    let include = tmp.join("include");
    if include.exists() {
        let _ = std::fs::remove_dir_all(dest.join("include"));
        std::fs::rename(&include, dest.join("include"))?;
    }
    let _ = std::fs::remove_dir_all(&tmp);

    if is_verbose() {
        eprintln!("Library installed to {:?}", dest);
    }

    Ok(())
}

/// Get the library version string from the native library.
pub fn version() -> Option<String> {
    unsafe {
        let ptr = ffi::nextimage_version();
        if ptr.is_null() {
            None
        } else {
            Some(std::ffi::CStr::from_ptr(ptr).to_string_lossy().into_owned())
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_platform() {
        let platform = get_platform();
        assert!(PLATFORMS.iter().any(|p| platform.contains(p.split('-').next().unwrap())));
    }

    #[test]
    fn test_get_cache_dir() {
        let cache_dir = get_cache_dir();
        assert!(cache_dir.is_some());
    }
}
