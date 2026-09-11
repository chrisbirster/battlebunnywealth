# Roadmap

## v0.1 — Foundation

- Go binary embeds Vite/Solid SPA.
- SolidJS 2 + Solid Router + StyleX shell.
- Game landing page, Warren Command prototype, NPC/lore page.
- Proof-of-Play package: blocks, chain validation, participants, challenges, proofs, config, status API.
- Architecture/security/economy documentation.
- feature -> dev -> main/tag CI discipline.

## v0.2 — Playable Warren Wars vertical slice

- Deterministic grid simulation independent of rendering.
- One original arena map.
- Movement, collision, bomb placement/fuse/blast.
- Destructible blocks and standardized pickups.
- Player-created bunny representation in the arena.
- Equal-start ranked rules: no account/season/CARROT/authority combat advantages.
- Local bots for repeatable testing.
- Match seed/replay foundation for debugging and fairness checks.

## v0.3 — Player profile + seasonal idle economy

- Persistent player-created Battle Bunny profile.
- Bunny name/callsign/cosmetics.
- NPC-driven onboarding with First Sergeant Hard-as-Nails, Private Stuffy, Captain Cashmere, Corporal Boomboom, Da Champ, and Doc Flopsy.
- First businesses and offline earnings.
- Seasonal Bunny Bucks economy.
- Business upgrades/synergies.
- Season standings.
- Season turn-in for cosmetic/collection/progression rewards.
- No persistent ranked combat power from idle wealth.

## v0.4 — Seasons + story progression

- Formal season lifecycle.
- Chapter/location progression.
- Seasonal NPC story beats and missions.
- Medals/trophies/profile history.
- Warren customization.
- Competitive season reset/archival rules.
- Economy telemetry to tune idle progression before connecting valuable protocol rewards.

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
