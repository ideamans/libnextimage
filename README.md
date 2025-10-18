# libnextimage

High-performance WebP and AVIF encoding/decoding library with FFI interface for Go.

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/ideamans/libnextimage)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0--alpha-orange)](DEPENDENCIES.txt)

## Overview

`libnextimage` provides direct FFI access to WebP and AVIF encoding/decoding functionality, eliminating the overhead of spawning separate processes for image conversion operations.

### Key Features

- **Zero Process Overhead**: Direct library calls instead of spawning cwebp/dwebp/avifenc/avifdec processes
- **Multi-plane Support**: Full support for YUV planar formats (4:2:0, 4:2:2, 4:4:4)
- **High Bit Depth**: Support for 8-bit, 10-bit, and 12-bit AVIF encoding
- **Memory Safe**: Comprehensive testing with AddressSanitizer and UndefinedBehaviorSanitizer
- **Thread Safe**: Thread-local error handling and concurrent encoding/decoding
- **Go Integration**: Idiomatic Go API with explicit functions for bytes, files, and streams

## Development Status

🚧 **Phase 1 Complete** - Foundation layer implemented

- ✅ Core FFI interface
- ✅ Memory management (alloc/into separation)
- ✅ Error handling (thread-local messages)
- ✅ CMake build system (normal/Debug/ASan/UBSan)
- ✅ Basic test suite

🔜 **Phase 2 Planned** - WebP integration (see [SPEC.md](SPEC.md))

## Quick Start

### Building from Source

#### Prerequisites

- CMake 3.15 or later
- C11-compatible compiler (GCC, Clang, or MSVC)
- Git (for submodule management)

#### Basic Build

```bash
# Clone the repository
git clone --recursive https://github.com/ideamans/libnextimage.git
cd libnextimage

# Build the C library
cd c
mkdir build && cd build
cmake ..
cmake --build .

# Run tests
ctest --output-on-failure
```

#### Debug Build (with Leak Counter)

```bash
cd c
mkdir build-debug && cd build-debug
cmake -DCMAKE_BUILD_TYPE=Debug ..
cmake --build .
ctest --output-on-failure
```

#### AddressSanitizer Build

```bash
cd c
mkdir build-asan && cd build-asan
cmake -DENABLE_ASAN=ON ..
cmake --build .
ctest --output-on-failure
```

#### UndefinedBehaviorSanitizer Build

```bash
cd c
mkdir build-ubsan && cd build-ubsan
cmake -DENABLE_UBSAN=ON ..
cmake --build .
ctest --output-on-failure
```

## Project Structure

```
libnextimage/
├── c/                        # C FFI layer
│   ├── include/              # Public headers
│   │   └── nextimage.h       # Main FFI interface
│   ├── src/                  # Implementation
│   │   ├── common.c          # Memory & error handling
│   │   └── internal.h        # Internal helpers
│   ├── test/                 # C tests
│   │   ├── basic_test.c
│   │   ├── leak_counter_test.c
│   │   └── sanitizer/
│   └── CMakeLists.txt
├── deps/                     # Dependencies (git submodules, TBD)
├── golang/                   # Go bindings (TBD)
├── lib/                      # Pre-compiled libraries (TBD)
├── SPEC.md                   # Detailed specification
├── DEPENDENCIES.txt          # Bill of Materials
└── LICENSE                   # MIT License
```

## Documentation

- [SPEC.md](SPEC.md) - Comprehensive specification and development plan
- [DEPENDENCIES.txt](DEPENDENCIES.txt) - Bill of Materials with all dependencies
- [CLAUDE.md](CLAUDE.md) - AI assistant context for development

## Testing

### Test Categories

1. **Basic Tests** - Core functionality verification
2. **Leak Counter Tests** - C heap leak detection (Debug build only)
3. **Sanitizer Tests** - Memory safety validation (ASan/UBSan builds)

### Running Tests

```bash
# All tests
ctest --output-on-failure

# Specific test
./basic_test

# Debug build with leak counter
cd build-debug && ./leak_counter_test

# ASan build
cd build-asan && ./sanitizer_test
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

This project is currently in active development (Phase 1). Contributions will be welcome once the core API is stabilized.

## Roadmap

- [x] **Phase 1**: Foundation (Complete)
  - Core FFI interface
  - Memory management
  - Error handling
  - Build system
- [ ] **Phase 2**: WebP Integration (Weeks 3-4)
  - libwebp integration
  - Go bindings for WebP
- [ ] **Phase 3**: AVIF Integration (Weeks 5-6)
  - libavif integration
  - Go bindings for AVIF
- [ ] **Phase 4**: New Features (Week 7)
  - webp2gif conversion
- [ ] **Phase 5**: Security & Fuzzing (Weeks 8-9)
- [ ] **Phase 6**: Optimization (Weeks 10-11)
- [ ] **Phase 7**: Release (Week 12)

See [SPEC.md](SPEC.md) for the complete development plan.

## Credits

Developed by [Ideamans Inc.](https://www.ideamans.com/)

Built on top of:
- [libwebp](https://github.com/webmproject/libwebp) (planned)
- [libavif](https://github.com/AOMediaCodec/libavif) (planned)
