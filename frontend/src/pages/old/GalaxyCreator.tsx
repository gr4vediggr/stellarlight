import { useState } from "react";
import type { Galaxy } from "../../proto/galaxy";
import { generateGalaxy, type GalaxyGenerationConfig } from "../../services/api";
import * as React from "react";
import { GalaxyMap } from "../../components/GalaxyMap";

export default function GalaxyCreator() {
  const [config, setConfig] = useState<GalaxyGenerationConfig>({
    numStarSystems: 100,
    shape: "spiral",
    hyperlaneDensity: 0.5,
    maxHyperlanesPerSystem: 4,
  });
  const [loading, setLoading] = useState(false);
  const [galaxy, setGalaxy] = useState<Galaxy | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedSystemId, setSelectedSystemId] = React.useState<string | undefined>(undefined);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setConfig((prev) => ({
      ...prev,
      [name]:
        name === "numStarSystems" || name === "maxHyperlanesPerSystem"
          ? parseInt(value)
          : name === "hyperlaneDensity"
          ? parseFloat(value)
          : value,
    }));
  };

  const handleGenerate = async () => {
    setLoading(true);
    setError(null);
    setGalaxy(null);
    try {
      const data = await generateGalaxy(config);
      setGalaxy(data);
    } catch (err: any) {
      setError(err.message || "Unknown error");
    } finally {
      setLoading(false);
    }
  };

  return (
<div style={{
      display: "flex",
      height : "100%",
      overflow: "hidden",
      background: "#222",
      color: "#fff"
    }}>
      {/* Sidebar */}
        <aside style={{
        width: 320,
        background: "#181818",
        padding: "2rem 1.5rem",
        height: "100%",
        flexDirection: "column",
        gap: "1.5rem",
        borderRight: "1px solid #333",
      }}>
        <h2 style={{ fontSize: "1.7rem", marginBottom: "1rem" }}>Galaxy Creator</h2>
        <form
          style={{ display: "flex", flexDirection: "column", gap: "1rem" }}
          onSubmit={(e) => {
            e.preventDefault();
            handleGenerate();
          }}
        >
          <label>
            Number of Star Systems:
            <input
              type="number"
              name="numStarSystems"
              min={1}
              max={10000}
              value={config.numStarSystems}
              onChange={handleChange}
              style={{ marginLeft: 8, padding: "0.4rem", borderRadius: 6, border: "none", background: "#333", color: "#fff", width: "100%" }}
            />
          </label>
          <label>
            Shape:
            <select
              name="shape"
              value={config.shape}
              onChange={handleChange}
              style={{ marginLeft: 8, padding: "0.4rem", borderRadius: 6, background: "#333", color: "#fff", width: "100%" }}
            >
              <option value="spiral">Spiral</option>
              <option value="elliptical">Elliptical</option>
              <option value="ring">Ring</option>
              <option value="irregular">Irregular</option>
            </select>
          </label>
          <label>
            Hyperlane Density:
            <input
              type="number"
              name="hyperlaneDensity"
              min={0}
              max={1}
              step={0.01}
              value={config.hyperlaneDensity}
              onChange={handleChange}
              style={{ marginLeft: 8, padding: "0.4rem", borderRadius: 6, border: "none", background: "#333", color: "#fff", width: "100%" }}
            />
          </label>
          <label>
            Max Hyperlanes Per System:
            <input
              type="number"
              name="maxHyperlanesPerSystem"
              min={1}
              max={20}
              value={config.maxHyperlanesPerSystem}
              onChange={handleChange}
              style={{ marginLeft: 8, padding: "0.4rem", borderRadius: 6, border: "none", background: "#333", color: "#fff", width: "100%" }}
            />
          </label>
          <button
            type="submit"
            disabled={loading}
            style={{
              background: "#2d7dff",
              color: "#fff",
              border: "none",
              borderRadius: 8,
              padding: "0.6rem 1rem",
              fontWeight: 600,
              cursor: "pointer",
              marginTop: "1rem"
            }}
          >
            {loading ? "Generating..." : "Generate Galaxy"}
          </button>
        </form>
        {error && <div style={{ color: "#ff6b6b", marginBottom: "1rem" }}>{error}</div>}
      </aside>
      {/* Main content */}
     <main style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        minHeight: 0, // critical to prevent flex child overflow
        overflow: "hidden", // prevent scrollbars
      }}>

        

        {galaxy ? (

        <GalaxyMap galaxy={galaxy} 
          selectedSystemId={selectedSystemId}
          onSelectSystem={setSelectedSystemId}/>
        ) : <div style={{ flex: 1 }}></div>}
        
      </main>
    </div>
  );
}
