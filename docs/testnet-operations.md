# v0.12 testnet operations

No paid multi-server deployment is required for local development. The deterministic cluster harness runs independent consensus engines with separate logical state, while `pop-node` is the long-running node binary.

Protocol v3 is a testnet reset from protocol v2 because TEST-CARROT is now replicated consensus state. Do not reuse a protocol-v2 data directory with a protocol-v3 genesis.

## Local security gates

```bash
go test ./...
go vet ./...
go run ./cmd/pop-cluster -blocks 20
go run ./cmd/pop-bootstrap-sim -phones 100 -gate
go run ./cmd/carrot-spec -check
go run ./cmd/pop-public-smoke -attackers 100 -gate
```

The cluster smoke test exercises the four-validator 3-of-4 path. The bootstrap gate compares the activation policy against a 100-real-phone attacker cohort. The public smoke gate attacks peer discovery, transaction admission, and validator-candidate admission.

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

1. loads and validates genesis and the CARROT policy hash;
2. constructs the deterministic TEST-CARROT consensus state;
3. replays every local finalized block through the state machine;
4. rejects corrupt or incompatible history;
5. begins peer catch-up and HTTP service.

## Consensus endpoints

- `GET /v1/node/status`
- `GET /v1/node/genesis`
- `GET /v1/node/blocks?from=HEIGHT`
- `GET /v1/node/metrics`
- `GET /v1/committee/draft?validator=ID&round=0`
- `POST /v1/committee/proposals`
- `POST /v1/committee/votes`
- signed relay endpoints under `/v1/peer/*`.

Phone clients submit their own signed proposal/vote. A receiving full node verifies the committee signature and may relay the same signed object using the node's separate relay identity.

## Public-testnet endpoints

The public wrapper adds:

- `GET /v1/public/status`
- `GET /v1/public/peers`
- `POST /v1/public/peers`
- `GET /v1/public/mempool`
- `POST /v1/public/transactions`
- `GET /v1/public/ledger?account=tcarrot1...`
- `GET /v1/public/funding-policy`
- `POST /v1/public/validators/apply` when an admission verifier is configured.

`POST /v1/public/transactions` is admission only. A transfer becomes authoritative only after it appears in a finalized block and executes through the replicated state machine.

## TEST-CARROT settlement

For height greater than one, a proposal must include the prior block's finality-settlement operation. That operation chooses the exact valid vote set used for Proof-of-Play release and pending-fee distribution.

If a node cannot validate the settlement, the proposal is invalid.

## Public test funding

`GET /v1/public/funding-policy` returns the hashed operational policy.

v0.12 does not run an automatic faucet and does not enable treasury spending. Operators who choose to provide test funds must use ordinary signed transfers from an operator-owned TEST-CARROT wallet containing existing balances. The default suggested cap is 100 TEST-CARROT per address per 24-hour window.

There is no public mint endpoint.

## Transition commitments

Finalized blocks may contain:

- `validator-set-commitment` plans with at least 144 blocks of notice;
- `protocol-upgrade-commitment` plans with at least 1,008 blocks of notice.

These plans are quorum-finalized and survive recovery replay, but protocol v3 does not automatically activate a new committee set or executable protocol version. That controlled activation path is v0.13 work.

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

## Internet exposure

Public discovery and transaction endpoints are adversarial-test surfaces, not a production DDoS guarantee. Request-size limits, per-IP concurrency limits, signed relay envelopes, replay protection, network-ID checks, validator signatures, per-host limits, and IP-prefix diversity limits are layered defenses.

TEST-CARROT remains valueless. Do not advertise the v0.12 network as a production cryptocurrency or production financial network.
