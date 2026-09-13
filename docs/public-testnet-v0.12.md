# v0.12 public testnet execution

v0.12 moves TEST-CARROT from admission-only testing into deterministic replicated consensus state. TEST-CARROT remains valueless and is not production CARROT.

## Replicated state

`internal/publictestnet.ConsensusState` is the protocol-v3 state machine used by `pop-node`.

It deterministically tracks:

- fixed genesis CARROT allocation buckets;
- TEST-CARROT wallet balances;
- exact per-wallet nonces;
- pending transaction fees;
- Proof-of-Play release already settled;
- the current state height.

Every proposal previews the same transition and commits the resulting state root into the block. The transition mutates state only after finality. Restart and peer catch-up replay finalized blocks through the same state machine and must reproduce the same state root.

## Transaction execution

The public mempool remains bounded and accepts only correctly signed, network-bound, unexpired transactions using the exact next nonce. A proposer converts the deterministic mempool ordering into `test-carrot-transfer` operations.

Consensus then independently rechecks the transaction against the state at the proposed height. Invalid balance, invalid nonce, replay, signature tampering, wrong network, duplicate transaction ID, or expiry causes the proposal transition to fail.

Transactions finalized in a block are removed from the local mempool. Mempool contents themselves are not consensus state.

## Canonical finality settlement

A node can locally reach finality after receiving any valid quorum subset. Different honest nodes may therefore hold different valid vote subsets for the same block.

Rewards cannot be derived directly from each node's local subset without risking state divergence. v0.12 resolves this by requiring height `H+1` to include a `test-carrot-finality-settlement` operation for height `H`.

The next block therefore reaches consensus on the exact valid prior-finality vote set used for:

- deterministic Proof-of-Play release;
- fee distribution;
- reward recipients.

Selected validator IDs are sorted before equal distribution. Remainder atoms are assigned deterministically in sorted order. The participation reserve and total fixed supply remain conserved.

## Validator reward addresses

A validator may bind a `tcarrot1` reward address. That address is part of the signed validator application message. An intermediary cannot replace it while forwarding an otherwise valid application.

Genesis validators without a TEST-CARROT reward address use an internal deterministic validator account in research networks.

## Finalized network-transition commitments

v0.12 adds two consensus operation types:

- `validator-set-commitment`
- `protocol-upgrade-commitment`

A validator-set plan includes the full sorted proposed validator set, activation height, network ID, and deterministic plan hash. It requires at least 144 blocks of notice.

A protocol-upgrade plan commits the current/next protocol version, activation height, CARROT policy hash, minimum software version, network ID, and deterministic plan hash. The existing minimum notice is 1,008 blocks.

These are **finalized commitments, not automatic hot-swaps**. Protocol v3 validates and finalizes the plan metadata, and recovery replay revalidates the exact operation. Applying a new executable protocol or committee set still requires an explicitly supported activation path in a later milestone. v0.12 does not silently mutate consensus rules underneath a running node.

## Public peer diversity

Public discovery retains per-host and total-directory limits and adds network-prefix limits:

- IPv4 peers are grouped by `/24`;
- IPv6 peers are grouped by `/48`;
- at most 16 peers from one network prefix are admitted by the default directory.

This is only one eclipse-resistance layer. It does not prove geographic, ASN, provider, or organizational diversity.

## Public TEST-CARROT funding policy

`GET /v1/public/funding-policy` returns the hashed v0.12 funding policy.

The default policy is deliberately conservative:

- asset is `TEST-CARROT`;
- economic value is explicitly false;
- no automatic faucet is enabled;
- no mint path exists;
- CARROT treasury spending remains disabled;
- test grants, when operators choose to provide them, must be ordinary signed transfers from an operator-owned TEST-CARROT wallet containing already-existing test balances;
- suggested cap is 100 TEST-CARROT per address per 24 hours, one grant per window.

The policy is operational guidance, not a consensus minting rule. Operators may choose to fund nobody.

## Protocol reset

v0.12 uses testnet protocol version 3. Existing protocol-v2 data directories are intentionally incompatible because replicated state execution changes the meaning of the committed state root.

This is a testnet reset, not a production migration.

## CI gates

CI verifies:

- protocol-v3 genesis and CARROT policy commitment;
- legacy voting, quorum, equivocation, replay, transport, restart, and peer catch-up tests;
- replicated TEST-CARROT balance/nonce convergence;
- fee and reward settlement;
- replay and insufficient-balance rejection;
- restart state-root reproduction;
- finalized validator/upgrade commitment recovery;
- public funding-policy invariants;
- peer host/prefix anti-flood behavior;
- four-phone bootstrap versus phone-farm gate;
- adversarial public-testnet smoke;
- Go vet and all binaries.

## Explicit non-goals

v0.12 does not activate economically valuable CARROT, exchange integration, token sales, treasury spending, slashing, formal BFT verification, autonomous protocol upgrades, or automatic validator-set hot swaps.
