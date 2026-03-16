// Agent adapter interface for Monocle
// See ARCHITECTURE.md section 5.1

import type { HookMessage, HookResponse } from "../types/protocol";

export interface AgentCapabilities {
  preToolUse: boolean;
  postToolUse: boolean;
  stopBlocking: boolean;
  asyncHooks: boolean;
}

export interface HookOutput {
  stdout: string;
  exitCode: number;
}

export interface SetupOptions {
  socketPath: string;
  agent: string;
  mode: string;
}

export interface AgentAdapter {
  parseHookInput(event: string, raw: Record<string, unknown>): HookMessage | null;
  formatHookOutput(response: HookResponse): HookOutput;
  generateConfig(options: SetupOptions): string;
  capabilities: AgentCapabilities;
}
