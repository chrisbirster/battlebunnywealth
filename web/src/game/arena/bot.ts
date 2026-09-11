import { bombAt, isDangerous } from "./engine.ts";
import { tileAt } from "./map.ts";
import type { ArenaState, Direction, PlayerInput } from "./types.ts";

const DIRECTIONS: Direction[] = ["up", "left", "down", "right"];
const DELTA: Record<Direction, { x: number; y: number }> = {
  up: { x: 0, y: -1 },
  down: { x: 0, y: 1 },
  left: { x: -1, y: 0 },
  right: { x: 1, y: 0 },
};

export function chooseBotInput(state: ArenaState, playerID: string): PlayerInput {
  const player = state.players.find((value) => value.id === playerID);
  if (!player?.alive || state.status !== "running") return {};

  const orderOffset = Math.abs(hash(`${state.seed}:${state.tick}:${playerID}`)) % DIRECTIONS.length;
  const directions = [...DIRECTIONS.slice(orderOffset), ...DIRECTIONS.slice(0, orderOffset)];
  const safeMoves = directions.filter((direction) => {
    const d = DELTA[direction];
    const x = player.x + d.x;
    const y = player.y + d.y;
    return tileAt(state.tiles, state.config, x, y) === "floor" && !bombAt(state, x, y) && !isDangerous(state, x, y);
  });

  if (isDangerous(state, player.x, player.y) && safeMoves.length > 0) return { move: safeMoves[0] };

  const nearbyTarget = directions.some((direction) => {
    const d = DELTA[direction];
    const x = player.x + d.x;
    const y = player.y + d.y;
    return tileAt(state.tiles, state.config, x, y) === "crate" || state.players.some((other) => other.alive && other.id !== player.id && other.x === x && other.y === y);
  });

  if (nearbyTarget && state.bombs.filter((bomb) => bomb.ownerId === player.id).length < player.bombCapacity) {
    return { bomb: true };
  }

  if (player.remoteDetonator && state.bombs.some((bomb) => bomb.ownerId === player.id && bomb.remote && bomb.fuse < state.config.bombFuseTicks * 3)) {
    return { remote: true, move: safeMoves[0] };
  }

  return { move: safeMoves[0] };
}

function hash(value: string): number {
  let result = 2166136261;
  for (let i = 0; i < value.length; i += 1) {
    result ^= value.charCodeAt(i);
    result = Math.imul(result, 16777619);
  }
  return result | 0;
}
