import { useEffect, useRef, useState } from "react";
import { createGameSocket } from "../../services/wsClient";
import { useUserStore } from "../../state/userStore";

export default function Play() {
  const { username } = useUserStore();
  const [status, setStatus] = useState("Connecting…");
  const [messages, setMessages] = useState<string[]>([]);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const ws = createGameSocket();
    wsRef.current = ws;

    ws.onopen = () => setStatus("Connected");
    ws.onmessage = (ev) => {
      // If you send protobuf frames, decode here.
      setMessages((m) => [...m, typeof ev.data === "string" ? ev.data : `[binary ${((ev.data as ArrayBuffer).byteLength)}b]`]);
    };
    ws.onerror = () => setStatus("Error");
    ws.onclose = () => setStatus("Closed");

    return () => ws.close();
  }, []);

  const sendPing = () => wsRef.current?.send(username ? `PING from ${username}` : "PING");

  return (
    <section>
      <h2 style={{ fontSize: "1.6rem", marginBottom: "0.5rem" }}>Sector Alpha</h2>
      <div style={{ marginBottom: "1rem", opacity: 0.8 }}>WebSocket: <b>{status}</b></div>
      <div style={{ display: "flex", gap: 8 }}>
        <button onClick={sendPing} style={{ padding: "0.5rem 0.9rem", borderRadius: 8, border: "none", background: "#2d7dff", color: "#fff" }}>
          Send Ping
        </button>
      </div>
      <div style={{ marginTop: "1rem", padding: "0.75rem", border: "1px solid #263042", borderRadius: 8 }}>
        <div style={{ fontWeight: 600, marginBottom: 6 }}>Events</div>
        <pre style={{ whiteSpace: "pre-wrap", margin: 0, opacity: 0.9 }}>
          {messages.length ? messages.join("\n") : "No events yet."}
        </pre>
      </div>
    </section>
  );
}
