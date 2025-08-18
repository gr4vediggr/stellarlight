import { useState } from "react";
import { Galaxy } from "../../proto/galaxy";
import { generateGalaxy, type GalaxyGenerationConfig } from "../../services/api";
import * as React from "react";
import { GalaxyMap } from "../../components/GalaxyMap";
import { GalaxySettings } from "../../components/GalaxySettigs";

export default function GalaxyCreatorV2() {
  const [config, setConfig] = useState<GalaxyGenerationConfig>({
    numStarSystems: 100,
    shape: "spiral",
    hyperlaneDensity: 0.5,
    maxHyperlanesPerSystem: 4,
  });
  const [loading, setLoading] = useState(false);
  const [galaxy, setGalaxy] = useState<Galaxy | null>(null);
  const [error, setError] = useState<string | null>(null);
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
    <div className="flex bg-[#222] h-full text-white overflow-hidden">
      {/* Sidebar */}
      <aside className="w-80 bg-[#181818] p-8 flex flex-col gap-6 border-r border-[#333] h-full">
        <h2 className="text-[1.7rem] mb-4">Galaxy Creator</h2>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            handleGenerate();
          }}
        >
          <GalaxySettings value={config} onChange={setConfig} />
          <button
            type="submit"
            disabled={loading}
            className="bg-[#2d7dff] text-white border-none rounded px-4 py-2 font-semibold cursor-pointer mt-4 disabled:opacity-60"
          >
            {loading ? "Generating..." : "Generate Galaxy"}
          </button>
        </form>
        {error && <div className="text-[#ff6b6b] mb-4">{error}</div>}
      </aside>
      {/* Main content */}
      <main className="flex-1 flex flex-col min-h-0 h-full overflow-hidden">
        {galaxy ? (
          <GalaxyMap galaxy={galaxy} />
        ) : <div className="flex-1"></div>}
      </main>
    </div>
  );
}
