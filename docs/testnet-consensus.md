# Testnet consensus

## Identities

`NodeKey` is an Ed25519 infrastructure identity used to authenticate peer-to-peer relay traffic. A node key cannot create a committee vote.

`Validator` is a player/device consensus identity. The protocol supports Ed25519 fixtures for deterministic testing and P-256 SPKI keys for phone/device signing. A valid vote must verify against a validator selected for the exact height and round.

Validator reward addresses are separate from validator signing keys. Public validator applications bind the reward address into the validator-signed application payload.

## Genesis

Genesis commits to a network ID, protocol configuration, creation time, sorted initial validator set, and executable CARROT policy hash. Its deterministic hash is the initial chain head. A data directory is permanently bound to one genesis hash.

The chain genesis remains protocol v3. v0.13 adds controlled support for one explicitly understood compatibility upgrade step, protocol v4. This does not mean nodes can download arbitrary consensus code from the chain.

The intended bootstrap still begins with four genesis phone identities and a 3-of-4 quorum.

## Height-scoped consensus rules

Consensus no longer assumes the genesis validator set or block protocol version forever. Before validating height `H`, the engine asks replicated state for:

- the validator set active at `H`;
- the block protocol version active at `H`.

Those answers are derived only from genesis plus previously finalized transition commitments. The same lookup is used for live proposals, votes, imported blocks, restart replay, and peer catch-up.

Historical validator-set/version boundaries are retained. A validator removed at height `H` cannot vote at `H`, but its signatures on blocks `< H` remain verifiable.

## Round

For height `H` and round `R`:

1. resolve the validator set and supported protocol version for height `H`;
2. derive the committee deterministically from that validator set, previous finalized hash, height, round, and authority weights;
3. derive the expected proposer from that committee;
4. proposer builds an operation list containing the prior-finality settlement when `H > 1`, contiguous executable mempool transactions, and any valid transition commitments;
5. the replicated state machine previews those operations and returns the deterministic next state root;
6. proposer constructs the block with protocol version, previous block hash, previous state root, operations, committee commitment, and next state root;
7. proposer signs the block hash;
8. every node independently replays the preview and validates the proposal;
9. selected committee devices sign `commit` votes;
10. when the configured quorum of distinct valid members is reached, nodes create a finality certificate;
11. only then does the state machine commit the exact transition and the block becomes finalized.

Authority is used only for committee selection. It is not applied again to vote weight. One selected committee member equals one vote.

## Quorum and failure behavior

For committees up to 15 members, the bootstrap uses `ceil(3/4 * N)`. Four members therefore require three signatures. The mature policy uses `ceil(2/3 * N)`.

A node never lowers quorum because validators are offline. If the remaining committee cannot meet quorum, the network stops finalizing.

## Finality certificate

A finalized block carries the original proposer signature and individual signed committee votes. Nodes independently check network ID, the protocol version active at that height, height, parent, state transition, committee commitment, proposer, vote membership, vote signatures, distinct validator IDs, and quorum before importing it.

No BLS aggregation is used yet; the explicit vote list is intentionally simple and auditable.

## Canonical reward settlement

Different honest nodes may observe different valid quorum subsets for the same finalized block. Those subsets cannot be used directly as a local reward source or balances could diverge.

Height `H+1` must therefore contain a `test-carrot-finality-settlement` operation for height `H`. Consensus on the next block chooses the exact valid prior-finality vote set used for Proof-of-Play release and fee distribution.

The settlement validator signatures are checked against the validator set that was active at rewarded height `H`, even if a different validator set activates at `H+1`.

## TEST-CARROT transaction state

Consensus state tracks balances, exact nonces, pending fees, participation-reserve release, and state height.

Public transaction admission is not execution. The state machine revalidates every `test-carrot-transfer` against exact consensus state. Invalid signatures, network IDs, nonces, expiry, duplicate IDs, or insufficient balances make the proposal invalid.

v0.13 permits each sender to queue up to 16 sequential/future nonces in the mempool. Proposals include only the contiguous executable prefix beginning at the sender's consensus nonce. Same-nonce replacement requires a 12.5% fee increase with a minimum one-atom bump. These are mempool rules only; consensus still requires exact sequential nonces.

TEST-CARROT remains explicitly valueless.

## Validator-set activation

A `validator-set-commitment` requires at least 144 blocks of notice and commits the complete sorted future active set.

Once the commitment finalizes:

- heights before `activationHeight` use the historical prior set;
- `activationHeight` and later use the committed replacement set;
- removed validators immediately lose voting power at that boundary;
- historical committees/certificates remain verifiable;
- the transition and its plan hash are part of deterministic state and therefore the state root;
- restart/catch-up reproduces the same boundary from finalized history.

Only one future validator-set plan may be pending at a time in v0.13.

## Protocol-upgrade activation

A `protocol-upgrade-commitment` requires at least 1,008 blocks of notice and commits target version, policy hash, activation height, and minimum software identifier.

At the activation height the expected block version changes deterministically. This binary supports only one controlled compatibility step beyond genesis (v3 -> v4). A node that cannot execute the scheduled target must fail closed rather than reinterpret the block.

The testnet does not load arbitrary code, shell out to an updater, or allow a chain message to replace the executable. Real future protocol versions require explicit reviewed implementation before nodes can support them.

## Persistence and recovery

Finalized blocks are appended to `blocks.ndjson`. On restart the engine begins from genesis and verifies every stored finalized block in order through the same state-machine path. Corrupt, incompatible, or unsupported state fails startup rather than being trusted.

A stale node can request blocks from a peer, but it verifies the remote genesis hash, every finality certificate, every historical validator/version boundary, every replicated state transition, and every resulting state root locally before advancing.

## Snapshots and checkpoints

`/v1/public/snapshot` exposes the deterministic consensus snapshot for observability.

`/v1/public/checkpoint` exposes a compact commitment to network/genesis, height, finalized hash, state root, protocol version, validator-set hash, and CARROT policy hash.

A checkpoint is **not** trusted state. `VerifyCheckpoint` reconstructs a fresh node from genesis and finalized blocks through the checkpoint height, then compares all checkpoint commitments. Faster trusted/state-proof snapshot sync remains future work.

## Public peer boundary

Public observer/full-node discovery does not create votes. The directory enforces total, per-host, IPv4 `/24`, and IPv6 `/48` admission limits.

v0.13 adds an optional server-side `NetworkClassifier` for ASN/provider-aware limits. Network metadata must come from a trusted operator-controlled source; peers do not get to self-declare their ASN/provider for diversity enforcement. Unknown classification falls back to existing IP controls.

## Long-running deterministic soak

`cmd/pop-soak` finalizes at least 1,100 blocks across five independent engines. It crosses both transition activation heights, periodically isolates and catches up one node, varies valid quorum subsets, reconstructs/replays nodes from genesis, and checks height/finalized-hash/state-root convergence plus fixed-supply conservation after every block.

The soak is a deterministic CI reliability gate, not a substitute for geographically distributed Internet soak testing.

## Limitations

This is not a formally verified BFT protocol. It still lacks a full multi-phase Tendermint/HotStuff-style round-change protocol, adaptive-corruption defenses, economic slashing, arbitrary future protocol semantics, fast state-proof snapshot sync, and economically valuable CARROT. The four-device bootstrap and real-device Sybil problem remain explicit research assumptions.
