# Distributed public-testnet evidence

v0.14 adds the tooling needed to collect comparable evidence from independently operated public-testnet nodes. This repository does **not** claim that a geographically or provider-distributed live test has already occurred.

That distinction is intentional: CI can prove deterministic behavior on isolated logical nodes, but it cannot substitute for independent networks, operators, routes, providers, outages, or real Internet latency.

## Evidence targets

A live test should eventually include independently operated nodes across multiple failure domains where practical:

- different physical hosts;
- different operators;
- more than one network/provider/ASN;
- more than one geographic region;
- independent process/data directories/node keys;
- the same network ID/genesis/protocol policy.

The goal is not to maximize server count. Observer/full-node count creates zero consensus votes.

## Checkpoint comparison

Each public node exposes:

```text
GET /v1/public/checkpoint
```

A checkpoint commits to network, genesis, finalized height/hash, state root, active protocol version, validator-set hash, and CARROT policy hash.

Compare two or more operators:

```bash
go run ./cmd/pop-checkpoint-compare \
  -url https://operator-a.example \
  -url https://operator-b.example
```

Comparison is strict and requires the same height. A disagreement in finalized hash or state root is a consensus incident, not a cosmetic monitoring difference.

The checkpoint is not trusted by itself. `VerifyCheckpoint` reconstructs the chain from genesis and finalized history to verify the claimed commitment.

## Evidence collector

`pop-network-evidence` records availability and request latency together with public status/checkpoint material:

```bash
go run ./cmd/pop-network-evidence \
  -url https://operator-a.example \
  -url https://operator-b.example \
  -url https://operator-c.example \
  -min-available 2 \
  -gate > evidence.json
```

The report includes:

- collection timestamp;
- available/unavailable node count;
- per-node status latency;
- per-node checkpoint latency;
- public status response;
- checkpoint response;
- same-height checkpoint comparison when at least two nodes are comparable.

Different node heights are treated as lag, not silently compared as equal consensus states.

## Reproducible ASN/provider metadata

Self-declared provider/ASN labels from a peer are untrusted. v0.14 therefore supports a server-side, versioned CIDR map:

```json
{
  "version": 1,
  "entries": [
    {"cidr": "203.0.113.0/24", "asn": 64500, "provider": "example-a"},
    {"cidr": "2001:db8::/32", "asn": 64501, "provider": "example-b"}
  ],
  "hash": "..."
}
```

The loader canonicalizes CIDRs, rejects duplicates, performs longest-prefix matching, and verifies the document hash. A node may load the map with:

```bash
go run ./cmd/pop-node \
  -config ./testnet/node-a.json \
  -data ./data/testnet/node-a \
  -network-map ./testnet/network-map.json \
  -max-peers-per-asn 32 \
  -max-peers-per-provider 64
```

The exact metadata source is operational policy and must be recorded alongside the evidence artifact. A stale or incomplete map can produce false negatives or false positives, so the map hash should always accompany test results.

## What a live evidence record should contain

For each test window, retain:

- repository commit SHA;
- genesis hash;
- CARROT policy hash;
- network-map hash and provenance;
- operator/node IDs;
- broad provider/ASN/region categories without exposing unnecessary personal location data;
- start/end timestamps;
- checkpoint samples;
- latency/availability reports;
- restart/catch-up events;
- observed round changes;
- any equivocation evidence;
- incidents and remediation notes.

## Current v0.14 evidence

The repository provides deterministic local evidence through:

```bash
go test ./...
go run ./cmd/pop-soak -blocks 1100 -gate
go run ./cmd/pop-round-chaos -blocks 240 -seed 42 -gate
```

These exercise independent logical state machines, transition activation, crash/replay, temporary partitions, round changes, reordered/duplicate messages, timestamp skew, checkpoint convergence, and supply conservation.

They are **not** evidence of multiple real Internet providers or geographic regions.

## External review gate

Before economically valuable CARROT or a production-security claim:

1. freeze a specific commit;
2. run a multi-operator public-testnet evidence window;
3. retain the artifacts described above;
4. resolve any finalized-hash/state-root disagreement;
5. provide the frozen commit, protocol docs, adversarial tests, known limitations, and evidence artifacts to an independent security reviewer;
6. separately complete appropriate legal/tax/privacy/app-store review.

TEST-CARROT remains valueless while those gates are open.
