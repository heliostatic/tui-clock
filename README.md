# 🌍 TUI World Clock

A terminal-based world clock application for tracking time across multiple timezones. Perfect for remote teams to know when colleagues are available.

![Built with Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![TUI](https://img.shields.io/badge/TUI-Bubbletea-FF69B4)

## Features

- ⏰ Real-time clocks for multiple timezones
- 🌐 Time offset display from your local timezone
- 💼 Working hours indicator (weekdays vs weekends)
- 📅 Date and day of week
- 🔄 Toggle between 12h/24h format
- ✏️ Interactive editing (add/edit/delete colleagues)
- 💾 Persistent configuration
- 📜 Scrolling support for 8+ colleagues

## Installation

### Prerequisites

- Go 1.21 or later

### Build from source

```bash
git clone <repository-url>
cd tui-clock
go build -o tui-clock
./tui-clock
```

Or run directly:
```bash
go run .
```

## Usage

### Basic Usage

```bash
./tui-clock
```

On first run, a default configuration file will be created at `~/.config/tui-clock/config.yaml` with example colleagues.

### Custom Config Location

```bash
./tui-clock -config /path/to/your/config.yaml
```

## Keyboard Controls

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `a` | Add new colleague |
| `e` | Edit selected colleague |
| `d` | Delete selected colleague |
| `f` | Toggle time format (12h ↔ 24h) |
| `?` / `h` | Show help screen |
| `q` / `Esc` | Quit |
| `Ctrl+C` | Force quit |

## Status Indicators

- **● Green** - Working hours (9am-5pm, weekdays)
- **○ Gray** - Off hours (outside 9am-5pm or weekends)
- **◆ Purple** - Weekend (Saturday/Sunday)

## Configuration

The configuration file uses YAML format:

```yaml
time_format: "24h"  # Options: "12h" or "24h"

colleagues:
  - name: "Alice (New York)"
    timezone: "America/New_York"
    work_start: 9   # 9am in 24h format
    work_end: 17    # 5pm in 24h format

  - name: "Bob (London)"
    timezone: "Europe/London"
    work_start: 9
    work_end: 17
```

### Common Timezones

**Americas**: `America/New_York`, `America/Los_Angeles`, `America/Chicago`, `America/Denver`

**Europe**: `Europe/London`, `Europe/Paris`, `Europe/Berlin`, `Europe/Moscow`

**Asia**: `Asia/Tokyo`, `Asia/Shanghai`, `Asia/Hong_Kong`, `Asia/Singapore`, `Asia/Kolkata`

**Pacific**: `Australia/Sydney`, `Pacific/Auckland`

**Africa**: `Africa/Cairo`, `Africa/Johannesburg`

See `config.example.yaml` for a full example configuration.

## How It Works

- **Auto-detection**: Your local timezone is automatically detected
- **IANA Database**: Uses Go's timezone database for accurate time calculations
- **DST Support**: Automatically handles Daylight Saving Time
- **Persistent**: All changes (add/edit/delete) are immediately saved to the config file

## Built With

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling and layout
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Development

### Running Tests

This project has comprehensive unit tests for core functionality:

```bash
# Run all tests
go test

# Run tests with verbose output
go test -v

# Run tests with coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

The test suite covers:
- **Timezone calculations** - Time formatting, offset calculation, working hours detection
- **Search functionality** - City/country/abbreviation search, ranking, case sensitivity
- **Configuration** - Config loading/saving, defaults, validation
- **Input helpers** - Text input creation, mode transitions, navigation
- **Business logic** - Add/edit/delete operations, search state management

### Project Structure

```
tui-clock/
├── main.go              # Entry point
├── types.go             # Data structures
├── model.go             # Bubbletea model & business logic
├── update.go            # Input handling & state updates
├── view.go              # UI rendering
├── config.go            # YAML config management
├── timezone.go          # Time calculations
├── timezones_data.go    # City database (200+ cities)
├── timezone_search.go   # Search & ranking logic
├── styles.go            # UI styling
├── inputs.go            # Input helpers & utilities
└── *_test.go           # Unit tests
```

## Contributing

We welcome contributions! Here's how to get started:

### Setting Up Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/tui-clock.git
   cd tui-clock
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build and run:
   ```bash
   go build -o tui-clock
   ./tui-clock
   ```

### Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. Make your changes
3. **Add tests** for new functionality
4. Run tests to ensure everything works:
   ```bash
   go test -v
   ```
5. Run the app to test manually:
   ```bash
   go run .
   ```

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format code
- Keep functions focused and well-named
- Add comments for non-obvious logic
- Maintain the DRY principle (see `inputs.go` for helper patterns)

### Submitting Changes

1. Ensure all tests pass: `go test`
2. Commit your changes with a clear message:
   ```bash
   git commit -m "Add feature: description of what you did"
   ```
3. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
4. Open a Pull Request with:
   - Clear description of changes
   - Why the change is needed
   - Any breaking changes
   - Screenshots (if UI changes)

### Adding New Features

**For new timezones/cities:**
- Add entries to `timezones_data.go` (alphabetically within region)
- Include city, country, IANA timezone, common abbreviations, and popularity rank

**For new features:**
- Update types in `types.go` if needed
- Add business logic to `model.go`
- Handle UI updates in `update.go` and `view.go`
- **Write tests** in corresponding `*_test.go` files
- Update this README if user-facing
- Update `CLAUDE.md` if architecture changes

### Bug Reports

Found a bug? Please open an issue with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Go version, terminal)
- Relevant config.yaml (if applicable)

## License

MIT
