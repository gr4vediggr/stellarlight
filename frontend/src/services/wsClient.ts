import { WS_URL } from "./config";

export function createGameSocket(): WebSocket {
  const ws = new WebSocket(WS_URL);
  ws.binaryType = "arraybuffer"; // for protobuf frames if needed
  return ws;
}
