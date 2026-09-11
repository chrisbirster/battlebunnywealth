import { createArena, stateFingerprint, stepArena } from "./engine.ts";
import type { ArenaConfig, ArenaPlayerSpec, ArenaState, InputFrame } from "./types.ts";

export type ArenaReplay = {
  version: 1;
  seed: number;
  players: ArenaPlayerSpec[];
  config: ArenaConfig;
  frames: InputFrame[];
};

export function createReplay(state: ArenaState): ArenaReplay {
  return {
    version: 1,
    seed: state.seed,
    players: state.players.map(({ id, name, icon, bot }) => ({ id, name, icon, bot })),
    config: { ...state.config },
    frames: [],
  };
}

export function recordFrame(replay: ArenaReplay, inputs: InputFrame): void {
  replay.frames.push(structuredClone(inputs));
}

export function runReplay(replay: ArenaReplay): ArenaState {
  let state = createArena(replay.seed, replay.players, replay.config);
  for (const frame of replay.frames) state = stepArena(state, frame);
  return state;
}

export function verifyReplay(replay: ArenaReplay, expected: ArenaState): boolean {
  return stateFingerprint(runReplay(replay)) === stateFingerprint(expected);
}
