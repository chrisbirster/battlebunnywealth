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

## v0.5 — Account and portable identity — implemented on `dev`

- WebAuthn/passkey account creation and usernameless passkey login.
- Short-lived passkey ceremonies with RP ID, origin, challenge, and signature validation.
- ES256/P-256 passkey support with signature-counter rollback detection.
- Random opaque sessions stored as hashes and delivered through HttpOnly SameSite=Strict cookies.
- Per-account persistent Warren/game state with automatic claiming of the legacy single-player save.
- Multi-account local standings across account-owned game states.
- Optional `did:plc` ATProto profile resolution stored explicitly as `resolved-unverified`.
- Separation of player profile, account identity, authentication identity, portable/social identity, and device identity.
- Multi-device enrollment and revocation data model.
- Browser-generated P-256 device key persisted locally in IndexedDB; server stores the public key only.
- All v0.5 browser device keys remain `unattested`.
- `/account` UI for registration, sign-in, additional passkeys, ATProto linking, device enrollment, and revocation.

The v0.5 WebAuthn verifier is deliberately narrow alpha code and requires security/interoperability review before public production authentication.

## v0.6 — Proof-of-Play missions and authority — implemented on `dev`

- Optional NPC-presented network missions with Recon Patrol, Secure the Supply Line, and Verify Intel templates.
- Mission assignment bound to account, active enrolled device key, current epoch, current chain head, protocol operation, and cryptographically random challenge.
- Five-minute mission expiry plus a hard four-issued-missions-per-UTC-day ceiling.
- Device-key signature verification and replay/binding/expiry defenses.
- Persistent authority/service history independent of game economy and account storage.
- +25 authority per valid mission, capped at 1,000 for the current research policy.
- Prototype committee eligibility threshold at 100 authority.
- 72-hour inactivity grace followed by slow 0.5%-per-day authority decay.
- Fourteen-day newcomer ramp from 10% to 100% effective committee weight.
- Deterministic weighted committee sampling without replacement for research/testing.
- `/proof-of-play` mission UI with local device signing, authority status, committee weight, daily limits, and completion history.
- No mission requirement for idle progression, campaign progression, or Warren Wars access.
- Unattested devices can exercise the v0.6 prototype, but `productionEligible` and production committee selection remain hard-disabled until v0.7.

The v0.6 constants and selection algorithm are explicit research parameters for attack simulation, not frozen consensus rules.

## v0.7 — Attested participation prototype — implemented on `dev`

- Provider-agnostic, persistent, single-use attestation challenge service.
- Attestation challenge bound to the enrolled P-256 device key through a SHA-256 key commitment and canonical request binding.
- Native iOS App Attest + Secure Enclave mission-key source spike.
- Native Android Play Integrity + Android Keystore/StrongBox mission-key source spike.
- Apple App Attest verifier boundary that fails closed until a concrete server validator is configured.
- Google Play Integrity REST decoder and policy checks for request hash, package, app recognition, signing certificate, and device integrity.
- Provider evidence projected into account device records without merging account identity and device identity.
- First permissioned-testnet eligibility rule: authority threshold plus at least one active verified hardware-backed device signal.
- Deterministic attested committee-selection path for future testnet work while distributed production consensus remains disabled.
- Multi-device authority awards reduced to 100% / 25% / 10% / 2% for first / second / third / later active devices.
- Concurrent replay defense: the attestation challenge is persisted as consumed before external provider verification.
- Development provider available only under dev controls and permanently excluded from attested eligibility.
- Account UI displays provider, attestation status, hardware-backed signal, testnet eligibility, and device weight.

The native iOS/Android files are integration spikes rather than CI-built mobile applications. App Attest still needs a concrete reviewed server validator, and production Google integration needs renewable OAuth credentials rather than the development static-token source.

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
