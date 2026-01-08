.PHONY: help build test test-verbose test-coverage lint fmt clean run install screenshots clean-screenshots

# Default target
help:
	@echo "TUI Clock - Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make test           - Run tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run golangci-lint (requires golangci-lint installed)"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make run            - Build and run the application"
	@echo "  make install        - Install binary to GOPATH/bin"
	@echo "  make screenshots    - Generate README screenshots (requires VHS: brew install vhs)"

# Build the binary
build:
	@echo "Building tui-clock..."
	go build -o tui-clock

# Run tests
test:
	@echo "Running tests..."
	go test

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report, run: go tool cover -html=coverage.out"

# Run linter (requires golangci-lint: https://golangci-lint.run/usage/install/)
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "Error: golangci-lint not installed"; \
		echo "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "Formatting code with gofmt..."
	gofmt -w .
	@echo "Done!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f tui-clock
	rm -f coverage.out
	rm -f *.coverprofile
	@echo "Done!"

# Build and run
run: build
	./tui-clock

# Install to GOPATH/bin
install:
	@echo "Installing tui-clock to GOPATH/bin..."
	go install
	@echo "Done! Binary installed to: $$(go env GOPATH)/bin/tui-clock"

# Generate README screenshots (requires VHS: brew install vhs)
screenshots: build
	@echo "Generating screenshots with VHS..."
	@mkdir -p assets
	@if command -v vhs >/dev/null 2>&1; then \
		vhs demo/demo.tape; \
	else \
		echo "Error: VHS not installed"; \
		echo "Install with: brew install vhs"; \
		exit 1; \
	fi

# Clean generated screenshots
clean-screenshots:
	rm -rf assets/*.png
