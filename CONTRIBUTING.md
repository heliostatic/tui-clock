# Contributing to TUI World Clock

We welcome contributions! This guide will help you get started.

## Table of Contents

- [Setting Up Development Environment](#setting-up-development-environment)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Adding New Features](#adding-new-features)
- [Bug Reports](#bug-reports)

## Setting Up Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork:**
   ```bash
   git clone https://github.com/your-username/tui-clock.git
   cd tui-clock
   ```

3. **Install dependencies:**
   ```bash
   go mod download
   ```

4. **Verify your setup:**
   ```bash
   make build
   make test
   ```

## Development Workflow

We use a `Makefile` for common development tasks:

```bash
make help           # Show all available commands
make build          # Build the binary
make test           # Run tests
make test-verbose   # Run tests with verbose output
make test-coverage  # Run tests with coverage report
make lint           # Run golangci-lint
make fmt            # Format code with gofmt
make clean          # Remove build artifacts
make run            # Build and run the application
```

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow existing patterns in the codebase
   - Add tests for new functionality

3. **Format and lint your code:**
   ```bash
   make fmt
   make lint
   ```

4. **Run tests:**
   ```bash
   make test-verbose
   ```

5. **Test manually:**
   ```bash
   make run
   ```

## Code Style

- **Follow standard Go conventions** - Use `gofmt` for formatting
- **Keep functions focused** - Single responsibility principle
- **Use descriptive names** - Variables, functions, and types should be self-documenting
- **Add comments for non-obvious logic** - Especially for complex algorithms
- **Maintain DRY principle** - See `inputs.go` for helper function patterns
- **Extract constants** - No magic numbers; use constants from `types.go`

### Project Structure

```
tui-clock/
├── main.go              # Entry point & CLI argument parsing
├── types.go             # Data structures & constants
├── model.go             # Bubbletea model & business logic
├── update.go            # Input handling & state updates
├── view.go              # UI rendering
├── config.go            # YAML config management
├── timezone.go          # Time calculations
├── timezones_data.go    # City database (200+ cities)
├── timezone_search.go   # Search & ranking logic
├── styles.go            # UI styling with Lipgloss
├── inputs.go            # Input helpers & utilities
└── *_test.go            # Unit tests
```

## Testing

### Test Coverage

Current coverage: ~37% (appropriate for a TUI app)

The test suite covers:
- **Timezone calculations** - Time formatting, offset calculation, working hours detection
- **Search functionality** - City/country/abbreviation search, ranking, case sensitivity
- **Configuration** - Config loading/saving, defaults, validation
- **Input helpers** - Text input creation, mode transitions, navigation
- **Business logic** - Add/edit/delete operations, search state management

### Writing Tests

- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Focus on pure functions (business logic, calculations, parsing)
- UI handlers (Bubbletea Update/View) are hard to unit test - avoid over-testing these

Example:
```go
func TestYourFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "output1"},
        {"case 2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := YourFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Submitting Changes

1. **Ensure all tests pass:**
   ```bash
   make test
   ```

2. **Format and lint your code:**
   ```bash
   make fmt
   make lint
   ```

3. **Commit with a clear message:**
   ```bash
   git commit -m "Add feature: description of what you did"
   ```

   Good commit messages:
   - Start with a verb (Add, Fix, Update, Refactor)
   - Be concise but descriptive
   - Explain *why* not just *what*

4. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Open a Pull Request** with:
   - Clear description of changes
   - Why the change is needed
   - Any breaking changes
   - Screenshots (if UI changes)
   - Link to related issues

## Adding New Features

### For New Timezones/Cities

Add entries to `timezones_data.go` (alphabetically within region):

```go
{"City Name", "Country", "IANA/Timezone", []string{"TZ", "ABBREV"}, popularity},
```

### For New Features

1. **Update types** in `types.go` if needed (new constants, structs, or enums)
2. **Add business logic** to `model.go` (data manipulation, state management)
3. **Handle UI updates** in `update.go` (input handling) and `view.go` (rendering)
4. **Write tests** in corresponding `*_test.go` files
5. **Update README.md** if user-facing changes
6. **Update CLAUDE.md** if architecture changes

### Examples

**Adding a constant:**
```go
// types.go
const (
    AutoHideTimeout  = 3 * time.Second
    YourNewConstant  = 10  // Add your constant
)
```

**Adding a helper function:**
```go
// inputs.go or model.go
func yourHelperFunction() {
    // Implementation
}
```

**Following DRY principles:**
- Extract duplicate code into helper functions
- Use constants for magic numbers
- Reuse existing patterns where possible

## Bug Reports

Found a bug? Please [open an issue](https://github.com/heliostatic/tui-clock/issues) with:

- **Clear description** of the problem
- **Steps to reproduce** (numbered list)
- **Expected behavior** vs **actual behavior**
- **Environment:**
  - OS and version
  - Go version (`go version`)
  - Terminal emulator
- **Config file** (if relevant) - sanitize any sensitive data
- **Screenshots/recordings** (if UI-related)

## Questions?

- Check existing issues and pull requests
- Open a new issue with the "question" label
- Be respectful and patient - maintainers are volunteers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
