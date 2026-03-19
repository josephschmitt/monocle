# Monocle

Terminal-based code review companion for AI coding agents (Claude Code, Codex, Gemini CLI). Developers run it alongside their agent — the agent writes code, the developer reviews diffs and leaves structured feedback, and Monocle delivers that feedback to the agent via skills.

## Quick Start

```bash
devbox shell                          # Sets up Go + lefthook
devbox run -- make build              # Build binary → bin/
devbox run -- make test               # Run tests
devbox run -- make lint               # Vet + build check
```

**Always use `devbox run --` for Go commands.** Never use the global `go` binary.

## Architecture

Single binary with CLI subcommands for agent communication:
- **`monocle`** — TUI + CLI (Kong). Manages sessions, renders diffs/plans, collects comments, delivers reviews.
- **`monocle review-status`** — Check for pending feedback (invoked by agent via skill).
- **`monocle get-feedback [--wait]`** — Retrieve review feedback (invoked by agent via skill).
- **`monocle submit-content --title TITLE`** — Submit plans/docs for review (invoked by agent via skill).
- **`monocle install`** — Install skill files for detected agents.

### Integration Model: Skills

Agents integrate with Monocle via **skills** (SKILL.md files), not hooks. The agent auto-invokes the skill at natural breakpoints, which instructs it to run `monocle` CLI subcommands. CLI subcommands communicate with the TUI via a Unix domain socket.

**Key design:**
- **Asynchronous by default** — agent polls for feedback, never blocked unless reviewer requests a pause
- **User-initiated review** — reviewer works at their own pace, submits when ready
- **Pause flow** — reviewer can request a pause; agent sees "pause_requested" on next status check and blocks on `get-feedback --wait`

### Package Layout

```
cmd/monocle/          Main CLI entry point (Kong commands)
internal/
  types/              Domain types (ReviewSession, ChangedFile, ReviewComment, Config)
  protocol/           NDJSON message types + marshal/unmarshal (GetReviewStatus, PollFeedback, SubmitContent)
  db/                 SQLite layer (schema, migrations, typed queries)
  core/               Engine, git client, feedback queue, formatter, session manager, socket server
  adapters/           Agent-specific skill installers (Claude, Gemini, Codex, OpenCode), socket client
  tui/                Bubble Tea v2 UI (app shell, sidebar, diff view, plan view, modals, theme)
```

### Key Interfaces

- **`core.EngineAPI`** (`internal/core/engine.go`) — Contract between TUI and engine. TUI never imports engine internals.
- **`adapters.SkillInstaller`** (`internal/adapters/adapter.go`) — Agent-specific skill file installation/uninstallation.

### Data Flow

```
Agent invokes skill → runs monocle CLI subcommand → Unix socket → SocketServer → Engine
Engine → emits events → BridgeEngineEvents → tea.Program.Send() → TUI updates
User submits review → Engine → FeedbackQueue → agent polls via CLI → gets formatted feedback
```

### Pause Flow

```
User presses P in TUI → Engine.RequestPause() → sets pause flag
Agent runs `monocle review-status` → sees "pause_requested"
Agent runs `monocle get-feedback --wait` → blocks until user submits
User reviews, adds comments, submits → FeedbackQueue releases → agent gets feedback
```

## Tech Stack

- **Go** (1.23 via devbox, module requires 1.25+)
- **Bubble Tea v2** — TUI framework. Uses `tea.Model` interface, `tea.View` struct (not string), `tea.KeyPressMsg` (not KeyMsg). Alt-screen set via `v.AltScreen = true` in View().
- **Lipgloss v2** — Styling. `lipgloss.Color()` is a function returning `color.Color`, not a type.
- **Bubbles v2** — UI components (key bindings)
- **Kong** — CLI parsing (not Cobra)
- **modernc.org/sqlite** — Pure Go SQLite (no CGo)
- **16-color ANSI** base theme for terminal compatibility, with true color for icons

## Bubble Tea v2 Gotchas

- `KeyPressMsg.String()` returns `"esc"` not `"escape"`, `"enter"` not `"return"`
- `View()` returns `tea.View` struct, not `string`
- `tea.Program` is not generic (no type parameter)
- `tea.Quit` is a `func() Msg`, usable directly as a `tea.Cmd`

## Conventions

- **Error handling**: Wrap with context: `fmt.Errorf("description: %w", err)`
- **Tests**: White-box, co-located in same package. Use `t.TempDir()` for isolation.
- **DB tests**: Use `:memory:` SQLite
- **Git tests**: Create temp repos with `setupTestRepo(t)`
- **Nerd Font icons**: Glyphs render wider than `lipgloss.Width()` measures. Use `iconSlack` compensation in layout math.
- **Conventional commits**: **All commit messages MUST use conventional commit format.** Release-please uses these to determine version bumps and generate changelogs.
  - `feat: ...` — New feature (minor version bump)
  - `fix: ...` — Bug fix (patch version bump)
  - `chore: ...` — Maintenance, deps, CI (no release)
  - `refactor: ...` — Code restructuring (no release)
  - `docs: ...` — Documentation only (no release)
  - `test: ...` — Test changes (no release)
  - `feat!: ...` or `BREAKING CHANGE:` in body — Breaking change (major version bump)
  - Scope is optional: `feat(tui): ...`, `fix(db): ...`

## Common Tasks

### Add a new TUI component
1. Create `internal/tui/yourcomponent.go` with a model struct + `Init`/`Update`/`View`
2. Define message types for communication
3. Wire into `appModel` in `app.go` (add field, init in `NewApp`, handle messages in `Update`, render in `View`)

### Add a new agent adapter
1. Create `internal/adapters/youragent.go` implementing `SkillInstaller`
2. Register in `AllInstallers()` in `adapter.go`

### Add a new CLI command
1. Add a struct to `cmd/monocle/main.go` with Kong tags
2. Add it as a field on the `CLI` struct
3. Implement `Run() error` method

### Add a new DB table
1. Add DDL to `schemaSQL` in `internal/db/schema.go`
2. Bump `schemaVersion`
3. Add query functions to `queries.go`
4. Add tests to `db_test.go`

## Release Process

Automated via release-please + goreleaser:
1. Push conventional commits to `main`
2. Release-please creates/updates a release PR
3. Merge the PR → tag is created
4. Goreleaser builds linux/darwin/windows (amd64+arm64), publishes to GitHub Releases + Homebrew tap
