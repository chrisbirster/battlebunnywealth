# v0.15 live distributed testnet

v0.15 is the evidence milestone between local/adversarial protocol work and a v1.0 candidate. Deployment is intentionally **Fly.io only**.

The goal is to observe the same frozen consensus implementation on separately addressed Fly nodes across different real regions and preserve enough material for an external reviewer to reproduce what happened.

## Boundary

Repository code can make evidence collection deterministic and tamper-evident. It cannot manufacture independent cloud providers or an external security review. The Fly-only decision means provider-wide failures remain a correlated risk.

v0.15 therefore has two states:

1. **repository-ready** — Fly configs, deployment image, live sampler, evidence schema, CI/manual workflows, and review-freeze manifest are implemented and tested;
2. **evidence-complete** — the four Fly regional nodes have run the retained window, failure exercises have been recorded, and the resulting artifacts are frozen for review.

Do not call v0.15 complete until state 2 exists.

## Fly topology

Use one Fly app per evidence node so every node has its own stable `*.fly.dev` endpoint and persistent volume.

Canonical regional layout:

- `iad`
- `ord`
- `dfw`
- `lax`

All four entries declare `provider: "fly.io"`. Provider diversity is deliberately **not** a v0.15 gate anymore; regional separation is.

See `deploy/public-testnet/README.md` and `deploy/fly/` for bootstrap/deployment commands.

## Evidence artifact

`pop-live-evidence` repeatedly samples public status/checkpoints and commits to:

- exact repository SHA;
- network ID and genesis hash;
- CARROT policy hash;
- optional canonical network-map hash;
- operator IDs plus declared Fly provider/region/ASN metadata;
- timestamps;
- availability and latency samples;
- checkpoints;
- any same-height finalized-hash/state-root disagreement;
- deterministic SHA-256 evidence hash.

A same-height disagreement is preserved as a `checkpoint-mismatch` incident. It makes review readiness fail but does not destroy the evidence artifact.

## Operator inventory

Start from `deploy/public-testnet/operators.example.json` and replace the four placeholder Fly hostnames.

Provider/region labels are evidence metadata, not consensus inputs. The reviewer should corroborate region placement from retained Fly deployment/status output.

## Local invocation

```bash
go run ./cmd/pop-live-evidence \
  -operators ./operators.json \
  -repo-sha <exact-merged-dev-sha> \
  -samples 12 \
  -interval 5m \
  -min-providers 1 \
  -min-regions 4 \
  -min-samples-per-operator 12 \
  -min-window 30m \
  -gate \
  -out live-evidence.json
```

For the independent-review freeze, use a substantially longer window than this smoke profile and record the chosen duration/rationale.

## GitHub workflows

- `.github/workflows/deploy-fly-testnet.yml` manually deploys one existing Fly app from the selected merged `dev` revision. It requires `FLY_API_TOKEN`; node config/key remain Fly app secrets and the deployment artifact retains the raw Machine inventory with immutable image digest.
- `.github/workflows/capture-fly-topology.yml` captures all four Fly Machine inventories, validates the canonical `iad`/`ord`/`dfw`/`lax` layout, binds it to `GITHUB_SHA`, and emits a hash-committed `fly-topology.json`.
- `.github/workflows/live-testnet-evidence.yml` collects retained consensus evidence and defaults to one provider plus four Fly regions.

Neither workflow changes consensus voting power.

## Failure exercises

A serious retained window should include controlled tests on your own Fly apps:

- stop/restart one node and verify catch-up;
- lose one regional node while the other three remain reachable;
- rolling restart all four nodes;
- peer churn;
- delayed recovery/height lag;
- bounded partial partitions you control;
- post-recovery same-height checkpoint comparison.

Do not attack Fly infrastructure or third-party systems.

## Timeout tuning

Do not tune the consensus pacemaker from localhost timings. Retain real Fly regional latency distributions, then choose timeout/backoff margins from observed tails. Any executable timeout change requires a new SHA and evidence window.

## Completion criteria

v0.15 is evidence-complete only when all of the following are true:

- four separately addressable Fly node apps are exercised;
- those apps run in four distinct Fly regions;
- all retained evidence validates and hashes are recorded;
- no unresolved same-height checkpoint disagreement exists;
- regional node-loss/recovery evidence is retained;
- the exact deployed `dev` SHA, validated Fly topology, immutable Fly image digest(s), genesis hash, CARROT policy hash, and network-map provenance are recorded;
- the retained topology includes the raw four `flyctl machines list --json` responses and their byte hashes;
- the review package explicitly lists Fly-wide correlated failure as an untested/provider-concentration risk;
- `pop-review-freeze` validates that the Fly topology SHA and image digest set match the frozen target and produces a valid manifest;
- the package is ready for an independent consensus/security reviewer.

TEST-CARROT remains valueless. Economically valuable CARROT remains blocked on external security review and appropriate legal/tax/privacy/app-store review.
