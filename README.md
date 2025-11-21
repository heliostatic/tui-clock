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

### Quick Start

```bash
git clone https://github.com/heliostatic/tui-clock.git
cd tui-clock
make build
./tui-clock
```

Or use `make run` to build and run in one step:
```bash
make run
```

### Install to System

```bash
make install
# Binary will be installed to $(go env GOPATH)/bin/tui-clock
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

### Quick Commands

This project uses a `Makefile` for common development tasks:

```bash
make help           # Show all available commands
make build          # Build the binary
make test           # Run tests
make test-coverage  # Run tests with coverage report
make lint           # Run golangci-lint (requires golangci-lint installed)
make fmt            # Format code with gofmt
make clean          # Remove build artifacts
```

### Test Coverage

The project has comprehensive unit tests (~37% coverage, appropriate for a TUI app):

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

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on:

- Setting up your development environment
- Code style and best practices
- Running tests and linters
- Submitting pull requests
- Adding new features
- Reporting bugs

Quick start for contributors:
```bash
git clone https://github.com/your-username/tui-clock.git
cd tui-clock
make build
make test
```

## License

MIT
