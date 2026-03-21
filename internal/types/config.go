package types

type Config struct {
	IgnorePatterns []string          `json:"ignore_patterns"`
	Keybindings    map[string]string `json:"keybindings"`
	DiffStyle      string            `json:"diff_style"`
	SidebarStyle   string            `json:"sidebar_style"`
	Layout         string            `json:"layout"`
	Theme          string            `json:"theme"`
	Wrap           bool              `json:"wrap"`
	TabSize        int               `json:"tab_size"`
	ContextLines   int               `json:"context_lines"`
	ReviewFormat   ReviewFormatConfig `json:"review_format"`
}

type ReviewFormatConfig struct {
	IncludeSnippets bool `json:"include_snippets"`
	MaxSnippetLines int  `json:"max_snippet_lines"`
	IncludeSummary  bool `json:"include_summary"`
}
