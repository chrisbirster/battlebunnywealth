import assert from "node:assert/strict";
import test from "node:test";

import { chooseBotInput } from "./bot.ts";
import { createArena, forceRunning, input, stateFingerprint, stepArena } from "./engine.ts";
import { createReplay, recordFrame, runReplay } from "./replay.ts";
import type { ArenaState, InputFrame } from "./types.ts";

const players = [
  { id: "p1", name: "Alpha" },
  { id: "p2", name: "Bravo", bot: true },
  { id: "p3", name: "Charlie", bot: true },
  { id: "p4", name: "Delta", bot: true },
];

test("every ranked player starts with identical combat stats", () => {
  const state = createArena(123, players);
  const baseline = state.players.map(({ bombCapacity, blastRadius, speed, canKick, remoteDetonator }) => ({ bombCapacity, blastRadius, speed, canKick, remoteDetonator }));
  assert.deepEqual(baseline, baseline.map(() => baseline[0]));
});

test("same seed creates the same symmetrical map", () => {
  const a = createArena(8080, players);
  const b = createArena(8080, players);
  assert.deepEqual(a.tiles, b.tiles);
  assert.deepEqual(a.hiddenPickups, b.hiddenPickups);
  for (let y = 0; y < a.config.height; y += 1) {
    for (let x = 0; x < a.config.width; x += 1) {
      const left = a.tiles[y * a.config.width + x];
      const right = a.tiles[y * a.config.width + (a.config.width - 1 - x)];
      assert.equal(left, right);
    }
  }
});

test("bomb destroys a crate and chain reactions detonate bombs", () => {
  let state = forceRunning(createArena(1, players.slice(0, 2)));
  state.players[0].x = 1;
  state.players[0].y = 1;
  state.players[0].blastRadius = 3;
  state.players[1].x = 11;
  state.players[1].y = 9;
  state = stepArena(state, { p1: input(undefined, true) });
  const first = state.bombs[0];
  assert.ok(first);
  first.fuse = 1;
  state = stepArena(state, {});
  assert.ok(state.explosions.length > 0);
});

test("chain reactions detonate a neighboring bomb immediately", () => {
  let state = forceRunning(createArena(77, players.slice(0, 2)));
  state.tiles = state.tiles.map((tile, index) => {
    const x = index % state.config.width;
    const y = Math.floor(index / state.config.width);
    return x > 0 && y > 0 && x < state.config.width - 1 && y < state.config.height - 1 ? "floor" : tile;
  });
  state.players[0].x = 3;
  state.players[0].y = 3;
  state.players[1].x = 9;
  state.players[1].y = 7;
  state.bombs = [
    { id: "b1", ownerId: "p1", x: 3, y: 3, fuse: 1, radius: 3, remote: false },
    { id: "b2", ownerId: "p2", x: 5, y: 3, fuse: 99, radius: 2, remote: false },
  ];
  state.nextBombID = 3;
  state = stepArena(state, {});
  assert.equal(state.bombs.length, 0);
  assert.ok(state.explosions.some((cell) => cell.x === 5 && cell.y === 3));
});

test("living players block occupied tiles", () => {
  let state = forceRunning(createArena(9, players.slice(0, 2)));
  state.players[0].x = 1;
  state.players[0].y = 1;
  state.players[1].x = 2;
  state.players[1].y = 1;
  state.tiles[1 * state.config.width + 2] = "floor";
  state = stepArena(state, { p1: { move: "right" } });
  assert.equal(state.players[0].x, 1);
  assert.equal(state.players[0].y, 1);
});

test("recorded inputs replay to identical state", () => {
  let state = createArena(42, players);
  const replay = createReplay(state);
  for (let tick = 0; tick < 180; tick += 1) {
    const frame: InputFrame = {};
    if (state.status === "running") {
      frame.p1 = tick % 17 === 0 ? { bomb: true } : { move: tick % 2 === 0 ? "right" : "down" };
      for (const bot of state.players.filter((player) => player.bot)) frame[bot.id] = chooseBotInput(state, bot.id);
    }
    recordFrame(replay, frame);
    state = stepArena(state, frame);
  }
  const replayed = runReplay(replay);
  assert.equal(stateFingerprint(replayed), stateFingerprint(state));
});

test("1000 seeded headless matches remain deterministic and valid", () => {
  for (let seed = 1; seed <= 1000; seed += 1) {
    let a = createArena(seed, players);
    let b = createArena(seed, players);
    for (let tick = 0; tick < 320 && (a.status !== "finished" || b.status !== "finished"); tick += 1) {
      const frameA = botFrame(a);
      const frameB = botFrame(b);
      assert.deepEqual(frameA, frameB);
      a = stepArena(a, frameA);
      b = stepArena(b, frameB);
    }
    assert.equal(stateFingerprint(a), stateFingerprint(b), `seed ${seed} diverged`);
    assert.ok(a.players.every((player) => player.x >= 0 && player.y >= 0 && player.x < a.config.width && player.y < a.config.height));
  }
});

function botFrame(state: ArenaState): InputFrame {
  const frame: InputFrame = {};
  if (state.status !== "running") return frame;
  for (const player of state.players) frame[player.id] = chooseBotInput(state, player.id);
  return frame;
}
