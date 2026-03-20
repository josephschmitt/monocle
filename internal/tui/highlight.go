package tui

import (
	"image/color"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// highlighter provides syntax highlighting for diff content using chroma.
type highlighter struct {
	lexerCache map[string]chroma.Lexer
	style      *chroma.Style
}

func newHighlighter() *highlighter {
	return &highlighter{
		lexerCache: make(map[string]chroma.Lexer),
		style:      styles.Get("monokai"),
	}
}

func (h *highlighter) getLexer(path string) chroma.Lexer {
	ext := filepath.Ext(path)
	if l, ok := h.lexerCache[ext]; ok {
		return l
	}
	l := lexers.Match(path)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	h.lexerCache[ext] = l
	return l
}

// changeRange represents a range of changed characters within a line (byte offsets).
type changeRange struct {
	start, end int
}

// highlightLine renders content with syntax highlighting and diff backgrounds.
// bg is the base background for the line (nil for context lines).
// changeBg is the background for changed characters (nil if no intra-line changes).
// changes specifies byte ranges within content that are changed.
// width is the visual width to pad to.
func (h *highlighter) highlightLine(path, content string, bg, changeBg color.Color, changes []changeRange, width int) string {
	lexer := h.getLexer(path)
	iter, err := lexer.Tokenise(nil, content)
	if err != nil {
		return h.renderPlain(content, bg, width)
	}

	var b strings.Builder
	pos := 0

	for _, tok := range iter.Tokens() {
		tokLen := len(tok.Value)
		tokEnd := pos + tokLen

		// Get syntax style for this token type
		entry := h.style.Get(tok.Type)

		// Render token, splitting at change boundaries
		segStart := pos
		for segStart < tokEnd {
			segEnd := tokEnd
			inChange := false

			for _, cr := range changes {
				if segStart >= cr.start && segStart < cr.end {
					// Inside a change range
					inChange = true
					if cr.end < segEnd {
						segEnd = cr.end
					}
					break
				}
				if cr.start > segStart && cr.start < segEnd {
					// Next change boundary splits this segment
					segEnd = cr.start
					break
				}
			}

			segText := tok.Value[segStart-pos : segEnd-pos]
			style := lipgloss.NewStyle()

			// Apply syntax foreground
			if entry.Colour.IsSet() {
				style = style.Foreground(lipgloss.Color(entry.Colour.String()))
			}
			if entry.Bold == chroma.Yes {
				style = style.Bold(true)
			}
			if entry.Italic == chroma.Yes {
				style = style.Italic(true)
			}

			// Apply background
			if inChange && changeBg != nil {
				style = style.Background(changeBg)
			} else if bg != nil {
				style = style.Background(bg)
			}

			b.WriteString(style.Render(segText))
			segStart = segEnd
		}

		pos = tokEnd
	}

	// Pad to fill width with background
	result := b.String()
	visWidth := lipgloss.Width(result)
	if visWidth < width {
		padStyle := lipgloss.NewStyle()
		if bg != nil {
			padStyle = padStyle.Background(bg)
		}
		result += padStyle.Render(strings.Repeat(" ", width-visWidth))
	}

	return result
}

// renderPlain renders content without syntax highlighting, just with background.
func (h *highlighter) renderPlain(content string, bg color.Color, width int) string {
	style := lipgloss.NewStyle()
	if bg != nil {
		style = style.Background(bg)
	}
	text := style.Render(content)
	visWidth := lipgloss.Width(text)
	if visWidth < width {
		text += style.Render(strings.Repeat(" ", width-visWidth))
	}
	return text
}

// computeChangeRanges finds changed character ranges between two lines
// using common prefix/suffix matching. Returns ranges for oldLine and newLine.
func computeChangeRanges(oldLine, newLine string) ([]changeRange, []changeRange) {
	if oldLine == newLine {
		return nil, nil
	}
	if oldLine == "" {
		return nil, []changeRange{{0, len(newLine)}}
	}
	if newLine == "" {
		return []changeRange{{0, len(oldLine)}}, nil
	}

	// Find common prefix
	prefixLen := 0
	minLen := len(oldLine)
	if len(newLine) < minLen {
		minLen = len(newLine)
	}
	for prefixLen < minLen && oldLine[prefixLen] == newLine[prefixLen] {
		prefixLen++
	}

	// Find common suffix (not overlapping prefix)
	suffixLen := 0
	oldRemain := len(oldLine) - prefixLen
	newRemain := len(newLine) - prefixLen
	minRemain := oldRemain
	if newRemain < minRemain {
		minRemain = newRemain
	}
	for suffixLen < minRemain &&
		oldLine[len(oldLine)-1-suffixLen] == newLine[len(newLine)-1-suffixLen] {
		suffixLen++
	}

	oldStart := prefixLen
	oldEnd := len(oldLine) - suffixLen
	newStart := prefixLen
	newEnd := len(newLine) - suffixLen

	var oldRanges, newRanges []changeRange
	if oldStart < oldEnd {
		oldRanges = []changeRange{{oldStart, oldEnd}}
	}
	if newStart < newEnd {
		newRanges = []changeRange{{newStart, newEnd}}
	}

	return oldRanges, newRanges
}

// clipChangeRanges clips change ranges to a maximum byte length.
func clipChangeRanges(changes []changeRange, maxLen int) []changeRange {
	var result []changeRange
	for _, cr := range changes {
		if cr.start >= maxLen {
			continue
		}
		clipped := cr
		if clipped.end > maxLen {
			clipped.end = maxLen
		}
		result = append(result, clipped)
	}
	return result
}
