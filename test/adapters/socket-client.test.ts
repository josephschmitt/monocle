import { describe, it, expect, beforeEach, afterEach } from "vitest";
import net from "node:net";
import os from "node:os";
import path from "node:path";
import fs from "node:fs";
import { SocketClient } from "../../src/adapters/socket-client.js";
import type { HookMessage, HookResponse } from "../../src/types/protocol.js";

function tmpSocketPath(): string {
  return path.join(
    fs.mkdtempSync(path.join(os.tmpdir(), "monocle-test-")),
    "test.sock",
  );
}

function makeMessage(
  overrides: Partial<HookMessage> = {},
): HookMessage {
  return {
    type: "post_tool_use",
    request_id: "req-1",
    session_id: "sess-1",
    timestamp: Date.now(),
    tool_name: "write_file",
    tool_input: { path: "foo.ts" },
    tool_output: "ok",
    ...overrides,
  } as HookMessage;
}

describe("SocketClient", () => {
  let server: net.Server;
  let serverSockets: net.Socket[];
  let socketPath: string;
  let client: SocketClient;

  beforeEach(() => {
    socketPath = tmpSocketPath();
    serverSockets = [];
  });

  afterEach(async () => {
    client?.close();
    for (const s of serverSockets) {
      s.destroy();
    }
    await new Promise<void>((resolve) => {
      if (server?.listening) {
        server.close(() => resolve());
      } else {
        resolve();
      }
    });
  });

  function startServer(
    onConnection?: (socket: net.Socket) => void,
  ): Promise<void> {
    return new Promise((resolve) => {
      server = net.createServer((socket) => {
        serverSockets.push(socket);
        onConnection?.(socket);
      });
      server.listen(socketPath, resolve);
    });
  }

  it("connects to a UDS server", async () => {
    await startServer();
    client = new SocketClient(socketPath);

    expect(client.isConnected()).toBe(false);
    await client.connect();
    expect(client.isConnected()).toBe(true);
  });

  it("rejects connect when socket does not exist", async () => {
    client = new SocketClient("/tmp/monocle-nonexistent-socket.sock");

    await expect(client.connect()).rejects.toThrow();
    expect(client.isConnected()).toBe(false);
  });

  it("sends a message as NDJSON", async () => {
    const received: string[] = [];

    await startServer((socket) => {
      socket.on("data", (chunk) => received.push(chunk.toString()));
    });

    client = new SocketClient(socketPath);
    await client.connect();

    const msg = makeMessage();
    client.send(msg);

    // Give the message time to arrive
    await new Promise((r) => setTimeout(r, 50));

    expect(received.length).toBe(1);
    const parsed = JSON.parse(received[0].trim());
    expect(parsed.request_id).toBe("req-1");
    expect(parsed.type).toBe("post_tool_use");
  });

  it("throws when sending on a disconnected client", () => {
    client = new SocketClient(socketPath);
    expect(() => client.send(makeMessage())).toThrow("Socket not connected");
  });

  it("sendAndWait resolves when server responds with matching request_id", async () => {
    await startServer((socket) => {
      let buf = "";
      socket.on("data", (chunk) => {
        buf += chunk.toString();
        const lines = buf.split("\n");
        buf = lines.pop()!;
        for (const line of lines) {
          if (!line.trim()) continue;
          const msg = JSON.parse(line);
          const response: HookResponse = {
            type: "pre_tool_use_response",
            request_id: msg.request_id,
            timestamp: Date.now(),
            allowed: true,
          } as HookResponse;
          socket.write(JSON.stringify(response) + "\n");
        }
      });
    });

    client = new SocketClient(socketPath);
    await client.connect();

    const msg = makeMessage({
      type: "pre_tool_use",
      request_id: "req-wait-1",
    } as Partial<HookMessage>);
    const response = await client.sendAndWait(msg);
    expect(response.request_id).toBe("req-wait-1");
    expect(response.type).toBe("pre_tool_use_response");
  });

  it("sendAndWait times out if no response arrives", async () => {
    await startServer(() => {
      // Server never responds
    });

    client = new SocketClient(socketPath);
    await client.connect();

    const msg = makeMessage({ request_id: "req-timeout" });
    await expect(client.sendAndWait(msg, 100)).rejects.toThrow("Timeout");
  });

  it("sendAndWait rejects when socket is not connected", async () => {
    client = new SocketClient(socketPath);
    await expect(client.sendAndWait(makeMessage())).rejects.toThrow(
      "Socket not connected",
    );
  });

  it("handles multiple concurrent sendAndWait calls", async () => {
    await startServer((socket) => {
      let buf = "";
      socket.on("data", (chunk) => {
        buf += chunk.toString();
        const lines = buf.split("\n");
        buf = lines.pop()!;
        for (const line of lines) {
          if (!line.trim()) continue;
          const msg = JSON.parse(line);
          // Respond in reverse order with a small delay
          setTimeout(() => {
            const response: HookResponse = {
              type: "stop_response",
              request_id: msg.request_id,
              timestamp: Date.now(),
              continue: true,
            } as HookResponse;
            socket.write(JSON.stringify(response) + "\n");
          }, msg.request_id === "req-a" ? 50 : 10);
        }
      });
    });

    client = new SocketClient(socketPath);
    await client.connect();

    const [resA, resB] = await Promise.all([
      client.sendAndWait(makeMessage({ request_id: "req-a" })),
      client.sendAndWait(makeMessage({ request_id: "req-b" })),
    ]);

    expect(resA.request_id).toBe("req-a");
    expect(resB.request_id).toBe("req-b");
  });

  it("close() cleans up and rejects pending requests", async () => {
    await startServer(() => {
      // Server never responds
    });

    client = new SocketClient(socketPath);
    await client.connect();

    const pending = client.sendAndWait(makeMessage({ request_id: "req-close" }));
    client.close();

    await expect(pending).rejects.toThrow("Socket closed");
    expect(client.isConnected()).toBe(false);
  });

  it("ignores malformed JSON lines from server", async () => {
    await startServer((socket) => {
      let buf = "";
      socket.on("data", (chunk) => {
        buf += chunk.toString();
        const lines = buf.split("\n");
        buf = lines.pop()!;
        for (const line of lines) {
          if (!line.trim()) continue;
          const msg = JSON.parse(line);
          // Send garbage then a valid response
          socket.write("not valid json\n");
          const response: HookResponse = {
            type: "pre_tool_use_response",
            request_id: msg.request_id,
            timestamp: Date.now(),
            allowed: false,
          } as HookResponse;
          socket.write(JSON.stringify(response) + "\n");
        }
      });
    });

    client = new SocketClient(socketPath);
    await client.connect();

    const msg = makeMessage({ request_id: "req-malformed" });
    const response = await client.sendAndWait(msg);
    expect(response.request_id).toBe("req-malformed");
  });

  it("connect() is idempotent when already connected", async () => {
    await startServer();
    client = new SocketClient(socketPath);

    await client.connect();
    await client.connect(); // Should not throw
    expect(client.isConnected()).toBe(true);
  });
});
