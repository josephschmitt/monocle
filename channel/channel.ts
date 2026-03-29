import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { connect } from "net";
import { createHash } from "crypto";
import { statSync, existsSync } from "fs";
import { resolve, join, dirname } from "path";

// -- Socket path computation (mirrors Go's FindRepoRoot + DefaultSocketPath) --

function findRepoRoot(startDir: string): string {
  let dir = resolve(startDir);
  while (true) {
    try {
      statSync(join(dir, ".git"));
      return dir;
    } catch {
      const parent = dirname(dir);
      if (parent === dir) return resolve(startDir);
      dir = parent;
    }
  }
}

function defaultSocketPath(dir: string): string {
  const abs = resolve(dir);
  const hash = createHash("sha256").update(abs).digest("hex").slice(0, 12);
  return "/tmp/monocle-" + hash + ".sock";
}

// -- Types --

type Message = {
  type: string;
  [key: string]: any;
};

// -- Engine Connection --
// Maintains a persistent socket connection to the Monocle engine for receiving
// push event notifications. Pull-based operations (get-feedback, send-artifact,
// etc.) are handled by the `monocle review` CLI commands, not this channel.

class EngineConnection {
  private socketPath: string;
  private conn: ReturnType<typeof connect> | null = null;
  private onEvent: (event: string, payload: Record<string, any>) => void;
  private onConnectionChange: (connected: boolean) => void;
  private lineBuffer = "";
  private reconnecting = false;
  private closed = false;
  private _connected = false;
  private hasConnected = false;
  private agentName: string | null = null;
  private connectionResolvers: Array<() => void> = [];

  get isConnected(): boolean {
    return this._connected;
  }

  private setConnected(value: boolean) {
    if (this._connected !== value) {
      this._connected = value;
      this.onConnectionChange(value);
      if (value) {
        this.hasConnected = true;
        const resolvers = this.connectionResolvers.splice(0);
        for (const r of resolvers) r();
      }
    }
  }

  constructor(
    socketPath: string,
    onEvent: (event: string, payload: Record<string, any>) => void,
    onConnectionChange: (connected: boolean) => void,
  ) {
    this.socketPath = socketPath;
    this.onEvent = onEvent;
    this.onConnectionChange = onConnectionChange;
  }

  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!existsSync(this.socketPath)) {
        reject(new Error("Socket not found: " + this.socketPath));
        return;
      }

      this.conn = connect(this.socketPath, () => {
        const msg = JSON.stringify({
          type: "connect",
          events: [
            "feedback_submitted",
            "pause_changed",
            "content_item_added",
            "additional_file_added",
          ],
        });
        this.conn!.write(msg + "\n");
      });

      this.conn.setEncoding("utf8");
      this.lineBuffer = "";
      let gotAck = false;

      this.conn.on("data", (chunk: string) => {
        this.lineBuffer += chunk;
        const lines = this.lineBuffer.split("\n");
        this.lineBuffer = lines.pop() || "";

        for (const line of lines) {
          if (!line.trim()) continue;
          try {
            const msg: Message = JSON.parse(line);
            if (!gotAck && msg.type === "connect_response") {
              gotAck = true;
              this.setConnected(true);
              resolve();
              continue;
            }
            this.handleMessage(msg);
          } catch {
            // ignore malformed lines
          }
        }
      });

      this.conn.on("error", (err: Error) => {
        if (!gotAck) {
          reject(err);
        } else {
          this.setConnected(false);
          this.scheduleReconnect();
        }
      });

      this.conn.on("close", () => {
        if (gotAck && !this.closed) {
          this.setConnected(false);
          this.scheduleReconnect();
        }
      });
    });
  }

  private handleMessage(msg: Message) {
    if (msg.type === "event_notification") {
      this.onEvent(msg.event, msg.payload || {});
    }
  }

  private scheduleReconnect() {
    if (this.reconnecting || this.closed) return;
    this.reconnecting = true;

    const attempt = (delay: number) => {
      if (this.closed) return;
      const timer = setTimeout(async () => {
        try {
          await this.connect();
          this.reconnecting = false;
          if (this.agentName) {
            this.identify(this.agentName);
          }
        } catch {
          if (this.hasConnected) {
            attempt(Math.min(delay * 2, 10000));
          } else {
            attempt(2000);
          }
        }
      }, delay);
      timer.unref();
    };

    attempt(this.hasConnected ? 1000 : 2000);
  }

  connectInBackground() {
    this.scheduleReconnect();
  }

  identify(agent: string) {
    this.agentName = agent;
    if (this.conn && !this.conn.destroyed) {
      this.conn.write(JSON.stringify({ type: "identify", agent }) + "\n");
    }
  }

  waitForConnection(timeoutMs: number): Promise<boolean> {
    if (this._connected) return Promise.resolve(true);
    return new Promise<boolean>((resolve) => {
      const onConnect = () => {
        clearTimeout(timer);
        resolve(true);
      };
      this.connectionResolvers.push(onConnect);
      const timer = setTimeout(() => {
        const idx = this.connectionResolvers.indexOf(onConnect);
        if (idx >= 0) this.connectionResolvers.splice(idx, 1);
        resolve(false);
      }, timeoutMs);
      timer.unref();
    });
  }

  close() {
    this.closed = true;
    this.setConnected(false);
    if (this.conn) {
      this.conn.destroy();
      this.conn = null;
    }
  }
}

// -- Instructions --

const INSTRUCTIONS = [
  "When you receive a feedback_submitted event, run `monocle review get-feedback` to retrieve the review.",
  "When you receive a pause_requested event, run `monocle review get-feedback --wait` to block until the reviewer submits feedback.",
].join("\n");

// -- Main --

const cwd = process.cwd();
const repoRoot = findRepoRoot(cwd);
const socketPath = process.env.MONOCLE_SOCKET || defaultSocketPath(repoRoot);

// MCP server with channel capability for push notifications only.
// Pull-based operations are handled by `monocle review` CLI commands.
const mcp = new Server(
  { name: "monocle", version: "1.0.0" },
  {
    capabilities: {
      experimental: { "claude/channel": {} },
    },
    instructions: INSTRUCTIONS,
  },
);

// Engine connection — forwards push events as channel notifications.
// These are fire-and-forget hints: the agent runs `monocle review get-feedback`
// to actually retrieve the feedback.
const engine = new EngineConnection(
  socketPath,
  // onEvent
  (event, payload) => {
    switch (event) {
      case "feedback_submitted":
        mcp
          .notification({
            method: "notifications/claude/channel",
            params: {
              content:
                payload.message ||
                "Your reviewer has submitted feedback. Run `monocle review get-feedback` to retrieve it.",
              meta: { event: "feedback_submitted" },
            },
          })
          .catch(() => {});
        break;
      case "pause_changed":
        if (payload.status === "pause_requested") {
          mcp
            .notification({
              method: "notifications/claude/channel",
              params: {
                content:
                  "Your reviewer has requested you pause and wait for feedback. " +
                  "Run `monocle review get-feedback --wait` to block until feedback is ready.",
                meta: { event: "pause_requested" },
              },
            })
            .catch(() => {});
        }
        break;
    }
  },
  // onConnectionChange
  (_connected: boolean) => {},
);

// No tools — all operations are CLI commands now.
mcp.setRequestHandler(ListToolsRequestSchema, async () => {
  return { tools: [] };
});

// -- Start --

async function main() {
  try {
    await engine.connect();
  } catch {
    engine.connectInBackground();
  }

  const transport = new StdioServerTransport();
  await mcp.connect(transport);

  const clientVersion = mcp.getClientVersion();
  if (clientVersion?.name) {
    engine.identify(clientVersion.name);
  }

  process.on("SIGINT", () => {
    engine.close();
    process.exit(0);
  });
  process.on("SIGTERM", () => {
    engine.close();
    process.exit(0);
  });

  const exitOnStdinClose = () => {
    engine.close();
    process.exit(0);
  };
  process.stdin.on("end", exitOnStdinClose);
  process.stdin.on("close", exitOnStdinClose);
}

main().catch((err) => {
  console.error("Fatal:", err);
  process.exit(1);
});
