import * as stylex from "@stylexjs/stylex";
import { For, Show, createSignal, onCleanup } from "solid-js";

import Shell from "../components/Shell";
import {
  chooseBotInput,
  controllerInput,
  createArena,
  createReplay,
  explosionAt,
  pickupAt,
  playerAt,
  recordFrame,
  runReplay,
  stepArena,
  tileAt,
  type ArenaPlayer,
  type ArenaReplay,
  type ArenaState,
  type ControllerSnapshot,
  type Direction,
  type InputFrame,
} from "../game/arena";

const PLAYERS = [
  { id: "you", name: "Your Bunny", icon: "🐰" },
  { id: "bot-1", name: "Dustpaw", icon: "🐇", bot: true },
  { id: "bot-2", name: "Cinder", icon: "🐇", bot: true },
  { id: "bot-3", name: "Hopper", icon: "🐇", bot: true },
];

const CELL_COUNT = 13 * 11;

export default function ArenaRoute() {
  const [seed, setSeed] = createSignal(20260911);
  const [arena, setArena] = createSignal(createArena(seed(), PLAYERS));
  const [lastReplay, setLastReplay] = createSignal<ArenaReplay>();
  const [replayMode, setReplayMode] = createSignal(false);
  const held = new Set<string>();
  let bombPressed = false;
  let remotePressed = false;
  let controllerBombDown = false;
  let controllerRemoteDown = false;
  let recorder = createReplay(arena());
  let timer: ReturnType<typeof setInterval> | undefined;

  const start = (nextSeed = seed() + 1) => {
    if (timer) clearInterval(timer);
    held.clear();
    bombPressed = false;
    remotePressed = false;
    controllerBombDown = false;
    controllerRemoteDown = false;
    setReplayMode(false);
    setSeed(nextSeed);
    const next = createArena(nextSeed, PLAYERS);
    recorder = createReplay(next);
    setArena(next);
    timer = setInterval(tick, 120);
  };

  const tick = () => {
    const current = arena();
    if (current.status === "finished") {
      if (timer) clearInterval(timer);
      setLastReplay(structuredClone(recorder));
      return;
    }

    const controller = controllerInput(browserControllerSnapshot());
    const controllerBombPressed = controller.bombDown && !controllerBombDown;
    const controllerRemotePressed = controller.remoteDown && !controllerRemoteDown;
    controllerBombDown = controller.bombDown;
    controllerRemoteDown = controller.remoteDown;

    const frame: InputFrame = {};
    frame.you = {
      move: keyboardDirection(held) ?? controller.move,
      bomb: bombPressed || controllerBombPressed,
      remote: remotePressed || controllerRemotePressed,
    };
    bombPressed = false;
    remotePressed = false;
    for (const player of current.players.filter((value) => value.bot)) frame[player.id] = chooseBotInput(current, player.id);
    recordFrame(recorder, frame);
    setArena(stepArena(current, frame));
  };

  const playReplay = () => {
    const replay = lastReplay();
    if (!replay) return;
    if (timer) clearInterval(timer);
    setReplayMode(true);
    let frameIndex = 0;
    let current = createArena(replay.seed, replay.players, replay.config);
    setArena(current);
    timer = setInterval(() => {
      if (frameIndex >= replay.frames.length) {
        if (timer) clearInterval(timer);
        setReplayMode(false);
        setArena(runReplay(replay));
        return;
      }
      current = stepArena(current, replay.frames[frameIndex]);
      frameIndex += 1;
      setArena(current);
    }, 45);
  };

  const touchMove = (move: Direction) => {
    held.clear();
    held.add(`touch-${move}`);
    window.setTimeout(() => held.delete(`touch-${move}`), 160);
  };

  const down = (event: KeyboardEvent) => {
    const key = event.key.toLowerCase();
    if (["arrowup", "arrowdown", "arrowleft", "arrowright", "w", "a", "s", "d", " ", "e"].includes(key)) event.preventDefault();
    held.add(key);
    if (key === " ") bombPressed = true;
    if (key === "e") remotePressed = true;
  };
  const up = (event: KeyboardEvent) => held.delete(event.key.toLowerCase());
  window.addEventListener("keydown", down);
  window.addEventListener("keyup", up);
  timer = setInterval(tick, 120);
  onCleanup(() => {
    if (timer) clearInterval(timer);
    window.removeEventListener("keydown", down);
    window.removeEventListener("keyup", up);
  });

  return (
    <Shell>
      <section {...stylex.props(styles.wrap)}>
        <div {...stylex.props(styles.heading)}>
          <div>
            <div {...stylex.props(styles.eyebrow)}>WARREN WARS / v0.2</div>
            <h1 {...stylex.props(styles.title)}>Equal start. Pure skill.</h1>
            <p {...stylex.props(styles.lede)}>Every ranked bunny enters with one bomb, blast radius two, and identical movement. All advantages are earned inside this match.</p>
          </div>
          <div {...stylex.props(styles.actions)}>
            <button type="button" onClick={() => start()} {...stylex.props(styles.primary)}>New seeded match</button>
            <button type="button" disabled={!lastReplay()} onClick={playReplay} {...stylex.props(styles.secondary)}>Replay last match</button>
          </div>
        </div>

        <div {...stylex.props(styles.statusbar)}>
          <span>Seed <strong>{seed()}</strong></span>
          <span>Tick <strong>{arena().matchTick}</strong></span>
          <span>Status <strong>{replayMode() ? "replay" : arena().status}</strong></span>
          <span>Starting stats <strong>1 bomb · radius 2 · speed 1</strong></span>
        </div>

        <div {...stylex.props(styles.layout)}>
          <div>
            <div {...stylex.props(styles.board)} style={{ "grid-template-columns": `repeat(${arena().config.width}, minmax(20px, 1fr))` }}>
              <For each={Array.from({ length: CELL_COUNT }, (_, index) => index)}>{(index) => {
                const x = index % arena().config.width;
                const y = Math.floor(index / arena().config.width);
                return <ArenaCell state={arena()} x={x} y={y} />;
              }}</For>
              <Show when={arena().status === "countdown"}>
                <div {...stylex.props(styles.overlay)}>{Math.max(1, Math.ceil(arena().countdownRemaining / 8))}</div>
              </Show>
              <Show when={arena().status === "finished"}>
                <div {...stylex.props(styles.overlay, styles.finish)}>{winnerText(arena())}</div>
              </Show>
            </div>

            <div {...stylex.props(styles.controls)}>
              <div {...stylex.props(styles.controlGrid)}>
                <span />
                <button type="button" onClick={() => touchMove("up")}>↑</button>
                <span />
                <button type="button" onClick={() => touchMove("left")}>←</button>
                <button type="button" onClick={() => touchMove("down")}>↓</button>
                <button type="button" onClick={() => touchMove("right")}>→</button>
              </div>
              <button type="button" onClick={() => { bombPressed = true; }} {...stylex.props(styles.bombButton)}>💣 Bomb</button>
              <button type="button" onClick={() => { remotePressed = true; }} {...stylex.props(styles.remoteButton)}>⚡ Remote</button>
            </div>
            <p {...stylex.props(styles.help)}>Keyboard: WASD / arrows · Space bomb · E remote. Controller: left stick / D-pad · A bomb · B remote. Touch controls work on mobile.</p>
          </div>

          <aside {...stylex.props(styles.sidebar)}>
            <div {...stylex.props(styles.panel)}>
              <div {...stylex.props(styles.eyebrow)}>COMBATANTS</div>
              <For each={arena().players}>{(player) => <CombatantRow player={player} />}</For>
            </div>
            <div {...stylex.props(styles.panel)}>
              <div {...stylex.props(styles.eyebrow)}>PICKUPS</div>
              <p>💣 +bomb · ✹ +blast · ↯ +speed</p>
              <p>🥾 kick bombs · 📡 remote detonator</p>
              <p {...stylex.props(styles.muted)}>Pickups exist only for the current match. Bunny Bucks, CARROT, authority, and account age never alter ranked starting power.</p>
            </div>
          </aside>
        </div>
      </section>
    </Shell>
  );
}

function CombatantRow(props: { player: ArenaPlayer }) {
  if (!props.player.alive) {
    return (
      <div {...stylex.props(styles.playerRow, styles.dead)}>
        <span>{props.player.icon ?? "🐰"}</span>
        <div><strong>{props.player.name}</strong><small>{props.player.bot ? "BOT" : "YOU"}</small></div>
        <div {...stylex.props(styles.stats)}>💣{props.player.bombCapacity} ✹{props.player.blastRadius} ↯{props.player.speed}</div>
      </div>
    );
  }
  return (
    <div {...stylex.props(styles.playerRow)}>
      <span>{props.player.icon ?? "🐰"}</span>
      <div><strong>{props.player.name}</strong><small>{props.player.bot ? "BOT" : "YOU"}</small></div>
      <div {...stylex.props(styles.stats)}>💣{props.player.bombCapacity} ✹{props.player.blastRadius} ↯{props.player.speed}</div>
    </div>
  );
}

function ArenaCell(props: { state: ArenaState; x: number; y: number }) {
  const tile = () => tileAt(props.state.tiles, props.state.config, props.x, props.y);
  const player = () => playerAt(props.state, props.x, props.y);
  const bomb = () => props.state.bombs.find((value) => value.x === props.x && value.y === props.y);
  const pickup = () => pickupAt(props.state, props.x, props.y);
  const explosion = () => explosionAt(props.state, props.x, props.y);
  const glyph = () => {
    if (player()) return player()?.icon ?? "🐰";
    if (bomb()) return "💣";
    if (pickup()?.kind === "bomb") return "💣";
    if (pickup()?.kind === "blast") return "✹";
    if (pickup()?.kind === "speed") return "↯";
    if (pickup()?.kind === "kick") return "🥾";
    if (pickup()?.kind === "remote") return "📡";
    return "";
  };

  if (explosion()) return <div {...stylex.props(styles.cell, styles.explosion)}>{glyph()}</div>;
  if (tile() === "wall") return <div {...stylex.props(styles.cell, styles.wall)}>{glyph()}</div>;
  if (tile() === "crate") return <div {...stylex.props(styles.cell, styles.crate)}>{glyph()}</div>;
  return <div {...stylex.props(styles.cell, styles.floor)}>{glyph()}</div>;
}

function browserControllerSnapshot(): ControllerSnapshot | undefined {
  const pad = navigator.getGamepads?.()[0];
  if (!pad) return undefined;
  return {
    axes: Array.from(pad.axes),
    buttons: pad.buttons.map((button) => button.pressed),
  };
}

function keyboardDirection(keys: Set<string>): Direction | undefined {
  if (keys.has("arrowup") || keys.has("w") || keys.has("touch-up")) return "up";
  if (keys.has("arrowdown") || keys.has("s") || keys.has("touch-down")) return "down";
  if (keys.has("arrowleft") || keys.has("a") || keys.has("touch-left")) return "left";
  if (keys.has("arrowright") || keys.has("d") || keys.has("touch-right")) return "right";
  return undefined;
}

function winnerText(state: ArenaState): string {
  if (state.winnerIds.length === 0) return "DRAW";
  const winner = state.players.find((player) => player.id === state.winnerIds[0]);
  return `${winner?.name ?? "Bunny"} WINS`;
}

const styles = stylex.create({
  wrap: { maxWidth: "1220px", margin: "0 auto", padding: "56px 24px" },
  heading: { display: "flex", justifyContent: "space-between", gap: "24px", alignItems: "end", flexWrap: "wrap" },
  eyebrow: { color: "#d9f13b", fontSize: "11px", fontWeight: 900, letterSpacing: "0.16em" },
  title: { fontSize: "clamp(44px, 7vw, 78px)", lineHeight: 0.95, letterSpacing: "-0.055em", margin: "10px 0" },
  lede: { maxWidth: "760px", color: "#aeb4a4", lineHeight: 1.6, fontSize: "18px" },
  actions: { display: "flex", gap: "10px", flexWrap: "wrap" },
  primary: { border: 0, borderRadius: "10px", backgroundColor: "#d9f13b", color: "#10120e", fontWeight: 900, padding: "12px 15px", cursor: "pointer" },
  secondary: { border: "1px solid #49513f", borderRadius: "10px", backgroundColor: "#20241c", color: "#f1efdc", fontWeight: 800, padding: "12px 15px", cursor: "pointer", ":disabled": { opacity: 0.4, cursor: "not-allowed" } },
  statusbar: { display: "flex", flexWrap: "wrap", gap: "18px", margin: "30px 0 14px", padding: "12px 14px", backgroundColor: "#171a14", border: "1px solid #333a2c", borderRadius: "10px", color: "#8f9785", fontSize: "12px" },
  layout: { display: "grid", gridTemplateColumns: "minmax(0, 780px) minmax(230px, 1fr)", gap: "18px", alignItems: "start" },
  board: { position: "relative", display: "grid", aspectRatio: "13 / 11", width: "100%", overflow: "hidden", border: "3px solid #68704f", borderRadius: "16px", backgroundColor: "#11140f", boxShadow: "0 24px 80px rgba(0,0,0,.35)" },
  cell: { minWidth: 0, display: "grid", placeItems: "center", fontSize: "clamp(12px, 2.4vw, 28px)", userSelect: "none", transition: "transform 80ms ease, background-color 100ms ease, opacity 100ms ease" },
  floor: { backgroundColor: "#283026", border: "1px solid #30382d" },
  wall: { backgroundColor: "#626a58", border: "2px solid #747d68", boxShadow: "inset 0 0 0 2px #515949" },
  crate: { backgroundColor: "#80643e", border: "2px solid #9a7950", boxShadow: "inset 0 0 0 2px #684e31" },
  explosion: { backgroundColor: "#f1c84a", transform: "scale(1.08)", boxShadow: "inset 0 0 18px #fff1a6, 0 0 18px #f1c84a", zIndex: 2 },
  overlay: { position: "absolute", inset: 0, display: "grid", placeItems: "center", backgroundColor: "rgba(8,10,7,.56)", fontSize: "clamp(72px, 18vw, 180px)", fontWeight: 950, color: "#d9f13b", zIndex: 4, pointerEvents: "none" },
  finish: { fontSize: "clamp(38px, 9vw, 100px)", textAlign: "center" },
  controls: { display: "flex", gap: "14px", alignItems: "center", flexWrap: "wrap", marginTop: "16px" },
  controlGrid: { display: "grid", gridTemplateColumns: "repeat(3, 44px)", gap: "4px" },
  bombButton: { minHeight: "48px", border: 0, borderRadius: "12px", backgroundColor: "#d9f13b", color: "#11130f", padding: "0 18px", fontWeight: 900, cursor: "pointer" },
  remoteButton: { minHeight: "48px", border: "1px solid #59624d", borderRadius: "12px", backgroundColor: "#282e24", color: "#f1efdc", padding: "0 18px", fontWeight: 900, cursor: "pointer" },
  help: { color: "#777f70", fontSize: "12px", lineHeight: 1.5 },
  sidebar: { display: "grid", gap: "14px" },
  panel: { border: "1px solid #343b2e", backgroundColor: "#181c16", borderRadius: "14px", padding: "18px" },
  playerRow: { display: "grid", gridTemplateColumns: "30px 1fr auto", gap: "10px", alignItems: "center", padding: "12px 0", borderBottom: "1px solid #2d3328" },
  dead: { opacity: 0.35, textDecoration: "line-through" },
  stats: { color: "#d9f13b", fontSize: "11px", whiteSpace: "nowrap" },
  muted: { color: "#838b7a", fontSize: "12px", lineHeight: 1.55 },
});
