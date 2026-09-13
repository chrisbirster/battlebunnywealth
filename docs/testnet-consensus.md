# Testnet consensus

## Identities

`NodeKey` is an Ed25519 infrastructure identity used to authenticate peer-to-peer relay traffic. A node key cannot create a committee vote.

`Validator` is a player/device consensus identity. The protocol supports Ed25519 fixtures for deterministic testing and P-256 SPKI keys for phone/device signing. A valid vote must verify against a validator selected for the exact height and round.

Validator reward addresses are separate from validator signing keys. Public validator applications bind the reward address into the validator-signed application payload.

## Genesis

Genesis commits to a network ID, protocol configuration, creation time, sorted initial validator set, and the executable CARROT policy hash. Its deterministic hash is the initial chain head. A data directory is permanently bound to one genesis hash.

v0.12 uses protocol version 3 because TEST-CARROT balances/nonces and finality settlement are now deterministic replicated state. Protocol-v2 data directories are intentionally incompatible with protocol v3.

The intended bootstrap still begins with four genesis phone identities and a 3-of-4 quorum.

## Round

For height `H` and round `R`:

1. derive the committee deterministically from the active validator set, previous finalized hash, height, round, and authority weights;
2. derive the expected proposer from that committee;
3. proposer builds an operation list containing the prior-finality settlement when `H > 1`, plus deterministic mempool transactions and any valid transition commitments;
4. the replicated state machine previews those operations and returns the deterministic next state root;
5. proposer constructs the block with previous block hash, previous state root, operations, committee commitment, and next state root;
6. proposer signs the block hash;
7. every node independently replays the preview and validates the proposal;
8. selected committee devices sign `commit` votes;
9. when the configured quorum of distinct valid members is reached, nodes create a finality certificate;
10. only then does the state machine commit the exact transition and the block becomes finalized.

Authority is used only for committee selection. It is not applied again to vote weight. One selected committee member equals one vote.

## Quorum and failure behavior

For committees up to 15 members, the bootstrap uses `ceil(3/4 * N)`. Four members therefore require three signatures. The mature policy uses `ceil(2/3 * N)`.

A node never lowers quorum because validators are offline. If the remaining committee cannot meet quorum, the network stops finalizing.

## Finality certificate

A finalized block carries the original proposer signature and individual signed committee votes. Nodes independently check network ID, protocol version, height, parent, state transition, committee commitment, proposer, vote membership, vote signatures, distinct validator IDs, and quorum before importing it.

No BLS aggregation is used yet; the explicit vote list is intentionally simple and auditable.

## Canonical reward settlement

Different honest nodes may observe different valid quorum subsets for the same finalized block. Those subsets cannot be used directly as a local reward source or balances could diverge.

For that reason, height `H+1` must contain a `test-carrot-finality-settlement` operation for height `H`. Consensus on the next block therefore chooses the exact valid prior-finality vote set used for Proof-of-Play release and fee distribution.

The state machine validates those signatures, sorts validator IDs, distributes rewards deterministically, and preserves the fixed supply.

## TEST-CARROT transaction state

Protocol v3 tracks balances, nonces, pending fees, participation-reserve release, and state height.

Public transaction admission is not sufficient for execution. The consensus state machine revalidates every `test-carrot-transfer` operation against the exact state at that block height. Invalid signatures, network IDs, nonces, expiry, duplicate IDs, or insufficient balances make the proposal invalid.

TEST-CARROT remains explicitly valueless.

## Network-transition commitments

Protocol v3 recognizes two finalized plan operations:

- `validator-set-commitment`
- `protocol-upgrade-commitment`

Validator-set commitments require at least 144 blocks of notice. Protocol-upgrade commitments require at least 1,008 blocks of notice.

These operations commit transition metadata into a quorum-finalized block and are revalidated during restart/catch-up. They do **not** automatically hot-swap the active committee set or executable consensus code. A later milestone must implement and review an explicit activation procedure.

## Persistence and recovery

Finalized blocks are appended to `blocks.ndjson`. On restart the engine begins from genesis and verifies every stored finalized block in order through the same state-machine transition path. Corrupt or incompatible state fails startup rather than being trusted.

A stale node can request blocks from a peer, but it verifies the remote genesis hash, every finality certificate, every replicated state transition, and every resulting state root locally before advancing.

Finalized validator-set and protocol-upgrade commitments therefore survive recovery as part of the verified block history.

## Public peer boundary

Public observer/full-node discovery does not create votes. The directory enforces total, per-host, and IP-prefix admission limits. IPv4 peers are grouped by `/24`; IPv6 peers are grouped by `/48`.

These controls make simple eclipse floods more expensive but do not prove ASN, geographic, cloud-provider, or organizational diversity.

## Limitations

This is not a formally verified BFT protocol. It still lacks a full multi-phase Tendermint-style round-change protocol, adaptive-corruption defenses, economic slashing, automatic finalized validator-set activation, automatic executable protocol upgrades, and economically valuable CARROT.
