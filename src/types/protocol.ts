// Protocol message types for hook ↔ Monocle communication over UDS (NDJSON)

// ---------------------------------------------------------------------------
// Message types
// ---------------------------------------------------------------------------

export type HookMessageType =
  | "pre_tool_use"
  | "post_tool_use"
  | "stop"
  | "prompt_submit"
  | "content_review";

export type HookResponseType =
  | "pre_tool_use_response"
  | "stop_response"
  | "prompt_submit_response";

// ---------------------------------------------------------------------------
// Inbound messages (Hook → Monocle)
// ---------------------------------------------------------------------------

export interface BaseHookMessage {
  request_id: string;
  session_id: string;
  timestamp: number;
}

export interface PreToolUseMessage extends BaseHookMessage {
  type: "pre_tool_use";
  tool_name: string;
  tool_input: Record<string, unknown>;
}

export interface PostToolUseMessage extends BaseHookMessage {
  type: "post_tool_use";
  tool_name: string;
  tool_input: Record<string, unknown>;
  tool_output: string;
}

export interface StopMessage extends BaseHookMessage {
  type: "stop";
  reason: string;
}

export interface PromptSubmitMessage extends BaseHookMessage {
  type: "prompt_submit";
  prompt: string;
}

export interface ContentReviewMessage extends BaseHookMessage {
  type: "content_review";
  content_id: string;
  title: string;
  content: string;
  content_type: string;
}

export type HookMessage =
  | PreToolUseMessage
  | PostToolUseMessage
  | StopMessage
  | PromptSubmitMessage
  | ContentReviewMessage;

// ---------------------------------------------------------------------------
// Outbound responses (Monocle → Hook)
// ---------------------------------------------------------------------------

export interface BaseHookResponse {
  request_id: string;
  timestamp: number;
}

export interface PreToolUseResponse extends BaseHookResponse {
  type: "pre_tool_use_response";
  allowed: boolean;
  reason?: string;
}

export interface StopResponse extends BaseHookResponse {
  type: "stop_response";
  continue: boolean;
  system_message?: string;
}

export interface PromptSubmitResponse extends BaseHookResponse {
  type: "prompt_submit_response";
  additional_context?: string;
}

export type HookResponse =
  | PreToolUseResponse
  | StopResponse
  | PromptSubmitResponse;
