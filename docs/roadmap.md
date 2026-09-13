# Roadmap

## v0.1 — Foundation — implemented on `dev`

- Go binary embeds the Vite/Solid SPA.
- Proof-of-Play protocol scaffold, architecture/security/economy docs, and feature -> dev -> main/tag discipline.

## v0.2 — Warren Wars vertical slice — implemented on `dev`

- Deterministic grid simulation, original arena, bombs/blasts/crates/pickups, equal-start ranked rules, bots, replay, keyboard/touch/gamepad, and 1,000 seeded CI matches.

## v0.3 — Player profile + seasonal idle economy — implemented on `dev`

- Persistent player-created bunny profile, three businesses, offline earnings, Bunny Bucks, upgrades/synergies, NPC orientation, season standings, and cosmetic-only turn-in.

## v0.4 — Seasons + story progression — implemented on `dev`

- Timed season lifecycle, archives, NPC story beats, locations, Warren themes, persistent awards, season resets, telemetry, campaign UI, and no ranked-power rewards.

## v0.5 — Account and portable identity — implemented on `dev`

- Passkeys/WebAuthn, hashed sessions, per-account game state, optional `did:plc`, device enrollment/revocation, and browser P-256 device keys.

## v0.6 — Proof-of-Play missions and authority — implemented on `dev`

- Signed optional missions, daily limits, bounded/decaying authority, newcomer weighting, deterministic weighted committee sampling, and Proof-of-Play UI.

## v0.7 — Attested participation prototype — implemented on `dev`

- Key-bound attestation challenges, native Apple/Android spikes, Play Integrity policy, App Attest integration boundary, attested eligibility, and diminishing multi-device authority.

## v0.8 — Proof-of-Play simulator — implemented on `dev`

- Deterministic attacker/population simulator, bot/phone farms, delayed attackers, outages/churn, collusion, capture/liveness/concentration/cost metrics, attack sweeps, and CI gates.

## v0.9 — Permissioned testnet — implemented on `dev`

- Four-device genesis with 3-of-4 finality.
- Validator maturation and rate-limited activation.
- Full-node identity separated from participant/device voting identity.
- Deterministic committee/proposer selection, signed proposals/votes/finality certificates, persistence/restart/catch-up, permissioned signed peer relay, and five-engine local smoke harness.
- Four-phone vs. 100-real-phone bootstrap security gate.
- Hardened Google Play Integrity and complete Apple App Attest enrollment/assertion verification.
- No economically valuable CARROT.

## v0.10 — CARROT protocol specification — implemented on `dev`

- `carrot/1` executable fixed-supply policy.
- 21,000,000 maximum CARROT with 8 decimal places.
- Genesis-reserve accounting: 20% founder/admin, 60% Proof-of-Play issuance, 10% ecosystem, 5% community, 5% security/public goods.
- One-year founder cliff and four-year total linear vest.
- 32 declining four-year issuance eras that exhaust the participation reserve exactly using integer arithmetic.
- Finality-signer equal reward model; ordinary missions do not directly pay CARROT.
- Supply-neutral fee model; testnet minimum fee remains zero.
- 2-of-3 founder custody target, 3-of-5 treasury target, seven-day key-rotation delay; treasury spending disabled in v0.10.
- Deterministic supply/circulation reporting and conservation tests.
- Permissioned-testnet protocol v2 genesis commits the frozen CARROT policy hash.
- CI drift gate for policy hash, supply, allocation, vesting, issuance, fee, and conservation invariants.
- Still no economically valuable CARROT or public token transaction system.

## v0.11 — Adversarial public testnet

- open enrollment
- valueless/test CARROT represented in consensus transactions
- wallet/transaction signature format
- public Sybil/device-farm attempts
- authority grinding tests
- committee bribery/collusion experiments
- treasury/governance test implementation
- bug bounty
- protocol upgrade exercises
- economics and legal review

## v1.0 candidate — Game + network readiness

### Game

- compelling seasonal idle loop
- stable player-created bunny/profile system
- strong NPC story/lore
- fair equal-start Warren Wars
- matchmaking/rating/replay/anti-cheat baseline

### Network

- attested enrollment
- mission/authority model validated by simulation and public testing
- robust committee/finality behavior
- deterministic fixed-supply CARROT accounting
- founder/admin allocation and lock policy publicly documented
- wallet/treasury/governance path reviewed
- security/legal review appropriate for an economically valuable launch

Economic activation should happen only when those gates are satisfied, not merely because the code can represent CARROT.
