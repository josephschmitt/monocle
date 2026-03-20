package tui

import "testing"

func TestComputeChangeRanges(t *testing.T) {
	tests := []struct {
		name      string
		oldLine   string
		newLine   string
		wantOld   []changeRange
		wantNew   []changeRange
	}{
		{
			name:    "identical lines",
			oldLine: "hello world",
			newLine: "hello world",
		},
		{
			name:    "empty old",
			oldLine: "",
			newLine: "hello",
			wantNew: []changeRange{{0, 5}},
		},
		{
			name:    "empty new",
			oldLine: "hello",
			newLine: "",
			wantOld: []changeRange{{0, 5}},
		},
		{
			name:    "single char change in middle",
			oldLine: "hello world",
			newLine: "hello World",
			wantOld: []changeRange{{6, 7}},
			wantNew: []changeRange{{6, 7}},
		},
		{
			name:    "word replacement",
			oldLine: "func foo() int {",
			newLine: "func bar() int {",
			wantOld: []changeRange{{5, 8}},
			wantNew: []changeRange{{5, 8}},
		},
		{
			name:    "suffix change",
			oldLine: "return nil",
			newLine: "return err",
			wantOld: []changeRange{{7, 10}},
			wantNew: []changeRange{{7, 10}},
		},
		{
			name:    "prefix change",
			oldLine: "var x = 1",
			newLine: "const x = 1",
			wantOld: []changeRange{{0, 3}},
			wantNew: []changeRange{{0, 5}},
		},
		{
			name:    "insertion",
			oldLine: "ab",
			newLine: "aXb",
			wantNew: []changeRange{{1, 2}},
		},
		{
			name:    "deletion",
			oldLine: "aXb",
			newLine: "ab",
			wantOld: []changeRange{{1, 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOld, gotNew := computeChangeRanges(tt.oldLine, tt.newLine)
			if !rangesEqual(gotOld, tt.wantOld) {
				t.Errorf("old ranges = %v, want %v", gotOld, tt.wantOld)
			}
			if !rangesEqual(gotNew, tt.wantNew) {
				t.Errorf("new ranges = %v, want %v", gotNew, tt.wantNew)
			}
		})
	}
}

func TestClipChangeRanges(t *testing.T) {
	ranges := []changeRange{{5, 15}, {20, 25}}

	clipped := clipChangeRanges(ranges, 10)
	if len(clipped) != 1 || clipped[0].start != 5 || clipped[0].end != 10 {
		t.Errorf("clipped = %v, want [{5 10}]", clipped)
	}

	clipped = clipChangeRanges(ranges, 22)
	if len(clipped) != 2 || clipped[1].end != 22 {
		t.Errorf("clipped = %v, want [{5 15} {20 22}]", clipped)
	}
}

func TestHighlighterTokenizes(t *testing.T) {
	h := newHighlighter()

	// Should not panic and should return non-empty result
	result := h.highlightLine("test.go", "func main() {}", nil, nil, nil, 30)
	if result == "" {
		t.Error("expected non-empty highlighted result")
	}

	// Unknown file type should still work
	result = h.highlightLine("unknown.xyz", "some content", nil, nil, nil, 20)
	if result == "" {
		t.Error("expected non-empty result for unknown file type")
	}
}

func rangesEqual(a, b []changeRange) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
