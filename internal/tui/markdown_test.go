package tui

import (
	"strings"
	"testing"
)

func TestStyleLineHeaders(t *testing.T) {
	ms := newMarkdownStyler(DefaultTheme())

	tests := []struct {
		name  string
		input string
		want  string // text content after stripping markers
	}{
		{"h1", "# Title", "Title"},
		{"h2", "## Subtitle", "Subtitle"},
		{"h3", "### Section", "Section"},
		{"h4 uses h3 style", "#### Deep", "Deep"},
		{"not a header", "#no space", "#no space"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ms.StyleLine(tt.input)
			if !strings.Contains(result, tt.want) {
				t.Errorf("StyleLine(%q) = %q, want it to contain %q", tt.input, result, tt.want)
			}
			// Should not contain the raw marker
			if strings.Contains(tt.input, "# ") && strings.HasPrefix(result, "#") {
				t.Errorf("StyleLine(%q) should strip # markers", tt.input)
			}
		})
	}
}

func TestStyleLineBullets(t *testing.T) {
	ms := newMarkdownStyler(DefaultTheme())

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"dash bullet", "- item", "•"},
		{"star bullet", "* item", "•"},
		{"indented bullet", "  - nested", "•"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ms.StyleLine(tt.input)
			if !strings.Contains(result, tt.want) {
				t.Errorf("StyleLine(%q) = %q, want bullet %q", tt.input, result, tt.want)
			}
			if !strings.Contains(result, "item") && !strings.Contains(result, "nested") {
				t.Errorf("StyleLine(%q) should contain item text", tt.input)
			}
		})
	}
}

func TestStyleLineBlockquote(t *testing.T) {
	ms := newMarkdownStyler(DefaultTheme())

	result := ms.StyleLine("> quoted text")
	if !strings.Contains(result, "│") {
		t.Errorf("blockquote should have │ border, got %q", result)
	}
	if !strings.Contains(result, "quoted text") {
		t.Errorf("blockquote should contain text, got %q", result)
	}
}

func TestStyleLineRule(t *testing.T) {
	ms := newMarkdownStyler(DefaultTheme())

	for _, rule := range []string{"---", "***", "___"} {
		result := ms.StyleLine(rule)
		if !strings.Contains(result, "─") {
			t.Errorf("StyleLine(%q) should render as horizontal rule, got %q", rule, result)
		}
	}
}

func TestStyleLineNumberedList(t *testing.T) {
	ms := newMarkdownStyler(DefaultTheme())

	result := ms.StyleLine("1. First item")
	if !strings.Contains(result, "1.") {
		t.Errorf("numbered list should keep number, got %q", result)
	}
	if !strings.Contains(result, "First item") {
		t.Errorf("numbered list should keep text, got %q", result)
	}
}

func TestStyleInline(t *testing.T) {
	ms := newMarkdownStyler(DefaultTheme())

	tests := []struct {
		name    string
		input   string
		want    string
		notWant string
	}{
		{"inline code", "use `fmt.Println`", "fmt.Println", "`"},
		{"bold", "this is **bold** text", "bold", "**"},
		{"plain text", "no formatting here", "no formatting here", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ms.styleInline(tt.input)
			if !strings.Contains(result, tt.want) {
				t.Errorf("styleInline(%q) = %q, want %q", tt.input, result, tt.want)
			}
			if tt.notWant != "" && strings.Contains(result, tt.notWant) {
				t.Errorf("styleInline(%q) should not contain raw %q", tt.input, tt.notWant)
			}
		})
	}
}

func TestIsMarkdownContent(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"content.md", true},
		{"content.markdown", true},
		{"content.MD", true},
		{"content.go", false},
		{"content.ts", false},
		{"some-uuid-id", true}, // no extension = plan default
		{"content.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isMarkdownContent(tt.input)
			if got != tt.want {
				t.Errorf("isMarkdownContent(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsMarkdownFile(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"README.md", true},
		{"docs/guide.markdown", true},
		{"CHANGELOG.MD", true},
		{"src/main.go", false},
		{"Makefile", false},      // extensionless — must NOT match
		{"LICENSE", false},       // extensionless — must NOT match
		{"some-uuid-id", false},  // unlike isMarkdownContent, this is false
		{"Dockerfile", false},
		{"notes.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isMarkdownFile(tt.input)
			if got != tt.want {
				t.Errorf("isMarkdownFile(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
