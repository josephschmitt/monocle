package core

import (
	"fmt"
	"strings"

	"github.com/anthropics/monocle/internal/types"
)

// ContentProvider is a callback to get file content for code snippets.
type ContentProvider func(path string, start, end int) string

// ReviewFormatter formats review comments into structured markdown.
type ReviewFormatter struct {
	getContent ContentProvider
}

// NewReviewFormatter creates a formatter with a content provider callback.
func NewReviewFormatter(getContent ContentProvider) *ReviewFormatter {
	return &ReviewFormatter{getContent: getContent}
}

// Format produces a FormattedReview from a session and its comments.
// The action is explicitly provided by the caller (user-selected status).
// The body is an optional general review comment included at the top.
func (rf *ReviewFormatter) Format(session *types.ReviewSession, comments []types.ReviewComment, action types.SubmitAction, body string) *FormattedReview {
	hasComments := false
	for _, c := range comments {
		if !c.Outdated {
			hasComments = true
			break
		}
	}

	if !hasComments && strings.TrimSpace(body) == "" {
		header := "## Code Review — Approved\n\nNo issues found. Code looks good!"
		if action == types.ActionRequestChanges {
			header = "## Code Review — Changes Requested\n\nNo specific comments."
		}
		return &FormattedReview{
			Formatted:    header,
			CommentCount: 0,
			Action:       string(action),
		}
	}

	var b strings.Builder

	// Header
	switch action {
	case types.ActionRequestChanges:
		b.WriteString("## Code Review — Changes Requested\n\n")
	default:
		b.WriteString("## Code Review — Feedback\n\n")
	}

	// General review body
	if trimmed := strings.TrimSpace(body); trimmed != "" {
		b.WriteString(trimmed)
		b.WriteString("\n\n")
		if hasComments {
			b.WriteString("---\n\n")
		}
	}

	// Count by type
	issueCt, suggestionCt, noteCt, praiseCt := countByType(comments)

	// Group comments by target
	fileComments := map[string][]types.ReviewComment{}
	contentComments := map[string][]types.ReviewComment{}
	for _, c := range comments {
		if c.Outdated {
			continue
		}
		switch c.TargetType {
		case types.TargetFile:
			fileComments[c.TargetRef] = append(fileComments[c.TargetRef], c)
		case types.TargetContent:
			contentComments[c.TargetRef] = append(contentComments[c.TargetRef], c)
		}
	}

	// File comments
	for path, cmts := range fileComments {
		for _, c := range cmts {
			lineRef := ""
			if c.LineStart > 0 {
				if c.LineEnd > c.LineStart {
					lineRef = fmt.Sprintf(":%d-%d", c.LineStart, c.LineEnd)
				} else {
					lineRef = fmt.Sprintf(":%d", c.LineStart)
				}
			}

			typeLabel := strings.ToUpper(string(c.Type))
			b.WriteString(fmt.Sprintf("### [%s] %s%s\n", typeLabel, path, lineRef))

			// Code snippet
			if c.CodeSnippet != "" {
				b.WriteString("```\n")
				if c.LineStart > 0 {
					b.WriteString(fmt.Sprintf("// Lines %d-%d:\n", c.LineStart, c.LineEnd))
				}
				b.WriteString(c.CodeSnippet)
				if !strings.HasSuffix(c.CodeSnippet, "\n") {
					b.WriteString("\n")
				}
				b.WriteString("```\n")
			} else if rf.getContent != nil && c.LineStart > 0 {
				end := c.LineEnd
				if end == 0 {
					end = c.LineStart
				}
				snippet := rf.getContent(path, c.LineStart, end)
				if snippet != "" {
					b.WriteString("```\n")
					b.WriteString(fmt.Sprintf("// Lines %d-%d:\n", c.LineStart, end))
					b.WriteString(snippet)
					if !strings.HasSuffix(snippet, "\n") {
						b.WriteString("\n")
					}
					b.WriteString("```\n")
				}
			}

			b.WriteString(c.Body)
			b.WriteString("\n\n---\n\n")
		}
	}

	// Content item comments (plans, docs) — with line references and snippets
	for itemID, cmts := range contentComments {
		// Find the content item title from session
		itemTitle := ""
		for _, item := range session.ContentItems {
			if item.ID == itemID {
				itemTitle = item.Title
				break
			}
		}

		for _, c := range cmts {
			typeLabel := strings.ToUpper(string(c.Type))

			lineRef := ""
			if c.LineStart > 0 {
				if c.LineEnd > c.LineStart {
					lineRef = fmt.Sprintf(":%d-%d", c.LineStart, c.LineEnd)
				} else {
					lineRef = fmt.Sprintf(":%d", c.LineStart)
				}
			}

			// Use "Plan: Title" if we have a title, otherwise "Content: itemID"
			var header string
			if itemTitle != "" {
				header = fmt.Sprintf("### [%s] Plan: %s%s\n", typeLabel, itemTitle, lineRef)
			} else {
				header = fmt.Sprintf("### [%s] Content: %s%s\n", typeLabel, itemID, lineRef)
			}
			b.WriteString(header)

			// Snippet from content item
			if c.CodeSnippet != "" {
				b.WriteString("```\n")
				if c.LineStart > 0 {
					b.WriteString(fmt.Sprintf("// Lines %d-%d:\n", c.LineStart, c.LineEnd))
				}
				b.WriteString(c.CodeSnippet)
				if !strings.HasSuffix(c.CodeSnippet, "\n") {
					b.WriteString("\n")
				}
				b.WriteString("```\n")
			}

			b.WriteString(c.Body)
			b.WriteString("\n\n---\n\n")
		}
	}

	// Summary (only if there are inline comments)
	if hasComments {
		b.WriteString("**Summary:** ")
		parts := []string{}
		if issueCt > 0 {
			parts = append(parts, fmt.Sprintf("%d issue(s) to fix", issueCt))
		}
		if suggestionCt > 0 {
			parts = append(parts, fmt.Sprintf("%d suggestion(s) to consider", suggestionCt))
		}
		if noteCt > 0 {
			parts = append(parts, fmt.Sprintf("%d note(s)", noteCt))
		}
		if praiseCt > 0 {
			parts = append(parts, fmt.Sprintf("%d praise", praiseCt))
		}
		b.WriteString(strings.Join(parts, ", "))
		b.WriteString(".\n")

		if issueCt > 0 {
			b.WriteString("Please address the issues and re-present your changes.\n")
		}
	}

	return &FormattedReview{
		Formatted:    b.String(),
		CommentCount: len(comments),
		Action:       string(action),
	}
}

func countByType(comments []types.ReviewComment) (issue, suggestion, note, praise int) {
	for _, c := range comments {
		if c.Outdated {
			continue
		}
		switch c.Type {
		case types.CommentIssue:
			issue++
		case types.CommentSuggestion:
			suggestion++
		case types.CommentNote:
			note++
		case types.CommentPraise:
			praise++
		}
	}
	return
}
