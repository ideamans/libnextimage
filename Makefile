.PHONY: help build-c test-c install-c clean-c test-go test-all clean-all

# Default target
help:
	@echo "libnextimage Build and Test Targets"
	@echo ""
	@echo "C Library:"
	@echo "  make build-c      - Build C library (static .a and shared .so/.dylib/.dll)"
	@echo "  make test-c       - Run C tests"
	@echo "  make install-c    - Build and install C libraries to lib/"
	@echo "  make clean-c      - Clean C build artifacts"
	@echo ""
	@echo "Go Package:"
	@echo "  make test-go      - Run Go tests"
	@echo ""
	@echo "Combined:"
	@echo "  make test-all     - Run both C and Go tests"
	@echo "  make clean-all    - Clean all build artifacts"
	@echo ""

# C Library targets
build-c:
	@echo "Building C library (static and shared)..."
	@mkdir -p c/build
	@cd c/build && cmake .. && $(MAKE) nextimage nextimage_shared

test-c: build-c
	@echo "Running C tests..."
	@cd c/build && $(MAKE) basic_test simple_test command_interface_test decoder_test header_test
	@echo "Running available tests..."
	@cd c/build && ./basic_test && ./simple_test && ./command_interface_test && ./decoder_test && ./header_test
	@echo "C tests completed (note: some test programs may have compilation issues)"

install-c:
	@echo "Building and installing C library (static and shared)..."
	@mkdir -p c/build
	@cd c/build && cmake .. && $(MAKE) nextimage nextimage_shared && $(MAKE) install
	@echo "Installed to lib/"
	@ls -lh lib/*/libnextimage.a
	@echo ""
	@echo "Shared libraries:"
	@find lib -name "*.so" -o -name "*.dylib" -o -name "*.dll" | xargs ls -lh 2>/dev/null || echo "No shared libraries found"

clean-c:
	@echo "Cleaning C build artifacts..."
	@rm -rf c/build c/build-debug c/build-asan

# Go Package targets
test-go:
	@echo "Running Go tests..."
	@cd golang && go test -v -timeout 30s

# Combined targets
test-all: test-c test-go
	@echo ""
	@echo "All tests completed successfully!"

clean-all: clean-c
	@echo "Cleaning Go cache..."
	@cd golang && go clean -testcache
	@echo "All artifacts cleaned"
