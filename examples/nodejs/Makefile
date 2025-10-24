.PHONY: all install clean test test-basic test-batch test-decode test-all help

# Default target
all: install test

# Install dependencies
install:
	@echo "Installing Node.js dependencies..."
	@# Build the TypeScript package first
	@cd ../../typescript && npm run build
	@# Pack the package
	@cd ../../typescript && npm pack && mv libnextimage-*.tgz ../examples/nodejs/libnextimage-0.4.0.tgz
	@# Install in examples
	@npm install
	@echo "✓ Installation complete"

# Run all tests
test: test-basic test-batch test-decode test-all
	@echo ""
	@echo "✓ All Node.js example tests passed!"

# Run basic encoding test
test-basic:
	@echo ""
	@echo "Running basic encoding test..."
	@npm run test:basic

# Run batch processing test
test-batch:
	@echo ""
	@echo "Running batch processing test..."
	@npm run test:batch

# Run decode test
test-decode:
	@echo ""
	@echo "Running decode test..."
	@npm run test:decode

# Run all features test
test-all:
	@echo ""
	@echo "Running all features test..."
	@npm run test:all

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -rf node_modules
	@rm -f package-lock.json
	@rm -f libnextimage-*.tgz
	@rm -f ideamans-libnextimage-*.tgz
	@rm -rf output
	@echo "✓ Cleanup complete"

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Install dependencies and run all tests (default)"
	@echo "  install    - Install dependencies"
	@echo "  test       - Run all tests"
	@echo "  test-basic - Run basic encoding test"
	@echo "  test-batch - Run batch processing test"
	@echo "  test-decode - Run decode test"
	@echo "  test-all   - Run all features test"
	@echo "  clean      - Remove installed dependencies and outputs"
	@echo "  help       - Show this help message"
