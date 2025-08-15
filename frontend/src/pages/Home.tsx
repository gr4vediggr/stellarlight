import { useNavigate } from "react-router-dom";
import { useUserStore } from "../state/userStore";
import { getHealth } from "../services/api";
import { useState } from "react";

export default function Home() {
  const { username, setUsername } = useUserStore();
  const [health, setHealth] = useState<string>("unknown");
  const navigate = useNavigate();

  const checkBackend = async () => {
    try {
      const res = await getHealth();
      setHealth(res.ok ? "ok" : "down");
    } catch {
      setHealth("down");
    }
  };

  return (
    <section>
      <h1 style={{ fontSize: "2rem", marginBottom: "1rem" }}>Welcome, Commander</h1>
      <p style={{ opacity: 0.8, marginBottom: "1.5rem" }}>
        Prepare your fleet. Log in with a codename and jump into the galaxy.
      </p>

      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <input
          value={username ?? ""}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="Enter codename"
          style={{ padding: "0.6rem 0.75rem", borderRadius: 8, border: "1px solid #263042", background: "#0f1420", color: "#e6e6e6", width: 260 }}
        />
        <button
          onClick={() => navigate("/play")}
          style={{ padding: "0.6rem 1rem", borderRadius: 8, border: "none", cursor: "pointer", background: "#2d7dff", color: "#fff", fontWeight: 600 }}
        >
          Enter Sector
        </button>
      </div>

      <button
        onClick={checkBackend}
        style={{ padding: "0.5rem 0.9rem", borderRadius: 8, border: "1px solid #2a3750", background: "transparent", color: "#e6e6e6" }}
      >
        Check Backend Health: <b style={{ marginLeft: 6 }}>{health}</b>
      </button>
    </section>
  );
}
