package tui

import (
	"strings"
	"testing"

	"github.com/josephschmitt/monocle/internal/types"
)

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		name  string
		hunks []types.DiffHunk
		want  bool
	}{
		{
			name:  "empty hunks",
			hunks: nil,
			want:  false,
		},
		{
			name: "normal text content",
			hunks: []types.DiffHunk{{
				Lines: []types.DiffLine{
					{Content: "func main() {"},
					{Content: "\tfmt.Println(\"hello\")"},
					{Content: "}"},
				},
			}},
			want: false,
		},
		{
			name: "null byte in content",
			hunks: []types.DiffHunk{{
				Lines: []types.DiffLine{
					{Content: "hello\x00world"},
				},
			}},
			want: true,
		},
		{
			name: "control character 0x01",
			hunks: []types.DiffHunk{{
				Lines: []types.DiffLine{
					{Content: "binary\x01data"},
				},
			}},
			want: true,
		},
		{
			name: "control character 0x1f",
			hunks: []types.DiffHunk{{
				Lines: []types.DiffLine{
					{Content: "data\x1fmore"},
				},
			}},
			want: true,
		},
		{
			name: "tab and newline are not binary",
			hunks: []types.DiffHunk{{
				Lines: []types.DiffLine{
					{Content: "line\twith\ttabs"},
				},
			}},
			want: false,
		},
		{
			name: "binary in second hunk",
			hunks: []types.DiffHunk{
				{Lines: []types.DiffLine{{Content: "normal text"}}},
				{Lines: []types.DiffLine{{Content: "has\x00null"}}},
			},
			want: true,
		},
		{
			name: "binary beyond sampling limit is missed",
			hunks: []types.DiffHunk{
				{Lines: []types.DiffLine{{Content: "normal"}}},
				{Lines: []types.DiffLine{{Content: "normal"}}},
				{Lines: []types.DiffLine{{Content: "has\x00null"}}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBinaryContent(tt.hunks)
			if got != tt.want {
				t.Errorf("isBinaryContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		width   int
		want    []string
	}{
		{
			name:    "fits within width",
			content: "hello world",
			width:   20,
			want:    []string{"hello world"},
		},
		{
			name:    "wraps at space boundary",
			content: "hello world foo",
			width:   12,
			want:    []string{"hello world ", "foo"},
		},
		{
			name:    "long word falls back to char wrap",
			content: "abcdefghijklmnop",
			width:   5,
			want:    []string{"abcde", "fghij", "klmno", "p"},
		},
		{
			name:    "mixed word wrap and char fallback",
			content: "hi abcdefghijklmno",
			width:   10,
			want:    []string{"hi ", "abcdefghij", "klmno"},
		},
		{
			name:    "empty string",
			content: "",
			width:   10,
			want:    []string{""},
		},
		{
			name:    "width zero returns as-is",
			content: "hello",
			width:   0,
			want:    []string{"hello"},
		},
		{
			name:    "negative width returns as-is",
			content: "hello",
			width:   -1,
			want:    []string{"hello"},
		},
		{
			name:    "exactly at width",
			content: "abcde",
			width:   5,
			want:    []string{"abcde"},
		},
		{
			name:    "break at last possible space",
			content: "aaa bbb ccc",
			width:   8,
			want:    []string{"aaa bbb ", "ccc"},
		},
		{
			name:    "leading indentation preserved",
			content: "    return nil",
			width:   10,
			want:    []string{"    return ", "nil"},
		},
		{
			name:    "multiple consecutive spaces",
			content: "a  b  c",
			width:   4,
			want:    []string{"a  b ", " c"},
		},
		{
			name:    "single character width",
			content: "abc",
			width:   1,
			want:    []string{"a", "b", "c"},
		},
		{
			name:    "space at exact boundary",
			content: "abcd efgh",
			width:   5,
			want:    []string{"abcd ", "efgh"},
		},
		{
			name:    "multiple wraps at word boundaries",
			content: "the quick brown fox jumps",
			width:   10,
			want:    []string{"the quick ", "brown fox ", "jumps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapContent(tt.content, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapContent(%q, %d) returned %d chunks, want %d\ngot:  %q\nwant: %q",
					tt.content, tt.width, len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("chunk[%d] = %q, want %q\nfull got:  %q\nfull want: %q",
						i, got[i], tt.want[i], got, tt.want)
				}
			}
		})
	}
}

func TestScreenLinesForConsistency(t *testing.T) {
	m := diffViewModel{
		wrap:  true,
		width: 50,
		lines: []diffViewLine{
			{content: "short line"},
			{content: "this is a longer line that should wrap at word boundaries when displayed"},
			{content: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"},
			{content: ""},
			{content: "    indented content with some extra words to wrap around"},
		},
	}

	for i, line := range m.lines {
		cw := m.contentWidthFor(line)
		expected := len(wrapContent(line.content, cw))
		got := m.screenLinesFor(i)
		if got != expected {
			t.Errorf("line %d: screenLinesFor=%d but len(wrapContent)=%d (content=%q, width=%d)",
				i, got, expected, line.content, cw)
		}
	}
}

func TestRenderWrappedLineMarkdownContent(t *testing.T) {
	theme := DefaultTheme()
	m := diffViewModel{
		theme:       &theme,
		hl:          newHighlighter(),
		mdStyler:    newMarkdownStyler(theme),
		contentMode: true,
		path:        "some-plan-id", // extensionless — content mode treats as markdown
		wrap:        true,
		width:       80,
	}

	tests := []struct {
		name    string
		content string
		// wantRaw is the raw markdown marker that should NOT appear in styled output
		wantRaw string
		// wantStyled is a substring that should appear in the styled output
		wantStyled string
	}{
		{
			name:       "header is styled",
			content:    "# Hello World",
			wantRaw:    "# ",
			wantStyled: "Hello World",
		},
		{
			name:       "bullet is styled",
			content:    "- list item",
			wantRaw:    "- ",
			wantStyled: "list item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := diffViewLine{content: tt.content, newLineNum: 1}
			result := m.renderWrappedLine("1   ", tt.content, 4, 76,
				nil, nil, false, &line)
			if strings.Contains(result, tt.wantRaw) {
				t.Errorf("expected raw markdown %q to be styled away, got: %s", tt.wantRaw, result)
			}
			if !strings.Contains(result, tt.wantStyled) {
				t.Errorf("expected styled output to contain %q, got: %s", tt.wantStyled, result)
			}
		})
	}
}

func TestRenderWrappedLineMarkdownFile(t *testing.T) {
	theme := DefaultTheme()
	m := diffViewModel{
		theme:       &theme,
		hl:          newHighlighter(),
		mdStyler:    newMarkdownStyler(theme),
		contentMode: false,
		path:        "README.md",
		wrap:        true,
		width:       80,
	}

	line := diffViewLine{content: "# Header", newLineNum: 1}
	result := m.renderWrappedLine("1   ", "# Header", 4, 76,
		nil, nil, false, &line)

	if strings.Contains(result, "# ") {
		t.Errorf("expected markdown header to be styled, got raw: %s", result)
	}
	if !strings.Contains(result, "Header") {
		t.Errorf("expected output to contain 'Header', got: %s", result)
	}
}

func TestRenderWrappedLineNonMarkdown(t *testing.T) {
	theme := DefaultTheme()
	m := diffViewModel{
		theme:       &theme,
		hl:          newHighlighter(),
		mdStyler:    newMarkdownStyler(theme),
		contentMode: false,
		path:        "main.go",
		wrap:        true,
		width:       80,
	}

	// "# comment" in a Go file should NOT be styled as a markdown header
	line := diffViewLine{content: "# comment", newLineNum: 1}
	result := m.renderWrappedLine("1   ", "# comment", 4, 76,
		nil, nil, false, &line)

	// The raw content should pass through (not transformed into a styled header)
	if !strings.Contains(result, "#") {
		t.Errorf("non-markdown file should preserve raw content, got: %s", result)
	}
}

func TestExtractSuggestionCode(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
		wantOK   bool
	}{
		{
			name:     "simple suggestion",
			body:     "```suggestion\nfoo := bar()\n```",
			wantCode: "foo := bar()",
			wantOK:   true,
		},
		{
			name:     "multi-line suggestion",
			body:     "```suggestion\nline1\nline2\nline3\n```",
			wantCode: "line1\nline2\nline3",
			wantOK:   true,
		},
		{
			name:     "suggestion with surrounding text",
			body:     "Consider this change:\n```suggestion\nnewCode()\n```\nThis is better.",
			wantCode: "newCode()",
			wantOK:   true,
		},
		{
			name:   "no suggestion block",
			body:   "This is a regular comment",
			wantOK: false,
		},
		{
			name:     "empty suggestion",
			body:     "```suggestion\n\n```",
			wantCode: "",
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := extractSuggestionCode(tt.body)
			if ok != tt.wantOK {
				t.Fatalf("extractSuggestionCode() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && code != tt.wantCode {
				t.Errorf("extractSuggestionCode() code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestFormatExpandedCommentSuggestionDiff(t *testing.T) {
	comment := &types.ReviewComment{
		Type: types.CommentSuggestion,
		Body: "```suggestion\nnewFunc()\n```",
	}

	// With original code: should render diff lines
	result := formatExpandedComment(comment, 80, "oldFunc()", true)
	if !strings.Contains(result, "- oldFunc()") {
		t.Errorf("expected diff to contain removed line '- oldFunc()', got:\n%s", result)
	}
	if !strings.Contains(result, "+ newFunc()") {
		t.Errorf("expected diff to contain added line '+ newFunc()', got:\n%s", result)
	}
	// Should NOT contain the raw fence markers
	if strings.Contains(result, "```suggestion") {
		t.Errorf("expected suggestion fence to be replaced by diff, got:\n%s", result)
	}

	// Without original code: should fall back to raw body
	resultNoOrig := formatExpandedComment(comment, 80, "", true)
	if !strings.Contains(resultNoOrig, "```suggestion") {
		t.Errorf("expected raw fence when no original code, got:\n%s", resultNoOrig)
	}
}

func TestFormatExpandedCommentSuggestionDiffWithSurroundingText(t *testing.T) {
	comment := &types.ReviewComment{
		Type: types.CommentSuggestion,
		Body: "Consider this:\n```suggestion\nnewFunc()\n```\nBetter approach.",
	}

	result := formatExpandedComment(comment, 80, "oldFunc()", true)
	if !strings.Contains(result, "Consider this:") {
		t.Errorf("expected text before suggestion, got:\n%s", result)
	}
	if !strings.Contains(result, "- oldFunc()") {
		t.Errorf("expected removed line, got:\n%s", result)
	}
	if !strings.Contains(result, "+ newFunc()") {
		t.Errorf("expected added line, got:\n%s", result)
	}
	if !strings.Contains(result, "Better approach.") {
		t.Errorf("expected text after suggestion, got:\n%s", result)
	}
}

func TestOriginalCodeForComment(t *testing.T) {
	m := diffViewModel{
		lines: []diffViewLine{
			{newLineNum: 1, content: "line one"},
			{newLineNum: 2, content: "line two"},
			{newLineNum: 3, content: "line three"},
			{newLineNum: 4, content: "line four"},
			{isComment: true, comment: &types.ReviewComment{ID: "c1"}},
			{newLineNum: 5, content: "line five"},
		},
	}

	comment := &types.ReviewComment{LineStart: 2, LineEnd: 4}
	got := m.originalCodeForComment(comment)
	want := "line two\nline three\nline four"
	if got != want {
		t.Errorf("originalCodeForComment() = %q, want %q", got, want)
	}

	// Single line
	comment2 := &types.ReviewComment{LineStart: 3, LineEnd: 0}
	got2 := m.originalCodeForComment(comment2)
	if got2 != "line three" {
		t.Errorf("single line: got %q, want %q", got2, "line three")
	}

	// File-level comment (LineStart=0)
	comment3 := &types.ReviewComment{LineStart: 0}
	got3 := m.originalCodeForComment(comment3)
	if got3 != "" {
		t.Errorf("file-level comment: got %q, want empty", got3)
	}
}

func TestRenderContentLineWrapModeMarkdown(t *testing.T) {
	theme := DefaultTheme()
	m := diffViewModel{
		theme:       &theme,
		hl:          newHighlighter(),
		mdStyler:    newMarkdownStyler(theme),
		contentMode: true,
		path:        "plan-id",
		wrap:        true,
		width:       80,
		focused:     true,
	}

	line := diffViewLine{content: "## Section Title", newLineNum: 1}
	result := m.renderContentLine(line, 0, 76, false, false)

	if strings.Contains(result, "## ") {
		t.Errorf("expected markdown header to be styled in wrap mode, got raw: %s", result)
	}
	if !strings.Contains(result, "Section Title") {
		t.Errorf("expected output to contain 'Section Title', got: %s", result)
	}
}

func TestLineNumAtSplitMode(t *testing.T) {
	m := diffViewModel{
		lines: []diffViewLine{
			// Split-mode added line: new-file number lives in rightLineNum.
			{rightLineNum: 42, newLineNum: 0, content: "added"},
			// Unified/file-mode line: new-file number lives in newLineNum.
			{rightLineNum: 0, newLineNum: 7, content: "context"},
			// Removed-only split line: neither side has a new-file number.
			{rightLineNum: 0, newLineNum: 0, content: "removed"},
		},
	}

	// Split line resolves to the right (new-file) line number, so the comment
	// editor opens against line 42 instead of silently no-opping.
	if got := m.lineNumAt(0); got != 42 {
		t.Errorf("lineNumAt(0) = %d, want 42", got)
	}
	m.cursor = 0
	if got := m.currentDiffLine(); got != 42 {
		t.Errorf("currentDiffLine() = %d, want 42", got)
	}

	// Unified line still resolves via newLineNum (regression guard).
	if got := m.lineNumAt(1); got != 7 {
		t.Errorf("lineNumAt(1) = %d, want 7", got)
	}

	// Removed-only split line stays non-commentable.
	if got := m.lineNumAt(2); got != 0 {
		t.Errorf("lineNumAt(2) = %d, want 0", got)
	}
}

// TestIsSelectableSplitInPlaceChange verifies that an in-place change in split
// mode — built as a removed line (kind == DiffLineRemoved, newLineNum == 0)
// paired with an added right side (rightLineNum > 0) — is selectable so the
// cursor can land on it and `c` opens the comment editor. A pure removed split
// line (rightLineNum == 0) must stay non-selectable.
func TestIsSelectableSplitInPlaceChange(t *testing.T) {
	m := diffViewModel{
		lines: []diffViewLine{
			// In-place split change: removed left, added right.
			{isSplit: true, kind: types.DiffLineRemoved, newLineNum: 0, rightLineNum: 11, content: "old", rightContent: "new"},
			// Pure removed split line: no right side.
			{isSplit: true, kind: types.DiffLineRemoved, newLineNum: 0, rightLineNum: 0, rightEmpty: true, content: "gone"},
			// Plain added split line.
			{isSplit: true, kind: types.DiffLineContext, leftEmpty: true, newLineNum: 0, rightLineNum: 12, rightContent: "added"},
		},
	}

	// The in-place change is selectable and resolves to its new-file line number.
	if !m.isSelectable(0) {
		t.Errorf("isSelectable(0) = false, want true for in-place split change")
	}
	if got := m.lineNumAt(0); got != 11 {
		t.Errorf("lineNumAt(0) = %d, want 11", got)
	}
	m.cursor = 0
	if got := m.currentDiffLine(); got != 11 {
		t.Errorf("currentDiffLine() = %d, want 11", got)
	}

	// A pure removed line (no right-side line number) stays non-selectable.
	if m.isSelectable(1) {
		t.Errorf("isSelectable(1) = true, want false for pure removed split line")
	}

	// nextSelectable from the in-place change skips the pure removed line and
	// lands on the next line that has a right-side number.
	if got := m.nextSelectable(0, 1); got != 2 {
		t.Errorf("nextSelectable(0, 1) = %d, want 2 (skipping pure removed line)", got)
	}
}
