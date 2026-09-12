# v0.9 permissioned Proof-of-Play testnet

v0.9 turns the Proof-of-Play research model into a real replicated, non-economic testnet. It is intentionally small enough to start with four phones and one development computer.

## Bootstrap topology

The genesis validator set contains four known participant devices. While the active committee is 15 members or fewer, finality uses a 3/4 quorum; therefore the four-device bootstrap requires three distinct valid committee signatures. Two unavailable genesis devices halt finality rather than lowering the threshold.

A validator device and a full node are different identities. Participant devices sign committee proposals/votes with their enrolled device keys. Full nodes store state, verify signatures, relay messages, synchronize blocks, and serve clients. Running 100 extra full nodes creates zero additional committee votes.

The mature committee target remains 64, but v0.9 does not require 64 phones or 64 servers. When fewer than the target are active, every active validator is selected. Above the target, authority influences deterministic committee-selection probability. Once selected, every committee member has exactly one vote.

## Validator activation

A newly attested device is a validator candidate, not an active validator. The v0.9 default research policy is:

- minimum authority: 100;
- maturation: 30 days;
- activation window: 7 days;
- maximum new activations per window: 1.

This is a bootstrap safety mechanism, not a unique-human proof. A sufficiently patient real-device farm can still accumulate candidates. The rate limit exists to make sudden capture impossible and slow long-horizon capture enough to observe and study it before permissionless enrollment is attempted.

The game remains immediately usable by new players. Validator maturation only limits consensus influence.

## Device eligibility

The protocol normalizes provider-specific attestation into one device-eligibility view.

Android consensus candidates require a verified Google Play Integrity result, the configured package/signing certificate, `PLAY_RECOGNIZED`, `LICENSED`, and the strong hardware-backed integrity tier for `productionEligible`.

Apple consensus candidates require verified App Attest evidence tied to the configured Team ID + bundle ID, a valid certificate chain to the configured App Attest trust root, the App Attest nonce, environment AAGUID, credential/key binding, and a P-256 attested public key. The Apple trust root is supplied by the operator so it can be rotated without recompiling the binary.

Neither provider proves that one device equals one independent human.

## Non-economic boundary

v0.9 has no economically valuable CARROT, token rewards, staking, or purchases that affect consensus. The milestone exists to break consensus, persistence, networking, and activation assumptions before economic incentives are introduced.

## Security properties under test

The testnet rejects wrong-network messages, invalid previous hashes, invalid state roots, unknown validators, invalid signatures, duplicate votes, equivocation, unknown permissioned peers, replayed peer envelopes, oversized bodies, and excessive concurrent requests from one source address. Stored chains are reverified on restart, and stale nodes only catch up by importing finalized blocks that independently pass certificate verification.

These controls do not make the system production-secure. v0.9 remains a permissioned research network.
