import { useState } from "react";
import { GalaxySettings } from "../components/GalaxySettigs";
import type { GalaxyGenerationConfig } from "../services/api";

type Player = {
  id: string;
  name: string;
  color: string;
};

const COLORS = [
  "#4f46e5", "#dc2626", "#059669", "#d97706", "#7c3aed", "#eab308", "#14b8a6", "#f43f5e"
];

const PlayerCard = ({ id, name, color, handleColorChange }: PlayerCardProps) => {
  return (
    <div className="flex items-center gap-4 bg-[#222] rounded p-3">
      <div
        className="w-8 h-8 rounded-full border-2"
        style={{ background: color, borderColor: color }}
      />
      <span className="font-medium">{name}</span>
      <div className="flex gap-2 ml-auto">
        {COLORS.map(c => (
          <button
            key={c}
            className={`w-6 h-6 rounded-full border-2 ${color === c ? "border-white" : "border-[#333]"}`}
            style={{ background: c }}
            onClick={() => handleColorChange(id, c)}
            aria-label={`Select color ${c}`}
          />
        ))}
      </div>
    </div>
  );
}

type PlayerCardProps = Player & {
  handleColorChange: (playerId: string, color: string) => void;
};


export default function NewGame() {
  const [players, setPlayers] = useState<Player[]>([
    { id: "1", name: "Host", color: COLORS[0] }
  ]);
  const [newPlayerName, setNewPlayerName] = useState("");
  const [galaxyConfig, setGalaxyConfig] = useState<GalaxyGenerationConfig>({
    numStarSystems: 100,
    shape: "spiral",
    hyperlaneDensity: 0.5,
    maxHyperlanesPerSystem: 4,
  });

  const handleAddPlayer = () => {
    if (!newPlayerName.trim()) return;
    const color = COLORS[players.length % COLORS.length];
    setPlayers([...players, { id: String(players.length + 1), name: newPlayerName.trim(), color }]);
    setNewPlayerName("");
  };

  const handleColorChange = (playerId: string, color: string) => {
    setPlayers(players.map(p => p.id === playerId ? { ...p, color } : p));
  };

  return (
    <div className="min-h-screen bg-[#222] text-white flex flex-col items-center py-8">
      <h1 className="text-3xl font-bold mb-6">Create New Game</h1>
      <div className="flex flex-row gap-8 w-full max-w-5xl">
        {/* Players Section */}
        <section className="flex-1 bg-[#181818] rounded-lg p-6 shadow-lg">
          <h2 className="text-xl font-semibold mb-4">Players</h2>
          <div className="flex flex-col gap-4 mb-6">
            {players.map(player => (
              <PlayerCard key={player.id} {...player} handleColorChange={handleColorChange} />
            ))}
          </div>
          <div className="flex gap-2 mt-2">
            <input
              type="text"
              placeholder="Player name"
              value={newPlayerName}
              onChange={e => setNewPlayerName(e.target.value)}
              className="p-2 rounded bg-[#333] text-white border-none flex-1"
            />
            <button
              onClick={handleAddPlayer}
              className="bg-blue-600 px-4 py-2 rounded text-white font-semibold"
            >
              Add Player
            </button>
          </div>
        </section>
        {/* Galaxy Settings Section */}
        <section className="flex-1">
          <GalaxySettings value={galaxyConfig} onChange={setGalaxyConfig} />
        </section>
      </div>
      <button
        className="mt-8 px-6 py-3 bg-green-600 rounded text-white font-bold text-lg"
        onClick={() => {/* TODO: Start game logic */}}
      >
        Start Game
      </button>
    </div>
  );
}
