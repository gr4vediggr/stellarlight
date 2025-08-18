import * as React from "react";

import type { Galaxy } from "../../types/galaxy";
import { GalaxyMap } from "../../components/GalaxyMap";



type GalaxyScreenProps = {
  galaxy: Galaxy;
};

export function GalaxyScreen({ galaxy }: GalaxyScreenProps) {
  const [selectedSystemId, setSelectedSystemId] = React.useState<string | undefined>(undefined);
  const selectedSystem = galaxy.starSystems.find(s => s.id === selectedSystemId);

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "row",
        width: "100vw",
        height: "100%",
        overflow: "hidden",
        background: "#101010",
        fontFamily: "sans-serif",
      }}
    >
      {/* Galaxy Map */}
      <div style={{ flex: 1, position: "relative" }}>
        <GalaxyMap
          galaxy={galaxy}
          selectedSystemId={selectedSystemId}
          onSelectSystem={setSelectedSystemId}

        />
      </div>

      {/* Info Panel */}
      <div
        style={{
          width: 300,
          background: "#1c1c1c",
          color: "#fff",
          padding: 16,
          boxSizing: "border-box",
          overflowY: "auto",
        }}
      >
        {selectedSystem ? (
          <div>
            <h2 style={{ marginTop: 0 }}>{selectedSystem.name}</h2>
            <p>
              <strong>Owner:</strong> {selectedSystem.ownerId || "Neutral"}
            </p>
            <p>
              <strong>Stars:</strong> {selectedSystem.stars.length}
            </p>
            <p>
              <strong>Colonies:</strong> {selectedSystem.colonies?.length || 0}
            </p>
            <h3>Connected Systems</h3>
            <ul>
              {selectedSystem.connectedSystems.map(connId => {
                const conn = galaxy.starSystems.find(s => s.id === connId);
                if (!conn) return null;
                return <li key={connId}>{conn.name}</li>;
              })}
            </ul>
          </div>
        ) : (
          <p>Select a system to see details.</p>
        )}
      </div>
    </div>
  );
}
