// NDJSON protocol types for Monocle hook communication
// See ARCHITECTURE.md section 4

/** Supported AI coding agents */
export type AgentName = "claude" | "codex" | "gemini";

// ---------------------------------------------------------------------------
// Inbound hook messages (agent → Monocle)
// ---------------------------------------------------------------------------

export interface PreToolUseMessage {
  type: "pre_tool_use";
  agent: AgentName;
  tool: string;
  file_path: string;
  content: string;
  tool_input: Record<string, unknown>;
  request_id: string;
}

export interface PostToolUseMessage {
  type: "post_tool_use";
  agent: AgentName;
  tool: string;
  file_path: string;
  tool_input: Record<string, unknown>;
  tool_output: Record<string, unknown>;
}

export interface StopMessage {
  type: "stop";
  agent: AgentName;
  stop_reason: string;
  request_id: string;
}

export interface PromptSubmitMessage {
  type: "prompt_submit";
  agent: AgentName;
  prompt: string;
  request_id: string;
}

export interface ContentReviewMessage {
  type: "content_review";
  /** Nullable — used for replacement semantics */
  id: string | null;
  title: string;
  content: string;
  content_type: "markdown" | "text" | "code";
  request_id: string;
}

export type HookMessage =
  | PreToolUseMessage
  | PostToolUseMessage
  | StopMessage
  | PromptSubmitMessage
  | ContentReviewMessage;

// ---------------------------------------------------------------------------
// Outbound hook responses (Monocle → agent)
// ---------------------------------------------------------------------------

export interface PreToolUseResponse {
  type: "pre_tool_use_response";
  request_id: string;
  decision: "allow" | "deny";
  reason: string | null;
}

export interface StopResponse {
  type: "stop_response";
  request_id: string;
  continue: boolean;
  system_message: string | null;
}

export interface PromptSubmitResponse {
  type: "prompt_submit_response";
  request_id: string;
  additional_context: string | null;
}

export type HookResponse =
  | PreToolUseResponse
  | StopResponse
  | PromptSubmitResponse;
