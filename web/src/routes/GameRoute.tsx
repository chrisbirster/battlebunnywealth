import * as stylex from "@stylexjs/stylex";
import { For, Show, createSignal, onCleanup } from "solid-js";

import Shell from "../components/Shell";
import { gameApi, type GameSnapshot, type Standing } from "../lib/gameApi";

const briefings = [
  { npc: "First Sergeant Hard-as-Nails", icon: "🎖️", title: "Welcome to the Warren", text: "You are broke, recruit. Fix that. Build the businesses, keep the operation moving, and do not embarrass the unit." },
  { npc: "Private Stuffy", icon: "🧸", title: "Your first hustle", text: "Questionable Carrot Logistics is definitely a legitimate business. Nobody has proven otherwise." },
  { npc: "Captain Cashmere", icon: "📈", title: "Idle income", text: "Your businesses keep earning while you are away. Command currently banks up to eight hours of offline income." },
  { npc: "Corporal Boomboom", icon: "💣", title: "Separate economies", text: "All this money stays outside Warren Wars. When you enter ranked combat, every bunny starts equal." },
  { npc: "Da Champ", icon: "🥊", title: "Season target", text: "Hit the Season Zero earnings target, turn in the campaign, collect your status reward, and prove you can build it again." },
  { npc: "Doc Flopsy", icon: "🩹", title: "You are cleared", text: "Profile saved. Businesses running. Competitive firewall intact. Try not to stand on your own bomb." },
];

export default function GameRoute() {
  const [state, setState] = createSignal<GameSnapshot>();
  const [standings, setStandings] = createSignal<Standing[]>([]);
  const [message, setMessage] = createSignal("Loading Warren Command…");
  const [busy, setBusy] = createSignal("");

  const refresh = async (silent = false) => {
    try {
      const [next, board] = await Promise.all([gameApi.state(), gameApi.standings()]);
      setState(next);
      setStandings(board);
      if (!silent) setMessage(next.accruedBunnyBucks > 0 ? `Your businesses earned $${money(next.accruedBunnyBucks)} while you were away.` : "Warren Command online.");
    } catch (error) {
      if (!silent) setMessage(error instanceof Error ? error.message : "Could not load Warren Command.");
    }
  };

  const upgrade = async (id: string) => {
    setBusy(id);
    try {
      const next = await gameApi.upgrade(id);
      setState(next);
      setMessage("Business upgraded. Passive income increased.");
      setStandings(await gameApi.standings());
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Upgrade failed.");
    } finally { setBusy(""); }
  };

  const advanceBriefing = async () => {
    setBusy("briefing");
    try {
      const next = await gameApi.advanceOnboarding();
      setState(next);
      setMessage(next.onboardingStep >= briefings.length ? "Orientation complete. Build the warren your way." : "Briefing acknowledged.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Briefing update failed.");
    } finally { setBusy(""); }
  };

  const turnIn = async () => {
    setBusy("turn-in");
    try {
      const next = await gameApi.turnIn();
      setState(next);
      setMessage("Season turn-in complete. Cosmetic service reward added; businesses reset for another run.");
      setStandings(await gameApi.standings());
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Turn-in failed.");
    } finally { setBusy(""); }
  };

  void refresh();
  const timer = window.setInterval(() => void refresh(true), 5000);
  onCleanup(() => window.clearInterval(timer));

  const progress = () => {
    const current = state();
    if (!current) return 0;
    return Math.min(100, Math.round((current.season.earnings / current.seasonTurnInTarget) * 100));
  };

  const currentBriefing = () => {
    const step = state()?.onboardingStep ?? briefings.length;
    return step < briefings.length ? briefings[step] : undefined;
  };

  return (
    <Shell>
      <section {...stylex.props(styles.wrap)}>
        <div {...stylex.props(styles.topline)}>
          <div>
            <div {...stylex.props(styles.eyebrow)}>WARREN COMMAND / v0.3</div>
            <h1 {...stylex.props(styles.title)}>Build the warren.</h1>
            <p {...stylex.props(styles.lede)}>Businesses earn Bunny Bucks automatically—even while you are away. Upgrade the operation, trigger synergies, and push toward the Season Zero turn-in.</p>
          </div>
          <Show when={state()}>
            <div {...stylex.props(styles.wallet)}><span>${money(state()!.wallet.bunnyBucks)}</span><small>+${money(state()!.incomePerSecond)}/sec</small></div>
          </Show>
        </div>

        <div {...stylex.props(styles.notice)}>{message()}</div>

        <Show when={state()} fallback={<div {...stylex.props(styles.loading)}>Establishing command uplink…</div>}>
          <Show when={currentBriefing()}>
            <section {...stylex.props(styles.briefing)}>
              <div {...stylex.props(styles.npcIcon)}>{currentBriefing()!.icon}</div>
              <div>
                <div {...stylex.props(styles.small)}>ORIENTATION / {currentBriefing()!.npc}</div>
                <h2 {...stylex.props(styles.businessTitle)}>{currentBriefing()!.title}</h2>
                <p {...stylex.props(styles.description)}>{currentBriefing()!.text}</p>
              </div>
              <button type="button" disabled={busy() !== ""} onClick={() => void advanceBriefing()} {...stylex.props(styles.acknowledge)}>Acknowledge →</button>
            </section>
          </Show>

          <div {...stylex.props(styles.playerStrip)}>
            <div {...stylex.props(styles.avatar)}>🐰</div>
            <div><div {...stylex.props(styles.small)}>YOUR BATTLE BUNNY</div><strong>{state()!.player.name}</strong><span {...stylex.props(styles.callsign)}> “{state()!.player.callsign}”</span></div>
            <a href="/profile" {...stylex.props(styles.secondaryLink)}>Customize profile →</a>
          </div>

          <div {...stylex.props(styles.board)}>
            <For each={state()!.businesses}>{(business) => (
              <article {...stylex.props(styles.operation)}>
                <div {...stylex.props(styles.icon)}>{business.icon}</div>
                <div>
                  <div {...stylex.props(styles.small)}>LEVEL {business.level}</div>
                  <h2 {...stylex.props(styles.businessTitle)}>{business.name}</h2>
                  <p {...stylex.props(styles.description)}>{business.description}</p>
                  <div {...stylex.props(styles.meta)}><span>+${money(business.incomePerSecond)}/sec</span><span>{business.synergy}</span></div>
                </div>
                <button type="button" disabled={busy() !== "" || state()!.wallet.bunnyBucks < business.upgradeCost} onClick={() => void upgrade(business.id)} {...stylex.props(styles.upgrade)}>
                  {busy() === business.id ? "Upgrading…" : `Upgrade $${money(business.upgradeCost)}`}
                </button>
              </article>
            )}</For>
          </div>

          <div {...stylex.props(styles.seasonGrid)}>
            <section {...stylex.props(styles.season)}>
              <div {...stylex.props(styles.seasonTop)}>
                <div><div {...stylex.props(styles.eyebrow)}>SEASON ZERO / ALPHA</div><h2 {...stylex.props(styles.seasonTitle)}>Cash out the campaign.</h2><p {...stylex.props(styles.description)}>Turn-in rewards collection and status only. Ranked combat power stays equal.</p></div>
                <div {...stylex.props(styles.seasonNumbers)}><strong>${money(state()!.season.earnings)}</strong><span>of ${money(state()!.seasonTurnInTarget)}</span></div>
              </div>
              <div {...stylex.props(styles.progressTrack)}><div style={{ width: `${progress()}%` }} {...stylex.props(styles.progressFill)} /></div>
              <div {...stylex.props(styles.seasonActions)}><span>{progress()}% ready · {state()!.season.turnIns} prior turn-ins</span><button type="button" disabled={busy() !== "" || state()!.season.earnings < state()!.seasonTurnInTarget} onClick={() => void turnIn()} {...stylex.props(styles.turnIn)}>{busy() === "turn-in" ? "Turning in…" : "Turn in season"}</button></div>
            </section>

            <aside {...stylex.props(styles.standings)}>
              <div {...stylex.props(styles.eyebrow)}>LOCAL ALPHA STANDINGS</div>
              <For each={standings()}>{(entry) => <div {...stylex.props(styles.standingRow)}><strong>#{entry.rank}</strong><span>{entry.name}<small> “{entry.callsign}”</small></span><b>${money(entry.bunnyBucks)}</b></div>}</For>
              <p {...stylex.props(styles.muted)}>v0.3 has one local persistent player. Multi-account standings arrive after account identity exists.</p>
            </aside>
          </div>

          <section {...stylex.props(styles.fairness)}>
            <div><div {...stylex.props(styles.eyebrow)}>COMPETITIVE FIREWALL</div><h2 {...stylex.props(styles.arenaTitle)}>Money buys status, not power.</h2><p {...stylex.props(styles.description)}>Bunny Bucks, businesses, cosmetics, season wealth, CARROT, and Proof-of-Play authority remain outside the ranked combat ruleset.</p></div>
            <a href="/arena" {...stylex.props(styles.deploy)}>Enter equal-start arena →</a>
          </section>
        </Show>
      </section>
    </Shell>
  );
}

function money(value: number): string { return Math.floor(value).toLocaleString(); }

const styles = stylex.create({
  wrap: { maxWidth: "1180px", margin: "0 auto", padding: "60px 24px" },
  topline: { display: "flex", justifyContent: "space-between", gap: "24px", alignItems: "end", flexWrap: "wrap" },
  eyebrow: { color: "#d9f13b", fontSize: "11px", fontWeight: 900, letterSpacing: "0.16em" },
  title: { fontSize: "clamp(46px, 8vw, 84px)", margin: "10px 0", letterSpacing: "-0.055em", lineHeight: 0.95 },
  lede: { maxWidth: "760px", color: "#aeb4a4", lineHeight: 1.6, fontSize: "17px" },
  wallet: { display: "grid", gap: "3px", padding: "15px 18px", border: "1px solid #4b543d", borderRadius: "12px", backgroundColor: "#1b1f18", fontWeight: 950, fontSize: "24px", textAlign: "right" },
  notice: { marginTop: "24px", padding: "12px 14px", border: "1px solid #343b2e", borderRadius: "10px", backgroundColor: "#171a14", color: "#aeb4a4", fontSize: "13px" },
  loading: { marginTop: "30px", padding: "40px", border: "1px dashed #444c3b", borderRadius: "14px", color: "#899180" },
  briefing: { display: "grid", gridTemplateColumns: "60px 1fr auto", gap: "16px", alignItems: "center", marginTop: "20px", padding: "18px", border: "1px solid #68704f", borderRadius: "14px", backgroundColor: "#24291e" },
  npcIcon: { width: "56px", height: "56px", display: "grid", placeItems: "center", fontSize: "32px", backgroundColor: "#30382b", borderRadius: "14px" },
  acknowledge: { border: "1px solid #d9f13b", borderRadius: "10px", backgroundColor: "transparent", color: "#d9f13b", padding: "11px 14px", fontWeight: 900, cursor: "pointer", whiteSpace: "nowrap", ":disabled": { opacity: 0.45, cursor: "not-allowed" } },
  playerStrip: { display: "grid", gridTemplateColumns: "54px 1fr auto", gap: "14px", alignItems: "center", marginTop: "20px", padding: "16px", border: "1px solid #343b2e", borderRadius: "14px", backgroundColor: "#1a1e17" },
  avatar: { width: "52px", height: "52px", display: "grid", placeItems: "center", borderRadius: "50%", backgroundColor: "#30382b", fontSize: "30px" },
  callsign: { color: "#899180", marginLeft: "6px" },
  secondaryLink: { color: "#d9f13b", textDecoration: "none", fontWeight: 800, fontSize: "13px" },
  board: { display: "grid", gap: "14px", marginTop: "20px" },
  operation: { display: "grid", gridTemplateColumns: "70px 1fr auto", gap: "20px", alignItems: "center", backgroundColor: "#1a1e17", border: "1px solid #343a2c", borderRadius: "16px", padding: "22px" },
  icon: { fontSize: "38px", width: "64px", height: "64px", display: "grid", placeItems: "center", borderRadius: "14px", backgroundColor: "#292f24" },
  small: { color: "#818878", fontSize: "11px", letterSpacing: "0.12em", fontWeight: 900 },
  businessTitle: { margin: "4px 0", fontSize: "24px" },
  description: { color: "#aeb4a4", lineHeight: 1.55, margin: "6px 0" },
  meta: { display: "flex", gap: "12px", color: "#d9f13b", fontWeight: 800, fontSize: "12px", flexWrap: "wrap" },
  upgrade: { border: 0, borderRadius: "10px", backgroundColor: "#d9f13b", color: "#11130f", padding: "13px 16px", fontWeight: 900, cursor: "pointer", whiteSpace: "nowrap", ":disabled": { opacity: 0.35, cursor: "not-allowed" } },
  seasonGrid: { display: "grid", gridTemplateColumns: "minmax(0, 1fr) 300px", gap: "14px", marginTop: "20px", alignItems: "stretch" },
  season: { padding: "28px", border: "1px solid #5a633e", borderRadius: "18px", backgroundColor: "#20251b" },
  seasonTop: { display: "flex", justifyContent: "space-between", gap: "24px", flexWrap: "wrap" },
  seasonTitle: { fontSize: "36px", margin: "7px 0" },
  seasonNumbers: { display: "grid", textAlign: "right", alignContent: "center" },
  progressTrack: { marginTop: "18px", height: "12px", backgroundColor: "#11140f", borderRadius: "999px", overflow: "hidden" },
  progressFill: { height: "100%", backgroundColor: "#d9f13b", borderRadius: "999px" },
  seasonActions: { display: "flex", justifyContent: "space-between", gap: "14px", alignItems: "center", marginTop: "14px", color: "#899180", fontSize: "12px", flexWrap: "wrap" },
  turnIn: { border: "1px solid #d9f13b", borderRadius: "10px", backgroundColor: "transparent", color: "#d9f13b", padding: "11px 14px", fontWeight: 900, cursor: "pointer", ":disabled": { opacity: 0.35, cursor: "not-allowed" } },
  standings: { padding: "22px", border: "1px solid #343b2e", borderRadius: "18px", backgroundColor: "#171a14" },
  standingRow: { display: "grid", gridTemplateColumns: "34px 1fr auto", gap: "8px", alignItems: "center", marginTop: "18px" },
  muted: { color: "#838b7a", fontSize: "12px", lineHeight: 1.55 },
  fairness: { marginTop: "20px", padding: "28px", borderRadius: "18px", border: "1px solid #41493a", background: "linear-gradient(135deg, #24291e, #161914)", display: "flex", justifyContent: "space-between", gap: "24px", alignItems: "center", flexWrap: "wrap" },
  arenaTitle: { fontSize: "34px", margin: "7px 0" },
  deploy: { padding: "13px 16px", borderRadius: "10px", border: "1px solid #d9f13b", backgroundColor: "#d9f13b", color: "#11130f", fontWeight: 900, textDecoration: "none" },
});
