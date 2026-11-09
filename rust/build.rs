use std::env;
use std::path::PathBuf;

fn main() {
    let target_os = env::var("CARGO_CFG_TARGET_OS").unwrap();
    let target_arch = env::var("CARGO_CFG_TARGET_ARCH").unwrap();

    // Map Rust target arch to libnextimage platform names
    let arch = match target_arch.as_str() {
        "x86_64" => "amd64",
        "aarch64" => "arm64",
        _ => panic!("Unsupported architecture: {}", target_arch),
    };

    // Map OS names to libnextimage convention
    let os = match target_os.as_str() {
        "macos" => "darwin",
        os => os,
    };

    let platform = format!("{}-{}", os, arch);

    // Search paths for the library
    let mut search_paths = Vec::new();

    // 1. Check LIBNEXTIMAGE_LIB_DIR environment variable
    if let Ok(lib_dir) = env::var("LIBNEXTIMAGE_LIB_DIR") {
        search_paths.push(PathBuf::from(lib_dir));
    }

    // 2. Check cache directory
    if let Some(cache_dir) = get_cache_dir() {
        search_paths.push(cache_dir.join("lib").join(&platform));
    }

    // 3. Check relative to this crate (for development)
    let manifest_dir = PathBuf::from(env::var("CARGO_MANIFEST_DIR").unwrap());
    search_paths.push(manifest_dir.join("..").join("lib").join(&platform));
    search_paths.push(manifest_dir.join("lib").join(&platform));

    // Find the library
    #[allow(unused_mut)]
    let mut lib_path = search_paths
        .iter()
        .find(|path| path.join("libnextimage.a").exists())
        .cloned();

    // 4. Download the prebuilt library from GitHub Releases (auto-download feature).
    //    The static library is needed at link time, so this has to happen here, not at runtime.
    #[cfg(feature = "auto-download")]
    if lib_path.is_none() {
        if let Some(cache_dir) = get_cache_dir() {
            let dest = cache_dir.join("lib").join(&platform);
            match download_library(&platform, &dest) {
                Ok(()) => lib_path = Some(dest),
                Err(e) => println!("cargo:warning=failed to download libnextimage: {}", e),
            }
        }
    }

    if let Some(lib_dir) = lib_path {
        println!("cargo:rustc-link-search=native={}", lib_dir.display());
        println!("cargo:rustc-link-lib=static=nextimage");
    } else {
        // Library not found - print warning but don't fail
        // The library will be downloaded at runtime if auto-download feature is enabled
        println!(
            "cargo:warning=libnextimage.a not found. Searched paths: {:?}",
            search_paths
        );
        println!("cargo:warning=The library will need to be downloaded at runtime or installed manually.");
    }

    // Link system libraries required by libnextimage
    match target_os.as_str() {
        "macos" => {
            println!("cargo:rustc-link-lib=framework=CoreFoundation");
            println!("cargo:rustc-link-lib=framework=CoreGraphics");
            println!("cargo:rustc-link-lib=framework=ImageIO");
            println!("cargo:rustc-link-lib=c++");
            println!("cargo:rustc-link-lib=z");
        }
        "linux" => {
            println!("cargo:rustc-link-lib=stdc++");
            println!("cargo:rustc-link-lib=m");
            println!("cargo:rustc-link-lib=pthread");
            println!("cargo:rustc-link-lib=z");
        }
        "windows" => {
            println!("cargo:rustc-link-lib=user32");
            println!("cargo:rustc-link-lib=gdi32");
        }
        _ => {}
    }

    // Rerun if environment changes
    println!("cargo:rerun-if-env-changed=LIBNEXTIMAGE_LIB_DIR");
    println!("cargo:rerun-if-env-changed=LIBNEXTIMAGE_CACHE_DIR");
}

/// Release asset URL for the static library of this crate's version.
/// The asset is `libnextimage-v<version>-<platform>.tar.gz` and contains `lib/libnextimage.a` and `include/`.
#[cfg(feature = "auto-download")]
fn download_url(platform: &str) -> String {
    let version = env::var("CARGO_PKG_VERSION").unwrap();
    format!(
        "https://github.com/ideamans/libnextimage/releases/download/v{0}/libnextimage-v{0}-{1}.tar.gz",
        version, platform
    )
}

/// Download the release asset and place `libnextimage.a` (and the headers) in `dest`.
#[cfg(feature = "auto-download")]
fn download_library(platform: &str, dest: &std::path::Path) -> Result<(), String> {
    use std::io::Read;

    let url = download_url(platform);
    let response = ureq::get(&url)
        .call()
        .map_err(|e| format!("{}: {}", url, e))?;
    let mut data = Vec::new();
    response
        .into_reader()
        .read_to_end(&mut data)
        .map_err(|e| format!("{}: {}", url, e))?;

    // Unpack into a temporary directory next to dest, then move only what we need.
    let tmp = dest.with_extension("download");
    let _ = std::fs::remove_dir_all(&tmp);
    std::fs::create_dir_all(&tmp).map_err(|e| e.to_string())?;
    tar::Archive::new(flate2::read::GzDecoder::new(&data[..]))
        .unpack(&tmp)
        .map_err(|e| format!("unpack {}: {}", url, e))?;

    let lib = tmp.join("lib").join("libnextimage.a");
    if !lib.exists() {
        let _ = std::fs::remove_dir_all(&tmp);
        return Err(format!("{} does not contain lib/libnextimage.a", url));
    }
    std::fs::create_dir_all(dest).map_err(|e| e.to_string())?;
    std::fs::copy(&lib, dest.join("libnextimage.a")).map_err(|e| e.to_string())?;
    let include = tmp.join("include");
    if include.exists() {
        let _ = std::fs::remove_dir_all(dest.join("include"));
        std::fs::rename(&include, dest.join("include")).map_err(|e| e.to_string())?;
    }
    let _ = std::fs::remove_dir_all(&tmp);
    Ok(())
}

fn get_cache_dir() -> Option<PathBuf> {
    // Check LIBNEXTIMAGE_CACHE_DIR first
    if let Ok(cache_dir) = env::var("LIBNEXTIMAGE_CACHE_DIR") {
        return Some(PathBuf::from(cache_dir));
    }

    // Check XDG_CACHE_HOME
    if let Ok(xdg_cache) = env::var("XDG_CACHE_HOME") {
        return Some(PathBuf::from(xdg_cache).join("libnextimage"));
    }

    // Fall back to home directory
    dirs::cache_dir().map(|d| d.join("libnextimage"))
}
