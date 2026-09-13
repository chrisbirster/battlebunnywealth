# v0.13 testnet operations

No paid multi-server deployment is required for local development. The deterministic cluster and soak harnesses run independent consensus engines with separate logical state, while `pop-node` is the long-running node binary.

Genesis remains protocol v3. v0.13 can activate one explicitly supported compatibility step to protocol v4 at a quorum-finalized height. Do not reuse incompatible data directories across different genesis hashes.

## Local security and reliability gates

```bash
go test ./...
go vet ./...
go run ./cmd/pop-cluster -blocks 20
go run ./cmd/pop-bootstrap-sim -phones 100 -gate
go run ./cmd/carrot-spec -check
go run ./cmd/pop-public-smoke -attackers 100 -gate
go run ./cmd/pop-soak -blocks 1100 -gate
```

`task testnet:soak` runs the final command.

The 1,100-block soak crosses the validator-set and protocol-version activation heights, changes valid quorum subsets, periodically isolates/catches up a node, reconstructs/replays nodes from genesis, and checks finalized hash/state-root convergence plus fixed-supply conservation.

## Node key

Generate a full-node Ed25519 relay identity:

```bash
go run ./cmd/pop-node -keygen ./data/testnet/node-a/node.key
```

The private key file is written with mode `0600` and is deliberately separate from JSON chain state. A node identity cannot create a committee vote.

## Run a node

A node config contains `listen`, a complete hashed protocol-v3 `genesis`, peer entries (`nodeId`, `url`, `publicKey`), and `syncIntervalSeconds`.

```bash
go run ./cmd/pop-node \
  -config ./testnet/node-a.json \
  -data ./data/testnet/node-a
```

At startup the node:

1. loads and validates genesis and CARROT policy hash;
2. constructs deterministic TEST-CARROT/transition consensus state;
3. replays every local finalized block through the state machine;
4. reconstructs historical validator-set and protocol-version boundaries;
5. rejects corrupt, incompatible, or unsupported history;
6. begins peer catch-up and HTTP service.

## Consensus endpoints

- `GET /v1/node/status`
- `GET /v1/node/genesis`
- `GET /v1/node/blocks?from=HEIGHT`
- `GET /v1/node/metrics`
- `GET /v1/committee/draft?validator=ID&round=0`
- `POST /v1/committee/proposals`
- `POST /v1/committee/votes`
- signed relay endpoints under `/v1/peer/*`.

Phone clients submit their own signed proposal/vote. A receiving full node verifies the committee signature against the validator set active at that exact height and may relay the same signed object using the node's separate relay identity.

## Public-testnet endpoints

The public wrapper adds:

- `GET /v1/public/status`
- `GET /v1/public/peers`
- `POST /v1/public/peers`
- `GET /v1/public/mempool`
- `POST /v1/public/transactions`
- `GET /v1/public/ledger?account=tcarrot1...`
- `GET /v1/public/snapshot`
- `GET /v1/public/checkpoint`
- `GET /v1/public/funding-policy`
- `POST /v1/public/validators/apply` when an admission verifier is configured.

`/v1/public/status` reports active protocol/validator transition state from replicated consensus state.

`POST /v1/public/transactions` is admission only. A transfer becomes authoritative only after it appears in a finalized block and executes through the replicated state machine.

## Sequential transaction queueing

A sender can queue at most 16 nonce positions beginning at the account's next consensus nonce. Proposals include only contiguous executable nonces, so a gap cannot make the entire proposal invalid.

A same-sender/same-nonce replacement must increase the fee by at least 12.5%, with a minimum one-atom increase. These are mempool rules only. Finalized consensus still requires the exact account nonce.

## TEST-CARROT settlement

For height greater than one, a proposal must include the prior block's finality-settlement operation. That operation chooses the exact valid vote set used for Proof-of-Play release and pending-fee distribution.

Settlement verification uses the validator set that was active at the rewarded historical height. A validator-set activation in the current block therefore cannot rewrite who was eligible to sign the prior block.

## Validator-set transition

A `validator-set-commitment` needs at least 144 blocks of notice.

At its exact activation height:

- the committed set becomes the committee-selection source;
- removed validators can no longer sign valid current votes;
- new validators can participate with their committed keys;
- prior blocks remain verified against their historical set;
- restart/catch-up reconstructs the same transition from finalized history.

If nodes disagree on the active set, stop and follow [Testnet incident response](testnet-incident-response.md). Do not edit node-local config to force a different consensus set.

## Protocol-upgrade transition and recovery

A `protocol-upgrade-commitment` needs at least 1,008 blocks of notice. v0.13 supports one controlled compatibility transition from protocol v3 to v4.

The activation procedure is:

1. finalize the upgrade commitment;
2. deploy software that explicitly supports the scheduled target before activation;
3. preserve a checkpoint before the activation height;
4. at activation, require blocks to use the committed protocol version;
5. incompatible software fails closed;
6. restart/replay from genesis or preserved chain data to verify the same activation boundary and post-upgrade checkpoint.

There is no admin rollback that rewrites finalized history. If the supported upgrade fails after finality, halt and use the incident runbook rather than inventing a local fork.

## Snapshots and checkpoints

Fetch the current observability snapshot:

```bash
curl http://127.0.0.1:9101/v1/public/snapshot
```

Fetch the compact checkpoint:

```bash
curl http://127.0.0.1:9101/v1/public/checkpoint
```

The checkpoint commits to genesis/network, finalized height/hash, state root, protocol version, validator-set hash, and CARROT policy hash. It is **not** trusted state. `VerifyCheckpoint` independently replays finalized blocks through that height before accepting it.

For incident evidence, capture checkpoints from multiple independently operated nodes and compare them before changing software or node data.

## Public test funding

`GET /v1/public/funding-policy` returns the hashed operational policy.

There is no automatic faucet, public mint, or treasury-spending endpoint. Operators who provide test funds must use ordinary signed transfers from existing TEST-CARROT balances under the documented test-funding cap.

## Peer diversity

Public discovery applies total, per-host, IPv4 `/24`, and IPv6 `/48` limits.

v0.13 also provides a server-side ASN/provider classifier boundary. Operators may connect a trusted local classification data source and configure ASN/provider caps. Never accept peer-self-declared ASN/provider labels as security evidence.

The classifier is research hardening, not proof of independent operators or geographic decentralization.

## Apple App Attest

Set the Apple Team ID, bundle ID and environment. To enable the concrete verifier, provide the trusted App Attest root certificate PEM as a file:

```text
BBWEALTH_APPLE_TEAM_ID
BBWEALTH_APPLE_BUNDLE_ID
BBWEALTH_APPLE_ATTEST_ENV
BBWEALTH_APPLE_ATTEST_ROOT_CA=/secure/path/app-attest-root.pem
```

If the validator/trust root is not configured, Apple enrollment remains fail-closed.

## Google Play Integrity

For a long-running server, configure the service-account JSON file so OAuth access tokens are minted and cached automatically:

```text
BBWEALTH_ANDROID_PACKAGE
BBWEALTH_ANDROID_CERTIFICATES
BBWEALTH_ANDROID_REQUIRE_STRONG_INTEGRITY=1
BBWEALTH_GOOGLE_SERVICE_ACCOUNT_FILE=/secure/path/service-account.json
```

`BBWEALTH_PLAY_INTEGRITY_ACCESS_TOKEN` remains a development fallback, not the preferred long-running credential source.

## Incident response

Use [Testnet incident response](testnet-incident-response.md) for validator/device key compromise, relay-node compromise, finality halt, transition failure, corrupted nodes, attestation outages, and eclipse incidents.

The default recovery rule is to preserve finalized history and reconstruct/verify state, not silently rewrite the chain.

## Internet exposure

Public discovery and transaction endpoints are adversarial-test surfaces, not a production DDoS guarantee. Request-size limits, concurrency limits, signed relay envelopes, replay protection, network-ID checks, validator signatures, IP-prefix diversity and optional trusted ASN/provider classification are layered defenses.

TEST-CARROT remains valueless. Do not advertise the v0.13 network as a production cryptocurrency or financial network.
