import { API_URL } from "./config";

export async function getHealth(): Promise<{ ok: boolean }> {
  const res = await fetch(`${API_URL}/health`, {
    credentials: "include",
    headers: { "Content-Type": "application/json" }
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}
