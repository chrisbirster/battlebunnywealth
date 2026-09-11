# Roadmap

## v0.1 — Foundation

- Go binary embeds Vite/Solid SPA.
- SolidJS 2 + Solid Router + StyleX shell.
- Game landing page, squad page, Warren Command prototype.
- Proof of Play package: blocks, chain validation, participants, challenges, proofs, config, status API.
- Architecture/security/economy documentation.
- feature -> dev -> main/tag CI discipline.

## v0.2 — Playable arena vertical slice

- Deterministic grid simulation independent of rendering.
- One arena map.
- Movement, collision, bomb placement/fuse/blast.
- Destructible blocks and pickups.
- First Sergeant + Stuffy playable.
- Local bots for repeatable testing.

## v0.3 — Idle economy vertical slice

- Persistent player profile.
- Operations and offline earnings.
- Upgrades/research.
- Squad progression.
- Idle rewards feed arena loadout without pay-to-win damage scaling.

## v0.4 — Account/identity

- Passkey authentication.
- Optional ATProto DID/profile binding.
- Device enrollment model and revocation.

## v0.5 — Attested participation prototype

- Native iOS App Attest spike.
- Native Android Play Integrity spike.
- Hardware-backed device signing.
- Challenge issuance/replay defense.
- No transferable token.

## v0.6 — Proof of Play simulator

- Honest-user model.
- bot/device-farm model.
- multi-device household model.
- reputation/decay simulations.
- committee capture probability analysis.

## v0.7 — Permissioned testnet

- peer networking
- deterministic state machine
- persistent chain/state
- committee selection
- votes/finality
- metrics and operator docs

## v0.8+ — Public experiments

Open, valueless public testnet first. Token/economic decisions come only after adversarial data.
