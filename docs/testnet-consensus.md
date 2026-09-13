# Permissioned testnet consensus

## Identities

`NodeKey` is an Ed25519 infrastructure identity used to authenticate permissioned peer-to-peer relay traffic. A node key cannot create a committee vote.

`Validator` is a player/device consensus identity. The protocol supports Ed25519 fixtures for deterministic testing and P-256 SPKI keys for phone/device signing. A valid vote must verify against a validator selected for the exact height and round.

## Genesis

Genesis commits to a network ID, protocol configuration, creation time, sorted initial validator set, and the executable CARROT policy hash. Its deterministic hash is the initial chain head. A data directory is permanently bound to one genesis hash.

v0.10 bumps permissioned-testnet genesis to protocol version 2 because CARROT policy is now part of network identity. Nodes reject a genesis whose `carrotPolicyHash` differs from `carrot.DefaultPolicy().Hash()`.

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

For committees up to 15 members, the bootstrap uses `ceil(3/4 * N)`. Four members therefore require three signatures. The mature policy uses `ceil(2/3 * N)`.

A node never lowers quorum because validators are offline. If the remaining committee cannot meet quorum, the network stops finalizing.

## Finality certificate

A finalized block carries the original proposer signature and individual signed committee votes. Nodes independently check network ID, height, parent, state transition, committee commitment, proposer, vote membership, vote signatures, distinct validator IDs, and quorum before importing it.

No BLS aggregation is used yet; the explicit vote list is intentionally simple and auditable.

## CARROT boundary in v0.10

v0.10 commits the fixed CARROT policy into genesis and provides a deterministic ledger/reward state-transition package plus supply reporting. Public wallet transactions and CARROT balance operations are **not yet consensus block operations**. v0.11 is the milestone intended to wire valueless/test CARROT into the adversarial public testnet.

This keeps v0.10 focused on freezing and testing supply mathematics without accidentally turning the permissioned research network into an economically active token system.

## Persistence and recovery

Finalized blocks are appended to `blocks.ndjson`. On restart the engine begins from genesis and verifies every stored finalized block in order. Corrupt or incompatible state fails startup rather than being trusted.

A stale node can request blocks from a peer, but it verifies the remote genesis hash and every finality certificate locally before advancing.

## Limitations

This is not a formally verified BFT protocol. It still lacks a full multi-phase Tendermint-style round-change protocol, adaptive-corruption defenses, public peer discovery, economic slashing, permissionless validator admission, and economically valuable CARROT.
