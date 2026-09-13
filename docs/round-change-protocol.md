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

A validator that signs a commit vote records a local value lock:

```text
validator id
round
block value hash
```

The value hash covers the state transition rather than the proposer-specific envelope. It commits to protocol version, network, height, parent, prior state, resulting state root, and operations. It deliberately excludes proposer ID, timestamp, round number, and round certificate so the same transition may be reproposed by a later proposer.

## Timeout and round change

The networking/client layer decides that the current round has timed out. The consensus engine does not trust a local clock timeout as consensus proof.

Each selected validator may sign a `RoundChange`:

```text
network
height
from round
to round
validator id
highest local locked round
highest local locked value
```

The engine verifies the signature and requires a validator that is already locally locked to disclose that exact lock.

When a quorum of distinct committee members requests the same next round, the node builds a `RoundCertificate` containing those signed messages.

The certificate is independently verifiable and records the highest disclosed lock. Conflicting values at the same highest locked round fail closed.

## Later-round proposal

A proposal for round `R > 0` must carry the quorum certificate that authorizes round `R`.

If the certificate carries a locked value, the new proposer must re-propose that same state transition. A different transaction/operation set is rejected.

If the certificate contains no lock, the new proposer may propose a new valid transition.

A block finalized in a later round permanently contains the round certificate, so restart, catch-up, checkpoints, and independent operators can verify why that round was authorized.

## Validator locks

A validator that has voted for value `A` at a height cannot later vote for value `B` at that height.

The engine persists locks in `round-state.json`. A crash/restart reloads the current round, lock set, and verified round-certificate chain before the validator can continue.

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
- partial value locks;
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
- no automatic slashing exists;
- adaptive corruption and sophisticated network scheduling remain research threats;
- the CI chaos harness is deterministic simulation, not evidence from independent Internet operators.

TEST-CARROT remains valueless throughout v0.14.
