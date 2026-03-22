package core

import (
	"strings"
	"testing"

	"github.com/anthropics/monocle/internal/types"
)

func defaultFormatCfg() types.ReviewFormatConfig {
	return types.ReviewFormatConfig{
		IncludeSnippets: true,
		MaxSnippetLines: 10,
		IncludeSummary:  true,
	}
}

func TestFormatNoComments(t *testing.T) {
	f := NewReviewFormatter(nil, defaultFormatCfg())
	result := f.Format(&types.ReviewSession{}, nil, types.ActionApprove, "")

	if result.CommentCount != 0 {
		t.Errorf("expected 0 comments, got %d", result.CommentCount)
	}
	if !strings.Contains(result.Formatted, "Approved") {
		t.Error("expected Approved in output")
	}
}

func TestFormatWithIssue(t *testing.T) {
	f := NewReviewFormatter(nil, defaultFormatCfg())
	comments := []types.ReviewComment{
		{
			ID:          "c1",
			TargetType:  types.TargetFile,
			TargetRef:   "src/auth/handler.ts",
			LineStart:   42,
			LineEnd:     45,
			Type:        types.CommentIssue,
			Body:        "This function doesn't handle the error case.",
			CodeSnippet: "func handle() {\n\terr := doSomething()\n}",
			ReviewRound: 1,
		},
	}

	result := f.Format(&types.ReviewSession{}, comments, types.ActionRequestChanges, "")

	if result.Action != string(types.ActionRequestChanges) {
		t.Errorf("expected request_changes, got %q", result.Action)
	}
	if !strings.Contains(result.Formatted, "[ISSUE]") {
		t.Error("expected [ISSUE] header")
	}
	if !strings.Contains(result.Formatted, "src/auth/handler.ts:42-45") {
		t.Error("expected file:line reference")
	}
	if !strings.Contains(result.Formatted, "Lines 42-45") {
		t.Error("expected line numbers in snippet")
	}
	if !strings.Contains(result.Formatted, "1 issue(s) to fix") {
		t.Error("expected issue count in summary")
	}
	if !strings.Contains(result.Formatted, "re-present your changes") {
		t.Error("expected re-present prompt")
	}
}

func TestFormatMixedTypes(t *testing.T) {
	f := NewReviewFormatter(nil, defaultFormatCfg())
	comments := []types.ReviewComment{
		{
			ID:         "c1",
			TargetType: types.TargetFile,
			TargetRef:  "main.go",
			LineStart:  10,
			Type:       types.CommentIssue,
			Body:       "Bug here",
		},
		{
			ID:         "c2",
			TargetType: types.TargetFile,
			TargetRef:  "main.go",
			LineStart:  20,
			Type:       types.CommentSuggestion,
			Body:       "Consider refactoring",
		},
		{
			ID:         "c3",
			TargetType: types.TargetContent,
			TargetRef:  "item-1",
			Type:       types.CommentNote,
			Body:       "Nice approach",
		},
	}

	result := f.Format(&types.ReviewSession{}, comments, types.ActionRequestChanges, "")

	if result.CommentCount != 3 {
		t.Errorf("expected 3 comments, got %d", result.CommentCount)
	}
	if !strings.Contains(result.Formatted, "[ISSUE]") {
		t.Error("missing ISSUE")
	}
	if !strings.Contains(result.Formatted, "[SUGGESTION]") {
		t.Error("missing SUGGESTION")
	}
	if !strings.Contains(result.Formatted, "[NOTE]") {
		t.Error("missing NOTE")
	}
	if !strings.Contains(result.Formatted, "Content: item-1") {
		t.Error("missing content item reference")
	}
}

func TestFormatOutdatedSkipped(t *testing.T) {
	f := NewReviewFormatter(nil, defaultFormatCfg())
	comments := []types.ReviewComment{
		{
			ID:         "c1",
			TargetType: types.TargetFile,
			TargetRef:  "main.go",
			Type:       types.CommentIssue,
			Body:       "Old bug",
			Outdated:   true,
		},
		{
			ID:         "c2",
			TargetType: types.TargetFile,
			TargetRef:  "main.go",
			Type:       types.CommentNote,
			Body:       "Current note",
		},
	}

	result := f.Format(&types.ReviewSession{}, comments, types.ActionApprove, "")

	if strings.Contains(result.Formatted, "Old bug") {
		t.Error("outdated comment should be skipped")
	}
	if !strings.Contains(result.Formatted, "Current note") {
		t.Error("current comment should be included")
	}
}

func TestFormatSnippetsDisabled(t *testing.T) {
	cfg := types.ReviewFormatConfig{
		IncludeSnippets: false,
		MaxSnippetLines: 10,
		IncludeSummary:  true,
	}
	f := NewReviewFormatter(nil, cfg)
	comments := []types.ReviewComment{
		{
			ID:          "c1",
			TargetType:  types.TargetFile,
			TargetRef:   "main.go",
			LineStart:   10,
			LineEnd:     12,
			Type:        types.CommentIssue,
			Body:        "Fix this",
			CodeSnippet: "func broken() {}",
		},
	}
	result := f.Format(&types.ReviewSession{}, comments, types.ActionRequestChanges, "")

	if strings.Contains(result.Formatted, "func broken()") {
		t.Error("snippet should not be included when IncludeSnippets=false")
	}
	if !strings.Contains(result.Formatted, "Fix this") {
		t.Error("comment body should still be included")
	}
	if !strings.Contains(result.Formatted, "[ISSUE]") {
		t.Error("comment header should still be included")
	}
}

func TestFormatSummaryDisabled(t *testing.T) {
	cfg := types.ReviewFormatConfig{
		IncludeSnippets: true,
		MaxSnippetLines: 10,
		IncludeSummary:  false,
	}
	f := NewReviewFormatter(nil, cfg)
	comments := []types.ReviewComment{
		{
			ID:         "c1",
			TargetType: types.TargetFile,
			TargetRef:  "main.go",
			Type:       types.CommentIssue,
			Body:       "Bug here",
		},
	}
	result := f.Format(&types.ReviewSession{}, comments, types.ActionRequestChanges, "")

	if strings.Contains(result.Formatted, "**Summary:**") {
		t.Error("summary should not be included when IncludeSummary=false")
	}
	if !strings.Contains(result.Formatted, "Bug here") {
		t.Error("comment body should still be included")
	}
}

func TestFormatMaxSnippetLines(t *testing.T) {
	cfg := types.ReviewFormatConfig{
		IncludeSnippets: true,
		MaxSnippetLines: 2,
		IncludeSummary:  true,
	}
	f := NewReviewFormatter(nil, cfg)
	comments := []types.ReviewComment{
		{
			ID:          "c1",
			TargetType:  types.TargetFile,
			TargetRef:   "main.go",
			LineStart:   1,
			LineEnd:     5,
			Type:        types.CommentIssue,
			Body:        "Too long",
			CodeSnippet: "line1\nline2\nline3\nline4\nline5\n",
		},
	}
	result := f.Format(&types.ReviewSession{}, comments, types.ActionRequestChanges, "")

	if !strings.Contains(result.Formatted, "line1") {
		t.Error("first line should be included")
	}
	if !strings.Contains(result.Formatted, "line2") {
		t.Error("second line should be included")
	}
	if strings.Contains(result.Formatted, "line3") {
		t.Error("third line should be truncated")
	}
	if !strings.Contains(result.Formatted, "// ... truncated") {
		t.Error("should have truncation indicator")
	}
}

func TestTruncateSnippet(t *testing.T) {
	tests := []struct {
		name     string
		snippet  string
		max      int
		wantTrun bool
	}{
		{"no limit", "a\nb\nc\n", 0, false},
		{"under limit", "a\nb\n", 3, false},
		{"at limit", "a\nb\nc\n", 3, false},
		{"over limit", "a\nb\nc\nd\n", 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateSnippet(tt.snippet, tt.max)
			hasTrunc := strings.Contains(result, "// ... truncated")
			if hasTrunc != tt.wantTrun {
				t.Errorf("truncated=%v, want %v; result=%q", hasTrunc, tt.wantTrun, result)
			}
		})
	}
}
