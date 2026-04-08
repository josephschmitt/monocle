# Documentation project instructions

## About this project

- This is the Monocle documentation site built on [Mintlify](https://mintlify.com)
- **Monocle is a local CLI/TUI application**, not a web service or SaaS product. There are no REST APIs, HTTP endpoints, authentication flows, API keys, webhooks, or server deployments. All communication happens over local Unix sockets between processes on the same machine.
- Pages are MDX files with YAML frontmatter in `docs/`
- Configuration lives in `docs/docs.json`
- Run `mint dev` from the `docs/` directory to preview locally
- Run `mint broken-links` to check links

## Terminology

- **monocle** — lowercase in prose, not "Monocle" (except at sentence start)
- **skills** — the agentskills.io `SKILL.md` files, not "plugins" or "tools"
- **register** / **unregister** — not "install" / "uninstall" (those are deprecated)
- **push notifications** — via MCP channels (Claude Code only)
- **pull-based feedback** — all other agents polling via `/get-feedback`
- **artifact** — plans, architecture docs, or other non-file content sent via `send-artifact`
- **review** / **submit** — the user's batch of comments, not individual comments

## Style preferences

- Use active voice and second person ("you")
- Keep sentences concise — one idea per sentence
- Use sentence case for headings
- Bold for UI elements: Click **Settings**
- Code formatting for file names, commands, paths, and code references
- Use backtick-quoted key names for keybindings: `S`, `Ctrl+g`, `Space`

## Keeping docs in sync with the codebase

**When you change any of the following in the codebase, update the corresponding docs pages:**

### Keybindings
- **Source of truth**: `internal/tui/keys.go` (KeyMap struct + defaults), `internal/tui/help.go` (help overlay sections)
- **Docs to update**:
  - `docs/reference/keybindings.mdx` — complete keyboard shortcut reference
  - `docs/configuration/keybindings.mdx` — available actions table for custom overrides

### Config options
- **Source of truth**: `internal/types/config.go` (Config struct), `internal/core/config.go` (DefaultConfig)
- **Docs to update**:
  - `docs/configuration/config-file.mdx` — example JSON and settings table

### CLI commands and flags
- **Source of truth**: `cmd/monocle/main.go` (Kong CLI structs)
- **Docs to update**:
  - `docs/reference/cli.mdx` — top-level commands (`monocle`, `register`, `unregister`)
  - `docs/reference/agent-commands.mdx` — `monocle review` subcommands (`status`, `get-feedback`, `send-artifact`, `add-files`)

### Skills
- **Source of truth**: `skills/` directory (embedded `SKILL.md` files)
- **Docs to update**:
  - `docs/concepts/agent-integration.mdx` — agent integration modes, available skills, and commands

### Supported agents
- **Source of truth**: `internal/adapters/` (agent adapter implementations)
- **Docs to update**:
  - `docs/guides/agent-setup.mdx` — per-agent setup tabs and config paths table
  - `docs/index.mdx` — supported agents table
  - `docs/introduction.mdx` — agent list

### TUI features (diff viewer, commenting, review flow)
- **Docs to update**:
  - `docs/concepts/review-loop.mdx` — diff modes, comment types, feedback queue, submit flow
  - `docs/concepts/review-state.mdx` — review state tracking, snapshots, rounds, change detection
  - `docs/guides/plan-review.mdx` — plan/artifact review, version history, focus mode
  - `docs/guides/review-gating.mdx` — plan gating and pause flow
  - `docs/guides/sessions.mdx` — session management, file tracking, commands

### Navigation structure
- **Source of truth**: `docs/docs.json` (Mintlify nav config)
- If you add a new page, add it to the appropriate group in `docs.json`

## Content boundaries

- Document user-facing features only — not internal architecture or implementation details
- Do not document deprecated commands (`monocle install`, `monocle uninstall`) — they exist for backwards compatibility only
- The `monocle serve-mcp-channel` command is hidden/internal and should not be documented
- **Do not create API reference pages.** Monocle has no REST API, GraphQL API, or HTTP endpoints. The `monocle review` subcommands are CLI commands, not API calls — document them as such in `reference/agent-commands.mdx`.
- **Do not use Mintlify API-oriented components** like `<ParamField>`, `<ResponseField>`, `<Expandable>`, or OpenAPI/Swagger integration. These are for web APIs and do not apply to this project.
- **Do not reference authentication, API keys, tokens, rate limits, or SDKs.** Monocle uses no authentication — it connects to a local Unix socket.
- **Do not create "endpoint" or "request/response" style documentation.** CLI flags and arguments should be documented in plain tables, not as API parameters.
