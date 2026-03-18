package core

import (
	"strings"
	"testing"

	"github.com/anthropics/monocle/internal/types"
)

func TestFormatNoComments(t *testing.T) {
	f := NewReviewFormatter(nil)
	result := f.Format(&types.ReviewSession{}, nil)

	if result.CommentCount != 0 {
		t.Errorf("expected 0 comments, got %d", result.CommentCount)
	}
	if !strings.Contains(result.Formatted, "Approved") {
		t.Error("expected Approved in output")
	}
}

func TestFormatWithIssue(t *testing.T) {
	f := NewReviewFormatter(nil)
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

	result := f.Format(&types.ReviewSession{}, comments)

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
	f := NewReviewFormatter(nil)
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

	result := f.Format(&types.ReviewSession{}, comments)

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
	f := NewReviewFormatter(nil)
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

	result := f.Format(&types.ReviewSession{}, comments)

	if strings.Contains(result.Formatted, "Old bug") {
		t.Error("outdated comment should be skipped")
	}
	if !strings.Contains(result.Formatted, "Current note") {
		t.Error("current comment should be included")
	}
}
