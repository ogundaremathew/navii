# Navii Makefile

.PHONY: build install download-data clean help

# Default target
all: build

# Build the CLI tool
build:
	@echo "🔨 Building Navii CLI..."
	go build -o bin/navii ./cmd/navii
	@echo "✅ Build complete! Binary available at: bin/navii"

# Install the CLI tool globally
install:
	@echo "📦 Installing Navii CLI globally..."
	go install ./cmd/navii
	@echo "✅ Navii CLI installed! You can now use 'navii' command globally."

# Download geographical data using the CLI
download-data: build
	@echo "🌍 Downloading geographical data..."
	./bin/navii -download-data

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	@echo "✅ Clean complete!"

# Show help
help:
	@echo "🌍 Navii Makefile Commands"
	@echo "=========================="
	@echo ""
	@echo "Available commands:"
	@echo "  make build        - Build the Navii CLI tool"
	@echo "  make install      - Install the CLI tool globally"
	@echo "  make download-data - Build and run data download"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make install"
	@echo "  2. navii -download-data"
	@echo ""