package types

type Config struct {
	IgnorePatterns []string          `json:"ignore_patterns"`
	Keybindings    map[string]string `json:"keybindings"`
	DiffStyle      string            `json:"diff_style"`
	SidebarStyle   string            `json:"sidebar_style"`
	Layout         string            `json:"layout"`
	Wrap           bool              `json:"wrap"`
	TabSize        int               `json:"tab_size"`
	ContextLines   int               `json:"context_lines"`
	ReviewFormat      ReviewFormatConfig `json:"review_format"`
	AutoFocusMode     bool               `json:"auto_focus_mode"`
	Mouse              *bool `json:"mouse"`
	MinDiffWidth       int   `json:"min_diff_width"`
	CommentExpand         *bool  `json:"comment_expand"`
	CommentExpandDelay    int    `json:"comment_expand_delay"`
	MarkReviewedOnSubmit  string `json:"mark_reviewed_on_submit"` // "all" (default), "commented", "manual"
}

type ReviewFormatConfig struct {
	IncludeSnippets bool `json:"include_snippets"`
	MaxSnippetLines int  `json:"max_snippet_lines"`
	IncludeSummary  bool `json:"include_summary"`
}
