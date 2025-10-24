.PHONY: all install clean test test-basic help

# Default target
all: install test

# Install dependencies
install:
	@echo "Installing Bun dependencies..."
	@# Build the TypeScript package first
	@cd ../../typescript && npm run build
	@# Pack the package
	@cd ../../typescript && npm pack && mv libnextimage-*.tgz ../examples/bun/ideamans-libnextimage-0.4.0.tgz
	@# Install in examples
	@bun install
	@echo "✓ Installation complete"

# Run all tests
test: test-basic
	@echo ""
	@echo "✓ All Bun example tests passed!"

# Run basic encoding test
test-basic:
	@echo ""
	@echo "Running basic encoding test..."
	@bun run test:basic

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -rf node_modules
	@rm -f bun.lock
	@rm -f ideamans-libnextimage-*.tgz
	@rm -f libnextimage-*.tgz
	@rm -rf output
	@echo "✓ Cleanup complete"

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Install dependencies and run all tests (default)"
	@echo "  install    - Install dependencies"
	@echo "  test       - Run all tests"
	@echo "  test-basic - Run basic encoding test"
	@echo "  clean      - Remove installed dependencies and outputs"
	@echo "  help       - Show this help message"
