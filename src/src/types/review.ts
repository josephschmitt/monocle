// Data model types for Monocle review sessions
// See ARCHITECTURE.md section 7.2

// ---------------------------------------------------------------------------
// Union / enum types
// ---------------------------------------------------------------------------

export type ReviewState =
  | "idle"
  | "watching"
  | "reviewing"
  | "gating"
  | "submitted";

export type ReviewMode = "review" | "gate" | "hybrid";

export type CommentType = "issue" | "suggestion" | "note" | "praise";

export type CommentTargetType = "file" | "content";

export type FileStatus = "pending" | "reviewed" | "approved";

export type ChangeStatus = "added" | "modified" | "deleted" | "renamed";

export type ContentType = "markdown" | "text" | "code";

export type SubmissionAction = "approve" | "request_changes" | "comment";

// ---------------------------------------------------------------------------
// Data model
// ---------------------------------------------------------------------------

export interface DiffHunk {
  old_start: number;
  old_count: number;
  new_start: number;
  new_count: number;
  content: string;
}

export interface ReviewComment {
  id: string;
  target_type: CommentTargetType;
  target_ref: string;
  line_start: number;
  line_end: number;
  type: CommentType;
  body: string;
  code_snippet: string;
  resolved: boolean;
  created_at: number;
}

export interface ChangedFile {
  path: string;
  status: ChangeStatus;
  hunks: DiffHunk[];
  reviewed: boolean;
  comments: ReviewComment[];
}

export interface ContentItem {
  id: string;
  title: string;
  content: string;
  content_type: ContentType;
  reviewed: boolean;
  comments: ReviewComment[];
  created_at: number;
}

export interface ReviewSession {
  id: string;
  mode: ReviewMode;
  state: ReviewState;
  agent: import("./protocol").AgentName;
  repo_root: string;
  base_ref: string;
  changed_files: ChangedFile[];
  content_items: ContentItem[];
  comments: ReviewComment[];
  file_statuses: Map<string, FileStatus>;
  gate_patterns: string[];
  ignore_patterns: string[];
  created_at: number;
  updated_at: number;
}

export interface ReviewSubmission {
  id: string;
  session_id: string;
  action: SubmissionAction;
  formatted_review: string;
  comment_count: number;
  submitted_at: number;
}
