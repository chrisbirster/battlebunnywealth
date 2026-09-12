# ADR 0003 — Validator activation and committee vote weight

Status: accepted for v0.9 permissioned-testnet research.

## Decision

Proof-of-Play authority affects the probability that an active validator is selected for a committee. Once selected, every committee member has exactly one vote. Authority is not applied again to vote weight.

New attested devices do not become active validators immediately. They enter a validator-candidate queue, must satisfy the configured maturation and authority policy, and are activated at a bounded rate. The v0.9 bootstrap starts from four known genesis devices and uses a 3-of-4 quorum while the active set remains small.

Full-node identity is separate from player/device validator identity. Running additional servers provides no additional consensus votes.

## Consequences

- A server Sybil can consume network resources but cannot manufacture committee signatures.
- A real-device farm remains a serious Sybil threat; activation throttling delays rather than proves resistance to that attack.
- The bootstrap network is intentionally permissioned and non-economic.
- CARROT issuance remains disabled until later adversarial testnet milestones.
