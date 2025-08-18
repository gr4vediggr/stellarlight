import { getApiUrl } from "./config";

import { Galaxy } from "../proto/galaxy";


export async function getHealth(): Promise<{ ok: boolean }> {
  const res = await fetch(`${getApiUrl()}/health`, {
    headers: { "Content-Type": "application/json" }
  });

  console.log(`Health check response: ${res.status}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  
  return res.json();
}

export type GalaxyShape = "spiral" | "elliptical" | "ring" | "irregular";

export type GalaxyGenerationConfig = {
  numStarSystems: number;
  shape: GalaxyShape;
  hyperlaneDensity: number;
  maxHyperlanesPerSystem: number;
};


export async function generateGalaxy(config: GalaxyGenerationConfig): Promise<Galaxy> {

  const res = await fetch(`${getApiUrl()}/galaxy-generate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(config),
      });
      if (!res.ok) throw new Error("Failed to generate galaxy");
      const data = await res.json();

  return Galaxy.fromJSON(data);
}
