package tui

import (
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
)

// markdownStyler applies lightweight line-preserving styling to markdown content.
// Each source line maps to exactly one styled output line, preserving line numbers
// for commenting.
type markdownStyler struct {
	theme Theme
}

func newMarkdownStyler(theme Theme) *markdownStyler {
	return &markdownStyler{theme: theme}
}

var (
	// Block-level patterns
	h1Pattern         = regexp.MustCompile(`^(#{1})\s+(.*)$`)
	h2Pattern         = regexp.MustCompile(`^(#{2})\s+(.*)$`)
	h3Pattern         = regexp.MustCompile(`^(#{3,})\s+(.*)$`)
	bulletPattern     = regexp.MustCompile(`^(\s*)([-*])\s+(.*)$`)
	numberedPattern   = regexp.MustCompile(`^(\s*)(\d+\.)\s+(.*)$`)
	blockquotePattern = regexp.MustCompile(`^>\s?(.*)$`)
	codeFencePattern  = regexp.MustCompile("^```(\\w*)\\s*$")
	rulePattern       = regexp.MustCompile(`^[-*_]{3,}\s*$`)

	// Inline patterns
	boldPattern       = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicPattern     = regexp.MustCompile(`(?:^|[^*])\*([^*]+?)\*(?:[^*]|$)`)
	inlineCodePattern = regexp.MustCompile("`([^`]+)`")
)

// StyleLine processes a single non-code-block markdown line and returns styled content.
// Code blocks and fences are handled by the caller (renderContentLine) since they
// need access to the syntax highlighter.
func (ms *markdownStyler) StyleLine(line string) string {
	// Horizontal rule
	if rulePattern.MatchString(line) {
		return ms.theme.MarkdownRule.Render(strings.Repeat("─", 40))
	}

	// Headers (check h3 before h2 before h1 since h3 pattern is more specific)
	if m := h3Pattern.FindStringSubmatch(line); m != nil {
		return ms.theme.MarkdownH3.Render(ms.styleInline(m[2]))
	}
	if m := h2Pattern.FindStringSubmatch(line); m != nil {
		return ms.theme.MarkdownH2.Render(ms.styleInline(m[2]))
	}
	if m := h1Pattern.FindStringSubmatch(line); m != nil {
		return ms.theme.MarkdownH1.Render(ms.styleInline(m[2]))
	}

	// Blockquote
	if m := blockquotePattern.FindStringSubmatch(line); m != nil {
		content := ms.styleInline(m[1])
		return ms.theme.MarkdownBullet.Render("│") + " " + ms.theme.MarkdownBlockquote.Render(content)
	}

	// Bullet list
	if m := bulletPattern.FindStringSubmatch(line); m != nil {
		indent := m[1]
		content := ms.styleInline(m[3])
		return indent + ms.theme.MarkdownBullet.Render("•") + " " + content
	}

	// Numbered list
	if m := numberedPattern.FindStringSubmatch(line); m != nil {
		indent := m[1]
		num := m[2]
		content := ms.styleInline(m[3])
		return indent + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(num) + " " + content
	}

	// Regular line: apply inline styling only
	return ms.styleInline(line)
}

// styleInline applies bold, italic, and inline code styling within a line.
func (ms *markdownStyler) styleInline(line string) string {
	// Process inline code first (to prevent bold/italic from matching inside code)
	line = inlineCodePattern.ReplaceAllStringFunc(line, func(match string) string {
		m := inlineCodePattern.FindStringSubmatch(match)
		if m == nil {
			return match
		}
		return ms.theme.MarkdownCode.Render(m[1])
	})

	// Bold
	line = boldPattern.ReplaceAllStringFunc(line, func(match string) string {
		m := boldPattern.FindStringSubmatch(match)
		if m == nil {
			return match
		}
		return lipgloss.NewStyle().Bold(true).Render(m[1])
	})

	// Italic — more careful to not interfere with bold markers
	line = italicPattern.ReplaceAllStringFunc(line, func(match string) string {
		m := italicPattern.FindStringSubmatch(match)
		if m == nil {
			return match
		}
		// Preserve surrounding characters that were captured by the lookaround
		prefix := ""
		suffix := ""
		if len(match) > 0 && match[0] != '*' {
			prefix = string(match[0])
		}
		if len(match) > 0 && match[len(match)-1] != '*' {
			suffix = string(match[len(match)-1])
		}
		return prefix + lipgloss.NewStyle().Italic(true).Render(m[1]) + suffix
	})

	return line
}

// isMarkdownContent returns true if the path or content type indicates markdown.
// Accepts paths like "content.md", extensions like "md" or ".md", or empty string
// (plans default to markdown when no content type is specified).
func isMarkdownContent(pathOrType string) bool {
	// No extension means it's a plan ID with no content type — treat as markdown
	if !strings.Contains(pathOrType, ".") {
		return true
	}
	lower := strings.ToLower(pathOrType)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}
