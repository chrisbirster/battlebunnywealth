# v0.14 round-change protocol

v0.14 replaces the single-round research loop with an explicit, validator-signed round-change path. It remains a research BFT protocol, not a formally verified implementation.

## Goal

A missing, censored, partitioned, or crashed proposer must not permanently halt one height. The protocol may rotate to a new proposer only after a quorum of the height's committee signs a transition to the next round.

Safety is preferred over aggressive liveness. If the node observes incompatible lock evidence, it fails closed rather than guessing which value is safe.

## Committee and proposer

Committee membership is fixed for the entire height. The committee is selected using the active validator set, previous finalized hash, height, and the round-zero selection seed.

Only the proposer rotates by round. `ExpectedProposer(committee, height, round)` selects the proposer.

Fixing committee membership preserves quorum intersection across round changes. Re-selecting a materially different committee every round would weaken the assumptions behind carrying locks from one round into another.

## Round zero

Every height begins at round zero.

A valid round-zero block has no `roundCertificate`. The expected proposer signs the block. Committee members independently validate the block and may sign `commit` votes.

New live validator votes sign both the concrete block hash and a consensus `valueHash`. The value hash covers the state transition rather than the proposer-specific envelope. It commits to protocol version, network, height, parent, prior state, resulting state root, and operations. It deliberately excludes proposer ID, timestamp, round number, and round certificate so the same transition may be reproposed by a later proposer.

A validator that signs a commit vote records a local lock containing:

```text
validator id
round
value hash
signed vote proof
```

Legacy finalized testnet history that predates v0.14 can still replay its older vote-signature domain, but new live votes must carry the signed value hash.

## Timeout and round change

The networking/client layer decides that the current round has timed out. The consensus engine does not trust a local clock timeout as consensus proof.

Each selected validator may sign a `RoundChange` containing:

```text
network
height
from round
to round
validator id
highest local locked round
highest local locked value
signed vote proof for that lock, if any
```

A lock claim without its validator-signed vote proof is invalid. The proof must match the same network, height, validator, locked round, and value hash, and its vote signature must verify against the validator key.

A round-change certificate does **not** globally adopt a value merely because one validator reports a lock. For a committee of `N` with quorum `Q`, a value is carried forward only when matching lock proofs reach the quorum-intersection threshold:

```text
max(1, 2Q - N)
```

This is the minimum number of prior-finality voters guaranteed to appear in any later quorum-sized timeout certificate if that value may already have finalized elsewhere. In the four-phone 3-of-4 bootstrap, two matching lock proofs are therefore required. A single malicious validator can prove that it signed an arbitrary value, but that lone proof cannot dictate the next-round proposal.

When a quorum of distinct committee members requests the same next round, the node builds a `RoundCertificate` containing those signed messages. Among lock groups that meet the intersection threshold, the highest locked round wins. Conflicting qualifying values at the same highest round fail closed.

## Later-round proposal

A proposal for round `R > 0` must carry the quorum certificate that authorizes round `R`.

If the certificate carries a locked value, the new proposer must re-propose that same state transition. A different transaction/operation set is rejected.

If the certificate contains no qualifying lock, the new proposer may propose a new valid transition. Individual validators remain bound by their own proven local locks and cannot sign a conflicting value.

A block finalized in a later round permanently contains the round certificate, including its lock proofs, so restart, catch-up, checkpoints, and independent operators can verify why that round and any carried value were authorized.

## Validator locks

A validator that has voted for value `A` at a height cannot later vote for value `B` at that height.

The engine persists locks and their signed vote proofs in `round-state.json`. A crash/restart verifies those proofs before restoring the current round, lock set, and round-certificate chain.

Finalized history supersedes stale in-progress state. If a crash happens after `blocks.ndjson` is synced but before old round state is removed, restart discards the stale lower-height round snapshot.

## Equivocation evidence

v0.14 can construct and independently verify evidence for:

- one proposer signing two different blocks for the same height/round/proposer slot;
- one validator signing commit votes for two different block hashes at the same height/round.

Evidence contains the original signed messages. The protocol does **not** automatically slash, ban, confiscate funds, or remove a validator based on this experimental evidence path. Governance/incident handling can review it later.

## Persistence and peer catch-up

`Store` persists finalized blocks plus in-progress round state. Peer catch-up verifies every imported finalized block and now durably appends verified remote history before returning success.

A synced node therefore retains imported history after restart instead of relying on another peer to serve it again.

## Chaos gate

`pop-round-chaos` deterministically exercises:

- abandoned round-zero proposals;
- partial value locks with signed lock proofs;
- quorum-intersection lock selection;
- quorum round changes;
- proposer rotation;
- message reordering;
- duplicate messages;
- temporary node isolation and catch-up;
- proposer timestamp skew;
- rolling reconstruction from genesis/finalized history;
- state-root/finalized-hash convergence;
- TEST-CARROT supply conservation.

Run:

```bash
go run ./cmd/pop-round-chaos -blocks 240 -seed 42 -gate
```

## Deliberate limitations

This is not a claim that the protocol is production BFT. In particular:

- there is no formal safety proof;
- timeout scheduling/adaptive backoff is still an operator/client concern;
- the lock model is deliberately conservative and may sacrifice liveness rather than unlock ambiguously;
- quorum-intersection lock proofs reduce the single-validator liveness attack but do not replace a formal multi-phase BFT proof;
- no automatic slashing exists;
- adaptive corruption and sophisticated network scheduling remain research threats;
- the CI chaos harness is deterministic simulation, not evidence from independent Internet operators.

TEST-CARROT remains valueless throughout v0.14.
