# Keep in sync with the golangci-lint version in .github/workflows/ci.yml
golangci_version := "v2.5.0"

# List available recipes
default:
    @just --list

# Install dev tools (golangci-lint pinned to the CI version) and dependencies
bootstrap:
    go mod download
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{{golangci_version}}
    @echo ""
    @echo "Installed golangci-lint {{golangci_version}} to $(go env GOPATH)/bin — make sure that's on your PATH."
    @echo "Optional, for 'just screenshots': go install github.com/charmbracelet/vhs@latest (also needs ttyd and ffmpeg)."

# (hidden) Fail with a pointer to bootstrap when golangci-lint is missing
_ensure-golangci:
    @command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found — run 'just bootstrap' first"; exit 1; }

# Build the binary
build:
    go build -o tui-clock

# Build and run the application
run: build
    ./tui-clock

# Full verification gate — identical to CI; run before every commit
check: _ensure-golangci
    go build ./...
    go vet ./...
    @test -z "$(gofmt -l .)" || (gofmt -l . && exit 1)
    golangci-lint run
    go test -race ./...

# Run tests (race detector on, like CI)
test:
    go test -race ./...

# Run tests with verbose output
test-verbose:
    go test -race -v ./...

# Run tests with a coverage report
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    @echo ""
    @echo "HTML report: go tool cover -html=coverage.out"

# Run golangci-lint
lint: _ensure-golangci
    golangci-lint run

# Format code with gofmt
fmt:
    gofmt -w .

# Remove build artifacts
clean:
    rm -f tui-clock coverage.out

# Install binary to GOPATH/bin
install:
    go install

# Regenerate README screenshots (requires VHS: brew install vhs)
screenshots: build
    mkdir -p assets
    vhs demo/demo.tape

# Remove generated screenshots
clean-screenshots:
    rm -rf assets/*.png
