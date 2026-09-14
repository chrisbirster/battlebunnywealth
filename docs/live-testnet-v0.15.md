# v0.15 live distributed testnet

v0.15 is the evidence milestone between the local/adversarial protocol work and a v1.0 candidate. Its purpose is to observe the same frozen consensus implementation on independent real Internet paths and preserve enough material for an external reviewer to reproduce what happened.

## Boundary

Repository code can make evidence collection deterministic and tamper-evident. It cannot manufacture independent providers, operators, routes, outages, or an external security review. Therefore v0.15 has two states:

1. **repository-ready** — deployment image, live sampler, evidence schema, CI/manual workflow, and review-freeze manifest are implemented and tested;
2. **evidence-complete** — real independent nodes have run the retained window, failure exercises have been recorded, and the resulting artifacts are frozen for review.

Do not call v0.15 complete until state 2 exists.

## Evidence artifact

`pop-live-evidence` reads an operator inventory and repeatedly samples public status/checkpoints. The resulting artifact commits to:

- exact repository SHA;
- network ID and genesis hash;
- CARROT policy hash;
- optional canonical network-map hash;
- operator IDs plus declared provider/region/ASN metadata;
- collection start/end timestamps;
- every availability and latency sample;
- every checkpoint observed;
- any same-height finalized-hash/state-root disagreement;
- a deterministic SHA-256 evidence hash.

A same-height disagreement is preserved as a `checkpoint-mismatch` incident. It makes review readiness fail but does not make the evidence file itself disappear.

## Operator inventory

Start from `deploy/public-testnet/operators.example.json`:

```json
[
  {"id":"operator-a","url":"https://a.example","provider":"provider-a","region":"region-a"},
  {"id":"operator-b","url":"https://b.example","provider":"provider-b","region":"region-b"}
]
```

Provider/region labels are declared metadata and must be independently corroborated during review. They are not consensus inputs.

## Local invocation

```bash
go run ./cmd/pop-live-evidence \
  -operators ./operators.json \
  -repo-sha <exact-merged-dev-sha> \
  -samples 12 \
  -interval 5m \
  -min-providers 2 \
  -min-regions 2 \
  -min-samples-per-operator 12 \
  -min-window 30m \
  -gate \
  -out live-evidence.json
```

For the independent-review freeze, use a substantially longer window than a smoke run. The repository deliberately does not pretend one universal duration proves safety; record the chosen duration and rationale for the reviewer.

## GitHub manual evidence workflow

`.github/workflows/live-testnet-evidence.yml` is `workflow_dispatch` only. Run it from the exact commit deployed to the operators. It stores `operators.json` and `live-evidence.json` as a retained Actions artifact.

The workflow does not deploy nodes and does not make provider metadata trustworthy. It simply gives the collection run an independent execution environment and binds the artifact to `GITHUB_SHA`.

## Failure exercises

A serious window should include controlled, authorized tests on infrastructure you own:

- observer/full-node process loss and catch-up;
- provider/path loss;
- rolling restart;
- peer churn;
- delayed recovery/height lag;
- sustained but bounded partial partitions;
- post-recovery same-height checkpoint comparison.

Record the start/end of each exercise as incident/operations notes alongside the evidence file. Do not perform disruptive testing against infrastructure you do not own or have permission to test.

## Timeout tuning

Do not tune the consensus pacemaker from localhost timings. Retain latency distributions from the live window, then choose timeout/backoff values with explicit margin for observed tails and provider failures. Any timeout-policy change after evidence collection changes the reviewed system and therefore requires a new frozen SHA/evidence window.

## Completion criteria

v0.15 is evidence-complete only when all of the following are true:

- at least two genuinely independent provider/network paths were exercised;
- all retained evidence files validate and their hashes are recorded;
- no unresolved same-height checkpoint disagreement exists;
- node/provider loss and recovery evidence is retained;
- the exact deployed `dev` SHA, image digest, genesis hash, CARROT policy hash, and network-map provenance are recorded;
- `pop-review-freeze` produces a valid manifest over the retained artifacts;
- the freeze package is ready to hand to an independent consensus/security reviewer.

TEST-CARROT remains valueless. Economically valuable CARROT is still blocked on external security and appropriate legal/tax/privacy/app-store review.
