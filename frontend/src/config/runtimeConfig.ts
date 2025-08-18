export interface RuntimeConfig {
  apiUrl: string;
  wsUrl: string;
}

let runtimeConfig: RuntimeConfig = {
  apiUrl: "",
  wsUrl: "",
};

export async function loadRuntimeConfig() {
  const response = await fetch("/config.json");
  runtimeConfig = await response.json();
}

export function getRuntimeConfig(): RuntimeConfig {
  return runtimeConfig;
}

