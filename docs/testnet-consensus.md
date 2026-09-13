# v0.9 testnet consensus

## Identities

`NodeKey` is an Ed25519 infrastructure identity used to authenticate permissioned peer-to-peer relay traffic. A node key cannot create a committee vote.

`Validator` is a player/device consensus identity. The protocol supports Ed25519 fixtures for deterministic testing and P-256 SPKI keys for phone/device signing. A valid vote must verify against a validator selected for the exact height and round.

## Genesis

Genesis commits to a network ID, protocol configuration, creation time, and sorted initial validator set. Its deterministic hash is the initial chain head. A data directory is permanently bound to one genesis hash.

The intended first Battle Bunny testnet uses four genesis phone identities and a 3-of-4 bootstrap quorum.

## Round

For height `H` and round `R`:

1. derive the committee deterministically from the active validator set, previous finalized hash, height, round, and authority weights;
2. derive the expected proposer from that committee;
3. proposer constructs a block with previous block hash, previous state root, operations, committee commitment, and deterministic next state root;
4. proposer signs the block hash;
5. every node independently validates the proposal;
6. selected committee devices sign `commit` votes;
7. when the configured quorum of distinct valid members is reached, nodes create a finality certificate and finalize the same block.

Authority is used only for committee selection. It is not applied again to vote weight. One selected committee member equals one vote.

## Quorum and failure behavior

For committees up to 15 members, v0.9 uses `ceil(3/4 * N)`. Four members therefore require three signatures. The mature policy uses `ceil(2/3 * N)`.

A node never lowers quorum because validators are offline. If the remaining committee cannot meet quorum, the network stops finalizing. That is the desired safety behavior for the research protocol.

## Finality certificate

A finalized block carries the original proposer signature and individual signed committee votes. Nodes independently check network ID, height, parent, state transition, committee commitment, proposer, vote membership, vote signatures, distinct validator IDs, and quorum before importing it.

No BLS aggregation is used in v0.9; the explicit vote list is intentionally simple and auditable.

## Persistence and recovery

Finalized blocks are appended to `blocks.ndjson`. On restart the engine begins from genesis and verifies every stored finalized block in order. Corrupt or incompatible state fails startup rather than being trusted.

A stale node can request blocks from a peer, but it verifies the remote genesis hash and every finality certificate locally before advancing.

## Limitations

This is not a formally verified BFT protocol. v0.9 does not yet implement a full multi-phase Tendermint-style round-change protocol, adaptive-corruption defenses, public peer discovery, economic slashing, CARROT, or permissionless validator admission. Those remain later research milestones.
