# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A terminal-based world clock application for tracking time across multiple timezones. Built with Go using the Charm TUI libraries (Bubbletea, Lipgloss, Bubbles).

## Commands

### Build and Run
```bash
go build -o tui-clock
./tui-clock
```

### Run directly
```bash
go run .
```

### Run with custom config
```bash
./tui-clock -config /path/to/config.yaml
```

### Install dependencies
```bash
go mod download
```

## Architecture

### Core Components

**Bubbletea Architecture (Elm-style)**
- `types.go`: Core data structures (Model, Config, Colleague, ColleagueTime, InputMode)
- `model.go`: Model initialization and business logic methods (Init, helper functions)
- `update.go`: Update function handling all messages and keyboard input
- `view.go`: View function rendering the UI with Lipgloss styling

**Supporting Modules**
- `config.go`: YAML configuration loading/saving with auto-creation of default config
- `timeline.go`: Timeline visualization rendering (individual & shared modes, color schemes)
- `timezone.go`: Timezone calculations (current time, offset, working hours detection)
- `timezones_data.go`: Comprehensive database of 200+ cities worldwide with timezone info
- `timezone_search.go`: Fuzzy search and smart name formatting for timezone lookup
- `styles.go`: Lipgloss style definitions (colors, formatting, ColorScheme type)
- `main.go`: Entry point with CLI flag parsing

### Data Flow

1. **Initialization**: Load config from `~/.config/tui-clock/config.yaml` (or custom path via `-config` flag)
2. **Auto-detection**: Local timezone detected automatically using `time.Now().Location()`
3. **Tick Loop**: Every second, `TickMsg` triggers time recalculation for all colleagues
4. **State Updates**: User input modifies model state, changes are persisted to config file
5. **Rendering**: View function renders current state with scrolling support (max 8 visible)

### Key Features

**Display**
- Real-time clocks updating every second
- Time offset from local timezone (+5h, -8h)
- Working hours indicator: ● (working), ○ (off-hours), ◆ (weekend)
- Date and day of week display
- Configurable 12h/24h time format
- Scrolling for >8 colleagues
- Timeline visualization mode with individual and shared views
- Three color schemes (classic, dark, high-contrast)

**Interactions**
- `↑/k, ↓/j`: Navigate
- `a`: Add colleague (prompts for name, then interactive timezone search)
- `e`: Edit selected colleague
- `d`: Delete selected colleague
- `f`: Toggle time format (12h ↔ 24h)
- `t`: Toggle timeline visualization mode
- `m`: Toggle timeline mode (individual ↔ shared) - only in timeline view
- `c`: Cycle color schemes - only in timeline view
- `?`: Help screen
- `q/Esc`: Quit

**Configuration**
- Config file: `~/.config/tui-clock/config.yaml`
- Auto-created with example colleagues on first run
- Changes saved immediately on add/edit/delete/format toggle/color scheme change/timeline mode toggle
- `time_format`: "12h" or "24h"
- `color_scheme`: "classic", "dark", or "high-contrast"
- `timeline_mode`: "individual" or "shared"
- `location_display_format`: "auto", "city", "timezone", or "abbreviation"
- Colleague fields: `work_start`, `work_end`, `sleep_start`, `sleep_end` (all optional, defaults provided)
- See `config.example.yaml` for structure

### Timezone Search Feature

**Comprehensive City Database**
- 200+ cities worldwide including:
  - All 50 US state capitals + 50+ major US cities
  - Major cities across Canada, Mexico, Central/South America
  - European cities (Western, Northern, Eastern regions)
  - Middle East, Africa, Asia (East, Southeast, South), Oceania

**Smart Search**
- Search by city name: "new york", "london", "tokyo"
- Search by abbreviation: "cst", "est", "pst" (shows all matching cities)
- Search by country: "japan", "germany", "australia"
- Search by state: "nebraska", "california"
- Fuzzy matching with real-time filtering
- Results ranked by popularity and relevance

**Auto-Append Location**
- Configurable via `location_display_format` in config:
  - `"auto"` (default): Append city if user searched by city, abbreviation if searched by abbrev
  - `"city"`: Always append city name
  - `"timezone"`: Always append IANA timezone
  - `"abbreviation"`: Always append abbreviation (EST, PST, etc.)
- Example: Search "london" → adds "Alice (London)"
- Example: Search "est" → adds "Alice (EST)"

**Search UX**
- Type to filter results (no need to press Enter while searching)
- Shows current time for each result to verify correctness
- `↑/↓` or `k/j` to navigate results
- `Enter` to select
- Scrolling for >10 results
- Handles ambiguous abbreviations (CST = Chicago, Shanghai, or Havana)

### Timezone Handling

- Uses Go's `time.LoadLocation()` with IANA timezone database
- Validation on add/edit prevents invalid timezone strings
- Working hours: Configurable per-colleague (default 9am-5pm)
- Weekend detection: Saturday/Sunday shown in purple
- Offset calculation: Accounts for DST automatically

### UI Layout

**Normal Mode:**
```
🌍 World Clock - Local Time: 15:30:45 (Mon, Jan 20)

  ▲ 2 more above
▶ ● Alice (New York)  10:30:45  -5h  Mon, Jan 20
  ○ Bob (London)      15:30:45  same  Mon, Jan 20
  ◆ Charlie (Tokyo)   00:30:45  +9h  Tue, Jan 21
  ▼ 3 more below

↑/k up • ↓/j down • a add • e edit • d delete • f format • ? help • q quit
```

**Timeline Mode:**
```
🌍 Timeline View - Local Time: 15:30:45 (Mon, Jan 20)

Katherine (EST)      14:49:34  [░░░▓▓▓████████████|█████▓▓▓░░░░░]
Austin (Lincoln)     13:49:34  [░░░▓▓▓█████████████|████▓▓▓░░░░░]
Ben (EST)            14:49:34  [░░░▓▓▓████████████|█████▓▓▓░░░░░]
                                0    6    12   18   24

░ sleep • ▓ off-hours • █ work • █ now

t normal mode • m individual • ↑/↓ scroll • c cycle colors • ? help • q quit
```

### Timeline Visualization Architecture

**Purpose**: Visualize each colleague's full 24-hour day with sleep/work/off-hours blocks and current time marker.

**Files**:
- `timeline.go` (~500 lines) - All timeline rendering logic
- `styles.go` - ColorScheme type and three built-in schemes
- `types.go` - ModeTimeline enum, timeline constants, Colleague accessor methods

**Two Visualization Modes:**

1. **Individual Mode** (`timeline_mode: "individual"`):
   - Each bar represents 0-24 hours in that person's local timezone
   - Current time marker highlights their local time position
   - Hour labels (0, 6, 12, 18, 24) represent hours in their timezone
   - Use case: "What time is it for them right now? What are they doing?"

2. **Shared Mode** (`timeline_mode: "shared"`):
   - Each bar represents 0-24 hours in YOUR (local) timezone
   - Their activities are shifted by timezone offset to align with your day
   - Current time marker at YOUR local time (same position for everyone)
   - Hour labels represent hours in your timezone
   - Use case: "Who's available RIGHT NOW? Who's working during my morning?"

**Key Insight**: Both modes use identical visual layout (Name + Time + Bar + Labels). Only the bar content and label interpretation differ.

**Bar Generation:**

Characters used:
- `░` - Sleep hours (default: 23:00-07:00, configurable per colleague)
- `▓` - Off-hours (awake but not working)
- `█` - Work hours (default: 09:00-17:00 weekdays, configurable per colleague)

Current time marker: Highlighted character (cyan/bold) instead of replacement.

**Key Functions:**

**Orchestration:**
```go
// Main entry point - handles both modes
func (m Model) renderTimeline() string
```

**Individual Mode:**
```go
// Renders one person's row (name + time + bar)
func (m Model) renderTimelineRow(index int, ct ColleagueTime) string

// Generates their 24-hour bar (0-24 in their timezone)
func (m Model) renderIndividualBar(ct ColleagueTime, barWidth int) string
```

**Shared Mode:**
```go
// Renders one person's row (same format as individual)
func (m Model) renderSharedTimelineRow(index int, ct ColleagueTime) string

// Generates bar shifted by offset (0-24 in local timezone)
func (m Model) renderSharedBar(ct ColleagueTime, offsetHours float64, barWidth int) string
```

**Shared Utilities:**
```go
// Applies color scheme to bar characters
func (m Model) colorizeBar(bar []rune, ct ColleagueTime, markerIndex int) string

// Renders hour labels (0, 6, 12, 18, 24) with precise alignment
func (m Model) renderHourLabels(barWidth int, leftPadding int) string

// Checks if hour falls in range (handles wraparound like 23:00-07:00)
func isInTimeRange(hour, start, end int) bool

// Colleague accessor methods with defaults (DRY principle)
func (c Colleague) GetWorkStart() int
func (c Colleague) GetWorkEnd() int
func (c Colleague) GetSleepStart() int
func (c Colleague) GetSleepEnd() int
```

**Color Schemes:**

Six built-in schemes defined in `colorSchemes` map:
- **Classic**: Vibrant (cyan, green, purple) - default
- **Dark**: Muted night-mode colors
- **High Contrast**: Accessibility-focused
- **Nord**: Nordic-inspired bluish theme
- **Solarized**: Light variant with reduced eye strain
- **Solarized Dark**: Warm dark tones for solarized users

**Phase 1 System** (Implemented): Self-discovering color scheme system
- Add new schemes by adding to `colorSchemes` map only (one place)
- `GetAvailableColorSchemes()` - Auto-discovers all registered schemes
- `GetNextColorScheme()` - Cycles alphabetically through all schemes
- `ValidateColorScheme()` - Validates scheme completeness

Each ColorScheme defines:
- `SleepColor`, `AwakeOffColor`, `WorkColor` - Bar character colors
- `MarkerColor` - Current time highlight color
- `WeekendTint` - Work block color on weekends
- `Primary`, `Secondary`, `Success`, `Warning`, `Error`, `Muted` - UI colors

**Bar Width Calculation:**

Adaptive based on terminal width:
```go
func (m Model) calculateTimelineBarWidth() int
```

Constants:
- `MinBarWidth = 24` - Minimum (1 char per hour)
- `IdealBarWidth = 48` - Target (2 chars per hour)
- `MaxBarWidth = 72` - Maximum (3 chars per hour)

Reserved space: NameFieldWidth (25) + TimeFieldWidth (12) + padding + brackets

**Critical Implementation Details:**

1. **Offset Direction** (shared mode):
   ```go
   // CORRECT: If they're +3h ahead, add offset to shift right
   theirHourFraction := localHourFraction + (offsetHours / 24.0)
   ```

2. **Marker Highlighting**:
   - Keep original character (░, █, ▓)
   - Apply cyan/bold styling to that position
   - Do NOT replace with '|' or other characters

3. **Hour Label Alignment**:
   - Character-by-character positioning for multi-char labels
   - Center labels around target hour position
   - Account for bracket in position calculations

4. **Unified Layout**:
   - Both modes render: Name + Time + Bar
   - Hour labels always at bottom (not per row)
   - Same scroll indicators, same styling

**Testing Considerations:**

When adding tests (see Phase 5.3 in plans/remaining-work.md):
- Test `isInTimeRange()` with wraparound cases (23:00-07:00)
- Test bar width calculations at various terminal sizes
- Test individual bar generation (character distribution, marker position)
- Test shared bar generation (offset shifting, wraparound at day boundaries)
- Test color scheme validity and cycling
- Integration tests for complete timeline rendering

**Common Patterns:**

- Use `truncateOrPad()` for fixed-width fields
- Use `isInTimeRange()` for hour checks (handles wraparound automatically)
- Use ColorScheme getters for colors, never hardcode
- Use Colleague accessor methods for work/sleep hours (provides defaults)
