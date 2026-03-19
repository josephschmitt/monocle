package types

type Config struct {
	DefaultAgent   string            `json:"default_agent"`
	IgnorePatterns []string          `json:"ignore_patterns"`
	Keybindings    map[string]string `json:"keybindings"`
	DiffStyle      string            `json:"diff_style"`
	SidebarStyle   string            `json:"sidebar_style"`
	Theme          string            `json:"theme"`
	ReviewFormat   ReviewFormatConfig `json:"review_format"`
}

type ReviewFormatConfig struct {
	IncludeSnippets bool `json:"include_snippets"`
	MaxSnippetLines int  `json:"max_snippet_lines"`
	IncludeSummary  bool `json:"include_summary"`
}
