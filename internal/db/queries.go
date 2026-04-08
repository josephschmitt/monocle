package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/josephschmitt/monocle/internal/types"
)

// CreateSession inserts a new review session.
func (d *DB) CreateSession(s *types.ReviewSession) error {
	patterns, _ := json.Marshal(s.IgnorePatterns)
	_, err := d.Exec(
		`INSERT INTO sessions (id, agent, repo_root, base_ref, ignore_patterns, review_round, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Agent, s.RepoRoot, s.BaseRef, string(patterns), s.ReviewRound, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

// GetSession retrieves a session by ID.
func (d *DB) GetSession(id string) (*types.ReviewSession, error) {
	s := &types.ReviewSession{}
	var patterns string
	err := d.QueryRow(
		`SELECT id, agent, repo_root, base_ref, ignore_patterns, review_round, created_at, updated_at
		 FROM sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.Agent, &s.RepoRoot, &s.BaseRef, &patterns, &s.ReviewRound, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(patterns), &s.IgnorePatterns)
	s.FileStatuses = make(map[string]bool)
	return s, nil
}

// UpdateSession updates mutable session fields.
func (d *DB) UpdateSession(s *types.ReviewSession) error {
	patterns, _ := json.Marshal(s.IgnorePatterns)
	_, err := d.Exec(
		`UPDATE sessions SET base_ref = ?, review_round = ?, ignore_patterns = ?, updated_at = ? WHERE id = ?`,
		s.BaseRef, s.ReviewRound, string(patterns), time.Now(), s.ID,
	)
	return err
}

// DeleteChangedFiles removes all changed file records for a session.
func (d *DB) DeleteChangedFiles(sessionID string) error {
	_, err := d.Exec(`DELETE FROM changed_files WHERE session_id = ?`, sessionID)
	return err
}

// DeleteContentItems removes all content item records and their versions for a session.
func (d *DB) DeleteContentItems(sessionID string) error {
	if _, err := d.Exec(`DELETE FROM content_versions WHERE session_id = ?`, sessionID); err != nil {
		return err
	}
	_, err := d.Exec(`DELETE FROM content_items WHERE session_id = ?`, sessionID)
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

// MarkContentItemReviewed sets the reviewed flag for a content item.
func (d *DB) MarkContentItemReviewed(sessionID, id string, reviewed bool) error {
	_, err := d.Exec(
		`UPDATE content_items SET reviewed = ? WHERE session_id = ? AND id = ?`,
		boolToInt(reviewed), sessionID, id,
	)
	return err
}

// UpsertContentItem inserts or updates a content item and records a new version.
func (d *DB) UpsertContentItem(sessionID string, item *types.ContentItem) error {
	_, err := d.Exec(
		`INSERT INTO content_items (id, session_id, title, content, content_type, is_plan, reviewed, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id, session_id) DO UPDATE SET title = excluded.title, content = excluded.content, content_type = excluded.content_type, is_plan = excluded.is_plan, updated_at = excluded.updated_at`,
		item.ID, sessionID, item.Title, item.Content, item.ContentType, boolToInt(item.IsPlan), boolToInt(item.Reviewed), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Record a new version
	_, err = d.Exec(
		`INSERT INTO content_versions (content_item_id, session_id, version, title, content, created_at)
		 VALUES (?, ?, COALESCE((SELECT MAX(version) FROM content_versions WHERE content_item_id = ? AND session_id = ?), 0) + 1, ?, ?, ?)`,
		item.ID, sessionID, item.ID, sessionID, item.Title, item.Content, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert content version: %w", err)
	}

	// Update the version count on the item
	err = d.QueryRow(
		`SELECT COUNT(*) FROM content_versions WHERE content_item_id = ? AND session_id = ?`, item.ID, sessionID,
	).Scan(&item.VersionCount)
	return err
}

// GetContentItems returns all content items for a session.
func (d *DB) GetContentItems(sessionID string) ([]types.ContentItem, error) {
	rows, err := d.Query(
		`SELECT c.id, c.title, c.content, c.content_type, c.is_plan, c.reviewed, c.created_at, c.updated_at,
		 (SELECT COUNT(*) FROM content_versions WHERE content_item_id = c.id AND session_id = c.session_id) AS version_count
		 FROM content_items c WHERE c.session_id = ? ORDER BY c.created_at`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.ContentItem
	for rows.Next() {
		var item types.ContentItem
		var isPlan, reviewed int
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.ContentType, &isPlan, &reviewed, &item.CreatedAt, &item.UpdatedAt, &item.VersionCount); err != nil {
			return nil, err
		}
		item.IsPlan = isPlan != 0
		item.Reviewed = reviewed != 0
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetContentItem returns a single content item by ID within a session.
func (d *DB) GetContentItem(sessionID, id string) (*types.ContentItem, error) {
	item := &types.ContentItem{}
	var isPlan, reviewed int
	err := d.QueryRow(
		`SELECT c.id, c.title, c.content, c.content_type, c.is_plan, c.reviewed, c.created_at, c.updated_at,
		 (SELECT COUNT(*) FROM content_versions WHERE content_item_id = c.id AND session_id = c.session_id) AS version_count
		 FROM content_items c WHERE c.id = ? AND c.session_id = ?`, id, sessionID,
	).Scan(&item.ID, &item.Title, &item.Content, &item.ContentType, &isPlan, &reviewed, &item.CreatedAt, &item.UpdatedAt, &item.VersionCount)
	if err != nil {
		return nil, err
	}
	item.IsPlan = isPlan != 0
	item.Reviewed = reviewed != 0
	return item, nil
}

// GetContentVersions returns all versions of a content item within a session, ordered by version ascending.
func (d *DB) GetContentVersions(sessionID, contentItemID string) ([]types.ContentVersion, error) {
	rows, err := d.Query(
		`SELECT content_item_id, version, title, content, created_at
		 FROM content_versions WHERE content_item_id = ? AND session_id = ? ORDER BY version ASC`, contentItemID, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []types.ContentVersion
	for rows.Next() {
		var v types.ContentVersion
		if err := rows.Scan(&v.ContentItemID, &v.Version, &v.Title, &v.Content, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// GetContentVersion returns a single version of a content item within a session.
func (d *DB) GetContentVersion(sessionID, contentItemID string, version int) (*types.ContentVersion, error) {
	v := &types.ContentVersion{}
	err := d.QueryRow(
		`SELECT content_item_id, version, title, content, created_at
		 FROM content_versions WHERE content_item_id = ? AND session_id = ? AND version = ?`, contentItemID, sessionID, version,
	).Scan(&v.ContentItemID, &v.Version, &v.Title, &v.Content, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// CreateComment inserts a new comment.
func (d *DB) CreateComment(sessionID string, c *types.ReviewComment) error {
	_, err := d.Exec(
		`INSERT INTO comments (id, session_id, target_type, target_ref, line_start, line_end, type, body, code_snippet, resolved, outdated, review_round, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, sessionID, string(c.TargetType), c.TargetRef, c.LineStart, c.LineEnd,
		string(c.Type), c.Body, c.CodeSnippet, boolToInt(c.Resolved), 0,
		c.ReviewRound, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

// UpdateComment updates a comment's type, body, and updated_at.
func (d *DB) UpdateComment(c *types.ReviewComment) error {
	_, err := d.Exec(
		`UPDATE comments SET type = ?, body = ?, updated_at = ? WHERE id = ?`,
		string(c.Type), c.Body, time.Now(), c.ID,
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
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// ResolveComment sets the resolved flag for a comment.
func (d *DB) ResolveComment(id string, resolved bool) error {
	_, err := d.Exec(
		`UPDATE comments SET resolved = ?, updated_at = ? WHERE id = ?`,
		boolToInt(resolved), time.Now(), id,
	)
	return err
}

// ClearComments deletes all comments in the session.
func (d *DB) ClearComments(sessionID string) error {
	_, err := d.Exec(`DELETE FROM comments WHERE session_id = ?`, sessionID)
	return err
}

// CreateSubmission inserts a review submission record.
func (d *DB) CreateSubmission(sessionID string, sub *types.ReviewSubmission) error {
	_, err := d.Exec(
		`INSERT INTO review_submissions (id, session_id, action, formatted_review, comment_count, review_round, submitted_at, delivered_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sub.ID, sessionID, string(sub.Action), sub.FormattedReview, sub.CommentCount, sub.ReviewRound, sub.SubmittedAt, sub.DeliveredAt,
	)
	return err
}

// GetSubmissions returns all submissions for a session.
func (d *DB) GetSubmissions(sessionID string) ([]types.ReviewSubmission, error) {
	rows, err := d.Query(
		`SELECT id, session_id, action, formatted_review, comment_count, review_round, submitted_at, delivered_at
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
		if err := rows.Scan(&s.ID, &s.SessionID, &action, &s.FormattedReview, &s.CommentCount, &s.ReviewRound, &s.SubmittedAt, &s.DeliveredAt); err != nil {
			return nil, err
		}
		s.Action = types.SubmitAction(action)
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// MarkSubmissionsDelivered sets delivered_at on all undelivered submissions for a session.
func (d *DB) MarkSubmissionsDelivered(sessionID string) error {
	_, err := d.Exec(
		`UPDATE review_submissions SET delivered_at = ? WHERE session_id = ? AND delivered_at IS NULL`,
		time.Now(), sessionID,
	)
	return err
}

// GetUndeliveredSubmissions returns all undelivered submissions for a session, ordered by submission time.
func (d *DB) GetUndeliveredSubmissions(sessionID string) ([]types.ReviewSubmission, error) {
	rows, err := d.Query(
		`SELECT id, session_id, action, formatted_review, comment_count, review_round, submitted_at
		 FROM review_submissions WHERE session_id = ? AND delivered_at IS NULL ORDER BY submitted_at`, sessionID,
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

// UpsertAdditionalFile inserts or updates an additional file record.
func (d *DB) UpsertAdditionalFile(sessionID string, af *types.AdditionalFile) error {
	_, err := d.Exec(
		`INSERT INTO additional_files (session_id, path, name, reviewed)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(session_id, path) DO UPDATE SET name = excluded.name`,
		sessionID, af.Path, af.Name, boolToInt(af.Reviewed),
	)
	return err
}

// GetAdditionalFiles returns all additional files for a session.
func (d *DB) GetAdditionalFiles(sessionID string) ([]types.AdditionalFile, error) {
	rows, err := d.Query(
		`SELECT path, name, reviewed FROM additional_files WHERE session_id = ? ORDER BY name`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []types.AdditionalFile
	for rows.Next() {
		var f types.AdditionalFile
		var reviewed int
		if err := rows.Scan(&f.Path, &f.Name, &reviewed); err != nil {
			return nil, err
		}
		f.Reviewed = reviewed != 0
		files = append(files, f)
	}
	return files, rows.Err()
}

// MarkAdditionalFileReviewed sets the reviewed flag for an additional file.
func (d *DB) MarkAdditionalFileReviewed(sessionID, path string, reviewed bool) error {
	_, err := d.Exec(
		`UPDATE additional_files SET reviewed = ? WHERE session_id = ? AND path = ?`,
		boolToInt(reviewed), sessionID, path,
	)
	return err
}

// DeleteAdditionalFiles removes all additional file records for a session.
func (d *DB) DeleteAdditionalFiles(sessionID string) error {
	_, err := d.Exec(`DELETE FROM additional_files WHERE session_id = ?`, sessionID)
	return err
}

// MarkAllReviewed sets the reviewed flag on all files, content items, and additional files for a session.
func (d *DB) MarkAllReviewed(sessionID string) error {
	for _, query := range []string{
		`UPDATE changed_files SET reviewed = 1 WHERE session_id = ?`,
		`UPDATE content_items SET reviewed = 1 WHERE session_id = ?`,
		`UPDATE additional_files SET reviewed = 1 WHERE session_id = ?`,
	} {
		if _, err := d.Exec(query, sessionID); err != nil {
			return err
		}
	}
	return nil
}

// ResetAllReviewed resets the reviewed flag on all files, content items, and additional files for a session.
func (d *DB) ResetAllReviewed(sessionID string) error {
	for _, query := range []string{
		`UPDATE changed_files SET reviewed = 0 WHERE session_id = ?`,
		`UPDATE content_items SET reviewed = 0 WHERE session_id = ?`,
		`UPDATE additional_files SET reviewed = 0 WHERE session_id = ?`,
	} {
		if _, err := d.Exec(query, sessionID); err != nil {
			return err
		}
	}
	return nil
}

// CreateSnapshot inserts a review snapshot with its file records.
func (d *DB) CreateSnapshot(sessionID, submissionID string, reviewRound int, headRef, baseRef string, files []types.SnapshotFile) (int64, error) {
	res, err := d.Exec(
		`INSERT INTO review_snapshots (session_id, submission_id, review_round, head_ref, base_ref)
		 VALUES (?, ?, ?, ?, ?)`,
		sessionID, submissionID, reviewRound, headRef, baseRef,
	)
	if err != nil {
		return 0, fmt.Errorf("insert snapshot: %w", err)
	}

	snapshotID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get snapshot id: %w", err)
	}

	for _, f := range files {
		_, err := d.Exec(
			`INSERT INTO review_snapshot_files (snapshot_id, path, status, reviewed, blob_sha, content)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			snapshotID, f.Path, string(f.Status), boolToInt(f.Reviewed), f.BlobSHA, f.Content,
		)
		if err != nil {
			return 0, fmt.Errorf("insert snapshot file %s: %w", f.Path, err)
		}
	}

	return snapshotID, nil
}

// GetSnapshots returns all snapshots for a session, ordered by round descending (most recent first).
// Files are not loaded — use GetSnapshot to load a snapshot with its files.
func (d *DB) GetSnapshots(sessionID string) ([]types.ReviewSnapshot, error) {
	rows, err := d.Query(
		`SELECT id, session_id, submission_id, review_round, head_ref, base_ref, created_at
		 FROM review_snapshots WHERE session_id = ? ORDER BY review_round DESC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []types.ReviewSnapshot
	for rows.Next() {
		var s types.ReviewSnapshot
		if err := rows.Scan(&s.ID, &s.SessionID, &s.SubmissionID, &s.ReviewRound, &s.HeadRef, &s.BaseRef, &s.CreatedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// GetSnapshot returns a single snapshot with its files loaded.
func (d *DB) GetSnapshot(snapshotID int) (*types.ReviewSnapshot, error) {
	s := &types.ReviewSnapshot{}
	err := d.QueryRow(
		`SELECT id, session_id, submission_id, review_round, head_ref, base_ref, created_at
		 FROM review_snapshots WHERE id = ?`, snapshotID,
	).Scan(&s.ID, &s.SessionID, &s.SubmissionID, &s.ReviewRound, &s.HeadRef, &s.BaseRef, &s.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := d.Query(
		`SELECT path, status, reviewed, blob_sha, content
		 FROM review_snapshot_files WHERE snapshot_id = ? ORDER BY path`, snapshotID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var f types.SnapshotFile
		var status string
		var reviewed int
		if err := rows.Scan(&f.Path, &status, &reviewed, &f.BlobSHA, &f.Content); err != nil {
			return nil, err
		}
		f.Status = types.FileChangeStatus(status)
		f.Reviewed = reviewed != 0
		s.Files = append(s.Files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build path lookup map for O(1) access
	s.FilesByPath = make(map[string]*types.SnapshotFile, len(s.Files))
	for i := range s.Files {
		s.FilesByPath[s.Files[i].Path] = &s.Files[i]
	}

	return s, nil
}

// DeleteSnapshots removes all snapshots and their files for a session.
func (d *DB) DeleteSnapshots(sessionID string) error {
	// Delete files first (child rows)
	_, err := d.Exec(
		`DELETE FROM review_snapshot_files WHERE snapshot_id IN
		 (SELECT id FROM review_snapshots WHERE session_id = ?)`, sessionID,
	)
	if err != nil {
		return fmt.Errorf("delete snapshot files: %w", err)
	}

	// Delete content item snapshots
	_, err = d.Exec(
		`DELETE FROM review_snapshot_content WHERE snapshot_id IN
		 (SELECT id FROM review_snapshots WHERE session_id = ?)`, sessionID,
	)
	if err != nil {
		return fmt.Errorf("delete snapshot content: %w", err)
	}

	// Delete snapshots
	_, err = d.Exec(`DELETE FROM review_snapshots WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("delete snapshots: %w", err)
	}

	return nil
}

// HasSnapshots returns true if any snapshots exist for the session.
func (d *DB) HasSnapshots(sessionID string) (bool, error) {
	var count int
	err := d.QueryRow(
		`SELECT COUNT(*) FROM review_snapshots WHERE session_id = ?`, sessionID,
	).Scan(&count)
	return count > 0, err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

