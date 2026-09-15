# Public-testnet operator deployment — Fly.io only

Battle Bunny Wealth v0.15 deploys public-testnet full nodes only on Fly.io. A full node is infrastructure only: running more Fly Machines or Fly apps does not create validator votes.

The evidence topology is **one Fly app per node**. Do not place several evidence nodes behind one Fly app hostname: each retained operator entry must address one independently deployed node endpoint.

Canonical test regions:

- `iad` — Ashburn, Virginia
- `ord` — Chicago, Illinois
- `dfw` — Dallas, Texas
- `lax` — Los Angeles, California

This gives four regional failure domains while intentionally accepting Fly.io as a single-provider concentration risk.

## 1. Freeze the code revision

All nodes in an evidence window run the same merged `dev` SHA and the same genesis/CARROT commitments. Use `Dockerfile.pop-node`; never use an unpinned `latest` image for a review window.

The final review freeze records the exact Fly-deployed image digest in addition to the Git SHA.

## 2. Generate four independent node keys locally

Never commit private node keys. `.fly-private/` is ignored by Git.

```bash
mkdir -p .fly-private

go run ./cmd/pop-node -keygen .fly-private/iad.node.key
go run ./cmd/pop-node -keygen .fly-private/ord.node.key
go run ./cmd/pop-node -keygen .fly-private/dfw.node.key
go run ./cmd/pop-node -keygen .fly-private/lax.node.key
```

Record each printed node ID and public key. Those public values are required when building the four node configs.

## 3. Create one Fly app and volume per node

Choose globally unique app names:

```bash
export APP_IAD=bbwealth-pop-iad
export APP_ORD=bbwealth-pop-ord
export APP_DFW=bbwealth-pop-dfw
export APP_LAX=bbwealth-pop-lax

fly apps create "$APP_IAD"
fly apps create "$APP_ORD"
fly apps create "$APP_DFW"
fly apps create "$APP_LAX"

fly volumes create pop_data -a "$APP_IAD" -r iad --size 1
fly volumes create pop_data -a "$APP_ORD" -r ord --size 1
fly volumes create pop_data -a "$APP_DFW" -r dfw --size 1
fly volumes create pop_data -a "$APP_LAX" -r lax --size 1
```

Each volume is independent and region-local. Consensus/catch-up replicates finalized history; Fly Volumes do not replicate it for us.

## 4. Generate one canonical genesis and all four node configs

Do not hand-edit genesis hashes or peer lists. Prepare two local public manifests under `.fly-private/`:

- `validators.json` — exactly four genesis validator/device public identities;
- `nodes.json` — exactly four full-node public identities and Fly endpoints.

Example node entry:

```json
{
  "name": "iad",
  "nodeId": "<node id printed by pop-node -keygen>",
  "publicKey": "<public key printed by pop-node -keygen>",
  "url": "https://bbwealth-pop-iad.fly.dev"
}
```

Then generate all configs in one operation:

```bash
go run ./cmd/pop-bootstrap-config \
  -validators .fly-private/validators.json \
  -nodes .fly-private/nodes.json \
  -network-id bbw-pop-fly-testnet-v1 \
  -out .fly-private/bootstrap
```

The tool validates node IDs against their Ed25519 public keys, requires bare HTTPS URLs, creates one canonical CARROT-committed genesis, and gives every node the other three peers. It writes:

```text
.fly-private/bootstrap/genesis.json
.fly-private/bootstrap/iad.node.json
.fly-private/bootstrap/ord.node.json
.fly-private/bootstrap/dfw.node.json
.fly-private/bootstrap/lax.node.json
.fly-private/bootstrap/bootstrap-summary.json
```

All four `*.node.json` files contain the exact same genesis hash. Full-node identities remain distinct from validator/device identities; running more Fly apps still creates zero validator votes.

## 5. Store node config and node key as Fly secrets

`fly.toml` mounts both secrets as files. The secret values must be base64 encoded. Stage them before first deployment:

```bash
fly secrets set --stage -a "$APP_IAD" \
  POP_NODE_CONFIG="$(base64 < .fly-private/bootstrap/iad.node.json | tr -d '\n')" \
  POP_NODE_KEY="$(base64 < .fly-private/iad.node.key | tr -d '\n')"
```

Repeat for `ord`, `dfw`, and `lax` using their own config/key files. Never reuse one node private key across apps.

## 6. Deploy

```bash
fly deploy . --remote-only --ha=false -a "$APP_IAD" -c deploy/fly/fly.iad.toml --dockerfile Dockerfile.pop-node
fly deploy . --remote-only --ha=false -a "$APP_ORD" -c deploy/fly/fly.ord.toml --dockerfile Dockerfile.pop-node
fly deploy . --remote-only --ha=false -a "$APP_DFW" -c deploy/fly/fly.dfw.toml --dockerfile Dockerfile.pop-node
fly deploy . --remote-only --ha=false -a "$APP_LAX" -c deploy/fly/fly.lax.toml --dockerfile Dockerfile.pop-node
```

`--ha=false` is intentional: each Fly app represents one evidence node with one region-local volume. Cross-node resilience comes from the four independent apps, not from Fly creating a hidden spare behind one hostname.

You can also run the manual **Deploy Fly Testnet Node** GitHub Action from the merged `dev` revision. It requires the repository secret `FLY_API_TOKEN`; node config/key secrets remain stored on the target Fly app.

## 7. Verify each node

```bash
fly status -a "$APP_IAD"
fly checks list -a "$APP_IAD"
curl "https://${APP_IAD}.fly.dev/v1/public/status"
curl "https://${APP_IAD}.fly.dev/v1/public/checkpoint"
```

Repeat for all four nodes. Copy `deploy/public-testnet/operators.example.json`, replace hostnames, and use that inventory for the retained evidence workflow.

## 8. Evidence policy

The Fly-only evidence gate requires:

- four separately addressable Fly apps;
- four distinct Fly regions;
- one provider (`fly.io`);
- independently persisted volumes;
- the exact same merged `dev` SHA/genesis/CARROT policy;
- no unresolved same-height checkpoint disagreement.

This does **not** prove provider independence. A Fly-wide control-plane, backbone, or platform failure remains a correlated risk and must be listed in the external-review package.

## 9. Failure exercises

During the retained window, intentionally record authorized tests such as:

- stop/restart one Fly app and verify catch-up;
- stop one regional Machine and verify the other three continue;
- rolling restart all four apps one at a time;
- delay one node long enough to create measurable height lag;
- temporarily isolate a peer path you control;
- compare same-height checkpoints after recovery.

Do not perform disruptive testing against Fly.io infrastructure itself or anything you do not own/control.

TEST-CARROT remains valueless throughout v0.15.
