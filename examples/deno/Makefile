.PHONY: all install clean test test-basic help

# Default target
all: install test

# Install/setup dependencies
install:
	@echo "Setting up Deno environment..."
	@# Build the TypeScript package first
	@cd ../../typescript && npm run build
	@# Verify Deno is available
	@which deno > /dev/null 2>&1 || (echo "Error: Deno is not installed. Please install it from https://deno.land/" && exit 1)
	@echo "✓ Deno environment ready"

# Run all tests
test: test-basic
	@echo ""
	@echo "✓ All Deno example tests passed!"

# Run basic encoding test
test-basic:
	@echo ""
	@echo "Running basic encoding test..."
	@deno task test:basic

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -f ideamans-libnextimage-*.tgz
	@rm -f libnextimage-*.tgz
	@rm -rf output
	@echo "✓ Cleanup complete"

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Setup environment and run all tests (default)"
	@echo "  install    - Setup Deno environment"
	@echo "  test       - Run all tests"
	@echo "  test-basic - Run basic encoding test"
	@echo "  clean      - Remove outputs"
	@echo "  help       - Show this help message"
	@echo ""
	@echo "Note: Deno must be installed separately from https://deno.land/"
