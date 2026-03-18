package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/monocle/internal/types"
)

// CreateSession inserts a new review session.
func (d *DB) CreateSession(s *types.ReviewSession) error {
	patterns, _ := json.Marshal(s.IgnorePatterns)
	_, err := d.Exec(
		`INSERT INTO sessions (id, agent, agent_status, repo_root, base_ref, ignore_patterns, review_round, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Agent, string(s.AgentStatus), s.RepoRoot, s.BaseRef, string(patterns), s.ReviewRound, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

// GetSession retrieves a session by ID.
func (d *DB) GetSession(id string) (*types.ReviewSession, error) {
	s := &types.ReviewSession{}
	var agentStatus, patterns string
	err := d.QueryRow(
		`SELECT id, agent, agent_status, repo_root, base_ref, ignore_patterns, review_round, created_at, updated_at
		 FROM sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.Agent, &agentStatus, &s.RepoRoot, &s.BaseRef, &patterns, &s.ReviewRound, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	s.AgentStatus = types.AgentStatus(agentStatus)
	json.Unmarshal([]byte(patterns), &s.IgnorePatterns)
	s.FileStatuses = make(map[string]bool)
	return s, nil
}

// UpdateSession updates mutable session fields.
func (d *DB) UpdateSession(s *types.ReviewSession) error {
	patterns, _ := json.Marshal(s.IgnorePatterns)
	_, err := d.Exec(
		`UPDATE sessions SET agent_status = ?, review_round = ?, ignore_patterns = ?, updated_at = ? WHERE id = ?`,
		string(s.AgentStatus), s.ReviewRound, string(patterns), time.Now(), s.ID,
	)
	return err
}

// ListSessions returns session summaries, optionally filtered by repo root.
func (d *DB) ListSessions(repoRoot string, limit int) ([]types.SessionSummary, error) {
	query := `SELECT s.id, s.agent, s.repo_root, s.review_round, s.created_at, s.updated_at,
			  (SELECT COUNT(*) FROM changed_files WHERE session_id = s.id) as file_count,
			  (SELECT COUNT(*) FROM comments WHERE session_id = s.id) as comment_count
			  FROM sessions s`
	var args []any
	if repoRoot != "" {
		query += " WHERE s.repo_root = ?"
		args = append(args, repoRoot)
	}
	query += " ORDER BY s.updated_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.SessionSummary
	for rows.Next() {
		var s types.SessionSummary
		if err := rows.Scan(&s.ID, &s.Agent, &s.RepoRoot, &s.ReviewRound, &s.CreatedAt, &s.UpdatedAt, &s.FileCount, &s.CommentCount); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// UpsertChangedFile inserts or updates a changed file record.
func (d *DB) UpsertChangedFile(sessionID string, f *types.ChangedFile) error {
	_, err := d.Exec(
		`INSERT INTO changed_files (session_id, path, status, reviewed)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(session_id, path) DO UPDATE SET status = excluded.status`,
		sessionID, f.Path, string(f.Status), boolToInt(f.Reviewed),
	)
	return err
}

// GetChangedFiles returns all changed files for a session.
func (d *DB) GetChangedFiles(sessionID string) ([]types.ChangedFile, error) {
	rows, err := d.Query(
		`SELECT path, status, reviewed FROM changed_files WHERE session_id = ? ORDER BY path`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []types.ChangedFile
	for rows.Next() {
		var f types.ChangedFile
		var status string
		var reviewed int
		if err := rows.Scan(&f.Path, &status, &reviewed); err != nil {
			return nil, err
		}
		f.Status = types.FileChangeStatus(status)
		f.Reviewed = reviewed != 0
		files = append(files, f)
	}
	return files, rows.Err()
}

// MarkFileReviewed sets the reviewed flag for a file.
func (d *DB) MarkFileReviewed(sessionID, path string, reviewed bool) error {
	_, err := d.Exec(
		`UPDATE changed_files SET reviewed = ? WHERE session_id = ? AND path = ?`,
		boolToInt(reviewed), sessionID, path,
	)
	return err
}

// UpsertContentItem inserts or updates a content item.
func (d *DB) UpsertContentItem(sessionID string, item *types.ContentItem) error {
	_, err := d.Exec(
		`INSERT INTO content_items (id, session_id, title, content, content_type, reviewed, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET title = excluded.title, content = excluded.content, content_type = excluded.content_type, updated_at = excluded.updated_at`,
		item.ID, sessionID, item.Title, item.Content, item.ContentType, boolToInt(item.Reviewed), item.CreatedAt, item.UpdatedAt,
	)
	return err
}

// GetContentItems returns all content items for a session.
func (d *DB) GetContentItems(sessionID string) ([]types.ContentItem, error) {
	rows, err := d.Query(
		`SELECT id, title, content, content_type, reviewed, created_at, updated_at
		 FROM content_items WHERE session_id = ? ORDER BY created_at`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.ContentItem
	for rows.Next() {
		var item types.ContentItem
		var reviewed int
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.ContentType, &reviewed, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Reviewed = reviewed != 0
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetContentItem returns a single content item by ID.
func (d *DB) GetContentItem(id string) (*types.ContentItem, error) {
	item := &types.ContentItem{}
	var reviewed int
	err := d.QueryRow(
		`SELECT id, title, content, content_type, reviewed, created_at, updated_at
		 FROM content_items WHERE id = ?`, id,
	).Scan(&item.ID, &item.Title, &item.Content, &item.ContentType, &reviewed, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	item.Reviewed = reviewed != 0
	return item, nil
}

// CreateComment inserts a new comment.
func (d *DB) CreateComment(sessionID string, c *types.ReviewComment) error {
	_, err := d.Exec(
		`INSERT INTO comments (id, session_id, target_type, target_ref, line_start, line_end, type, body, code_snippet, resolved, outdated, review_round, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, sessionID, string(c.TargetType), c.TargetRef, c.LineStart, c.LineEnd,
		string(c.Type), c.Body, c.CodeSnippet, boolToInt(c.Resolved), boolToInt(c.Outdated),
		c.ReviewRound, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

// UpdateComment updates a comment's body and updated_at.
func (d *DB) UpdateComment(c *types.ReviewComment) error {
	_, err := d.Exec(
		`UPDATE comments SET body = ?, updated_at = ? WHERE id = ?`,
		c.Body, time.Now(), c.ID,
	)
	return err
}

// DeleteComment removes a comment by ID.
func (d *DB) DeleteComment(id string) error {
	_, err := d.Exec(`DELETE FROM comments WHERE id = ?`, id)
	return err
}

// GetComments returns all comments for a session, optionally filtered.
func (d *DB) GetComments(sessionID string) ([]types.ReviewComment, error) {
	rows, err := d.Query(
		`SELECT id, target_type, target_ref, line_start, line_end, type, body, code_snippet, resolved, outdated, review_round, created_at, updated_at
		 FROM comments WHERE session_id = ? ORDER BY created_at`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []types.ReviewComment
	for rows.Next() {
		var c types.ReviewComment
		var targetType, commentType string
		var resolved, outdated int
		if err := rows.Scan(&c.ID, &targetType, &c.TargetRef, &c.LineStart, &c.LineEnd, &commentType,
			&c.Body, &c.CodeSnippet, &resolved, &outdated, &c.ReviewRound, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.TargetType = types.TargetType(targetType)
		c.Type = types.CommentType(commentType)
		c.Resolved = resolved != 0
		c.Outdated = outdated != 0
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// MarkOutdated marks all non-outdated comments in the session as outdated.
func (d *DB) MarkOutdated(sessionID string) error {
	_, err := d.Exec(
		`UPDATE comments SET outdated = 1, updated_at = ? WHERE session_id = ? AND outdated = 0`,
		time.Now(), sessionID,
	)
	return err
}

// DismissOutdated deletes all outdated comments in the session.
func (d *DB) DismissOutdated(sessionID string) error {
	_, err := d.Exec(`DELETE FROM comments WHERE session_id = ? AND outdated = 1`, sessionID)
	return err
}

// CreateSubmission inserts a review submission record.
func (d *DB) CreateSubmission(sessionID string, sub *types.ReviewSubmission) error {
	_, err := d.Exec(
		`INSERT INTO review_submissions (id, session_id, action, formatted_review, comment_count, review_round, submitted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sub.ID, sessionID, string(sub.Action), sub.FormattedReview, sub.CommentCount, sub.ReviewRound, sub.SubmittedAt,
	)
	return err
}

// GetSubmissions returns all submissions for a session.
func (d *DB) GetSubmissions(sessionID string) ([]types.ReviewSubmission, error) {
	rows, err := d.Query(
		`SELECT id, session_id, action, formatted_review, comment_count, review_round, submitted_at
		 FROM review_submissions WHERE session_id = ? ORDER BY submitted_at`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []types.ReviewSubmission
	for rows.Next() {
		var s types.ReviewSubmission
		var action string
		if err := rows.Scan(&s.ID, &s.SessionID, &action, &s.FormattedReview, &s.CommentCount, &s.ReviewRound, &s.SubmittedAt); err != nil {
			return nil, err
		}
		s.Action = types.SubmitAction(action)
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Ensure sql.ErrNoRows is accessible for callers without importing database/sql directly.
var ErrNoRows = sql.ErrNoRows
