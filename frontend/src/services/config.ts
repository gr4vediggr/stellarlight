import { getRuntimeConfig } from "../config/runtimeConfig";

export function getApiUrl() {
  return getRuntimeConfig().apiUrl;
}

export function getWsUrl() {
  return getRuntimeConfig().wsUrl;
}
