package types

type Config struct {
	DefaultAgent   string            `json:"default_agent"`
	IgnorePatterns []string          `json:"ignore_patterns"`
	Keybindings    map[string]string `json:"keybindings"`
	DiffStyle      string            `json:"diff_style"`
	SidebarStyle   string            `json:"sidebar_style"`
	Theme          string            `json:"theme"`
	ReviewFormat   ReviewFormatConfig `json:"review_format"`
	Hooks          HooksConfig       `json:"hooks"`
}

type ReviewFormatConfig struct {
	IncludeSnippets bool `json:"include_snippets"`
	MaxSnippetLines int  `json:"max_snippet_lines"`
	IncludeSummary  bool `json:"include_summary"`
}

type HooksConfig struct {
	SocketPath string `json:"socket_path"`
	TimeoutMs  int    `json:"timeout_ms"`
	Scope      string `json:"scope"` // "repo" (default) or "cwd"
}
