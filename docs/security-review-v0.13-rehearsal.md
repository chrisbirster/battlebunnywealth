# v0.13 security-review rehearsal

This document is an **internal pre-audit rehearsal**, not an external security audit and not evidence that an independent reviewer has approved Battle Bunny Wealth. It exercises the v0.12 security-review handoff against the v0.13 feature branch before merge and records the issues that should be handed to a future independent reviewer.

Review target:

- repository: `chrisbirster/battlebunnywealth`
- branch: `feature/v0.13-transition-activation-soak`
- PR: #16
- base: v0.12 `dev` commit `b67e81ee1d8949817238eb9b8d50d6c8ef5b5af9`
- economic status: TEST-CARROT only; explicitly valueless

The exact final feature-head and squash-merge SHAs are recorded by GitHub/CI and should be supplied to an external reviewer together with this document.

## Invariants exercised

### Consensus safety

- Node count does not create validator votes.
- Committee membership is derived from the validator set active at the exact block height.
- A removed validator remains valid for historical certificates but cannot vote after its activation boundary.
- A new validator cannot vote before the committed activation height.
- Protocol version changes only at a finalized, notice-delayed activation height.
- Unsupported scheduled protocol versions fail closed.
- Restart and catch-up re-run the same historical validator/protocol rules.
- Quorum is not lowered because nodes or validators are offline.

### State and TEST-CARROT

- State roots are deterministic across independent engines.
- Every TEST-CARROT transfer uses an exact consensus nonce.
- Mempool future-nonce queueing does not weaken consensus nonce validation.
- Replacement transactions require a bounded fee bump and cannot create an extra pool slot.
- Fixed supply is conserved through transfers, fees, Proof-of-Play release, restart, and transition activation.
- Finality rewards use the validator set that was active at the rewarded height.

### Networking

- Public observer/full-node count is not voting power.
- Public discovery enforces total, per-host, IPv4 `/24`, and IPv6 `/48` limits.
- v0.13 can additionally enforce trusted server-side ASN/provider caps. Peer-declared network-provider metadata is deliberately not trusted.
- Signed permissioned relay messages remain network-bound and replay protected.

### Recovery and independent verification

- The soak harness crosses validator and protocol activation boundaries while isolating/catching up nodes, changing valid quorum subsets, restarting nodes, and replaying the chain.
- Public checkpoints commit finalized hash, state root, protocol version, validator-set hash, genesis, and CARROT policy.
- Checkpoint verification independently replays finalized blocks; a checkpoint is not a shortcut around finality verification.
- Consensus snapshots are observable but are not authoritative without chain/checkpoint verification.

## Findings / known limitations

### R-01 — No formally verified BFT round-change protocol

**Severity for economic launch:** high.

The current research consensus handles deterministic committee/proposer/finality and safe quorum failure, but it is not a complete formally analyzed Tendermint/HotStuff-class implementation. Adversarial round changes, adaptive corruption, proposer censorship, network asynchrony, and long-lived partial partitions need deeper design and review.

**Disposition:** release blocker for economically valuable CARROT; continue public-testnet research.

### R-02 — Protocol v4 is a compatibility activation exercise, not arbitrary hot code loading

**Severity for economic launch:** medium/high.

v0.13 proves that a finalized upgrade commitment can change the accepted block protocol version at an exact height and survive replay. The binary supports one known compatibility step. It does not download or execute arbitrary code based on chain data.

**Disposition:** correct fail-closed behavior for the testnet. Future upgrades need explicit versioned executable semantics and independent review.

### R-03 — Emergency rollback is intentionally absent

**Severity for operations:** medium.

Finalized history is not silently rewritten. This is safer than an admin rollback, but the project does not yet have a formally specified emergency governance/recovery mechanism for a catastrophic implementation bug after finality.

**Disposition:** document and design before economic launch; current incident runbook preserves evidence and favors halt/replay over ad-hoc rollback.

### R-04 — ASN/provider classification needs a trusted data source

**Severity:** medium.

The code provides a classifier boundary and limits, but production-quality ASN/cloud-provider attribution is external data. Incorrect or stale mappings can reduce diversity or unfairly reject peers.

**Disposition:** research feature only. Compare multiple data sources and measure false-positive/evasion behavior before enforcing strict production limits.

### R-05 — Device attestation does not prove unique humans

**Severity for Proof-of-Play decentralization:** high.

App Attest and Play Integrity raise the cost of fake clients/devices but do not turn one device into one independent person. A patient real-phone farm remains a primary Sybil threat.

**Disposition:** validator maturation/rate-limited activation and attack simulation remain mandatory; do not claim personhood security.

### R-06 — Bootstrap trust remains explicit

**Severity:** expected bootstrap limitation.

The network begins from four known genesis participant devices with 3-of-4 bootstrap finality. The long-term objective is to safely reduce dependence on those identities, not conceal the initial trust assumption.

**Disposition:** continue measuring validator-set growth and define protocol-encoded bootstrap-privilege sunset conditions.

### R-07 — Mempool ordering is deterministic but intentionally simple

**Severity:** low/medium.

Sequential nonce queueing is bounded to 16 transactions per sender and replacement requires a 12.5% fee increase (minimum one atom). Ordering is sender/nonce deterministic, not a production fee market.

**Disposition:** adequate for valueless TEST-CARROT; revisit economics/fair ordering before production fees have value.

### R-08 — Checkpoints are verification aids, not weak subjectivity or trusted snapshots

**Severity:** informational.

A checkpoint only verifies after replaying the chain segment from genesis under current implementation. Faster snapshot sync/state proofs are future work.

**Disposition:** keep the stronger verification semantics until a snapshot trust/state-proof design is reviewed.

### R-09 — No economic slashing or stake-based Sybil defense

**Severity:** high if future economics depend on it.

Proof of Play currently uses attestation, participation, authority, maturation, and randomized committee selection. There is no slashing system and CARROT does not buy validator power.

**Disposition:** intentional. Do not add stake/slashing casually; model the security/economic consequences first.

## Review artifacts to provide externally

An independent reviewer should receive:

- exact v0.13 squash commit and reproducible build instructions;
- `docs/security-review-package.md` plus this rehearsal;
- `docs/testnet-incident-response.md`;
- CARROT policy/hash and supply-invariant tests;
- Proof-of-Play simulator/bootstrap attack scenarios;
- App Attest/Play Integrity server validation code;
- consensus, transition activation, checkpoint, mempool, peer-diversity, and soak tests;
- all GitHub Actions logs for the exact reviewed SHA;
- disclosure of bootstrap validators/operators and testnet topology used for Internet testing.

## Economic release decision

Nothing in this rehearsal authorizes economic activation. Production CARROT remains blocked on independent security review and the separate legal, tax, privacy, and app-store reviews already listed in the roadmap/security package.
