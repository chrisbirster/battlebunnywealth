# v0.15 independent-review freeze

The review freeze is a reproducibility boundary, not an audit result.

## Create the manifest

After the live evidence files are finalized and the exact deployed commit and immutable image digest are known:

```bash
go run ./cmd/pop-review-freeze \
  -repo-sha <merged-dev-sha> \
  -image-digest sha256:<64-hex-digest> \
  -evidence ./evidence/window-a.json \
  -evidence ./evidence/window-b.json \
  -out ./evidence/review-freeze.json
```

Repeat `-image-digest` when retained operators used more than one image build for the same exact source target. The manifest sorts/deduplicates image digests and rejects malformed SHA-256 digests.

The command rejects evidence created from a different repository SHA, network, genesis, or CARROT policy. It records each file's byte-level SHA-256 plus the evidence artifact's own deterministic hash.

## Why both source SHA and image digest matter

The repository SHA pins source, but a container build can also depend on base-image bytes and toolchain layers. The immutable deployed image digest therefore belongs in the review package alongside the source SHA. An unpinned mutable image tag such as `latest` is not an acceptable review identifier.

## Freeze package

Provide the reviewer:

- `review-freeze.json`;
- every evidence artifact named by the manifest;
- exact Git commit and container image digest(s);
- genesis configuration;
- CARROT policy hash/specification;
- network-map file, hash, and provenance if used;
- `docs/round-change-protocol.md`;
- `docs/testnet-consensus.md`;
- v0.12/v0.13/v0.14 security-review handoffs;
- v0.15 live-testnet runbook and incident notes;
- CI run identifiers for the frozen commit;
- known limitations and unresolved research questions.

## Immutability rule

Any consensus, state-transition, attestation, validator-selection, wallet, peer-authentication, timeout/pacemaker, or CARROT-accounting code change after the freeze creates a new review target. Generate a new manifest and, where the change can affect network behavior, collect a new live evidence window.

Documentation-only corrections may be handled separately if the reviewer agrees they do not alter the executable target, but the reviewed source SHA must remain explicit.

## What this does not prove

A hash manifest proves which bytes and commitments were presented. It does not prove the protocol is secure, provider labels are truthful, a participant is a unique human, or CARROT is legally ready for economic use.
