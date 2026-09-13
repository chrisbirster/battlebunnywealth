# v0.9 testnet operations

No paid multi-server deployment is required for development. The deterministic cluster harness runs five independent consensus engines on one machine, each with separate logical state.

## Local security gates

```bash
go test ./...
go run ./cmd/pop-cluster -blocks 20
go run ./cmd/pop-bootstrap-sim -phones 100 -gate
```

The cluster smoke test exercises the four-validator 3-of-4 path across five independent engines. The bootstrap gate compares unsafe immediate activation with the 30-day, one-per-seven-days activation policy for a 100-real-phone attacker cohort.

## Node key

Generate a full-node Ed25519 identity:

```bash
go run ./cmd/pop-node -keygen ./data/testnet/node-a/node.key
```

The private key file is written with mode `0600` and is deliberately separate from JSON chain state. Record the printed node ID and public key in the permissioned peer configuration.

## Run a node

A node config contains `listen`, a complete hashed `genesis`, a permissioned peer array (`nodeId`, `url`, `publicKey`), and `syncIntervalSeconds`.

```bash
go run ./cmd/pop-node \
  -config ./testnet/node-a.json \
  -data ./data/testnet/node-a
```

Consensus-facing endpoints include:

- `GET /v1/node/status`
- `GET /v1/node/genesis`
- `GET /v1/node/blocks?from=HEIGHT`
- `GET /v1/node/metrics`
- `GET /v1/committee/draft?validator=ID&round=0`
- `POST /v1/committee/proposals`
- `POST /v1/committee/votes`
- signed permissioned relay endpoints under `/v1/peer/*`.

Phone clients submit their own signed proposal/vote. A receiving full node verifies the committee signature and then relays the same signed object to its permissioned peers using the node's separate relay identity.

## Apple App Attest

Set the existing Apple Team ID, bundle ID and environment. To enable the concrete verifier, provide the current trusted App Attest root certificate PEM as a file:

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

v0.9 peer ingress is permissioned. Do not treat it as a public production P2P network. Request-size limits, per-IP concurrency limits, signed peer envelopes, replay protection, network-ID checks, and validator-signature checks are defense layers, not a DDoS guarantee.
