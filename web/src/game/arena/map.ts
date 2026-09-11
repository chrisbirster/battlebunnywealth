import { randomInt } from "./rng.ts";
import type { ArenaConfig, ArenaPickup, Point, TileKind } from "./types.ts";

export const STARTS: Point[] = [
  { x: 1, y: 1 },
  { x: 11, y: 1 },
  { x: 1, y: 9 },
  { x: 11, y: 9 },
];

const PICKUPS = ["bomb", "blast", "speed", "kick", "remote"] as const;

export function tileIndex(config: ArenaConfig, x: number, y: number): number {
  return y * config.width + x;
}

export function tileAt(tiles: TileKind[], config: ArenaConfig, x: number, y: number): TileKind {
  if (x < 0 || y < 0 || x >= config.width || y >= config.height) return "wall";
  return tiles[tileIndex(config, x, y)] ?? "wall";
}

export function setTile(tiles: TileKind[], config: ArenaConfig, x: number, y: number, value: TileKind): void {
  tiles[tileIndex(config, x, y)] = value;
}

function spawnSafe(x: number, y: number, config: ArenaConfig): boolean {
  return STARTS.some((spawn) => {
    if (spawn.x >= config.width || spawn.y >= config.height) return false;
    const dx = Math.abs(spawn.x - x);
    const dy = Math.abs(spawn.y - y);
    return dx + dy <= 2;
  });
}

function canonical(x: number, y: number, config: ArenaConfig): string {
  const mx = config.width - 1 - x;
  const my = config.height - 1 - y;
  const points = [
    `${x},${y}`,
    `${mx},${y}`,
    `${x},${my}`,
    `${mx},${my}`,
  ].sort();
  return points[0];
}

export function generateArena(seed: number, config: ArenaConfig): {
  tiles: TileKind[];
  hiddenPickups: ArenaPickup[];
  rngState: number;
} {
  const tiles = Array<TileKind>(config.width * config.height).fill("floor");
  const hiddenPickups: ArenaPickup[] = [];
  const groupRolls = new Map<string, { crate: boolean; pickup?: (typeof PICKUPS)[number] }>();
  let rngState = seed >>> 0 || 0x51f15e;

  for (let y = 0; y < config.height; y += 1) {
    for (let x = 0; x < config.width; x += 1) {
      if (x === 0 || y === 0 || x === config.width - 1 || y === config.height - 1 || (x % 2 === 0 && y % 2 === 0)) {
        setTile(tiles, config, x, y, "wall");
        continue;
      }
      if (spawnSafe(x, y, config)) continue;

      const key = canonical(x, y, config);
      let roll = groupRolls.get(key);
      if (!roll) {
        const crateRoll = randomInt(rngState, 100);
        rngState = crateRoll.state;
        const crate = crateRoll.value < 66;
        roll = { crate };
        if (crate) {
          const pickupRoll = randomInt(rngState, 100);
          rngState = pickupRoll.state;
          if (pickupRoll.value < 42) {
            const kindRoll = randomInt(rngState, PICKUPS.length);
            rngState = kindRoll.state;
            roll.pickup = PICKUPS[kindRoll.value];
          }
        }
        groupRolls.set(key, roll);
      }

      if (roll.crate) {
        setTile(tiles, config, x, y, "crate");
        if (roll.pickup) hiddenPickups.push({ x, y, kind: roll.pickup });
      }
    }
  }

  return { tiles, hiddenPickups, rngState };
}
