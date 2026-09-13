# v0.14 testnet operations

No paid multi-server deployment is required for local development. The deterministic cluster, transition-soak, and round-chaos harnesses run independent consensus engines with separate logical state. `pop-node` is the long-running node binary.

Genesis remains protocol v3. The binary may activate only explicitly implemented protocol versions; unsupported scheduled versions fail closed.

## Local security and reliability gates

```bash
go test ./...
go vet ./...
go run ./cmd/pop-cluster -blocks 20
go run ./cmd/pop-bootstrap-sim -phones 100 -gate
go run ./cmd/carrot-spec -check
go run ./cmd/pop-public-smoke -attackers 100 -gate
go run ./cmd/pop-soak -blocks 1100 -gate
go run ./cmd/pop-round-chaos -blocks 240 -seed 42 -gate
```

`task testnet:soak` runs the v0.13 transition soak. `task testnet:round-chaos` runs the v0.14 multi-round disorder/restart gate.

## Node key

Generate a full-node Ed25519 relay identity:

```bash
go run ./cmd/pop-node -keygen ./data/testnet/node-a/node.key
```

The private key is mode `0600` and is separate from player/device validator keys. Running a node creates no committee vote.

## Run a node

```bash
go run ./cmd/pop-node \
  -config ./testnet/node-a.json \
  -data ./data/testnet/node-a
```

At startup the node validates genesis/CARROT policy, rebuilds deterministic state from finalized blocks, reconstructs validator/protocol transitions, restores only valid in-progress round state for the next height, and then begins peer catch-up.

Peer catch-up independently verifies imported finality and persists accepted blocks locally. A later restart therefore does not depend on the same peer still being online.

## Optional reproducible network metadata

Public peer diversity may use a hash-pinned local CIDR→ASN/provider map:

```bash
go run ./cmd/pop-node \
  -config ./testnet/node-a.json \
  -data ./data/testnet/node-a \
  -network-map ./testnet/network-map.json \
  -max-peers-per-asn 32 \
  -max-peers-per-provider 64
```

The map is canonicalized, hash-verified, and longest-prefix matched. Record the map hash and provenance with every live evidence run. Peer-supplied ASN/provider labels are never trusted.

## Consensus endpoints

- `GET /v1/node/status`
- `GET /v1/node/genesis`
- `GET /v1/node/blocks?from=HEIGHT`
- `GET /v1/node/metrics`
- `GET /v1/committee/draft?validator=ID&round=R`
- `GET /v1/committee/round-certificate?round=R`
- `POST /v1/committee/proposals`
- `POST /v1/committee/votes`
- `POST /v1/committee/round-changes`
- signed relay endpoints under `/v1/peer/*`, including round changes.

The networking/client layer decides when its current round has timed out. A local timeout does not advance consensus by itself. A quorum of validator-signed round-change messages is required.

## Round-change recovery

In-progress state lives in `round-state.json` and includes current round, validator value locks, and verified round-certificate history.

After a crash:

1. finalized `blocks.ndjson` is replayed first;
2. a round snapshot is accepted only when it targets the exact next height;
3. stale lower-height round state is deleted because finalized history wins;
4. future-height round state fails startup;
5. a restored validator remains unable to vote against its persisted lock.

See [Round-change protocol](round-change-protocol.md).

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

A TEST-CARROT transfer is authoritative only after finalization and deterministic state execution.

## Checkpoint comparison

Compare independent operators at the same height:

```bash
go run ./cmd/pop-checkpoint-compare \
  -url https://operator-a.example \
  -url https://operator-b.example
```

A finalized-hash or state-root disagreement is an incident. Do not choose whichever node looks newest and continue.

## Distributed evidence collection

Collect public status, availability, latency, and checkpoint evidence:

```bash
go run ./cmd/pop-network-evidence \
  -url https://operator-a.example \
  -url https://operator-b.example \
  -url https://operator-c.example \
  -min-available 2 \
  -gate > evidence.json
```

The repository supplies this tooling but does not claim that a real multi-provider/geographic run has already happened. See [Distributed testnet evidence](distributed-testnet-evidence.md).

## Sequential transaction queueing

A sender can queue at most 16 nonce positions beginning at the account's next consensus nonce. Proposals include only a contiguous executable prefix.

Same-nonce replacement requires at least a 12.5% fee increase with a minimum one-atom bump. Consensus itself still requires exact sequential nonces.

## TEST-CARROT settlement

For height greater than one, a proposal includes the prior block's finality-settlement operation. Settlement verification uses the validator set active at the rewarded historical height.

TEST-CARROT remains valueless. There is no public mint, automatic faucet, or treasury-spending endpoint.

## Validator/protocol transitions

Validator-set commitments need at least 144 blocks of notice. At activation, the committed set becomes authoritative for current voting while old blocks retain their historical set.

Protocol-upgrade commitments need at least 1,008 blocks of notice. Nodes must explicitly implement the target version before activation. No chain message can download or install executable code.

## Equivocation evidence

Conflicting signed proposals or votes can be packaged as independently verifiable evidence. Preserve the original messages and checkpoint context.

Evidence is non-punitive in v0.14: there is no automatic slashing, balance confiscation, validator deletion, or permanent ban.

## Apple App Attest

```text
BBWEALTH_APPLE_TEAM_ID
BBWEALTH_APPLE_BUNDLE_ID
BBWEALTH_APPLE_ATTEST_ENV
BBWEALTH_APPLE_ATTEST_ROOT_CA=/secure/path/app-attest-root.pem
```

If the trust root/verifier is not configured, validator eligibility remains fail-closed.

## Google Play Integrity

```text
BBWEALTH_ANDROID_PACKAGE
BBWEALTH_ANDROID_CERTIFICATES
BBWEALTH_ANDROID_REQUIRE_STRONG_INTEGRITY=1
BBWEALTH_GOOGLE_SERVICE_ACCOUNT_FILE=/secure/path/service-account.json
```

The static access-token environment variable remains development fallback only.

## Incident response

Use [Testnet incident response](testnet-incident-response.md) for validator/device compromise, relay compromise, finality halt, checkpoint disagreement, transition failure, corrupted nodes, attestation outage, or eclipse symptoms.

Preserve finalized history and signed evidence. Reconstruct and verify state rather than editing node-local consensus data.

## Internet exposure

Public endpoints are adversarial-test surfaces, not a production DDoS guarantee. Request limits, signatures, replay protection, network binding, peer diversity controls, and round-change certificates are layered defenses.

TEST-CARROT remains valueless. Do not describe the v0.14 testnet as a production cryptocurrency or financial network.
