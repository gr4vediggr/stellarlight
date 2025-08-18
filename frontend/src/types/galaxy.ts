export type Galaxy = {
  id: string;
  name: string;
  starSystems: StarSystem[];
};

export type StarSystem = {
  id: string;
  name: string;
  ownerId?: string;
  locationX: number;
  locationY: number;
  connectedSystems: string[];
  stars: Star[];
  colonies?: Colony[];
};

export type Star = {
  id: string;
  name: string;
  type: string;
  size: number;
  locationX?: number;
  locationY?: number;
  planets: Planet[];
};

export type Planet = {
  id: string;
  name: string;
  type: string;
  size: number;
  orbitRadius: number;
  angle: number;
  colonies?: Colony[];
};

export type Colony = {
  id: string;
  name: string;
  ownerId: string;
  population: number;
  planetId: string;
  resources?: Record<string, number>;
};

export type Fleet = {
  id: string;
  starSystemId: string;
};

export type FleetMovement = {
  fleetId: string;
  destinationId: string;
};