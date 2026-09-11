import * as stylex from "@stylexjs/stylex";
import { For, Show, createSignal, onSettled } from "solid-js";

import Shell from "../components/Shell";

type Status = {
  protocol: string;
  phase: string;
  epoch: number;
  chainHeight: number;
  committeeSize: number;
  quorum: string;
  epochSeconds: number;
  challengeTtlSeconds: number;
  maxHumanChallengesPerDay: number;
  attestationProviders: string[];
};

async function loadStatus(): Promise<Status> {
  const response = await fetch("/api/v1/proof-of-play");
  if (!response.ok) throw new Error(`status request failed: ${response.status}`);
  return response.json();
}

export default function ProofOfPlayRoute() {
  const [status, setStatus] = createSignal<Status>();
  const [error, setError] = createSignal<string>();

  onSettled(() => void (async () => {
    try {
      setStatus(await loadStatus());
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "unable to load protocol status");
    }
  })());

  return (
    <Shell>
      <section {...stylex.props(styles.wrap)}>
        <div {...stylex.props(styles.eyebrow)}>EXPERIMENTAL CONSENSUS LAYER</div>
        <h1 {...stylex.props(styles.title)}>Proof of Play</h1>
        <p {...stylex.props(styles.lede)}>
          The thesis: make long-lived, attested participation by ordinary devices useful to network security — then hide the cryptography behind a game people actually want to play.
        </p>

        <div {...stylex.props(styles.flow)}>
          <For each={["ATProto DID / passkey", "Attested device key", "Random epoch challenge", "Participation proof", "Committee selection", "Block finality"]}>
            {(item, index) => <div {...stylex.props(styles.step)}><span>{String(index() + 1).padStart(2, "0")}</span>{item}</div>}
          </For>
        </div>

        <Show when={error()}>{(message) => <div {...stylex.props(styles.error)}>Protocol status unavailable: {message()}</div>}</Show>
        <Show when={status()} fallback={<div {...stylex.props(styles.status)}>Connecting to local Proof of Play coordinator…</div>}>
          {(network) => (
            <section {...stylex.props(styles.network)}>
              <div {...stylex.props(styles.networkHeader)}>
                <div>
                  <div {...stylex.props(styles.eyebrow)}>LIVE SERVER SCAFFOLD</div>
                  <h2>Network status</h2>
                </div>
                <span {...stylex.props(styles.badge)}>{network().phase}</span>
              </div>
              <div {...stylex.props(styles.metrics)}>
                <Metric label="Protocol" value={network().protocol} />
                <Metric label="Epoch" value={String(network().epoch)} />
                <Metric label="Chain height" value={String(network().chainHeight)} />
                <Metric label="Committee" value={`${network().committeeSize} bunnies`} />
                <Metric label="Quorum" value={network().quorum} />
                <Metric label="Human checks" value={`${network().maxHumanChallengesPerDay}/day max`} />
              </div>
            </section>
          )}
        </Show>

        <div {...stylex.props(styles.note)}>
          <strong>Important:</strong> this is a protocol scaffold, not a production cryptocurrency. Apple App Attest and Google Play Integrity are modeled as pluggable attestation providers; no token issuance or real-money reward system exists yet.
        </div>
      </section>
    </Shell>
  );
}

function Metric(props: { label: string; value: string }) {
  return <div {...stylex.props(styles.metric)}><span>{props.label}</span><strong>{props.value}</strong></div>;
}

const styles = stylex.create({
  wrap: { maxWidth: "1180px", margin: "0 auto", padding: "70px 24px" },
  eyebrow: { color: "#d9f13b", fontSize: "12px", fontWeight: 900, letterSpacing: "0.16em" },
  title: { fontSize: "clamp(58px, 10vw, 112px)", lineHeight: 0.9, letterSpacing: "-0.065em", margin: "14px 0 22px" },
  lede: { color: "#aeb4a4", maxWidth: "850px", lineHeight: 1.55, fontSize: "22px" },
  flow: { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(170px, 1fr))", marginTop: "42px", gap: "8px" },
  step: { minHeight: "120px", padding: "18px", backgroundColor: "#1b1f18", border: "1px solid #343a2c", borderRadius: "12px", display: "flex", flexDirection: "column", justifyContent: "space-between", fontWeight: 800 },
  status: { marginTop: "24px", color: "#858c78" },
  error: { marginTop: "24px", color: "#ffb4a8", backgroundColor: "#2a1917", border: "1px solid #6f3933", borderRadius: "10px", padding: "14px" },
  network: { marginTop: "36px", padding: "28px", backgroundColor: "#151813", border: "1px solid #414936", borderRadius: "16px" },
  networkHeader: { display: "flex", justifyContent: "space-between", alignItems: "center", gap: "20px", flexWrap: "wrap" },
  badge: { color: "#d9f13b", backgroundColor: "#272d21", border: "1px solid #4a543d", borderRadius: "999px", padding: "8px 10px", fontSize: "12px", fontWeight: 900 },
  metrics: { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))", gap: "10px", marginTop: "20px" },
  metric: { backgroundColor: "#20251c", padding: "14px", borderRadius: "10px", display: "flex", flexDirection: "column", gap: "6px", color: "#929989", fontSize: "12px" },
  note: { marginTop: "22px", padding: "18px", borderLeft: "3px solid #d9f13b", backgroundColor: "#1a1d17", color: "#b8bdad", lineHeight: 1.6 },
});
