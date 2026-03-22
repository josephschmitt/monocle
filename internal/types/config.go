package types

type Config struct {
	IgnorePatterns []string `json:"ignore_patterns"`
	DiffStyle      string   `json:"diff_style"`
	SidebarStyle   string   `json:"sidebar_style"`
	Layout         string   `json:"layout"`
	Wrap           bool     `json:"wrap"`
	TabSize        int      `json:"tab_size"`
	ContextLines   int      `json:"context_lines"`
}
