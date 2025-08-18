import React from "react";
import type { GalaxyGenerationConfig, GalaxyShape } from "../services/api";

interface GalaxySettingsProps {
  value: GalaxyGenerationConfig;
  onChange: (config: GalaxyGenerationConfig) => void;
}

export class GalaxySettings extends React.Component<GalaxySettingsProps> {
  handleChange = (field: keyof GalaxyGenerationConfig, value: any) => {
    const newConfig = { ...this.props.value, [field]: value };
    this.props.onChange(newConfig);
  };

  render() {
    const config = this.props.value;
    return (
      <div className="galaxy-settings bg-[#181818] rounded-lg p-6 shadow-lg max-w-md w-full mx-auto text-white">
        <h2 className="text-2xl font-bold mb-6 text-center">Galaxy Settings</h2>
        <form className="flex flex-col gap-4">
          <label className="flex flex-col gap-1">
            <span className="font-medium">Number of Star Systems:</span>
            <input
              type="number"
              value={config.numStarSystems}
              onChange={(e) => this.handleChange("numStarSystems", +e.target.value)}
              className="p-2 rounded bg-[#222] text-white border border-[#333] focus:outline-none focus:ring-2 focus:ring-blue-500"
              min={1}
              max={10000}
            />
          </label>
          <label className="flex flex-col gap-1">
            <span className="font-medium">Shape:</span>
            <select
              value={config.shape}
              onChange={(e) => this.handleChange("shape", e.target.value as GalaxyShape)}
              className="p-2 rounded bg-[#222] text-white border border-[#333] focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="spiral">Spiral</option>
              <option value="elliptical">Elliptical</option>
              <option value="irregular">Irregular</option>
            </select>
          </label>
          <label className="flex flex-col gap-1">
            <span className="font-medium">Hyperlane Density:</span>
            <input
              type="number"
              step="0.1"
              value={config.hyperlaneDensity}
              onChange={(e) => this.handleChange("hyperlaneDensity", +e.target.value)}
              className="p-2 rounded bg-[#222] text-white border border-[#333] focus:outline-none focus:ring-2 focus:ring-blue-500"
              min={0}
              max={1}
            />
          </label>
          <label className="flex flex-col gap-1">
            <span className="font-medium">Max Hyperlanes per System:</span>
            <input
              type="number"
              value={config.maxHyperlanesPerSystem}
              onChange={(e) => this.handleChange("maxHyperlanesPerSystem", +e.target.value)}
              className="p-2 rounded bg-[#222] text-white border border-[#333] focus:outline-none focus:ring-2 focus:ring-blue-500"
              min={1}
              max={20}
            />
          </label>
        </form>
      </div>
    );
  }
}

