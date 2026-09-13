# Testnet consensus

## Identities

`NodeKey` is an Ed25519 infrastructure identity used to authenticate peer-to-peer relay traffic. A node key cannot create a committee vote.

`Validator` is a player/device consensus identity. The protocol supports Ed25519 fixtures for deterministic testing and P-256 SPKI keys for phone/device signing. A valid vote must verify against a validator selected for the exact height.

Validator reward addresses are separate from validator signing keys. Public validator applications bind the reward address into the validator-signed application payload.

## Genesis

Genesis commits to network ID, protocol configuration, creation time, sorted initial validator set, and executable CARROT policy hash. Its deterministic hash is the initial chain head. A data directory is permanently bound to one genesis hash.

The chain genesis remains protocol v3. The binary supports the explicitly reviewed v4 compatibility activation path from v0.13; it cannot download arbitrary consensus code from the chain.

The intended bootstrap still begins with four genesis phone identities and a 3-of-4 quorum.

## Height-scoped rules

Before validating height `H`, consensus resolves from replicated state:

- validator set active at `H`;
- block protocol version active at `H`.

Historical boundaries are retained. A validator removed at `H` cannot vote at `H`, but its signatures on blocks before `H` remain verifiable.

## Committee and rounds

Committee membership is fixed for the full height. It is selected from the active validator set using the previous finalized hash, height, and the round-zero committee seed. Authority affects this selection only.

Proposer selection then rotates within that fixed committee by round.

Fixing the committee across rounds preserves quorum-intersection assumptions. One selected committee member equals one vote.

Every height starts at round zero. The expected proposer builds operations, previews the deterministic state transition, signs the resulting block hash, and committee members may sign `commit` votes.

## Value locks and signed proofs

New live commit votes bind both the concrete block hash and a proposer-independent consensus `valueHash`. The value hash commits to protocol version, network, height, parent, prior state, resulting state root, and operations.

It excludes proposer ID, timestamp, round, and round certificate. A later proposer can therefore re-propose the same transition without changing the locked value.

When a validator signs a live commit vote it records a local lock containing the validator ID, locked round, value hash, and the signed vote itself as cryptographic proof. A validator may not vote for a different value later in the same height.

Legacy finalized round-zero testnet history from before v0.14 can still replay the older vote-signature domain without a value hash. New live votes use the value-bound signature domain, and later-round finality requires value-bound votes.

## Round change

If the current proposer does not lead to finality, validators may sign `RoundChange` messages. The engine does not treat one node's local timeout as consensus proof.

A round-change message commits to:

- network and height;
- old and requested next round;
- validator ID;
- validator's highest local locked round/value;
- the signed value-bound vote proving that lock, if a lock is claimed.

A claimed lock without its signed vote proof is invalid. The proof must match the same network, height, validator, locked round, and value hash, and its validator signature must verify.

A quorum of distinct committee round-change signatures creates a `RoundCertificate`, but one validator's lock is not enough to constrain the entire next round. For committee size `N` and quorum `Q`, matching lock proofs must reach:

```text
max(1, 2Q - N)
```

before the certificate carries that value as a global lock. This is the minimum intersection of two quorum-sized sets. In the four-device 3-of-4 bootstrap, two matching proven locks are required.

Among qualifying lock groups, the highest locked round is carried forward. Conflicting qualifying values at the same highest round fail closed. Individual validators remain locally bound by their own proven locks even when those locks do not reach the certificate-wide threshold.

A block for round greater than zero must embed the round certificate that authorizes that round. If the certificate carries a qualifying lock, the new proposer must propose the same consensus value.

The full certificate and its lock proofs remain inside a later-round finalized block so imported history can be verified without trusting the importing node's prior in-memory timeout state.

See [Round-change protocol](round-change-protocol.md) for the detailed safety model.

## Quorum and failure behavior

For committees up to 15 members, bootstrap quorum is `ceil(3/4 * N)`. Four members therefore require three signatures. The mature policy uses `ceil(2/3 * N)`.

A node never lowers quorum because validators are offline. If quorum cannot be reached, finality stops.

v0.14 deliberately prefers a halt to unsafe unlocking when qualifying lock evidence is ambiguous.

## Finality certificate

A finalized block carries the proposer signature and individual signed committee votes. Nodes independently check network, active protocol version, height, parent, transition, committee, proposer, round authorization, vote signatures, value hashes, distinct validator IDs, and quorum before import.

Later-round blocks additionally carry the round certificate and signed lock proofs that authorized proposer rotation and any carried value.

No BLS aggregation is used yet; the explicit vote list is intentionally auditable.

## Equivocation evidence

v0.14 can package signed evidence when:

- one proposer signs two different blocks for the same height/round slot;
- one validator signs two commit votes for different block hashes at the same height/round.

Evidence is independently signature-verifiable. It does not trigger automatic slashing, confiscation, banning, or validator removal.

## Canonical reward settlement

Different honest nodes may observe different valid quorum subsets for the same finalized block. Those subsets cannot directly drive local rewards or balances could diverge.

Height `H+1` therefore includes a `test-carrot-finality-settlement` selecting the exact prior-finality vote set used for Proof-of-Play release and fee distribution.

Settlement signatures are checked against the validator set active at rewarded height `H`, even when a different set activates at `H+1`.

## TEST-CARROT transaction state

Consensus state tracks balances, exact nonces, pending fees, participation-reserve release, and state height.

Public admission is not execution. Every `test-carrot-transfer` is revalidated against exact consensus state.

The mempool may queue up to 16 future sequential nonces per sender. Proposals include only the contiguous executable prefix. Same-nonce replacement requires a 12.5% fee increase with a minimum one-atom bump. Consensus still requires exact sequential nonces.

TEST-CARROT remains explicitly valueless.

## Validator and protocol activation

A validator-set commitment requires at least 144 blocks of notice and commits the complete sorted future active set. At `activationHeight`, the new set becomes authoritative while prior heights retain their historical set.

A protocol-upgrade commitment requires at least 1,008 blocks of notice. At the activation height, expected block version changes deterministically. A node that does not explicitly implement that version fails closed.

No chain operation downloads or executes arbitrary replacement code.

## Persistence and recovery

Finalized blocks are appended to `blocks.ndjson`. In-progress round state is atomically stored in `round-state.json`, including current round, validator locks with signed vote proofs, and the verified round-certificate chain.

On restart, the engine reconstructs finalized history from genesis, re-verifies persisted lock proofs, and then restores only a round snapshot for the exact next height. If a crash leaves a stale lower-height round snapshot after a finalized block was already fsynced, finalized history wins and the stale snapshot is deleted.

Peer catch-up independently verifies imported blocks, round certificates, lock proofs, finality certificates, state transitions, and state roots, then durably appends verified history locally. A restarted node therefore retains peer-synced finality.

## Snapshots and checkpoints

`/v1/public/snapshot` exposes deterministic consensus state for observability.

`/v1/public/checkpoint` commits to network/genesis, height, finalized hash, state root, protocol version, validator-set hash, and CARROT policy hash.

A checkpoint is not trusted state. `VerifyCheckpoint` rebuilds from genesis and finalized history. `pop-checkpoint-compare` compares same-height commitments from independent operators and fails on finalized-hash/state-root disagreement.

## Public peer boundary

Observer/full-node discovery creates no votes. The directory enforces total, per-host, IPv4 `/24`, and IPv6 `/48` limits.

An optional trusted server-side network map adds ASN/provider caps. The map is canonicalized and hash-pinned; peers cannot self-declare trusted metadata.

## Reliability gates

`pop-soak` runs the v0.13 1,100-block transition/restart soak.

`pop-round-chaos` adds deterministic round changes, partial proven locks, quorum-intersection lock selection, proposer rotation, reordered and duplicate messages, temporary partitions, clock-skewed proposal timestamps, rolling full replay, convergence checks, and TEST-CARROT supply checks.

These are CI simulations, not claims about real geographic/provider diversity.

## Limitations

The protocol is not formally verified. v0.14 adds explicit round changes, signed lock proofs, and a quorum-intersection carried-lock rule, but it is still a conservative research design rather than a formally reviewed Tendermint/HotStuff implementation. Adaptive corruption, sophisticated scheduler attacks, a fully specified adaptive pacemaker, automatic slashing policy, production timeout tuning, arbitrary future upgrade semantics, and real multi-provider Internet evidence remain open before a production-security claim.
