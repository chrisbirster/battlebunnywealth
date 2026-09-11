import * as stylex from "@stylexjs/stylex";
import { For, Show, createSignal } from "solid-js";

import Shell from "../components/Shell";
import { gameApi, type PlayerProfile } from "../lib/gameApi";

const furOptions = ["cinnamon", "snow", "charcoal", "spotted"];
const earOptions = ["upright", "lop", "battle-worn"];
const uniformOptions = ["field-olive", "desert-tan", "night-black", "medic-white"];

export default function ProfileRoute() {
  const [profile, setProfile] = createSignal<PlayerProfile>();
  const [message, setMessage] = createSignal("Loading service record…");
  const [saving, setSaving] = createSignal(false);

  const load = async () => {
    try {
      const snapshot = await gameApi.state();
      setProfile(snapshot.player);
      setMessage("Your bunny belongs to you. NPCs belong to the story.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Could not load profile.");
    }
  };
  void load();

  const patch = (value: Partial<PlayerProfile>) => setProfile((current) => current ? { ...current, ...value } : current);
  const save = async () => {
    const current = profile();
    if (!current) return;
    setSaving(true);
    try {
      const snapshot = await gameApi.updateProfile(current);
      setProfile(snapshot.player);
      setMessage("Service record updated.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Profile update failed.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Shell>
      <section {...stylex.props(styles.wrap)}>
        <div {...stylex.props(styles.eyebrow)}>PLAYER IDENTITY / v0.3</div>
        <h1 {...stylex.props(styles.title)}>Build your bunny.</h1>
        <p {...stylex.props(styles.lede)}>First Sergeant Hard-as-Nails, Private Stuffy, and the rest of the named cast are NPCs. This is the bunny other players will know as you.</p>
        <div {...stylex.props(styles.notice)}>{message()}</div>

        <Show when={profile()} fallback={<div {...stylex.props(styles.loading)}>Opening personnel file…</div>}>
          <div {...stylex.props(styles.layout)}>
            <div {...stylex.props(styles.preview)}>
              <div {...stylex.props(styles.bunny)}>🐰</div>
              <h2>{profile()!.name}</h2>
              <strong>“{profile()!.callsign || "No callsign"}”</strong>
              <dl {...stylex.props(styles.details)}>
                <div><dt>Fur</dt><dd>{profile()!.fur}</dd></div>
                <div><dt>Ears</dt><dd>{profile()!.ears}</dd></div>
                <div><dt>Uniform</dt><dd>{profile()!.uniform}</dd></div>
              </dl>
            </div>

            <form onSubmit={(event) => { event.preventDefault(); void save(); }} {...stylex.props(styles.form)}>
              <label {...stylex.props(styles.field)}><span>Name</span><input value={profile()!.name} maxLength={24} onInput={(event) => patch({ name: event.currentTarget.value })} /></label>
              <label {...stylex.props(styles.field)}><span>Callsign</span><input value={profile()!.callsign} maxLength={20} onInput={(event) => patch({ callsign: event.currentTarget.value })} /></label>
              <label {...stylex.props(styles.field)}><span>Fur</span><select value={profile()!.fur} onInput={(event) => patch({ fur: event.currentTarget.value })}><For each={furOptions}>{(option) => <option value={option}>{option}</option>}</For></select></label>
              <label {...stylex.props(styles.field)}><span>Ears</span><select value={profile()!.ears} onInput={(event) => patch({ ears: event.currentTarget.value })}><For each={earOptions}>{(option) => <option value={option}>{option}</option>}</For></select></label>
              <label {...stylex.props(styles.field)}><span>Uniform</span><select value={profile()!.uniform} onInput={(event) => patch({ uniform: event.currentTarget.value })}><For each={uniformOptions}>{(option) => <option value={option}>{option}</option>}</For></select></label>
              <button type="submit" disabled={saving()} {...stylex.props(styles.save)}>{saving() ? "Saving…" : "Save Battle Bunny"}</button>
            </form>
          </div>

          <section {...stylex.props(styles.cosmetics)}>
            <div {...stylex.props(styles.eyebrow)}>SERVICE LOCKER</div>
            <h2>Cosmetics & status</h2>
            <div {...stylex.props(styles.badges)}><For each={profile()!.cosmetics}>{(item) => <span {...stylex.props(styles.badge)}>{item}</span>}</For></div>
            <p {...stylex.props(styles.muted)}>Cosmetics, outfits, trophies, and season rewards never change ranked Warren Wars statistics.</p>
          </section>
        </Show>
      </section>
    </Shell>
  );
}

const styles = stylex.create({
  wrap: { maxWidth: "1000px", margin: "0 auto", padding: "60px 24px" },
  eyebrow: { color: "#d9f13b", fontSize: "11px", fontWeight: 900, letterSpacing: "0.16em" },
  title: { fontSize: "clamp(46px, 8vw, 78px)", margin: "10px 0", letterSpacing: "-0.055em", lineHeight: 0.95 },
  lede: { maxWidth: "760px", color: "#aeb4a4", lineHeight: 1.6, fontSize: "17px" },
  notice: { marginTop: "20px", padding: "12px 14px", border: "1px solid #343b2e", borderRadius: "10px", backgroundColor: "#171a14", color: "#aeb4a4", fontSize: "13px" },
  loading: { marginTop: "24px", padding: "36px", color: "#899180" },
  layout: { display: "grid", gridTemplateColumns: "minmax(260px, .8fr) minmax(320px, 1.2fr)", gap: "18px", marginTop: "22px" },
  preview: { padding: "28px", border: "1px solid #41493a", borderRadius: "18px", backgroundColor: "#1a1e17", textAlign: "center" },
  bunny: { fontSize: "96px", marginBottom: "10px" },
  details: { display: "grid", gap: "8px", marginTop: "24px", textAlign: "left" },
  form: { padding: "24px", border: "1px solid #343b2e", borderRadius: "18px", backgroundColor: "#171a14", display: "grid", gap: "14px" },
  field: { display: "grid", gap: "7px", color: "#aeb4a4", fontSize: "12px", fontWeight: 800 },
  save: { marginTop: "8px", border: 0, borderRadius: "10px", backgroundColor: "#d9f13b", color: "#11130f", padding: "13px 16px", fontWeight: 900, cursor: "pointer", ":disabled": { opacity: 0.5, cursor: "not-allowed" } },
  cosmetics: { marginTop: "18px", padding: "24px", border: "1px solid #343b2e", borderRadius: "16px", backgroundColor: "#1a1e17" },
  badges: { display: "flex", gap: "8px", flexWrap: "wrap", margin: "14px 0" },
  badge: { border: "1px solid #5d6747", borderRadius: "999px", padding: "7px 10px", color: "#d9f13b", fontSize: "12px", fontWeight: 800 },
  muted: { color: "#838b7a", lineHeight: 1.55, fontSize: "12px" },
});
