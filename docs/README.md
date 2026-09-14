# Battle Bunny Wealth docs

This directory is the source of truth for product, game, economy, and protocol decisions.

Start with [Product decisions](product-decisions.md) for the compact list of what is already settled versus what remains open.

- [Product decisions](product-decisions.md) — canonical player, network, CARROT, and ranked-fairness decisions.
- [Architecture](architecture.md) — executable system boundaries, storage separation, and deployment model.
- [Game design](game-design.md) — player-created bunny, NPC/lore roles, seasonal idle businesses, missions, and equal-start Warren Wars.
- [Idle economy](idle-economy.md) — player profile, businesses, offline earnings, synergies, and seasonal turn-in.
- [Seasons and story](seasons-story.md) — timed seasons, archives, NPC story beats, awards, Warren customization, telemetry, and development controls.
- [Account and portable identity](account-identity.md) — passkeys, account-owned game state, ATProto links, device enrollment/revocation, and authentication boundaries.
- [Warren Wars](warren-wars.md) — deterministic arena rules, bombs, pickups, ranked fairness, replay format, inputs, and bots.
- [Proof of Play](proof-of-play.md) — authority, optional network missions, committee-selection research, and consensus phases.
- [Proof-of-Play missions and authority](proof-of-play-authority.md) — signed missions, rate limits, bounded authority, decay, newcomer weighting, and committee lottery.
- [Attested participation](attested-participation.md) — provider adapters, key-bound attestation, mobile integration, testnet gate, and multi-device weighting.
- [Proof-of-Play simulator](proof-of-play-simulator.md) — deterministic attacker/population simulation, capture probabilities, liveness, churn, provider outages, and attack-cost assumptions.
- [Permissioned testnet](permissioned-testnet.md) — four-phone bootstrap, validator activation, node separation, peer defenses, and non-economic v0.9 testnet.
- [Testnet consensus](testnet-consensus.md) — deterministic state execution, historical validator sets, protocol activation, signed lock proofs, round change, finality settlement, persistence, and recovery.
- [Round-change protocol](round-change-protocol.md) — v0.14 proposer rotation, timeout certificates, value-bound votes, quorum-intersection lock proofs, and crash-safe round recovery.
- [Distributed testnet evidence](distributed-testnet-evidence.md) — checkpoint comparison, latency/availability collection, pinned network metadata, and the live-operator evidence gate.
- [v0.15 live distributed testnet](live-testnet-v0.15.md) — evidence artifact, operator inventory, manual collector, failure exercises, and completion criteria.
- [v0.15 review freeze](review-freeze-v0.15.md) — exact-SHA evidence manifest and independent-review handoff boundary.
- [v0.12 public testnet](public-testnet-v0.12.md) — consensus TEST-CARROT execution, peer diversity, transition commitments, and public funding policy.
- [v0.12 security-review package](security-review-v0.12.md) — external-review scope, invariants, adversarial checklist, commands, known limitations, and production release blockers.
- [v0.13 security-review rehearsal](security-review-v0.13-rehearsal.md) — internal pre-audit exercise, findings, known limitations, and external-review handoff notes.
- [v0.14 security-review handoff](security-review-v0.14-handoff.md) — round-change/lock-proof review scope, remediated findings, and remaining production blockers.
- [Testnet incident response](testnet-incident-response.md) — validator/node/wallet compromise, finality halts, upgrades, replay recovery, and evidence preservation.
- [Testnet operations](testnet-operations.md) — local node/cluster operation, soak tests, round-change chaos, snapshots, checkpoints, and evidence collection.
- [Identity and attestation](identity-and-attestation.md) — player/account/device separation, ATProto, passkeys, App Attest, and Play Integrity.
- [Economy](economy.md) — Bunny Bucks vs. CARROT and ranked-fairness boundaries.
- [CARROT protocol](carrot-protocol.md) — v0.10 fixed supply, allocations, vesting, issuance, fees, custody, policy hash, and supply reporting.
- [Security](security.md) — Sybil resistance, device farms, authority grinding, CARROT risks, anti-cheat, and threat model.
- [Development](development.md) — local workflow, repository layout, API conventions, and testing.
- [Releases](releases.md) — `feature/* -> dev -> main -> tag` lifecycle.
- [Roadmap](roadmap.md) — staged path through Proof-of-Play public testing and economic-readiness gates.
- [ADR 0001](adr/0001-proof-of-play-boundary.md) — Proof of Play is a separate protocol package.
- [ADR 0002](adr/0002-portable-identity-and-device-attestation.md) — portable account identity plus device attestation.
- [ADR 0003](adr/0003-validator-activation-and-vote-weight.md) — authority selects committees; selected members get one vote.
- [ADR 0004](adr/0004-carrot-fixed-supply.md) — CARROT uses genesis reserves and deterministic release.

## Current non-negotiable design rules

1. Players create and name their own bunny; named Battle Bunny characters are NPCs for story/lore/missions.
2. Seasonal businesses generate Bunny Bucks; Bunny Bucks are not CARROT.
3. Network missions are optional for ordinary game progression but can increase bounded Proof-of-Play authority.
4. Active honest participants can receive better committee-selection odds than inactive participants; activity never guarantees permanent control.
5. Ranked Warren Wars begins from an equal competitive state.
6. CARROT, Bunny Bucks, purchases, authority, account age, and story progress cannot buy ranked combat power.
7. CARROT is capped at 21,000,000 with 8 decimals and a fixed 20/60/10/5/5 allocation; the founder allocation has a one-year cliff/four-year vest.
8. CARROT missions do not directly pay tokens; committee finality is the modeled Proof-of-Play issuance event.
9. Proof of Play remains a research protocol until its Sybil resistance and finality survive public adversarial testing.
10. Account login, portable/social identity, device identity, platform attestation, and Proof-of-Play authority are separate security concepts; none automatically proves unique humanity.
11. v0.15 still executes **TEST-CARROT only**. It does not activate economically valuable CARROT, token sales, exchange integration, automatic treasury spending, or production consensus.
12. A finalized validator-set commitment changes voting power only at its committed activation height; historical blocks continue to verify against the validator set that was active when they finalized.
13. A finalized protocol-upgrade commitment cannot load arbitrary code. Nodes must explicitly support the scheduled version or fail closed.
14. Public snapshots/checkpoints are verification aids, not trusted substitutes for finality verification or genesis replay.
15. Committee membership is fixed for one height across consensus rounds; proposer rotation requires a quorum-signed round certificate.
16. A claimed value lock must carry the validator's signed value-bound vote. Certificate-wide carry-forward requires matching proofs at `max(1, 2Q-N)` rather than one validator's claim.
17. Validators remain individually bound by their proven locks and may not vote for a conflicting value at the same height.
18. Equivocation evidence is cryptographically verifiable but does not automatically slash, confiscate, ban, or remove a validator.
19. Local chaos testing and evidence tooling do not count as real multi-provider/geographic public-testnet evidence.
20. A v0.15 review freeze is tied to an exact executable SHA and evidence hashes; changing executable consensus behavior requires a new review target and, when behavior can change live-network results, a new evidence window.
