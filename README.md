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
- 📊 **Timeline visualization mode** - visualize everyone's day at a glance
- 🎨 **Multiple color schemes** - classic, dark, and high-contrast modes

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

### Normal Mode

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `a` | Add new colleague |
| `e` | Edit selected colleague |
| `d` | Delete selected colleague |
| `f` | Toggle time format (12h ↔ 24h) |
| `t` | Enter timeline visualization mode |
| `?` / `h` | Show help screen |
| `q` / `Esc` | Quit |
| `Ctrl+C` | Force quit |

### Timeline Mode

| Key | Action |
|-----|--------|
| `t` | Return to normal mode |
| `m` | Toggle mode (individual ↔ shared) |
| `c` | Cycle color schemes |
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |
| `?` | Show help screen |
| `q` / `Esc` | Quit |

## Status Indicators

- **● Green** - Working hours (9am-5pm, weekdays)
- **○ Gray** - Off hours (outside 9am-5pm or weekends)
- **◆ Purple** - Weekend (Saturday/Sunday)

## Timeline Visualization

Press `t` to enter **timeline mode** - visualize everyone's day at a glance!

### Two Visualization Modes

**Individual Mode** - See each person's local day:
```
Katherine (EST)      14:49:34  [░░░▓▓▓████████████|█████▓▓▓░░░░░]
Austin (Lincoln)     13:49:34  [░░░▓▓▓█████████████|████▓▓▓░░░░░]
Ben (EST)            14:49:34  [░░░▓▓▓████████████|█████▓▓▓░░░░░]
                                0    6    12   18   24
```
- Each bar shows their full 24-hour day (0-24 in their timezone)
- Current time marker highlights their local time
- See what they're doing throughout their entire day

**Shared Mode** - See who's available RIGHT NOW:
```
Katherine (EST)      14:49:34  [    ████████████|█████        ]
Austin (Lincoln)     13:49:34  [  █████████████|████          ]
Ben (EST)            14:49:34  [    ████████████|█████        ]
                                0    6    12   18   24
```
- Bars aligned to **your** local time (0-24 in your timezone)
- Activities shifted to show when they're working/sleeping relative to you
- Current time marker shows your local time (same position for everyone)
- Instantly see who's available during your working hours

### Timeline Legend

- **░** - Sleep hours (default: 11pm-7am)
- **▓** - Off-hours (awake but not working)
- **█** - Work hours (default: 9am-5pm, weekdays only)
- **Highlighted character** - Current time (cyan/bold)

### Color Schemes

Press `c` to cycle through five built-in color schemes:

1. **Classic** - Vibrant true colors (cyan, green, purple) for maximum visibility
2. **Dark** - Muted night-mode true colors for low-light environments
3. **High Contrast** - Accessibility-focused with strong differentiation using ANSI colors
4. **Nord** - Nordic-inspired theme using official Nord colors with adaptive backgrounds
5. **Solarized** - Precision-crafted adaptive scheme using official Solarized colors, automatically adjusts for light/dark terminal backgrounds

**Adaptive Schemes:** Nord and Solarized use adaptive colors that automatically adjust background tones based on your terminal's light or dark theme, while keeping accent colors vibrant and consistent.

Your color scheme preference is automatically saved to your config file.

### How It Works

- **Individual mode**: Each timeline shows 0-24 hours in that person's timezone. When it's 3pm for them, the marker is at the 3pm position on their bar.

- **Shared mode**: All timelines show 0-24 hours in YOUR timezone. Activities are shifted so you can see what each person is doing at any given hour of YOUR day.

**Example**: If Katherine is in EST (+5h ahead of PST) and it's noon in PST:
- In individual mode: Her marker is at 17:00 (5pm) on her 0-24 bar
- In shared mode: Her work block appears shifted right, marker at 12:00 (your noon)

Press `m` to toggle between modes and see the difference!

## Configuration

The configuration file uses YAML format:

```yaml
time_format: "24h"        # Options: "12h" or "24h"
color_scheme: "classic"   # Options: "classic", "dark", "high-contrast", "nord", "solarized"
timeline_mode: "individual"  # Options: "individual", "shared"

colleagues:
  - name: "Alice (New York)"
    timezone: "America/New_York"
    work_start: 9    # 9am in 24h format (default if omitted)
    work_end: 17     # 5pm in 24h format (default if omitted)
    sleep_start: 23  # 11pm in 24h format (default if omitted)
    sleep_end: 7     # 7am in 24h format (default if omitted)

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
├── timeline.go          # Timeline visualization (individual & shared modes)
├── config.go            # YAML config management
├── timezone.go          # Time calculations
├── timezones_data.go    # City database (200+ cities)
├── timezone_search.go   # Search & ranking logic
├── styles.go            # UI styling (including color schemes)
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
