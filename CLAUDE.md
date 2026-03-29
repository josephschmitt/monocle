# Monocle

Terminal-based code review companion for AI coding agents. Developers run it alongside their agent — the agent writes code, the developer reviews diffs and leaves structured feedback, and Monocle delivers that feedback via CLI commands and push notifications.

## Quick Start

```bash
devbox shell                          # Sets up Go + lefthook
devbox run -- make build              # Build binary → bin/
devbox run -- make test               # Run tests
devbox run -- make lint               # Vet + build check
```

**Always use `devbox run --` for Go commands.** Never use the global `go` binary.

## Architecture

Single binary with CLI subcommands:
- **`monocle`** — TUI (Kong). Manages sessions, renders diffs/plans, collects comments, delivers reviews.
- **`monocle review`** — Agent-facing CLI commands: `status`, `get-feedback`, `send-artifact`, `add-files`.
- **`monocle register`** — Install skills and (for Claude) MCP channel config.
- **`monocle unregister`** — Remove skills and MCP channel config.
- **`monocle serve-mcp-channel`** — (hidden) Run the MCP channel server for push notifications. Called by Claude Code, not users.

### Integration Model: CLI + Push Notifications

Agents interact with Monocle via **CLI commands** (`monocle review ...`) for pull-based operations and **MCP channel notifications** (Claude Code only) for push-based events.

- **CLI commands** — `monocle review status`, `get-feedback`, `send-artifact`, `add-files` connect to the engine's Unix socket, send a request, print the response, and exit. All agents use these via skills.
- **MCP channel** (Claude Code only) — a slim stdio MCP server that forwards push notifications (`feedback_submitted`, `pause_requested`) as channel events. No tools — all operations are CLI commands.
- **Skills** — standardized `SKILL.md` files (agentskills.io format) embedded in the binary and installed by `monocle register`. Skills instruct agents to run the CLI commands.

**Key design:**
- **Push+pull** — Claude Code gets push notifications via MCP channel, then runs CLI commands to retrieve feedback. Other agents use skills to poll.
- **User-initiated review** — reviewer works at their own pace, submits when ready
- **Pause flow** — reviewer can request a pause; agent runs `monocle review get-feedback --wait` to block

### Package Layout

```
cmd/monocle/          Main CLI entry point (Kong commands, including monocle review subcommands)
channel/              MCP channel source (TypeScript) + esbuild bundling (push notifications only)
skills/               Embedded SKILL.md files (agentskills.io format) shared by all agents
internal/
  types/              Domain types (ReviewSession, ChangedFile, ReviewComment, Config)
  protocol/           NDJSON message types + marshal/unmarshal (GetReviewStatus, PollFeedback, SubmitContent)
  client/             Socket client for CLI commands (connects to engine socket)
  db/                 SQLite layer (schema, migrations, typed queries)
  core/               Engine, git client, feedback queue, formatter, session manager, socket server
  adapters/           Agent registration, skill installation, repo/socket utilities
  tui/                Bubble Tea v2 UI (app shell, sidebar, diff view, plan view, modals, theme)
```

### Key Interfaces

- **`core.EngineAPI`** (`internal/core/engine.go`) — Contract between TUI and engine. TUI never imports engine internals.

### Data Flow

```
Agent runs `monocle review send-artifact` → CLI → Unix socket → SocketServer → Engine
Agent runs `monocle review get-feedback` → CLI → Unix socket → SocketServer → Engine
Engine → emits events → BridgeEngineEvents → tea.Program.Send() → TUI updates
User submits review → Engine → FeedbackQueue → channel.ts sends MCP notification → Agent
Agent runs `monocle review get-feedback` → retrieves feedback
```

### Pause Flow

```
User presses P in TUI → Engine.RequestPause() → sets pause flag
Agent runs `monocle review status` → sees "pause_requested"
Agent runs `monocle review get-feedback --wait` → blocks until user submits
User reviews, adds comments, submits → FeedbackQueue releases → notification sent → Agent retrieves
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

## Documentation

**Keep README.md up to date when adding or changing user-facing features.** Specifically:

- **Keybindings** — If you add, remove, or change a keybinding in the TUI, update both `internal/tui/help.go` and the keybinding table in `README.md`.
- **Config options** — If you add or change a field in `types.Config`, update the Configuration section in `README.md` (the example JSON, the settings table, and defaults in `internal/core/config.go`).
- **CLI commands** — If you add or change a subcommand, update the CLI section in `README.md`.
- **Features list** — If you add a significant new capability, add it to the Features bullet list in `README.md`.

## Common Tasks

### Add a new TUI component
1. Create `internal/tui/yourcomponent.go` with a model struct + `Init`/`Update`/`View`
2. Define message types for communication
3. Wire into `appModel` in `app.go` (add field, init in `NewApp`, handle messages in `Update`, render in `View`)

### Add a new CLI command
1. Add a struct to `cmd/monocle/main.go` with Kong tags
2. Add it as a field on the `CLI` struct
3. Implement `Run() error` method

### Add a new DB table
1. Add DDL to `schemaSQL` in `internal/db/schema.go`
2. Bump `schemaVersion`
3. Add query functions to `queries.go`
4. Add tests to `db_test.go`

## Monocle Integration

When Monocle is running, use the `/review-plan` or `/review-plan-wait` skills to send plans for review. These skills run CLI commands under the hood:
- `monocle review send-artifact --title "..." --file <path> --id <filename>` — send content for the reviewer to see
- `monocle review send-artifact --title "..." --file <path> --id <filename> --wait` — send and block until the reviewer responds
- `monocle review get-feedback` — retrieve pending feedback
- `monocle review status` — check if feedback is pending or a pause was requested

## Release Process

Automated via release-please + goreleaser:
1. Push conventional commits to `main`
2. Release-please creates/updates a release PR
3. Merge the PR → tag is created
4. Goreleaser builds linux/darwin/windows (amd64+arm64), publishes to GitHub Releases + Homebrew tap
