import * as stylex from "@stylexjs/stylex";
import { For } from "solid-js";

import Shell from "../components/Shell";
import { squad } from "../lib/squad";

export default function SquadRoute() {
  return (
    <Shell>
      <section {...stylex.props(styles.wrap)}>
        <div {...stylex.props(styles.eyebrow)}>PERSONNEL FILES</div>
        <h1 {...stylex.props(styles.title)}>Meet the Battle Bunnies.</h1>
        <p {...stylex.props(styles.lede)}>Underfunded. Overarmed. Extremely motivated by money.</p>
        <div {...stylex.props(styles.grid)}>
          <For each={squad}>{(member) => (
            <article {...stylex.props(styles.card)}>
              <div {...stylex.props(styles.icon)}>{member.icon}</div>
              <div {...stylex.props(styles.rank)}>{member.rank}</div>
              <h2 {...stylex.props(styles.name)}>{member.callsign}</h2>
              <div {...stylex.props(styles.role)}>{member.role}</div>
              <blockquote {...stylex.props(styles.quote)}>“{member.quote}”</blockquote>
              <div {...stylex.props(styles.perk)}>{member.perk}</div>
            </article>
          )}</For>
        </div>
      </section>
    </Shell>
  );
}

const styles = stylex.create({
  wrap: { maxWidth: "1180px", margin: "0 auto", padding: "70px 24px" },
  eyebrow: { color: "#d9f13b", fontSize: "12px", fontWeight: 900, letterSpacing: "0.16em" },
  title: { fontSize: "clamp(48px, 8vw, 92px)", lineHeight: 0.94, letterSpacing: "-0.055em", margin: "14px 0" },
  lede: { color: "#9fa695", fontSize: "20px" },
  grid: { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(280px, 1fr))", gap: "16px", marginTop: "40px" },
  card: { backgroundColor: "#1a1e17", border: "1px solid #343a2c", borderRadius: "18px", padding: "26px" },
  icon: { fontSize: "40px" },
  rank: { color: "#d9f13b", marginTop: "16px", textTransform: "uppercase", letterSpacing: "0.14em", fontSize: "11px", fontWeight: 900 },
  name: { fontSize: "30px", margin: "6px 0" },
  role: { color: "#929a89" },
  quote: { margin: "24px 0", paddingLeft: "14px", borderLeft: "2px solid #d9f13b", color: "#c9cdbf", fontStyle: "italic" },
  perk: { display: "inline-block", backgroundColor: "#292f24", color: "#d9f13b", padding: "8px 10px", borderRadius: "8px", fontSize: "12px", fontWeight: 800 },
});
