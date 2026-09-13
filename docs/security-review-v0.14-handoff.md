# v0.14 external security-review handoff

This is an internal handoff package for a future independent security review. It is not an external audit report and does not claim production readiness.

## Scope added since v0.13

v0.14 addresses several pre-audit findings:

- **Single-round liveness:** explicit validator-signed round changes and rotating proposers now exist.
- **Round safety across restart:** current round, value locks, signed lock proofs, and verified round certificates are persisted and restored.
- **Unsafe proposer replacement:** later-round proposals carry quorum certificates and must preserve any value whose matching lock proofs meet the quorum-intersection threshold.
- **Unproved-lock liveness attack:** a round-change lock now requires the validator's signed value-bound vote, and a lone valid lock cannot globally constrain the next round. The carried-value threshold is `max(1, 2Q-N)` for committee size `N` and quorum `Q`.
- **Proposer/validator equivocation evidence:** conflicting signed proposals/votes can be packaged and independently verified without automatic punishment.
- **Peer catch-up durability:** independently verified blocks imported from peers are durably persisted before a later restart.
- **Operator disagreement visibility:** checkpoint comparison fails on same-height finalized-hash/state-root disagreement.
- **Opaque peer diversity data:** ASN/provider classification can use a canonical, hash-pinned CIDR map rather than peer-declared labels or undocumented runtime results.
- **Network-disorder coverage:** deterministic chaos exercises reordered/duplicate messages, partial locks, lock proofs, round changes, partitions, clock skew, and rolling replay.

## Security invariants reviewers should challenge

1. A node cannot advance to round `R+1` without a quorum certificate from the fixed committee for that height.
2. A validator that has voted for value `A` at height `H` cannot later vote for value `B` at `H` after restart or round change.
3. A claimed lock is ignored unless its signed vote proves the same network, height, validator, round, and value hash.
4. One malicious validator's valid lock proof cannot globally force a carried value in the four-validator 3-of-4 bootstrap; two matching proofs are required by quorum intersection.
5. If a value may already have reached a prior finality quorum, every later quorum-sized round-change certificate contains at least `2Q-N` members from that prior quorum.
6. A later-round proposer cannot replace a value selected by the certificate's qualifying highest lock group.
7. Round certificates embedded in finalized blocks can be verified by a fresh node replaying history from genesis.
8. Committee membership is stable across rounds at a height; only the proposer rotates.
9. Historical validator-set and protocol-version boundaries remain authoritative during replay.
10. Peer synchronization never bypasses block/finality/state-transition verification and imported finality survives restart.
11. TEST-CARROT fixed-supply conservation remains true through round changes, restarts, and partitions.
12. Full-node count still creates no validator votes.
13. Equivocation evidence cannot itself mutate consensus state or confiscate balances.

## Primary commands

```bash
go test ./...
go vet ./...
go run ./cmd/pop-bootstrap-sim -phones 100 -gate
go run ./cmd/pop-public-smoke -attackers 100 -gate
go run ./cmd/pop-soak -blocks 1100 -gate
go run ./cmd/pop-round-chaos -blocks 240 -seed 42 -gate
```

Independent-operator tooling:

```bash
go run ./cmd/pop-checkpoint-compare -url URL_A -url URL_B
go run ./cmd/pop-network-evidence -url URL_A -url URL_B -gate
```

## Remaining high-priority review questions

- Does the `2Q-N` carried-lock rule preserve safety under every adversarial message schedule supported by this one-phase commit model, including hidden finality and selective timeout participation?
- Can multiple qualifying lock groups or selectively omitted proofs still create permanent liveness loss near the tolerated Byzantine bound?
- Are timeout/backoff rules sufficiently specified outside the engine to avoid synchronized churn or denial-of-service amplification?
- Can a compromised validator or proposer use message ordering to create evidence that appears valid against an honest participant?
- Are validator-set changes near round changes safe at exact activation boundaries?
- Are checkpoint and evidence APIs exposing enough information for diagnosis without leaking unnecessary participant/device identity?
- Can public discovery/network metadata controls be bypassed cheaply through multi-provider hosting, proxies, or stale CIDR data?
- Does the real-device/phone-farm activation model remain sufficiently costly over months rather than only over the CI simulation horizon?

## Known limitations / blockers

The following remain blockers for an economically valuable network:

- no formal proof or independent consensus-protocol audit;
- no completed real multi-provider/geographic public-testnet evidence window;
- timeout policy remains operator/client-side rather than a fully specified adaptive pacemaker;
- the current round-change model is a conservative research design rather than a formally reviewed Tendermint/HotStuff implementation;
- no automatic slashing design, intentionally;
- no mature DDoS/eclipsing guarantee;
- device attestation is not unique-human proof;
- the patient real-phone-farm problem remains open research;
- no production treasury execution;
- no production-value CARROT issuance;
- external legal/tax/privacy/app-store review remains outstanding.

## Frozen-review procedure

The independent review target should be a specific merged `dev` commit after v0.14 CI and post-merge CI are green. Record that commit SHA together with:

- this handoff;
- `round-change-protocol.md`;
- `distributed-testnet-evidence.md`;
- the v0.12/v0.13 review packages;
- genesis and CARROT policy hashes;
- network-map hash/provenance used during live evidence collection;
- the CI/chaos artifacts;
- live public-testnet evidence artifacts when they exist.

Do not label the internal CI result as an independent audit.

TEST-CARROT remains valueless.
