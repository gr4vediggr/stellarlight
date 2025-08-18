import { getWsUrl } from "./config";

export function createGameSocket(): WebSocket {
  const ws = new WebSocket(getWsUrl());
  ws.binaryType = "arraybuffer"; // for protobuf frames if needed
  return ws;
}
