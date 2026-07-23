# List available recipes
default:
    @just --list

# Build the binary
build:
    go build -o tui-clock

# Build and run the application
run: build
    ./tui-clock

# Full verification gate — identical to CI; run before every commit
check:
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
lint:
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
