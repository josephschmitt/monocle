import { McpServer } from "@anthropic-ai/sdk/mcp";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
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

class EngineConnection {
  private socketPath: string;
  private conn: ReturnType<typeof connect> | null = null;
  private pendingRequests = new Map<
    string,
    { resolve: (msg: Message) => void; reject: (err: Error) => void }
  >();
  private onEvent: (event: string, payload: Record<string, any>) => void;
  private lineBuffer = "";
  private reconnecting = false;
  private closed = false;

  constructor(
    socketPath: string,
    onEvent: (event: string, payload: Record<string, any>) => void,
  ) {
    this.socketPath = socketPath;
    this.onEvent = onEvent;
  }

  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!existsSync(this.socketPath)) {
        reject(new Error("Socket not found: " + this.socketPath));
        return;
      }

      this.conn = connect(this.socketPath, () => {
        // Send subscribe message
        const sub = JSON.stringify({
          type: "subscribe",
          events: [
            "feedback_submitted",
            "pause_changed",
            "content_item_added",
          ],
        });
        this.conn!.write(sub + "\n");
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
            if (!gotAck && msg.type === "subscribe_response") {
              gotAck = true;
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
          this.scheduleReconnect();
        }
      });

      this.conn.on("close", () => {
        if (gotAck && !this.closed) {
          this.scheduleReconnect();
        }
      });
    });
  }

  private handleMessage(msg: Message) {
    if (msg.type === "event_notification") {
      this.onEvent(msg.event, msg.payload || {});
      return;
    }

    // Response to a request — match by type
    const key = msg.type;
    const pending = this.pendingRequests.get(key);
    if (pending) {
      this.pendingRequests.delete(key);
      pending.resolve(msg);
    }
  }

  async request(msg: Message): Promise<Message> {
    if (!this.conn || this.conn.destroyed) {
      throw new Error("Not connected to monocle engine");
    }

    return new Promise((resolve, reject) => {
      const responseType = msg.type + "_response";
      this.pendingRequests.set(responseType, { resolve, reject });
      this.conn!.write(JSON.stringify(msg) + "\n");

      // Timeout after 30s
      setTimeout(() => {
        if (this.pendingRequests.has(responseType)) {
          this.pendingRequests.delete(responseType);
          reject(new Error("Request timed out: " + msg.type));
        }
      }, 30000);
    });
  }

  private scheduleReconnect() {
    if (this.reconnecting || this.closed) return;
    this.reconnecting = true;

    const attempt = (delay: number) => {
      if (this.closed) return;
      setTimeout(async () => {
        try {
          await this.connect();
          this.reconnecting = false;
        } catch {
          attempt(Math.min(delay * 2, 10000));
        }
      }, delay);
    };

    attempt(1000);
  }

  close() {
    this.closed = true;
    if (this.conn) {
      this.conn.destroy();
      this.conn = null;
    }
  }
}

// -- Blocking connection for get_feedback --wait --

async function blockingGetFeedback(socketPath: string): Promise<Message> {
  return new Promise((resolve, reject) => {
    const conn = connect(socketPath, () => {
      const msg = JSON.stringify({ type: "poll_feedback", wait: true });
      conn.write(msg + "\n");
    });

    conn.setEncoding("utf8");
    let buf = "";

    conn.on("data", (chunk: string) => {
      buf += chunk;
      const lines = buf.split("\n");
      buf = lines.pop() || "";

      for (const line of lines) {
        if (!line.trim()) continue;
        try {
          const msg = JSON.parse(line);
          resolve(msg);
          conn.destroy();
        } catch {
          // ignore
        }
      }
    });

    conn.on("error", (err: Error) => reject(err));
    conn.on("close", () => reject(new Error("Connection closed")));
  });
}

// -- Main --

const cwd = process.cwd();
const repoRoot = findRepoRoot(cwd);
const socketPath = defaultSocketPath(repoRoot);

// Channel notification callback
let pushNotification: ((params: { method: string; params: { level?: string; data?: any } }) => Promise<void>) | null =
  null;

// Engine connection with event handler
const engine = new EngineConnection(socketPath, (event, payload) => {
  if (!pushNotification) return;

  switch (event) {
    case "feedback_submitted":
      pushNotification({
        method: "notifications/message",
        params: {
          level: "warning",
          data: payload.message || "Your reviewer has submitted feedback.",
        },
      }).catch(() => {});
      break;
    case "pause_changed":
      if (payload.status === "pause_requested") {
        pushNotification({
          method: "notifications/message",
          params: {
            level: "warning",
            data:
              "Your reviewer has requested you pause and wait for feedback. " +
              "Use the get_feedback tool with wait=true to block until feedback is ready.",
          },
        }).catch(() => {});
      }
      break;
    case "content_item_added":
      // Informational — no push needed, the agent submitted this
      break;
  }
});

// Create MCP server
const server = new McpServer({
  name: "monocle",
  version: "1.0.0",
});

// -- Tools --

server.tool(
  "review_status",
  "Check if your reviewer has pending feedback or has requested a pause",
  {},
  async () => {
    try {
      const resp = await engine.request({ type: "get_review_status" });
      return {
        content: [
          { type: "text" as const, text: resp.summary || "No reviewer connected." },
        ],
      };
    } catch {
      return {
        content: [{ type: "text" as const, text: "No reviewer connected." }],
      };
    }
  },
);

server.tool(
  "get_feedback",
  "Retrieve review feedback from your reviewer. Use wait=true to block until feedback is available (pause flow).",
  {
    wait: {
      type: "boolean",
      description: "Block until feedback is available",
    },
  },
  async ({ wait }: { wait?: boolean }) => {
    try {
      let resp: Message;
      if (wait) {
        resp = await blockingGetFeedback(socketPath);
      } else {
        resp = await engine.request({ type: "poll_feedback", wait: false });
      }

      if (resp.has_feedback) {
        return {
          content: [{ type: "text" as const, text: resp.feedback }],
        };
      }
      return {
        content: [{ type: "text" as const, text: "No feedback pending." }],
      };
    } catch {
      return {
        content: [{ type: "text" as const, text: "No reviewer connected." }],
      };
    }
  },
);

server.tool(
  "submit_plan",
  "Submit a plan, architecture decision, or other content for your reviewer to see and comment on",
  {
    title: {
      type: "string",
      description: "Title for the plan or content",
    },
    content: {
      type: "string",
      description: "The plan or content body (markdown supported)",
    },
    id: {
      type: "string",
      description: "Optional ID for updating existing content",
    },
  },
  async ({
    title,
    content,
    id,
  }: {
    title: string;
    content: string;
    id?: string;
  }) => {
    try {
      const resp = await engine.request({
        type: "submit_content",
        id: id || "",
        title,
        content,
      });
      return {
        content: [
          {
            type: "text" as const,
            text: resp.message || "Content submitted for review.",
          },
        ],
      };
    } catch {
      return {
        content: [{ type: "text" as const, text: "No reviewer connected." }],
      };
    }
  },
);

// -- Start --

async function main() {
  // Connect to monocle engine (retry if not yet started)
  let connected = false;
  for (let i = 0; i < 5; i++) {
    try {
      await engine.connect();
      connected = true;
      break;
    } catch {
      await new Promise((r) => setTimeout(r, 2000));
    }
  }

  if (!connected) {
    console.error(
      "Warning: Could not connect to monocle engine. Will retry in background.",
    );
  }

  // Connect to Claude Code via stdio
  const transport = new StdioServerTransport();
  await server.connect(transport);

  // Register notification sender for pushing events
  pushNotification = async (params) => {
    try {
      await (server as any).server?.notification(params);
    } catch {
      // notification sending failed, ignore
    }
  };

  // Handle graceful shutdown
  process.on("SIGINT", () => {
    engine.close();
    process.exit(0);
  });
  process.on("SIGTERM", () => {
    engine.close();
    process.exit(0);
  });
}

main().catch((err) => {
  console.error("Fatal:", err);
  process.exit(1);
});
