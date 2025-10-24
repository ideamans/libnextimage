.PHONY: help build-c test-c install-c clean-c test-go test-typescript test-all clean-all

# Default target
help:
	@echo "libnextimage Build and Test Targets"
	@echo ""
	@echo "C Library:"
	@echo "  make build-c        - Build C library (static .a and shared .so/.dylib/.dll)"
	@echo "  make test-c         - Run C tests"
	@echo "  make install-c      - Build and install C libraries to lib/"
	@echo "  make clean-c        - Clean C build artifacts"
	@echo ""
	@echo "Go Package:"
	@echo "  make test-go        - Run Go tests"
	@echo ""
	@echo "TypeScript Package:"
	@echo "  make test-typescript - Run TypeScript tests (includes environment setup)"
	@echo ""
	@echo "Combined:"
	@echo "  make test-all       - Run C, Go, and TypeScript tests"
	@echo "  make clean-all      - Clean all build artifacts"
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
	@echo ""
	@echo "Static library:"
	@ls -lh lib/static/libnextimage.a 2>/dev/null || echo "  (not found)"
	@echo ""
	@echo "Shared libraries:"
	@find lib/shared -name "*.so" -o -name "*.dylib" -o -name "*.dll" 2>/dev/null | xargs ls -lh 2>/dev/null || echo "  (not found)"
	@echo ""
	@echo "Header files:"
	@ls lib/include/*.h 2>/dev/null | wc -l | xargs echo "  " "files in lib/include/" || echo "  (not found)"

clean-c:
	@echo "Cleaning C build artifacts..."
	@rm -rf c/build c/build-debug c/build-asan

# Go Package targets
test-go:
	@echo "Running Go tests..."
	@cd golang && go test -v -timeout 30s

# TypeScript Package targets
test-typescript: install-c
	@echo "Running TypeScript tests..."
	@echo "Checking for Node.js..."
	@command -v node >/dev/null 2>&1 || { echo "Error: Node.js is not installed. Please install Node.js >= 18.0.0"; exit 1; }
	@echo "Node.js version: $$(node --version)"
	@echo ""
	@echo "Setting up TypeScript environment..."
	@cd typescript && npm install
	@echo ""
	@echo "Building TypeScript project..."
	@cd typescript && npm run build
	@echo ""
	@echo "Running TypeScript tests..."
	@cd typescript && npm test
	@echo ""
	@echo "TypeScript tests completed successfully!"

# Combined targets
test-all: test-c test-go test-typescript
	@echo ""
	@echo "All tests completed successfully!"

clean-all: clean-c
	@echo "Cleaning Go cache..."
	@cd golang && go clean -testcache
	@echo "Cleaning TypeScript build artifacts..."
	@rm -rf typescript/dist typescript/node_modules
	@echo "All artifacts cleaned"
