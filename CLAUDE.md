# Monocle

Terminal-based code review tool for the agentic coding era. Runs alongside AI coding agents (Claude Code, Codex CLI, Gemini CLI) and provides a structured, bidirectional code review workflow between human and agent via lifecycle hooks.

## Quick Reference

- **Language**: TypeScript 5.7+ (strict mode, ES modules)
- **Runtime**: Node.js (ES2022 target)
- **TUI Framework**: Rezi
- **Database**: SQLite via better-sqlite3 (WAL mode, foreign keys)
- **CLI Framework**: Commander.js
- **Git**: simple-git wrapper
- **Testing**: Vitest
- **Build**: `npm run build` (tsc)
- **Dev**: `npm run dev` (tsx)
- **Test**: `npm run test` (vitest)

## Architecture

Three independent layers — the core protocol and logic are language-agnostic. The TUI frontend and hook shim are replaceable.

```
Agent Terminal                    Monocle Terminal
┌─────────────────┐              ┌──────────────────────────┐
│ Claude / Codex  │              │  TUI Frontend (Rezi)     │
│ / Gemini CLI    │              │         │                │
│                 │── UDS ──────►│  Core Engine             │
│ Hooks fire on:  │   (NDJSON)   │  - Hook Server (UDS)     │
│ Pre/PostToolUse │              │  - Review State Machine  │
│ Stop            │◄── UDS ─────│  - Git Integration       │
│ PromptSubmit    │              │  - Session Persistence   │
└─────────────────┘              └──────────────────────────┘
```

**Socket**: Unix Domain Socket at `/tmp/monocle-<session-id>.sock`
**Protocol**: Newline-delimited JSON (NDJSON)

### Layer Boundaries

| Layer | Directory | Responsibility |
|-------|-----------|----------------|
| **Types** | `src/types/` | Protocol messages, review data model, config schema |
| **Core Engine** | `src/core/` | State machine, review logic, session persistence, git ops, config |
| **Database** | `src/db/` | SQLite schema, migrations, typed query functions |
| **Adapters** | `src/adapters/` | Agent-specific translation (Claude, Gemini, Codex, OpenCode) + socket client |
| **Hook Shim** | `src/hooks/` | Thin CLI subprocess: stdin → adapter → socket → adapter → stdout |
| **TUI Frontend** | `src/tui/` | Rezi screens, components, keybindings. **Must not contain business logic.** |

### Key Constraint: Frontend-Agnostic Core

The `core/`, `db/`, and `types/` directories **must not import anything from `src/tui/`**. All business logic lives in the core — the TUI is a thin rendering layer over it. This enables future web UI or alternative TUI implementations.

Similarly, `adapters/` is shared between the hook shim and OpenCode plugin — it must not depend on the TUI or core engine internals beyond the protocol types.

## Path Aliases

Configured in tsconfig.json:

- `@core/*` → `src/core/*`
- `@adapters/*` → `src/adapters/*`
- `@db/*` → `src/db/*`
- `@types/*` → `src/types/*`
- `@tui/*` → `src/tui/*`
- `@hooks/*` → `src/hooks/*`

Always use these aliases in imports rather than relative paths.

## Review State Machine

Two modes with distinct state flows:

**Batch Review (Mode 2, default):**
```
IDLE → WATCHING → REVIEWING → SUBMITTED → WATCHING → ...
```
- IDLE: No active agent session
- WATCHING: Agent is working, PostToolUse events streaming in, file list updates live
- REVIEWING: Agent hit Stop hook, agent is blocked, user reviews diffs and leaves comments
- SUBMITTED: Feedback formatted and sent to agent via StopResponse, returns to WATCHING

**Gate Mode (Mode 1):**
```
WATCHING → GATING → WATCHING → ...
```
- GATING: PreToolUse fired for a gated file pattern, agent blocked, user approves/denies

**Hybrid**: Combines both — gate_patterns trigger GATING, everything else flows through batch review.

## Data Model

Core entities (defined in `src/types/review.ts`):

- **ReviewSession**: Mode, state, agent type, repo root, base ref, gate/ignore patterns
- **ChangedFile**: Path, status (added/modified/deleted/renamed), hunks, review status
- **ContentItem**: Non-diff content (plans, proposals) piped via `monocle review`. Has id, title, content, content_type
- **ReviewComment**: Anchored to file or content item lines. Types: issue, suggestion, note, praise
- **ReviewSubmission**: The formatted markdown review sent to the agent

## Protocol Messages

**Inbound (Hook → Monocle)**: `pre_tool_use`, `post_tool_use`, `stop`, `prompt_submit`, `content_review`
**Outbound (Monocle → Hook)**: `pre_tool_use_response`, `stop_response`, `prompt_submit_response`

Blocking semantics:
- `pre_tool_use`: Blocking. Hook waits for allow/deny response.
- `post_tool_use`: Non-blocking. Fire and forget.
- `stop`: Blocking. Response includes continue flag and optional system_message with review comments.
- `prompt_submit`: Blocking. Response includes optional additional_context.

If socket is unreachable, hooks exit 0 (allow) — agent must never be permanently blocked.

## Review Formatter Output

Comments are formatted as structured markdown for agent consumption:

```markdown
## Code Review — Changes Requested
### [ISSUE] src/auth/handler.ts:42-45
<code snippet>
<comment body>
---
### [SUGGESTION] src/api/routes.ts:18
...
**Summary:** N issues to fix, N suggestions to consider...
```

- Include code snippets inline so agent has context without re-reading files
- ISSUE vs SUGGESTION vs NOTE vs PRAISE gives clear priority signals
- Content item comments use title instead of file path: `### [ISSUE] Implementation Plan:15-22`

## TUI Screens

1. **Dashboard** — Session overview, file list with change stats. Entry point.
2. **Diff Browser** — Main review interface. File tree sidebar + diff/content pane. Supports unified and side-by-side (`t` to toggle). Inline comments shown expanded by default (`z` to collapse).
3. **Comment Editor** — Modal overlay. Type selector (issue/suggestion/note/praise), multiline body, code snippet preview.
4. **Review Summary** — Pre-submission confirmation. Comments grouped by file/content item.
5. **Gate Prompt** — Quick approve/deny for PreToolUse events. Minimal UI, speed-focused.
6. **Config** — Mode toggle, gate patterns, ignore patterns.

### Keybindings (vim-inspired)

| Key | Action |
|-----|--------|
| `j`/`k` | Navigate / scroll |
| `Ctrl-d`/`Ctrl-u` | Page down / up |
| `[`/`]` | Previous / next file |
| `c` | Comment on current line(s) |
| `v`/`V` | Visual mode for line selection |
| `r` | Mark file reviewed |
| `t` | Toggle unified / side-by-side |
| `z` | Toggle expand / collapse comments |
| `a` | Approve (gate mode) |
| `d` | Deny with comment (gate mode) |
| `:submit` | Open review summary |
| `:approve` | Approve all and end review |

### Real-Time Behavior

In WATCHING state, the file list updates live as PostToolUse events arrive. New files get a `~` indicator that fades. Diff view refreshes only when viewing an actively modified file. Never interrupt the user's reading.

## CLI Interface

```
monocle                           # Start TUI with auto-detected session
monocle start [--mode] [--agent]  # Start new session
monocle setup <agent>             # Install hooks (claude|gemini|codex|opencode)
monocle review [--id] [--title] <file|->  # Pipe content for review
monocle resume [session-id]       # Resume session
monocle sessions [--repo]         # List sessions
```

## Configuration

- Global: `~/.config/monocle/config.json` (XDG)
- Project: `.monocle/config.json`
- Database: `~/.local/share/monocle/monocle.db` (XDG)

Config merges: defaults → global → project. Validated with helpful error messages.

## Testing Conventions

- Test runner: Vitest with globals enabled
- Test files: `test/<module>/<name>.test.ts` mirroring `src/<module>/<name>.ts`
- Fixtures: `test/fixtures/diffs/` and `test/fixtures/hook-payloads/`
- Database tests use in-memory SQLite (`:memory:`)
- Run all tests: `npm test`
- Run specific: `npx vitest run test/core/config.test.ts`

## Implementation Status

**Implemented:**
- `src/types/` — Full data model and config types
- `src/core/config.ts` — XDG-compliant config loading with validation
- `src/db/schema.ts` — SQLite schema with WAL mode
- `src/db/queries.ts` — Complete typed CRUD query layer

**Stub (comment placeholders only):**
- `src/core/` — git.ts, hook-server.ts, review-formatter.ts, review-state.ts, session.ts
- `src/adapters/` — All adapters and socket-client
- `src/hooks/` — monocle-hook.ts, bundle-plugin.ts, templates
- `src/tui/` — All screens and components
- `src/index.ts` — CLI entry point

## Design Principles

- **The agent is a coworker, not a tool.** The interaction model is PR review, not configuration.
- **Don't own the agent lifecycle.** Monocle observes via hooks, never wraps or spawns the agent.
- **Hooks are the universal interface.** Build on the shared hook abstraction across all CLI agents.
- **Terminal-native.** Users are already in the terminal. Don't make them leave.
- **Graceful degradation.** Socket unreachable → hooks allow. No hooks → standalone diff viewer. Unknown agent → generic adapter.
