# Public-testnet operator deployment

This directory contains provider-neutral examples for v0.15. A full node is infrastructure only: running more node processes does not create validator votes.

## 1. Freeze the code revision

All operators in one evidence window should run the same merged `dev` commit and the same genesis/CARROT policy commitments. Record the image digest as well as the Git SHA.

```bash
docker build -f Dockerfile.pop-node -t battlebunnywealth-pop-node:<sha> .
docker image inspect battlebunnywealth-pop-node:<sha> --format '{{.Id}}'
```

Do not use an unpinned `latest` image during a review evidence window.

## 2. Generate an independent node key

Each operator owns a different full-node Ed25519 key and persistent data volume.

```bash
docker run --rm -v "$PWD/data:/data" battlebunnywealth-pop-node:<sha> \
  -keygen /data/node.key
```

The output contains the node ID and public key. Exchange only those public values with the other operators. Never copy a node private key between providers.

## 3. Build `node.json`

Every node config contains the exact same `genesis` value and a peer list containing the other permissioned relay nodes. The local listen address should normally be `:9101` inside the container.

Each peer entry needs:

```json
{
  "nodeId": "peer node id",
  "publicKey": "peer Ed25519 public key",
  "url": "https://peer.example.net"
}
```

The peer list authenticates relay traffic. It does not grant committee voting power.

## 4. Persist state

Mount `/data` on durable provider storage. The node stores its private node key, genesis binding, finalized `blocks.ndjson`, and in-progress round state there. Losing the volume is a node-recovery event and should be recorded in the evidence window.

`docker-compose.example.yml` shows the expected mounts. The node container also accepts the optional canonical network map at `/config/network-map.json`.

## 5. Expose HTTPS

Terminate TLS in the provider load balancer/reverse proxy and route public HTTPS to container port `9101`. Evidence tooling uses:

- `/v1/public/status`
- `/v1/public/checkpoint`
- `/v1/node/genesis`
- `/v1/node/blocks`

Do not expose provider control-plane credentials or private node/validator keys through the HTTP service.

## 6. Operator independence

For a useful v0.15 window, prefer different provider/network failure domains and independent persistent volumes. Two regions on one provider are useful but do not count as two providers. Provider/region labels in `operators.json` are evidence metadata, not cryptographic truth; reviewers should independently verify them.

## 7. Failure exercises

During a retained evidence window, record intentionally induced events rather than hiding them:

- stop one observer/full node and verify catch-up after restart;
- lose one provider path temporarily;
- restart nodes one at a time;
- create peer churn without changing validator voting power;
- delay one node long enough to produce measurable height lag;
- compare same-height checkpoints after recovery.

Never induce an outage against third-party systems you do not own or have permission to test.

TEST-CARROT remains valueless throughout v0.15.
