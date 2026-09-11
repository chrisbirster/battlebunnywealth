# Roadmap

## v0.1 — Foundation

- Go binary embeds Vite/Solid SPA.
- SolidJS 2 + Solid Router + StyleX shell.
- Game landing page, Warren Command prototype, NPC/lore page.
- Proof-of-Play package: blocks, chain validation, participants, challenges, proofs, config, status API.
- Architecture/security/economy documentation.
- feature -> dev -> main/tag CI discipline.

## v0.2 — Playable Warren Wars vertical slice — implemented on `dev`

- Deterministic grid simulation independent of rendering.
- One original arena map.
- Movement, collision, bomb placement/fuse/blast.
- Destructible blocks and standardized pickups.
- Player-created bunny representation in the arena.
- Equal-start ranked rules: no account/season/CARROT/authority combat advantages.
- Local bots for repeatable testing.
- Match seed/replay foundation for debugging and fairness checks.
- Keyboard, touch, and standard-gamepad input.
- 1,000 seeded deterministic headless-match CI gate.

## v0.3 — Player profile + seasonal idle economy — implemented on `dev`

- Persistent player-created Battle Bunny profile.
- Bunny name, callsign, fur, ears, uniform, and cosmetic locker.
- Single-player alpha state persisted by the Go server with schema-versioned atomic storage.
- First three businesses and automatic offline earnings with an eight-hour cap.
- Seasonal Bunny Bucks economy.
- Business upgrades and first synergies.
- Persistent NPC orientation sequence across the named story cast.
- Preliminary local season standings API.
- Season Zero manual turn-in for cosmetic/service-record rewards.
- Server-authoritative validation for profile/economy mutations.
- No persistent ranked combat power from idle wealth.

## v0.4 — Seasons + story progression — implemented on `dev`

- Server-timed `active -> turn-in -> archive -> next season` lifecycle.
- Seven-day alpha season window and 24-hour turn-in window.
- Schema-v2 migration for existing v0.3/schema-v1 saves.
- Archived season history with final earnings, chapter/location, trophies, medals, turn-ins, and completion timestamps.
- Chapter/location progression through Broken Burrow, Scrap Row, and Carrot District.
- Persistent NPC story beats with First Sergeant Hard-as-Nails, Private Stuffy, Captain Cashmere, Corporal Boomboom, Da Champ, and Doc Flopsy.
- Warren themes plus persistent season pennants/trophy decorations.
- Service-record awards that survive seasonal economy resets.
- Competitive season resets that return Bunny Bucks/businesses to the baseline while preserving identity/history/status.
- Economy telemetry for lifetime earnings, upgrades, offline returns, and completed seasons.
- Accelerated season transitions gated behind `BBWEALTH_DEV_CONTROLS=1` for local testing.
- `/campaign` UI plus story/Warren APIs.
- No story/season/Warren reward affects ranked Warren Wars combat power.

## v0.5 — Account and portable identity

- Passkey authentication.
- Optional ATProto DID/profile binding.
- Separation of player profile, account identity, authentication identity, and device identity.
- Device enrollment/revocation data model.
- Multi-device account policy.

## v0.6 — Proof-of-Play missions and authority

- Optional NPC-presented network missions.
- Deterministic mission qualification.
- Authority accrual rules.
- Bounded authority weighting.
- Slow inactivity decay.
- No mission requirement for ordinary idle progression or Warren Wars access.
- Active honest participation increases committee-selection probability rather than guaranteeing selection.

## v0.7 — Attested participation prototype

- Native iOS App Attest spike.
- Native Android Play Integrity spike.
- Hardware-backed device signing.
- Challenge issuance/replay defense.
- Provider adapter abstraction.
- Multi-device diminishing-weight experiment.

## v0.8 — Proof-of-Play simulator

- honest-user model
- inactive-player model
- highly active participant model
- mission-completion distributions
- bot/mission-farm model
- real-device farm model
- multi-device household model
- authority/decay simulations
- committee-capture probability analysis
- attestation-provider outage model

## v0.9 — Permissioned testnet

- peer networking
- deterministic state machine
- persistent chain/state
- authority state replication
- committee selection
- votes/finality
- metrics and operator docs
- no economically valuable CARROT yet

## v0.10 — CARROT protocol specification

Before issuance becomes valuable, freeze and review:

- fixed maximum CARROT supply
- exact divisibility
- declining issuance schedule
- Proof-of-Play reward allocation
- explicit 20% founder/admin allocation
- remaining 80% allocation categories
- founder vesting/lock policy
- fee model
- genesis state
- treasury/key management
- supply/circulation reporting

This milestone is specification and testnet implementation, not necessarily public economic activation.

## v0.11 — Adversarial public testnet

- open enrollment
- valueless/test CARROT representation
- public Sybil/device-farm attempts
- authority grinding tests
- committee bribery/collusion experiments
- bug bounty
- protocol upgrade exercises
- economics and legal review

## v1.0 candidate — Game + network readiness

A v1.0 candidate should require evidence that both halves are independently good:

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
- security/legal review appropriate for an economically valuable launch

Economic activation should happen only when those gates are satisfied, not merely because the code can mint a token.
