# Public testnet incident response

This runbook applies to the valueless Battle Bunny Proof-of-Play public testnet. It is not a production CARROT recovery policy. The default rule is **do not rewrite finalized history**. Recovery should restore honest verification from genesis/checkpoints and preserve auditable evidence.

## Severity

- **SEV-1:** safety failure, conflicting finalized blocks, unauthorized validator-set/protocol activation, private validator key compromise with active signing, or fixed-supply/state-root divergence.
- **SEV-2:** prolonged finality halt, widespread peer eclipse, corrupted node stores, attestation-provider outage that blocks validator participation, or upgrade activation failure without conflicting finality.
- **SEV-3:** single-node outage, public peer churn, mempool spam, degraded discovery diversity, or isolated client/API failure.

## First actions

1. Preserve the current finalized height/hash/state root and public checkpoint from at least two independently operated nodes.
2. Stop automation that could submit new transition commitments or rotate keys.
3. Preserve logs, node data directories, peer metadata, signed proposals/votes, and the exact binary/commit hash.
4. Compare `/v1/node/status`, `/v1/public/status`, `/v1/public/snapshot`, and `/v1/public/checkpoint` across independent nodes.
5. Do not delete chain data or force a replacement genesis while evidence is being collected.

## Validator signing-key compromise

A compromised committee/device key can sign only as that validator; it cannot become additional validators or alter the active set by itself.

- Revoke the affected device in the account/attestation system.
- Stop the validator client and preserve the compromised public key, device ID, and first-known compromise time.
- Do not accept an out-of-band replacement key as consensus truth.
- Prepare a normal validator-set commitment that removes the compromised key and, if appropriate, introduces an independently attested replacement. The normal notice/activation rules still apply unless a future formally specified emergency mechanism exists.
- Monitor for equivocation and retain both signed votes as evidence.
- If enough validators are compromised to violate the BFT assumption, halt economic plans and treat the testnet as unsafe. Do not claim the chain is secure merely because software continues producing blocks.

## Full-node relay-key compromise

A node relay key has **zero consensus voting power**.

- Remove the node key from permissioned relay configuration where applicable.
- Generate a fresh node key and reannounce the public observer/full node.
- Treat spam, censorship, or malformed relay traffic as a networking incident, not a validator-set change.
- Verify the node from genesis/checkpoint before returning it to service.

## TEST-CARROT wallet-key compromise

TEST-CARROT is explicitly valueless.

- Stop using the wallet.
- Do not introduce an admin balance rewrite or mint to compensate it.
- A replacement wallet receives future funds only through normal consensus-valid transfers/rewards/funding policy.
- Record the compromised address so test operators can distinguish expected attacker transfers from consensus faults.

## Finality halt

A halt is preferable to unsafe finality.

1. Determine the active validator set for the stalled height, not merely the current genesis list.
2. Check whether quorum is online and whether validators agree on parent hash, state root, protocol version, and committee hash.
3. Check for a scheduled validator-set or protocol transition at the stalled height.
4. Restart individual nodes from preserved data or reconstruct them from genesis + finalized blocks/checkpoint.
5. Do not lower quorum or bypass signature validation to restore liveness.

## Protocol-upgrade activation failure

v0.13 supports one controlled compatibility upgrade step beyond protocol v3. A node that does not support the scheduled target must fail closed rather than reinterpret blocks.

- Confirm the finalized upgrade commitment, activation height, minimum software version, and plan hash.
- Nodes on incompatible software should remain stopped at the pre-activation head.
- Upgrade the binary, replay from genesis/finalized history, and compare the resulting checkpoint with independent nodes.
- There is no automatic rollback that erases finalized v4 blocks. A future rollback/recovery mechanism must itself be specified and consensus-authorized rather than improvised operationally.

## Validator-set activation failure

- Compare the finalized validator-set commitment and plan hash.
- Verify the set used at `activationHeight-1` and `activationHeight` separately.
- Removed validators must remain valid for historical certificates but must not vote at or after activation.
- Replacement validators must possess the committed keys; no node-local config may silently substitute a key.
- Reconstruct a clean node from genesis and finalized blocks to distinguish local corruption from deterministic protocol disagreement.

## Corrupted or stale node

- Quarantine the data directory; do not overwrite it before diagnosis.
- Start a clean node from the same genesis.
- Import finalized blocks in order and verify every finality certificate/state transition.
- Compare the resulting public checkpoint/state root with at least two peers.
- If replay diverges deterministically across clean nodes, escalate to SEV-1.

## Attestation-provider outage

Apple/Google availability affects candidate/validator eligibility, not the validity of already-finalized blocks.

- Existing active validators continue under the protocol's current eligibility rules unless an explicit state transition says otherwise.
- Do not fail open by converting unattested devices into eligible validators.
- Pause new validator activation if eligibility cannot be verified reliably.
- Ordinary game access can remain separate from consensus eligibility.

## Eclipse / peer-diversity incident

- Inspect per-host, IPv4 `/24`, IPv6 `/48`, ASN, and provider distribution.
- Prefer independently sourced outbound peers and known-good checkpoints.
- Rotate connections without changing consensus voting power.
- Remember that 100 hostile full nodes still have zero votes unless they also control selected validator keys.

## Evidence and closure

An incident is closed only after:

- affected keys/nodes are identified;
- a clean genesis replay reaches the accepted finalized head and state root;
- TEST-CARROT fixed-supply conservation passes;
- active/historical validator sets and protocol version agree across independent nodes;
- a post-incident checkpoint is recorded;
- a regression test is added for any protocol/software defect discovered.

Any event that demonstrates conflicting valid finality, supply creation, or unauthorized transition activation is a release blocker for economically valuable CARROT.
