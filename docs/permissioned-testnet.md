# v0.9 permissioned Proof-of-Play testnet

v0.9 turns the Proof-of-Play research model into a real replicated, non-economic testnet. It is intentionally small enough to start with four phones and one development computer.

## Bootstrap topology

The genesis validator set contains four known participant devices. While the active committee is 15 members or fewer, finality uses a 3/4 quorum; therefore the four-device bootstrap requires three distinct valid committee signatures. Two unavailable genesis devices halt finality rather than lowering the threshold.

A validator device and a full node are different identities. Participant devices sign committee proposals/votes with their enrolled device keys. Full nodes store state, verify signatures, relay messages, synchronize blocks, and serve clients. Running 100 extra full nodes creates zero additional committee votes.

The mature committee target remains 64, but v0.9 does not require 64 phones or 64 servers. When fewer than the target are active, every active validator is selected. Above the target, authority influences deterministic committee-selection probability. Once selected, every committee member has exactly one vote.

## Validator activation

A newly attested device is a validator candidate, not an active validator. The v0.9 default research policy is minimum authority 100, maturation 30 days, a 7-day activation window, and at most one new activation per window. This rate limit makes sudden phone-farm capture impossible during bootstrap while leaving long-horizon capture as an explicit research risk.

## Device eligibility

Android candidates require the configured package/signing certificate, `PLAY_RECOGNIZED`, `LICENSED`, and strong hardware-backed Play Integrity for `productionEligible`.

Apple candidates require verified App Attest evidence tied to the configured Team ID and bundle ID, a valid certificate chain to the configured trust root, nonce/AAGUID/credential checks, and the P-256 public key certified at enrollment.

After enrollment, the server persists that certified App Attest public key and issues one-time, purpose-bound assertion challenges. The iOS client hashes the canonical challenge payload and calls `generateAssertion`. The server verifies the assertion signature, App ID/RP-ID hash, exact challenge, and a strictly increasing assertion counter. Used challenges and steady/decreasing counters are rejected as replays.

Assertion endpoints are `POST /api/v1/proof-of-play/attestations/assertions/begin` and `POST /api/v1/proof-of-play/attestations/assertions/complete`.

Neither platform signal proves that one device equals one independent human.

## Non-economic boundary

v0.9 has no economically valuable CARROT, token rewards, staking, or purchases that affect consensus.

## Security properties under test

The testnet rejects wrong-network messages, invalid previous hashes, invalid state roots, unknown validators, invalid signatures, duplicate votes, equivocation, unknown permissioned peers, replayed peer envelopes, oversized bodies, and excessive concurrent requests from one source address. Stored chains are reverified on restart, and stale nodes only catch up by importing finalized blocks that independently pass certificate verification.

The attestation layer also rejects reused enrollment/assertion challenges, App Attest signatures that do not verify with the enrollment-certified key, App-ID mismatches, and non-increasing assertion counters.

These controls do not make the system production-secure. v0.9 remains a permissioned research network.
