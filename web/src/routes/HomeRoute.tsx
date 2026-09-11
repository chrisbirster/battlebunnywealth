import { A } from "@solidjs/router";
import * as stylex from "@stylexjs/stylex";
import { For } from "solid-js";

import Shell from "../components/Shell";
import { squad } from "../lib/squad";

export default function HomeRoute() {
  return (
    <Shell>
      <section {...stylex.props(styles.hero)}>
        <div {...stylex.props(styles.eyebrow)}>THE WARREN NEEDS YOU</div>
        <h1 {...stylex.props(styles.title)}>Build wealth.<br />Build bombs.<br /><span {...stylex.props(styles.accent)}>Win the warren.</span></h1>
        <p {...stylex.props(styles.lede)}>
          An idle military-bunny hustle game with skill-based bomb battles — powered underneath by an experimental Proof of Play network.
        </p>
        <div {...stylex.props(styles.actions)}>
          <A href="/game" {...stylex.props(styles.primary)}>Enter Warren Command</A>
          <A href="/proof-of-play" {...stylex.props(styles.secondary)}>How Proof of Play works</A>
        </div>
      </section>

      <section {...stylex.props(styles.grid)}>
        <article {...stylex.props(styles.panel)}>
          <div {...stylex.props(styles.kicker)}>01 / HUSTLE</div>
          <h2>Run ridiculous operations.</h2>
          <p>Scrap yards, carrot logistics, questionable supply contracts, and increasingly absurd businesses keep earning while you're away.</p>
        </article>
        <article {...stylex.props(styles.panel)}>
          <div {...stylex.props(styles.kicker)}>02 / ARM</div>
          <h2>Turn profit into firepower.</h2>
          <p>Research bomb blueprints, squad perks, gadgets, and base upgrades. The economy feeds the battle game, not the other way around.</p>
        </article>
        <article {...stylex.props(styles.panel)}>
          <div {...stylex.props(styles.kicker)}>03 / BATTLE</div>
          <h2>Settle it in the arena.</h2>
          <p>Fast grid battles reward positioning, timing, bluffing, and mastery. No pay-to-win stat wall decides who can dodge a blast.</p>
        </article>
      </section>

      <section {...stylex.props(styles.squadSection)}>
        <div>
          <div {...stylex.props(styles.eyebrow)}>MEET THE UNIT</div>
          <h2 {...stylex.props(styles.sectionTitle)}>A highly questionable military organization.</h2>
        </div>
        <div {...stylex.props(styles.squadGrid)}>
          <For each={squad.slice(0, 3)}>{(member) => (
            <article {...stylex.props(styles.member)}>
              <div {...stylex.props(styles.avatar)}>{member.icon}</div>
              <div {...stylex.props(styles.rank)}>{member.rank}</div>
              <h3>{member.callsign}</h3>
              <p>{member.role}</p>
            </article>
          )}</For>
        </div>
      </section>
    </Shell>
  );
}

const styles = stylex.create({
  hero: { maxWidth: "1180px", margin: "0 auto", padding: "86px 24px 72px" },
  eyebrow: { color: "#d9f13b", fontSize: "12px", fontWeight: 900, letterSpacing: "0.18em" },
  title: { margin: "16px 0 22px", maxWidth: "850px", fontSize: "clamp(54px, 9vw, 116px)", lineHeight: 0.86, letterSpacing: "-0.065em", textTransform: "uppercase" },
  accent: { color: "#d9f13b" },
  lede: { maxWidth: "700px", color: "#aeb4a4", fontSize: "clamp(18px, 2.2vw, 24px)", lineHeight: 1.5 },
  actions: { display: "flex", flexWrap: "wrap", gap: "12px", marginTop: "34px" },
  primary: { backgroundColor: "#d9f13b", color: "#11130f", padding: "14px 18px", borderRadius: "10px", textDecoration: "none", fontWeight: 900 },
  secondary: { border: "1px solid #454c3c", color: "#f3f0db", padding: "14px 18px", borderRadius: "10px", textDecoration: "none", fontWeight: 800 },
  grid: { maxWidth: "1180px", margin: "0 auto", padding: "0 24px", display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))", gap: "14px" },
  panel: { backgroundColor: "#1c2019", border: "1px solid #343a2c", borderRadius: "16px", padding: "28px", minHeight: "250px" },
  kicker: { color: "#858c78", fontWeight: 800, fontSize: "12px", letterSpacing: "0.12em" },
  squadSection: { maxWidth: "1180px", margin: "80px auto 0", padding: "0 24px" },
  sectionTitle: { marginTop: "10px", fontSize: "clamp(34px, 5vw, 62px)", letterSpacing: "-0.045em", maxWidth: "760px" },
  squadGrid: { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))", gap: "14px", marginTop: "26px" },
  member: { backgroundColor: "#141713", border: "1px solid #343a2c", borderRadius: "16px", padding: "22px" },
  avatar: { width: "58px", height: "58px", display: "grid", placeItems: "center", borderRadius: "50%", backgroundColor: "#2c3226", fontSize: "30px" },
  rank: { color: "#d9f13b", fontSize: "11px", fontWeight: 900, letterSpacing: "0.14em", textTransform: "uppercase", marginTop: "18px" },
});
