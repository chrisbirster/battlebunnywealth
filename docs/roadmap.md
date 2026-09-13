# Roadmap

## v0.1–v0.8 — Foundation through simulation — implemented on `dev`

Game shell, Warren Wars, seasonal economy/story, passkey identity, Proof-of-Play missions/authority, device attestation, and deterministic attack simulation are implemented.

## v0.9 — Permissioned testnet — implemented on `dev`

- four-device genesis with 3-of-4 bootstrap finality
- validator maturation and rate-limited activation
- node identity separated from participant/device voting identity
- deterministic committee/proposer selection, signed finality, persistence/restart/catch-up, signed peer relay, and local cluster smoke tests
- four-phone versus 100-real-phone security gate
- hardened Play Integrity and complete App Attest enrollment/assertion verification

## v0.10 — CARROT protocol specification — implemented on `dev`

- fixed 21,000,000 CARROT maximum with 8 decimals
- 20/60/10/5/5 genesis-reserve allocation
- one-year founder cliff / four-year vest
- deterministic declining Proof-of-Play release
- supply-neutral fees, custody targets, supply reporting, and CARROT policy-hash commitment in testnet genesis
- no economically valuable activation

## v0.11 — Adversarial public testnet — implemented on `dev`

- Ed25519 TEST-CARROT wallets and `tcarrot1` addresses
- signed network-bound transfers with exact nonce, expiry, transaction ID, and replay protection
- bounded public mempool plus deterministic `test-carrot-transfer` operation representation
- permissionless self-signed observer/full-node discovery with expiry and per-host/total anti-flood limits; node count creates zero votes
- public validator applications with key-possession proof and server-side attestation/authority eligibility
- existing 30-day maturation and rate-limited activation preserved
- dry-run signed governance for protocol-upgrade and treasury proposals; treasury execution disabled
- deterministic upgrade plans with seven-day minimum notice
- 100-attacker CI gate for node flooding, invalid transaction spam, and validator-candidate flooding
- TEST-CARROT remains explicitly valueless

v0.11 deliberately stops at adversarial transaction admission and deterministic block-operation representation. Arbitrary public wallet transfers are not yet executed inside replicated consensus state.

## v0.12 — Consensus transaction execution + public-network hardening

- execute TEST-CARROT transfers in deterministic replicated state
- mempool-to-block selection rules and consensus-visible balances/nonces
- tie deterministic Proof-of-Play release to finality signer identities in the state transition
- public peer-diversity and eclipse testing across independent hosts
- finalized validator-set transition commitments
- finalized protocol-upgrade commitments and recovery exercises
- public test-funding policy
- prepare external security-review package

## v1.0 candidate — Game + network readiness

### Game

- compelling seasonal idle loop
- stable player-created bunny/profile system
- strong NPC story/lore
- fair equal-start Warren Wars
- matchmaking/rating/replay/anti-cheat baseline

### Network

- attested enrollment and public adversarial evidence
- robust committee/finality behavior
- deterministic fixed-supply CARROT accounting
- reviewed wallet/treasury/governance path
- external security/legal/tax/privacy/app-store review appropriate for an economically valuable launch

Economic activation happens only after those gates are satisfied, not merely because the code can represent CARROT.
