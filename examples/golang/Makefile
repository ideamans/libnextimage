.PHONY: all install clean test test-webp test-avif test-batch help

# Default target
all: install test

# Install dependencies and setup native libraries
install:
	@echo "Installing Go dependencies..."
	@# For local testing, uncomment the replace directive in go.mod
	@if ! grep -q "^replace github.com/ideamans/libnextimage/golang" go.mod; then \
		echo "Enabling local development mode..."; \
		sed -i.bak 's|^// replace github.com/ideamans/libnextimage/golang|replace github.com/ideamans/libnextimage/golang|' go.mod && rm -f go.mod.bak; \
	fi
	@# Download dependencies
	@go mod tidy
	@echo "✓ Installation complete"

# Run all tests
test: test-webp test-avif test-batch
	@echo ""
	@echo "✓ All Go example tests passed!"

# Run JPEG to WebP conversion test
test-webp:
	@echo ""
	@echo "Running JPEG to WebP conversion test..."
	@mkdir -p output
	@if [ ! -f ../../testdata/jpeg-source/landscape-like.jpg ]; then \
		echo "Error: Test image not found"; \
		exit 1; \
	fi
	@go run jpeg_to_webp.go ../../testdata/jpeg-source/landscape-like.jpg output/test.webp 80
	@if [ -f output/test.webp ]; then \
		echo "✓ WebP conversion test passed"; \
	else \
		echo "✗ WebP conversion test failed"; \
		exit 1; \
	fi

# Run JPEG to AVIF conversion test
test-avif:
	@echo ""
	@echo "Running JPEG to AVIF conversion test..."
	@mkdir -p output
	@if [ ! -f ../../testdata/jpeg-source/landscape-like.jpg ]; then \
		echo "Error: Test image not found"; \
		exit 1; \
	fi
	@go run jpeg_to_avif.go ../../testdata/jpeg-source/landscape-like.jpg output/test.avif 60 6
	@if [ -f output/test.avif ]; then \
		echo "✓ AVIF conversion test passed"; \
	else \
		echo "✗ AVIF conversion test failed"; \
		exit 1; \
	fi

# Run batch processing test
test-batch:
	@echo ""
	@echo "Running batch processing test..."
	@mkdir -p output/batch
	@if [ ! -d ../../testdata/jpeg-source ]; then \
		echo "Error: Test images directory not found"; \
		exit 1; \
	fi
	@go run batch_convert.go ../../testdata/jpeg-source output/batch webp
	@if [ -f output/batch/*.webp ]; then \
		echo "✓ Batch processing test passed"; \
	else \
		echo "✗ Batch processing test failed"; \
		exit 1; \
	fi

# Clean up
clean:
	@echo "Cleaning up..."
	@# Restore go.mod to use published version
	@if grep -q "^replace github.com/ideamans/libnextimage/golang" go.mod; then \
		sed -i.bak 's|^replace github.com/ideamans/libnextimage/golang|// replace github.com/ideamans/libnextimage/golang|' go.mod && rm -f go.mod.bak; \
	fi
	@rm -rf output
	@rm -f go.sum
	@echo "✓ Cleanup complete"

# Help
help:
	@echo "Available targets:"
	@echo "  all         - Install dependencies and run all tests (default)"
	@echo "  install     - Install Go dependencies and setup native libraries"
	@echo "  test        - Run all tests"
	@echo "  test-webp   - Run JPEG to WebP conversion test"
	@echo "  test-avif   - Run JPEG to AVIF conversion test"
	@echo "  test-batch  - Run batch processing test"
	@echo "  clean       - Remove outputs and restore go.mod"
	@echo "  help        - Show this help message"
