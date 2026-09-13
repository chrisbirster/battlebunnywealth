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

## v0.12 — Consensus transaction execution + public-network hardening — implemented on `dev`

- protocol-v3 deterministic replicated state-machine boundary
- TEST-CARROT transfers execute inside consensus with exact balances/nonces and fixed-supply conservation
- deterministic mempool-to-block selection and removal on finality
- canonical next-block settlement of the previous finality vote set so reward recipients cannot diverge across honest nodes
- deterministic Proof-of-Play release and fee distribution to the finalized settlement signers
- restart/store replay and peer catch-up reproduce the same consensus state root
- validator reward addresses are covered by the validator-signed admission payload
- public peer diversity adds IPv4 `/24` and IPv6 `/48` prefix limits in addition to per-host/total limits
- quorum-finalized validator-set and protocol-upgrade commitments with minimum notice and recovery replay
- hashed public TEST-CARROT funding policy with no mint, automatic faucet, or treasury-spending path
- external security-review handoff package and explicit known-limitations/release-blocker list
- TEST-CARROT remains explicitly valueless; no production CARROT activation

## v0.13 — Controlled transition activation + long-running public testnet — implemented on `dev`

- finalized validator-set commitments activate deterministically at the committed height
- historical validator sets remain available for committee, finality-certificate, and reward-settlement verification
- removed validators lose current voting power without invalidating historical signatures
- finalized protocol-upgrade commitments activate at the exact committed height when the binary explicitly supports the target version; unsupported versions fail closed
- protocol-v3 genesis can exercise the reviewed v4 compatibility activation path without arbitrary chain-loaded code
- transition schedules/history are committed into replicated state and survive restart/catch-up replay
- deterministic 1,100-block multi-node soak crosses validator/protocol activation, alternate quorum subsets, partitions/catch-up, and repeated full replay/restart while checking state-root convergence and fixed-supply conservation
- public discovery can apply trusted server-side ASN/provider diversity caps in addition to host and IP-prefix limits
- mempool supports up to 16 queued sequential/future nonces per sender, contiguous proposal selection, and bounded same-nonce fee replacement
- public consensus snapshots and compact checkpoints are exposed; checkpoints require independent genesis/finality replay to verify
- incident-response and key-compromise runbook added
- v0.13 internal security-review rehearsal records unresolved risks without claiming an external audit
- TEST-CARROT remains non-economic

## v0.14 — Round-change protocol + distributed-evidence tooling — implemented on feature branch / merge-gated

- fixed committee membership across rounds at a height, with proposer rotation by round
- validator-signed round-change messages and quorum round certificates
- persistent value locks prevent a validator from voting for conflicting state transitions later in the same height
- later-round finalized blocks embed their round certificate for independent replay verification
- in-progress round/lock/certificate state survives crashes; stale round snapshots lose to durably finalized history
- peer catch-up now durably persists independently verified imported blocks
- cryptographically verifiable proposer/vote equivocation evidence without automatic slashing or confiscation
- deterministic round-chaos harness covers partial locks, reordered/duplicate messages, temporary partitions, proposer timestamp skew, rolling replay, convergence, and fixed-supply conservation
- canonical hash-pinned CIDR-to-ASN/provider metadata maps support reproducible peer-diversity policy
- checkpoint comparison tooling detects same-height finalized-hash/state-root disagreement across operators
- distributed evidence collector records availability, latency, public status, and checkpoints from real endpoints when supplied
- v0.14 security-review handoff records remediated findings and remaining blockers
- TEST-CARROT remains non-economic

The repository does **not** claim a real geographically/network-provider-distributed evidence run in v0.14. Tooling and local adversarial evidence are implemented; real independent infrastructure is an operational/external gate.

## v0.15 — Live distributed testnet + independent review freeze

- run independent public nodes across multiple operators/providers/regions where practical
- retain repository SHA, genesis hash, CARROT policy hash, network-map hash/provenance, node/operator IDs, availability, latency, catch-up events, round changes, and checkpoint samples
- compare same-height checkpoints continuously and treat finalized-hash/state-root disagreement as a consensus incident
- exercise node loss, provider loss, rolling restart, peer churn, delayed recovery, and long partial partitions on real Internet paths
- measure ASN/provider diversity false positives, stale-data behavior, and inexpensive evasion strategies
- tune timeout/backoff policy from measured latency rather than local assumptions
- freeze one merged `dev` commit plus live-evidence artifacts for independent consensus/security review
- remediate independent-review findings before any production-security claim
- separately complete appropriate legal/tax/privacy/app-store review for economically valuable CARROT
- keep TEST-CARROT valueless throughout the live evidence/review window

## v1.0 candidate — Game + network readiness

### Game

- compelling seasonal idle loop
- stable player-created bunny/profile system
- strong NPC story/lore
- fair equal-start Warren Wars
- matchmaking/rating/replay/anti-cheat baseline

### Network

- attested enrollment and public adversarial evidence
- robust committee/finality behavior through validator-set changes, protocol upgrades, and round changes
- deterministic fixed-supply CARROT accounting
- reviewed wallet/treasury/governance path
- completed live distributed-testnet evidence window
- independent external security review plus legal/tax/privacy/app-store review appropriate for an economically valuable launch

Economic activation happens only after those gates are satisfied, not merely because the code can represent CARROT.
