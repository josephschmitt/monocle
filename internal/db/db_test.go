package db

import (
	"strings"
	"testing"
	"time"

	"github.com/josephschmitt/monocle/internal/types"
)

func TestDBPath_EnvOverride(t *testing.T) {
	t.Setenv("MONOCLE_DB", "/custom/path/test.db")
	got := DBPath()
	if got != "/custom/path/test.db" {
		t.Errorf("expected /custom/path/test.db, got %q", got)
	}
}

func TestDBPath_Default(t *testing.T) {
	t.Setenv("MONOCLE_DB", "")
	got := DBPath()
	if got == "" {
		t.Error("expected non-empty default path")
	}
	if !strings.HasSuffix(got, "monocle/monocle.db") {
		t.Errorf("expected path ending in monocle/monocle.db, got %q", got)
	}
}

func testDB(t *testing.T) *DB {
	t.Helper()
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestSessionCRUD(t *testing.T) {
	d := testDB(t)
	now := time.Now()

	s := &types.ReviewSession{
		ID:          "sess-1",
		Agent:       "claude",
		RepoRoot:    "/tmp/repo",
		BaseRef:     "abc123",
		ReviewRound: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := d.CreateSession(s); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := d.GetSession("sess-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Agent != "claude" {
		t.Errorf("got agent=%q", got.Agent)
	}

	s.BaseRef = "def456"
	if err := d.UpdateSession(s); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ = d.GetSession("sess-1")
	if got.BaseRef != "def456" {
		t.Errorf("expected updated base_ref, got %q", got.BaseRef)
	}

	summaries, err := d.ListSessions("", 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(summaries) != 1 {
		t.Errorf("expected 1 session, got %d", len(summaries))
	}
}

func TestChangedFiles(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	f := &types.ChangedFile{Path: "main.go", Status: types.FileModified}
	if err := d.UpsertChangedFile("sess-1", f); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	files, err := d.GetChangedFiles("sess-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(files) != 1 || files[0].Path != "main.go" {
		t.Errorf("unexpected files: %+v", files)
	}

	if err := d.MarkFileReviewed("sess-1", "main.go", true); err != nil {
		t.Fatalf("mark: %v", err)
	}

	files, _ = d.GetChangedFiles("sess-1")
	if !files[0].Reviewed {
		t.Error("expected reviewed")
	}
}

func TestComments(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	c := &types.ReviewComment{
		ID:          "cmt-1",
		TargetType:  types.TargetFile,
		TargetRef:   "main.go",
		LineStart:   10,
		LineEnd:     15,
		Type:        types.CommentIssue,
		Body:        "Fix this bug",
		ReviewRound: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := d.CreateComment("sess-1", c); err != nil {
		t.Fatalf("create: %v", err)
	}

	comments, err := d.GetComments("sess-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(comments) != 1 || comments[0].Body != "Fix this bug" {
		t.Errorf("unexpected: %+v", comments)
	}

	c.Body = "Updated body"
	if err := d.UpdateComment(c); err != nil {
		t.Fatalf("update: %v", err)
	}

	comments, _ = d.GetComments("sess-1")
	if comments[0].Body != "Updated body" {
		t.Errorf("expected updated body, got %q", comments[0].Body)
	}

	if err := d.ClearComments("sess-1"); err != nil {
		t.Fatalf("clear comments: %v", err)
	}
	comments, _ = d.GetComments("sess-1")
	if len(comments) != 0 {
		t.Errorf("expected 0 comments after clear, got %d", len(comments))
	}
}

func TestContentItems(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	item := &types.ContentItem{
		ID:          "item-1",
		Title:       "Test Plan",
		Content:     "# Test Plan\nSteps...",
		ContentType: "markdown",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := d.UpsertContentItem("sess-1", item); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	items, err := d.GetContentItems("sess-1")
	if err != nil {
		t.Fatalf("get items: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Test Plan" {
		t.Errorf("unexpected: %+v", items)
	}

	got, err := d.GetContentItem("sess-1", "item-1")
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if got.Content != "# Test Plan\nSteps..." {
		t.Errorf("unexpected content: %q", got.Content)
	}
	if got.VersionCount != 1 {
		t.Errorf("expected version count 1 for first version, got %d", got.VersionCount)
	}

	// Update the same item — version count should increase
	item2 := &types.ContentItem{
		ID:          "item-1",
		Title:       "Updated Plan",
		Content:     "# Updated Plan\nNew steps...",
		ContentType: "markdown",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := d.UpsertContentItem("sess-1", item2); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	got2, err := d.GetContentItem("sess-1", "item-1")
	if err != nil {
		t.Fatalf("get updated item: %v", err)
	}
	if got2.Content != "# Updated Plan\nNew steps..." {
		t.Errorf("unexpected content after update: %q", got2.Content)
	}
	if got2.VersionCount != 2 {
		t.Errorf("expected version count 2, got %d", got2.VersionCount)
	}

	// Update again — version count should be 3
	item3 := &types.ContentItem{
		ID:          "item-1",
		Title:       "Plan v3",
		Content:     "# Plan v3\nFinal steps...",
		ContentType: "markdown",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := d.UpsertContentItem("sess-1", item3); err != nil {
		t.Fatalf("upsert v3: %v", err)
	}

	got3, err := d.GetContentItem("sess-1", "item-1")
	if err != nil {
		t.Fatalf("get v3 item: %v", err)
	}
	if got3.VersionCount != 3 {
		t.Errorf("expected version count 3, got %d", got3.VersionCount)
	}

	// Verify only one item exists (upsert, not insert)
	items2, err := d.GetContentItems("sess-1")
	if err != nil {
		t.Fatalf("get items after updates: %v", err)
	}
	if len(items2) != 1 {
		t.Errorf("expected 1 item after upserts, got %d", len(items2))
	}
	if items2[0].VersionCount != 3 {
		t.Errorf("expected version count 3 from GetContentItems, got %d", items2[0].VersionCount)
	}

	// Verify version history
	versions, err := d.GetContentVersions("sess-1", "item-1")
	if err != nil {
		t.Fatalf("get versions: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	if versions[0].Version != 1 || versions[0].Content != "# Test Plan\nSteps..." {
		t.Errorf("version 1 mismatch: %+v", versions[0])
	}
	if versions[1].Version != 2 || versions[1].Content != "# Updated Plan\nNew steps..." {
		t.Errorf("version 2 mismatch: %+v", versions[1])
	}
	if versions[2].Version != 3 || versions[2].Content != "# Plan v3\nFinal steps..." {
		t.Errorf("version 3 mismatch: %+v", versions[2])
	}

	// Verify single version fetch
	v2, err := d.GetContentVersion("sess-1", "item-1", 2)
	if err != nil {
		t.Fatalf("get version 2: %v", err)
	}
	if v2.Content != "# Updated Plan\nNew steps..." {
		t.Errorf("version 2 content mismatch: %q", v2.Content)
	}
}

func TestContentItems_CrossSession(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})
	d.CreateSession(&types.ReviewSession{ID: "sess-2", Agent: "claude", RepoRoot: "/tmp", BaseRef: "def", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	// Submit plan to session 1 twice (v1 and v2)
	item1 := &types.ContentItem{ID: "plan", Title: "Plan v1", Content: "content v1", ContentType: "markdown", CreatedAt: now, UpdatedAt: now}
	if err := d.UpsertContentItem("sess-1", item1); err != nil {
		t.Fatalf("upsert sess-1 v1: %v", err)
	}
	item1b := &types.ContentItem{ID: "plan", Title: "Plan v2", Content: "content v2", ContentType: "markdown", CreatedAt: now, UpdatedAt: now}
	if err := d.UpsertContentItem("sess-1", item1b); err != nil {
		t.Fatalf("upsert sess-1 v2: %v", err)
	}

	// Submit same plan ID to session 2 — should start at v1
	item2 := &types.ContentItem{ID: "plan", Title: "Plan v1", Content: "content v1 new session", ContentType: "markdown", CreatedAt: now, UpdatedAt: now}
	if err := d.UpsertContentItem("sess-2", item2); err != nil {
		t.Fatalf("upsert sess-2 v1: %v", err)
	}

	// Session 1 should have 2 versions
	v1, err := d.GetContentVersions("sess-1", "plan")
	if err != nil {
		t.Fatalf("get versions sess-1: %v", err)
	}
	if len(v1) != 2 {
		t.Errorf("expected 2 versions for sess-1, got %d", len(v1))
	}
	if v1[0].Version != 1 || v1[1].Version != 2 {
		t.Errorf("expected versions 1,2 for sess-1, got %d,%d", v1[0].Version, v1[1].Version)
	}

	// Session 2 should have 1 version starting at v1
	v2, err := d.GetContentVersions("sess-2", "plan")
	if err != nil {
		t.Fatalf("get versions sess-2: %v", err)
	}
	if len(v2) != 1 {
		t.Errorf("expected 1 version for sess-2, got %d", len(v2))
	}
	if v2[0].Version != 1 {
		t.Errorf("expected version 1 for sess-2, got %d", v2[0].Version)
	}

	// Version counts should be independent
	got1, err := d.GetContentItem("sess-1", "plan")
	if err != nil {
		t.Fatalf("get item sess-1: %v", err)
	}
	if got1.VersionCount != 2 {
		t.Errorf("expected version count 2 for sess-1, got %d", got1.VersionCount)
	}

	got2, err := d.GetContentItem("sess-2", "plan")
	if err != nil {
		t.Fatalf("get item sess-2: %v", err)
	}
	if got2.VersionCount != 1 {
		t.Errorf("expected version count 1 for sess-2, got %d", got2.VersionCount)
	}

	// GetContentItems should return independent items per session
	items1, _ := d.GetContentItems("sess-1")
	items2, _ := d.GetContentItems("sess-2")
	if len(items1) != 1 || items1[0].VersionCount != 2 {
		t.Errorf("sess-1 items unexpected: %+v", items1)
	}
	if len(items2) != 1 || items2[0].VersionCount != 1 {
		t.Errorf("sess-2 items unexpected: %+v", items2)
	}
}

func TestDeleteComment(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	c := &types.ReviewComment{
		ID:          "cmt-1",
		TargetType:  types.TargetFile,
		TargetRef:   "main.go",
		LineStart:   10,
		LineEnd:     15,
		Type:        types.CommentIssue,
		Body:        "Fix this bug",
		ReviewRound: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := d.CreateComment("sess-1", c); err != nil {
		t.Fatalf("create comment: %v", err)
	}

	if err := d.DeleteComment("cmt-1"); err != nil {
		t.Fatalf("delete comment: %v", err)
	}

	comments, err := d.GetComments("sess-1")
	if err != nil {
		t.Fatalf("get comments: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments after delete, got %d", len(comments))
	}
}

func TestClearComments(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	c1 := &types.ReviewComment{ID: "cmt-a1", TargetType: types.TargetFile, TargetRef: "a.go", Type: types.CommentIssue, Body: "first", ReviewRound: 1, CreatedAt: now, UpdatedAt: now}
	c2 := &types.ReviewComment{ID: "cmt-a2", TargetType: types.TargetFile, TargetRef: "b.go", Type: types.CommentIssue, Body: "second", ReviewRound: 1, CreatedAt: now, UpdatedAt: now}
	d.CreateComment("sess-1", c1)
	d.CreateComment("sess-1", c2)

	if err := d.ClearComments("sess-1"); err != nil {
		t.Fatalf("clear comments: %v", err)
	}

	comments, err := d.GetComments("sess-1")
	if err != nil {
		t.Fatalf("get comments: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments after clear, got %d", len(comments))
	}
}

func TestDeleteChangedFiles(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	f1 := &types.ChangedFile{Path: "main.go", Status: types.FileModified}
	f2 := &types.ChangedFile{Path: "util.go", Status: types.FileAdded}
	d.UpsertChangedFile("sess-1", f1)
	d.UpsertChangedFile("sess-1", f2)

	if err := d.DeleteChangedFiles("sess-1"); err != nil {
		t.Fatalf("delete changed files: %v", err)
	}

	files, err := d.GetChangedFiles("sess-1")
	if err != nil {
		t.Fatalf("get changed files: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 changed files after delete, got %d", len(files))
	}
}

func TestCreateAndGetSubmissions(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	sub := &types.ReviewSubmission{
		ID:              "sub-1",
		SessionID:       "sess-1",
		Action:          types.ActionRequestChanges,
		FormattedReview: "Please fix the error handling",
		CommentCount:    3,
		ReviewRound:     1,
		SubmittedAt:     now,
	}
	if err := d.CreateSubmission("sess-1", sub); err != nil {
		t.Fatalf("create submission: %v", err)
	}

	subs, err := d.GetSubmissions("sess-1")
	if err != nil {
		t.Fatalf("get submissions: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(subs))
	}
	got := subs[0]
	if got.ID != "sub-1" {
		t.Errorf("expected ID sub-1, got %q", got.ID)
	}
	if got.Action != types.ActionRequestChanges {
		t.Errorf("expected action request_changes, got %q", got.Action)
	}
	if got.FormattedReview != "Please fix the error handling" {
		t.Errorf("expected formatted review text, got %q", got.FormattedReview)
	}
	if got.CommentCount != 3 {
		t.Errorf("expected comment_count 3, got %d", got.CommentCount)
	}
	if got.ReviewRound != 1 {
		t.Errorf("expected review_round 1, got %d", got.ReviewRound)
	}
}

func TestAdditionalFiles(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	af := &types.AdditionalFile{Path: "/tmp/extra.go", Name: "extra.go"}
	if err := d.UpsertAdditionalFile("sess-1", af); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	files, err := d.GetAdditionalFiles("sess-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(files) != 1 || files[0].Path != "/tmp/extra.go" || files[0].Name != "extra.go" {
		t.Errorf("unexpected files: %+v", files)
	}

	// Upsert same path updates name
	af2 := &types.AdditionalFile{Path: "/tmp/extra.go", Name: "renamed.go"}
	if err := d.UpsertAdditionalFile("sess-1", af2); err != nil {
		t.Fatalf("upsert update: %v", err)
	}
	files, _ = d.GetAdditionalFiles("sess-1")
	if len(files) != 1 || files[0].Name != "renamed.go" {
		t.Errorf("expected updated name, got %+v", files)
	}

	// Mark reviewed
	if err := d.MarkAdditionalFileReviewed("sess-1", "/tmp/extra.go", true); err != nil {
		t.Fatalf("mark reviewed: %v", err)
	}
	files, _ = d.GetAdditionalFiles("sess-1")
	if !files[0].Reviewed {
		t.Error("expected reviewed")
	}

	// Unmark reviewed
	if err := d.MarkAdditionalFileReviewed("sess-1", "/tmp/extra.go", false); err != nil {
		t.Fatalf("unmark reviewed: %v", err)
	}
	files, _ = d.GetAdditionalFiles("sess-1")
	if files[0].Reviewed {
		t.Error("expected not reviewed")
	}

	// Delete
	if err := d.DeleteAdditionalFiles("sess-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	files, _ = d.GetAdditionalFiles("sess-1")
	if len(files) != 0 {
		t.Errorf("expected 0 files after delete, got %d", len(files))
	}
}

func TestListSessions_WithFilter(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp/repo-a", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})
	d.CreateSession(&types.ReviewSession{ID: "sess-2", Agent: "claude", RepoRoot: "/tmp/repo-b", BaseRef: "def", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	summaries, err := d.ListSessions("/tmp/repo-a", 0)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 session with filter, got %d", len(summaries))
	}
	if summaries[0].ID != "sess-1" {
		t.Errorf("expected sess-1, got %q", summaries[0].ID)
	}
	if summaries[0].RepoRoot != "/tmp/repo-a" {
		t.Errorf("expected repo root /tmp/repo-a, got %q", summaries[0].RepoRoot)
	}
}

func TestListSessions_WithLimit(t *testing.T) {
	d := testDB(t)
	now := time.Now()
	d.CreateSession(&types.ReviewSession{ID: "sess-1", Agent: "claude", RepoRoot: "/tmp", BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})
	d.CreateSession(&types.ReviewSession{ID: "sess-2", Agent: "claude", RepoRoot: "/tmp", BaseRef: "def", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})
	d.CreateSession(&types.ReviewSession{ID: "sess-3", Agent: "claude", RepoRoot: "/tmp", BaseRef: "ghi", ReviewRound: 1, CreatedAt: now, UpdatedAt: now})

	summaries, err := d.ListSessions("", 2)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("expected 2 sessions with limit, got %d", len(summaries))
	}
}

func TestSnapshotCRUD(t *testing.T) {
	d := testDB(t)
	now := time.Now()

	// Create a session and submission
	d.CreateSession(&types.ReviewSession{
		ID: "sess-1", Agent: "claude", RepoRoot: "/tmp",
		BaseRef: "abc123", ReviewRound: 1, CreatedAt: now, UpdatedAt: now,
	})
	d.CreateSubmission("sess-1", &types.ReviewSubmission{
		ID: "sub-1", SessionID: "sess-1", Action: types.ActionRequestChanges,
		FormattedReview: "review", ReviewRound: 1, SubmittedAt: now,
	})

	// Create a snapshot with files
	files := []types.SnapshotFile{
		{Path: "main.go", Status: types.FileModified, Reviewed: true, BlobSHA: "abc123def"},
		{Path: "utils.go", Status: types.FileAdded, Reviewed: false, BlobSHA: "def456abc"},
	}
	snapID, err := d.CreateSnapshot("sess-1", "sub-1", 1, "headabc", "baseabc", files)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	if snapID == 0 {
		t.Fatal("expected non-zero snapshot ID")
	}

	// List snapshots
	snapshots, err := d.GetSnapshots("sess-1")
	if err != nil {
		t.Fatalf("get snapshots: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}
	if snapshots[0].ReviewRound != 1 {
		t.Errorf("expected round 1, got %d", snapshots[0].ReviewRound)
	}
	if snapshots[0].HeadRef != "headabc" {
		t.Errorf("expected head ref headabc, got %q", snapshots[0].HeadRef)
	}

	// Get snapshot with files
	snap, err := d.GetSnapshot(int(snapID))
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if len(snap.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(snap.Files))
	}
	if snap.Files[0].Path != "main.go" {
		t.Errorf("expected main.go, got %q", snap.Files[0].Path)
	}
	if !snap.Files[0].Reviewed {
		t.Error("expected main.go to be reviewed")
	}
	if snap.Files[0].BlobSHA != "abc123def" {
		t.Errorf("expected blob SHA abc123def, got %q", snap.Files[0].BlobSHA)
	}

	// HasSnapshots
	has, err := d.HasSnapshots("sess-1")
	if err != nil {
		t.Fatalf("has snapshots: %v", err)
	}
	if !has {
		t.Error("expected HasSnapshots to return true")
	}

	// Delete snapshots
	if err := d.DeleteSnapshots("sess-1"); err != nil {
		t.Fatalf("delete snapshots: %v", err)
	}
	has, _ = d.HasSnapshots("sess-1")
	if has {
		t.Error("expected HasSnapshots to return false after delete")
	}
}

func TestSnapshotMultipleRounds(t *testing.T) {
	d := testDB(t)
	now := time.Now()

	d.CreateSession(&types.ReviewSession{
		ID: "sess-1", Agent: "claude", RepoRoot: "/tmp",
		BaseRef: "abc", ReviewRound: 1, CreatedAt: now, UpdatedAt: now,
	})
	d.CreateSubmission("sess-1", &types.ReviewSubmission{
		ID: "sub-1", SessionID: "sess-1", Action: types.ActionRequestChanges,
		FormattedReview: "r1", ReviewRound: 1, SubmittedAt: now,
	})
	d.CreateSubmission("sess-1", &types.ReviewSubmission{
		ID: "sub-2", SessionID: "sess-1", Action: types.ActionRequestChanges,
		FormattedReview: "r2", ReviewRound: 2, SubmittedAt: now,
	})

	// Create snapshots for two rounds
	d.CreateSnapshot("sess-1", "sub-1", 1, "head1", "base1", []types.SnapshotFile{
		{Path: "main.go", Status: types.FileModified, BlobSHA: "sha1"},
	})
	d.CreateSnapshot("sess-1", "sub-2", 2, "head2", "base2", []types.SnapshotFile{
		{Path: "main.go", Status: types.FileModified, BlobSHA: "sha2"},
	})

	// GetSnapshots returns most recent first
	snapshots, _ := d.GetSnapshots("sess-1")
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}
	if snapshots[0].ReviewRound != 2 {
		t.Errorf("expected round 2 first (most recent), got %d", snapshots[0].ReviewRound)
	}
	if snapshots[1].ReviewRound != 1 {
		t.Errorf("expected round 1 second, got %d", snapshots[1].ReviewRound)
	}
}
