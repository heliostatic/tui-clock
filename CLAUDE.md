# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A terminal-based world clock application for tracking time across multiple timezones. Built with Go using the Charm TUI libraries (Bubbletea, Lipgloss, Bubbles).

## Commands

### Build and Run
```bash
go build -o tui-clock
./tui-clock
./tui-clock -config /path/to/config.yaml   # custom config
go run .                                    # run directly
```

### Verify (run all of these before every commit)
```bash
go build ./...
go vet ./...
gofmt -l .            # must print nothing
golangci-lint run     # v2 config in .golangci.yml; must report 0 issues
go test -race ./...
```

CI (`.github/workflows/ci.yml`) runs the same gate plus a build matrix on linux/macos/windows. Releases: pushing a `v*` tag runs goreleaser (`.goreleaser.yaml`) and publishes binaries. The IANA timezone DB is embedded via `time/tzdata`, so builds work on systems without one.

## Development Practices

These are the practices that have repeatedly caught real bugs here. Follow them for any non-trivial change.

### 1. Drive the real app, don't just unit test

Unit tests missed a bug where pasted input was silently dropped (Bubbletea batches fast keystrokes into one multi-rune message; tests fed single chars). Only driving the actual binary caught it. For any behavior change, launch the app in tmux and interact with it:

```bash
tmux new-session -d -s app -x 120 -y 25 "env TZ=UTC ./tui-clock -config /tmp/test-config.yaml"
timeout 10 bash -c 'until tmux capture-pane -t app -p | grep -q "World Clock"; do sleep 0.2; done'
tmux send-keys -t app 'a'        # drive keys; send strings for text entry
tmux capture-pane -t app -p      # read the screen
tmux kill-session -t app
```

Craft config files for the scenario under test (invalid timezones, half-hour zones like Asia/Kolkata, CJK names, midnight work hours) and inspect the config file on disk afterwards — several bugs only showed up as wrong YAML, not wrong pixels.

### 2. Adversarial review before merge

Before merging a substantive PR, run an independent adversarial review of the diff whose explicit job is to break it: for each claimed fix, find inputs where the old bug still manifests or a new one appears. Require findings in the form file:line + concrete failure scenario + confidence (CONFIRMED = reproduced/proved, PLAUSIBLE = suspicion). This has caught, among others: a silent config regression for existing users, a code path that could overwrite the user's config with defaults during an editor's rename-save, and wrong dates for midnight DST transitions.

### 3. Pin critical predicates with mutation checks

If a test suite passes both with a predicate and with its plausible-but-wrong variant (e.g. `mtime.Equal` vs `!mtime.After`), the behavior is not actually tested. When a comparison or boundary matters, temporarily apply the wrong variant and confirm a test fails; if none does, write the test that kills the mutant (see `TestMaybeReloadConfigTriggersOnOlderMtime`, which uses a same-size edit so the size fallback can't mask the mtime comparison).

### 4. Deterministic tests

- Never depend on wall-clock "now" for assertions. Use fixed dates in known-DST periods (e.g. US transitions Mar 8 / Nov 1 2026), `time.FixedZone`, and `time.UTC`.
- For file-watching tests, set mtimes explicitly with `os.Chtimes` — never rely on filesystem timestamp granularity.
- When "now" is unavoidable (working-hours detection), construct ranges around `time.Now()`'s hour so the assertion holds at any time of day while still exercising the wraparound branch.

### 5. Issues, commits, and merges

- File a GitHub issue for every discovered bug, even if fixing it immediately; one commit per issue with `Fixes #N` so merges close them.
- User-visible design changes (render precedence, selection behavior) get an issue describing the trade-off rather than a unilateral change, unless the owner has already authorized it.
- PRs open as drafts; merge only when CI is green and the adversarial review's confirmed findings are fixed.
- Report honestly: if something is unverified or deferred, the PR body says so.

## Code Invariants

Each of these encodes a fixed bug. Violating them reintroduces it.

1. **Hour fields are `*int`; always use the accessors.** `Colleague.WorkStart/WorkEnd/SleepStart/SleepEnd` use nil = "default", so 0 (midnight) is a real value. Never read the raw fields — use `GetWorkStart()` etc.; construct values with `HourPtr(n)`. `parseConfig` migrates the legacy (0, 0) pair (old binaries' "use defaults" sentinel) back to nil; don't remove that migration while old configs exist.
2. **Config mutations go through `ColleagueTime.ConfigIndex`, never the display cursor.** The display list can include invalid-timezone entries and is not guaranteed to align index-for-index with `Config.Colleagues` semantics; using the cursor once deleted the wrong colleague.
3. **All config writes go through `m.saveConfig()`** — it stamps mtime/size so hot-reload doesn't re-read our own writes. Never call `SaveConfig` directly from update handlers. The reload path (`maybeReloadConfig`) must never write to the file: it reads + `parseConfig` (pure, no filesystem), specifically avoiding `LoadConfig`'s create-if-missing side effect, which once made a mid-rename editor save resurrect defaults.
4. **Measure display width in cells, not bytes.** Use `truncateOrPad` (go-runewidth) for fixed-width fields; `len()` on names broke alignment for CJK/accented text and could split UTF-8 sequences.
5. **Text-input modes must not steal typeable keys, and must accept batched runes.** Search navigation is arrow-keys only (`k`/`j` are letters in "tokyo"). Check `msg.Type == tea.KeyRunes || tea.KeySpace` and append `string(msg.Runes)` — Bubbletea coalesces fast typing/paste into one multi-rune message; a `len(key) == 1` check silently drops it.
6. **Time ranges are half-open `[start, end)` with wraparound.** `isInTimeRange` (int hours, for "what is happening now") and `isInTimeRangeFrac` (fractional hours, for bar positions — keeps half/quarter-hour offsets exact). `start == end` is an empty range. In bars, configured **work hours take precedence over sleep** (`barCharForHour`) so night shifts render as working.
7. **Offsets display exactly.** `formatOffsetString` uses shortest-exact-decimal ("+5.5h", "+5.75h"); `%.1f` once showed Nepal as "+5.8h". Offset math is `float64` seconds/3600 — integer division truncated half-hour zones.
8. **Ticks align to wall-clock second boundaries** (`time.Until(now.Truncate(time.Second).Add(time.Second))`), otherwise the seconds display drifts.
9. **Marker highlighting recolors the existing bar character** (cyan/bold); it never replaces it with `|` or anything else.
10. **Use ColorScheme getters, never hardcoded colors**, and add new schemes only via the `colorSchemes` map — discovery, cycling (`GetNextColorScheme`), and validation (`ValidateColorScheme`) pick them up automatically.

## Architecture

### Core Components

**Bubbletea Architecture (Elm-style)**
- `types.go`: Core data structures (Model, Config, Colleague, ColleagueTime, InputMode, constants)
- `model.go`: Model init, config save/hot-reload, scrub helpers, business logic methods
- `update.go`: Update function handling all messages and keyboard input, per-mode handlers
- `view.go`: Normal-mode/help rendering with Lipgloss styling
- `inputs.go`: Text-input construction, hour-range parsing, search navigation

**Supporting Modules**
- `config.go`: YAML load/save; `parseConfig` (pure parse+defaults+migration) vs `LoadConfig` (adds create-if-missing, startup only)
- `timeline.go`: Timeline rendering (both modes, overlap row, hour labels, bar math)
- `timezone.go`: Time computation per colleague, offsets, working hours, DST transition detection
- `timezones_data.go`: Database of 200+ cities with IANA zones, abbreviations, popularity
- `timezone_search.go`: Fuzzy search scoring and display-name formatting
- `styles.go`: Lipgloss styles, ColorScheme type, five built-in schemes
- `main.go`: Entry point, CLI flags, tzdata embed

### Data Flow

1. **Initialization**: Load config from `~/.config/tui-clock/config.yaml` (or `-config` path); local timezone from `time.Now().Location()`
2. **Tick Loop**: Every second `TickMsg` recomputes all colleague times and runs the config hot-reload check (`maybeReloadConfig`): external edits are detected by mtime/size, deferred while a modal flow is open, and torn/invalid/vanished files are retried on later ticks. Conflict semantics are whole-file last-writer-wins: an in-app save overwrites external edits made since the last reload
3. **State Updates**: Keyboard input mutates the model; config changes persist immediately via `saveConfig`
4. **Rendering**: View renders current state; lists scroll at >8 colleagues (>10 search results)

### Key Features

**Display**
- Real-time clocks; offset from local time ("+5.5h" for half-hour zones)
- Status dot: ● working, ○ off-hours, ◆ weekend, ⚠ invalid timezone
- DST warning ("⚡-1h Nov 1") when a colleague's offset changes within 7 days
- 12h/24h format; five color schemes (classic, dark, high-contrast, nord, solarized) with true color and adaptive light/dark support
- Timeline mode: individual and shared views, team overlap row, time scrubbing

**Interactions**
- `↑/k, ↓/j`: Navigate (selection auto-hides after 3s; first keypress reactivates it)
- `a`: Add colleague (name prompt, then type-to-filter timezone search)
- `e`: Edit selected colleague's name and timezone
- `w`: Edit selected colleague's work/sleep hours (two-step prompt; staged, applied only on final confirm; blank keeps, `default` resets, `9-24` = until midnight)
- `d`: Delete selected colleague
- `f`: Toggle time format
- `t`: Toggle timeline mode
- `m` / `c`: Toggle individual↔shared / cycle color scheme (timeline only)
- `←/→`: Scrub time ±1h; Esc returns to now, second Esc exits (timeline only)
- `?`: Help; `q`/`Esc`: Quit
- Broken config entries stay visible as red `⚠ invalid timezone` rows and can be fixed with `e` or removed with `d`

**Configuration**
- `~/.config/tui-clock/config.yaml`, auto-created on first run, hot-reloaded on external edits
- `time_format`, `color_scheme`, `timeline_mode`, `location_display_format`
- Colleague fields `work_start`/`work_end`/`sleep_start`/`sleep_end`: optional, omitted = defaults (9-17 work, 23-7 sleep), 0 = midnight
- See `config.example.yaml`

### Timezone Search

- 200+ cities searchable by city, country, US state, or abbreviation (ambiguous abbreviations like CST list all matches); fuzzy, ranked by popularity and match quality
- Appended location label controlled by `location_display_format`: `auto` (city or abbreviation depending on what was searched), `city`, `timezone`, `abbreviation`
- Type to filter; `↑/↓` navigate (letters including `k`/`j` go to the query); Enter selects; each result shows its current time

### Timezone Handling

- `time.LoadLocation` against the embedded IANA database; validation on add/edit
- Offsets recomputed every tick, DST-aware; upcoming transitions found by `nextOffsetChange` (binary search on zone offset, ~530ns, runs per colleague per tick)
- Weekend = Saturday/Sunday in the colleague's zone

### UI Layout

**Normal Mode:**
```
🌍 World Clock - Local Time: 15:30:45 (Mon, Jan 20)

  ▲ 2 more above
▶ ● Alice (New York)  10:30:45  -5h  Mon, Jan 20  ⚡-1h Nov 1
  ○ Bob (London)      15:30:45  same  Mon, Jan 20
  ⚠ Carol  invalid timezone "Typo/Zone" — edit or delete
  ▼ 3 more below

↑/k up • ↓/j down • a add • e edit • w hours • d delete • f format • t timeline • ? help • q quit
```

**Timeline Mode (shared, with overlap row):**
```
🌍 Timeline View - Local Time: 15:30:45 (Mon, Jan 20)  ⏩ scrubbed +2h

Alice (New York)          10:30:45     [░░░░░░░░░░░░▓▓▓▓████████████████▓▓▓▓▓▓▓▓▓▓▓▓░░░░]
Ravi (Kolkata)            02:00:45     [░░░▓▓▓▓████████████████▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░]
Team overlap              1/2 now      [░░░░░░░▓▓▓▓▓███████████▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░]

                                       [0           6          12          18         24]

░ sleep • ▓ off-hours • █ work • █ now
overlap: █ everyone working • ▓ majority

t normal mode • m shared • ↑/↓ scroll • ←/→ scrub time • c cycle colors • ? help • q quit
```

The `⏩ scrubbed` header segment and the `esc back to now` footer hint appear only while scrubbed. The overlap row and its legend line appear only in shared mode with 2+ valid colleagues.

## Timeline Internals

**Bar characters**: `░` sleep, `▓` off-hours awake, `█` work. Classified by `barCharForHour(ct, hour float64)` — work wins over overlapping sleep; weekends drop work blocks. Positions map to fractional hours (`float64(i)/barWidth*24`), which keeps half/quarter-hour offsets exact at 2 chars/hour.

**Modes**: Individual bars are 0-24 in the colleague's zone (marker at their local time). Shared bars are 0-24 in the viewer's zone with the colleague's schedule shifted via `sharedBarHour(position, barWidth, offsetHours) float64` (marker at viewer's local time, same column for everyone). Both render Name + Time + Bar with shared hour labels at the bottom.

**Overlap row** (`renderOverlapRow`): `computeSharedOverlap` counts, per bar position, how many valid colleagues' `barCharForHour` is work. `█` (Success) where count == total, `▓` (Warning) where count ≥ half, `░` (Muted) otherwise; time field shows "N/M now" at the marker.

**Scrubbing**: `Model.timeOffset` shifts a virtual now. `displayNow()` gives the shifted local time; `scrubbed(ct)` shifts a `ColleagueTime` and recomputes `IsWeekend`/`IsWorkingTime` (scrubbing across midnight changes the weekday). All timeline rendering goes through these; leaving timeline mode resets the offset.

**Bar width**: `calculateTimelineBarWidth()` adapts to terminal width between `MinBarWidth` (24, 1 char/hour) and `IdealBarWidth` (48, 2 chars/hour — also the maximum), minus `NameFieldWidth` (25) + `TimeFieldWidth` (12) + padding.

**Color schemes**: defined solely in the `colorSchemes` map in styles.go; each provides bar colors (`SleepColor`, `AwakeOffColor`, `WorkColor`, `MarkerColor`, `WeekendTint`) and UI colors (`Primary`…`Muted`) as `lipgloss.TerminalColor` (plain or `AdaptiveColor`). Nord and Solarized are adaptive light/dark.

## Test Suite Map

- `inputs_test.go`: hour-range parsing, staged `w`-flow semantics (Esc cancels everything, blank Enter-Enter is a no-op), multi-rune paste handling, search navigation
- `timezone_test.go`: offsets (incl. +5.5h), ConfigIndex/invalid-entry flagging, overnight working hours, DST transition detection against fixed 2026 dates
- `timeline_test.go`: range checks incl. wraparound and fractional boundaries, bar precedence, shared-bar hour math, overlap counting with fixed zones, scrub flag recomputation, cell-width truncation (CJK/accented)
- `config_test.go`: defaults, round-trips, midnight preservation, legacy 0/0 sentinel migration
- `reload_test.go`: hot-reload — external pickup, own-save suppression, modal deferral, torn writes, deleted-file non-resurrection, same-size older-mtime (mutation pin), same-mtime size change, selection reset
