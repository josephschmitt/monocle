package tui

import (
	"testing"
)

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
