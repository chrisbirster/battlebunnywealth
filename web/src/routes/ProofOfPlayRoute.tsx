import * as stylex from "@stylexjs/stylex";
import { For, Show, createSignal, onSettled } from "solid-js";

import Shell from "../components/Shell";
import { authApi, type Account, type Device } from "../lib/authApi";
import { currentBrowserDevicePublicKey, signCurrentDevicePayload } from "../lib/deviceKey";
import { proofApi, type AuthoritySnapshot, type NetworkStatus } from "../lib/proofApi";

export default function ProofOfPlayRoute() {
  const [status, setStatus] = createSignal<NetworkStatus>();
  const [authority, setAuthority] = createSignal<AuthoritySnapshot>();
  const [account, setAccount] = createSignal<Account>();
  const [message, setMessage] = createSignal("Loading Proof of Play…");
  const [busy, setBusy] = createSignal(false);

  const load = async () => {
    try {
      setStatus(await proofApi.status());
      try {
        const [nextAccount, nextAuthority] = await Promise.all([authApi.me(), proofApi.me()]);
        setAccount(nextAccount);
        setAuthority(nextAuthority);
        setMessage("Network duty is optional. Your businesses and Warren Wars remain available whether you participate or not.");
      } catch (error) {
        const statusCode = (error as Error & { status?: number }).status;
        setMessage(statusCode === 401 ? "Sign in to volunteer for network duty and build prototype authority." : errorText(error));
      }
    } catch (error) {
      setMessage(errorText(error));
    }
  };

  onSettled(() => void load());

  const localDevice = async (): Promise<Device> => {
    const current = account();
    if (!current) throw new Error("Sign in before requesting network duty.");
    const publicKey = await currentBrowserDevicePublicKey();
    if (!publicKey) throw new Error("Enroll this browser from Account before requesting a mission.");
    const device = current.devices.find((candidate) => candidate.status === "active" && candidate.publicKeySpki === publicKey);
    if (!device) throw new Error("This browser key is not an active enrolled device. Re-enroll it from Account.");
    return device;
  };

  const requestMission = async () => {
    setBusy(true);
    try {
      const device = await localDevice();
      const next = await proofApi.issueMission(device.id);
      setAuthority(next);
      setMessage(next.currentMission ? `${next.currentMission.npc} assigned ${next.currentMission.title}. The challenge expires quickly.` : "No mission is currently available.");
    } catch (error) {
      setMessage(errorText(error));
    } finally {
      setBusy(false);
    }
  };

  const completeMission = async () => {
    const mission = authority()?.currentMission;
    if (!mission) return;
    setBusy(true);
    try {
      const device = await localDevice();
      if (device.id !== mission.deviceId) throw new Error("This mission belongs to a different enrolled device.");
      const signature = await signCurrentDevicePayload(mission.signedPayload);
      const next = await proofApi.completeMission(mission.id, device.id, signature);
      setAuthority(next);
      setMessage(`Mission verified. +${mission.authorityAward} prototype authority.`);
    } catch (error) {
      setMessage(errorText(error));
    } finally {
      setBusy(false);
    }
  };

  const authorityPercent = () => {
    const current = authority();
    return current ? Math.min(100, Math.round((current.authority.score / current.authority.maximum) * 100)) : 0;
  };

  return (
    <Shell>
      <section {...stylex.props(styles.wrap)}>
        <div {...stylex.props(styles.eyebrow)}>EXPERIMENTAL CONSENSUS LAYER / v0.6</div>
        <h1 {...stylex.props(styles.title)}>Proof of Play</h1>
        <p {...stylex.props(styles.lede)}>
          Volunteer for short network missions, sign unpredictable challenges with your enrolled device, and build bounded authority over time. Network duty is optional; idle progression and equal-start Warren Wars do not depend on it.
        </p>

        <div {...stylex.props(styles.flow)}>
          <For each={["Passkey account", "Enrolled device key", "Optional network mission", "Signed epoch challenge", "Bounded authority", "Committee lottery"]}>
            {(item, index) => <div {...stylex.props(styles.step)}><span>{String(index() + 1).padStart(2, "0")}</span>{item}</div>}
          </For>
        </div>

        <div {...stylex.props(styles.notice)}>{message()}</div>

        <Show when={status()}>
          {(network) => (
            <section {...stylex.props(styles.network)}>
              <div {...stylex.props(styles.networkHeader)}>
                <div><div {...stylex.props(styles.eyebrow)}>PROTOCOL STATUS</div><h2 {...stylex.props(styles.sectionTitle)}>Network scaffold</h2></div>
                <span {...stylex.props(styles.badge)}>{network().phase}</span>
              </div>
              <div {...stylex.props(styles.metrics)}>
                <Metric label="Protocol" value={network().protocol} />
                <Metric label="Epoch" value={String(network().epoch)} />
                <Metric label="Chain height" value={String(network().chainHeight)} />
                <Metric label="Target committee" value={`${network().committeeSize} bunnies`} />
                <Metric label="Quorum" value={network().quorum} />
                <Metric label="Mission ceiling" value={`${network().maxHumanChallengesPerDay}/day`} />
              </div>
            </section>
          )}
        </Show>

        <Show when={account()} fallback={
          <section {...stylex.props(styles.callout)}>
            <div><div {...stylex.props(styles.eyebrow)}>NETWORK SERVICE RECORD</div><h2 {...stylex.props(styles.sectionTitle)}>Sign in to volunteer.</h2><p {...stylex.props(styles.muted)}>A passkey account is required so missions can bind authority to a persistent participant rather than a disposable browser session.</p></div>
            <a href="/account" {...stylex.props(styles.primaryLink)}>Open Account →</a>
          </section>
        }>
          <Show when={authority()}>
            {(record) => (
              <>
                <section {...stylex.props(styles.authorityGrid)}>
                  <div {...stylex.props(styles.authorityCard)}>
                    <div {...stylex.props(styles.eyebrow)}>YOUR AUTHORITY</div>
                    <div {...stylex.props(styles.bigNumber)}>{record().authority.score}<small> / {record().authority.maximum}</small></div>
                    <div {...stylex.props(styles.progressTrack)}><div style={{ width: `${authorityPercent()}%` }} {...stylex.props(styles.progressFill)} /></div>
                    <div {...stylex.props(styles.authorityFacts)}>
                      <span>{record().authority.missionsCompleted} missions verified</span>
                      <span>{record().authority.newcomerWeightPercent}% newcomer weight</span>
                      <span>committee weight {record().authority.committeeWeight}</span>
                    </div>
                  </div>
                  <div {...stylex.props(styles.authorityCard)}>
                    <div {...stylex.props(styles.eyebrow)}>ELIGIBILITY</div>
                    <StatusRow label="Prototype lottery" value={record().authority.prototypeEligible ? "eligible" : `needs ${record().authority.eligibilityThreshold} authority`} />
                    <StatusRow label="Production validator" value="locked until attestation" />
                    <StatusRow label="Daily missions" value={`${record().dailyMissionsUsed} / ${record().dailyMissionLimit}`} />
                    <StatusRow label="Decay" value={`${record().authority.decayPerDayBasisPoints / 100}%/day after grace`} />
                  </div>
                </section>

                <section {...stylex.props(styles.mission)}>
                  <Show when={record().currentMission} fallback={
                    <div {...stylex.props(styles.missionEmpty)}>
                      <div><div {...stylex.props(styles.eyebrow)}>OPTIONAL NETWORK DUTY</div><h2 {...stylex.props(styles.sectionTitle)}>No active mission.</h2><p {...stylex.props(styles.muted)}>Requesting a mission does not affect businesses, story progression, or Warren Wars access.</p></div>
                      <button type="button" disabled={busy() || record().dailyMissionsUsed >= record().dailyMissionLimit} onClick={() => void requestMission()} {...stylex.props(styles.action)}>{busy() ? "Requesting…" : "Request mission"}</button>
                    </div>
                  }>
                    {(mission) => (
                      <div {...stylex.props(styles.missionActive)}>
                        <div {...stylex.props(styles.npc)}>📡</div>
                        <div>
                          <div {...stylex.props(styles.eyebrow)}>{mission().npc} / EPOCH {mission().epoch}</div>
                          <h2 {...stylex.props(styles.missionTitle)}>{mission().title}</h2>
                          <p {...stylex.props(styles.muted)}>{mission().briefing}</p>
                          <div {...stylex.props(styles.missionMeta)}><span>{mission().operation}</span><span>+{mission().authorityAward} authority</span><span>{mission().deviceAttestation}</span><span>expires {clock(mission().expiresAt)}</span></div>
                        </div>
                        <button type="button" disabled={busy()} onClick={() => void completeMission()} {...stylex.props(styles.action)}>{busy() ? "Signing…" : "Sign & complete"}</button>
                      </div>
                    )}
                  </Show>
                </section>

                <section {...stylex.props(styles.history)}>
                  <div><div {...stylex.props(styles.eyebrow)}>SERVICE HISTORY</div><h2 {...stylex.props(styles.sectionTitle)}>Recent verified missions</h2></div>
                  <Show when={record().recentCompletions.length > 0} fallback={<p {...stylex.props(styles.muted)}>No verified network missions yet.</p>}>
                    <div {...stylex.props(styles.historyRows)}>
                      <For each={record().recentCompletions}>{(completion) => <div {...stylex.props(styles.historyRow)}><span>{completion.kind}</span><span>epoch {completion.epoch}</span><strong>+{completion.authorityAward}</strong><time>{dateTime(completion.completedAt)}</time></div>}</For>
                    </div>
                  </Show>
                </section>
              </>
            )}
          </Show>
        </Show>

        <div {...stylex.props(styles.note)}>
          <strong>Security boundary:</strong> v0.6 proves control of an enrolled device key and implements authority mechanics, but current browser devices are still unattested. Authority can enter the prototype committee lottery, while production committee eligibility remains disabled until the v0.7 attestation work.
        </div>
      </section>
    </Shell>
  );
}

function Metric(props: { label: string; value: string }) { return <div {...stylex.props(styles.metric)}><span>{props.label}</span><strong>{props.value}</strong></div>; }
function StatusRow(props: { label: string; value: string }) { return <div {...stylex.props(styles.statusRow)}><span>{props.label}</span><strong>{props.value}</strong></div>; }
function errorText(error: unknown): string { return error instanceof Error ? error.message : "Proof of Play request failed."; }
function clock(value: string): string { return new Date(value).toLocaleTimeString([], { hour: "numeric", minute: "2-digit", second: "2-digit" }); }
function dateTime(value: string): string { return new Date(value).toLocaleString(); }

const styles = stylex.create({
  wrap: { maxWidth: "1180px", margin: "0 auto", padding: "70px 24px" },
  eyebrow: { color: "#d9f13b", fontSize: "12px", fontWeight: 900, letterSpacing: "0.16em" },
  title: { fontSize: "clamp(58px, 10vw, 112px)", lineHeight: 0.9, letterSpacing: "-0.065em", margin: "14px 0 22px" },
  lede: { color: "#aeb4a4", maxWidth: "900px", lineHeight: 1.55, fontSize: "21px" },
  flow: { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))", marginTop: "42px", gap: "8px" },
  step: { minHeight: "110px", padding: "18px", backgroundColor: "#1b1f18", border: "1px solid #343a2c", borderRadius: "12px", display: "flex", flexDirection: "column", justifyContent: "space-between", fontWeight: 800 },
  notice: { marginTop: "20px", padding: "13px 15px", border: "1px solid #343b2e", borderRadius: "10px", backgroundColor: "#171a14", color: "#aeb4a4", fontSize: "13px" },
  network: { marginTop: "24px", padding: "26px", backgroundColor: "#151813", border: "1px solid #414936", borderRadius: "16px" },
  networkHeader: { display: "flex", justifyContent: "space-between", alignItems: "center", gap: "20px", flexWrap: "wrap" },
  badge: { color: "#d9f13b", backgroundColor: "#272d21", border: "1px solid #4a543d", borderRadius: "999px", padding: "8px 10px", fontSize: "12px", fontWeight: 900 },
  metrics: { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(145px, 1fr))", gap: "10px", marginTop: "20px" },
  metric: { backgroundColor: "#20251c", padding: "14px", borderRadius: "10px", display: "flex", flexDirection: "column", gap: "6px", color: "#929989", fontSize: "12px" },
  sectionTitle: { fontSize: "30px", margin: "6px 0" },
  callout: { marginTop: "20px", display: "flex", justifyContent: "space-between", gap: "20px", alignItems: "center", flexWrap: "wrap", padding: "24px", border: "1px solid #4b543d", borderRadius: "16px", backgroundColor: "#1b1f18" },
  primaryLink: { color: "#11130f", backgroundColor: "#d9f13b", padding: "12px 16px", borderRadius: "10px", textDecoration: "none", fontWeight: 900 },
  muted: { color: "#aeb4a4", lineHeight: 1.55, margin: "6px 0" },
  authorityGrid: { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(300px, 1fr))", gap: "14px", marginTop: "20px" },
  authorityCard: { padding: "24px", border: "1px solid #343a2c", borderRadius: "16px", backgroundColor: "#1a1e17" },
  bigNumber: { marginTop: "10px", fontSize: "48px", fontWeight: 950, letterSpacing: "-0.04em" },
  progressTrack: { marginTop: "12px", height: "10px", backgroundColor: "#0f110e", borderRadius: "999px", overflow: "hidden" },
  progressFill: { height: "100%", backgroundColor: "#d9f13b", borderRadius: "999px" },
  authorityFacts: { display: "flex", gap: "12px", flexWrap: "wrap", marginTop: "12px", color: "#929989", fontSize: "12px" },
  statusRow: { display: "flex", justifyContent: "space-between", gap: "16px", padding: "12px 0", borderBottom: "1px solid #2d3328", color: "#929989", fontSize: "13px" },
  mission: { marginTop: "14px", padding: "24px", border: "1px solid #68704f", borderRadius: "16px", backgroundColor: "#22281d" },
  missionEmpty: { display: "flex", justifyContent: "space-between", gap: "20px", alignItems: "center", flexWrap: "wrap" },
  missionActive: { display: "grid", gridTemplateColumns: "64px minmax(0, 1fr) auto", gap: "18px", alignItems: "center" },
  npc: { width: "60px", height: "60px", display: "grid", placeItems: "center", borderRadius: "14px", backgroundColor: "#30382b", fontSize: "31px" },
  missionTitle: { margin: "5px 0", fontSize: "32px" },
  missionMeta: { display: "flex", gap: "9px", flexWrap: "wrap", marginTop: "10px", color: "#d9f13b", fontSize: "11px", fontWeight: 800 },
  action: { border: 0, borderRadius: "10px", backgroundColor: "#d9f13b", color: "#11130f", padding: "13px 16px", fontWeight: 900, cursor: "pointer", whiteSpace: "nowrap", ":disabled": { opacity: 0.4, cursor: "not-allowed" } },
  history: { marginTop: "14px", padding: "24px", border: "1px solid #343a2c", borderRadius: "16px", backgroundColor: "#151813" },
  historyRows: { display: "grid", gap: "7px", marginTop: "14px" },
  historyRow: { display: "grid", gridTemplateColumns: "1.4fr 0.7fr 0.4fr 1fr", gap: "12px", padding: "11px 12px", borderRadius: "9px", backgroundColor: "#1e231b", color: "#9da493", fontSize: "12px" },
  note: { marginTop: "22px", padding: "18px", borderLeft: "3px solid #d9f13b", backgroundColor: "#1a1d17", color: "#b8bdad", lineHeight: 1.6 },
});
