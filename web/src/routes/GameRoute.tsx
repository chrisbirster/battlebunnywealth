import * as stylex from "@stylexjs/stylex";
import { createSignal } from "solid-js";

import Shell from "../components/Shell";

export default function GameRoute() {
  const [cash, setCash] = createSignal(473);
  const [scrap, setScrap] = createSignal(12);

  const collect = () => {
    setCash((value) => value + 128);
    setScrap((value) => value + 3);
  };

  return (
    <Shell>
      <section {...stylex.props(styles.wrap)}>
        <div {...stylex.props(styles.topline)}>
          <div>
            <div {...stylex.props(styles.eyebrow)}>WARREN COMMAND / PROTOTYPE</div>
            <h1 {...stylex.props(styles.title)}>Your unit is broke.</h1>
          </div>
          <div {...stylex.props(styles.wallet)}>
            <span>${cash().toLocaleString()}</span>
            <span>{scrap()} scrap</span>
          </div>
        </div>

        <div {...stylex.props(styles.board)}>
          <article {...stylex.props(styles.operation)}>
            <div {...stylex.props(styles.icon)}>🥕</div>
            <div>
              <div {...stylex.props(styles.small)}>OPERATION 001</div>
              <h2>Questionable Carrot Logistics</h2>
              <p>Private Stuffy found a pallet nobody seems to be looking for.</p>
            </div>
            <button type="button" onClick={collect} {...stylex.props(styles.collect)}>Collect $128 + 3 scrap</button>
          </article>

          <article {...stylex.props(styles.operation)}>
            <div {...stylex.props(styles.icon)}>🔧</div>
            <div>
              <div {...stylex.props(styles.small)}>WORKSHOP</div>
              <h2>Trashcan TNT</h2>
              <p>Corporal Boomboom says it is safe enough. Legal has declined comment.</p>
            </div>
            <button type="button" disabled={scrap() < 15} onClick={() => setScrap((value) => value - 15)} {...stylex.props(styles.craft)}>
              Craft / 15 scrap
            </button>
          </article>
        </div>

        <section {...stylex.props(styles.arena)}>
          <div>
            <div {...stylex.props(styles.eyebrow)}>SKILL MODE</div>
            <h2 {...stylex.props(styles.arenaTitle)}>Warren Wars</h2>
            <p {...stylex.props(styles.arenaText)}>Grid-based bomb combat is the active half of the loop. Movement, blast timing, destructible terrain, pickups, and bunny-specific abilities come next.</p>
          </div>
          <button type="button" disabled {...stylex.props(styles.deploy)}>Arena prototype coming next</button>
        </section>
      </section>
    </Shell>
  );
}

const styles = stylex.create({
  wrap: { maxWidth: "1180px", margin: "0 auto", padding: "64px 24px" },
  topline: { display: "flex", justifyContent: "space-between", gap: "24px", alignItems: "end", flexWrap: "wrap" },
  eyebrow: { color: "#d9f13b", fontSize: "12px", fontWeight: 900, letterSpacing: "0.16em" },
  title: { fontSize: "clamp(46px, 8vw, 84px)", margin: "10px 0 0", letterSpacing: "-0.055em" },
  wallet: { display: "flex", gap: "10px", padding: "12px", border: "1px solid #3a4132", borderRadius: "12px", backgroundColor: "#1b1f18", fontWeight: 900 },
  board: { display: "grid", gap: "14px", marginTop: "40px" },
  operation: { display: "grid", gridTemplateColumns: "70px 1fr auto", gap: "20px", alignItems: "center", backgroundColor: "#1a1e17", border: "1px solid #343a2c", borderRadius: "16px", padding: "22px" },
  icon: { fontSize: "38px", width: "64px", height: "64px", display: "grid", placeItems: "center", borderRadius: "14px", backgroundColor: "#292f24" },
  small: { color: "#818878", fontSize: "11px", letterSpacing: "0.12em", fontWeight: 900 },
  collect: { border: 0, borderRadius: "10px", backgroundColor: "#d9f13b", color: "#11130f", padding: "13px 16px", fontWeight: 900, cursor: "pointer" },
  craft: { border: "1px solid #4a5241", borderRadius: "10px", backgroundColor: "#252a21", color: "#f3f0db", padding: "13px 16px", fontWeight: 900, cursor: "pointer", ":disabled": { opacity: 0.45, cursor: "not-allowed" } },
  arena: { marginTop: "20px", padding: "34px", borderRadius: "18px", border: "1px solid #5e663f", background: "linear-gradient(135deg, #262b1d, #171a14)", display: "flex", justifyContent: "space-between", gap: "24px", alignItems: "center", flexWrap: "wrap" },
  arenaTitle: { fontSize: "42px", margin: "8px 0" },
  arenaText: { color: "#aeb4a4", maxWidth: "700px", lineHeight: 1.6 },
  deploy: { padding: "14px 18px", borderRadius: "10px", border: "1px solid #535b48", backgroundColor: "transparent", color: "#a6ad9d", fontWeight: 800 },
});
