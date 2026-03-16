import net from "node:net";
import type { HookMessage, HookResponse } from "../types/protocol.js";

const DEFAULT_TIMEOUT_MS = 5 * 60 * 1000; // 5 minutes

interface PendingRequest {
  resolve: (response: HookResponse) => void;
  reject: (error: Error) => void;
  timer: ReturnType<typeof setTimeout>;
}

export class SocketClient {
  private socket: net.Socket | null = null;
  private buffer = "";
  private pending = new Map<string, PendingRequest>();
  private connected = false;

  constructor(private readonly socketPath: string) {}

  /**
   * Connect to the Unix domain socket.
   * Resolves when connected; rejects on ENOENT or other connection errors.
   */
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.connected) {
        resolve();
        return;
      }

      const socket = net.createConnection(this.socketPath);

      socket.on("connect", () => {
        this.socket = socket;
        this.connected = true;
        resolve();
      });

      socket.on("data", (chunk: Buffer) => {
        this.handleData(chunk.toString());
      });

      socket.on("error", (err: NodeJS.ErrnoException) => {
        if (!this.connected) {
          reject(err);
          return;
        }
        this.cleanup();
      });

      socket.on("close", () => {
        this.cleanup();
      });
    });
  }

  /**
   * Fire-and-forget: serialize message as NDJSON and write to socket.
   */
  send(message: HookMessage): void {
    if (!this.socket || !this.connected) {
      throw new Error("Socket not connected");
    }
    this.socket.write(JSON.stringify(message) + "\n");
  }

  /**
   * Send a message and wait for a response with matching request_id.
   * Throws on timeout (default 5 minutes).
   */
  sendAndWait(
    message: HookMessage,
    timeoutMs: number = DEFAULT_TIMEOUT_MS,
  ): Promise<HookResponse> {
    return new Promise((resolve, reject) => {
      if (!this.socket || !this.connected) {
        reject(new Error("Socket not connected"));
        return;
      }

      const timer = setTimeout(() => {
        this.pending.delete(message.request_id);
        reject(
          new Error(
            `Timeout waiting for response to request ${message.request_id}`,
          ),
        );
      }, timeoutMs);

      this.pending.set(message.request_id, { resolve, reject, timer });
      this.socket.write(JSON.stringify(message) + "\n");
    });
  }

  /**
   * Close the socket connection.
   */
  close(): void {
    if (this.socket) {
      this.socket.destroy();
    }
    this.cleanup();
  }

  /**
   * Whether the client has an active connection.
   */
  isConnected(): boolean {
    return this.connected;
  }

  private handleData(data: string): void {
    this.buffer += data;
    const lines = this.buffer.split("\n");
    // Keep the last (possibly incomplete) chunk in the buffer
    this.buffer = lines.pop()!;

    for (const line of lines) {
      if (line.trim() === "") continue;
      try {
        const response = JSON.parse(line) as HookResponse;
        const pending = this.pending.get(response.request_id);
        if (pending) {
          clearTimeout(pending.timer);
          this.pending.delete(response.request_id);
          pending.resolve(response);
        }
      } catch {
        // Ignore malformed JSON lines
      }
    }
  }

  private cleanup(): void {
    this.connected = false;
    this.socket = null;
    this.buffer = "";
    for (const [id, pending] of this.pending) {
      clearTimeout(pending.timer);
      pending.reject(new Error("Socket closed"));
      this.pending.delete(id);
    }
  }
}
