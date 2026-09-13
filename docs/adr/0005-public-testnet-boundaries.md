# ADR 0005: Public nodes, validators, and TEST-CARROT remain separate

## Status

Accepted for v0.11 adversarial testing.

## Decision

Public node discovery is permissionless but node count conveys no consensus power. Proof-of-Play validator applications may be submitted publicly, while attestation/authority verification and maturation/rate-limited activation remain mandatory. TEST-CARROT wallet keys are independent from infrastructure node keys and participant/device validator keys.

The public test asset is explicitly valueless. Its signed transaction/mempool layer exists to test replay, flood, signature, nonce, and accounting behavior without creating a financially valuable network.

Treasury governance is dry-run only while treasury spending remains disabled.

## Consequences

A node farm can consume bounded networking resources but cannot manufacture committee votes. A real-device farm remains the more important consensus threat and continues through the activation gate. Wallet compromise affects test balances, not validator identity.
