export type Direction = "up" | "down" | "left" | "right";
export type TileKind = "floor" | "wall" | "crate";
export type PickupKind = "bomb" | "blast" | "speed" | "kick" | "remote";
export type MatchStatus = "countdown" | "running" | "finished";

export type Point = { x: number; y: number };

export type ArenaPlayerSpec = {
  id: string;
  name: string;
  icon?: string;
  bot?: boolean;
};

export type ArenaPlayer = ArenaPlayerSpec & Point & {
  alive: boolean;
  bombCapacity: number;
  blastRadius: number;
  speed: number;
  moveCooldown: number;
  canKick: boolean;
  remoteDetonator: boolean;
  bombsPlaced: number;
};

export type ArenaBomb = Point & {
  id: string;
  ownerId: string;
  fuse: number;
  radius: number;
  remote: boolean;
};

export type ArenaExplosion = Point & {
  ttl: number;
};

export type ArenaPickup = Point & {
  kind: PickupKind;
};

export type PlayerInput = {
  move?: Direction;
  bomb?: boolean;
  remote?: boolean;
};

export type InputFrame = Record<string, PlayerInput>;

export type ArenaConfig = {
  width: number;
  height: number;
  countdownTicks: number;
  bombFuseTicks: number;
  explosionTicks: number;
  maxMatchTicks: number;
};

export type HiddenPickup = ArenaPickup;

export type ArenaState = {
  seed: number;
  tick: number;
  matchTick: number;
  status: MatchStatus;
  countdownRemaining: number;
  config: ArenaConfig;
  tiles: TileKind[];
  hiddenPickups: HiddenPickup[];
  players: ArenaPlayer[];
  bombs: ArenaBomb[];
  explosions: ArenaExplosion[];
  pickups: ArenaPickup[];
  winnerIds: string[];
  nextBombID: number;
  rngState: number;
};

export const DEFAULT_ARENA_CONFIG: ArenaConfig = {
  width: 13,
  height: 11,
  countdownTicks: 24,
  bombFuseTicks: 24,
  explosionTicks: 3,
  maxMatchTicks: 1_800,
};
