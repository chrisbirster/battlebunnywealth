import { generateArena, STARTS, setTile, tileAt } from "./map.ts";
import type {
  ArenaBomb,
  ArenaConfig,
  ArenaExplosion,
  ArenaPlayer,
  ArenaPlayerSpec,
  ArenaState,
  Direction,
  InputFrame,
  PlayerInput,
  Point,
} from "./types.ts";
import { DEFAULT_ARENA_CONFIG } from "./types.ts";

const DIR: Record<Direction, Point> = {
  up: { x: 0, y: -1 },
  down: { x: 0, y: 1 },
  left: { x: -1, y: 0 },
  right: { x: 1, y: 0 },
};

export function createArena(seed: number, specs: ArenaPlayerSpec[], config: ArenaConfig = DEFAULT_ARENA_CONFIG): ArenaState {
  if (specs.length < 2 || specs.length > 4) throw new Error("Warren Wars supports 2-4 players in v0.2");
  const generated = generateArena(seed, config);
  const players: ArenaPlayer[] = specs.map((spec, index) => {
    const spawn = STARTS[index];
    if (!spawn || spawn.x >= config.width || spawn.y >= config.height) throw new Error("arena does not have enough spawn points");
    return {
      ...spec,
      ...spawn,
      alive: true,
      bombCapacity: 1,
      blastRadius: 2,
      speed: 1,
      moveCooldown: 0,
      canKick: false,
      remoteDetonator: false,
      bombsPlaced: 0,
    };
  });

  return {
    seed,
    tick: 0,
    matchTick: 0,
    status: "countdown",
    countdownRemaining: config.countdownTicks,
    config,
    tiles: generated.tiles,
    hiddenPickups: generated.hiddenPickups,
    players,
    bombs: [],
    explosions: [],
    pickups: [],
    winnerIds: [],
    nextBombID: 1,
    rngState: generated.rngState,
  };
}

export function stepArena(previous: ArenaState, rawInputs: InputFrame = {}): ArenaState {
  const state = cloneState(previous);
  state.tick += 1;
  state.explosions = state.explosions.map((value) => ({ ...value, ttl: value.ttl - 1 })).filter((value) => value.ttl > 0);

  if (state.status === "finished") return state;
  if (state.status === "countdown") {
    state.countdownRemaining -= 1;
    if (state.countdownRemaining <= 0) state.status = "running";
    return state;
  }

  state.matchTick += 1;
  for (const player of state.players) player.moveCooldown = Math.max(0, player.moveCooldown - 1);

  const inputs = normalizeInputs(state, rawInputs);
  resolveMoves(state, inputs);
  resolveActions(state, inputs);
  tickBombs(state);
  applyExplosionDamage(state);
  collectPickups(state);
  resolveWinner(state);
  return state;
}

function normalizeInputs(state: ArenaState, inputs: InputFrame): InputFrame {
  const normalized: InputFrame = {};
  for (const player of state.players) {
    if (!player.alive) continue;
    const value = inputs[player.id] ?? {};
    normalized[player.id] = {
      move: value.move,
      bomb: Boolean(value.bomb),
      remote: Boolean(value.remote),
    };
  }
  return normalized;
}

function resolveMoves(state: ArenaState, inputs: InputFrame): void {
  const occupied = new Set(state.players.filter((p) => p.alive).map((p) => pointKey(p.x, p.y)));
  const proposed = new Map<string, Point>();
  const claims = new Map<string, number>();

  for (const player of [...state.players].sort((a, b) => a.id.localeCompare(b.id))) {
    const move = inputs[player.id]?.move;
    if (!player.alive || !move || player.moveCooldown > 0) continue;
    const delta = DIR[move];
    const target = { x: player.x + delta.x, y: player.y + delta.y };
    if (tileAt(state.tiles, state.config, target.x, target.y) !== "floor") continue;

    const bomb = bombAt(state, target.x, target.y);
    if (bomb) {
      if (!player.canKick || !tryKickBomb(state, bomb, move, occupied)) continue;
    }
    if (occupied.has(pointKey(target.x, target.y))) continue;
    proposed.set(player.id, target);
    claims.set(pointKey(target.x, target.y), (claims.get(pointKey(target.x, target.y)) ?? 0) + 1);
  }

  for (const player of state.players) {
    const target = proposed.get(player.id);
    if (!target || (claims.get(pointKey(target.x, target.y)) ?? 0) > 1) continue;
    player.x = target.x;
    player.y = target.y;
    player.moveCooldown = Math.max(1, 3 - Math.min(player.speed, 2));
  }
}

function tryKickBomb(state: ArenaState, bomb: ArenaBomb, direction: Direction, occupied: Set<string>): boolean {
  const delta = DIR[direction];
  const target = { x: bomb.x + delta.x, y: bomb.y + delta.y };
  if (tileAt(state.tiles, state.config, target.x, target.y) !== "floor") return false;
  if (occupied.has(pointKey(target.x, target.y)) || bombAt(state, target.x, target.y)) return false;
  bomb.x = target.x;
  bomb.y = target.y;
  return true;
}

function resolveActions(state: ArenaState, inputs: InputFrame): void {
  for (const player of state.players) {
    if (!player.alive) continue;
    const input = inputs[player.id] ?? {};
    if (input.remote && player.remoteDetonator) {
      const remote = state.bombs
        .filter((bomb) => bomb.ownerId === player.id && bomb.remote)
        .sort((a, b) => Number(a.id.slice(1)) - Number(b.id.slice(1)))[0];
      if (remote) remote.fuse = 0;
    }

    if (!input.bomb) continue;
    const active = state.bombs.filter((bomb) => bomb.ownerId === player.id).length;
    if (active >= player.bombCapacity || bombAt(state, player.x, player.y)) continue;
    state.bombs.push({
      id: `b${state.nextBombID++}`,
      ownerId: player.id,
      x: player.x,
      y: player.y,
      fuse: player.remoteDetonator ? state.config.bombFuseTicks * 4 : state.config.bombFuseTicks,
      radius: player.blastRadius,
      remote: player.remoteDetonator,
    });
    player.bombsPlaced += 1;
  }
}

function tickBombs(state: ArenaState): void {
  for (const bomb of state.bombs) bomb.fuse -= 1;
  const queue = state.bombs.filter((bomb) => bomb.fuse <= 0).map((bomb) => bomb.id);
  const exploded = new Set<string>();

  while (queue.length > 0) {
    const id = queue.shift();
    if (!id || exploded.has(id)) continue;
    const bomb = state.bombs.find((value) => value.id === id);
    if (!bomb) continue;
    exploded.add(id);

    const blast = blastCells(state, bomb);
    for (const cell of blast.cells) {
      addExplosion(state, cell.x, cell.y);
      const chained = bombAt(state, cell.x, cell.y);
      if (chained && !exploded.has(chained.id)) queue.push(chained.id);
    }
    for (const crate of blast.crates) destroyCrate(state, crate.x, crate.y);
  }

  if (exploded.size > 0) state.bombs = state.bombs.filter((bomb) => !exploded.has(bomb.id));
}

function blastCells(state: ArenaState, bomb: ArenaBomb): { cells: Point[]; crates: Point[] } {
  const cells: Point[] = [{ x: bomb.x, y: bomb.y }];
  const crates: Point[] = [];
  for (const direction of Object.values(DIR)) {
    for (let distance = 1; distance <= bomb.radius; distance += 1) {
      const x = bomb.x + direction.x * distance;
      const y = bomb.y + direction.y * distance;
      const tile = tileAt(state.tiles, state.config, x, y);
      if (tile === "wall") break;
      cells.push({ x, y });
      if (tile === "crate") {
        crates.push({ x, y });
        break;
      }
    }
  }
  return { cells, crates };
}

function destroyCrate(state: ArenaState, x: number, y: number): void {
  if (tileAt(state.tiles, state.config, x, y) !== "crate") return;
  setTile(state.tiles, state.config, x, y, "floor");
  const hidden = state.hiddenPickups.find((pickup) => pickup.x === x && pickup.y === y);
  if (hidden && !state.pickups.some((pickup) => pickup.x === x && pickup.y === y)) {
    state.pickups.push({ ...hidden });
  }
  state.hiddenPickups = state.hiddenPickups.filter((pickup) => !(pickup.x === x && pickup.y === y));
}

function addExplosion(state: ArenaState, x: number, y: number): void {
  const existing = state.explosions.find((value) => value.x === x && value.y === y);
  if (existing) existing.ttl = Math.max(existing.ttl, state.config.explosionTicks);
  else state.explosions.push({ x, y, ttl: state.config.explosionTicks });
}

function applyExplosionDamage(state: ArenaState): void {
  const blast = new Set(state.explosions.map((value) => pointKey(value.x, value.y)));
  for (const player of state.players) {
    if (player.alive && blast.has(pointKey(player.x, player.y))) player.alive = false;
  }
}

function collectPickups(state: ArenaState): void {
  for (const player of state.players) {
    if (!player.alive) continue;
    const index = state.pickups.findIndex((pickup) => pickup.x === player.x && pickup.y === player.y);
    if (index < 0) continue;
    const [pickup] = state.pickups.splice(index, 1);
    if (pickup.kind === "bomb") player.bombCapacity = Math.min(6, player.bombCapacity + 1);
    if (pickup.kind === "blast") player.blastRadius = Math.min(8, player.blastRadius + 1);
    if (pickup.kind === "speed") player.speed = Math.min(2, player.speed + 1);
    if (pickup.kind === "kick") player.canKick = true;
    if (pickup.kind === "remote") player.remoteDetonator = true;
  }
}

function resolveWinner(state: ArenaState): void {
  const alive = state.players.filter((player) => player.alive);
  if (alive.length <= 1) {
    state.status = "finished";
    state.winnerIds = alive.map((player) => player.id);
    return;
  }
  if (state.matchTick >= state.config.maxMatchTicks) {
    state.status = "finished";
    state.winnerIds = [];
  }
}

export function bombAt(state: ArenaState, x: number, y: number): ArenaBomb | undefined {
  return state.bombs.find((bomb) => bomb.x === x && bomb.y === y);
}

export function explosionAt(state: ArenaState, x: number, y: number): ArenaExplosion | undefined {
  return state.explosions.find((value) => value.x === x && value.y === y);
}

export function playerAt(state: ArenaState, x: number, y: number): ArenaPlayer | undefined {
  return state.players.find((player) => player.alive && player.x === x && player.y === y);
}

export function pickupAt(state: ArenaState, x: number, y: number) {
  return state.pickups.find((pickup) => pickup.x === x && pickup.y === y);
}

export function stateFingerprint(state: ArenaState): string {
  return JSON.stringify({
    seed: state.seed,
    tick: state.tick,
    matchTick: state.matchTick,
    status: state.status,
    tiles: state.tiles,
    players: state.players,
    bombs: state.bombs,
    explosions: state.explosions,
    pickups: state.pickups,
    winners: state.winnerIds,
  });
}

export function isDangerous(state: ArenaState, x: number, y: number): boolean {
  if (state.explosions.some((value) => value.x === x && value.y === y)) return true;
  for (const bomb of state.bombs) {
    if (bomb.fuse > 8) continue;
    if (bomb.x === x && Math.abs(bomb.y - y) <= bomb.radius && clearLine(state, bomb.x, bomb.y, x, y)) return true;
    if (bomb.y === y && Math.abs(bomb.x - x) <= bomb.radius && clearLine(state, bomb.x, bomb.y, x, y)) return true;
  }
  return false;
}

function clearLine(state: ArenaState, fromX: number, fromY: number, toX: number, toY: number): boolean {
  const dx = Math.sign(toX - fromX);
  const dy = Math.sign(toY - fromY);
  let x = fromX + dx;
  let y = fromY + dy;
  while (x !== toX || y !== toY) {
    if (tileAt(state.tiles, state.config, x, y) !== "floor") return false;
    x += dx;
    y += dy;
  }
  return true;
}

function pointKey(x: number, y: number): string {
  return `${x}:${y}`;
}

function cloneState(state: ArenaState): ArenaState {
  return {
    ...state,
    config: { ...state.config },
    tiles: [...state.tiles],
    hiddenPickups: state.hiddenPickups.map((value) => ({ ...value })),
    players: state.players.map((value) => ({ ...value })),
    bombs: state.bombs.map((value) => ({ ...value })),
    explosions: state.explosions.map((value) => ({ ...value })),
    pickups: state.pickups.map((value) => ({ ...value })),
    winnerIds: [...state.winnerIds],
  };
}

export function forceRunning(state: ArenaState): ArenaState {
  const next = cloneState(state);
  next.status = "running";
  next.countdownRemaining = 0;
  return next;
}

export function input(move?: Direction, bomb = false, remote = false): PlayerInput {
  return { move, bomb, remote };
}
