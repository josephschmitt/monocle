// Configuration schema for Monocle
// See ARCHITECTURE.md section 11

export interface ReviewFormatConfig {
  include_snippets: boolean;
  max_snippet_lines: number;
  include_summary: boolean;
}

export interface HooksConfig {
  socket_path: string;
  timeout_ms: number;
}

export interface MonocleConfig {
  default_mode: string;
  default_agent: string;
  gate_patterns: string[];
  ignore_patterns: string[];
  keybindings: string;
  diff_style: string;
  theme: string;
  review_format: ReviewFormatConfig;
  hooks: HooksConfig;
}
